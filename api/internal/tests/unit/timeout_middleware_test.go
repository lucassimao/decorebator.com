package unit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	httpPkg "decorebator.com/internal/http"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test the TimeoutMiddleware function in isolation
func TestTimeoutMiddleware_NormalRequest(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	// Create middleware
	middlewareFunc := httpPkg.TimeoutMiddleware(100 * time.Millisecond)

	// Create a simple handler to test with
	handler := func(c *gin.Context) {
		// Verify context has timeout
		deadline, hasDeadline := c.Request.Context().Deadline()
		assert.True(t, hasDeadline, "Context should have a deadline")
		assert.True(t, deadline.After(time.Now()), "Deadline should be in the future")

		c.JSON(http.StatusOK, gin.H{"message": "success"})
	}

	// Create router and apply middleware
	router := gin.New()
	router.Use(middlewareFunc)
	router.GET("/test", handler)

	// Test
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
}

func TestTimeoutMiddleware_RequestTimeout(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	// Create middleware with short timeout
	middlewareFunc := httpPkg.TimeoutMiddleware(50 * time.Millisecond)

	// Create a slow handler that takes longer than timeout
	handler := func(c *gin.Context) {
		// Simulate slow operation
		select {
		case <-time.After(100 * time.Millisecond):
			c.JSON(http.StatusOK, gin.H{"message": "should not reach here"})
		case <-c.Request.Context().Done():
			// Context canceled, but don't respond here - let middleware handle it
			return
		}
	}

	router := gin.New()
	router.Use(middlewareFunc)
	router.GET("/slow", handler)

	// Test
	req := httptest.NewRequest("GET", "/slow", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusRequestTimeout, w.Code)
	assert.Contains(t, w.Body.String(), "Request timeout - please try again")
}

func TestTimeoutMiddleware_ContextPropagation(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	timeoutDuration := 100 * time.Millisecond
	middlewareFunc := httpPkg.TimeoutMiddleware(timeoutDuration)

	var contextReceived context.Context
	handler := func(c *gin.Context) {
		contextReceived = c.Request.Context()
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	}

	router := gin.New()
	router.Use(middlewareFunc)
	router.GET("/test", handler)

	// Test
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	require.NotNil(t, contextReceived)
	deadline, hasDeadline := contextReceived.Deadline()
	assert.True(t, hasDeadline, "Context should have deadline")

	// Verify deadline is approximately correct (within reasonable tolerance)
	expectedDeadline := time.Now().Add(timeoutDuration)
	timeDiff := deadline.Sub(expectedDeadline)
	assert.True(t, timeDiff < 50*time.Millisecond && timeDiff > -50*time.Millisecond,
		"Deadline should be approximately %v from now, got %v (diff: %v)",
		timeoutDuration, time.Until(deadline), timeDiff)
}

func TestTimeoutMiddleware_NoDoubleResponse(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	middlewareFunc := httpPkg.TimeoutMiddleware(50 * time.Millisecond)

	// Handler that responds quickly but then tries to do more work
	handler := func(c *gin.Context) {
		// Respond immediately
		c.JSON(http.StatusOK, gin.H{"message": "first response"})

		// Then simulate more work that might timeout
		time.Sleep(100 * time.Millisecond)
	}

	router := gin.New()
	router.Use(middlewareFunc)
	router.GET("/test", handler)

	// Test
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert - should get the handler's response, not timeout response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "first response")
	assert.NotContains(t, w.Body.String(), "timeout")
}

func TestTimeoutMiddleware_ContextCancellation(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	middlewareFunc := httpPkg.TimeoutMiddleware(50 * time.Millisecond)

	var contextError error
	handler := func(c *gin.Context) {
		// Wait for context cancellation
		<-c.Request.Context().Done()
		contextError = c.Request.Context().Err()
		// Don't respond - let middleware handle timeout
	}

	router := gin.New()
	router.Use(middlewareFunc)
	router.GET("/test", handler)

	// Test
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusRequestTimeout, w.Code)
	assert.Equal(t, context.DeadlineExceeded, contextError)
}

func TestTimeoutMiddleware_DifferentTimeouts(t *testing.T) {
	tests := []struct {
		name            string
		timeoutDuration time.Duration
		handlerDelay    time.Duration
		expectedStatus  int
	}{
		{
			name:            "Short timeout, fast handler",
			timeoutDuration: 50 * time.Millisecond,
			handlerDelay:    10 * time.Millisecond,
			expectedStatus:  http.StatusOK,
		},
		{
			name:            "Short timeout, slow handler",
			timeoutDuration: 50 * time.Millisecond,
			handlerDelay:    100 * time.Millisecond,
			expectedStatus:  http.StatusRequestTimeout,
		},
		{
			name:            "Long timeout, slow handler",
			timeoutDuration: 200 * time.Millisecond,
			handlerDelay:    100 * time.Millisecond,
			expectedStatus:  http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			gin.SetMode(gin.TestMode)
			middlewareFunc := httpPkg.TimeoutMiddleware(tt.timeoutDuration)

			handler := func(c *gin.Context) {
				select {
				case <-time.After(tt.handlerDelay):
					c.JSON(http.StatusOK, gin.H{"message": "completed"})
				case <-c.Request.Context().Done():
					return
				}
			}

			router := gin.New()
			router.Use(middlewareFunc)
			router.GET("/test", handler)

			// Test
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code, "Test case: %s", tt.name)
		})
	}
}

func TestTimeoutMiddleware_ZeroTimeout(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	middlewareFunc := httpPkg.TimeoutMiddleware(0) // Zero timeout should timeout immediately

	handler := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "should timeout"})
	}

	router := gin.New()
	router.Use(middlewareFunc)
	router.GET("/test", handler)

	// Test
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert - zero timeout should result in immediate timeout
	assert.Equal(t, http.StatusRequestTimeout, w.Code)
}

func TestTimeoutMiddleware_ConcurrentRequests(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	middlewareFunc := httpPkg.TimeoutMiddleware(50 * time.Millisecond)

	fastHandler := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "fast"})
	}

	slowHandler := func(c *gin.Context) {
		time.Sleep(100 * time.Millisecond)
		c.JSON(http.StatusOK, gin.H{"message": "slow"})
	}

	router := gin.New()
	router.Use(middlewareFunc)
	router.GET("/fast", fastHandler)
	router.GET("/slow", slowHandler)

	// Test multiple requests concurrently
	req1 := httptest.NewRequest("GET", "/fast", nil)
	w1 := httptest.NewRecorder()

	req2 := httptest.NewRequest("GET", "/slow", nil)
	w2 := httptest.NewRecorder()

	// Process requests
	router.ServeHTTP(w1, req1)
	router.ServeHTTP(w2, req2)

	// Assert
	assert.Equal(t, http.StatusOK, w1.Code)
	assert.Contains(t, w1.Body.String(), "fast")

	assert.Equal(t, http.StatusRequestTimeout, w2.Code)
	assert.Contains(t, w2.Body.String(), "Request timeout")
}
