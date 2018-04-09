package pep

import (
	"sync"
	"sync/atomic"
	"time"
)

const (
	crpIdle uint32 = iota
	crpWorking
	crpFull
	crpStopping
)

type connRetryPool struct {
	timeout time.Duration
	state   *uint32

	ch    chan *streamConn
	count *uint32

	c *sync.Cond
	m *sync.RWMutex
}

func newConnRetryPool(conns []*streamConn, timeout time.Duration) *connRetryPool {
	state := crpIdle

	ch := make(chan *streamConn, len(conns))
	for _, c := range conns {
		ch <- c
	}
	count := uint32(len(conns))

	return &connRetryPool{
		timeout: timeout,
		state:   &state,
		ch:      ch,
		count:   &count,
		c:       sync.NewCond(new(sync.Mutex)),
		m:       new(sync.RWMutex),
	}
}

func (p *connRetryPool) tryStart() {
	if atomic.CompareAndSwapUint32(p.state, crpIdle, crpFull) {
		go p.worker(p.ch)
	}
}

func (p *connRetryPool) stop() {
	atomic.StoreUint32(p.state, crpStopping)
	p.c.Broadcast()

	p.m.Lock()
	defer p.m.Unlock()

	close(p.ch)
	p.ch = nil
}

func (p *connRetryPool) check() bool {
	return atomic.LoadUint32(p.state) == crpWorking
}

func (p *connRetryPool) wait() bool {
	state := atomic.LoadUint32(p.state)
	if p.timeout < 0 {
		p.c.L.Lock()
		defer p.c.L.Unlock()

		for state != crpWorking && state != crpStopping {
			p.c.Wait()
			state = atomic.LoadUint32(p.state)
		}
	} else if p.timeout > 0 {
		sch := make(chan bool)

		done := make(chan bool)
		defer close(done)
		go func(state uint32) {
			p.c.L.Lock()
			defer func() {
				p.c.L.Unlock()
				close(sch)
			}()

			for state != crpWorking && state != crpStopping {
				p.c.Wait()
				select {
				default:
				case <-done:
					break
				}

				state = atomic.LoadUint32(p.state)
			}
		}(state)

		select {
		case <-time.After(p.timeout):
		case <-sch:
		}

		state = atomic.LoadUint32(p.state)
	}

	return state == crpWorking
}

func (p *connRetryPool) put(c *streamConn) {
	if c.markDisconnected() {
		p.m.RLock()
		defer p.m.RUnlock()

		if p.ch == nil {
			return
		}

		if atomic.AddUint32(p.count, 1) == uint32(cap(p.ch)) {
			atomic.CompareAndSwapUint32(p.state, crpWorking, crpFull)
		}

		p.ch <- c
	}
}

func (p *connRetryPool) worker(ch chan *streamConn) {
	for c := range ch {
		go p.reconnect(c)
	}
}

func (p *connRetryPool) reconnect(c *streamConn) {
	if err := c.connect(); err != nil {
		return
	}

	atomic.AddUint32(p.count, ^uint32(0))
	if atomic.CompareAndSwapUint32(p.state, crpFull, crpWorking) {
		p.c.Broadcast()
	}
}
