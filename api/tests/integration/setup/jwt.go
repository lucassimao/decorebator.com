package setup

import (
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

// DecodeJWT parses a JWT token string without validation and returns the claims.
// It's a test helper and will fail the test if the token is malformed.
func DecodeJWT(t *testing.T, tokenString string) jwt.MapClaims {
	// The actual token is usually prefixed with "Bearer "
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	parser := new(jwt.Parser)
	token, _, err := parser.ParseUnverified(tokenString, jwt.MapClaims{})
	require.NoError(t, err, "Failed to parse JWT token unverified")

	claims, ok := token.Claims.(jwt.MapClaims)
	require.True(t, ok, "Failed to cast claims to jwt.MapClaims")

	return claims
}
