package auth

import (
	"context"
	"errors"
	"fmt"

	jwt "github.com/dgrijalva/jwt-go"
)

const (

	// TODO: Field is tentatively called "TenantID" but will probably need to be
	// changed. We don't know what the JWT will look like, so we're giving it our
	// best guess for the time being.
	TENANT_ID_FIElD = "TenantID"
)

var (
	errMissingTenantID = errors.New(
		fmt.Sprintf("unable to extract %s from token", TENANT_ID_FIElD),
	)
	errInvalidAssertion = errors.New("unable to assert value as jwt.MapClaims")
)

func GetTenantID(ctx context.Context, keyfunc jwt.Keyfunc) (string, error) {
	token, err := getToken(ctx, keyfunc)
	if err != nil {
		return "", err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errInvalidAssertion
	}
	tenantID, ok := claims[TENANT_ID_FIElD]
	if !ok {
		return "", errMissingTenantID
	}
	return fmt.Sprint(tenantID), nil
}
