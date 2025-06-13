# Decorebator API Testing Framework

This directory contains comprehensive testing infrastructure for the Decorebator API, including unit tests, integration tests, and test utilities.

## Overview

The testing framework follows Go testing best practices and provides:
- **Unit Tests**: Fast, isolated tests for individual components
- **Integration Tests**: End-to-end tests with real API routes and database integration
- **Test Utilities**: Shared helpers, mocks, and fixtures for consistent testing
- **Performance Testing**: Benchmarks and load testing capabilities
- **Coverage Reporting**: Comprehensive coverage analysis and threshold enforcement

## Quick Start

### Prerequisites

- Go 1.22+
- Docker and Docker Compose
- PostgreSQL client tools (for manual debugging)

### Running Tests

```bash
# Run all tests (recommended for CI/CD)
./scripts/run-tests.sh all

# Run only unit tests (fast, for development)
./scripts/run-tests.sh unit

# Run only integration tests
./scripts/run-tests.sh integration

# Run tests in watch mode (auto-reload on file changes)
./scripts/run-tests.sh watch

# Check coverage thresholds
./scripts/run-tests.sh coverage

# First-time setup
./scripts/run-tests.sh setup
```

### Manual Test Execution

```bash
# Copy test environment file (required)
cp .env.test.example .env.test

# Start test services
docker-compose -f docker-compose.test.yml up -d

# Source environment and run tests
cd tests/integration
set -a && source ../../.env.test && set +a
go test -v -run TestSignupLoginFlow

# Run specific test packages
go test -v ./internal/service/...
go test -v ./tests/integration/...
```

## Test Structure

### Directory Organization

```
tests/
├── README.md                 # This file
├── integration/              # Integration tests
│   ├── auth_test.go          # Authentication flow tests
│   ├── signup_login_flow_test.go # Complete user registration and login
│   ├── helpers/              # Integration test helpers
│   ├── mocks/                # Mock servers for external APIs
│   ├── setup/                # Test environment setup
│   │   ├── database.go       # Database test utilities
│   │   ├── fixtures.go       # Test data fixtures
│   │   └── test_server.go    # HTTP test server setup (uses real API routes)
│   └── testdata/             # Static test data files
└── testutils/                # Shared test utilities (if needed)
```

### Unit Tests

Unit tests are co-located with source code in each package:

```
internal/
├── service/
│   ├── user.go
│   ├── user_test.go          # Unit tests for user service
│   ├── word.go
│   └── word_test.go          # Unit tests for word service
├── repository/
│   ├── user.go
│   └── user_test.go          # Unit tests for user repository
└── http/
    ├── user.go
    └── user_test.go          # Unit tests for user HTTP handlers
```

### Integration Tests

Integration tests are located in the `tests/integration/` directory and test complete workflows using **real API routes** from `internal/http/setup.go`:

- **Authentication**: User registration, login, JWT handling, session management
- **API Endpoints**: Full HTTP request/response cycles with real database
- **Database Operations**: CRUD operations, migrations, data integrity
- **External Services**: Mocked integrations with OpenAI, Stripe, SendGrid
- **Performance**: Response times, concurrent operations, resource usage

**Important**: Integration tests now use the actual API routes instead of mock routes, ensuring tests match production behavior exactly.

## Environment Configuration

### Environment Variables Setup

Tests require external environment configuration (no hardcoded values in test files):

1. **Copy the environment template**:
   ```bash
   cp .env.test.example .env.test
   ```

2. **Environment is sourced externally**:
   ```bash
   # For manual testing
   set -a && source .env.test && set +a
   go test -v ./tests/integration/...
   
   # For script usage
   ./scripts/run-tests.sh integration
   ```

### Test vs Development Environments

**Development Environment** (`.env`):
- Database: `localhost:5432/decorebator`
- Redis: `localhost:6379`
- MinIO: `localhost:9000`

**Test Environment** (`.env.test`):
- Database: `localhost:5433/decorebator_test`
- Redis: `localhost:6380`
- MinIO: `localhost:9001`

### Key Environment Variables

```bash
# Database (Test)
POSTGRES_USER=test_user
POSTGRES_PASSWORD=test_pass
POSTGRES_DB=decorebator_test
POSTGRES_HOST=localhost
POSTGRES_PORT=5433
DATABASE_URL=postgres://test_user:test_pass@localhost:5433/decorebator_test?sslmode=disable

# External Service Mocks
OPENAI_API_KEY=test-openai-key-mock
STRIPE_API_KEY=test-stripe-key-mock
STRIPE_WEBHOOK_SECRET=test-stripe-webhook-secret
STRIPE_SUCCESS_URL=http://localhost:8080/subscription/success
STRIPE_CANCEL_URL=http://localhost:8080/subscription/cancel
SENDGRID_API_KEY=test-sendgrid-key-mock
RESET_PASSWORD_PRIVATE_KEY=test-reset-password-key-32-chars

# Test Configuration
TEST_MODE=true
JWT_SECRET=test-jwt-secret-key-for-testing-only
ENV=test
DISABLE_WORKERS=true
DISABLE_EMAIL_SENDING=true
```

### Test Database

Integration tests use a dedicated PostgreSQL database that:
- Runs on port 5433 to avoid conflicts with development
- Automatically applies migrations using Go migration runner
- Cleans data between test runs for isolation
- Uses transactions for atomic test operations

## Architecture Changes

### Environment Variable Migration

**Previous (❌ Deprecated)**:
- Used `common.Env` global variable with `init()` functions
- Environment variables hardcoded in test files
- Mock routes instead of real API routes

**Current (✅ Modern)**:
- Direct `os.Getenv()` calls throughout codebase
- External environment configuration via `.env.test`
- Real API routes from `internal/http/setup.go`
- Removed `internal/common/env.go` completely

### Migration Runner

**Previous**: Used `migrate` CLI tool
**Current**: Uses Go-based migration runner (`go run ./cmd/migrate/main.go`)

This ensures consistent environment variable handling between tests and migration execution.

## Testing Patterns

### AAA Pattern

All tests follow the Arrange-Act-Assert pattern:

```go
func TestUserService_CreateUser_WithValidData_ReturnsCreatedUser(t *testing.T) {
    // Arrange
    service := setupUserService(t)
    userData := &model.User{
        Email:     "test@example.com",
        FirstName: "John",
        LastName:  "Doe",
    }
    
    // Act
    result, err := service.CreateUser(userData)
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, userData.Email, result.Email)
}
```

### Test Naming Convention

Format: `Test[Subject]_[Scenario]_[Expected]`

Examples:
- `TestUserService_CreateUser_WithValidData_ReturnsCreatedUser`
- `TestWordlistRepository_GetByUserID_WhenUserHasNoWordlists_ReturnsEmptySlice`
- `TestLoginHandler_WithInvalidCredentials_Returns401`

### Integration Test Setup

Integration tests use real API routes and external environment:

```go
func TestSignupLoginFlow(t *testing.T) {
    // Environment variables must be set externally
    // Test server uses real API routes from internal/http/setup.go
    server := setup.NewTestServer(t)
    defer server.Cleanup()
    
    t.Run("complete signup and login flow", func(t *testing.T) {
        // Test actual API endpoints
        signupResponse := server.Expect.POST("/users").
            WithJSON(testUser).
            Expect().
            Status(http.StatusCreated)
            
        // Verify JWT tokens in headers/cookies
        authToken := signupResponse.Header("Authorization").NotEmpty().Raw()
        // ... more assertions
    })
}
```

### Real vs Mock Routes

**Integration Tests Now Use Real Routes**:
```go
// ✅ Current: Uses actual production routes
engine := httphandlers.SetupRoutes()
server := httptest.NewServer(engine)

// ❌ Previous: Used mock routes
router.POST("/users", mockUserHandler)
```

This ensures integration tests verify the exact same code paths users will experience.

## Coverage Requirements

The project enforces coverage thresholds:

- **Unit Tests**: 70% minimum coverage
- **Integration Tests**: 80% minimum coverage
- **Critical Business Logic**: 95% coverage (authentication, subscriptions, core features)

### Generating Coverage Reports

```bash
# HTML coverage reports
./scripts/run-tests.sh all
open coverage-unit.html
open coverage-integration.html

# Console coverage summary
go test -cover ./internal/...
go test -cover ./tests/integration/...
```

## Performance Testing

### Benchmarks

```bash
# Run benchmarks
go test -bench=. -benchmem ./internal/...

# Profile CPU usage
go test -bench=. -cpuprofile=cpu.prof ./internal/service/

# Profile memory usage
go test -bench=. -memprofile=mem.prof ./internal/service/
```

### Load Testing

Integration tests include performance assertions:

```go
func TestAPIPerformance(t *testing.T) {
    start := time.Now()
    
    // Perform operation
    response := server.POST("/api/users").WithJSON(userData).Expect().Status(201)
    
    duration := time.Since(start)
    assert.Less(t, duration, 500*time.Millisecond, "API should respond within 500ms")
}
```

## Debugging Tests

### Common Debug Commands

```bash
# Run specific test with verbose output and environment
set -a && source .env.test && set +a
go test -v -run TestSpecificFunction ./tests/integration/

# Run test with race detection
go test -race ./internal/...

# Debug with Delve
dlv test ./internal/service/ -- -test.run TestSpecificFunction
```

### Test Database Access

```bash
# Connect to test database
psql postgres://test_user:test_pass@localhost:5433/decorebator_test

# View test data during debugging
SELECT * FROM users;
SELECT * FROM wordlists;
```

### Viewing Test Logs

```bash
# Docker logs for test services
docker-compose -f docker-compose.test.yml logs postgres
docker-compose -f docker-compose.test.yml logs redis
docker-compose -f docker-compose.test.yml logs minio
```

## CI/CD Integration

### GitHub Actions

The project includes automated testing via GitHub Actions (`.github/workflows/test.yml`):

- **Unit Tests**: Run on every push and PR
- **Integration Tests**: Full database setup with real API routes
- **Environment Configuration**: Uses `.env.test` with proper external service mocks
- **Coverage Enforcement**: Fail CI if coverage drops below thresholds
- **Security Scanning**: `gosec` and `golangci-lint`
- **Build Verification**: Ensure all binaries compile

### Coverage Reporting

Coverage results are uploaded to Codecov for tracking over time.

## Test Data Management

### Fixtures

**IMPORTANT**: Use production model types from `api/internal/model/` instead of custom test types.

❌ **Current (Deprecated)**: Custom test types that duplicate production models
```go
// DON'T DO THIS - creates maintenance burden and type drift
type TestUser struct {
    ID               int64  `json:"id"`
    Email            string `json:"email"`
    FirstName        string `json:"firstName"`  // Different from production
    SubscriptionPlan string `json:"subscriptionPlan"`  // Missing proper enum type
}
```

✅ **Recommended**: Use actual production models
```go
import "decorebator.com/internal/model"

// Use real production types for realistic testing
func GenerateTestUser() *model.User {
    fake := gofakeit.New(0)
    return &model.User{
        FirstName:        fake.FirstName(),
        LastName:         fake.LastName(),
        Email:           fake.Email(),
        PasswordHash:    "$2a$10$...", // Use real bcrypt hash
        SubscriptionPlan: model.SubscriptionPlanFree,  // Use proper enum
    }
}

func GenerateTestWordlist(userID int64) *model.Wordlist {
    fake := gofakeit.New(0)
    return &model.Wordlist{
        Name:         fake.Sentence(3),
        Description:  fake.Sentence(10),
        UserID:       userID,
        LanguageCode: "en",
        // pgtype fields will be set by database operations
    }
}
```

**Benefits of Using Production Models**:
- **Type Safety**: Tests use exact same types as production
- **Automatic Updates**: Model changes automatically propagate to tests
- **Realistic Testing**: Tests exercise actual JSON marshaling/unmarshaling logic
- **No Duplication**: Eliminates duplicate type definitions
- **Better Coverage**: Tests cover model serialization code paths

### Test Data Generation
```go
// Load predefined test data using production models
server.SeedTestData(t, "users", "wordlists", "words")

// Generate random test data with production types
user := setup.GenerateTestUser()          // Returns *model.User
wordlist := setup.GenerateTestWordlist(userID)  // Returns *model.Wordlist
```

### Data Isolation

Each test run gets a clean database:
- Transactions are used for atomic operations
- Data is cleaned between tests
- Test database is separate from development

### Realistic Test Data

Tests use data that closely resembles production:
- Valid email formats and realistic names
- Proper password hashing with bcrypt
- Realistic word counts and vocabulary
- Valid JWT tokens with proper expiration

## Best Practices

### Migration from Test Types to Production Models

**Action Required**: The current `tests/integration/setup/fixtures.go` file contains custom test types that should be replaced with production models.

**Migration Steps**:

1. **Update fixture generation functions**:
   ```go
   // Before: Returns map[string]interface{}
   func GenerateTestUser() map[string]interface{} { ... }
   
   // After: Returns actual production type
   func GenerateTestUser() *model.User { ... }
   ```

2. **Update test code to use production types**:
   ```go
   // Before: Manual type casting and field access
   userMap := setup.GenerateTestUser()
   email := userMap["email"].(string)
   
   // After: Type-safe field access
   user := setup.GenerateTestUser()
   email := user.Email
   ```

3. **Remove custom test types**:
   - Delete `TestUser`, `TestWordlist`, `TestWord`, `TestDefinition` from fixtures.go
   - Update all references to use `model.User`, `model.Wordlist`, etc.

4. **Handle complex fields properly**:
   ```go
   // Production models may have fields that need special handling
   user := &model.User{
       // Simple fields work as before
       FirstName: fake.FirstName(),
       Email:     fake.Email(),
       
       // Handle enum types properly
       SubscriptionPlan: model.SubscriptionPlanFree,
       
       // pgtype fields will be handled by database operations
       // Don't set CreatedAt/UpdatedAt manually
   }
   ```

### Do's ✅

- **Production Models**: Always use `api/internal/model/` types instead of custom test types
- **External Environment**: Always source environment variables externally, never hardcode in tests
- **Real API Routes**: Use actual production routes in integration tests
- **Descriptive Naming**: Follow the `Test[Subject]_[Scenario]_[Expected]` convention
- **Test Both Paths**: Test both happy path and error conditions
- **Mock External Dependencies**: Use mocks for external APIs
- **Clean Resources**: Always use `defer` or `t.Cleanup()`

### Don'ts ❌

- Don't set environment variables inside test files
- Don't use mock routes when real routes are available  
- Don't test private methods directly
- Don't rely on external services in unit tests
- Don't ignore flaky tests - fix them immediately
- Don't mix unit and integration test concerns

### Performance Considerations

- Unit tests should complete in < 50ms each
- Integration tests should complete in < 500ms each
- Use `t.Parallel()` for independent tests
- Minimize database operations in setup/teardown
- Cache expensive setup operations when possible

## Troubleshooting

### Common Issues

**Environment Configuration Errors**:
```bash
# Ensure .env.test exists
cp .env.test.example .env.test

# Verify environment is sourced
set -a && source .env.test && set +a
env | grep POSTGRES
```

**Database Connection Errors**:
```bash
# Ensure test services are running
docker-compose -f docker-compose.test.yml up -d

# Check service health
docker-compose -f docker-compose.test.yml ps
```

**Port Conflicts**:
```bash
# Check if ports are in use
lsof -i :5433  # Test PostgreSQL
lsof -i :6380  # Test Redis
lsof -i :9001  # Test MinIO
```

**Migration Failures**:
```bash
# Run migrations manually
cd api
set -a && source .env.test && set +a
go run ./cmd/migrate/main.go
```

### Getting Help

1. Verify `.env.test` file exists and contains all required variables
2. Check test output for specific error messages
3. Run with `-v` flag for verbose output
4. Verify test services are running via Docker Compose
5. Check service logs for integration test issues

## Contributing

When adding new tests:

1. **Environment**: Never set environment variables in test code - use external configuration
2. **Real Routes**: Use actual API routes for integration tests, not mocks
3. **Patterns**: Follow established patterns and naming conventions
4. **Coverage**: Add both unit and integration tests for new features
5. **Documentation**: Update this README if adding new testing utilities

The testing framework now provides true integration testing with real API routes and external environment configuration, ensuring tests accurately represent production behavior.