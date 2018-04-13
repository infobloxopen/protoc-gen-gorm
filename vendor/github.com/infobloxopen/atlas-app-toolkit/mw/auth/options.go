package auth

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/grpc-ecosystem/go-grpc-middleware/auth"
	pdp "github.com/infobloxopen/themis/pdp-service"
	"google.golang.org/grpc/transport"
)

// functional options for the defaultBuilder
type option func(*defaultBuilder)

// WithJWT allows for token-based authorization using JWT. When WithJWT has been
// added as a build parameter, every field in the token payload will be included
// in the request to Themis
func WithJWT(keyfunc jwt.Keyfunc) option {
	withTokenJWTFunc := func(ctx context.Context) ([]*pdp.Attribute, error) {
		attributes := []*pdp.Attribute{}
		token, err := getToken(ctx, keyfunc)
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

// getToken parses the token into a jwt.Token type from the grpc metadata.
// WARNING: if keyfunc is nil, the token will get parsed but not verified
// because it has been checked previously in the stack. More information
// here: https://godoc.org/github.com/dgrijalva/jwt-go#Parser.ParseUnverified
func getToken(ctx context.Context, keyfunc jwt.Keyfunc) (jwt.Token, error) {
	tokenStr, err := grpc_auth.AuthFromMD(ctx, "token")
	if err != nil {
		return jwt.Token{}, ErrUnauthorized
	}
	parser := jwt.Parser{}
	if keyfunc != nil {
		token, err := parser.Parse(tokenStr, keyfunc)
		if err != nil {
			return jwt.Token{}, ErrUnauthorized
		}
		return *token, nil
	}
	token, _, err := parser.ParseUnverified(tokenStr, jwt.MapClaims{})
	if err != nil {
		return jwt.Token{}, ErrUnauthorized
	}
	return *token, nil
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

// WithRequest takes metadata from the incoming request and passes it
// to Themis in the authorization request. Specifically, this includes the gRPC
// service name (e.g. AddressBook) and the corresponding function that is
// called by the client (e.g. ListPersons)
func WithRequest() option {
	withRequestFunc := func(ctx context.Context) ([]*pdp.Attribute, error) {
		stream, ok := transport.StreamFromContext(ctx)
		if !ok {
			return nil, errors.New("failed getting stream from context")
		}
		service, method := getRequestDetails(*stream)
		service = stripPackageName(service)
		attributes := []*pdp.Attribute{
			&pdp.Attribute{"operation", "string", method},
			// lowercase the service to match PARG naming conventions
			&pdp.Attribute{"application", "string", strings.ToLower(service)},
		}
		return attributes, nil
	}
	return func(d *defaultBuilder) {
		d.getters = append(d.getters, withRequestFunc)
	}
}

// stripPackageName removes the package name prefix from a fully-qualified
// proto service name
func stripPackageName(service string) string {
	fields := strings.Split(service, ".")
	return fields[len(fields)-1]
}

func getRequestDetails(stream transport.Stream) (service, method string) {
	fullMethodString := stream.Method()
	return path.Dir(fullMethodString)[1:], path.Base(fullMethodString)
}

func combineAttributes(first, second []*pdp.Attribute) []*pdp.Attribute {
	for _, attr := range second {
		first = append(first, attr)
	}
	return first
}
