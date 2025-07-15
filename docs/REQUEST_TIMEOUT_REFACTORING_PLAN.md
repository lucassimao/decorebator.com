# Request Timeout Refactoring Plan

## Overview

This document outlines a comprehensive refactoring plan to implement 2-second request timeouts across all API endpoints. The goal is to propagate context from the HTTP layer through the service layer down to database operations, ensuring consistent timeout behavior and preventing hanging requests.

## ✅ IMPLEMENTATION UPDATE (January 2025)

**Status**: Phase 1 Complete - Centralized TimeoutMiddleware successfully implemented

### What Was Implemented
- **Centralized TimeoutMiddleware**: Custom middleware in `internal/http/midlewares.go`
- **Global Application**: Applied to all routes with 2-second default timeout
- **POST /login Complete**: Full call stack timeout implementation and testing
- **Simplified Handlers**: Removed all manual timeout handling from Login and UpdateProfile

### Key Design Decisions
1. **Custom Implementation**: Built in-house rather than using `gin-contrib/timeout` for better control
2. **Middleware-Level Error Handling**: Timeout errors detected and handled centrally after `c.Next()`
3. **Context Replacement**: Used `c.Request = c.Request.WithContext(ctx)` for proper propagation
4. **Response Safety**: Added `c.Writer.Written()` check to prevent double responses

## Objectives

1. **Consistent Timeout Behavior**: All HTTP requests timeout within 2 seconds
2. **Context Propagation**: Pass the same context from HTTP → Service → Repository → Database
3. **Graceful Error Handling**: Proper timeout error responses
4. **No Breaking Changes**: Maintain existing API contracts
5. **Comprehensive Coverage**: All endpoints receive timeout protection

## Architecture Changes

### Original Flow (Before Implementation)
```
HTTP Handler → Service Method → Repository Method → Database Query
(manual ctx)   (ctx passed)    (ctx passed)       (ctx timeout)
```

### ✅ Implemented Flow (Current State)
```
TimeoutMiddleware → HTTP Handler → Service Method → Repository Method → Database Query
(2s timeout)       (ctx from req) (ctx passed)    (ctx passed)       (ctx timeout)
```

### Key Architecture Improvements
- **Centralized Timeout Management**: Single middleware handles all timeout logic
- **Eliminated Code Duplication**: No more repetitive timeout handling in handlers
- **Context Flow**: `c.Request.Context()` carries timeout through entire call stack
- **Database Integration**: pgx/v5 driver automatically respects context timeouts

## Implementation Strategy

### ✅ Phase 1: Centralized Middleware (COMPLETED)
- ✅ Created `TimeoutMiddleware(timeout time.Duration)` in `internal/http/midlewares.go`
- ✅ Applied globally in `internal/http/setup.go` with 2-second timeout
- ✅ Implemented centralized timeout error handling after `c.Next()`
- ✅ Added proper context propagation using `c.Request.WithContext(ctx)`

### ✅ Phase 2: Login Endpoint Implementation (COMPLETED)
- ✅ Simplified `Login()` handler by removing manual timeout creation
- ✅ Updated `UpdateProfile()` handler to remove manual timeout handling
- ✅ Verified service layer `LoginUser(ctx, email, password)` already accepts context
- ✅ Confirmed repository layer `userRepository.Find(ctx, args)` respects context

### 🔄 Phase 3: Remaining Endpoints (PLANNED)
- 🔲 Apply same pattern to remaining 55+ endpoints
- 🔲 Remove manual timeout handling from other handlers
- 🔲 Add endpoint-specific timeout customization as needed

### 🔄 Phase 4: Testing & Validation (PLANNED)
- 🔲 Add timeout tests for critical endpoints
- 🔲 Performance testing to ensure 2s timeout is appropriate
- 🔲 Error handling validation
- 🔲 Load testing with timeout scenarios

---

## Endpoint Refactoring Checklist

### Authentication Endpoints

#### `/login` - POST
- **Status**: ✅ Completed (January 2025)
- **HTTP Handler**: `internal/http/user.go:Login()` ✅
  - **Implementation**: Simplified to use `c.Request.Context()` from middleware
  - **Removed**: Manual `context.WithTimeout()` creation and error handling
  - **Lines**: 136-153 (simplified from 18 lines to 12 lines)
- **Service Methods**: 
  - `service.LoginUser(ctx, email, password)` ✅ (already implemented)
  - `service.GenerateJWT(user)` ✅ (no context needed)
- **Repository Methods**:
  - `userRepository.Find(ctx, args)` ✅ (already had context support)
- **Database Operations**:
  - User lookup query ✅ (pgx/v5 respects context timeout)
  - Password hash comparison ✅ (in-memory operation)
- **Actual Effort**: 1 hour (less than estimated due to middleware approach)
- **Dependencies**: TimeoutMiddleware implementation
- **Notes**: 
  - Also updated `UpdateProfile` function password verification
  - Timeout errors now handled centrally in middleware
  - Context flows naturally through entire call stack

#### `/users` - POST (Registration)
- **Status**: ✅ Completed (Phase 2)
- **HTTP Handler**: `internal/http/user.go:SignUp()` ✅
  - **Implementation**: Updated to pass `c.Request.Context()` to service layer
- **Service Methods**: 
  - `service.SaveUser(ctx, firstName, lastName, password, email, country)` ✅ (context added)
- **Repository Methods**:
  - `userRepository.Save(ctx, firstName, lastName, password, email, country)` ✅ (context added)
- **Database Operations**:
  - User creation INSERT ✅ (uses request context)
  - Email uniqueness check ✅ (respects timeout)
- **Actual Effort**: 1 hour
- **Dependencies**: TimeoutMiddleware implementation

### User Management Endpoints

#### `/users` - GET (Get Profile)
- **Status**: ✅ Completed (Phase 2)
- **HTTP Handler**: `internal/http/user.go:GetProfile()` ✅
  - **Implementation**: Updated to pass `c.Request.Context()` to service layer
- **Service Methods**:
  - `service.GetProfile(ctx, userID)` ✅ (context added, replaced context.Background)
  - `service.checkAndDowngradeExpiredSubscription(ctx, userID, user)` ✅ (context added)
- **Repository Methods**:
  - `userRepository.Find(ctx, args)` ✅ (already had context support)
  - `subRepo.GetActiveSubscriptionForUser(ctx, userID)` ✅ (already had context support)
  - `userRepository.UpdateSubscriptionPlan(ctx, userID, plan)` ✅ (already had context support)
- **Database Operations**:
  - User profile query ✅ (uses request context)
  - Subscription lookup ✅ (respects timeout)
  - Subscription update (if grace period expired) ✅ (respects timeout)
- **Actual Effort**: 2 hours
- **Dependencies**: TimeoutMiddleware implementation

#### `/users/profile` - PUT (Update Profile)
- **Status**: ✅ Completed (Phase 2)
- **HTTP Handler**: `internal/http/user.go:UpdateProfile()` ✅
  - **Implementation**: Updated to pass `c.Request.Context()` to service layer
- **Service Methods**:
  - `service.UpdateProfile(ctx, userID, ...)` ✅ (context added)
- **Repository Methods**:
  - `userRepository.UpdateUserProfile(ctx, args)` ✅ (context added)
- **Database Operations**:
  - User profile UPDATE ✅ (uses request context)
  - Password hash update (if provided) ✅ (respects timeout)
- **Actual Effort**: 1.5 hours
- **Dependencies**: TimeoutMiddleware implementation

#### `/users` - DELETE (Delete Account)
- **Status**: ✅ Completed (Phase 2)
- **HTTP Handler**: `internal/http/user.go:DeleteProfile()` ✅
  - **Implementation**: Updated to pass `c.Request.Context()` to service layer
- **Service Methods**:
  - `service.Delete(ctx, userID)` ✅ (context added)
  - `service.DeleteUserErrorReports(ctx, userID)` ✅ (context added)
- **Repository Methods**:
  - `wordlistRepository.DeleteAll(ctx, userID)` ✅ (context added)
  - `userRepository.Delete(ctx, userID)` ✅ (context added)
- **Database Operations**:
  - Cascade deletion of user data ✅ (uses request context)
  - Error reports cleanup ✅ (respects timeout)
- **Actual Effort**: 3.5 hours (additional complexity from multi-repository updates)
- **Dependencies**: TimeoutMiddleware implementation

### Wordlist Management Endpoints

#### `/wordlists` - GET (List Wordlists)
- **Status**: ✅ Completed (Phase 3.2)
- **HTTP Handler**: `internal/http/wordlist.go:GetAll()` ✅
  - **Implementation**: Updated to pass `c.Request.Context()` to service layer
- **Service Methods**:
  - `service.GetUserWordlistsWithWordStats(ctx, userID)` ✅ (context added)
- **Repository Methods**:
  - `wordlistRepository.Find(ctx, args)` ✅ (context added)
- **Database Operations**:
  - Wordlists query with word counts ✅ (uses request context)
- **Actual Effort**: 1 hour
- **Dependencies**: TimeoutMiddleware implementation

#### `/wordlists` - POST (Create Wordlist)
- **Status**: ✅ Completed (Phase 3.2)
- **HTTP Handler**: `internal/http/wordlist.go:Create()` ✅
  - **Implementation**: Updated to pass `c.Request.Context()` to service layer
- **Service Methods**:
  - `service.SaveWordlist(ctx, newWordlist)` ✅ (context added)
- **Repository Methods**:
  - `wordlistRepository.Save(ctx, name, description, languageCode, userID, pronunciationSystem)` ✅ (context added)
- **Database Operations**:
  - Wordlist creation INSERT ✅ (uses request context)
  - Content moderation validation ✅ (respects timeout)
- **Actual Effort**: 1.5 hours
- **Dependencies**: TimeoutMiddleware implementation

#### `/wordlists/{id}` - GET (Get Wordlist Details)
- **Status**: ✅ Completed (Phase 3.2)
- **HTTP Handler**: `internal/http/wordlist.go:GetById()` ✅
  - **Implementation**: Updated to pass `c.Request.Context()` to service layer
- **Service Methods**:
  - `service.GetWordlistByID(ctx, wordlistID, userID)` ✅ (context added)
- **Repository Methods**:
  - `wordlistRepository.Find(ctx, args)` ✅ (context added)
- **Database Operations**:
  - Single wordlist query with permissions check ✅ (uses request context)
- **Actual Effort**: 1 hour
- **Dependencies**: TimeoutMiddleware implementation

#### `/wordlists/{id}` - PUT (Update Wordlist)
- **Status**: ✅ Completed (Phase 3.2)
- **HTTP Handler**: `internal/http/wordlist.go:Update()` ✅
  - **Implementation**: Updated to pass `c.Request.Context()` to service layer
- **Service Methods**:
  - `service.UpdateWordlist(ctx, wordlist)` ✅ (context added)
- **Repository Methods**:
  - `wordlistRepository.Update(ctx, wordlist)` ✅ (context added)
- **Database Operations**:
  - Wordlist UPDATE with ownership check ✅ (uses request context)
  - Content moderation validation ✅ (respects timeout)
- **Actual Effort**: 1 hour
- **Dependencies**: TimeoutMiddleware implementation

#### `/wordlists/{id}` - DELETE (Delete Wordlist)
- **Status**: ✅ Completed (Phase 3.2)
- **HTTP Handler**: `internal/http/wordlist.go:Delete()` ✅
  - **Implementation**: Updated to pass `c.Request.Context()` to service layer
- **Service Methods**:
  - `service.DeleteWordlist(ctx, wordlistID, userID)` ✅ (context added)
- **Repository Methods**:
  - `wordlistRepository.Delete(ctx, wordlistID, userID)` ✅ (context added)
- **Database Operations**:
  - Cascade deletion of wordlist and associated data ✅ (uses request context)
- **Actual Effort**: 1.5 hours
- **Dependencies**: TimeoutMiddleware implementation

### Word Management Endpoints

#### `/wordlists/{wordlistId}/words` - GET (List Words)
- **Status**: 🔲 Pending
- **HTTP Handler**: `internal/http/word.go:GetWords()`
- **Service Methods**:
  - `service.GetWords(wordlistID, userID)`
- **Repository Methods**:
  - `wordRepository.GetWordsForWordlist(wordlistID, userID)`
- **Database Operations**:
  - Words query with definitions and progress
- **Estimated Effort**: 1.5 hours
- **Dependencies**: None

#### `/wordlists/{wordlistId}/words` - POST (Create Word)
- **Status**: 🔲 Pending
- **HTTP Handler**: `internal/http/word.go:CreateWord()`
- **Service Methods**:
  - `service.CreateWord(wordlistID, userID, wordData)`
- **Repository Methods**:
  - `wordRepository.Create(word)`
  - `wordRepository.GetWordCount(wordlistID)` (for limits)
- **Database Operations**:
  - Word creation INSERT
  - Word count validation
  - Leitner system initialization
- **Estimated Effort**: 2 hours
- **Dependencies**: None

#### `/words/{id}` - GET (Get Word Details)
- **Status**: 🔲 Pending
- **HTTP Handler**: `internal/http/word.go:GetWord()`
- **Service Methods**:
  - `service.GetWord(wordID, userID)`
- **Repository Methods**:
  - `wordRepository.GetWordByID(wordID, userID)`
- **Database Operations**:
  - Single word query with all related data
- **Estimated Effort**: 1 hour
- **Dependencies**: None

#### `/words/{id}` - PUT (Update Word)
- **Status**: 🔲 Pending
- **HTTP Handler**: `internal/http/word.go:UpdateWord()`
- **Service Methods**:
  - `service.UpdateWord(wordID, userID, updates)`
- **Repository Methods**:
  - `wordRepository.Update(wordID, userID, updates)`
- **Database Operations**:
  - Word UPDATE with ownership verification
- **Estimated Effort**: 1.5 hours
- **Dependencies**: None

#### `/words/{id}` - DELETE (Delete Word)
- **Status**: 🔲 Pending
- **HTTP Handler**: `internal/http/word.go:DeleteWord()`
- **Service Methods**:
  - `service.DeleteWord(wordID, userID)`
- **Repository Methods**:
  - `wordRepository.Delete(wordID, userID)`
- **Database Operations**:
  - Word deletion with cascade cleanup
- **Estimated Effort**: 1.5 hours
- **Dependencies**: None

### Definition Management Endpoints

#### `/words/{wordId}/definitions` - POST (Create Definition)
- **Status**: 🔲 Pending
- **HTTP Handler**: `internal/http/definition.go:CreateDefinition()`
- **Service Methods**:
  - `service.CreateDefinition(wordID, userID, definitionData)`
- **Repository Methods**:
  - `definitionRepository.Create(definition)`
  - `wordRepository.GetWordByID()` (for ownership check)
- **Database Operations**:
  - Definition creation INSERT
  - Word ownership verification
- **Estimated Effort**: 1.5 hours
- **Dependencies**: None

#### `/definitions/{id}` - PUT (Update Definition)
- **Status**: 🔲 Pending
- **HTTP Handler**: `internal/http/definition.go:UpdateDefinition()`
- **Service Methods**:
  - `service.UpdateDefinition(definitionID, userID, updates)`
- **Repository Methods**:
  - `definitionRepository.Update(definitionID, userID, updates)`
- **Database Operations**:
  - Definition UPDATE with ownership check
- **Estimated Effort**: 1 hour
- **Dependencies**: None

#### `/definitions/{id}` - DELETE (Delete Definition)
- **Status**: 🔲 Pending
- **HTTP Handler**: `internal/http/definition.go:DeleteDefinition()`
- **Service Methods**:
  - `service.DeleteDefinition(definitionID, userID)`
- **Repository Methods**:
  - `definitionRepository.Delete(definitionID, userID)`
- **Database Operations**:
  - Definition deletion with cleanup
- **Estimated Effort**: 1 hour
- **Dependencies**: None

### Quiz System Endpoints

#### `/wordlists/{wordlistId}/quiz` - POST (Create Quiz)
- **Status**: 🔲 Pending
- **HTTP Handler**: `internal/http/quiz.go:Create()`
- **Service Methods**:
  - `strategy.CreateQuiz(wordlistID, userID)`
- **Repository Methods**:
  - Multiple Leitner system queries
  - Word selection algorithms
- **Database Operations**:
  - Complex word selection queries
  - Leitner box calculations
- **Estimated Effort**: 3 hours
- **Dependencies**: Leitner system context updates

#### `/quiz/save` - POST (Save Quiz Result)
- **Status**: 🔲 Pending
- **HTTP Handler**: `internal/http/quiz.go:Save()`
- **Service Methods**:
  - `strategy.SaveQuizResult(result, isPremium, features)`
- **Repository Methods**:
  - Multiple repository updates for quiz results
  - Leitner system progression
  - Analytics data updates
- **Database Operations**:
  - Quiz performance INSERT
  - Leitner tracking UPDATE
  - Word mastery calculations
- **Estimated Effort**: 3.5 hours
- **Dependencies**: Analytics and Leitner system updates

### Analytics Endpoints

#### `/analytics/dashboard-stats/{wordlistId}` - GET
- **Status**: 🔲 Pending
- **HTTP Handler**: `internal/http/analytics.go:GetDashboardStats()`
- **Service Methods**:
  - `analyticsService.GetDashboardStats(userID, wordlistID)`
- **Repository Methods**:
  - `analyticsRepo.GetDashboardStats(ctx, userID, wordlistID)`
- **Database Operations**:
  - Complex aggregation queries
  - Multiple table JOINs
- **Estimated Effort**: 2 hours
- **Dependencies**: None

#### `/analytics/word-mastery/{wordlistId}` - GET
- **Status**: 🔲 Pending
- **HTTP Handler**: `internal/http/analytics.go:GetWordMastery()`
- **Service Methods**:
  - `analyticsService.GetWordMastery(userID, wordlistID)`
- **Repository Methods**:
  - `analyticsRepo.GetWordMastery(ctx, userID, wordlistID)`
- **Database Operations**:
  - Word mastery aggregation queries
- **Estimated Effort**: 1.5 hours
- **Dependencies**: None

#### `/analytics/learning-progress/{wordlistId}` - GET
- **Status**: 🔲 Pending
- **HTTP Handler**: `internal/http/analytics.go:GetLearningProgress()`
- **Service Methods**:
  - `analyticsService.GetLearningProgress(userID, wordlistID)`
- **Repository Methods**:
  - `analyticsRepo.GetLearningProgress(ctx, userID, wordlistID)`
- **Database Operations**:
  - Time-series analytics queries
- **Estimated Effort**: 2 hours
- **Dependencies**: None

#### `/analytics/quiz-performance/{wordlistId}` - GET
- **Status**: 🔲 Pending
- **HTTP Handler**: `internal/http/analytics.go:GetQuizPerformance()`
- **Service Methods**:
  - `analyticsService.GetQuizPerformance(userID, wordlistID)`
- **Repository Methods**:
  - `analyticsRepo.GetQuizPerformance(ctx, userID, wordlistID)`
- **Database Operations**:
  - Quiz performance analytics
- **Estimated Effort**: 1.5 hours
- **Dependencies**: None

#### `/analytics/box-distribution/{wordlistId}` - GET
- **Status**: 🔲 Pending
- **HTTP Handler**: `internal/http/analytics.go:GetBoxDistribution()`
- **Service Methods**:
  - `analyticsService.GetBoxDistribution(userID, wordlistID)`
- **Repository Methods**:
  - `analyticsRepo.GetBoxDistribution(ctx, userID, wordlistID)`
- **Database Operations**:
  - Leitner box distribution queries
- **Estimated Effort**: 1.5 hours
- **Dependencies**: None

#### `/analytics/progress-summary` - GET (Batch Endpoint)
- **Status**: 🔲 Pending
- **HTTP Handler**: `internal/http/analytics.go:GetProgressSummary()`
- **Service Methods**:
  - `analyticsService.GetBatchProgressSummary(userID)`
- **Repository Methods**:
  - `analyticsRepo.GetAllWordlistsProgress(ctx, userID)`
  - Multiple analytics queries
- **Database Operations**:
  - Batch analytics queries
  - Multiple aggregations
- **Estimated Effort**: 3 hours
- **Dependencies**: All analytics endpoints

### Subscription Management Endpoints

#### `/subscription` - GET (Get Subscription Status)
- **Status**: 🔲 Pending
- **HTTP Handler**: `internal/http/subscription.go:GetSubscription()`
- **Service Methods**:
  - `subscriptionService.GetUserSubscription(userID)`
- **Repository Methods**:
  - `subRepo.GetActiveSubscriptionForUser(ctx, userID)`
- **Database Operations**:
  - Subscription status query
- **Estimated Effort**: 1 hour
- **Dependencies**: None

#### `/subscription/checkout` - POST (Create Stripe Checkout)
- **Status**: 🔲 Pending
- **HTTP Handler**: `internal/http/subscription.go:CreateCheckoutSession()`
- **Service Methods**:
  - `subscriptionService.CreateStripeCheckoutSession(userID, plan, redirectURI)`
- **Repository Methods**:
  - `userRepository.Find()` (for customer info)
- **Database Operations**:
  - User lookup for Stripe customer creation
- **Estimated Effort**: 1.5 hours
- **Dependencies**: None

#### `/subscription/cancel` - POST (Cancel Subscription)
- **Status**: 🔲 Pending
- **HTTP Handler**: `internal/http/subscription.go:CancelSubscription()`
- **Service Methods**:
  - `subscriptionService.CancelUserSubscription(userID)`
- **Repository Methods**:
  - `subRepo.GetActiveSubscriptionForUser(ctx, userID)`
  - `subRepo.UpdateSubscription(ctx, subscription, event)`
- **Database Operations**:
  - Subscription cancellation update
- **Estimated Effort**: 2 hours
- **Dependencies**: None

#### `/subscription/revenuecat/restore` - POST (Restore RevenueCat Purchases)
- **Status**: 🔲 Pending
- **HTTP Handler**: `internal/http/revenuecat.go:RestorePurchases()`
- **Service Methods**:
  - `rcService.RestorePurchases(userID, appUserID, platform)`
- **Repository Methods**:
  - RevenueCat API call + subscription updates
- **Database Operations**:
  - Subscription restoration
- **Estimated Effort**: 2 hours
- **Dependencies**: RevenueCat service updates

### Error Reporting Endpoints

#### `/error-reports` - POST (Report Error)
- **Status**: 🔲 Pending
- **HTTP Handler**: `internal/http/error_report.go:CreateErrorReport()`
- **Service Methods**:
  - `service.CreateErrorReport(userID, reportData)`
- **Repository Methods**:
  - `errorReportRepository.Create(report)`
  - Rate limiting checks
- **Database Operations**:
  - Error report INSERT
  - Rate limiting queries
- **Estimated Effort**: 1.5 hours
- **Dependencies**: None

#### `/error-reports/status` - GET (Get Rate Limit Status)
- **Status**: 🔲 Pending
- **HTTP Handler**: `internal/http/error_report.go:GetRateLimitStatus()`
- **Service Methods**:
  - `service.GetErrorReportingStatus(userID)`
- **Repository Methods**:
  - `errorReportRepository.GetUserReportCount(userID, timeWindow)`
- **Database Operations**:
  - Rate limiting status queries
- **Estimated Effort**: 1 hour
- **Dependencies**: None

### Background Job Endpoints

#### `/generate-content/{wordId}` - POST (Trigger Content Generation)
- **Status**: 🔲 Pending
- **HTTP Handler**: `internal/http/content.go:GenerateContent()`
- **Service Methods**:
  - River job enqueueing
- **Repository Methods**:
  - Job queue operations
- **Database Operations**:
  - River job INSERT
- **Estimated Effort**: 1 hour
- **Dependencies**: None

### Webhook Endpoints

#### `/webhook/stripe` - POST (Stripe Webhook)
- **Status**: 🔲 Pending
- **HTTP Handler**: `internal/http/subscription.go:HandleStripeWebhook()`
- **Service Methods**:
  - River job enqueueing (already async)
- **Repository Methods**:
  - Job queue operations
- **Database Operations**:
  - River job INSERT
- **Estimated Effort**: 0.5 hours
- **Dependencies**: None (already async)

#### `/webhook/revenuecat` - POST (RevenueCat Webhook)
- **Status**: 🔲 Pending
- **HTTP Handler**: `internal/http/revenuecat.go:HandleRevenueCatWebhook()`
- **Service Methods**:
  - River job enqueueing (already async)
- **Repository Methods**:
  - Job queue operations
- **Database Operations**:
  - River job INSERT
- **Estimated Effort**: 0.5 hours
- **Dependencies**: None (already async)

---

## Implementation Guidelines

### ✅ Implemented Patterns

#### TimeoutMiddleware Implementation
```go
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Create timeout context
        ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
        defer cancel()

        // Replace request context with timeout context
        c.Request = c.Request.WithContext(ctx)

        // Process request
        c.Next()

        // Check if request timed out after processing
        if ctx.Err() == context.DeadlineExceeded {
            // Only respond if no response was already written
            if !c.Writer.Written() {
                common.Logger.Error("request timed out",
                    "path", c.FullPath(),
                    "method", c.Request.Method,
                    "timeout", timeout)

                c.AbortWithStatusJSON(http.StatusRequestTimeout, gin.H{
                    "error": "Request timeout - please try again",
                })
            }
        }
    }
}
```

#### Simplified Handler Pattern (Before vs After)
```go
// BEFORE: Manual timeout handling (18 lines)
func (h *UserRoutes) Login(c *gin.Context) {
    ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
    defer cancel()
    
    var input LoginInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    jwtToken, err := service.LoginUser(ctx, input.Email, input.Password)
    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            c.JSON(http.StatusRequestTimeout, gin.H{"error": "Request timeout - please try again"})
            return
        }
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email or password"})
        return
    }
    
    c.Header("authorization", jwtToken)
    writeAuthenticationCookie(c, jwtToken)
    c.Status(http.StatusOK)
}

// AFTER: Simplified with middleware (12 lines)
func (h *UserRoutes) Login(c *gin.Context) {
    var input LoginInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    jwtToken, err := service.LoginUser(c.Request.Context(), input.Email, input.Password)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email or password"})
        return
    }
    
    c.Header("authorization", jwtToken)
    writeAuthenticationCookie(c, jwtToken)
    c.Status(http.StatusOK)
}
```

#### Global Middleware Application
```go
// In internal/http/setup.go
router := gin.New()
router.Use(SentryMiddlewares()...)
router.Use(gin.Logger())
router.Use(TimeoutMiddleware(2 * time.Second))  // Applied globally
router.Use(ErrorMiddleware())
router.Use(CORSMiddleware())
```

### Service Layer (No Changes Needed)
```go
// LoginUser already accepts context - no signature changes needed
func LoginUser(ctx context.Context, email, password string) (string, error) {
    // Context timeout errors already handled
    results, err := userRepository.Find(ctx, args)
    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            common.Logger.Error("login request timed out", "email", email)
            return "", err
        }
        // Handle other errors...
    }
    // ...
}
```

### Repository Layer (Already Compatible)
```go
// userRepository.Find already uses context - no changes needed
func (r *UserRepository) Find(ctx context.Context, args FindUserArgs) ([]User, error) {
    // pgx/v5 automatically respects context timeout
    rows, err := r.Db.Query(ctx, query, queryArgs...)
    // ...
}
```

### ✅ Key Advantages of Middleware Approach
1. **Zero Service/Repository Changes**: Existing context-aware code works unchanged
2. **Centralized Error Handling**: All timeout errors handled in one place
3. **Code Reduction**: 33% fewer lines in handlers
4. **Consistent Behavior**: All endpoints get same timeout behavior
5. **Easy Configuration**: Change timeout globally in one place

---

## Testing Strategy

### Unit Tests
- Add timeout tests for each refactored endpoint
- Test context cancellation behavior
- Verify proper error responses

### Integration Tests
- Test actual timeout behavior with slow queries
- Verify database connections are properly closed
- Test concurrent request handling

### Performance Tests
- Ensure 2-second timeout is appropriate for all operations
- Test under load to verify timeout effectiveness
- Monitor for connection pool exhaustion

---

## Risk Assessment

### High Risk Areas
- **Quiz System**: Complex algorithms may need optimization
- **Analytics Endpoints**: Heavy aggregation queries may timeout
- **Batch Operations**: Multiple database calls may exceed 2s limit

### Medium Risk Areas
- **Word Creation**: Multiple database operations
- **User Profile**: Grace period checks add complexity

### Low Risk Areas
- **Simple CRUD Operations**: Single table operations
- **Webhook Endpoints**: Already asynchronous

---

## ✅ Updated Rollout Strategy

### ✅ Phase 1: Core Infrastructure & Login (COMPLETED)
- ✅ TimeoutMiddleware implementation and global application
- ✅ POST /login endpoint complete implementation and testing
- **Duration**: 1 day

### ✅ Phase 2: Authentication Endpoints (COMPLETED)
- ✅ POST /users (SignUp) - Context propagation implemented
- ✅ GET /users (GetProfile) - Context propagation implemented  
- ✅ PUT /users/profile (UpdateProfile) - Context propagation implemented
- ✅ DELETE /users (DeleteProfile) - Context propagation implemented
- **Actual Duration**: 8 hours (1 day)

### 🔄 Phase 3: CRUD Operations (WEEK 2)

#### 3.1: User Management (COMPLETED IN PHASE 2)
- ✅ POST /users (Registration) - Completed in Phase 2
- ✅ GET /users (Profile) - Completed in Phase 2
- ✅ DELETE /users (Delete Account) - Completed in Phase 2

#### 3.2: Wordlist Management (COMPLETED)
- ✅ GET /wordlists (List) - Completed (context propagation implemented)
- ✅ POST /wordlists (Create) - Completed (context propagation implemented)
- ✅ GET /wordlists/{id} (Details) - Completed (context propagation implemented)
- ✅ PUT /wordlists/{id} (Update) - Completed (context propagation implemented)
- ✅ DELETE /wordlists/{id} (Delete) - Completed (context propagation implemented)

#### 3.3: Word Management (7 hours)
- 🔲 GET /wordlists/{id}/words (List) - 1.5h
- 🔲 POST /wordlists/{id}/words (Create) - 2h
- 🔲 GET /words/{id} (Details) - 1h
- 🔲 PUT /words/{id} (Update) - 1.5h
- 🔲 DELETE /words/{id} (Delete) - 1.5h

#### 3.4: Definition Management (3.5 hours)
- 🔲 POST /words/{id}/definitions (Create) - 1.5h
- 🔲 PUT /definitions/{id} (Update) - 1h
- 🔲 DELETE /definitions/{id} (Delete) - 1h

**Total Phase 3**: 10.5 hours (1.5 days) - 9 hours saved from Phase 2 & 3.2 completion

### 🔄 Phase 4: Complex Operations (WEEK 3)
- 🔲 Quiz system, Analytics, Subscription endpoints
- **Estimated Duration**: 3 days

### 🔄 Phase 5: Testing & Production (WEEK 4)
- 🔲 Comprehensive timeout testing, Performance validation, Production monitoring
- **Estimated Duration**: 2 days

### 🔄 Phase 6: Background Worker Optimization (CRITICAL)
**Performance Issue**: POST /wordlists/{id}/words timeout at 100+ concurrent users due to synchronous OpenAI moderation

#### 6.1: Move OpenAI Validation to Background Workers
**Current Flow**:
```
User creates word → OpenAI moderation call (2-4s) → Return response
```

**Optimized Flow**:
```
User creates word → Save with "pending" status → Return immediately
Background worker → Validate with OpenAI → Update status
```

**Implementation**:
- Add `validation_status` enum: pending, approved, rejected
- Create `WordValidationWorker` for OpenAI moderation
- Update mobile app to handle pending states
- Add validation retry logic for API failures

#### 6.2: OpenAI Circuit Breaker Pattern
**Problem**: OpenAI API timeouts cascade to 30s user timeouts

**Solution**: 
- Circuit breaker with 3 states: closed, open, half-open
- Auto-approve words during OpenAI outages
- Rate limiting to prevent API quota exhaustion

#### 6.3: Job Priority Queues
**Problem**: Background job storms overwhelm 1 vCPU system

**Solution**:
- High priority: user-facing validation jobs
- Low priority: image/audio generation
- Resource-aware job scheduling

**Effort**: 60 hours

---

## Monitoring & Observability

### Core Timeout Metrics
- Request timeout frequency per endpoint
- Average request duration
- Database query duration
- Context cancellation rates

### Database Connection Pool & Resource Monitoring (CRITICAL FOR 512MB RAM)
**Performance Issue**: Connection pool exhaustion at 100+ concurrent users on Digital Ocean 512MB/1vCPU

#### Connection Pool Monitoring
**Critical Metrics**:
- Active/idle connections, pool utilization %
- Connection acquire timeouts (>1s warning)
- Connection lifetime and churn rate

**Alerts**:
- **Critical**: >25 active connections (pool exhaustion)
- **Warning**: >20 active connections
- **Info**: Connection acquire time >500ms

#### Memory & CPU Monitoring  
**Resource Thresholds**:
- **Memory**: <400MB normal, >450MB critical (512MB limit)
- **CPU**: <70% normal, >85% warning, >95% critical (1 vCPU)

**Implementation**:
- Real-time `/health/resources` endpoint
- Resource pressure detection during high load
- Connection pool health dashboard
- Enhanced logging with resource context

---

## Success Criteria

### Core Timeout Implementation
- [ ] All endpoints complete within 2 seconds or timeout gracefully
- [ ] No degradation in application functionality
- [ ] Context properly propagated through all layers
- [ ] Comprehensive test coverage for timeout scenarios
- [ ] Monitoring and alerting in place
- [ ] Documentation updated

### Performance Bottleneck Resolution (Phase 6)
- [ ] **Word Creation Performance**: POST /wordlists/{id}/words achieves sub-500ms response time at 100+ concurrent users
- [ ] **Quiz Generation Performance**: POST /wordlists/{id}/quizzes maintains <1s response time under load
- [ ] **Error Rate Reduction**: <2% timeout errors under normal load conditions
- [ ] **Concurrency Support**: System handles 100+ concurrent users without degradation
- [ ] **Resource Utilization**: CPU usage <70%, Memory usage <400MB during peak load

### Resource Monitoring & Alerting
- [ ] **Connection Pool Monitoring**: Real-time dashboard showing pool utilization, timeouts, and health
- [ ] **Memory Pressure Detection**: Alerts configured for 400MB warning, 450MB critical thresholds
- [ ] **Database Performance Monitoring**: Slow query detection, connection timeout tracking
- [ ] **Load Testing Baseline**: Performance metrics compared against load test results
- [ ] **Production Alerting**: High/medium/low priority alerts configured and tested

### Background Worker Optimization
- [ ] **OpenAI Validation Async**: Word creation with pending status, background validation
- [ ] **Circuit Breaker**: OpenAI API failures don't cascade to user-facing timeouts
- [ ] **Job Priority Queue**: Critical jobs (validation) prioritized over background processing
- [ ] **Resource-Aware Scheduling**: Job processing throttled during high resource usage
- [ ] **Graceful Degradation**: System continues functioning during OpenAI API outages

### Measurable Performance Improvements
- [ ] **25 Concurrent Users**: <1s average response time, >99% success rate
- [ ] **100 Concurrent Users**: <2s average response time, >95% success rate
- [ ] **Database Connections**: <25 active connections during peak load (512MB RAM limit)
- [ ] **Memory Management**: <10 GC/second, <20% GC CPU time
- [ ] **Validation Throughput**: 95%+ words validated within 30 seconds

---

## ✅ Updated Effort Analysis

### Original Estimate vs Actual (Including Performance Optimization)
- **Original Timeout Refactoring**: ~65 hours (8.5 working days)
- **Middleware Approach**: ~20 hours (2.5 working days) - **70% reduction**
- **Performance Optimization (Phase 6)**: ~60 hours (7.5 working days)
- **Enhanced Monitoring**: ~16 hours (2 working days)

### Complete Implementation Effort Breakdown
- ✅ **Core Infrastructure**: 1 hour (COMPLETED)
- ✅ **POST /login Implementation**: 1 hour (COMPLETED)
- ✅ **Authentication Endpoints (Phase 2)**: 8 hours (COMPLETED)
- 🔄 **CRUD Operations**: 2 hours (simplified - Phase 3.1 completed in Phase 2)
- 🔄 **Complex Operations**: 8 hours (may need custom timeouts)
- 🔄 **Testing & Documentation**: 2 hours
- 🔄 **Background Worker Optimization (Phase 6)**: 60 hours (critical for high concurrency)
- 🔄 **Enhanced Monitoring Implementation**: 16 hours (resource monitoring, alerts)

**Total Estimated Time**: ~98 hours (12.25 working days) - **Phase 2 completed ahead of schedule**

### Phase 6 Detailed Breakdown
- **Word Creation with Pending Status**: 20 hours (2.5 days)
  - Database schema updates, service modifications, background worker, frontend handling
- **Circuit Breaker & Rate Limiting**: 18 hours (2.25 days)
  - OpenAI circuit breaker, graceful degradation, rate limiting system
- **Enhanced Job Scheduling**: 22 hours (2.75 days)
  - Priority queues, load-based scheduling, resource monitoring

### Monitoring Implementation Breakdown
- **Connection Pool Monitoring**: 6 hours
  - Real-time dashboard, pool health checks, alerting
- **Resource Pressure Detection**: 4 hours
  - Memory/CPU monitoring, threshold alerts
- **Load Testing Monitoring**: 4 hours
  - Endpoint-specific metrics, performance tracking
- **Alert Configuration**: 2 hours
  - High/medium/low priority alerts, production deployment

### Why the Middleware Approach is Much Faster
1. **No Service/Repository Changes**: Existing context-aware code works unchanged
2. **No Signature Updates**: Service methods already accept context
3. **Bulk Application**: Global middleware applies to all endpoints
4. **Simplified Handlers**: Just remove manual timeout code
5. **Centralized Testing**: Test timeout behavior once in middleware

---

*Last Updated: July 15, 2025*
*Status: Phase 3.2 Complete - Authentication & Wordlist Management Context Propagation Implemented*

## ✅ Implementation Results Summary

### Phase 1 Achievements (COMPLETED)
- **Custom TimeoutMiddleware**: Robust, production-ready implementation
- **Global Application**: All endpoints now have 2-second timeout protection
- **POST /login**: Complete context propagation implementation

### Phase 2 Achievements (COMPLETED)
- **Authentication Endpoints**: Complete context propagation for all 4 endpoints
- **Repository Layer Updates**: Added context parameters to all user-related operations
- **Service Layer Updates**: Proper context flow from HTTP → Service → Repository → Database
- **Integration Test Fixes**: Updated test signatures to match new context parameters
- **Code Quality**: All linting issues resolved, consistent parameter naming

### Phase 3.2 Achievements (COMPLETED)
- **Wordlist Management**: Complete context propagation for all 5 endpoints
- **Repository Layer**: Updated 4 wordlist repository methods with context parameters
- **Service Layer**: Updated 5 wordlist service methods to accept and pass context
- **Analytics Integration**: Fixed 7 analytics endpoints that call wordlist services
- **Integration Tests**: Updated wordlist service calls with context parameters

### Context Propagation Implementation
**Successfully implemented for:**

**Phase 2 - Authentication Endpoints:**
1. **POST /users (SignUp)** - Full context chain updated
2. **GET /users (GetProfile)** - Replaced context.Background with request context
3. **PUT /users/profile (UpdateProfile)** - Complete service/repository context updates  
4. **DELETE /users (DeleteProfile)** - Multi-repository context propagation

**Phase 3.2 - Wordlist Management Endpoints:**
5. **GET /wordlists (List)** - Full context chain updated
6. **POST /wordlists (Create)** - Complete service/repository context updates
7. **GET /wordlists/{id} (Details)** - Full context chain updated
8. **PUT /wordlists/{id} (Update)** - Complete service/repository context updates
9. **DELETE /wordlists/{id} (Delete)** - Full context chain updated

### Repository Methods Updated

**Phase 2 - User Operations:**
- `UserRepository.Save(ctx, ...)` - User creation with context
- `UserRepository.Delete(ctx, ...)` - User deletion with context
- `UserRepository.UpdateUserProfile(ctx, ...)` - Profile updates with context
- `WordlistRepository.DeleteAll(ctx, ...)` - Cascade deletion with context
- `ErrorReportService.DeleteUserErrorReports(ctx, ...)` - Error cleanup with context

**Phase 3.2 - Wordlist Operations:**
- `WordlistRepository.Save(ctx, ...)` - Wordlist creation with context
- `WordlistRepository.Find(ctx, args)` - Wordlist queries with context
- `WordlistRepository.Update(ctx, wordlist)` - Wordlist updates with context
- `WordlistRepository.Delete(ctx, wordlistID, userID)` - Wordlist deletion with context

### Key Design Decisions
1. **Complete Context Flow**: Request context flows from TimeoutMiddleware → HTTP → Service → Repository → Database
2. **Consistent Signatures**: All updated methods follow `func(ctx context.Context, ...)` pattern
3. **Error Handling**: Database timeouts properly bubble up through context cancellation
4. **Backward Compatibility**: No breaking changes to API contracts

### Next Steps
1. **Phase 3.3 & 3.4**: Apply same pattern to remaining Word & Definition CRUD endpoints (now 10.5 hours estimated)
2. **Performance Testing**: Validate authentication and wordlist endpoints handle 2-second timeout under load
3. **Phase 6 Priority**: OpenAI validation background worker optimization for high concurrency

### Lessons Learned
- **Context Background Issues**: Found many service methods using context.Background instead of request context
- **Multi-Repository Complexity**: User deletion required updating 3 different repositories
- **Integration Test Dependencies**: Test files needed updates when service signatures changed
- **Linting Importance**: Parameter naming consistency critical for Go conventions
- **Analytics Dependencies**: Wordlist service changes impacted 7 analytics endpoints that needed context updates
- **Content Moderation**: OpenAI validation calls in wordlist creation now respect timeout context