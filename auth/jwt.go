package auth

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	jwt "github.com/golang-jwt/jwt/v4"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
)

const (
	// MultiTenancyField the field name for a specific tenant
	MultiTenancyField = "account_id"

	// MultiCompartmentField the field name for a specific compartment
	MultiCompartmentField = "compartment_id"

	// AuthorizationHeader contains information about the header value for the token
	AuthorizationHeader = "Authorization"

	// DefaultTokenType is the name of the authorization token (e.g. "Bearer"
	// or "token")
	DefaultTokenType = "Bearer"
)

var (
	errMissingField     = errors.New("unable to get field from token")
	errMissingToken     = errors.New("unable to get token from context")
	errInvalidAssertion = errors.New("unable to assert token as jwt.MapClaims")

	// multiTenancyVariants all possible multi-tenant names
	multiTenancyVariants = []string{
		MultiTenancyField,
		"AccountID",
	}
)

// GetJWTFieldWithTokenType gets the JWT from a context and returns the
// specified field. The user must provide a token type, which prefixes the
// token itself (e.g. "Bearer" or "token")
func GetJWTFieldWithTokenType(ctx context.Context, tokenType, tokenField string, keyfunc jwt.Keyfunc) (string, error) {
	token, err := getToken(ctx, tokenType, keyfunc)
	if err != nil {
		return "", errMissingToken
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errInvalidAssertion
	}
	jwtField, ok := claims[tokenField]
	if !ok {
		return "", errMissingField
	}
	switch jwtField := jwtField.(type) {
	case float64:
		return strconv.FormatFloat(jwtField, 'f', -1, 64), nil
	default:
		return fmt.Sprint(jwtField), nil
	}
}

// GetJWTField gets the JWT from a context and returns the specified field
// using the DefaultTokenName
func GetJWTField(ctx context.Context, tokenField string, keyfunc jwt.Keyfunc) (string, error) {
	return GetJWTFieldWithTokenType(ctx, DefaultTokenType, tokenField, keyfunc)
}

// GetAccountID gets the JWT from a context and returns the AccountID field
func GetAccountID(ctx context.Context, keyfunc jwt.Keyfunc) (string, error) {
	for _, tenantField := range multiTenancyVariants {
		if val, err := GetJWTField(ctx, tenantField, keyfunc); err == nil {
			return val, nil
		}
	}
	return "", errMissingField
}

// GetCompartmentID gets the JWT from a context and returns the CompartmentID field
func GetCompartmentID(ctx context.Context, keyfunc jwt.Keyfunc) (string, error) {
	val, err := GetJWTField(ctx, MultiCompartmentField, keyfunc)
	if err == errMissingField {
		return "", nil
	}
	return val, err
}

// getToken parses the token into a jwt.Token type from the grpc metadata.
// WARNING: if keyfunc is nil, the token will get parsed but not verified
// because it has been checked previously in the stack. More information
// here: https://pkg.go.dev/github.com/golang-jwt/jwt/v4#Parser.ParseUnverified
func getToken(ctx context.Context, tokenField string, keyfunc jwt.Keyfunc) (jwt.Token, error) {
	if ctx == nil {
		return jwt.Token{}, errMissingToken
	}
	tokenStr, err := grpc_auth.AuthFromMD(ctx, tokenField)
	if err != nil {
		return jwt.Token{}, err
	}
	parser := jwt.Parser{}
	if keyfunc != nil {
		token, err := parser.Parse(tokenStr, keyfunc)
		if err != nil {
			return jwt.Token{}, err
		}
		return *token, nil
	}
	token, _, err := parser.ParseUnverified(tokenStr, jwt.MapClaims{})
	if err != nil {
		return jwt.Token{}, err
	}
	return *token, nil
}
