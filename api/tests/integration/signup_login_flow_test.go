package integration

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"decorebator.com/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSignupLoginFlow tests the complete user registration and login flow
func TestSignupLoginFlow(t *testing.T) {
	server := setup.NewTestServer(t)

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

		signupResponse.Header("Authorization").IsEmpty()
		signupResponse.Cookies().IsEmpty()
		signupResponse.JSON().Object().Value("message").String().NotEmpty()
		server.Expect.POST("/login").WithJSON(map[string]interface{}{
			"email": testUser["email"], "password": testUser["password"],
		}).Expect().Status(http.StatusBadRequest)
		server.VerifyTestSignup(t, testUser["email"].(string), testUser["password"].(string))

		t.Log("✓ User created pending mailbox verification")

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

		// Browser login is cookie-only and must not expose credentials in response headers.
		loginResponse.Header("Authorization").IsEmpty()
		loginResponse.Header("X-Refresh-Token").IsEmpty()
		loginResponse.Cookies().ContainsOnly("Authorization", "RefreshToken")
		loginCookie := loginResponse.Cookie("Authorization")
		loginCookie.Value().NotEmpty()
		loginToken := loginCookie.Value().Raw()

		// Verify the post-verification token is a valid JWT.
		assert.Greater(t, len(loginToken), 20, "Login token should be a valid JWT")

		t.Log("✓ User successfully logged in with cookie-only browser credentials")

		// Step 3: Verify the user exists in database with correct data
		t.Log("Step 3: Verifying user data in database")
		ctx := context.Background()
		var dbUserID int64
		var dbEmail, dbFirstName, dbLastName, dbPasswordHash string
		var dbCountry *string
		var dbCreatedAt time.Time

		err := server.DB.QueryRow(ctx,
			"SELECT id, email, first_name, last_name, password_hash, country, created_at FROM users WHERE email = $1",
			testUser["email"]).Scan(&dbUserID, &dbEmail, &dbFirstName, &dbLastName, &dbPasswordHash, &dbCountry, &dbCreatedAt)

		require.NoError(t, err, "User should exist in database")

		// Verify database data matches
		assert.Greater(t, dbUserID, int64(0), "User ID should be positive")
		assert.Equal(t, testUser["email"], dbEmail, "Email should match")
		assert.Equal(t, testUser["firstName"], dbFirstName, "First name should match")
		assert.Equal(t, testUser["lastName"], dbLastName, "Last name should match")
		assert.NotEqual(t, testUser["password"], dbPasswordHash, "Password should be hashed, not stored as plain text")
		assert.Contains(t, dbPasswordHash, "$2a$", "Password should be bcrypt hashed")
		assert.WithinDuration(t, time.Now(), dbCreatedAt, 5*time.Second, "Created timestamp should be recent")
		assert.Nil(t, dbCountry, "Country should be nil when not provided")

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
			Status(http.StatusCreated)

		response.Header("Authorization").IsEmpty()
		response.Cookies().IsEmpty()
		response.JSON().Object().Value("message").String().NotEmpty()

		t.Log("✓ Duplicate signup preserves the generic public contract")
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

	t.Run("signup with country code", func(t *testing.T) {
		testCases := []struct {
			name          string
			country       string
			expectedValue *string
		}{
			{
				name:          "signup with US country code",
				country:       "US",
				expectedValue: func() *string { s := "US"; return &s }(),
			},
			{
				name:          "signup with DE country code",
				country:       "DE",
				expectedValue: func() *string { s := "DE"; return &s }(),
			},
			{
				name:          "signup with JP country code",
				country:       "JP",
				expectedValue: func() *string { s := "JP"; return &s }(),
			},
			{
				name:          "signup with empty country",
				country:       "",
				expectedValue: nil,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				testUser := map[string]interface{}{
					"email":     fmt.Sprintf("country-test-%s@example.com", strings.ToLower(tc.country)),
					"password":  "securepassword123",
					"firstName": "John",
					"lastName":  "Doe",
					"country":   tc.country,
				}

				// Step 1: Sign up user with country
				server.Expect.POST("/users").
					WithJSON(testUser).
					Expect().
					Status(http.StatusCreated)

				// Step 2: Verify country is stored correctly in database
				ctx := context.Background()
				var dbUserID int64
				var dbEmail, dbFirstName, dbLastName, dbPasswordHash string
				var dbCountry *string
				var dbCreatedAt time.Time

				err := server.DB.QueryRow(ctx,
					"SELECT id, email, first_name, last_name, password_hash, country, created_at FROM users WHERE email = $1",
					testUser["email"]).Scan(&dbUserID, &dbEmail, &dbFirstName, &dbLastName, &dbPasswordHash, &dbCountry, &dbCreatedAt)

				require.NoError(t, err, "User should exist in database")

				// Verify country matches expected value
				if tc.expectedValue == nil {
					assert.Nil(t, dbCountry, "Country should be nil")
				} else {
					require.NotNil(t, dbCountry, "Country should not be nil")
					assert.Equal(t, *tc.expectedValue, *dbCountry, "Country should match")
				}

				t.Logf("✓ Country correctly stored for %s", tc.name)
			})
		}
	})

	t.Run("signup backward compatibility", func(t *testing.T) {
		// Test signup without country field in request
		testUser := map[string]interface{}{
			"email":     "backward-compat@example.com",
			"password":  "securepassword123",
			"firstName": "John",
			"lastName":  "Doe",
			// No country field at all
		}

		// Step 1: Sign up user without country field
		server.Expect.POST("/users").
			WithJSON(testUser).
			Expect().
			Status(http.StatusCreated)

		// Step 2: Verify country is nil in database
		ctx := context.Background()
		var dbUserID int64
		var dbEmail, dbFirstName, dbLastName, dbPasswordHash string
		var dbCountry *string
		var dbCreatedAt time.Time

		err := server.DB.QueryRow(ctx,
			"SELECT id, email, first_name, last_name, password_hash, country, created_at FROM users WHERE email = $1",
			testUser["email"]).Scan(&dbUserID, &dbEmail, &dbFirstName, &dbLastName, &dbPasswordHash, &dbCountry, &dbCreatedAt)

		require.NoError(t, err, "User should exist in database")
		assert.Nil(t, dbCountry, "Country should be nil when not provided in request")

		t.Log("✓ Backward compatibility maintained - signup works without country field")
	})

	t.Run("signup validation errors", func(t *testing.T) {
		validationCases := []struct {
			name    string
			payload map[string]interface{}
			status  int
		}{
			{
				name: "missing firstName",
				payload: map[string]interface{}{
					"email":    "missing-firstname@example.com",
					"password": "password123",
					"lastName": "Doe",
					"country":  "US",
				},
				status: http.StatusBadRequest,
			},
			{
				name: "missing lastName",
				payload: map[string]interface{}{
					"email":     "missing-lastname@example.com",
					"password":  "password123",
					"firstName": "John",
					"country":   "US",
				},
				status: http.StatusBadRequest,
			},
			{
				name: "empty firstName",
				payload: map[string]interface{}{
					"email":     "empty-firstname@example.com",
					"password":  "password123",
					"firstName": "",
					"lastName":  "Doe",
					"country":   "US",
				},
				status: http.StatusBadRequest,
			},
		}

		for _, tc := range validationCases {
			t.Run(tc.name, func(t *testing.T) {
				server.Expect.POST("/users").
					WithJSON(tc.payload).
					Expect().
					Status(tc.status)

				t.Logf("✓ Validation correctly rejected %s", tc.name)
			})
		}
	})
}

// TestSignupWithLanguageParameter tests user registration with preferred language
func TestSignupWithLanguageParameter(t *testing.T) {
	server := setup.NewTestServer(t)

	testCases := []struct {
		name              string
		preferredLanguage string
		expectedLanguage  *string
	}{
		{
			name:              "signup with English language",
			preferredLanguage: "en",
			expectedLanguage:  strPtr("en"),
		},
		{
			name:              "signup with German language",
			preferredLanguage: "de",
			expectedLanguage:  strPtr("de"),
		},
		{
			name:              "signup with Portuguese Brazil language",
			preferredLanguage: "pt_BR",
			expectedLanguage:  strPtr("pt_BR"),
		},
		{
			name:              "signup without language parameter",
			preferredLanguage: "",
			expectedLanguage:  nil, // Should be NULL in database
		},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create unique test user for each case
			testUser := map[string]interface{}{
				"email":     fmt.Sprintf("langtest%d@example.com", i),
				"password":  "securepassword123",
				"firstName": "Language",
				"lastName":  "Tester",
				"country":   "US",
			}

			// Add preferred language only if not empty
			if tc.preferredLanguage != "" {
				testUser["preferredLanguage"] = tc.preferredLanguage
			}

			// Step 1: Sign up user with language parameter
			t.Logf("Step 1: Creating user with preferredLanguage=%s", tc.preferredLanguage)
			server.Expect.POST("/users").
				WithJSON(testUser).
				Expect().
				Status(http.StatusCreated)

			server.VerifyTestSignup(t, testUser["email"].(string), testUser["password"].(string))
			authToken := server.LoginTestUser(testUser["email"].(string), testUser["password"].(string))

			t.Log("✓ User successfully created with language parameter")

			// Step 2: Verify language is stored correctly using GetProfile endpoint
			profileResp := server.Expect.GET("/users").
				WithHeader("Authorization", authToken).
				Expect().
				Status(http.StatusOK)

			// Parse the profile response
			profileData := profileResp.JSON().Object()

			// Verify user data matches
			profileData.Value("id").Number().Gt(0)
			profileData.Value("email").String().IsEqual(testUser["email"].(string))
			profileData.Value("firstName").String().IsEqual(testUser["firstName"].(string))
			profileData.Value("lastName").String().IsEqual(testUser["lastName"].(string))

			// Verify preferred language matches expected value
			if tc.expectedLanguage == nil {
				profileData.NotContainsKey("preferredLanguage")
			} else {
				profileData.Value("preferredLanguage").String().IsEqual(*tc.expectedLanguage)
			}

			t.Logf("✓ Language data verified via GetProfile API: %s", tc.preferredLanguage)
		})
	}
}

// Helper function to create string pointer
func strPtr(s string) *string {
	return &s
}
