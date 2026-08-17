package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"decorebator.com/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func TestRateLimitErrorReportsFailsClosedWhenQuotaStoreIsUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	pool, err := pgxpool.New(context.Background(), "postgres://localhost:1/unavailable?connect_timeout=1")
	require.NoError(t, err)
	defer pool.Close()

	called := false
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user", &model.User{ID: 42})
		c.Next()
	})
	router.GET("/errorReports", RateLimitErrorReports(pool), func(c *gin.Context) {
		called = true
		c.Status(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodGet, "/errorReports", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	require.False(t, called, "the protected handler must not run after a quota-store failure")
	require.Equal(t, http.StatusServiceUnavailable, recorder.Code)
	require.Equal(t, "1", recorder.Header().Get("Retry-After"))
	require.JSONEq(t, `{"error":"Error reporting temporarily unavailable","retryAfter":1}`, recorder.Body.String())
}
