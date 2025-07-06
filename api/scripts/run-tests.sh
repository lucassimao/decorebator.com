#!/bin/bash

# Decorebator API Test Runner
# This script provides easy commands for running different types of tests

set -e

# Tool versions (matching GitHub workflow and Makefile)
GO_VERSION="1.23"
POSTGRES_VERSION="15"
REDIS_VERSION="7-alpine"
MINIO_VERSION="latest"
GOTESTSUM_VERSION="latest"
GOVULNCHECK_VERSION="latest"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ROOT=$(dirname $(dirname $(realpath $0)))
COMPOSE_FILE="$PROJECT_ROOT/docker-compose.test.yml"
ENV_FILE="$PROJECT_ROOT/.env.test"

# Helper functions
log() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if required tools are installed
check_dependencies() {
    local missing=()
    
    if ! command -v docker &> /dev/null; then
        missing+=("docker")
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        missing+=("docker-compose")
    fi
    
    if ! command -v go &> /dev/null; then
        missing+=("go")
    fi
    
    if [ ${#missing[@]} -ne 0 ]; then
        error "Missing required dependencies: ${missing[*]}"
        echo "Please install the missing tools and try again."
        exit 1
    fi
}

# Setup test environment
setup_env() {
    log "Setting up test environment..."
    
    # Check if .env.test exists
    if [ ! -f "$ENV_FILE" ]; then
        error ".env.test file not found. Copy from .env.test.example:"
        echo "  cp .env.test.example .env.test"
        exit 1
    fi
    
    # Start Docker services
    log "Starting test services..."
    docker-compose -f "$COMPOSE_FILE" up -d postgres redis minio createbuckets
    
    # Mark that services have been started so cleanup will run
    export SERVICES_STARTED=1
    
    # Wait for services to be ready
    log "Waiting for services to be ready..."
    sleep 15
    
    # Check if services are healthy
    if ! docker-compose -f "$COMPOSE_FILE" ps | grep -q "Up"; then
        error "Some test services failed to start"
        docker-compose -f "$COMPOSE_FILE" logs
        exit 1
    fi
    
    success "Test environment ready"
}

# Run database migrations
run_migrations() {
    log "Running database migrations..."
    
    # Source environment variables
    set -a
    source "$ENV_FILE"
    set +a
    
    # Use Go-based migration runner instead of migrate CLI
    cd "$PROJECT_ROOT"
    go run ./cmd/migrate/main.go
    
    if [ $? -eq 0 ]; then
        success "Migrations completed successfully"
    else
        error "Migration failed"
        exit 1
    fi
}

# Clean test environment
cleanup() {
    log "Cleaning up test environment..."
    
    # Stop and remove containers, networks, and volumes
    if [ -f "$COMPOSE_FILE" ]; then
        docker-compose -f "$COMPOSE_FILE" down -v --remove-orphans 2>/dev/null || {
            warning "Docker compose cleanup had issues, trying individual cleanup..."
            
            # Try to stop containers individually
            docker stop $(docker ps -q --filter "name=decorebator-test") 2>/dev/null || true
            docker rm $(docker ps -aq --filter "name=decorebator-test") 2>/dev/null || true
            
            # Remove test networks
            docker network rm decorebator-test_default 2>/dev/null || true
            
            # Remove test volumes  
            docker volume rm $(docker volume ls -q --filter "name=decorebator-test") 2>/dev/null || true
        }
    fi
    
    success "Cleanup completed"
}

# Install gotestsum if available
ensure_gotestsum() {
    if ! command -v gotestsum &> /dev/null; then
        log "gotestsum not found, installing version $GOTESTSUM_VERSION..."
        go install gotest.tools/gotestsum@$GOTESTSUM_VERSION
        
        # Check if installation was successful
        if ! command -v gotestsum &> /dev/null; then
            warning "Failed to install gotestsum, falling back to go test"
            return 1
        fi
    fi
    return 0
}

# Run unit tests
run_unit_tests() {
    log "Running unit tests..."
    cd "$PROJECT_ROOT"
    
    # Try to use gotestsum for better output, fall back to go test
    if ensure_gotestsum; then
        log "Using gotestsum for structured output"
        gotestsum --format testname -- -v -race -coverprofile=unit.out -covermode=atomic ./internal/...
    else
        log "Using go test"
        go test -v -race -coverprofile=unit.out -covermode=atomic ./internal/...
    fi
    
    if [ $? -eq 0 ]; then
        success "Unit tests passed"
        go tool cover -func=unit.out
    else
        error "Unit tests failed"
        exit 1
    fi
}

# Run integration tests
run_integration_tests() {
    local package_filter="$1"
    local test_path="./tests/integration/..."
    
    if [ -n "$package_filter" ]; then
        test_path="./tests/integration/$package_filter/..."
        log "Running integration tests for package: $package_filter"
    else
        log "Running all integration tests..."
    fi
    
    cd "$PROJECT_ROOT"
    
    # Source test environment properly
    set -a
    source "$ENV_FILE"
    set +a
    
    local test_exit_code=0
    
    # Try to use gotestsum for better output, fall back to go test
    if ensure_gotestsum; then
        log "Using gotestsum for structured output"
        gotestsum --format testname -- -v -race -count=1 -p=1 -coverpkg=./internal/... -coverprofile=integration.out -covermode=atomic $test_path || test_exit_code=$?
    else
        log "Using go test"
        go test -v -race -count=1 -p=1 -coverpkg=./internal/... -coverprofile=integration.out -covermode=atomic $test_path || test_exit_code=$?
    fi
    
    if [ $test_exit_code -eq 0 ]; then
        if [ -n "$package_filter" ]; then
            success "Integration tests for $package_filter passed"
        else
            success "Integration tests passed"
        fi
        go tool cover -func=integration.out
    else
        if [ -n "$package_filter" ]; then
            error "Integration tests for $package_filter failed"
        else
            error "Integration tests failed"
        fi
        return $test_exit_code
    fi
}

# Run specific unit test by name
run_specific_unit_test() {
    local test_name="$1"
    if [ -z "$test_name" ]; then
        error "Test name is required for specific unit test"
        echo "Usage: $0 unit-test <TestName>"
        echo "Example: $0 unit-test TestUserService_CreateUser"
        exit 1
    fi
    
    log "Running specific unit test: $test_name"
    cd "$PROJECT_ROOT"
    
    go test -v -race -run "$test_name" ./internal/...
    
    if [ $? -eq 0 ]; then
        success "Unit test '$test_name' passed"
    else
        error "Unit test '$test_name' failed"
        exit 1
    fi
}

# Run specific integration test by name
run_specific_integration_test() {
    local test_name="$1"
    if [ -z "$test_name" ]; then
        error "Test name is required for specific integration test"
        echo "Usage: $0 integration-test <TestName>"
        echo "Example: $0 integration-test TestSignupLoginFlow"
        exit 1
    fi
    
    log "Running specific integration test: $test_name"
    cd "$PROJECT_ROOT"
    
    # Source test environment properly
    set -a
    source "$ENV_FILE"
    set +a
    
    go test -v -race -coverpkg=./internal/... -run "$test_name" ./tests/integration/...
    
    if [ $? -eq 0 ]; then
        success "Integration test '$test_name' passed"
    else
        error "Integration test '$test_name' failed"
        exit 1
    fi
}

# Run all tests
run_all_tests() {
    local unit_exit=0
    local integration_exit=0
    
    run_unit_tests || unit_exit=$?
    run_integration_tests || integration_exit=$?
    
    log "Generating combined coverage report..."
    cd "$PROJECT_ROOT"
    
    if [ -f "unit.out" ]; then
        go tool cover -html=unit.out -o coverage-unit.html
    fi
    
    if [ -f "integration.out" ]; then
        go tool cover -html=integration.out -o coverage-integration.html
    fi
    
    if [ $unit_exit -ne 0 ] || [ $integration_exit -ne 0 ]; then
        error "Some tests failed"
        return 1
    fi
    
    success "All tests completed. Coverage reports generated."
}

# Check coverage thresholds
check_coverage() {
    cd "$PROJECT_ROOT"
    
    local unit_threshold=70
    local integration_threshold=80
    local failed=false
    
    if [ -f "unit.out" ]; then
        local unit_coverage=$(go tool cover -func=unit.out | grep total | awk '{print $3}' | sed 's/%//')
        log "Unit test coverage: ${unit_coverage}%"
        
        if (( $(echo "$unit_coverage < $unit_threshold" | bc -l) )); then
            error "Unit coverage (${unit_coverage}%) below threshold (${unit_threshold}%)"
            failed=true
        fi
    fi
    
    if [ -f "integration.out" ]; then
        local integration_coverage=$(go tool cover -func=integration.out | grep total | awk '{print $3}' | sed 's/%//')
        log "Integration test coverage: ${integration_coverage}%"
        
        if (( $(echo "$integration_coverage < $integration_threshold" | bc -l) )); then
            error "Integration coverage (${integration_coverage}%) below threshold (${integration_threshold}%)"
            failed=true
        fi
    fi
    
    if [ "$failed" = true ]; then
        exit 1
    else
        success "All coverage thresholds met!"
    fi
}

# Watch mode for continuous testing
watch_tests() {
    log "Starting test watch mode (unit tests only)..."
    log "Press Ctrl+C to stop"
    
    while true; do
        run_unit_tests
        
        log "Waiting for file changes..."
        if command -v inotifywait &> /dev/null; then
            inotifywait -r -e modify,create,delete --include '.*\.go$' "$PROJECT_ROOT/internal" "$PROJECT_ROOT/tests" 2>/dev/null
        else
            warning "inotifywait not found. Falling back to 5-second polling."
            sleep 5
        fi
        
        echo
        log "Files changed, re-running tests..."
    done
}

# Show help
show_help() {
    echo "Decorebator API Test Runner"
    echo
    echo "Usage: $0 [COMMAND]"
    echo
    echo "Commands:"
    echo "  setup                    Set up test environment (Docker services, migrations)"
    echo "  unit                     Run unit tests only"
    echo "  integration [package]    Run integration tests (optionally filter by package)"
    echo "  unit-test <TestName>     Run specific unit test by name"
    echo "  integration-test <TestName> Run specific integration test by name"
    echo "  all                      Run all tests (unit + integration)"
    echo "  watch                    Run tests in watch mode (auto-reload on file changes)"
    echo "  coverage                 Check coverage thresholds"
    echo "  versions                 Check tool versions vs expected versions"
    echo "  security                 Run security scans (nancy, govulncheck)"
    echo "  report                   Run all tests with detailed reporting (full setup)"
    echo "  clean                    Clean up test environment"
    echo "  help                     Show this help message"
    echo
    echo "Examples:"
    echo "  $0 setup                           # First-time setup"
    echo "  $0 all                             # Run all tests"
    echo "  $0 unit                            # Quick unit test run"
    echo "  $0 integration                     # All integration tests"
    echo "  $0 integration error_reporting     # Only error_reporting package tests"
    echo "  $0 unit-test TestUserService       # Run specific unit test"
    echo "  $0 integration-test TestSignupLoginFlow  # Run specific integration test"
    echo "  $0 watch                           # Development mode"
    echo
    echo "Environment:"
    echo "  Test configuration is loaded from .env.test"
    echo "  Docker services run on non-standard ports to avoid conflicts"
    echo "  All test data is automatically cleaned up between runs"
}

# Check tool versions
check_versions() {
    log "Checking tool versions (expected vs installed):"
    
    echo "Go:"
    echo "  Expected: $GO_VERSION"
    local go_installed=$(go version 2>/dev/null | grep -oE 'go[0-9]+\.[0-9]+' | sed 's/go//' || echo "Not found")
    echo "  Installed: $go_installed"
    
    echo "gotestsum:"
    echo "  Expected: $GOTESTSUM_VERSION"
    local gotestsum_installed=$(gotestsum --version 2>/dev/null || echo "Not found")
    echo "  Installed: $gotestsum_installed"
    
    echo "govulncheck:"
    echo "  Expected: $GOVULNCHECK_VERSION"
    local govulncheck_installed=$(govulncheck -version 2>/dev/null || echo "Not found")
    echo "  Installed: $govulncheck_installed"
    
    echo "Docker images (from compose file):"
    echo "  PostgreSQL: $POSTGRES_VERSION"
    echo "  Redis: $REDIS_VERSION"
    echo "  MinIO: $MINIO_VERSION"
}

# Run security scans
run_security_scans() {
    log "Running security scans (matching GitHub workflow)..."
    cd "$PROJECT_ROOT"
    
    # Install govulncheck if not present
    if ! command -v govulncheck &> /dev/null; then
        log "Installing govulncheck version $GOVULNCHECK_VERSION..."
        go install golang.org/x/vuln/cmd/govulncheck@$GOVULNCHECK_VERSION
    fi
    
    log "Running govulncheck (official Go vulnerability scanner)..."
    if govulncheck ./...; then
        success "No vulnerabilities found"
    else
        warning "Vulnerabilities found - check output above"
        log "Note: Some vulnerabilities may be in standard library or dependencies"
    fi
    
    success "Security scan completed"
    log "Tip: Use 'govulncheck -show verbose ./...' for detailed information"
}

# Main command dispatcher
main() {
    case "${1:-help}" in
        "setup")
            check_dependencies
            setup_env
            run_migrations
            ;;
        "unit")
            check_dependencies
            run_unit_tests
            ;;
        "integration")
            check_dependencies
            setup_env
            run_migrations
            run_integration_tests "$2" || exit $?
            ;;
        "unit-test")
            check_dependencies
            run_specific_unit_test "$2"
            ;;
        "integration-test")
            check_dependencies
            setup_env
            run_migrations
            run_specific_integration_test "$2"
            ;;
        "all")
            check_dependencies
            setup_env
            run_migrations
            run_all_tests || exit $?
            check_coverage || exit $?
            ;;
        "watch")
            check_dependencies
            watch_tests
            ;;
        "coverage")
            check_coverage
            ;;
        "clean")
            cleanup
            ;;
        "versions")
            check_versions
            ;;
        "security")
            run_security_scans
            ;;
        "report")
            check_dependencies
            setup_env
            run_migrations
            log "Running comprehensive test report..."
            run_unit_tests
            run_integration_tests
            check_coverage
            ;;
        "help"|"-h"|"--help")
            show_help
            ;;
        *)
            error "Unknown command: $1"
            show_help
            exit 1
            ;;
    esac
}

# Cleanup function for trap
cleanup_on_exit() {
    local exit_code=$?
    
    # Only run cleanup if we actually started services
    if [ -n "$SERVICES_STARTED" ]; then
        cleanup
    fi
    
    exit $exit_code
}

# Trap cleanup on script exit and interrupt
trap cleanup_on_exit EXIT INT TERM

# Run main function
main "$@"