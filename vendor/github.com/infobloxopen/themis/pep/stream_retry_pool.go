package pep

import "sync"

type streamRetryPool struct {
	limit int
	index map[int]uint64
	ver   uint64

	sync.Mutex
}

func newStreamRetryPool(limit int) *streamRetryPool {
	return &streamRetryPool{
		limit: limit,
		index: make(map[int]uint64, limit+1),
	}
}

func (p *streamRetryPool) push(i int) (uint64, bool) {
	p.Lock()
	defer p.Unlock()

	p.ver++
	p.index[i] = p.ver
	return p.ver, len(p.index) < p.limit
}

func (p *streamRetryPool) pop(i int, ver uint64) {
	p.Lock()
	defer p.Unlock()

	if v, ok := p.index[i]; !ok || v == ver {
		delete(p.index, i)
	}
}

func (p *streamRetryPool) flush() {
	p.Lock()
	defer p.Unlock()

	p.index = make(map[int]uint64, p.limit+1)
}
