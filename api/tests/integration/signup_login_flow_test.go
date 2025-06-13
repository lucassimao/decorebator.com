package integration

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	httphandlers "decorebator.com/internal/http"
	"decorebator.com/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSignupLoginFlow tests the complete user registration and login flow
func TestSignupLoginFlow(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	// The login route is now handled by the test server setup

	// Test data
	testUser := map[string]interface{}{
		"email":     "testuser@example.com",
		"password":  "securepassword123",
		"firstName": "John",
		"lastName":  "Doe",
	}

	t.Run("complete signup and login flow", func(t *testing.T) {
		// Step 1: Sign up a new user
		t.Log("Step 1: Creating new user account")
		signupResponse := server.Expect.POST("/users").
			WithJSON(testUser).
			Expect().
			Status(http.StatusCreated)

		// Verify signup response contains JWT token in Authorization header
		authToken := signupResponse.Header("Authorization").NotEmpty().Raw()
		require.NotEmpty(t, authToken, "Authorization token should be present")

		// Verify JWT token is also set as cookie
		signupResponse.Cookies().ContainsOnly("Authorization")
		cookie := signupResponse.Cookie("Authorization")
		cookie.Value().NotEmpty()
		assert.Equal(t, authToken, cookie.Value().Raw(), "Token in header should match cookie")

		t.Log("✓ User successfully created and JWT token received")

		// Step 2: Login with the created user credentials
		t.Log("Step 2: Logging in with created user credentials")
		loginCredentials := map[string]interface{}{
			"email":    testUser["email"],
			"password": testUser["password"],
		}

		loginResponse := server.Expect.POST("/login").
			WithJSON(loginCredentials).
			Expect().
			Status(http.StatusOK)

		// Verify login response contains JWT token in Authorization header
		loginToken := loginResponse.Header("Authorization").NotEmpty().Raw()
		require.NotEmpty(t, loginToken, "Authorization token should be present after login")

		// Verify JWT token is also set as cookie
		loginResponse.Cookies().ContainsOnly("Authorization")
		loginCookie := loginResponse.Cookie("Authorization")
		loginCookie.Value().NotEmpty()
		assert.Equal(t, loginToken, loginCookie.Value().Raw(), "Login token in header should match cookie")

		// Verify both tokens are valid JWT tokens (they might be different due to different generation times)
		assert.Greater(t, len(authToken), 20, "Signup token should be a valid JWT")
		assert.Greater(t, len(loginToken), 20, "Login token should be a valid JWT")

		t.Logf("✓ User successfully logged in with token: %s", loginToken[:20]+"...")

		// Step 3: Verify the user exists in database with correct data
		t.Log("Step 3: Verifying user data in database")
		ctx := context.Background()
		var dbUserID int64
		var dbEmail, dbFirstName, dbLastName, dbPasswordHash string
		var dbCreatedAt time.Time

		err := server.DB.QueryRow(ctx,
			"SELECT id, email, first_name, last_name, password_hash, created_at FROM users WHERE email = $1",
			testUser["email"]).Scan(&dbUserID, &dbEmail, &dbFirstName, &dbLastName, &dbPasswordHash, &dbCreatedAt)

		require.NoError(t, err, "User should exist in database")

		// Verify database data matches
		assert.Greater(t, dbUserID, int64(0), "User ID should be positive")
		assert.Equal(t, testUser["email"], dbEmail, "Email should match")
		assert.Equal(t, testUser["firstName"], dbFirstName, "First name should match")
		assert.Equal(t, testUser["lastName"], dbLastName, "Last name should match")
		assert.NotEqual(t, testUser["password"], dbPasswordHash, "Password should be hashed, not stored as plain text")
		assert.Contains(t, dbPasswordHash, "$2a$", "Password should be bcrypt hashed")
		assert.WithinDuration(t, time.Now(), dbCreatedAt, 5*time.Second, "Created timestamp should be recent")

		t.Logf("✓ User data verified in database")

		t.Log("✅ Complete signup and login flow successful!")
	})

	t.Run("login with incorrect password", func(t *testing.T) {
		// First ensure user exists from previous test
		server.Expect.POST("/users").
			WithJSON(map[string]interface{}{
				"email":     "wrongpass@example.com",
				"password":  "correctpassword",
				"firstName": "Test",
				"lastName":  "User",
			}).
			Expect().
			Status(http.StatusCreated)

		// Try to login with wrong password
		wrongCredentials := map[string]interface{}{
			"email":    "wrongpass@example.com",
			"password": "wrongpassword",
		}

		server.Expect.POST("/login").
			WithJSON(wrongCredentials).
			Expect().
			Status(http.StatusBadRequest) // API returns 400 for wrong password

		t.Log("✓ Login correctly rejected for wrong password")
	})

	t.Run("login with non-existent user", func(t *testing.T) {
		nonExistentCredentials := map[string]interface{}{
			"email":    "nonexistent@example.com",
			"password": "anypassword",
		}

		server.Expect.POST("/login").
			WithJSON(nonExistentCredentials).
			Expect().
			Status(http.StatusBadRequest) // API returns 400 for non-existent user

		t.Log("✓ Login correctly rejected for non-existent user")
	})

	t.Run("signup with duplicate email", func(t *testing.T) {
		// Create first user
		firstUser := map[string]interface{}{
			"email":     "duplicate@example.com",
			"password":  "password123",
			"firstName": "First",
			"lastName":  "User",
		}

		server.Expect.POST("/users").
			WithJSON(firstUser).
			Expect().
			Status(http.StatusCreated)

		// Try to create second user with same email
		secondUser := map[string]interface{}{
			"email":     "duplicate@example.com", // Same email
			"password":  "differentpassword",
			"firstName": "Second",
			"lastName":  "User",
		}

		response := server.Expect.POST("/users").
			WithJSON(secondUser).
			Expect().
			Status(http.StatusInternalServerError) // API returns 500 for duplicate email

		errorJSON := response.JSON().Object()
		errorJSON.ContainsKey("error")
		errorJSON.Value("error").IsEqual("Email already exists.")

		t.Log("✓ Signup correctly rejected for duplicate email")
	})

	t.Run("signup with invalid data", func(t *testing.T) {
		invalidCases := []struct {
			name    string
			payload map[string]interface{}
			error   string
		}{
			{
				name: "missing email",
				payload: map[string]interface{}{
					"password":  "password123",
					"firstName": "John",
					"lastName":  "Doe",
				},
				error: "Missing required fields",
			},
			{
				name: "missing password",
				payload: map[string]interface{}{
					"email":     "test@example.com",
					"firstName": "John",
					"lastName":  "Doe",
				},
				error: "Missing required fields",
			},
			{
				name: "invalid email format",
				payload: map[string]interface{}{
					"email":     "invalid-email",
					"password":  "password123",
					"firstName": "John",
					"lastName":  "Doe",
				},
				error: "Invalid email format",
			},
			{
				name: "password too short",
				payload: map[string]interface{}{
					"email":     "short@example.com",
					"password":  "123", // Too short
					"firstName": "John",
					"lastName":  "Doe",
				},
				error: "Password too short",
			},
		}

		for _, tc := range invalidCases {
			t.Run(tc.name, func(t *testing.T) {
				response := server.Expect.POST("/users").
					WithJSON(tc.payload).
					Expect().
					Status(http.StatusBadRequest)

				// Now that content-type is fixed, we can properly check JSON responses
				errorJSON := response.JSON().Object()
				errorJSON.ContainsKey("validationErrors")

				t.Logf("✓ Signup correctly rejected for %s", tc.name)
			})
		}
	})

	t.Run("login with invalid data", func(t *testing.T) {
		invalidCases := []struct {
			name    string
			payload map[string]interface{}
		}{
			{
				name: "missing email",
				payload: map[string]interface{}{
					"password": "password123",
				},
			},
			{
				name: "missing password",
				payload: map[string]interface{}{
					"email": "test@example.com",
				},
			},
			{
				name:    "empty payload",
				payload: map[string]interface{}{},
			},
		}

		for _, tc := range invalidCases {
			t.Run(tc.name, func(t *testing.T) {
				server.Expect.POST("/login").
					WithJSON(tc.payload).
					Expect().
					Status(http.StatusBadRequest)

				// Login errors return empty body, just check status code

				t.Logf("✓ Login correctly rejected for %s", tc.name)
			})
		}
	})
}

// TestSignupLoginPerformance tests the performance of signup and login operations
func TestSignupLoginPerformance(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	// The login route is now handled by the test server setup

	t.Run("signup performance", func(t *testing.T) {
		start := time.Now()

		signupInput := setup.GenerateSignupInput()
		server.Expect.POST("/users").
			WithJSON(signupInput).
			Expect().
			Status(http.StatusCreated)

		duration := time.Since(start)

		// Signup should complete within 2 seconds (includes bcrypt hashing)
		assert.Less(t, duration, 2*time.Second, "Signup should be fast")
		t.Logf("✓ Signup completed in %v", duration)
	})

	t.Run("login performance", func(t *testing.T) {
		// Create user first
		signupInput := setup.GenerateSignupInput()
		server.Expect.POST("/users").
			WithJSON(signupInput).
			Expect().
			Status(http.StatusCreated)

		start := time.Now()

		credentials := httphandlers.LoginInput{
			Email:    signupInput.Email,
			Password: signupInput.Password,
		}

		server.Expect.POST("/login").
			WithJSON(credentials).
			Expect().
			Status(http.StatusOK)

		duration := time.Since(start)

		// Login should complete within 3 seconds (includes bcrypt verification)
		assert.Less(t, duration, 3*time.Second, "Login should be fast")
		t.Logf("✓ Login completed in %v", duration)
	})

	t.Run("concurrent signups", func(t *testing.T) {
		const numConcurrent = 10
		results := make(chan error, numConcurrent)

		start := time.Now()

		// Start concurrent signups
		for i := 0; i < numConcurrent; i++ {
			go func(index int) {
				signupInput := setup.GenerateSignupInput()
				signupInput.Email = fmt.Sprintf("concurrent%d@example.com", index)

				resp := server.Expect.POST("/users").
					WithJSON(signupInput).
					Expect()

				if resp.Raw().StatusCode != http.StatusCreated { //nolint:bodyclose // httpexpect handles body closing
					results <- fmt.Errorf("signup %d failed with status %d", index, resp.Raw().StatusCode) //nolint:bodyclose // httpexpect handles body closing
				} else {
					results <- nil
				}
			}(i)
		}

		// Collect results
		var errors []error
		for i := 0; i < numConcurrent; i++ {
			if err := <-results; err != nil {
				errors = append(errors, err)
			}
		}

		duration := time.Since(start)

		// All concurrent signups should succeed
		assert.Empty(t, errors, "All concurrent signups should succeed")

		// Should complete within 15 seconds for 10 concurrent signups (including bcrypt hashing)
		assert.Less(t, duration, 15*time.Second, "Concurrent signups should complete within reasonable time")

		t.Logf("✓ %d concurrent signups completed in %v", numConcurrent, duration)
	})
}
