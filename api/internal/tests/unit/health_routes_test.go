package unit

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	decorebatorhttp "decorebator.com/internal/http"
	"github.com/gin-gonic/gin"
)

type healthCheckerFunc func(context.Context) error

func (f healthCheckerFunc) Ping(ctx context.Context) error {
	return f(ctx)
}

func TestHealthzIsPublicAndDependencyFree(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	decorebatorhttp.RegisterHealthRoutes(router, healthCheckerFunc(func(context.Context) error {
		t.Fatal("liveness must not inspect the database")
		return nil
	}))

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if response.Body.String() != `{"status":"ok"}` {
		t.Fatalf("body = %q", response.Body.String())
	}
}

func TestReadyzPingsDatabaseWithBoundedContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	decorebatorhttp.RegisterHealthRoutes(router, healthCheckerFunc(func(ctx context.Context) error {
		deadline, ok := ctx.Deadline()
		if !ok {
			t.Fatal("readiness database ping has no deadline")
		}
		remaining := time.Until(deadline)
		if remaining <= 0 || remaining > 2*time.Second {
			t.Fatalf("readiness deadline remaining = %s, want (0, 2s]", remaining)
		}
		return nil
	}))

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if response.Body.String() != `{"status":"ready"}` {
		t.Fatalf("body = %q", response.Body.String())
	}
}

func TestReadyzReturnsFixedNonSecretFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	decorebatorhttp.RegisterHealthRoutes(router, healthCheckerFunc(func(context.Context) error {
		return errors.New("postgres://admin:secret@private-host/database")
	}))

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
	if response.Body.String() != `{"status":"not_ready"}` {
		t.Fatalf("body = %q", response.Body.String())
	}
	if strings.Contains(response.Body.String(), "secret") || strings.Contains(response.Body.String(), "private-host") {
		t.Fatalf("readiness response leaked database error: %q", response.Body.String())
	}
}

func TestReadyzHonorsRequestCancellation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	decorebatorhttp.RegisterHealthRoutes(router, healthCheckerFunc(func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	}))

	requestContext, cancel := context.WithCancel(context.Background())
	cancel()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/readyz", nil).WithContext(requestContext)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
	if response.Body.String() != `{"status":"not_ready"}` {
		t.Fatalf("body = %q", response.Body.String())
	}
}

func TestReadyzTimesOutSlowDatabase(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	deadlineObserved := make(chan error, 1)
	decorebatorhttp.RegisterHealthRoutes(router, healthCheckerFunc(func(ctx context.Context) error {
		<-ctx.Done()
		deadlineObserved <- ctx.Err()
		return ctx.Err()
	}))

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	started := time.Now()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
	if elapsed := time.Since(started); elapsed > 2*time.Second {
		t.Fatalf("readiness timeout took %s, want at most 2s", elapsed)
	}
	if err := <-deadlineObserved; !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("database ping context error = %v, want deadline exceeded", err)
	}
}
