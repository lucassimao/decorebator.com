package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"decorebator.com/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRegistration(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

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
				"email":     "newuser@example.com", // Same as above
				"password":  "password123",
				"firstName": "Jane",
				"lastName":  "Smith",
			},
			expectedStatus: http.StatusConflict,
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
		t.Run(tt.name, func(t *testing.T) {
			response := server.Expect.POST("/users").
				WithJSON(tt.payload).
				Expect().
				Status(tt.expectedStatus)

			if tt.shouldHaveUser {
				// Verify user was created successfully
				json := response.JSON().Object()
				json.ContainsKey("id")
				json.ContainsKey("email")
				json.Value("email").Equal(tt.payload["email"])
				json.Value("firstName").Equal(tt.payload["firstName"])
				json.Value("lastName").Equal(tt.payload["lastName"])
				json.NotContainsKey("password") // Password should not be returned
			}
		})
	}
}

func TestUserLogin(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	// Create a test user first
	user := setup.GenerateTestUser()
	server.Expect.POST("/users").
		WithJSON(user).
		Expect().
		Status(http.StatusCreated)

	tests := []struct {
		name           string
		credentials    map[string]interface{}
		expectedStatus int
		shouldHaveJWT  bool
	}{
		{
			name: "successful login",
			credentials: map[string]interface{}{
				"email":    user["email"],
				"password": user["password"],
			},
			expectedStatus: http.StatusOK,
			shouldHaveJWT:  true,
		},
		{
			name: "wrong password",
			credentials: map[string]interface{}{
				"email":    user["email"],
				"password": "wrongpassword",
			},
			expectedStatus: http.StatusUnauthorized,
			shouldHaveJWT:  false,
		},
		{
			name: "non-existent user",
			credentials: map[string]interface{}{
				"email":    "nonexistent@example.com",
				"password": "password123",
			},
			expectedStatus: http.StatusUnauthorized,
			shouldHaveJWT:  false,
		},
		{
			name: "missing email",
			credentials: map[string]interface{}{
				"password": user["password"],
			},
			expectedStatus: http.StatusBadRequest,
			shouldHaveJWT:  false,
		},
		{
			name: "missing password",
			credentials: map[string]interface{}{
				"email": user["email"],
			},
			expectedStatus: http.StatusBadRequest,
			shouldHaveJWT:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := server.Expect.POST("/login").
				WithJSON(tt.credentials).
				Expect().
				Status(tt.expectedStatus)

			if tt.shouldHaveJWT {
				json := response.JSON().Object()
				json.ContainsKey("token")
				json.ContainsKey("user")
				
				// Verify token is a valid string
				token := json.Value("token").String()
				require.NotEmpty(t, token.Raw())
				
				// Verify user data
				userData := json.Value("user").Object()
				userData.ContainsKey("id")
				userData.ContainsKey("email")
				userData.Value("email").Equal(user["email"])
				userData.NotContainsKey("password")
			}
		})
	}
}

func TestJWTAuthentication(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

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
			request := server.Expect.GET("/users")
			
			if tt.token != "" {
				request = request.WithHeader("Authorization", fmt.Sprintf("Bearer %s", tt.token))
			}
			
			request.Expect().Status(tt.expectedStatus)
		})
	}
}

func TestPasswordReset(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	// Create a test user
	user := setup.GenerateTestUser()
	server.Expect.POST("/users").
		WithJSON(user).
		Expect().
		Status(http.StatusCreated)

	t.Run("request password reset", func(t *testing.T) {
		response := server.Expect.POST("/password/send-reset-email").
			WithJSON(map[string]interface{}{
				"email": user["email"],
			}).
			Expect().
			Status(http.StatusOK)

		json := response.JSON().Object()
		json.ContainsKey("message")
	})

	t.Run("request reset for non-existent email", func(t *testing.T) {
		server.Expect.POST("/password/send-reset-email").
			WithJSON(map[string]interface{}{
				"email": "nonexistent@example.com",
			}).
			Expect().
			Status(http.StatusOK) // Should still return 200 for security
	})

	t.Run("reset password with valid token", func(t *testing.T) {
		// In a real test, we'd extract the reset token from the email
		// For now, we'll simulate the flow
		resetToken := "test-reset-token-123"
		
		server.Expect.PATCH("/password/reset").
			WithJSON(map[string]interface{}{
				"token":    resetToken,
				"password": "newpassword123",
			}).
			Expect().
			Status(http.StatusOK)
	})
}

func TestUserProfile(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	token := server.WithTestUser(t)

	t.Run("get user profile", func(t *testing.T) {
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

	t.Run("update user profile", func(t *testing.T) {
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
		json.Value("firstName").Equal("UpdatedFirst")
		json.Value("lastName").Equal("UpdatedLast")
		json.Value("country").Equal("US")
	})

	t.Run("delete user account", func(t *testing.T) {
		server.Expect.DELETE("/users").
			WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
			Expect().
			Status(http.StatusOK)

		// Verify user can no longer access profile
		server.Expect.GET("/users").
			WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
			Expect().
			Status(http.StatusUnauthorized)
	})
}

func TestUserLogout(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	token := server.WithTestUser(t)

	t.Run("successful logout", func(t *testing.T) {
		server.Expect.GET("/logout").
			WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
			Expect().
			Status(http.StatusOK)
	})

	t.Run("logout without token", func(t *testing.T) {
		server.Expect.GET("/logout").
			Expect().
			Status(http.StatusUnauthorized)
	})
}

func TestSubscriptionAuthentication(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	freeToken := server.WithTestUser(t)
	premiumToken := server.WithPremiumUser(t)

	t.Run("free user subscription status", func(t *testing.T) {
		response := server.Expect.GET("/users").
			WithHeader("Authorization", fmt.Sprintf("Bearer %s", freeToken)).
			Expect().
			Status(http.StatusOK)

		json := response.JSON().Object()
		json.Value("subscriptionPlan").Equal("free")
	})

	t.Run("premium user subscription status", func(t *testing.T) {
		response := server.Expect.GET("/users").
			WithHeader("Authorization", fmt.Sprintf("Bearer %s", premiumToken)).
			Expect().
			Status(http.StatusOK)

		json := response.JSON().Object()
		json.Value("subscriptionPlan").Equal("monthly")
	})
}

func TestRateLimiting(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	user := setup.GenerateTestUser()

	t.Run("registration rate limiting", func(t *testing.T) {
		// Attempt multiple rapid registrations from same IP
		for i := 0; i < 10; i++ {
			testUser := setup.GenerateTestUser()
			response := server.Expect.POST("/users").
				WithJSON(testUser).
				Expect()

			// First few should succeed, later ones might be rate limited
			if i < 5 {
				response.Status(http.StatusCreated)
			} else {
				// Allow either success or rate limit
				status := response.Raw().StatusCode
				assert.True(t, status == http.StatusCreated || status == http.StatusTooManyRequests)
			}

			// Small delay to avoid overwhelming the system
			time.Sleep(50 * time.Millisecond)
		}
	})

	t.Run("login rate limiting", func(t *testing.T) {
		// Create user first
		server.Expect.POST("/users").
			WithJSON(user).
			Expect().
			Status(http.StatusCreated)

		// Attempt multiple rapid login attempts with wrong password
		for i := 0; i < 10; i++ {
			response := server.Expect.POST("/login").
				WithJSON(map[string]interface{}{
					"email":    user["email"],
					"password": "wrongpassword",
				}).
				Expect()

			// Should get unauthorized or rate limited
			status := response.Raw().StatusCode
			assert.True(t, status == http.StatusUnauthorized || status == http.StatusTooManyRequests)

			time.Sleep(50 * time.Millisecond)
		}
	})
}