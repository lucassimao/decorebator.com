# Logging Architecture Improvement Plan

## 🔍 Current State Analysis

### **Good Examples Found:**
- `internal/mail/mail.go` - Excellent structured logging with context
- `internal/service/*_worker.go` - Good error handling in background workers
- `internal/common/logger.go` - Well-configured centralized logger

### **Major Issues Identified:**
1. **Inconsistent logging patterns** across layers
2. **Missing request correlation** and tracing
3. **Silent failures** in critical operations
4. **Poor error context** in repositories
5. **No standardized error handling** middleware

## 📋 Actionable Improvement Plan

### **Phase 1: Foundation (High Priority)**

#### 1.1 Create Request Context Middleware
**File to create:** `internal/http/middleware/logging.go`
- Generate unique request IDs
- Log request start/end with timing
- Inject user context into logger
- Sanitize sensitive data from logs

#### 1.2 Standardize Error Response Structure
**Files to modify:**
- `internal/http/setup.go` - Add error handling middleware
- All controllers in `internal/http/` - Standardize error responses

#### 1.3 Create Logging Utilities
**File to create:** `internal/common/logging_utils.go`
- Helper functions for consistent structured logging
- Error wrapping with automatic logging
- Business event logging helpers

### **Phase 2: HTTP Layer Improvements (High Priority)**

#### 2.1 Fix Silent Failures
**Specific files needing immediate attention:**

**`internal/http/user.go`:**
- Line 167: `service.UpdatePassword()` - No error handling
- Line 114: `go mail.SendWelcomeEmail()` - Goroutine errors ignored
- Add proper error logging for all authentication operations

**`internal/http/subscription.go`:**
- All endpoints return generic errors without logging actual causes
- Add structured logging with subscription context (plan, user ID, amount)

**`internal/http/quiz.go`:**
- Missing logging for quiz operations
- No tracking of quiz completion/failure rates

#### 2.2 Add Business Event Logging
**Files to enhance:**
- `internal/http/word.go` - Log word creation/deletion with context
- `internal/http/wordlist.go` - Log wordlist operations
- `internal/http/user.go` - Log authentication events

### **Phase 3: Service Layer Standardization (Medium Priority)**

#### 3.1 Implement Consistent Error Patterns
**Files needing improvement:**

**`internal/service/word.go`:**
```go
// Current: Silent failures
// Target: Log errors with full context (user, wordlist, word details)
```

**`internal/service/wordlist.go`:**
```go
// Current: Wrapped errors without logging
// Target: Log business operations with timing and success/failure
```

**`internal/service/user.go`:**
```go
// Current: Basic error logging
// Target: Enhanced context with operation details
```

#### 3.2 Add Performance Monitoring
- Log operation timing for critical business functions
- Track success/failure rates
- Monitor resource usage patterns

### **Phase 4: Repository Layer Enhancement (Medium Priority)**

#### 4.1 Add Database Operation Logging
**Files to modify:**
- `internal/repository/word.go` - Add query logging with parameters
- `internal/repository/wordlist.go` - Log database operations
- `internal/repository/user.go` - Add timing and error context
- `internal/repository/definition.go` - Enhance existing logging

#### 4.2 Database Error Context
- Add business context to database errors
- Log query parameters (sanitized)
- Track query performance

### **Phase 5: Advanced Features (Low Priority)**

#### 5.1 Distributed Tracing
**File to create:** `internal/common/tracing.go`
- OpenTelemetry integration
- Request correlation across services
- Performance tracing

#### 5.2 Error Aggregation
**File to create:** `internal/common/error_aggregation.go`
- Centralized error collection
- Error pattern analysis
- Alert thresholds

## 🎯 Implementation Priority

### **Immediate (This Week)**
1. Fix silent failures in `user.go`, `subscription.go`, `quiz.go`
2. Create logging middleware for request correlation
3. Standardize error responses across HTTP layer

### **Short Term (Next 2 Weeks)**
1. Implement consistent service layer logging
2. Add business event logging
3. Create logging utility helpers

### **Medium Term (Next Month)**
1. Enhance repository layer logging
2. Add performance monitoring
3. Implement distributed tracing foundation

## 🔧 Specific File Improvements Needed

### **Critical Files (Fix First):**
- `internal/http/user.go` - Lines 114, 167 (silent failures)
- `internal/http/subscription.go` - All endpoints lack error logging
- `internal/http/quiz.go` - Missing operation logging
- `internal/service/word.go` - Silent database operations
- `internal/service/wordlist.go` - Poor error context

### **Follow the Good Patterns From:**
- `internal/mail/mail.go` - Excellent structured logging example
- `internal/service/text_to_speech_worker.go` - Good error handling pattern
- `internal/service/definition_fetcher_worker.go` - Context-rich logging

## 📝 Implementation Guidelines

### **Logging Standards to Adopt:**

#### Request-Level Logging:
```go
logger := common.Logger.With(
    "requestID", requestID,
    "userID", userID,
    "endpoint", ctx.Request.URL.Path,
    "method", ctx.Request.Method,
)
```

#### Business Operation Logging:
```go
logger := common.Logger.With(
    "operation", "createWord",
    "userID", userID,
    "wordlistID", wordlistID,
    "wordName", input.Name,
)
logger.Info("creating word")
// ... operation ...
if err != nil {
    logger.Error("failed to create word", "error", err)
    return fmt.Errorf("failed to create word: %w", err)
}
logger.Info("word created successfully", "wordID", word.ID)
```

#### Database Operation Logging:
```go
logger := common.Logger.With(
    "operation", "saveWord",
    "table", "words",
    "userID", userID,
)
start := time.Now()
// ... database operation ...
if err != nil {
    logger.Error("database operation failed", 
        "error", err, 
        "duration", time.Since(start),
        "query", "INSERT INTO words...",
    )
    return fmt.Errorf("failed to save word: %w", err)
}
logger.Debug("database operation completed", "duration", time.Since(start))
```

### **Error Response Standards:**
```go
// HTTP layer should always log errors before returning
if err != nil {
    logger.Error("operation failed", "error", err)
    c.JSON(http.StatusInternalServerError, gin.H{
        "error": "Failed to create word",
        "requestID": requestID, // for correlation
    })
    return
}
```

This plan will create a unified, traceable error handling system that makes debugging and monitoring much more effective.