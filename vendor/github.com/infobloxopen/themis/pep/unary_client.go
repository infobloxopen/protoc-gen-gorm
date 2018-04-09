package pep

import (
	"fmt"
	"sync"

	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	ot "github.com/opentracing/opentracing-go"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/infobloxopen/themis/pdp-service"
)

type unaryClient struct {
	lock   *sync.RWMutex
	conn   *grpc.ClientConn
	client *pb.PDPClient

	opts options
}

func newUnaryClient(opts options) *unaryClient {
	return &unaryClient{
		lock: &sync.RWMutex{},
		opts: opts,
	}
}

func (c *unaryClient) Connect(addr string) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.conn != nil {
		return ErrorConnected
	}

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
	}

	if len(c.opts.addresses) > 0 {
		addr = virtualServerAddress
		switch c.opts.balancer {
		default:
			panic(fmt.Errorf("invalid balancer %d", c.opts.balancer))

		case roundRobinBalancer:
			opts = append(opts, grpc.WithBalancer(grpc.RoundRobin(newStaticResolver(addr, c.opts.addresses...))))

		case hotSpotBalancer:
			return ErrorHotSpotBalancerUnsupported
		}
	}

	if c.opts.tracer != nil {
		opts = append(opts,
			grpc.WithUnaryInterceptor(
				otgrpc.OpenTracingClientInterceptor(
					c.opts.tracer,
					otgrpc.IncludingSpans(
						func(parentSpanCtx ot.SpanContext, method string, req, resp interface{}) bool {
							return parentSpanCtx != nil
						},
					),
				),
			),
		)
	}

	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		return err
	}

	c.conn = conn

	client := pb.NewPDPClient(c.conn)
	c.client = &client

	return nil
}

func (c *unaryClient) Close() {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}

	c.client = nil
}

func (c *unaryClient) Validate(in, out interface{}) error {
	c.lock.RLock()
	uc := c.client
	c.lock.RUnlock()

	if uc == nil {
		return ErrorNotConnected
	}

	req, err := makeRequest(in)
	if err != nil {
		return err
	}

	res, err := (*uc).Validate(context.Background(), &req, grpc.FailFast(false))
	if err != nil {
		return err
	}

	return fillResponse(res, out)
}
