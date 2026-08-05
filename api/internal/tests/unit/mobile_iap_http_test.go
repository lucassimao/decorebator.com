package unit

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"decorebator.com/internal/app"
	decorebatorhttp "decorebator.com/internal/http"
	"decorebator.com/internal/model"
	"decorebator.com/internal/service"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMobileIAPDisabledRoutesReturnStableValidEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("JWT_KEY", "mobile-iap-test-key")
	router := gin.New()
	decorebatorhttp.RegisterMobileIAPRoutes(router, &app.Context{StoreIAPEnabled: false}, nil)
	request := httptest.NewRequest(http.MethodGet, "/subscription/iap/context?store=apple", nil)
	request.Header.Set("Authorization", "Bearer "+mobileIAPTestJWT(t, "mobile-iap-test-key"))
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusServiceUnavailable, response.Code)
	var body model.MobileIAPResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	require.NoError(t, body.Validate())
	assert.Empty(t, body.Products)
	assert.Nil(t, body.PurchaseContext)
	require.NotNil(t, body.Error)
	assert.Equal(t, "iap.error.unavailable", body.Error.MessageKey)
	assert.NotContains(t, response.Body.String(), "RevenueCat")
	assert.NotContains(t, response.Body.String(), "Stripe")
}

func mobileIAPTestJWT(t *testing.T, key string) string {
	t.Helper()
	claims := service.Claims{
		Email: "test@example.invalid", SubscriptionPlan: model.PlanFree,
		StandardClaims: jwt.StandardClaims{
			Subject: "7", ExpiresAt: time.Now().Add(time.Hour).Unix(), Issuer: "Decorebator",
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(key))
	require.NoError(t, err)
	return token
}
