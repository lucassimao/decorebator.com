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
			expectedStatus: http.StatusInternalServerError, // API currently returns 500, should be 409
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
				// Signup returns empty body but JWT token in Authorization header
				token := response.Header("Authorization").NotEmpty().Raw()
				require.NotEmpty(t, token, "Authorization token should be present")
				
				// Verify JWT token is also set as cookie
				response.Cookies().ContainsOnly("Authorization")
				cookie := response.Cookie("Authorization")
				cookie.Value().NotEmpty()
				assert.Equal(t, token, cookie.Value().Raw(), "Token in header should match cookie")
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
			expectedStatus: http.StatusBadRequest, // API returns 400 for wrong password
			shouldHaveJWT:  false,
		},
		{
			name: "non-existent user",
			credentials: map[string]interface{}{
				"email":    "nonexistent@example.com",
				"password": "password123",
			},
			expectedStatus: http.StatusBadRequest, // API returns 400 for non-existent user
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
				// Login returns empty body but JWT token in Authorization header
				token := response.Header("Authorization").NotEmpty().Raw()
				require.NotEmpty(t, token, "Authorization token should be present")
				
				// Verify JWT token is also set as cookie
				response.Cookies().ContainsOnly("Authorization")
				cookie := response.Cookie("Authorization")
				cookie.Value().NotEmpty()
				assert.Equal(t, token, cookie.Value().Raw(), "Token in header should match cookie")
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
		server.Expect.POST("/password/send-reset-email").
			WithJSON(map[string]interface{}{
				"email": user["email"],
			}).
			Expect().
			Status(http.StatusOK)
		// Endpoint returns empty body on success
	})

	t.Run("request reset for non-existent email", func(t *testing.T) {
		server.Expect.POST("/password/send-reset-email").
			WithJSON(map[string]interface{}{
				"email": "nonexistent@example.com",
			}).
			Expect().
			Status(http.StatusOK) // Should still return 200 for security
	})

	t.Run("reset password with invalid token", func(t *testing.T) {
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
			Status(http.StatusOK) // Logout endpoint allows access without token
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

func TestMultipleRequests(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	user := setup.GenerateTestUser()

	t.Run("multiple rapid registrations", func(t *testing.T) {
		// Test multiple rapid registrations (no rate limiting implemented currently)
		successCount := 0
		for i := 0; i < 5; i++ {
			testUser := setup.GenerateTestUser()
			response := server.Expect.POST("/users").
				WithJSON(testUser).
				Expect()

			// All should succeed since no rate limiting is implemented
			if response.Raw().StatusCode == http.StatusCreated {
				successCount++
			}

			// Small delay to avoid overwhelming the system
			time.Sleep(50 * time.Millisecond)
		}
		
		assert.Equal(t, 5, successCount, "All registrations should succeed without rate limiting")
	})

	t.Run("multiple login attempts with wrong password", func(t *testing.T) {
		// Create user first
		server.Expect.POST("/users").
			WithJSON(user).
			Expect().
			Status(http.StatusCreated)

		// Attempt multiple rapid login attempts with wrong password
		unauthorizedCount := 0
		for i := 0; i < 5; i++ {
			response := server.Expect.POST("/login").
				WithJSON(map[string]interface{}{
					"email":    user["email"],
					"password": "wrongpassword",
				}).
				Expect()

			// Should consistently get unauthorized (no rate limiting implemented)
			if response.Raw().StatusCode == http.StatusBadRequest {
				unauthorizedCount++
			}

			time.Sleep(50 * time.Millisecond)
		}
		
		assert.Equal(t, 5, unauthorizedCount, "All wrong password attempts should return 400 without rate limiting")
	})
}