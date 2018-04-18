package auth

import (
	"context"
	"fmt"
	"path"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	pdp "github.com/infobloxopen/themis/pdp-service"
	"google.golang.org/grpc"
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
		service, method, err := getRequestDetails(ctx)
		if err != nil {
			return nil, err
		}
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

func getRequestDetails(ctx context.Context) (string, string, error) {
	fullMethodString, ok := grpc.Method(ctx)
	if !ok {
		return "", "", ErrInternal
	}
	fmt.Println(fullMethodString)
	return path.Dir(fullMethodString)[1:], path.Base(fullMethodString), nil
}

func combineAttributes(first, second []*pdp.Attribute) []*pdp.Attribute {
	for _, attr := range second {
		first = append(first, attr)
	}
	return first
}
