package pep

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/transport"

	pb "github.com/infobloxopen/themis/pdp-service"
)

var (
	errStreamWrongState = errors.New("can't make operation with the stream")
	errStreamFailure    = errors.New("stream failed")
)

func closeStreams(toClose, toDrop []*stream) {
	wg := &sync.WaitGroup{}
	for _, s := range toClose {
		wg.Add(1)
		go s.closeStream(wg)
	}

	for _, s := range toDrop {
		s.drop()
	}

	wg.Wait()
}

type stream struct {
	parent *streamConn
	stream *atomic.Value
}

func (c *streamConn) newStream() *stream {
	s := &stream{
		parent: c,
		stream: new(atomic.Value),
	}
	s.drop()
	return s
}

func (s *stream) connect() error {
	sp := s.stream.Load().(*pb.PDP_NewValidationStreamClient)
	if sp != nil {
		return errStreamWrongState
	}

	ss, err := s.parent.newValidationStream()
	if err != nil {
		return err
	}

	sp = &ss
	s.stream.Store(sp)
	return nil
}

func (s *stream) closeStream(wg *sync.WaitGroup) {
	defer wg.Done()

	sp := s.stream.Load().(*pb.PDP_NewValidationStreamClient)
	if sp == nil {
		return
	}
	s.drop()

	if err := (*sp).CloseSend(); err != nil {
		return
	}

	done := make(chan int)
	go func() {
		defer close(done)
		(*sp).Recv()
	}()

	t := time.NewTimer(closeWaitDuration)
	select {
	case <-done:
		if !t.Stop() {
			<-t.C
		}
	case <-t.C:
	}
}

func (s *stream) drop() {
	var ssNil *pb.PDP_NewValidationStreamClient
	s.stream.Store(ssNil)
}

func (s *stream) validate(in, out interface{}) error {
	sp := s.stream.Load().(*pb.PDP_NewValidationStreamClient)
	if sp == nil {
		return errStreamWrongState
	}

	req, err := makeRequest(in)
	if err != nil {
		return err
	}

	err = (*sp).Send(&req)
	if err != nil {
		if err == transport.ErrConnClosing || err == balancer.ErrTransientFailure {
			return errConnFailure
		}

		return errStreamFailure
	}

	res, err := (*sp).Recv()
	if err != nil {
		if err == transport.ErrConnClosing || err == balancer.ErrTransientFailure {
			return errConnFailure
		}

		return errStreamFailure
	}

	return fillResponse(res, out)
}
