# Database Initialization Elimination Plan

## Executive Summary

This document outlines a comprehensive plan to eliminate all database initialization in `init()` functions and global variables throughout the Decorebator API codebase. The current architecture prevents proper unit testing and violates dependency injection principles by establishing database connections at package import time.

### Why This Matters
- **Unit tests fail** due to database connection attempts during package imports
- **Hidden dependencies** make the system difficult to reason about and test
- **Tight coupling** prevents modular development and testing
- **Violates SOLID principles**, specifically Dependency Inversion

### Current Impact
- Unit tests cannot run without database connections
- Integration tests are fragile due to global state
- Dependency injection refactoring is incomplete
- Code quality and maintainability are compromised

## Current State Analysis

### 🔴 Critical Issues (Blocking Unit Tests)

| Location | Issue | Impact |
|----------|-------|---------|
| `internal/http/quiz.go:16` | `var strategy = service.DefaultLeitnerSystemStrategy()` | Global variable triggers DB connection at import |
| `internal/service/wordlist.go:32` | `init()` function calls `common.GetDBConnection()` | Package import requires database |
| `internal/http/setup.go:72` | `applyDefaults()` calls `GetDBConnection()` | HTTP package setup requires database |

### 🟠 High Priority Issues (Major Architecture Problems)

| Location | Issue | Count |
|----------|-------|-------|
| Service layer | Direct `common.GetDBConnection()` calls | 24+ instances |
| HTTP handlers | Database connections in request handlers | 5+ instances |
| Worker services | Database initialization in worker functions | 8+ instances |

### 🟡 Medium Priority Issues (Code Quality)

| Location | Issue | Impact |
|----------|-------|---------|
| `internal/common/logger.go:12` | `init()` function for logger setup | Side effects during import |
| `internal/http/setup.go:31` | Sentry initialization in `init()` | Global state initialization |

## Implementation Plan

## Phase 1: Critical Fixes 🔴
**Priority**: CRITICAL - Must be completed first to enable unit testing

### Step 1.1: Remove Global Strategy Variable ✅ COMPLETED
- [x] **File**: `internal/http/quiz.go`
- [x] **Action**: Remove `var strategy = service.DefaultLeitnerSystemStrategy()` (line 16)
- [x] **Replace with**: Dependency injection in `QuizRoutes` struct
- [x] **Create**: `NewQuizRoutes(strategy common.SpacedRepetitionStrategy)` constructor
- [x] **Update**: All references to use injected strategy
- [x] **Test**: Verify HTTP package imports without database connection

**Implementation Details:**
- Removed global variable triggering database connection at import time
- Added `strategy` field to `QuizRoutes` struct
- Created `NewQuizRoutes(strategy)` constructor function
- Updated `h.strategy.CreateQuiz()` and `h.strategy.SaveQuizResult()` method calls
- Modified `internal/http/setup.go` to use constructor injection
- Fixed golangci-lint formatting issues
- Updated Makefile `test-unit` target to run `./internal/tests/unit/...`

### Step 1.2: Eliminate Service init() Functions
- [ ] **File**: `internal/service/wordlist.go`
- [ ] **Action**: Remove `init()` function (lines 32-37)
- [ ] **Replace with**: Lazy initialization or constructor injection
- [ ] **Update**: All `DefaultWordlistService()` callers to use explicit injection
- [ ] **Test**: Verify service package imports without database connection

### Step 1.3: Fix HTTP Setup Database Dependencies
- [ ] **File**: `internal/http/setup.go`
- [ ] **Action**: Remove `applyDefaults()` function database fallback (line 72)
- [ ] **Replace with**: Required explicit configuration
- [ ] **Update**: `SetupRoutes(config *Config)` to require all dependencies
- [ ] **Test**: Verify HTTP setup only uses injected dependencies

**Success Criteria**: Unit tests run without database connections

---

## Phase 2: Service Layer Dependency Injection 🟠
**Priority**: HIGH - Core architecture improvements

### Step 2.1: Create Application Context
- [ ] **File**: `internal/app/context.go` (new)
- [ ] **Action**: Create `AppContext` struct containing all services
- [ ] **Include**: Database, all services, configuration
- [ ] **Pattern**: Builder pattern `NewAppContext().WithDatabase(db).Build()`
- [ ] **Test**: Verify clean service initialization

### Step 2.2: Convert Service Constructors
- [ ] **Files**: All service files with `GetDBConnection()` calls
- [ ] **Action**: Update all 24+ instances to use injected database
- [ ] **Pattern**: `NewXXXService(db *pgxpool.Pool, deps...)` 
- [ ] **Remove**: All `DefaultXXXService()` functions that call `GetDBConnection()`
- [ ] **Test**: Verify all services use explicit dependencies

### Step 2.3: Update HTTP Handlers
- [ ] **Files**: `internal/http/*.go`
- [ ] **Action**: Remove direct `GetDBConnection()` calls (5+ instances)
- [ ] **Replace with**: Injected services or database connections
- [ ] **Update**: Route handlers to use dependency injection
- [ ] **Test**: Verify HTTP handlers work with injected dependencies

### Step 2.4: Convert Worker Services
- [ ] **Files**: `internal/service/*_worker.go`
- [ ] **Action**: Remove `GetDBConnection()` calls (8+ instances)
- [ ] **Replace with**: Constructor injection or service dependencies
- [ ] **Update**: Worker initialization to use explicit dependencies
- [ ] **Test**: Verify workers function with injected dependencies

**Success Criteria**: No direct database connections in service layer

---

## Phase 3: Remove Remaining GetDBConnection() Calls 🟡
**Priority**: MEDIUM - Code quality and consistency

### Step 3.1: Audit and Convert Remaining Calls
- [ ] **Files**: Scan entire codebase for remaining `GetDBConnection()` calls
- [ ] **Action**: Convert each to use dependency injection
- [ ] **Priority Order**:
  1. [ ] `internal/service/leitner_system_strategy.go` (11 instances)
  2. [ ] `internal/service/error_reporting.go` (2 instances)
  3. [ ] `internal/service/analytics.go` (1 instance)
  4. [ ] `internal/mail/mail.go` (1 instance)
  5. [ ] Other files as discovered
- [ ] **Test**: Verify each conversion maintains functionality

### Step 3.2: Update Main Application Bootstrap
- [ ] **File**: `main.go`
- [ ] **Action**: Create single database connection
- [ ] **Initialize**: All services with explicit dependencies
- [ ] **Pass**: Complete configuration to HTTP setup
- [ ] **Test**: Verify application starts cleanly

**Success Criteria**: Zero `GetDBConnection()` calls outside of main.go

---

## Phase 4: Safe Initialization and Cleanup 🟢
**Priority**: LOW - Polish and best practices

### Step 4.1: Make Logger Initialization Safe
- [ ] **File**: `internal/common/logger.go`
- [ ] **Action**: Convert `init()` to lazy initialization
- [ ] **Pattern**: `sync.Once` for thread-safe lazy loading
- [ ] **Test**: Verify logger works without import-time side effects

### Step 4.2: Optional Sentry Initialization
- [ ] **File**: `internal/http/setup.go`
- [ ] **Action**: Make Sentry initialization optional/lazy
- [ ] **Move**: From `init()` to explicit initialization
- [ ] **Test**: Verify Sentry works when explicitly initialized

### Step 4.3: Create Database Connection Interface
- [ ] **File**: `internal/common/database.go` (optional)
- [ ] **Action**: Create interface for database connections
- [ ] **Benefit**: Easier mocking and testing
- [ ] **Test**: Verify interface works with existing code

**Success Criteria**: All initialization is explicit and testable

---

## Testing Strategy

### Unit Test Verification
```bash
# After each phase, verify unit tests work without database
make test-unit

# Should complete without database connection errors
# Expected: Only timeout middleware tests run (or other pure unit tests)
```

### Integration Test Verification
```bash
# Verify integration tests still work
make test-integration

# Should work with proper database connections in test environment
```

### Smoke Test Application
```bash
# Verify application still starts and works
make watch

# Check that all endpoints respond correctly
curl http://localhost:8080/health  # if health endpoint exists
```

## Dependencies and Execution Order

### Critical Path
1. **Phase 1** must be completed before unit tests can work
2. **Step 1.1** (quiz.go) is the highest priority - most immediate impact
3. **Step 1.2** (wordlist.go) enables service layer testing
4. **Step 1.3** (setup.go) enables HTTP layer testing

### Parallel Work Opportunities
- **Phase 2.2** and **Phase 2.3** can be done in parallel
- **Phase 3.1** can be done incrementally alongside other phases
- **Phase 4** can be done independently after critical fixes

### Validation Checkpoints
- [ ] **After Phase 1**: Unit tests run without database
- [ ] **After Phase 2**: All services use dependency injection  
- [ ] **After Phase 3**: Zero global database connections
- [ ] **After Phase 4**: All initialization is explicit and safe

## Progress Tracking

### Overall Progress
- [x] **Phase 1**: Critical Fixes (1/3 steps completed) ✅ Step 1.1 DONE
- [ ] **Phase 2**: Service Layer DI (0/4 steps completed) 
- [ ] **Phase 3**: Remove Remaining Calls (0/2 steps completed)
- [ ] **Phase 4**: Safe Initialization (0/3 steps completed)

### Current Status
**Status**: 🟡 Phase 1 In Progress
**Completed**: Step 1.1 - Global strategy variable eliminated ✅
**Next Action**: Begin Phase 1, Step 1.2 (Remove service init() functions)
**Estimated Time**: 
- Phase 1: 2-4 hours
- Phase 2: 4-8 hours  
- Phase 3: 2-4 hours
- Phase 4: 1-2 hours

### Risk Mitigation
- **Backup**: Create feature branch before major changes
- **Incremental**: Test after each step completion
- **Rollback**: Keep previous working state available
- **Documentation**: Update CLAUDE.md after completion

---

## Expected Benefits After Completion

### ✅ Immediate Benefits
- Unit tests run without database dependencies
- Faster test execution
- Better development experience
- Proper separation of concerns

### ✅ Architecture Benefits  
- Clean dependency injection throughout
- Explicit dependency graph
- Better testability and maintainability
- SOLID principle compliance

### ✅ Development Benefits
- Easier to add new features
- Simpler testing strategies
- Better error isolation
- Improved debugging experience

---

**Last Updated**: 2025-07-13 (Step 1.1 Completed)
**Document Owner**: Development Team
**Status**: 🟡 Phase 1 In Progress - Step 1.1 ✅ COMPLETED