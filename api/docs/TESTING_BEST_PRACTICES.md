# Testing Best Practices for Decorebator API

This document outlines best practices, patterns, and guidelines for writing effective tests in the Decorebator API backend.

## 🎯 Testing Philosophy

### Core Principles

1. **Test Behavior, Not Implementation** - Focus on what the code does, not how it does it
2. **Fast, Reliable, Isolated** - Tests should be quick, deterministic, and independent
3. **Clear and Maintainable** - Tests should be easy to understand and modify
4. **Comprehensive Coverage** - Cover happy paths, edge cases, and error scenarios

### Test Pyramid

```
    /\
   /  \    E2E Tests (Few)
  /____\   
 /      \   Integration Tests (Some)
/__________\ Unit Tests (Many)
```

- **Unit Tests (70%)**: Fast, isolated, test individual functions
- **Integration Tests (25%)**: Test API endpoints with real database
- **E2E Tests (5%)**: Full user journey tests (handled by mobile app)

## 📝 Writing Effective Tests

### 1. Test Naming Conventions

**Format**: `Test[Subject]_[Scenario]_[Expected]`

```go
// ✅ Good Examples
func TestUserRegistration_WithValidData_ReturnsCreatedUser(t *testing.T)
func TestCreateWordlist_WhenFreePlanLimitExceeded_Returns403(t *testing.T)
func TestQuizGeneration_WithNoWordsAvailable_ReturnsEmptyResponse(t *testing.T)

// ❌ Bad Examples
func TestUser(t *testing.T)
func TestCreateWordlist(t *testing.T)
func TestQuiz(t *testing.T)
```

### 2. Test Structure (AAA Pattern)

```go
func TestCreateWordlist_WithValidData_ReturnsCreatedWordlist(t *testing.T) {
    // Arrange
    server := setup.NewTestServer(t)
    defer server.Cleanup()
    
    token := server.WithTestUser(t)
    payload := map[string]interface{}{
        "name":     "Travel Essentials",
        "language": "en",
    }
    
    // Act
    response := server.Expect.POST("/wordlists").
        WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
        WithJSON(payload).
        Expect()
    
    // Assert
    response.Status(http.StatusCreated)
    json := response.JSON().Object()
    json.Value("name").Equal("Travel Essentials")
    json.Value("language").Equal("en")
    json.ContainsKey("id")
}
```

### 3. Table-Driven Tests

Use table-driven tests for multiple scenarios:

```go
func TestUserValidation(t *testing.T) {
    tests := []struct {
        name           string
        payload        map[string]interface{}
        expectedStatus int
        expectedError  string
    }{
        {
            name: "valid user data",
            payload: map[string]interface{}{
                "email":     "test@example.com",
                "password":  "password123",
                "firstName": "John",
                "lastName":  "Doe",
            },
            expectedStatus: http.StatusCreated,
        },
        {
            name: "missing email",
            payload: map[string]interface{}{
                "password":  "password123",
                "firstName": "John",
                "lastName":  "Doe",
            },
            expectedStatus: http.StatusBadRequest,
            expectedError:  "email is required",
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
            expectedError:  "invalid email format",
        },
    }

    server := setup.NewTestServer(t)
    defer server.Cleanup()

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            response := server.Expect.POST("/users").
                WithJSON(tt.payload).
                Expect().
                Status(tt.expectedStatus)

            if tt.expectedError != "" {
                response.JSON().Object().Value("error").String().Contains(tt.expectedError)
            }
        })
    }
}
```

### 4. Test Helpers and Utilities

Create reusable helpers for common operations:

```go
// tests/helpers/auth.go
func CreateAuthenticatedUser(t *testing.T, server *setup.TestServer) (string, int64) {
    user := setup.GenerateTestUser()
    
    // Register user
    createResp := server.Expect.POST("/users").
        WithJSON(user).
        Expect().
        Status(http.StatusCreated)
    
    userID := createResp.JSON().Object().Value("id").Number().Raw()
    
    // Login to get token
    loginResp := server.Expect.POST("/login").
        WithJSON(map[string]interface{}{
            "email":    user["email"],
            "password": user["password"],
        }).
        Expect().
        Status(http.StatusOK)
    
    token := loginResp.JSON().Object().Value("token").String().Raw()
    
    return token, int64(userID)
}

// Usage
func TestSomeFeature(t *testing.T) {
    server := setup.NewTestServer(t)
    defer server.Cleanup()
    
    token, userID := helpers.CreateAuthenticatedUser(t, server)
    
    // Use token and userID in test...
}
```

## 🔧 Database Testing Patterns

### 1. Transaction-Based Isolation

```go
func TestUserService_UpdateProfile(t *testing.T) {
    server := setup.NewTestServer(t)
    defer server.Cleanup()
    
    // Each test runs in its own transaction (auto-rollback)
    setup.WithTransaction(t, server.DB, func(tx pgx.Tx) {
        // Create test user within transaction
        userID := createTestUser(tx, "test@example.com")
        
        // Test user update
        service := &UserService{db: tx}
        err := service.UpdateProfile(userID, profileData)
        
        assert.NoError(t, err)
        
        // Verify update
        user, err := service.GetUser(userID)
        assert.NoError(t, err)
        assert.Equal(t, "Updated Name", user.FirstName)
        
        // Transaction automatically rolls back after test
    })
}
```

### 2. Test Data Fixtures

```go
// Load predefined test data
func TestWordlistAnalytics(t *testing.T) {
    server := setup.NewTestServer(t)
    defer server.Cleanup()
    
    // Load test fixtures
    server.SeedTestData(t, "users", "wordlists", "words", "quiz_performance")
    
    token := server.WithTestUser(t)
    
    response := server.Expect.GET("/analytics/wordlists/1/progress").
        WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
        Expect().
        Status(http.StatusOK)
    
    // Verify analytics data
    json := response.JSON().Object()
    json.Value("totalWords").Number().Gt(0)
    json.Value("accuracy").Number().Between(0, 100)
}
```

### 3. Database State Assertions

```go
func TestCreateWordlist_DatabaseState(t *testing.T) {
    server := setup.NewTestServer(t)
    defer server.Cleanup()
    
    token := server.WithTestUser(t)
    
    // Create wordlist via API
    response := server.Expect.POST("/wordlists").
        WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
        WithJSON(wordlistData).
        Expect().
        Status(http.StatusCreated)
    
    wordlistID := response.JSON().Object().Value("id").Number().Raw()
    
    // Verify database state directly
    var count int
    err := server.DB.QueryRow(context.Background(),
        "SELECT COUNT(*) FROM wordlists WHERE id = $1", wordlistID).Scan(&count)
    
    assert.NoError(t, err)
    assert.Equal(t, 1, count)
}
```

## 🎭 Mocking External Services

### 1. HTTP Mock Server Pattern

```go
// tests/mocks/openai_mock.go
type OpenAIMock struct {
    Server     *httptest.Server
    Requests   []OpenAIRequest
    Responses  map[string]interface{}
}

func NewOpenAIMock() *OpenAIMock {
    mock := &OpenAIMock{
        Requests:  make([]OpenAIRequest, 0),
        Responses: make(map[string]interface{}),
    }
    
    mock.Server = httptest.NewServer(http.HandlerFunc(mock.handler))
    return mock
}

func (m *OpenAIMock) handler(w http.ResponseWriter, r *http.Request) {
    // Record request
    m.Requests = append(m.Requests, parseRequest(r))
    
    // Return mock response
    switch r.URL.Path {
    case "/v1/chat/completions":
        m.respondWithChatCompletion(w)
    case "/v1/images/generations":
        m.respondWithImageGeneration(w)
    default:
        http.NotFound(w, r)
    }
}

// Usage in tests
func TestDefinitionGeneration_WithOpenAI(t *testing.T) {
    server := setup.NewTestServer(t)
    defer server.Cleanup()
    
    // Setup OpenAI mock
    openaiMock := mocks.NewOpenAIMock()
    defer openaiMock.Close()
    
    openaiMock.MockChatCompletion(`{
        "meaning": "A greeting used when meeting someone",
        "partOfSpeech": "interjection",
        "examples": ["Hello, how are you?"]
    }`)
    
    // Configure service to use mock
    os.Setenv("OPENAI_BASE_URL", openaiMock.Server.URL)
    
    // Test definition generation
    token := server.WithTestUser(t)
    
    response := server.Expect.POST("/words/1/definitions/generate").
        WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
        Expect().
        Status(http.StatusOK)
    
    // Verify mock was called
    assert.Len(t, openaiMock.Requests, 1)
    assert.Contains(t, openaiMock.Requests[0].Prompt, "hello")
}
```

### 2. Interface-Based Mocking

```go
// internal/interfaces/openai.go
type OpenAIService interface {
    GenerateDefinition(ctx context.Context, word, language string) (*Definition, error)
    GenerateImage(ctx context.Context, prompt string) (string, error)
}

// tests/mocks/openai_service_mock.go
type MockOpenAIService struct {
    DefinitionResponses map[string]*Definition
    ImageResponses      map[string]string
    CallLog            []string
}

func (m *MockOpenAIService) GenerateDefinition(ctx context.Context, word, language string) (*Definition, error) {
    m.CallLog = append(m.CallLog, fmt.Sprintf("GenerateDefinition:%s:%s", word, language))
    
    if def, exists := m.DefinitionResponses[word]; exists {
        return def, nil
    }
    
    return &Definition{
        Meaning:      fmt.Sprintf("Mock definition for %s", word),
        PartOfSpeech: "noun",
        Examples:     []string{fmt.Sprintf("Example with %s", word)},
    }, nil
}

// Usage
func TestDefinitionService_GenerateDefinition(t *testing.T) {
    mockOpenAI := &MockOpenAIService{
        DefinitionResponses: map[string]*Definition{
            "hello": {
                Meaning:      "A greeting",
                PartOfSpeech: "interjection",
                Examples:     []string{"Hello, world!"},
            },
        },
    }
    
    service := &DefinitionService{
        openai: mockOpenAI,
        db:     testDB,
    }
    
    definition, err := service.GenerateDefinition(ctx, "hello", "en")
    
    assert.NoError(t, err)
    assert.Equal(t, "A greeting", definition.Meaning)
    assert.Contains(t, mockOpenAI.CallLog, "GenerateDefinition:hello:en")
}
```

## 🔍 Testing Error Scenarios

### 1. Network Errors

```go
func TestDefinitionGeneration_OpenAITimeout(t *testing.T) {
    server := setup.NewTestServer(t)
    defer server.Cleanup()
    
    // Setup mock that times out
    timeoutMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        time.Sleep(30 * time.Second) // Longer than client timeout
    }))
    defer timeoutMock.Close()
    
    os.Setenv("OPENAI_BASE_URL", timeoutMock.URL)
    os.Setenv("OPENAI_TIMEOUT", "1s")
    
    token := server.WithTestUser(t)
    
    response := server.Expect.POST("/words/1/definitions/generate").
        WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
        Expect().
        Status(http.StatusServiceUnavailable)
    
    response.JSON().Object().Value("error").String().Contains("timeout")
}
```

### 2. Database Errors

```go
func TestUserService_DatabaseFailure(t *testing.T) {
    // Close database connection to simulate failure
    server := setup.NewTestServer(t)
    server.DB.Close()
    
    token := server.WithTestUser(t) // This should fail gracefully
    
    response := server.Expect.GET("/users").
        WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
        Expect().
        Status(http.StatusInternalServerError)
    
    response.JSON().Object().Value("error").String().Contains("database")
}
```

### 3. Rate Limiting

```go
func TestRateLimiting_ErrorReports(t *testing.T) {
    server := setup.NewTestServer(t)
    defer server.Cleanup()
    
    token := server.WithTestUser(t)
    
    // Make requests up to the limit
    for i := 0; i < 3; i++ {
        server.Expect.POST("/errorReports").
            WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
            WithJSON(errorReportData).
            Expect().
            Status(http.StatusOK)
    }
    
    // Next request should be rate limited
    response := server.Expect.POST("/errorReports").
        WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
        WithJSON(errorReportData).
        Expect().
        Status(http.StatusTooManyRequests)
    
    json := response.JSON().Object()
    json.ContainsKey("retryAfter")
    json.Value("windowType").Equal("hourly")
}
```

## 🚀 Performance Testing

### 1. Benchmark Tests

```go
func BenchmarkAnalyticsService_CalculateMastery(b *testing.B) {
    service := &AnalyticsService{db: testDB}
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := service.CalculateMastery(context.Background(), 1, 1)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkQuizGeneration_ComplexQuery(b *testing.B) {
    server := setup.NewTestServer(b)
    defer server.Cleanup()
    
    token := server.WithTestUser(b)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        server.Expect.POST("/wordlists/1/quizzes").
            WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
            Expect().
            Status(http.StatusOK)
    }
}
```

### 2. Load Testing

```go
func TestConcurrentQuizGeneration(t *testing.T) {
    server := setup.NewTestServer(t)
    defer server.Cleanup()
    
    token := server.WithTestUser(t)
    
    const numGoroutines = 10
    const requestsPerGoroutine = 5
    
    var wg sync.WaitGroup
    errors := make(chan error, numGoroutines*requestsPerGoroutine)
    
    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            
            for j := 0; j < requestsPerGoroutine; j++ {
                response := server.Expect.POST("/wordlists/1/quizzes").
                    WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
                    Expect()
                
                if response.Raw().StatusCode != http.StatusOK {
                    errors <- fmt.Errorf("request failed with status %d", response.Raw().StatusCode)
                }
            }
        }()
    }
    
    wg.Wait()
    close(errors)
    
    // Check for errors
    for err := range errors {
        t.Error(err)
    }
}
```

## 📊 Coverage and Quality

### 1. Coverage Targets

```go
//go:build coverage
// +build coverage

// This file is only included when running coverage tests
func TestCoverageTargets(t *testing.T) {
    // Critical paths should have near 100% coverage
    criticalPaths := []string{
        "internal/service/user.go",
        "internal/service/auth.go", 
        "internal/service/subscription.go",
    }
    
    for _, path := range criticalPaths {
        coverage := getCoverageForFile(path)
        assert.True(t, coverage >= 95.0, 
            "Critical path %s has insufficient coverage: %.1f%% (required: 95%%)", 
            path, coverage)
    }
}
```

### 2. Quality Gates

```go
func TestCodeQuality(t *testing.T) {
    t.Run("no TODO comments in production code", func(t *testing.T) {
        files := findGoFiles("internal/")
        
        for _, file := range files {
            content := readFile(file)
            assert.NotContains(t, content, "TODO", 
                "File %s contains TODO comments", file)
        }
    })
    
    t.Run("all errors are properly handled", func(t *testing.T) {
        // Static analysis to ensure error handling
        // This could use tools like errcheck
    })
}
```

## 🔄 Continuous Integration

### 1. Pre-commit Hooks

```bash
#!/bin/sh
# .git/hooks/pre-commit

echo "Running pre-commit checks..."

# Run unit tests
make test-unit
if [ $? -ne 0 ]; then
    echo "❌ Unit tests failed"
    exit 1
fi

# Check test coverage
make coverage-threshold
if [ $? -ne 0 ]; then
    echo "❌ Coverage threshold not met"
    exit 1
fi

# Run linter
golangci-lint run
if [ $? -ne 0 ]; then
    echo "❌ Linting failed"
    exit 1
fi

echo "✅ All pre-commit checks passed"
```

### 2. CI Pipeline Configuration

```yaml
# .github/workflows/test.yml (excerpt)
- name: Run tests with coverage
  run: |
    go test -v -race -coverprofile=coverage.out ./...
    
- name: Check coverage threshold
  run: |
    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
    echo "Coverage: ${COVERAGE}%"
    if (( $(echo "$COVERAGE < 80" | bc -l) )); then
      echo "Coverage ${COVERAGE}% is below threshold 80%"
      exit 1
    fi

- name: Upload coverage to Codecov
  uses: codecov/codecov-action@v3
  with:
    file: ./coverage.out
```

## 🐛 Debugging Tests

### 1. Test Debugging Techniques

```go
func TestDebugExample(t *testing.T) {
    // Enable debug logging
    if testing.Verbose() {
        log.SetLevel(log.DebugLevel)
    }
    
    server := setup.NewTestServer(t)
    defer server.Cleanup()
    
    // Print request/response for debugging
    response := server.Expect.POST("/wordlists").
        WithJSON(payload).
        Expect().
        Status(http.StatusCreated)
    
    if testing.Verbose() {
        t.Logf("Response: %s", response.Body().Raw())
    }
}

// Run with: go test -v -run TestDebugExample
```

### 2. Test Data Inspection

```go
func TestWithDataInspection(t *testing.T) {
    server := setup.NewTestServer(t)
    defer server.Cleanup()
    
    // Create test data
    token := server.WithTestUser(t)
    
    // Inspect database state
    if testing.Verbose() {
        var count int
        server.DB.QueryRow(context.Background(), 
            "SELECT COUNT(*) FROM users").Scan(&count)
        t.Logf("Users in database: %d", count)
    }
    
    // ... rest of test
}
```

## 📚 Additional Resources

### Go Testing Resources
- [Go Testing Package](https://golang.org/pkg/testing/)
- [Advanced Testing Patterns](https://golang.org/doc/tutorial/add-a-test)
- [Table-Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)

### Testing Tools
- [testify](https://github.com/stretchr/testify) - Assertion library
- [httpexpect](https://github.com/gavv/httpexpect) - HTTP API testing
- [gomock](https://github.com/golang/mock) - Mock generation
- [golangci-lint](https://golangci-lint.run/) - Linting

### Best Practices References
- [Google Go Testing Best Practices](https://google.github.io/styleguide/go/best-practices.html#testing)
- [Effective Go Testing](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)
- [Testing in Go](https://blog.golang.org/testing)