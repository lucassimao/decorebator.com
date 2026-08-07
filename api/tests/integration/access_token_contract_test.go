package integration

import (
	"fmt"
	"os"
	"testing"
	"time"

	"decorebator.com/tests/integration/setup"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccessTokenUsesSubjectOnlyAndCurrentServerAuthorizationState(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	token := server.WithPremiumUser(t)
	claims := setup.DecodeJWT(t, token)
	assert.NotContains(t, claims, "email")
	assert.NotContains(t, claims, "subscriptionPlan")
	assert.Equal(t, "access", claims["tokenType"])
	assert.Equal(t, "test", claims["environment"])
	assert.Equal(t, "decorebator-api", claims["iss"])
	assert.Equal(t, []any{"decorebator-clients"}, claims["aud"])

	server.Expect.GET("/users").WithHeader("Authorization", token).Expect().Status(200).
		JSON().Object().Value("subscriptionPlan").String().IsEqual("monthly")
}

func TestLegacyYearLongClaimTokenForcesLogin(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	signup := setup.GenerateSignupInput()
	server.Expect.POST("/users").WithJSON(signup).Expect().Status(201)
	var userID int64
	require.NoError(t, server.DB.QueryRow(t.Context(), `SELECT id FROM users WHERE email=$1`, signup.Email).Scan(&userID))
	now := time.Now().UTC()
	legacy, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": "Decorebator", "sub": fmt.Sprint(userID), "iat": now.Unix(),
		"exp": now.Add(365 * 24 * time.Hour).Unix(), "email": signup.Email,
		"environment": "test", "subscriptionPlan": "free",
	}).SignedString([]byte(os.Getenv("JWT_KEY")))
	require.NoError(t, err)

	server.Expect.GET("/users").WithHeader("Authorization", legacy).Expect().Status(401)
}
