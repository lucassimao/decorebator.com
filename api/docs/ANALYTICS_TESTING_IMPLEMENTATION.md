# Analytics Testing Implementation Guide

This document describes the comprehensive analytics testing suite implemented for the Decorebator API, providing detailed guidance on testing analytics endpoints with proper data validation and calculation verification.

## 📊 Overview

The analytics testing suite covers all 8 analytics endpoints with focused, metric-specific test files that validate database calculations, query accuracy, and business logic implementation.

### Analytics Endpoints Tested

| Endpoint | Purpose | Key Metrics Tested |
|----------|---------|-------------------|
| `/analytics/wordlists/:id/mastery` | Word mastery statistics | Mastery levels, accuracy, streaks, highest box |
| `/analytics/wordlists/:id/progress` | Daily learning progress | Words studied, attempts, accuracy rates, response times |
| `/analytics/wordlists/:id/distribution` | Historical box distribution | Daily snapshots, box counts over time |
| `/analytics/wordlists/:id/current-distribution` | Current box distribution | Real-time box counts, MAX(box_id) logic |
| `/analytics/wordlists/:id/quiz-performance` | Quiz type performance | Success rates, response times, attempt counts |
| `/analytics/wordlists/:id/practice-time` | Daily practice time | Time calculations, outlier filtering |
| `/analytics/wordlists/:id/overview` | Dashboard statistics | Combined metrics, today's stats, streaks |
| `/analytics/progress-summary` | Multi-wordlist summary | Progress percentages, activity dates |

## 🗂️ Test Suite Organization

### Directory Structure
```
tests/integration/analytics/
├── helpers.go                    # Shared utilities and common test data setup
├── word_mastery_test.go          # Word mastery endpoint tests
├── quiz_performance_test.go      # Quiz performance endpoint tests  
├── box_distribution_test.go      # Box distribution endpoint tests
├── learning_progress_test.go     # Learning progress endpoint tests
├── practice_time_test.go         # Practice time endpoint tests
├── wordlist_overview_test.go     # Wordlist overview endpoint tests
└── progress_summary_test.go      # Progress summary endpoint tests
```

### Design Principles

1. **One File Per Metric** - Each test file focuses on a single analytics endpoint
2. **Comprehensive Coverage** - Tests calculations, edge cases, error handling, parameters
3. **Known Data Testing** - Uses predetermined test data to verify exact calculations
4. **Database Query Validation** - Tests complex SQL queries directly against PostgreSQL
5. **Shared Utilities** - Common helpers reduce code duplication

## 🧪 Testing Methodology

### 1. Data Seeding Strategy

Each test creates controlled test data with known expected outcomes:

```go
// Example: Word mastery test data with precise calculations
testWords := []struct {
    name            string
    masteryLevel    float64
    totalAttempts   int
    correctAttempts int
    streakCount     int
    boxLevel        int64
}{
    {"excellent", 0.95, 20, 19, 8, 7},    // 95% mastery, high performance
    {"good", 0.75, 16, 12, 4, 5},         // 75% mastery, good performance  
    {"struggling", 0.25, 12, 3, 0, 2},    // 25% mastery, poor performance
}
```

### 2. Calculation Verification

Tests verify that endpoint responses match expected database calculations:

```go
// Verify accuracy calculation: (correct_attempts / total_attempts) * 100
expectedAccuracy := float64(correctAttempts) / float64(totalAttempts) * 100
statObj.Value("accuracy").Number().InDelta(expectedAccuracy, 0.01)

// Verify progress percentage: (wordsMastered / totalWords) * 100  
expectedProgress := (wordsMastered / totalWords) * 100
wordlistObj.Value("progressPercent").Number().InDelta(expectedProgress, 0.1)
```

### 3. Complex Query Testing

Tests validate complex PostgreSQL queries including CTEs, window functions, and aggregations:

```go
// Tests MAX(box_id) logic for box distribution
// Verifies that words with multiple definitions use highest box level
distribution.Value("box5").Number().Equal(1) // One word at max box 5
distribution.Value("box7").Number().Equal(1) // One word at max box 7
```

### 4. Edge Case Coverage

Each test file includes comprehensive edge case testing:

- **Empty Data**: Tests with no words, no quiz data, no learning progress
- **Boundary Conditions**: Tests with exactly 0%, 50%, 100% values
- **Parameter Validation**: Tests invalid query parameters and defaults
- **Error Scenarios**: Tests unauthorized access, nonexistent resources

## 📋 Test File Details

### Word Mastery Tests (`word_mastery_test.go`)

**Purpose**: Tests word mastery analytics calculations and data accuracy.

**Key Test Cases**:
- `TestWordMasteryEndpoint_CalculationAccuracy` - Verifies mastery level, accuracy, streak calculations
- `TestWordMasteryEndpoint_MultipleBoxLevels` - Tests words distributed across Leitner boxes
- `TestWordMasteryEndpoint_EmptyWordlist` - Tests empty wordlist handling

**Critical Validations**:
- Mastery level calculations (weighted average: 70% box progress + 30% historical accuracy)
- Accuracy calculations (correct_attempts / total_attempts)
- Highest box level per word (MAX(box_id) from leitner_system_tracking)
- Result ordering by mastery level DESC

```go
// Example validation
statObj.Value("masteryLevel").Number().InDelta(expectedWord.MasteryLevel, 0.01)
statObj.Value("accuracy").Number().InDelta(expectedWord.Accuracy, 0.01)
statObj.Value("highestBox").ValueEqual("highestBox", expectedWord.BoxLevel)
```

### Quiz Performance Tests (`quiz_performance_test.go`)

**Purpose**: Tests quiz type performance metrics and response time filtering.

**Key Test Cases**:
- `TestQuizPerformanceEndpoint_MetricCalculations` - Tests success rates and response times
- `TestQuizPerformanceEndpoint_ResponseTimeFiltering` - Tests 200ms-30s outlier filtering
- `TestQuizPerformanceEndpoint_ResultOrdering` - Tests ordering by success rate DESC

**Critical Validations**:
- Success rate calculations (correct_attempts / total_attempts * 100)
- Response time averaging with outlier filtering (200ms ≤ time ≤ 30000ms)
- Quiz type aggregation across all attempts
- Result ordering by performance

```go
// Verify success rate calculation
expectedSuccessRate := float64(correctAttempts) / float64(totalAttempts) * 100
perfObj.Value("successRate").Number().InDelta(expectedSuccessRate, 0.1)

// Verify response time filtering
expectedAvg := (1500 + 2000) / 2 // Only valid times included
perfObj.Value("avgResponseMs").Number().InDelta(expectedAvg, 1.0)
```

### Box Distribution Tests (`box_distribution_test.go`)

**Purpose**: Tests current and historical Leitner box distribution calculations.

**Key Test Cases**:
- `TestCurrentBoxDistributionEndpoint_AccurateDistribution` - Tests real-time box counts
- `TestCurrentBoxDistributionEndpoint_MaxBoxLogic` - Tests MAX(box_id) logic for words with multiple definitions
- `TestBoxDistributionHistoryEndpoint_HistoricalData` - Tests historical snapshots

**Critical Validations**:
- Word counting by highest box level achieved (MAX(box_id) per word)
- Historical snapshot accuracy and date ordering
- Total word count consistency across all boxes
- Empty wordlist handling (all boxes = 0)

```go
// Verify MAX(box_id) logic - word should be counted at highest box only
distribution.Value("box5").Number().Equal(1) // Word 1: boxes 2,3,5 → counted at 5
distribution.Value("box7").Number().Equal(1) // Word 2: boxes 1,4,7 → counted at 7
distribution.Value("totalWords").Number().Equal(2) // Only 2 words total
```

### Learning Progress Tests (`learning_progress_test.go`)

**Purpose**: Tests daily learning progress aggregation and calculations.

**Key Test Cases**:
- `TestLearningProgressEndpoint_DailyAggregation` - Tests daily statistics aggregation
- `TestLearningProgressEndpoint_AccuracyCalculations` - Tests accuracy edge cases (0%, 100%)
- `TestLearningProgressEndpoint_ParameterValidation` - Tests date range parameters

**Critical Validations**:
- Daily aggregation accuracy (words studied, attempts, response times)
- Accuracy rate calculations with proper rounding
- Date range parameter handling and validation
- Result ordering by date DESC

```go
// Verify daily accuracy calculation
expectedAccuracy := float64(correctAttempts) / float64(totalAttempts) * 100
dayObj.Value("accuracyRate").Number().InDelta(expectedAccuracy, 0.1)

// Verify date range validation
json.Value("days").Number().Equal(expectedDays) // Parameter clamping
```

### Practice Time Tests (`practice_time_test.go`)

**Purpose**: Tests practice time calculations and response time filtering.

**Key Test Cases**:
- `TestPracticeTimeEndpoint_TimeCalculations` - Tests practice time aggregation
- `TestPracticeTimeEndpoint_ResponseTimeFiltering` - Tests outlier filtering (200ms-30s)
- `TestPracticeTimeEndpoint_CustomDateRange` - Tests date range limits (7 default, 30 max)

**Critical Validations**:
- Practice time aggregation from filtered response times
- Milliseconds to minutes conversion accuracy
- Outlier filtering consistency
- Quiz count vs practice time separation

```go
// Verify time conversion accuracy
expectedMinutes := float64(expectedPracticeTimeMs) / 60000.0
dayObj.Value("practiceTimeMinutes").Number().InDelta(expectedMinutes, 0.1)

// Verify outlier filtering
expectedFilteredTime := 3800 // Only valid response times included
dayObj.Value("practiceTimeMs").Number().Equal(expectedFilteredTime)
```

### Wordlist Overview Tests (`wordlist_overview_test.go`)

**Purpose**: Tests dashboard statistics combining multiple analytics metrics.

**Key Test Cases**:
- `TestWordlistOverviewEndpoint_DashboardStats` - Tests combined dashboard metrics
- `TestWordlistOverviewEndpoint_MasteryCalculations` - Tests mastery aggregations
- `TestWordlistOverviewEndpoint_TodayStats` - Tests today's activity metrics

**Critical Validations**:
- Mastery threshold application (>= 0.8 for mastered words)
- Today's statistics from learning_progress table
- Streak calculations using recursive CTE logic
- Optional field handling (NULL values)

```go
// Verify mastery threshold
stats.Value("wordsMastered").Number().Equal(expectedWordsMastered) // Count >= 0.8

// Verify today's stats
stats.Value("wordsStudiedToday").Number().Equal(expectedWordsToday)
stats.Value("accuracyToday").Number().InDelta(expectedAccuracy, 0.1)
```

### Progress Summary Tests (`progress_summary_test.go`)

**Purpose**: Tests multi-wordlist progress summary aggregation.

**Key Test Cases**:
- `TestProgressSummaryEndpoint_MultipleWordlists` - Tests multi-wordlist aggregation
- `TestProgressSummaryEndpoint_ProgressCalculations` - Tests progress percentage calculations
- `TestProgressSummaryEndpoint_LastActivityDate` - Tests activity date tracking

**Critical Validations**:
- Progress percentage calculations across multiple wordlists
- Last activity date accuracy and formatting
- Streak calculations per wordlist
- Wordlist metadata accuracy (name, language)

```go
// Verify progress calculation
expectedProgress := (wordsMastered / totalWords) * 100
wordlistObj.Value("progressPercent").Number().InDelta(expectedProgress, 0.1)

// Verify logical constraints
assert.True(t, wordsMastered <= totalWords, "Words mastered ≤ total words")
```

## 🔧 Shared Utilities (`helpers.go`)

### Common Helper Functions

```go
// createBasicWordStructure - Creates word, definition, and Leitner tracking
func createBasicWordStructure(t *testing.T, db *pgxpool.Pool, ctx context.Context, 
    userID, wordlistID int64, wordName string) (int64, int64, int64)

// getUserID - Gets most recently created user ID from WithTestUser()
func getUserID(t *testing.T, db *pgxpool.Pool, ctx context.Context) int64

// createTestWordlist - Creates basic wordlist for testing
func createTestWordlist(t *testing.T, db *pgxpool.Pool, ctx context.Context, 
    userID int64, name, description string) int64

// setupCommonTestData - Sets up basic test data structure
func setupCommonTestData(t *testing.T, db *pgxpool.Pool, ctx context.Context, 
    testName string) *CommonTestData
```

### Usage Pattern

```go
func TestAnalyticsEndpoint_SomeMetric(t *testing.T) {
    server := setup.NewTestServer(t)
    defer server.Cleanup()
    
    token := server.WithTestUser(t)
    ctx := context.Background()
    
    // Use shared helper to create basic test structure
    testData := setupCommonTestData(t, server.DB, ctx, "SomeMetric")
    
    // Add metric-specific test data
    // ... create specific quiz performance, learning progress, etc.
    
    // Test endpoint
    response := server.Expect.GET(fmt.Sprintf("/analytics/endpoint/%d", testData.WordlistID)).
        WithHeader("Authorization", token).
        Expect().
        Status(http.StatusOK)
    
    // Verify calculations
    // ... specific assertions
}
```

## 🚀 Running Analytics Tests

### Individual Test Execution

```bash
# Run specific analytics test file
go test -v ./tests/integration/analytics -run TestWordMastery

# Run specific test case
go test -v ./tests/integration/analytics -run TestWordMasteryEndpoint_CalculationAccuracy

# Run all analytics tests
go test -v ./tests/integration/analytics

# Run with coverage
go test -v -coverprofile=coverage.out ./tests/integration/analytics
```

### Database Requirements

Analytics tests require PostgreSQL with test database setup:

```bash
# Setup test database
make setup-test-db

# Run migrations
make migrate-up-test

# Run analytics tests with proper database
DATABASE_URL="postgres://user:pass@localhost:5433/decorebator_test" \
  go test -v ./tests/integration/analytics
```

## 📊 Database Query Validation

Following CLAUDE.md requirements, all analytics tests validate database queries directly:

### Query Testing Pattern

1. **Insert Known Data**: Create test data with predetermined values
2. **Execute Endpoint**: Call analytics endpoint to get calculated results
3. **Verify Calculations**: Compare endpoint results with expected calculations
4. **Test Edge Cases**: Validate boundary conditions and error scenarios

### Example: Streak Calculation Testing

```go
// Create learning progress data for consecutive days
for i := 0; i < expectedStreak; i++ {
    date := time.Now().AddDate(0, 0, -i)
    _, err := db.Exec(ctx,
        `INSERT INTO learning_progress (user_id, wordlist_id, date, words_studied, 
         total_quiz_attempts, correct_attempts, average_response_time_ms) 
         VALUES ($1, $2, $3::date, 2, 10, 8, 1500)`,
        userID, wordlistID, date.Format("2006-01-02"))
    require.NoError(t, err)
}

// Test endpoint calculation
response := server.Expect.GET(fmt.Sprintf("/analytics/wordlists/%d/overview", wordlistID))
stats := response.JSON().Object().Value("stats").Object()
stats.Value("currentStreak").Number().Equal(expectedStreak)
```

### Complex Query Validation

Analytics tests specifically validate complex PostgreSQL features:

- **Recursive CTEs** - Streak calculations with gap-and-islands logic
- **Window Functions** - ROW_NUMBER, LAG, LEAD for time series analysis  
- **Complex JOINs** - Multi-table aggregations across analytics tables
- **Conditional Aggregation** - COUNT(CASE WHEN...) patterns for box distributions
- **Type Casting** - Date/timestamp handling and interval calculations

## 🎯 Business Logic Validation

### Mastery Thresholds

```go
// Words with mastery_level >= 0.8 are considered mastered
expectedWordsMastered := 2 // Out of 4 words, 2 have >= 0.8 mastery
stats.Value("wordsMastered").Number().Equal(expectedWordsMastered)
```

### Response Time Filtering

```go
// Only response times between 200ms and 30000ms are included in calculations
validResponseTimes := []int{1500, 2000, 300} // Valid range
invalidResponseTimes := []int{100, 45000}     // Filtered out
expectedAverage := (1500 + 2000 + 300) / 3   // Only valid times
```

### Box Progression Logic

```go
// Words are counted at their highest box level (MAX(box_id))
// Word with definitions in boxes [2, 3, 5] is counted in box 5
// Word with definitions in boxes [1, 4, 7] is counted in box 7
```

## 📈 Performance Considerations

### Test Data Volume

Tests use realistic data volumes to validate query performance:

- **30 days** of learning progress data
- **Multiple quiz types** with varied attempt counts
- **Realistic response times** (200ms - 8000ms range)
- **Multiple words per wordlist** (5-20 words typical)

### Query Performance Testing

```go
// Example: Benchmark analytics query performance
func BenchmarkAnalyticsQuery_WordMastery(b *testing.B) {
    // Setup test data
    testData := setupLargeDataset(b, 1000) // 1000 words
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        // Execute analytics query
        _, err := analyticsService.WordMastery(ctx, userID, wordlistID)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

## 🔍 Error Handling Testing

### Comprehensive Error Coverage

Each test file includes error scenario testing:

```go
func TestAnalyticsEndpoint_ErrorCases(t *testing.T) {
    server := setup.NewTestServer(t)
    defer server.Cleanup()
    
    token := server.WithTestUser(t)
    
    t.Run("NonexistentWordlist", func(t *testing.T) {
        server.Expect.GET("/analytics/wordlists/999999/mastery").
            WithHeader("Authorization", token).
            Expect().
            Status(http.StatusNotFound)
    })
    
    t.Run("InvalidWordlistID", func(t *testing.T) {
        server.Expect.GET("/analytics/wordlists/invalid/mastery").
            WithHeader("Authorization", token).
            Expect().
            Status(http.StatusBadRequest)
    })
    
    t.Run("UnauthorizedAccess", func(t *testing.T) {
        server.Expect.GET("/analytics/wordlists/1/mastery").
            Expect().
            Status(http.StatusUnauthorized)
    })
}
```

## 🏆 Benefits of This Implementation

### 1. **Comprehensive Coverage**
- Tests all 8 analytics endpoints with dedicated test files
- Validates calculations, edge cases, parameters, and error handling
- Covers complex database queries and business logic

### 2. **Maintainability**
- Focused test files are easier to understand and modify
- Shared utilities reduce code duplication
- Clear naming conventions and documentation

### 3. **Reliability**
- Uses known test data to verify exact calculations
- Tests directly against PostgreSQL database
- Validates complex SQL query accuracy

### 4. **Performance**
- Tests can run in parallel for faster CI/CD
- Realistic data volumes validate query performance
- Proper database indexing verification

### 5. **Developer Experience**
- Easy to run individual metric tests during development
- Clear test failure messages with specific calculation details
- Comprehensive error scenario coverage

## 🚀 Running Analytics Tests

### Quick Commands (Updated 2025)

```bash
# Run all analytics tests with structured output
./scripts/run-tests.sh integration

# Run specific analytics test file
go test -v ./tests/integration/analytics/word_mastery_test.go

# Run analytics tests in watch mode during development
./scripts/run-tests.sh watch

# Generate coverage report
make coverage-html
```

### CI-Compatible Testing

```bash
# Full test suite with proper environment setup
./scripts/run-tests.sh all

# CI-style reporting with gotestsum
make test-ci
```

This analytics testing implementation provides a robust foundation for validating the accuracy and reliability of all analytics functionality in the Decorebator API.