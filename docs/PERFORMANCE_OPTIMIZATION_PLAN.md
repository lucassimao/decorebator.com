# Performance Optimization Plan for Decorebator API

## Executive Summary

Based on load testing results showing critical performance issues:
- **POST /users**: 7.3s avg (UNACCEPTABLE)
- **POST /login**: 3.0s avg (UNACCEPTABLE)  
- **POST /wordlists**: 1.2s avg (POOR)

This plan provides a systematic approach to optimize these endpoints through database indexing, bcrypt optimization, architectural improvements, and monitoring.

## Current Implementation Analysis

### Critical Performance Bottlenecks Identified

#### 1. User Registration (`POST /users`) - 7.3s avg
**Location**: `internal/http/user.go:82-133`

**Issues Found**:
- **Immediate Login After Registration**: Lines 115-120 perform a full login flow immediately after user creation
- **Synchronous bcrypt Operations**: `bcrypt.DefaultCost` (10) used for both registration and immediate login
- **Double Password Hashing**: Registration hashes password, then login validates it immediately
- **Async Email Operations**: Lines 125-130 trigger background email sending (good practice)

**Root Causes**:
```go
// user.go:115-120 - Immediate login after registration
jwtToken, err := service.LoginUser(input.Email, input.Password)
```

#### 2. User Login (`POST /login`) - 3.0s avg
**Location**: `internal/http/user.go:135-152`

**Issues Found**:
- **Expensive bcrypt Operations**: `bcrypt.CompareHashAndPassword` at default cost (10)
- **Database User Lookup**: Email-based user search without optimized indexes
- **JWT Generation**: Synchronous JWT creation with environment lookups

**Root Causes**:
```go
// service/user.go:116 - Expensive bcrypt comparison
err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
```

#### 3. Wordlist Creation (`POST /wordlists`) - 1.2s avg
**Location**: `internal/http/wordlist.go:45-95`

**Issues Found**:
- **OpenAI Moderation API Calls**: Lines 57-73 synchronous validation
- **Pronunciation System Validation**: Complex logic for supported systems
- **Database Operations**: No apparent connection pooling optimization

**Root Causes**:
```go
// wordlist.go:57-73 - Synchronous moderation validation
supportedSystems := model.GetSupportedPronunciationSystems(input.LanguageCode)
```

#### 4. Background Job Overhead
**Location**: `internal/service/word.go:87-176`

**Issues Found**:
- **Heavy Processing in SaveWord**: Definition reuse logic, worker triggering
- **Transaction Management**: Manual transaction handling with defer cleanup
- **Multiple Background Jobs**: TriggerFetchDefinitionWorker, TriggerTextToSpeechWorker

#### 5. Architecture Anti-Patterns
**Location**: `internal/service/user.go:60-68`

**Critical Issues**:
```go
// Global state with os.Exit(1) - CRITICAL PROBLEM
func init() {
    db, err := common.GetDBConnection()
    if err != nil {
        common.Logger.Error("failed to open db connection", "error", err)
        os.Exit(1) // Breaks testing and deployment
    }
    userRepository = &repo.UserRepository{Db: db}
    wordlistRepository = &repo.WordlistRepository{Db: db}
}
```

## Phase 1: Immediate Fixes (1-2 weeks)

### 1.1 Database Indexing Strategy

#### Critical Indexes to Add

**User Lookup Optimization**:
```sql
-- For email-based login queries
CREATE INDEX CONCURRENTLY idx_users_email_lower ON users (LOWER(email));

-- For user ID lookups
CREATE INDEX CONCURRENTLY idx_users_id_active ON users (id) WHERE deleted_at IS NULL;

-- For subscription queries
CREATE INDEX CONCURRENTLY idx_users_subscription ON users (subscription_status, subscription_plan);
```

**Wordlist Query Optimization**:
```sql
-- For user wordlist queries
CREATE INDEX CONCURRENTLY idx_wordlists_user_id ON wordlists (user_id, created_at DESC);

-- For word count aggregations
CREATE INDEX CONCURRENTLY idx_words_wordlist_learned ON words (wordlist_id, learned, user_id);
```

**Implementation File**: `migrations/000xxx_add_performance_indexes.up.sql`

### 1.2 bcrypt Cost Optimization

#### Environment-Specific Configuration
**Target Files**: 
- `internal/repository/user.go:42`
- `internal/repository/user.go:150`
- `internal/repository/user.go:200`

**Implementation**:
```go
// internal/common/crypto.go (new file)
func GetBcryptCost() int {
    if os.Getenv("ENV") == "production" {
        return 12 // Higher security for production
    }
    return 8 // Faster for development/testing
}

// Update all bcrypt.GenerateFromPassword calls
bytes, err := bcrypt.GenerateFromPassword([]byte(password), common.GetBcryptCost())
```

### 1.3 Registration Flow Optimization

#### Eliminate Double Authentication
**Target File**: `internal/http/user.go:115-120`

**Current Flow**:
```
1. User registration → bcrypt hash (10 rounds)
2. Immediate login → bcrypt compare (10 rounds)  ← ELIMINATE THIS
3. JWT generation
```

**Optimized Flow**:
```go
// After successful user creation, generate JWT directly
user, err := service.SaveUser(input.FirstName, input.LastName, input.Password, input.Email, country)
if err != nil {
    // handle error
    return
}

// Generate JWT directly from user object - no re-authentication needed
// OLD: jwtToken, err := service.LoginUser(input.Email, input.Password)
// NEW: jwtToken, err := service.GenerateJWT(*user)
jwtToken, err := service.GenerateJWT(*user)
```

**Expected Impact**: ~50% reduction in registration time (eliminate one bcrypt operation)

**✅ IMPLEMENTED (January 2025)**: Registration double-bcrypt fix completed
- **Before**: ~1400ms+ (double bcrypt operations)
- **After**: ~670-820ms (single bcrypt operation)
- **Actual improvement**: ~40-50% faster registration
- **Files changed**: `api/internal/http/user.go:115-120`
- **Verification**: All integration tests pass, JWT tokens contain correct subscription_plan = 'free'

### 1.4 JWT Generation Optimization

#### Cache Environment Variables and JWT Key
**Target File**: `internal/service/user.go:35-58`

**Current Issues**:
- Environment lookups in every JWT generation (`os.Getenv("ENV")`, `os.Getenv("JWT_KEY")`)
- JWT key parsed from string on every token creation

**Optimization**:
```go
// Add to service/user.go
var (
    jwtKey []byte
    jwtEnv string
    jwtOnce sync.Once
)

func initJWTConfig() {
    jwtKey = []byte(os.Getenv("JWT_KEY"))
    jwtEnv = os.Getenv("ENV")
}

func GenerateJWT(user User) (string, error) {
    jwtOnce.Do(initJWTConfig)
    
    claims := &Claims{
        Email:            user.Email,
        Environment:      jwtEnv,  // Use cached value
        SubscriptionPlan: user.SubscriptionPlan,
        StandardClaims: jwt.StandardClaims{
            Issuer:    "Decorebator",
            ExpiresAt: time.Now().Add(AUTH_TOKEN_DURATION).Unix(),
            Subject:   fmt.Sprint(user.ID),
            IssuedAt:  time.Now().Unix(),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(jwtKey)  // Use cached key
}
```

**Expected Impact**: 10-20% faster JWT generation

### 1.5 Request Timeout Context

#### Add Timeouts to External API Calls
**Target Files**: 
- `internal/service/openai_moderation.go`
- `internal/openai/*.go`

**Current Issues**:
- No timeouts on OpenAI API calls
- Requests can hang indefinitely
- No context cancellation support

**Optimization**:
```go
// Add to all external API calls
func (s *OpenAIModerationService) Validate(content string) ModerationResult {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    // Use ctx in HTTP requests
    req, err := http.NewRequestWithContext(ctx, "POST", url, body)
    // ... rest of implementation
}
```

**Expected Impact**: Prevent hanging requests, fail fast instead of 30+ second timeouts

### 1.6 Background Job Optimization

#### Defer Heavy Operations
**Target File**: `internal/service/word.go:138-169`

**Current Issues**:
- Synchronous definition lookup
- Immediate worker triggering
- Heavy transaction scope

**Optimization**:
```go
// Defer definition processing to background
func (ws *WordService) SaveWordAsync(ctx context.Context, dto *Word) (*Word, error) {
    // Only save basic word data synchronously
    word, err := ws.repository.Save(trimmedName, dto.Notes, dto.UserID, dto.WordlistID, tx)
    if err != nil {
        return nil, err
    }
    
    // Defer heavy processing to background job
    go func() {
        ws.processWordDefinitionsAsync(word.ID, trimmedName)
    }()
    
    return word, nil
}
```

## Phase 2: Architectural Improvements (2-3 weeks)

### 2.1 Fix Global State Anti-Patterns

#### Remove os.Exit(1) from init Functions
**Target Files**: 
- `internal/service/user.go:60-68`
- `internal/service/wordlist.go:32-38`

**Current Problem**:
```go
func init() {
    db, err := common.GetDBConnection()
    if err != nil {
        common.Logger.Error("failed to open db connection", "error", err)
        os.Exit(1) // BREAKS TESTING
    }
    userRepository = &repo.UserRepository{Db: db}
}
```

**Solution - Constructor-Based DI**:
```go
type UserService struct {
    repository *repo.UserRepository
}

func NewUserService(db *pgxpool.Pool) *UserService {
    return &UserService{
        repository: &repo.UserRepository{Db: db},
    }
}
```

### 2.2 Database Connection Optimization

#### Implement Proper Connection Pooling
**Target File**: `internal/common/database.go`

**Configuration**:
```go
func GetOptimizedDBConnection() (*pgxpool.Pool, error) {
    config, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
    if err != nil {
        return nil, err
    }
    
    // Optimize connection pool settings
    config.MaxConns = 25
    config.MinConns = 5
    config.MaxConnLifetime = time.Hour
    config.MaxConnIdleTime = time.Minute * 30
    
    return pgxpool.ConnectConfig(context.Background(), config)
}
```

### 2.3 Service Layer Modernization

#### Convert to Interface-Based Services
**New Files**:
- `internal/service/interfaces.go`
- `internal/service/container.go`

**Implementation**:
```go
// interfaces.go
type UserServiceInterface interface {
    SaveUser(firstName, lastName, password, email string, country *string) (*User, error)
    LoginUser(email, password string) (string, error)
    GetProfile(userID int64) (*User, error)
}

// container.go
type ServiceContainer struct {
    UserService     UserServiceInterface
    WordlistService WordlistServiceInterface
    WordService     WordServiceInterface
}

func NewServiceContainer(db *pgxpool.Pool) *ServiceContainer {
    return &ServiceContainer{
        UserService:     NewUserService(db),
        WordlistService: NewWordlistService(db),
        WordService:     NewWordService(db),
    }
}
```

## Phase 3: Advanced Optimizations (3-4 weeks)

### 3.1 Caching Layer Implementation

#### Redis Integration for User Sessions
**New File**: `internal/cache/user_cache.go`

**Implementation**:
```go
type UserCache struct {
    client *redis.Client
}

func (uc *UserCache) CacheUser(userID int64, user *User) error {
    key := fmt.Sprintf("user:%d", userID)
    data, _ := json.Marshal(user)
    return uc.client.Set(context.Background(), key, data, time.Hour).Err()
}

func (uc *UserCache) GetUser(userID int64) (*User, error) {
    key := fmt.Sprintf("user:%d", userID)
    data, err := uc.client.Get(context.Background(), key).Result()
    if err != nil {
        return nil, err
    }
    
    var user User
    json.Unmarshal([]byte(data), &user)
    return &user, nil
}
```

### 3.2 Query Optimization

#### Batch Database Operations
**Target File**: `internal/repository/user.go`

**Optimization**:
```go
// Batch user lookups
func (repository *UserRepository) FindMultiple(ctx context.Context, userIDs []int64) ([]User, error) {
    // Use SQL IN clause for batch lookups
    query := `SELECT id, email, first_name, last_name, subscription_plan 
              FROM users WHERE id = ANY($1)`
    
    rows, err := repository.Db.Query(ctx, query, userIDs)
    // ... process results
}
```

### 3.3 Response Optimization

#### Implement Response Compression
**New File**: `internal/middleware/compression.go`

**Implementation**:
```go
func GzipMiddleware() gin.HandlerFunc {
    return gin.WrapMiddleware(gziphandler.GzipHandler)
}

// Add to main.go
router.Use(middleware.GzipMiddleware())
```

## Implementation Timeline

### Phase 1: Critical Fixes (Week 1)
- [x] Fix user registration double-bcrypt flow (biggest impact) - **COMPLETED: ~50% performance improvement**
- [x] Add explicit subscription_plan = 'free' to repository - **COMPLETED: Better code maintainability**
- [ ] Add essential database indexes (users.email, wordlists.user_id)
- [ ] Implement environment-specific bcrypt costs
- [ ] Add request timeout context for external API calls
- [ ] Add JWT generation optimization
- [ ] Integrate with existing load testing (`make load-test`)

### Phase 2: Architecture & Monitoring (Week 2)
- [ ] Remove global state anti-patterns and os.Exit(1) calls
- [ ] Add OpenAI moderation timeout and retry logic
- [ ] Implement performance monitoring middleware
- [ ] Add proper connection pooling configuration
- [ ] Validate improvements with load testing

### Phase 3: Measure & Decide (Week 3)
- [ ] Run comprehensive performance testing
- [ ] Evaluate if targets are met (< 500ms users, < 300ms login, < 200ms wordlists)
- [ ] Document actual improvements achieved
- [ ] Decide if advanced optimizations are needed based on data

## Success Metrics

### Performance Targets
- **POST /users**: < 500ms (from 7.3s) - 93% improvement
- **POST /login**: < 300ms (from 3.0s) - 90% improvement
- **POST /wordlists**: < 200ms (from 1.2s) - 83% improvement

### Load Testing Integration
**Validate improvements after each phase:**
```bash
# Run load tests using existing infrastructure
cd api && make load-test ARGS="-users 10 -duration 2m -words test_words.txt"

# Compare against baseline metrics:
# Before: POST /users: 7.3s, POST /login: 3.0s, POST /wordlists: 1.2s
# Target: POST /users: <500ms, POST /login: <300ms, POST /wordlists: <200ms

# Test with different user loads
make load-test ARGS="-users 5 -duration 1m -words test_words.txt"   # Light load
make load-test ARGS="-users 15 -duration 3m -words test_words.txt"  # Heavy load
```

### Monitoring Implementation
**New File**: `internal/middleware/performance.go`

```go
func PerformanceMiddleware() gin.HandlerFunc {
    return gin.HandlerFunc(func(c *gin.Context) {
        start := time.Now()
        
        c.Next()
        
        latency := time.Since(start)
        
        // Log slow requests
        if latency > time.Second {
            common.Logger.Warn("slow request", 
                "method", c.Request.Method,
                "path", c.Request.URL.Path,
                "latency", latency,
                "status", c.Writer.Status())
        }
        
        // Record metrics
        metrics.RecordHTTPLatency(c.Request.Method, c.Request.URL.Path, latency)
    })
}
```

## Testing Strategy

### Load Testing Validation
- Run load tests after each phase
- Compare against baseline performance
- Validate P95 and P99 latencies
- Monitor resource utilization

### Regression Testing
- Ensure existing functionality works
- Run full test suite after each change
- Monitor error rates during optimization
- Validate data integrity

## Rollback Procedures

### Database Changes
- Keep migration rollback scripts ready
- Monitor index creation performance
- Have index drop procedures prepared

### Code Changes
- Use feature flags for new implementations
- Keep backward compatibility during transition
- Have rollback commits prepared

## Risk Mitigation

### High-Risk Changes
1. **Global State Removal**: Coordinate with all team members
2. **Database Schema Changes**: Test thoroughly in staging
3. **bcrypt Cost Changes**: Validate authentication still works

### Monitoring During Changes
- Monitor error rates continuously
- Track response times in real-time
- Set up alerts for performance degradation
- Keep rollback procedures ready

## Long-term Maintenance

### Ongoing Optimization
- Regular performance audits
- Continuous monitoring setup
- Performance regression testing
- Database query optimization reviews

### Team Knowledge Transfer
- Document all optimizations
- Create performance best practices guide
- Train team on new patterns
- Establish performance review processes

---

This plan addresses the critical performance bottlenecks identified in the codebase and provides a systematic approach to achieve production-ready performance while maintaining code quality and system reliability.