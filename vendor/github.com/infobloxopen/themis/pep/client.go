// Package pep implements gRPC client for Policy Decision Point (PDP) server.
// PEP package (Policy Enforcement Point) wraps service part of golang gRPC
// protocol implementation. The protocol is defined by
// github.com/infobloxopen/themis/proto/service.proto. Its golang implementation
// can be found at github.com/infobloxopen/themis/pdp-service. PEP is able
// to work with single server as well as multiple servers balancing requests
// using round-robin approach.
package pep

//go:generate bash -c "mkdir -p $GOPATH/src/github.com/infobloxopen/themis/pdp-service && protoc -I $GOPATH/src/github.com/infobloxopen/themis/proto/ $GOPATH/src/github.com/infobloxopen/themis/proto/service.proto --go_out=plugins=grpc:$GOPATH/src/github.com/infobloxopen/themis/pdp-service && ls $GOPATH/src/github.com/infobloxopen/themis/pdp-service"

import (
	"errors"
	"time"

	ot "github.com/opentracing/opentracing-go"
)

var (
	// ErrorConnected occurs if method connect is called after connection has been established.
	ErrorConnected = errors.New("connection has been already established")
	// ErrorNotConnected indicates that there is no connection established to PDP server.
	ErrorNotConnected = errors.New("no connection")
	// ErrorHotSpotBalancerUnsupported returned by attempt to make unary connection with
	// "hot spot" balancer.
	ErrorHotSpotBalancerUnsupported = errors.New("\"hot spot\" balancer isn't supported by unary gRPC client")
)

// Client defines abstract PDP service client interface.
//
// Marshalling and unmarshalling
//
// Validate method accepts as "in" argument any structure and pointer to
// any structure as "out" argument. If "in" argument is Request structure from
// github.com/infobloxopen/themis/pdp-service package, Validate passes it as is
// to server. Similarly if "out" argument is pointer to Response structure
// from the protocol package, Validate just copy data from server's response
// to the structure.
//
// If "in" argument is just a structure, Validate marshals it to list of PDP
// attributes. If no fields contains format string, Validate tries to convert
// all exported fields to attributes. Any bool field is converted to boolean
// attribute, string - to string attribute, net.IP - to address, net.IPNet or
// *net.IPNet to network. Fields of other types are silently ingnored.
//
// Marshalling can be ajusted more precisely with help of `pdp` key in format
// string. When some fields of "in" structure have format string, only fields
// with "pdp" key are converted to attributes. The key supports two option
// separated by comma. First is desired attribute name. Second - attribute type.
// Allowed types are: boolean, string, address, network and domain. Validate can
// convert only bool structure field to boolean attribute, string to string
// attribute, net.IP to address attribute, net.IPNet or *net.IPNet to network
// attribute and string to domain attribute.
//
// Validate is also able to unmarshal server's response to structure.
// It accepts pointer to the structure as "out" argument. If no fields
// of the structure has format string, Validate assigns effect to Effect field,
// reason to Reason field and obligation attributes to other fields
// according to their names and types. Effect field can be of bool type
// (and becomes true if effect is Permit or false otherwise), integer (it gets
// one of Response_* constants form pdp-service package) or string (gets
// Response_Effect_name value). Reason should be a string field. Obligation
// attributes are assigned to fields with corresponding names if
// types of fields allow assignment if there is no field with appropriate
// name and type response attribute silently dropped. The same as for marshaling
// `pdp` key can control unmarshaling.
type Client interface {
	// Connect establishes connection to given PDP server. It ignores address
	// parameter if balancer is provided.
	Connect(addr string) error
	// Close terminates previously established connection if any.
	// Close should silently return if connection hasn't been established yet or
	// if it has been already closed.
	Close()

	// Validate sends decision request to PDP server and fills out response.
	Validate(in, out interface{}) error
}

// An Option sets such options as balancer, tracer and number of streams.
type Option func(*options)

const virtualServerAddress = "pdp"

// WithRoundRobinBalancer returns an Option which sets round-robin balancer with given set of servers.
func WithRoundRobinBalancer(addresses ...string) Option {
	return func(o *options) {
		o.balancer = roundRobinBalancer
		o.addresses = addresses
	}
}

// WithHotSpotBalancer returns an Option which sets "hot spot" balancer with given set of servers
// (the balancer can be applied for gRPC streaming connection).
func WithHotSpotBalancer(addresses ...string) Option {
	return func(o *options) {
		o.balancer = hotSpotBalancer
		o.addresses = addresses
	}
}

// WithTracer returns an Option which sets OpenTracing tracer.
func WithTracer(tracer ot.Tracer) Option {
	return func(o *options) {
		o.tracer = tracer
	}
}

// WithStreams returns an Option which sets number of gRPC streams to run in parallel.
func WithStreams(n int) Option {
	return func(o *options) {
		o.maxStreams = n
	}
}

// WithConnectionTimeout returns an Option which sets validation timeout
// for the case when no connection can be established. Negative value means
// no timeout. Zero - don't wait for connection, fail immediately.
func WithConnectionTimeout(timeout time.Duration) Option {
	return func(o *options) {
		o.connTimeout = timeout
	}
}

// WithConnectionStateNotification returns an Option which sets connection
// state notification callback. The callback is called before connection
// attempt with state StreamingConnectionConnecting, on successfull connect
// with state StreamingConnectionEstablished. If connection attempt fails
// the callback is called with state StreamingConnectionFailure and with error
// occured during the attempt. State StreamingConnectionBroken is used when
// during request validation connection to any PDP server appears not working.
func WithConnectionStateNotification(callback ConnectionStateNotificationCallback) Option {
	return func(o *options) {
		o.connStateCb = callback
	}
}

const (
	noBalancer = iota
	roundRobinBalancer
	hotSpotBalancer
)

type options struct {
	addresses   []string
	balancer    int
	tracer      ot.Tracer
	maxStreams  int
	connTimeout time.Duration
	connStateCb ConnectionStateNotificationCallback
}

// NewClient creates client instance using given options.
func NewClient(opts ...Option) Client {
	o := options{
		connTimeout: -1,
	}
	for _, opt := range opts {
		opt(&o)
	}

	if o.maxStreams > 0 {
		return newStreamingClient(o)
	}

	return newUnaryClient(o)
}
