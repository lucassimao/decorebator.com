package integration

import (
	"fmt"
	"net/http"
	"testing"

	httphandlers "decorebator.com/internal/http"
	"decorebator.com/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRegistration(t *testing.T) {
	server := setup.NewTestServer(t)

	tests := []struct {
		name           string
		payload        map[string]interface{}
		expectedStatus int
		shouldHaveUser bool
	}{
		{
			name: "successful registration",
			payload: map[string]interface{}{
				"email":     "newuser@example.com",
				"password":  "password123",
				"firstName": "John",
				"lastName":  "Doe",
			},
			expectedStatus: http.StatusCreated,
			shouldHaveUser: true,
		},
		{
			name: "duplicate email registration",
			payload: map[string]interface{}{
				"email":     "duplicate@example.com", // Will be created first, then tested for duplicate
				"password":  "password123",
				"firstName": "Jane",
				"lastName":  "Smith",
			},
			expectedStatus: http.StatusCreated,
			shouldHaveUser: false,
		},
		{
			name: "missing required fields",
			payload: map[string]interface{}{
				"email": "incomplete@example.com",
				// Missing password, firstName, lastName
			},
			expectedStatus: http.StatusBadRequest,
			shouldHaveUser: false,
		},
		{
			name: "invalid email format",
			payload: map[string]interface{}{
				"email":     "invalid-email",
				"password":  "password123",
				"firstName": "John",
				"lastName":  "Doe",
			},
			expectedStatus: http.StatusBadRequest,
			shouldHaveUser: false,
		},
		{
			name: "weak password",
			payload: map[string]interface{}{
				"email":     "weakpass@example.com",
				"password":  "123", // Too short
				"firstName": "John",
				"lastName":  "Doe",
			},
			expectedStatus: http.StatusBadRequest,
			shouldHaveUser: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			// Special handling for duplicate email test
			if tt.name == "duplicate email registration" {
				// First create the user
				server.Expect.POST("/users").
					WithJSON(tt.payload).
					Expect().
					Status(http.StatusCreated)

				// Now test duplicate registration
				server.Expect.POST("/users").
					WithJSON(tt.payload).
					Expect().
					Status(tt.expectedStatus)

				// No further processing needed for duplicate test
				return
			}

			response := server.Expect.POST("/users").
				WithJSON(tt.payload).
				Expect().
				Status(tt.expectedStatus)

			if tt.shouldHaveUser {
				response.Header("Authorization").IsEmpty()
				response.JSON().Object().Value("message").String().NotEmpty()
			}
		})
	}
}

func TestUserLogin(t *testing.T) {
	server := setup.NewTestServer(t)

	// Create a test user first
	signupInput := setup.GenerateSignupInput()
	server.Expect.POST("/users").
		WithJSON(signupInput).
		Expect().
		Status(http.StatusCreated)
	server.VerifyTestSignup(t, signupInput.Email, signupInput.Password)

	tests := []struct {
		name           string
		credentials    interface{}
		expectedStatus int
		shouldHaveJWT  bool
	}{
		{
			name: "successful login",
			credentials: httphandlers.LoginInput{
				Email:    signupInput.Email,
				Password: signupInput.Password,
			},
			expectedStatus: http.StatusOK,
			shouldHaveJWT:  true,
		},
		{
			name: "wrong password",
			credentials: httphandlers.LoginInput{
				Email:    signupInput.Email,
				Password: "wrongpassword",
			},
			expectedStatus: http.StatusBadRequest, // API returns 400 for wrong password
			shouldHaveJWT:  false,
		},
		{
			name: "non-existent user",
			credentials: httphandlers.LoginInput{
				Email:    "nonexistent@example.com",
				Password: "password123",
			},
			expectedStatus: http.StatusBadRequest, // API returns 400 for non-existent user
			shouldHaveJWT:  false,
		},
		{
			name: "missing email",
			credentials: map[string]interface{}{
				"password": signupInput.Password,
			},
			expectedStatus: http.StatusBadRequest,
			shouldHaveJWT:  false,
		},
		{
			name: "missing password",
			credentials: map[string]interface{}{
				"email": signupInput.Email,
			},
			expectedStatus: http.StatusBadRequest,
			shouldHaveJWT:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			response := server.Expect.POST("/login").
				WithJSON(tt.credentials).
				Expect().
				Status(tt.expectedStatus)

			if tt.shouldHaveJWT {
				response.Header("Authorization").IsEmpty()
				response.Header("X-Refresh-Token").IsEmpty()
				response.Cookies().ContainsOnly("Authorization", "RefreshToken")
				cookie := response.Cookie("Authorization")
				cookie.Value().NotEmpty()
			}
		})
	}
}

func TestJWTAuthentication(t *testing.T) {
	server := setup.NewTestServer(t)

	// Create authenticated user
	token := server.WithTestUser(t)

	tests := []struct {
		name           string
		token          string
		expectedStatus int
	}{
		{
			name:           "valid token",
			token:          token,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing token",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid token",
			token:          "invalid.jwt.token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "malformed token",
			token:          "Bearer malformed-token",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, server.BaseURL+"/users", nil)
			require.NoError(t, err)
			if tt.token != "" {
				request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tt.token))
			}

			response, err := server.Server.Client().Do(request)
			require.NoError(t, err)
			defer response.Body.Close()
			assert.Equal(t, tt.expectedStatus, response.StatusCode)
		})
	}
}

func TestPasswordReset(t *testing.T) {
	server := setup.NewTestServer(t)

	// Create a test user
	signupInput := setup.GenerateSignupInput()
	server.Expect.POST("/users").
		WithJSON(signupInput).
		Expect().
		Status(http.StatusCreated)

	t.Run("request password reset", func(_ *testing.T) {
		server.Expect.POST("/password/send-reset-email").
			WithJSON(map[string]interface{}{
				"email": signupInput.Email,
			}).
			Expect().
			Status(http.StatusAccepted)
		// Endpoint returns empty body on success
	})

	t.Run("request reset for non-existent email", func(_ *testing.T) {
		server.Expect.POST("/password/send-reset-email").
			WithJSON(map[string]interface{}{
				"email": "nonexistent@example.com",
			}).
			Expect().
			Status(http.StatusAccepted)
	})

	t.Run("reset password with invalid token", func(t *testing.T) { //nolint:revive // t is used for testing
		// Test with invalid token format - should return 400 Bad Request
		resetToken := "invalid-token-123"

		server.Expect.PATCH("/password/reset").
			WithJSON(map[string]interface{}{
				"token":    resetToken,
				"password": "newpassword123",
			}).
			Expect().
			Status(http.StatusBadRequest)
	})
}

func TestUserProfile(t *testing.T) {
	server := setup.NewTestServer(t)

	token := server.WithTestUser(t)

	t.Run("get user profile", func(t *testing.T) { //nolint:revive // t is used for testing
		response := server.Expect.GET("/users").
			WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
			Expect().
			Status(http.StatusOK)

		json := response.JSON().Object()
		json.ContainsKey("id")
		json.ContainsKey("email")
		json.ContainsKey("firstName")
		json.ContainsKey("lastName")
		json.ContainsKey("subscriptionPlan")
		json.NotContainsKey("password")
	})

	t.Run("update user profile", func(t *testing.T) { //nolint:revive // t is used for testing
		updateData := map[string]interface{}{
			"firstName": "UpdatedFirst",
			"lastName":  "UpdatedLast",
			"country":   "US",
		}

		response := server.Expect.PATCH("/users").
			WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
			WithJSON(updateData).
			Expect().
			Status(http.StatusOK)

		json := response.JSON().Object()
		json.Value("firstName").IsEqual("UpdatedFirst")
		json.Value("lastName").IsEqual("UpdatedLast")
		json.Value("country").IsEqual("US")
	})

	t.Run("delete user account", func(t *testing.T) { //nolint:revive // t is used for testing
		server.Expect.DELETE("/users").
			WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
			Expect().
			Status(http.StatusNoContent) // API returns 204 No Content for delete

		// Verify user can no longer access profile
		server.Expect.GET("/users").
			WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
			Expect().
			Status(http.StatusUnauthorized)
	})
}

func TestUserLogout(t *testing.T) {
	server := setup.NewTestServer(t)

	token := server.WithTestUser(t)

	t.Run("successful logout", func(t *testing.T) { //nolint:revive // t is used for testing
		server.Expect.POST("/logout").
			WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
			Expect().
			Status(http.StatusOK)
	})

	t.Run("logout without token", func(t *testing.T) { //nolint:revive // t is used for testing
		server.Expect.POST("/logout").
			Expect().
			Status(http.StatusOK) // Logout endpoint allows access without token
	})
}

func TestSubscriptionAuthentication(t *testing.T) {
	server := setup.NewTestServer(t)

	freeToken := server.WithTestUser(t)
	premiumToken := server.WithPremiumUser(t)

	t.Run("free user subscription status", func(t *testing.T) { //nolint:revive // t is used for testing
		response := server.Expect.GET("/users").
			WithHeader("Authorization", fmt.Sprintf("Bearer %s", freeToken)).
			Expect().
			Status(http.StatusOK)

		json := response.JSON().Object()
		json.Value("subscriptionPlan").IsEqual("free")
	})

	t.Run("premium user subscription status", func(t *testing.T) { //nolint:revive // t is used for testing
		response := server.Expect.GET("/users").
			WithHeader("Authorization", fmt.Sprintf("Bearer %s", premiumToken)).
			Expect().
			Status(http.StatusOK)

		json := response.JSON().Object()
		json.Value("subscriptionPlan").IsEqual("monthly")
	})
}

func TestMultipleRequests(t *testing.T) {
	server := setup.NewTestServer(t)

	t.Run("multiple rapid registrations", func(t *testing.T) {
		// Test multiple rapid registrations (no rate limiting implemented currently)
		successCount := 0
		for i := 0; i < 5; i++ {
			testSignupInput := setup.GenerateSignupInput()
			response := server.Expect.POST("/users").
				WithJSON(testSignupInput).
				Expect()

			// All should succeed since no rate limiting is implemented
			if response.Raw().StatusCode == http.StatusCreated { //nolint:bodyclose // httpexpect handles body closing
				successCount++
			}
		}

		assert.Equal(t, 5, successCount, "All registrations should succeed without rate limiting")
	})

	t.Run("multiple login attempts with wrong password", func(t *testing.T) {
		// Generate a unique signup input for this subtest
		signupInput := setup.GenerateSignupInput()

		// Create user first
		server.Expect.POST("/users").
			WithJSON(signupInput).
			Expect().
			Status(http.StatusCreated)

		// Attempt multiple rapid login attempts with wrong password
		unauthorizedCount := 0
		for i := 0; i < 5; i++ {
			response := server.Expect.POST("/login").
				WithJSON(httphandlers.LoginInput{
					Email:    signupInput.Email,
					Password: "wrongpassword",
				}).
				Expect()

			// Should consistently get unauthorized (no rate limiting implemented)
			if response.Raw().StatusCode == http.StatusBadRequest { //nolint:bodyclose // httpexpect handles body closing
				unauthorizedCount++
			}
		}

		assert.Equal(t, 5, unauthorizedCount, "All wrong password attempts should return 400 without rate limiting")
	})
}
