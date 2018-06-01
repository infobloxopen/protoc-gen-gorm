package pep

import (
	"fmt"
	"sync/atomic"
)

const (
	scsDisconnected uint32 = iota
	scsConnecting
	scsConnected
	scsClosing
	scsClosed
)

type validator func(in, out interface{}) error

type streamingClient struct {
	opts options

	state    *uint32
	conns    []*streamConn
	counter  *uint64
	validate validator

	crp *connRetryPool
}

func newStreamingClient(opts options) *streamingClient {
	if opts.maxStreams <= 0 {
		panic(fmt.Errorf("streaming client must be created with at least 1 stream but got %d", opts.maxStreams))
	}

	state := scsDisconnected
	counter := uint64(0)
	return &streamingClient{
		opts:    opts,
		state:   &state,
		counter: &counter,
	}
}

func (c *streamingClient) Connect(addr string) error {
	if !atomic.CompareAndSwapUint32(c.state, scsDisconnected, scsConnecting) {
		return ErrorConnected
	}

	exitState := scsDisconnected
	defer func() { atomic.StoreUint32(c.state, exitState) }()

	addrs := c.opts.addresses
	c.validate = c.makeSimpleValidator()
	if len(addrs) > 1 {
		switch c.opts.balancer {
		default:
			panic(fmt.Errorf("invalid balancer %d", c.opts.balancer))

		case roundRobinBalancer:
			c.validate = c.makeRoundRobinValidator()

		case hotSpotBalancer:
			c.validate = c.makeHotSpotValidator()
		}
	} else if len(addrs) < 1 {
		addrs = []string{addr}
	}

	conns, crp := makeStreamConns(addrs, c.opts.maxStreams, c.opts.tracer, c.opts.connTimeout, c.opts.connStateCb)
	c.conns = conns
	c.crp = crp

	exitState = scsConnected
	return nil
}

func (c *streamingClient) Close() {
	if !atomic.CompareAndSwapUint32(c.state, scsConnected, scsClosing) {
		return
	}

	c.crp.stop()
	closeStreamConns(c.conns)
	atomic.StoreUint32(c.state, scsClosed)
}

func (c *streamingClient) Validate(in, out interface{}) error {
	for atomic.LoadUint32(c.state) == scsConnected {
		if !c.crp.check() {
			c.crp.tryStart()
			if !c.crp.wait() {
				return ErrorNotConnected
			}
		}

		for i := 0; i < len(c.conns); i++ {
			err := c.validate(in, out)
			if err == nil {
				return nil
			}

			if err != errConnFailure &&
				err != errStreamFailure &&
				err != errStreamConnWrongState &&
				err != errStreamWrongState {
				return err
			}
		}
	}

	return ErrorNotConnected
}

func (c *streamingClient) makeSimpleValidator() validator {
	return func(in, out interface{}) error {
		conn := c.conns[0]
		err := conn.validate(in, out)
		if err == errConnFailure {
			c.crp.put(conn)
		}

		return err
	}
}

func (c *streamingClient) makeRoundRobinValidator() validator {
	return func(in, out interface{}) error {
		i := int((atomic.AddUint64(c.counter, 1) - 1) % uint64(len(c.conns)))
		conn := c.conns[i]
		err := conn.validate(in, out)
		if err == errConnFailure {
			c.crp.put(conn)
		}

		return err
	}
}

func (c *streamingClient) makeHotSpotValidator() validator {
	return func(in, out interface{}) error {
		total := uint64(len(c.conns))
		start := atomic.LoadUint64(c.counter)
		i := int(start % total)
		for {
			conn := c.conns[i]
			ok, err := conn.tryValidate(in, out)
			if ok {
				if err == errConnFailure {
					c.crp.put(conn)
				}

				return err
			}

			new := atomic.AddUint64(c.counter, 1)
			if new-start >= total {
				break
			}

			i = int(new % total)
		}

		conn := c.conns[i]
		err := conn.validate(in, out)
		if err == errConnFailure {
			c.crp.put(conn)
		}

		return err
	}
}
