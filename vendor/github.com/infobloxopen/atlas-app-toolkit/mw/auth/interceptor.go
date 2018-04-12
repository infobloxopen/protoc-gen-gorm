package auth

import (
	"context"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/tags/logrus"
	pdp "github.com/infobloxopen/themis/pdp-service"
	"github.com/infobloxopen/themis/pep"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

var (
	ErrInternal     = grpc.Errorf(codes.Internal, "unable to process request")
	ErrUnauthorized = grpc.Errorf(codes.PermissionDenied, "unauthorized")
)

// Builder is responsible for creating requests to Themis. The response
// from Themis will determine if a request is authorized or unauthorized
type Builder interface {
	// build(...) uses the incoming request context to make a separate request
	// to Themis
	build(context.Context) (pdp.Request, error)
}

// Handler decides whether or not a request from Themis is authorized
type Handler interface {
	// handle(...) takes the response from Themis and return a boolean (true for
	// authorized, false for unauthorized)
	handle(context.Context, pdp.Response) (bool, error)
}

// attributer uses the context to build a slice of attributes
type attributer func(context.Context) ([]*pdp.Attribute, error)

// defaultBuilder provides a default implementation of the Builder interface
type defaultBuilder struct{ getters []attributer }

// build makes pdp.Request objects based on all the options provived by the
// user (e.g. WithJWT or WithRules)
func (d defaultBuilder) build(ctx context.Context) (pdp.Request, error) {
	attributes := []*pdp.Attribute{}
	for _, getter := range d.getters {
		attrs, err := getter(ctx)
		if err != nil {
			return pdp.Request{}, err
		}
		attributes = combineAttributes(attributes, attrs)
	}
	return pdp.Request{attributes}, nil
}

// NewBuilder returns an instance of the default Builder that includes all of
// of the user-provided options
func NewBuilder(opts ...option) Builder {
	db := defaultBuilder{}
	for _, opt := range opts {
		opt(&db)
	}
	return db
}

// defaultHandler provides a default implementation of the Handler interface
type defaultHandler struct{}

// handle denies all incoming requests that do not generate a PERMIT response
// from Themis
func (defaultHandler) handle(ctx context.Context, res pdp.Response) (bool, error) {
	if res.Effect != pdp.Response_PERMIT {
		return false, nil
	}
	return true, nil
}

// NewHandler returns an instance of the default handler
func NewHandler() Handler { return defaultHandler{} }

// Authorizer glues together a Builder and a Handler. It is responsible for
// sending requests and receiving responses to/from Themis
type Authorizer struct {
	PDPAddress string
	Bldr       Builder
	Hdlr       Handler
}

// AuthFunc builds the "AuthFunc" using the pep client that comes with Themis
func (a Authorizer) AuthFunc() grpc_auth.AuthFunc {
	clientFactory := func() pep.Client {
		return pep.NewClient(
			pep.WithConnectionTimeout(time.Second * 2),
		)
	}
	return a.authFunc(clientFactory)
}

// authFunc builds the "AuthFunc" type, which is the function that gets called
// by the authorization interceptor. The AuthFunc type is part of the gRPC
// authorization library, so a detailed explanation can be found here:
// https://github.com/grpc-ecosystem/go-grpc-middleware/blob/master/auth/auth.go
func (a Authorizer) authFunc(factory func() pep.Client) grpc_auth.AuthFunc {
	return func(ctx context.Context) (context.Context, error) {
		logger := ctx_logrus.Extract(ctx)
		pepClient := factory()
		// open connection to themis
		if err := pepClient.Connect(a.PDPAddress); err != nil {
			logger.Errorf("failed connecting to themis: %v", err)
			return ctx, ErrInternal
		}
		defer pepClient.Close()
		// build a pdp request and send it to themis
		req, err := a.Bldr.build(ctx)
		if err != nil {
			logger.Errorf("failed building themis request: %v", err)
			return ctx, err
		}
		res := pdp.Response{}
		if err := pepClient.Validate(req, &res); err != nil {
			logger.Errorf("error sending message to themis: %v", err)
			return ctx, ErrInternal
		}
		logger.Infof("themis response: %v", res)
		// handle response from themis
		authorized, err := a.Hdlr.handle(ctx, res)
		if err != nil {
			logger.Errorf("error handling response from themis: %v", err)
			return ctx, ErrInternal
		}
		if !authorized {
			logger.Info("request unauthorized")
			return ctx, ErrUnauthorized
		}
		logger.Info("request authorized")
		return ctx, nil
	}
}
