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

# Run unit tests with structured output
make test-report
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
make lint
if [ $? -ne 0 ]; then
    echo "❌ Linting failed"
    exit 1
fi

# Run security scan
make security-scan
if [ $? -ne 0 ]; then
    echo "⚠️  Security issues found"
    # Don't fail on security issues in pre-commit
fi

echo "✅ All pre-commit checks passed"
```

### 2. CI Pipeline Configuration

```yaml
# .github/workflows/test.yml (excerpt)
- name: Run unit tests with structured output
  run: |
    # Install gotestsum for better test output formatting
    go install gotest.tools/gotestsum@latest
    
    # Run tests with structured output and junit report
    gotestsum --junitfile unit-tests.xml --format testname -- \
      -v -race -coverprofile=unit.out -covermode=atomic ./internal/...
    
    # Display coverage summary
    go tool cover -func=unit.out
    
- name: Check coverage thresholds
  run: |
    # Unit test coverage (70% threshold)
    UNIT_COV=$(go tool cover -func=unit.out | grep total | awk '{print $3}' | sed 's/%//')
    echo "Unit coverage: ${UNIT_COV}%"
    if (( $(echo "$UNIT_COV < 70" | bc -l) )); then
      echo "❌ Unit coverage (${UNIT_COV}%) below threshold (70%)"
      exit 1
    fi
    
    # Integration test coverage (80% threshold) - if integration.out exists
    if [ -f integration.out ]; then
      INT_COV=$(go tool cover -func=integration.out | grep total | awk '{print $3}' | sed 's/%//')
      echo "Integration coverage: ${INT_COV}%"
      if (( $(echo "$INT_COV < 80" | bc -l) )); then
        echo "❌ Integration coverage (${INT_COV}%) below threshold (80%)"
        exit 1
      fi
    fi

- name: Upload coverage to Codecov
  uses: codecov/codecov-action@v4
  with:
    file: unit.out
    flags: unit
    name: unit-tests
    token: ${{ secrets.CODECOV_TOKEN }}
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

## 🧮 Analytics Testing Implementation

### Comprehensive Analytics Test Suite

The Decorebator API includes a comprehensive analytics testing suite with dedicated test files for each analytics endpoint. This implementation demonstrates advanced testing patterns for complex database calculations and business logic validation.

#### Analytics Test Coverage

- **8 dedicated test files** covering all analytics endpoints
- **Database query validation** with direct PostgreSQL testing
- **Complex calculation verification** for metrics, percentages, and aggregations
- **Edge case testing** including empty data, boundary conditions, and error scenarios

#### Key Files
- `tests/integration/analytics/` - Complete analytics test suite
- `api/docs/ANALYTICS_TESTING_IMPLEMENTATION.md` - Detailed implementation guide

#### Example: Analytics Calculation Testing

```go
func TestWordMasteryEndpoint_CalculationAccuracy(t *testing.T) {
    // Arrange: Create test data with known values
    testData := setupWordMasteryTestData(t, server.DB, ctx)
    
    // Act: Call analytics endpoint
    response := server.Expect.GET(fmt.Sprintf("/analytics/wordlists/%d/mastery", testData.WordlistID)).
        WithHeader("Authorization", token).
        Expect().Status(http.StatusOK)
    
    // Assert: Verify calculations match expected values
    stats := response.JSON().Object().Value("stats").Array()
    for i, stat := range stats.Iter() {
        expected := testData.ExpectedWords[i]
        statObj := stat.Object()
        
        statObj.Value("masteryLevel").Number().InDelta(expected.MasteryLevel, 0.01)
        statObj.Value("accuracy").Number().InDelta(expected.Accuracy, 0.01)
        statObj.Value("highestBox").ValueEqual("highestBox", expected.BoxLevel)
    }
}
```

#### Advanced Testing Patterns

1. **Controlled Data Seeding** - Creates test data with predetermined values for exact calculation verification
2. **Complex Query Testing** - Validates CTEs, window functions, and multi-table aggregations
3. **Metric-Specific Test Files** - Focused test suites for each analytics endpoint
4. **Shared Utilities** - Common helpers reduce duplication across test files

### Database Query Validation

Following CLAUDE.md requirements, all analytics tests validate database queries directly:

```go
// Insert known test data
_, err := db.Exec(ctx,
    `INSERT INTO word_mastery (user_id, word_id, mastery_level, total_attempts, correct_attempts, streak_count, max_streak) 
     VALUES ($1, $2, $3, $4, $5, $6, $7)`,
    userID, wordID, 0.75, 16, 12, 4, 6)

// Test endpoint calculation
response := server.Expect.GET(fmt.Sprintf("/analytics/wordlists/%d/mastery", wordlistID))

// Verify calculation accuracy
expectedAccuracy := float64(12) / float64(16) // 75%
statObj.Value("accuracy").Number().InDelta(expectedAccuracy, 0.01)
```

## 🔧 Modern Testing Infrastructure

### Updated Testing Commands (2025)

The project now includes modernized testing infrastructure with structured output and version-aligned tools:

#### Quick Testing Commands
```bash
# Fast unit tests with structured output
make test-report

# Comprehensive testing (all tests + coverage)
./scripts/run-tests.sh all

# Specific test types
./scripts/run-tests.sh unit        # Unit tests only
./scripts/run-tests.sh integration # Integration tests only
./scripts/run-tests.sh watch       # Watch mode for development

# Tool management
make setup                         # Install all tools with correct versions
make check-versions               # Verify tool versions match CI
make security-scan                # Run security scans (govulncheck)
```

#### Advanced Test Script Features
- **Version Alignment**: All tools match GitHub Actions workflow versions
- **Structured Output**: Uses `gotestsum` for better test reporting  
- **Security Scanning**: Integrated `govulncheck` for vulnerability detection
- **Clean Environment**: Automatic Docker cleanup between test runs
- **Coverage Reports**: HTML reports and threshold checking

#### Test Script Examples
```bash
# First-time setup
./scripts/run-tests.sh setup

# Development workflow
./scripts/run-tests.sh watch       # Auto-reload on file changes

# CI-style testing
./scripts/run-tests.sh report      # Full report with all metrics

# Cleanup after issues
./scripts/run-tests.sh clean       # Remove Docker containers/volumes
```

### Tool Versions (Matching CI)
- **Go**: 1.23
- **PostgreSQL**: 15  
- **Redis**: 7-alpine
- **gotestsum**: latest (structured test output)
- **govulncheck**: latest (security scanning)

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

### Project-Specific Resources
- `api/docs/ANALYTICS_TESTING_IMPLEMENTATION.md` - Comprehensive analytics testing guide
- `tests/integration/analytics/` - Analytics test suite implementation
- `tests/integration/setup/` - Test infrastructure and utilities

### Best Practices References
- [Google Go Testing Best Practices](https://google.github.io/styleguide/go/best-practices.html#testing)
- [Effective Go Testing](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)
- [Testing in Go](https://blog.golang.org/testing)