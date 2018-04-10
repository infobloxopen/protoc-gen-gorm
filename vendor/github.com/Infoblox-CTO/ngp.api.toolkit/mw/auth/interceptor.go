package auth

import (
	"context"
	"errors"
	"fmt"
	"path"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/tags/logrus"
	pdp "github.com/infobloxopen/themis/pdp-service"
	"github.com/infobloxopen/themis/pep"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/transport"
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

// WithJWT allows for token-based authorization using JWT. When WithJWT has been
// added as a build parameter, every field in the token payload will be included
// in the request to Themis
func WithJWT() option {
	withTokenJWTFunc := func(ctx context.Context) ([]*pdp.Attribute, error) {
		attributes := []*pdp.Attribute{}
		token, err := getToken(ctx)
		if err != nil {
			return attributes, ErrUnauthorized
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return attributes, ErrInternal
		}
		for k, v := range claims {
			attr := &pdp.Attribute{k, "string", fmt.Sprint(v)}
			attributes = append(attributes, attr)
		}
		return attributes, nil
	}
	return func(d *defaultBuilder) {
		d.getters = append(d.getters, withTokenJWTFunc)
	}
}

// WithRules enables operation-based authorization. Developers can map their
// backend operations to specific rules (e.g. read/write or CRUD) and WithRules
// will include the rule in its request to Themis
func WithRules(r rules) option {
	withRulesFunc := func(ctx context.Context) ([]*pdp.Attribute, error) {
		attributes := []*pdp.Attribute{}
		stream, ok := transport.StreamFromContext(ctx)
		if !ok {
			return nil, errors.New("failed getting stream from context")
		}
		_, method := getRequestDetails(*stream)
		operation, err := r.getRule(method)
		if err != nil {
			return attributes, err
		}
		attr := &pdp.Attribute{"operation", "string", operation}
		return append(attributes, attr), nil
	}
	return func(d *defaultBuilder) {
		d.getters = append(d.getters, withRulesFunc)
	}
}

// WithCallback allows developers to pass their own attributer to the
// authorization service. It gives them the flexibility to add customization to
// the auth process without needing to write a Builder from scratch.
func WithCallback(attr attributer) option {
	withCallbackFunc := func(ctx context.Context) ([]*pdp.Attribute, error) {
		return attr(ctx)
	}
	return func(d *defaultBuilder) {
		d.getters = append(d.getters, withCallbackFunc)
	}
}

// rules provide extra functionality to a basic map[string]string
type rules map[string]string

// getRule determines if a rule has been set for a given backend operation
func (r rules) getRule(rule string) (string, error) {
	val, ok := r[rule]
	if !ok {
		return val, errors.New("rule is not defined")
	}
	return val, nil
}

// defaultBuilder provides a default implementation of the Builder interface
type defaultBuilder struct {
	getters []attributer
}

// functional options for the defaultBuilder
type option func(*defaultBuilder)

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
func NewHandler() Handler {
	return defaultHandler{}
}

// Authorizer glues together a Builder and a Handler. It is responsible for
// sending requests and receiving responses to/from Themis
type Authorizer struct {
	PDPAddress string
	Bldr       Builder
	Hdlr       Handler
}

// AuthFunc builds the "AuthFunc" using the pep client that comes with Themis
func (a Authorizer) AuthFunc() grpc_auth.AuthFunc {
	pepClient := pep.NewClient()
	return a.authFunc(pepClient)
}

// authFunc builds the "AuthFunc" type, which is the function that gets called
// by the authorization interceptor. The AuthFunc type is part of the gRPC
// authorization library, so a detailed explanation can be found here:
// https://github.com/grpc-ecosystem/go-grpc-middleware/blob/master/auth/auth.go
func (a Authorizer) authFunc(pepClient pep.Client) grpc_auth.AuthFunc {
	return func(ctx context.Context) (context.Context, error) {
		logger := ctx_logrus.Extract(ctx)
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
			return ctx, ErrInternal
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

func combineAttributes(first, second []*pdp.Attribute) []*pdp.Attribute {
	for _, attr := range second {
		first = append(first, attr)
	}
	return first
}

func getRequestDetails(stream transport.Stream) (service, method string) {
	fullMethodString := stream.Method()
	return path.Dir(fullMethodString)[1:], path.Base(fullMethodString)
}

func getToken(ctx context.Context) (jwt.Token, error) {
	var token *jwt.Token
	tokenStr, err := grpc_auth.AuthFromMD(ctx, "token")
	if err != nil {
		return *token, ErrUnauthorized
	}
	// this parses the token into a jwt.Token type. if the token reaches the
	// interceptor, it has already been validated at the api gateway. this
	// is the reason why the token is parsed with a dummy secret.
	token, _ = jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte("dummy-secret"), nil
	})
	return *token, nil
}
