package unit

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"decorebator.com/internal/app"
	decorebatorhttp "decorebator.com/internal/http"
	"decorebator.com/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMobileIAPDisabledRoutesReturnStableValidEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	authenticate := func(c *gin.Context) {
		c.Set("userID", int64(7))
		c.Set("user", &model.User{ID: 7, SubscriptionPlan: model.PlanFree})
		c.Next()
	}
	decorebatorhttp.RegisterMobileIAPRoutes(router, &app.Context{StoreIAPEnabled: false}, nil, authenticate)
	request := httptest.NewRequest(http.MethodGet, "/subscription/iap/context?store=apple", nil)
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
