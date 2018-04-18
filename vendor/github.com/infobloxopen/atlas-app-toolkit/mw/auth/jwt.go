package auth

import (
	"context"
	"errors"
	"fmt"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/grpc-ecosystem/go-grpc-middleware/auth"
)

const (
	// TODO: Field is tentatively called "AccountID" but will probably need to be
	// changed. We don't know what the JWT will look like, so we're giving it our
	// best guess for the time being.
	MULTI_TENANCY_FIELD = "AccountID"
)

var (
	errMissingField     = errors.New("unable to get field from token")
	errMissingToken     = errors.New("unable to get token from context")
	errInvalidAssertion = errors.New("unable to assert value as jwt.MapClaims")
)

func GetJWTField(ctx context.Context, field string, keyfunc jwt.Keyfunc) (string, error) {
	token, err := getToken(ctx, keyfunc)
	if err != nil {
		return "", errMissingToken
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errInvalidAssertion
	}
	jwtField, ok := claims[field]
	if !ok {
		return "", errMissingField
	}
	return fmt.Sprint(jwtField), nil
}

func GetAccountID(ctx context.Context, keyfunc jwt.Keyfunc) (string, error) {
	return GetJWTField(ctx, MULTI_TENANCY_FIELD, keyfunc)
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
