package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"decorebator.com/internal/config"
	"decorebator.com/internal/model"
	"decorebator.com/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type failingLogoutSessions struct{}

func (failingLogoutSessions) Logout(context.Context, string) error {
	return errors.New("database unavailable")
}

func (failingLogoutSessions) Refresh(context.Context, string) (service.SessionCredentials, error) {
	return service.SessionCredentials{}, errors.New("not used")
}

func TestCORSUsesExplicitAllowlistAndStopsRejectedOrigins(t *testing.T) {
	gin.SetMode(gin.TestMode)
	policy := config.HTTPSecurityConfig{AllowedOrigins: []string{"https://app.example"}}
	runs := 0
	router := gin.New()
	router.Use(CORSMiddleware(policy))
	router.POST("/state", func(c *gin.Context) { runs++; c.Status(http.StatusNoContent) })

	allowed := httptest.NewRecorder()
	allowedRequest := httptest.NewRequest(http.MethodPost, "/state", nil)
	allowedRequest.Header.Set("Origin", "https://app.example")
	router.ServeHTTP(allowed, allowedRequest)
	assert.Equal(t, http.StatusNoContent, allowed.Code)
	assert.Equal(t, "https://app.example", allowed.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "Origin", allowed.Header().Get("Vary"))
	assert.NotContains(t, allowed.Header().Get("Access-Control-Allow-Headers"), "Cookie")
	assert.NotContains(t, allowed.Header().Get("Access-Control-Allow-Headers"), "Authorization")
	assert.NotContains(t, allowed.Header().Get("Access-Control-Expose-Headers"), "Authorization")
	assert.Contains(t, allowed.Header().Get("Access-Control-Expose-Headers"), "X-Next-Cursor")
	assert.Contains(t, allowed.Header().Get("Access-Control-Expose-Headers"), "X-Definitions-Continuation")

	rejected := httptest.NewRecorder()
	rejectedRequest := httptest.NewRequest(http.MethodPost, "/state", nil)
	rejectedRequest.Header.Set("Origin", "https://attacker.example")
	router.ServeHTTP(rejected, rejectedRequest)
	assert.Equal(t, http.StatusForbidden, rejected.Code)
	assert.Equal(t, 1, runs)
}

func TestCORSRejectsDuplicateOriginHeadersBeforeHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	runs := 0
	router := gin.New()
	router.Use(CORSMiddleware(config.HTTPSecurityConfig{AllowedOrigins: []string{"https://app.example"}}))
	router.POST("/", func(c *gin.Context) { runs++; c.Status(http.StatusNoContent) })
	request := httptest.NewRequest(http.MethodPost, "/", nil)
	request.Header.Add("Origin", "https://app.example")
	request.Header.Add("Origin", "https://attacker.example")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Zero(t, runs)
}

func TestCORSVariesResponsesWithoutOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CORSMiddleware(config.HTTPSecurityConfig{}))
	router.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	response := httptest.NewRecorder()

	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, http.StatusNoContent, response.Code)
	assert.Equal(t, "Origin", response.Header().Get("Vary"))
}

func TestSecurityHeadersCoverAPIResponsesAndHSTSIsProductionOnly(t *testing.T) {
	for _, test := range []struct {
		name       string
		production bool
		hsts       string
	}{
		{name: "development"},
		{name: "production", production: true, hsts: "max-age=31536000"},
	} {
		t.Run(test.name, func(t *testing.T) {
			router := gin.New()
			router.Use(SecurityHeaders(test.production))
			router.GET("/", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
			assert.Equal(t, "default-src 'none'; frame-ancestors 'none'; base-uri 'none'; form-action 'none'", response.Header().Get("Content-Security-Policy"))
			assert.Equal(t, "DENY", response.Header().Get("X-Frame-Options"))
			assert.Equal(t, "nosniff", response.Header().Get("X-Content-Type-Options"))
			assert.Equal(t, "no-referrer", response.Header().Get("Referrer-Policy"))
			assert.Equal(t, "no-store", response.Header().Get("Cache-Control"))
			assert.NotEmpty(t, response.Header().Get("Permissions-Policy"))
			assert.Equal(t, test.hsts, response.Header().Get("Strict-Transport-Security"))
		})
	}
}

func TestSecurityHeadersCoverRecoveredErrorResponses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(ErrorMiddleware(), SecurityHeaders(true))
	router.GET("/panic", func(*gin.Context) { panic("test panic") })
	response := httptest.NewRecorder()

	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/panic", nil))

	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Equal(t, "nosniff", response.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "max-age=31536000", response.Header().Get("Strict-Transport-Security"))
}

func TestFailedLogoutStillExpiresBrowserCookies(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &UserRoutes{
		authSessions: failingLogoutSessions{},
		httpSecurity: config.HTTPSecurityConfig{SecureCookies: true},
	}
	router := gin.New()
	router.POST("/logout", handler.Logout)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/logout", nil))

	assert.Equal(t, http.StatusServiceUnavailable, response.Code)
	cookies := response.Result().Cookies()
	require.Len(t, cookies, 2)
	assert.Equal(t, "__Host-Authorization", cookies[0].Name)
	assert.Equal(t, -1, cookies[0].MaxAge)
	assert.True(t, cookies[0].HttpOnly)
	assert.True(t, cookies[0].Secure)
	assert.Equal(t, "__Host-RefreshToken", cookies[1].Name)
	assert.Equal(t, -1, cookies[1].MaxAge)
	assert.True(t, cookies[1].HttpOnly)
	assert.True(t, cookies[1].Secure)
}

func TestStaticAuthenticationAbortsBeforeHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const token = "0123456789abcdef0123456789abcdef"
	for _, test := range []struct {
		name   string
		header string
		status int
		runs   int
	}{
		{name: "missing", status: http.StatusUnauthorized},
		{name: "wrong", header: strings.Repeat("x", len(token)), status: http.StatusUnauthorized},
		{name: "correct", header: token, status: http.StatusNoContent, runs: 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			runs := 0
			router := gin.New()
			router.Use(AuthenticateStatic(token))
			router.GET("/", func(c *gin.Context) { runs++; c.Status(http.StatusNoContent) })
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			request.Header.Set("Authorization", test.header)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			assert.Equal(t, test.status, response.Code)
			assert.Equal(t, test.runs, runs)
		})
	}
}

func TestSessionCookiesAreHostOnlyBoundedAndSymmetricallyCleared(t *testing.T) {
	gin.SetMode(gin.TestMode)
	policy := config.HTTPSecurityConfig{SecureCookies: true}

	assertCookie := func(value string, cleared bool) {
		t.Helper()
		response := httptest.NewRecorder()
		context, _ := gin.CreateTestContext(response)
		context.Request = httptest.NewRequest(http.MethodPost, "/", nil)
		writeAuthenticationCookie(context, policy, value)
		writeRefreshCookie(context, policy, value)
		cookies := response.Result().Cookies()
		require.Len(t, cookies, 2)
		assert.ElementsMatch(t, []string{
			"__Host-Authorization", "__Host-RefreshToken",
		}, []string{cookies[0].Name, cookies[1].Name})
		for _, cookie := range cookies {
			assert.Empty(t, cookie.Domain)
			assert.Equal(t, "/", cookie.Path)
			assert.True(t, cookie.Secure)
			assert.True(t, cookie.HttpOnly)
			assert.Equal(t, http.SameSiteStrictMode, cookie.SameSite)
			if cleared {
				assert.Less(t, cookie.MaxAge, 0)
			} else {
				if cookie.Name == authenticationCookieName(policy) {
					assert.Equal(t, int(config.AccessTokenDuration/time.Second), cookie.MaxAge)
				} else {
					assert.Equal(t, int(service.RefreshIdleDuration/time.Second), cookie.MaxAge)
				}
			}
		}
	}
	assertCookie("token", false)
	assertCookie("", true)
}

func TestDeployedRefreshIgnoresTossedGenericCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	response := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(response)
	context.Request = httptest.NewRequest(http.MethodPost, "/session/refresh", nil)
	context.Request.AddCookie(&http.Cookie{Name: "RefreshToken", Value: "tossed"})
	context.Request.AddCookie(&http.Cookie{Name: "__Host-RefreshToken", Value: "host-only"})

	assert.Equal(t, "host-only", readRefreshToken(
		context, config.HTTPSecurityConfig{SecureCookies: true},
	))
}

func TestSessionCookieWritesExpireLegacyParentDomainVariants(t *testing.T) {
	gin.SetMode(gin.TestMode)
	policy := config.HTTPSecurityConfig{
		SecureCookies: true, LegacyCookieDomain: "decorebator.com",
	}
	router := gin.New()
	router.Use(ExpireLegacyParentDomainCookies(policy))
	router.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))

	cookies := response.Result().Cookies()
	require.Len(t, cookies, 2)
	legacyNames := make([]string, 0, 2)
	for _, cookie := range cookies {
		assert.Equal(t, "decorebator.com", cookie.Domain)
		assert.Less(t, cookie.MaxAge, 0)
		assert.True(t, cookie.HttpOnly)
		assert.True(t, cookie.Secure)
		assert.Equal(t, http.SameSiteStrictMode, cookie.SameSite)
		legacyNames = append(legacyNames, cookie.Name)
	}
	assert.ElementsMatch(t, []string{"Authorization", "RefreshToken"}, legacyNames)
}

func TestBrowserOriginCannotRequestNativeRefreshCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	credentials := service.SessionCredentials{AccessToken: "access", RefreshToken: "refresh"}

	for _, test := range []struct {
		name              string
		origin            string
		wantRefreshHeader bool
	}{
		{name: "native transport", wantRefreshHeader: true},
		{name: "browser impersonation", origin: "https://app.example"},
	} {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			context, _ := gin.CreateTestContext(response)
			context.Request = httptest.NewRequest(http.MethodPost, "/login", nil)
			context.Request.Header.Set("X-Auth-Client", "native")
			if test.origin != "" {
				context.Request.Header.Set("Origin", test.origin)
			}

			writeSessionCredentials(context, config.HTTPSecurityConfig{}, credentials)

			assert.Equal(t, test.wantRefreshHeader, response.Header().Get("X-Refresh-Token") != "")
			assert.Equal(t, test.wantRefreshHeader, response.Header().Get("Authorization") != "")
			if !test.wantRefreshHeader {
				assert.NotEmpty(t, response.Header().Values("Set-Cookie"))
			}
		})
	}
}

func TestSubscriptionLimitInvalidUserTypeAborts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set("user", "wrong type") })
	router.Use(CheckSubscriptionLimits(nil, model.UserActionCreateWordlist))
	runs := 0
	router.GET("/", func(*gin.Context) { runs++ })
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Zero(t, runs)
}
