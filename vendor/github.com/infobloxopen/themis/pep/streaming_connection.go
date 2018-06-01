package pep

//go:generate bash -c "mkdir -p $GOPATH/src/github.com/infobloxopen/themis/pdp-service && protoc -I $GOPATH/src/github.com/infobloxopen/themis/proto/ $GOPATH/src/github.com/infobloxopen/themis/proto/service.proto --go_out=plugins=grpc:$GOPATH/src/github.com/infobloxopen/themis/pdp-service && ls $GOPATH/src/github.com/infobloxopen/themis/pdp-service"

import (
	"errors"
	"sync"
	"time"

	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/opentracing/opentracing-go"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/infobloxopen/themis/pdp-service"
)

// ConnectionStateNotificationCallback is a function type for connection state
// notifications.
type ConnectionStateNotificationCallback func(address string, state int, err error)

// StreamingConnection* constants designate different states of connection
// to particular PDP on notification callback call.
const (
	// StreamingConnectionEstablished stands for succesfully established connection.
	StreamingConnectionEstablished = iota
	// StreamingConnectionBroken is passed to notification callback when
	// connection appears broken during a validation call.
	StreamingConnectionBroken
	// StreamingConnectionConnecting marks a connection attempt.
	StreamingConnectionConnecting
	// StreamingConnectionFailure used when a connection attempt fails.
	// In the case err gets value of an error occured.
	StreamingConnectionFailure
)

const connectionResetPercent float64 = 0.3

func makeStreamConns(addrs []string, streams int, tracer opentracing.Tracer, timeout time.Duration, cb ConnectionStateNotificationCallback) ([]*streamConn, *connRetryPool) {
	total := len(addrs)
	if total > streams {
		total = streams
	}

	conns := make([]*streamConn, total)
	chunk := streams / total
	rem := streams % total
	for i := range conns {
		count := chunk
		if i < rem {
			count++
		}

		conns[i] = newStreamConn(addrs[i], count, tracer, cb)
	}

	crp := newConnRetryPool(conns, timeout)
	for _, c := range conns {
		c.crp = crp
	}

	return conns, crp
}

func closeStreamConns(conns []*streamConn) {
	for _, c := range conns {
		if c != nil {
			c.closeConn()
		}
	}
}

const (
	scisDisconnected uint32 = iota
	scisConnecting
	scisConnected
	scisClosing
)

var (
	errStreamConnWrongState = errors.New("can't make operation with the connection")
	errConnFailure          = errors.New("connection failed")
)

var closeWaitDuration = 5 * time.Second

type streamConn struct {
	addr   string
	tracer opentracing.Tracer
	crp    *connRetryPool
	limit  int

	state uint32
	lock  *sync.RWMutex

	conn    *grpc.ClientConn
	client  pb.PDPClient
	streams []*stream
	index   chan int
	retry   chan boundStream

	notify ConnectionStateNotificationCallback
}

func newStreamConn(addr string, streams int, tracer opentracing.Tracer, cb ConnectionStateNotificationCallback) *streamConn {
	c := &streamConn{
		addr:    addr,
		tracer:  tracer,
		limit:   int(float64(streams)*connectionResetPercent + 0.5),
		lock:    new(sync.RWMutex),
		streams: make([]*stream, streams),
		notify:  cb,
	}

	for i := range c.streams {
		c.streams[i] = c.newStream()
	}

	return c
}

func (c *streamConn) connect() error {
	addr, tracer, err := c.enterConnect()
	if err != nil {
		return err
	}

	for {
		if err := c.tryConnect(addr, tracer, time.Second); err == nil || err == errStreamConnWrongState {
			break
		}
	}

	return nil
}

func (c *streamConn) enterConnect() (string, opentracing.Tracer, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.state != scisDisconnected {
		return "", nil, errStreamConnWrongState
	}

	c.state = scisConnecting
	return c.addr, c.tracer, nil
}

func (c *streamConn) exitConnect() bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.state == scisClosing {
		closeStreams(c.streams, nil)
		c.closeConnInternal()
		return false
	}

	c.index = make(chan int, len(c.streams))
	for i := range c.streams {
		c.index <- i
	}

	c.retry = make(chan boundStream)
	go c.retryWorker(c.retry)

	c.state = scisConnected
	return true
}

func (c *streamConn) tryConnect(addr string, tracer opentracing.Tracer, timeout time.Duration) (err error) {
	if c.notify != nil {
		go c.notify(addr, StreamingConnectionConnecting, nil)
		defer func() {
			state := StreamingConnectionEstablished
			if err != nil {
				state = StreamingConnectionFailure
			}

			go c.notify(addr, state, err)
		}()
	}
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.FailOnNonTempDialError(true),
	}

	if tracer != nil {
		opts = append(opts,
			grpc.WithUnaryInterceptor(
				otgrpc.OpenTracingClientInterceptor(
					tracer,
					otgrpc.IncludingSpans(inclusionFunc),
				),
			),
		)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var conn *grpc.ClientConn
	conn, err = grpc.DialContext(ctx, addr, opts...)
	if err != nil {
		return
	}

	client := pb.NewPDPClient(conn)
	c.lock.Lock()
	c.conn = conn
	c.client = client
	c.lock.Unlock()

	var ready int
	ready, err = c.connectStreams()
	if err != nil {
		c.lock.Lock()
		defer c.lock.Unlock()

		closeStreams(c.streams[0:ready], nil)
		c.closeConnInternal()

		return
	}

	if !c.exitConnect() {
		err = errStreamConnWrongState
	}

	return
}

func (c *streamConn) newValidationStream() (pb.PDP_NewValidationStreamClient, error) {
	c.lock.RLock()
	state := c.state
	client := c.client
	c.lock.RUnlock()

	if (state != scisConnected && state != scisConnecting) || client == nil {
		return nil, errStreamConnWrongState
	}

	return client.NewValidationStream(context.TODO())
}

func (c *streamConn) connectStreams() (int, error) {
	for i, s := range c.streams {
		err := s.connect()
		if err != nil {
			return i, err
		}
	}

	return 0, nil
}

func (c *streamConn) withdrawStreams() ([]*stream, []*stream) {
	c.lock.RLock()
	index := c.index
	c.lock.RUnlock()

	if index == nil {
		return nil, nil
	}

	toClose := []*stream{}
	toDrop := []*stream{}
	ready := tryRead(index)
	for i, s := range c.streams {
		if _, ok := ready[i]; ok {
			toClose = append(toClose, s)
		} else {
			toDrop = append(toDrop, s)
		}
	}

	return toClose, toDrop
}

func tryRead(ch chan int) map[int]bool {
	streams := make(map[int]bool)

	for len(ch) > 0 {
		select {
		default:
		case i, ok := <-ch:
			if !ok {
				return streams
			}

			streams[i] = true
		}
	}

	if len(streams) < cap(ch) {
		t := time.NewTimer(closeWaitDuration)
		for len(streams) < cap(ch) {
			select {
			case i, ok := <-ch:
				if !ok {
					if !t.Stop() {
						<-t.C
					}

					return streams
				}

				streams[i] = true

			case <-t.C:
				return streams
			}
		}

		if !t.Stop() {
			<-t.C
		}
	}

	return streams
}

func (c *streamConn) closeConn() {
	if !c.enterCloseConn() {
		return
	}

	toClose, toDrop := c.withdrawStreams()

	c.lock.Lock()
	defer c.lock.Unlock()

	closeStreams(toClose, toDrop)
	c.closeConnInternal()
}

func (c *streamConn) enterCloseConn() bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.state == scisConnected {
		c.state = scisClosing

		close(c.retry)
		c.retry = nil

		return true
	}

	if c.state == scisConnecting {
		c.state = scisClosing

		if c.retry != nil {
			close(c.retry)
			c.retry = nil
		}
	}

	return false
}

func (c *streamConn) closeConnInternal() {
	if c.index != nil {
		close(c.index)
		c.index = nil
	}

	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	c.client = nil
	c.state = scisDisconnected
}

func (c *streamConn) markDisconnected() bool {
	if c.notify != nil {
		go c.notify(c.addr, StreamingConnectionBroken, nil)
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	if c.state != scisConnected {
		return false
	}

	close(c.retry)
	c.retry = nil

	close(c.index)
	c.index = nil

	closeStreams(nil, c.streams)
	c.closeConnInternal()
	return true
}

type boundStream struct {
	s     *stream
	idx   int
	index chan int
	retry chan boundStream
}

func (c *streamConn) getStream() (boundStream, error) {
	c.lock.RLock()
	state := c.state
	index := c.index
	retry := c.retry
	c.lock.RUnlock()

	if state == scisConnected && index != nil {
		if i, ok := <-index; ok {
			return boundStream{
				s:     c.streams[i],
				idx:   i,
				index: index,
				retry: retry,
			}, nil
		}
	}

	return boundStream{}, errStreamConnWrongState
}

func (c *streamConn) tryGetStream() (boundStream, bool, error) {
	c.lock.RLock()
	state := c.state
	index := c.index
	retry := c.retry
	c.lock.RUnlock()

	if state == scisConnected && index != nil {
		select {
		default:
			return boundStream{}, false, nil

		case i, ok := <-index:
			if ok {
				return boundStream{
					s:     c.streams[i],
					idx:   i,
					index: index,
					retry: retry,
				}, true, nil
			}
		}
	}

	return boundStream{}, false, errStreamConnWrongState
}

func (c *streamConn) putStream(s boundStream) error {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if c.state != scisConnected && c.state != scisClosing || c.index != s.index {
		return errStreamConnWrongState
	}

	s.index <- s.idx
	return nil
}

func (c *streamConn) validate(in, out interface{}) error {
	s, err := c.getStream()
	if err != nil {
		return err
	}

	err = s.s.validate(in, out)
	if err != nil {
		c.lock.RLock()
		defer c.lock.RUnlock()
		if err == errStreamFailure && c.retry == s.retry {
			s.retry <- s
		}

		return err
	}

	c.putStream(s)
	return nil
}

func (c *streamConn) tryValidate(in, out interface{}) (bool, error) {
	s, ok, err := c.tryGetStream()
	if err != nil {
		return false, err
	}

	if !ok {
		return false, nil
	}

	err = s.s.validate(in, out)
	if err != nil {
		c.lock.RLock()
		defer c.lock.RUnlock()
		if err == errStreamFailure && c.retry == s.retry {
			s.retry <- s
		}

		return true, err
	}

	c.putStream(s)
	return true, nil
}

func (c *streamConn) retryWorker(retry chan boundStream) {
	pool := newStreamRetryPool(c.limit)
	for s := range retry {
		c.lock.RLock()
		state := c.state
		c.lock.RUnlock()

		if state == scisClosing {
			c.putStream(s)
			continue
		}

		ver, ok := pool.push(s.idx)
		if !ok {
			c.putStream(s)
			pool.flush()
			go c.crp.put(c)
			continue
		}

		go func(s boundStream, ver uint64) {
			defer func() {
				c.putStream(s)
				pool.pop(s.idx, ver)
			}()

			s.s.drop()
			if err := s.s.connect(); err != nil {
				c.crp.put(c)
			}
		}(s, ver)
	}
}

func inclusionFunc(parentSpanCtx opentracing.SpanContext, method string, req, resp interface{}) bool {
	return parentSpanCtx != nil
}
