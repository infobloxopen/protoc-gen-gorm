package pep

import "google.golang.org/grpc/naming"

type staticWatcher struct {
	update []*naming.Update
	stop   chan bool
	sent   bool
}

func newStaticWatcher(addrs []string) *staticWatcher {
	w := &staticWatcher{
		update: make([]*naming.Update, len(addrs)),
		stop:   make(chan bool),
	}

	for i, addr := range addrs {
		w.update[i] = &naming.Update{Op: naming.Add, Addr: addr}
	}

	return w
}

func (w *staticWatcher) Next() ([]*naming.Update, error) {
	if w.sent {
		<-w.stop
		return nil, nil
	}

	w.sent = true
	return w.update, nil
}

func (w *staticWatcher) Close() {
	w.stop <- true
}
