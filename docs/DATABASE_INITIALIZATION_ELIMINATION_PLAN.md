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

### Step 1.2: Eliminate Service init() Functions ✅ COMPLETED
- [x] **File**: `internal/service/wordlist.go`
- [x] **Action**: Remove `init()` function and legacy `GetWordlistById()` function
- [x] **Replace with**: Dependency injection in analytics HTTP handlers
- [x] **Update**: All analytics routes to use injected `WordlistService`
- [x] **Test**: Verify service package imports without database connection

**Implementation Details:**
- Converted all analytics HTTP handlers to use dependency injection pattern
- Updated `RegisterAnalyticsRoutes()` to accept `WordlistService` parameter
- Converted handler functions to return `gin.HandlerFunc` with injected dependencies
- Removed legacy `GetWordlistById()` function that used `common.GetDBConnection()`
- Updated `internal/http/setup.go` to pass `WordlistService` to analytics routes
- Fixed variable shadowing lint issue in analytics handlers

### Step 1.3: Fix HTTP Setup Database Dependencies ✅ COMPLETED
- [x] **File**: `internal/http/setup.go`
- [x] **Action**: Remove `applyDefaults()` function database fallback (line 72)
- [x] **Replace with**: Required explicit configuration and dependency injection
- [x] **Update**: All HTTP handlers to use injected database connections
- [x] **Test**: Verify HTTP setup only uses injected dependencies

**Implementation Details:**
- Removed database fallback in `applyDefaults()` function with explicit error
- Converted `GetErrorReportStats()` to accept `*pgxpool.Pool` parameter
- Converted `GetUserErrorReportStatus()` to accept `*pgxpool.Pool` parameter  
- Converted `RateLimitErrorReports()` middleware to accept `*pgxpool.Pool` parameter
- Updated all route registrations in `SetupRoutes()` to pass database connections
- Verified no remaining `common.GetDBConnection()` calls in HTTP layer

**Success Criteria**: ✅ **ACHIEVED** - Unit tests run without database connections

---

## Phase 2: Service Layer Dependency Injection 🟠
**Priority**: HIGH - Core architecture improvements

### Step 2.1: Create Application Context ✅ COMPLETED
- [x] **File**: `internal/app/context.go` (new)
- [x] **Action**: Create `Context` struct containing all services
- [x] **Include**: Database, RiverClient, all services, configuration
- [x] **Pattern**: Builder pattern `NewContext().WithDatabase(db).Build()`
- [x] **Test**: Verify clean service initialization

**Implementation Details:**
- Created centralized `Context` struct with all application dependencies
- Implemented fluent builder pattern with `ContextBuilder`
- Added builder methods for all services: `WithWordService()`, `WithModerationService()`, etc.
- Added factory method support for test service injection (`WithRevenueCatServiceFunc()`)
- Updated main application (`cmd/api/server.go`) to use new AppContext
- Refactored HTTP setup (`internal/http/setup.go`) to accept AppContext parameter
- Converted integration test setup to use `AppContextConfigFunc` pattern
- Replaced deprecated `TestConfig` with clean AppContext dependency injection
- Added proper lifecycle management with `Context.Close()` method
- Ensured conditional service initialization (only create if not provided)

### Step 2.2: Convert Service Constructors ✅ IN PROGRESS
- [ ] **Files**: All service files with `GetDBConnection()` calls
- [ ] **Action**: Update all instances to use injected database
- [ ] **Pattern**: `NewXXXService(db *pgxpool.Pool, deps...)` 
- [ ] **Remove**: All `DefaultXXXService()` functions that call `GetDBConnection()`
- [ ] **Test**: Verify all services use explicit dependencies

**Remaining GetDBConnection() calls by priority:**
1. **leitner_system_strategy.go**: 12 instances (starting with DefaultLeitnerSystemStrategy)
2. **error_reporting.go**: 2 instances  
3. **analytics.go**: 1 instance
4. **definition_fetcher_worker.go**: 2 instances
5. **example_audio_worker.go**: 1 instance
6. **worker_validation.go**: 1 instance
7. **river.go**: 1 instance
8. **mail.go**: 1 instance

**Current Focus**: Step 2.2.1 - Refactor LeitnerSystemStrategy service constructor

#### Step 2.2.1: LeitnerSystemStrategy + ErrorReportService Refactoring ✅ COMPLETED
- [x] **Action**: Remove `DefaultLeitnerSystemStrategy()` function from `leitner_system_strategy.go`
- [x] **Update**: Add `LeitnerSystemStrategy` to `AppContext` struct and builder
- [x] **Update**: Initialize `LeitnerSystemStrategy` in `AppContext.initializeServices()`
- [x] **Update**: Use `AppContext.LeitnerSystemStrategy` in HTTP setup instead of creating new instance
- [x] **Bonus**: Complete ErrorReportService dependency injection refactoring
- [x] **Update**: Restructured ErrorReportService to separate dependencies from request data
- [x] **Update**: Added ErrorReportService to AppContext with proper initialization
- [x] **Update**: Simplified ErrorReportRoutes to only depend on ErrorReportService
- [x] **Update**: Fix integration test to use proper service creation
- [x] **Test**: Unit tests pass ✅, Linting passes ✅

**Files Modified**:
- `internal/service/leitner_system_strategy.go` - Removed `DefaultLeitnerSystemStrategy()` function
- `internal/app/context.go` - Added LeitnerSystemStrategy AND ErrorReportService with proper initialization
- `internal/http/setup.go` - Updated to use both services from AppContext
- `internal/service/error_reporting.go` - Completely restructured to separate dependencies from request data
- `internal/http/error_reporting.go` - Simplified to only depend on ErrorReportService
- `tests/integration/errorreporting/side_effects_test.go` - Updated to create services properly

**Result**: Eliminated multiple `GetDBConnection()` calls and converted both LeitnerSystemStrategy and ErrorReportService to proper dependency injection architecture

**Architecture Benefits Achieved**:
- Clean separation of concerns between services and HTTP handlers
- Proper dependency injection managed by AppContext
- Easier testing with mockable service dependencies
- Reduced parameter passing complexity
- Consistent architecture patterns across services

#### Step 2.2.2: AnalyticsService Refactoring
- [ ] **Action**: Remove `GetDBConnection()` call in `NewAnalyticsService()`
- [ ] **Update**: Update constructor to accept database connection as parameter
- [ ] **Update**: Enhance AnalyticsService management in AppContext if needed
- [ ] **Update**: Update HTTP handlers to use AnalyticsService from AppContext
- [ ] **Test**: Verify all tests pass and linting passes

**Target**: `internal/service/analytics.go:52` - Single `GetDBConnection()` call in service constructor

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
- [x] **After Phase 1**: Unit tests run without database ✅ **COMPLETED**
- [ ] **After Phase 2**: All services use dependency injection  
- [ ] **After Phase 3**: Zero global database connections
- [ ] **After Phase 4**: All initialization is explicit and safe

## Progress Tracking

### Overall Progress
- [x] **Phase 1**: Critical Fixes ✅ **COMPLETED** (3/3 steps completed)
- [ ] **Phase 2**: Service Layer DI (1/4 steps completed) 🟠 **Step 2.1 DONE** 
- [ ] **Phase 3**: Remove Remaining Calls (0/2 steps completed)
- [ ] **Phase 4**: Safe Initialization (0/3 steps completed)

### Current Status
**Status**: 🟠 Phase 2 IN PROGRESS - Service Constructor Refactoring
**Completed**: 
- ✅ **Phase 1**: All critical fixes completed (global variables, init() functions, HTTP dependencies)
- ✅ **Step 2.1**: Application Context system implemented with dependency injection

**In Progress**: Step 2.2.1 - Refactor LeitnerSystemStrategy service constructor

**Next Action**: Remove DefaultLeitnerSystemStrategy() function and update all callers
**Time Spent on Phase 1**: ~3 hours (within estimated range)
**Remaining Estimated Time**: 
- Phase 2: 4-8 hours  
- Phase 3: 2-4 hours
- Phase 4: 1-2 hours

**Key Achievement**: 🎉 Unit tests now run successfully without database connections!

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

**Last Updated**: 2025-07-13 (Phase 1 COMPLETED)
**Document Owner**: Development Team
**Status**: 🟢 Phase 1 COMPLETED - All critical fixes implemented ✅

## Phase 1 Completion Summary

**🎯 Mission Accomplished**: The primary goal of Phase 1 has been achieved - unit tests now run successfully without requiring database connections.

**What Was Fixed**:
1. **Global Variable Elimination**: Removed database-triggering global strategy variable in HTTP layer
2. **Service Init() Removal**: Eliminated legacy functions and converted analytics to dependency injection
3. **HTTP Dependency Injection**: All HTTP handlers now use explicit database dependencies

**Impact**:
- ✅ Unit tests run without database setup
- ✅ Faster test execution (no database overhead)
- ✅ Better separation of concerns in HTTP and service layers
- ✅ Foundation laid for further dependency injection improvements

**Test Verification**: 
```bash
make test-unit  # ✅ PASSES - No database connection required
make lint-changed  # ✅ PASSES - All code quality checks pass
```

**Next Steps**: Phase 2-4 are now unblocked and can be tackled incrementally as needed for broader architecture improvements.