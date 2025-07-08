# Load environment variables from the .env file
include .env

# Tool versions (matching GitHub workflow)
GO_VERSION := 1.23
POSTGRES_VERSION := 15
REDIS_VERSION := 7-alpine
MINIO_VERSION := latest
GOTESTSUM_VERSION := latest
GOVULNCHECK_VERSION := latest
GODOTENV_VERSION := latest
DELVE_VERSION := latest

GREEN  		  := $(shell tput -Txterm setaf 2)
YELLOW 		  := $(shell tput -Txterm setaf 3)
WHITE  		  := $(shell tput -Txterm setaf 7)
CYAN   		  := $(shell tput -Txterm setaf 6)
RED    		  := $(shell tput -Txterm setaf 1)
RESET  		  := $(shell tput -Txterm sgr0)
GODOTENV_PATH := ~/go/bin/godotenv
DIR_TO_WATCH  := $(PWD)

all: help

setup: ## Install all dev and build dependencies (matching GitHub workflow versions)
	@echo "${YELLOW}Installing development tools (versions matching GitHub workflow)...${RESET}"
	go install github.com/joho/godotenv/cmd/godotenv@$(GODOTENV_VERSION)
	go install github.com/go-delve/delve/cmd/dlv@$(DELVE_VERSION)
	go install gotest.tools/gotestsum@$(GOTESTSUM_VERSION)
	go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)
	go install github.com/riverqueue/river/cmd/river@latest
	@echo "${YELLOW}Installing system dependencies...${RESET}"
	sudo apt-get update && sudo apt-get install -y inotify-tools migrate
	@echo "${GREEN}✅ All dependencies installed successfully${RESET}"
	@echo "${CYAN}Tool versions:${RESET}"
	@echo "  Go: $(GO_VERSION)"
	@echo "  PostgreSQL: $(POSTGRES_VERSION)"  
	@echo "  Redis: $(REDIS_VERSION)"
	@echo "  MinIO: $(MINIO_VERSION)"

migrate-up: ## Run database migrations
	@POSTGRESQL_URL=$$($(GODOTENV_PATH) -f .env sh -c 'echo postgres://$$POSTGRES_USER:$$POSTGRES_PASSWORD@localhost:$$POSTGRES_PORT/$$POSTGRES_DB?sslmode=disable'); \
	migrate -database $$POSTGRESQL_URL -path cmd/migrate/migrations up && \
	river migrate-up --database-url $$POSTGRESQL_URL

migrate-down: ## Rollback the last database migration executed
	@POSTGRESQL_URL=$$($(GODOTENV_PATH) -f .env sh -c 'echo postgres://$$POSTGRES_USER:$$POSTGRES_PASSWORD@localhost:$$POSTGRES_PORT/$$POSTGRES_DB?sslmode=disable'); \
	migrate -database $$POSTGRESQL_URL -path cmd/migrate/migrations down 1

migrate-drop: ## Drop everything
	@POSTGRESQL_URL=$$($(GODOTENV_PATH) -f .env sh -c 'echo postgres://$$POSTGRES_USER:$$POSTGRES_PASSWORD@localhost:$$POSTGRES_PORT/$$POSTGRES_DB?sslmode=disable'); \
	migrate -database $$POSTGRESQL_URL -path cmd/migrate/migrations drop

test: ## run containerized integration tests
	rm -rf coverage.out coverage.html &&\
	docker compose -f docker-compose.test.yml rm -f -s -v &&\
	docker compose -f docker-compose.test.yml up --force-recreate --build --exit-code-from main; \
	docker compose -f docker-compose.test.yml down -v; \
	go tool cover -html=coverage.out -o=coverage.html

test-unit: ## Run unit tests
	@echo "Running unit tests..."
	go test -v -race -coverprofile=unit.out -covermode=atomic ./internal/...
	go tool cover -func=unit.out

test-integration: ## Run integration tests locally with proper cleanup
	@echo "Running integration tests via test script..."
	./scripts/run-tests.sh integration

test-all: ## Run all tests with proper cleanup  
	@echo "Running all tests via test script..."
	./scripts/run-tests.sh all

test-fast: ## Run tests without Docker (requires local services)
	@echo "Running fast integration tests (requires local services)..."
	go test -v -short -coverpkg=./internal/... ./tests/integration/...

test-clean: ## Clean up test environment (stop containers, remove volumes)
	@echo "Cleaning up test environment..."
	./scripts/run-tests.sh clean

migrate-up-test: ## Run database migrations for test database
	@POSTGRESQL_URL=$$($(GODOTENV_PATH) -f .env.test sh -c 'echo $$TEST_DATABASE_URL'); \
	migrate -database $$POSTGRESQL_URL -path cmd/migrate/migrations up && \
	river migrate-up --database-url $$POSTGRESQL_URL

migrate-down-test: ## Rollback test database migration
	@POSTGRESQL_URL=$$($(GODOTENV_PATH) -f .env.test sh -c 'echo $$TEST_DATABASE_URL'); \
	migrate -database $$POSTGRESQL_URL -path cmd/migrate/migrations down 1

coverage-html: ## Generate HTML coverage reports
	@if [ -f unit.out ]; then go tool cover -html=unit.out -o coverage-unit.html; fi
	@if [ -f integration.out ]; then go tool cover -html=integration.out -o coverage-integration.html; fi
	@echo "Coverage reports generated: coverage-unit.html, coverage-integration.html"

coverage-threshold: ## Check coverage thresholds
	@echo "Checking coverage thresholds..."
	@if [ -f unit.out ]; then \
		UNIT_COV=$$(go tool cover -func=unit.out | grep total | awk '{print $$3}' | sed 's/%//'); \
		echo "Unit coverage: $${UNIT_COV}%"; \
		if [ $$(echo "$$UNIT_COV < 70" | bc -l) -eq 1 ]; then \
			echo "❌ Unit coverage ($${UNIT_COV}%) below threshold (70%)"; \
			exit 1; \
		fi; \
	fi
	@if [ -f integration.out ]; then \
		INT_COV=$$(go tool cover -func=integration.out | grep total | awk '{print $$3}' | sed 's/%//'); \
		echo "Integration coverage: $${INT_COV}%"; \
		if [ $$(echo "$$INT_COV < 80" | bc -l) -eq 1 ]; then \
			echo "❌ Integration coverage ($${INT_COV}%) below threshold (80%)"; \
			exit 1; \
		fi; \
	fi
	@echo "✅ All coverage thresholds met!"

test-watch: ## Run tests in watch mode
	@echo "Watching for changes and running tests..."
	@while true; do \
		$(MAKE) test-unit; \
		inotifywait -r -e modify,create,delete --include '.*\.go$$' ./internal ./tests; \
		sleep 1; \
	done

test-report: ## Run tests with detailed reporting (requires gotestsum)
	@echo "${YELLOW}Running tests with detailed reporting...${RESET}"
	@command -v gotestsum >/dev/null 2>&1 || { echo "Installing gotestsum@$(GOTESTSUM_VERSION)..."; go install gotest.tools/gotestsum@$(GOTESTSUM_VERSION); }
	@echo "Unit tests:"
	@gotestsum --format testname -- -v -race -coverprofile=unit.out -covermode=atomic ./internal/... || true
	@echo "\n${GREEN}Unit tests completed.${RESET}"
	@echo "\n${YELLOW}Note: For integration tests, use 'make test-integration' or './scripts/run-tests.sh all'${RESET}"
	@echo "${YELLOW}Integration tests require database setup and can't run standalone.${RESET}"
	@echo "\n${GREEN}Coverage summary:${RESET}"
	@if [ -f unit.out ]; then echo "Unit coverage:"; go tool cover -func=unit.out | tail -1; fi

test-report-full: ## Run ALL tests with detailed reporting (includes integration tests)
	@echo "${YELLOW}Running full test suite with detailed reporting...${RESET}"
	@echo "${YELLOW}This will use the test script for proper environment setup...${RESET}"
	./scripts/run-tests.sh all

test-ci: ## Run tests with CI-style reporting (unit tests only)
	@echo "${YELLOW}Running CI-style tests with structured output...${RESET}"
	@command -v gotestsum >/dev/null 2>&1 || { echo "Installing gotestsum@$(GOTESTSUM_VERSION)..."; go install gotest.tools/gotestsum@$(GOTESTSUM_VERSION); }
	gotestsum --junitfile unit-tests.xml --format testname -- -v -race -coverprofile=unit.out -covermode=atomic ./internal/...
	@echo "\n${GREEN}Unit tests completed with JUnit XML output.${RESET}"
	@echo "${YELLOW}For integration tests with full CI setup, use './scripts/run-tests.sh all'${RESET}"

check-versions: ## Check installed tool versions vs expected versions
	@echo "${CYAN}Checking tool versions (expected vs installed):${RESET}"
	@echo "Go:"
	@echo "  Expected: $(GO_VERSION)"
	@go version 2>/dev/null | grep -oE 'go[0-9]+\.[0-9]+' | sed 's/go/  Installed: /' || echo "  Installed: Not found"
	@echo "gotestsum:"
	@echo "  Expected: $(GOTESTSUM_VERSION)"
	@gotestsum --version 2>/dev/null | sed 's/^/  Installed: /' || echo "  Installed: Not found (run 'make setup')"
	@echo "govulncheck:"
	@echo "  Expected: $(GOVULNCHECK_VERSION)"
	@govulncheck -version 2>/dev/null | sed 's/^/  Installed: /' || echo "  Installed: Not found (run 'make setup')"


run: ## HTTP server
	$(GODOTENV_PATH) -f .env go run ./cmd/api

workers: ## Start background workers
	$(GODOTENV_PATH) -f .env go run ./cmd/workers

watch:
	@while true; do \
		$(MAKE) run & \
		PID=$$!; \
		inotifywait -r -e modify,create,delete,move --include '.*\.go$$' $(DIR_TO_WATCH); \
		if kill -0 $$PID 2>/dev/null; then \
			kill $$PID; \
		fi; \
		sleep 1; \
	done

psql: ## Start psql session to development database
	@$(GODOTENV_PATH) -f .env sh -c 'PGPASSWORD=$(POSTGRES_PASSWORD) psql -p $(POSTGRES_PORT) -U $(POSTGRES_USER) -h localhost -d $(POSTGRES_DB)'

build: clean ## Build restful api and save it to bin/main
	go build -o bin/main ./cmd/api

clean: ## Remove build related file
	rm -fr ./bin

debug-workers: ## Starts a delve session for workers
	@$(GODOTENV_PATH) -f .env dlv debug cmd/workers/main.go

debug-api: ## Starts a delve session for api
	@$(GODOTENV_PATH) -f .env dlv debug cmd/api/server.go

load-test: ## Run load testing script with environment variables and pass-through arguments
	$(GODOTENV_PATH) -f .env go run tests/load_test_runner.go $(ARGS)

create-migration: ## Create a new database migration file with the provided file name
	@read -p "Enter the migration file name: " FILE_NAME; \
	if [ -z "$$FILE_NAME" ]; then \
		echo "Migration file name is required."; \
	else \
		migrate create -ext sql -dir cmd/migrate/migrations -seq $$FILE_NAME; \
	fi	

## Quality Assurance Targets

qa-install: ## Install QA tools matching GitHub workflow versions
	@echo "${YELLOW}Installing QA tools...${RESET}"
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.62.2; \
	fi
	@if ! command -v gosec >/dev/null 2>&1; then \
		echo "Installing gosec..."; \
		go install github.com/securego/gosec/v2/cmd/gosec@latest; \
	fi
	@if ! command -v govulncheck >/dev/null 2>&1; then \
		echo "Installing govulncheck..."; \
		go install golang.org/x/vuln/cmd/govulncheck@latest; \
	fi
	@if ! command -v gotestsum >/dev/null 2>&1; then \
		echo "Installing gotestsum..."; \
		go install gotest.tools/gotestsum@latest; \
	fi
	@echo "${GREEN}✅ All QA tools installed${RESET}"

qa: qa-install ## Run comprehensive quality assurance suite matching GitHub workflow
	@echo "${CYAN}🔍 Running comprehensive Quality Assurance suite...${RESET}"
	@echo "${YELLOW}This matches what runs in the GitHub workflow${RESET}\n"
	@$(MAKE) lint
	@$(MAKE) security-scan-full
	@$(MAKE) test-unit
	@$(MAKE) test-integration
	@echo "${GREEN}✅ All QA checks completed${RESET}"

qa-fast: qa-install ## Run fast QA checks (no integration tests)
	@echo "${CYAN}🚀 Running fast QA checks...${RESET}"
	@$(MAKE) lint
	@$(MAKE) security-scan-full  
	@$(MAKE) test-unit
	@echo "${GREEN}✅ Fast QA checks completed${RESET}"

qa-ci: qa-install ## Run QA checks with CI-style output (JUnit XML + structured logs)
	@echo "${CYAN}🤖 Running QA checks with CI-style output...${RESET}"
	@mkdir -p reports
	@$(MAKE) lint > reports/lint.log 2>&1 || true
	@$(MAKE) security-scan-full > reports/security.log 2>&1 || true
	@gotestsum --junitfile reports/unit-tests.xml --format testname -- -v -race -coverprofile=reports/unit.out -covermode=atomic ./internal/... || true
	@gotestsum --junitfile reports/integration-tests.xml --format testname -- -v -race -count=1 -p=1 -coverpkg=./internal/... -coverprofile=reports/integration.out -covermode=atomic ./tests/integration/... || true
	@echo "${GREEN}✅ CI-style QA checks completed - reports in ./reports/${RESET}"

security-scan-full: ## Run comprehensive security scanning (matching GitHub workflow)
	@echo "${CYAN}🔒 Running comprehensive security scans...${RESET}"
	@echo "\n${YELLOW}1. Running gosec security scanner...${RESET}"
	@gosec -fmt json -out gosec-results.json -stdout -verbose=text ./... || true
	@echo "\n${YELLOW}2. Running govulncheck vulnerability scanner...${RESET}"
	@govulncheck ./...
	@echo "\n${YELLOW}3. Running go vet security checks...${RESET}"
	@go vet ./...
	@echo "${GREEN}✅ Security scans completed${RESET}"

security-scan: ## Run basic security scan (gosec + govulncheck)
	@echo "${CYAN}🔒 Running security scans...${RESET}"
	@govulncheck ./...
	@gosec ./...
	@echo "${GREEN}✅ Security scan completed${RESET}"

## Code Quality Targets

lint: ## Run golangci-lint on all Go files
	@echo "${YELLOW}Running golangci-lint on all files...${RESET}"
	golangci-lint run --timeout=5m

lint-changed: ## Run golangci-lint on all changed files (committed + staged + working)
	@echo "${YELLOW}Running golangci-lint on all changed files (committed, staged, and working)...${RESET}"
	@{ \
		git diff --name-only --diff-filter=AM origin/master...HEAD 2>/dev/null || true; \
		git diff --cached --name-only --diff-filter=AM 2>/dev/null || true; \
		git diff --name-only --diff-filter=AM 2>/dev/null || true; \
	} | grep '\.go$$' | grep '^api/' | sed 's|^api/||' | sort | uniq > /tmp/changed_files.txt || true
	@if [ ! -s /tmp/changed_files.txt ]; then \
		echo "${GREEN}No Go files changed (committed, staged, or working directory)${RESET}"; \
	else \
		echo "${CYAN}Changed files: $$(tr '\n' ' ' < /tmp/changed_files.txt)${RESET}"; \
		echo "${CYAN}Running golangci-lint with --new-from-rev (includes all committed + uncommitted changes)${RESET}"; \
		golangci-lint run --timeout=5m --new-from-rev=$$(git merge-base origin/master HEAD 2>/dev/null || echo "HEAD~10"); \
	fi
	@rm -f /tmp/changed_files.txt

lint-staged: ## Run golangci-lint only on staged files
	@echo "${YELLOW}Running golangci-lint on staged files...${RESET}"
	@STAGED_FILES=$$(git diff --cached --name-only --diff-filter=AM | grep '\.go$$' | grep '^api/' || echo ""); \
	if [ -z "$$STAGED_FILES" ]; then \
		echo "${GREEN}No Go files staged${RESET}"; \
	else \
		echo "${CYAN}Staged files: $$STAGED_FILES${RESET}"; \
		golangci-lint run --timeout=5m --new-from-rev=HEAD~1; \
	fi

lint-branch: ## Run golangci-lint on files changed in current branch vs specified base (usage: make lint-branch BASE=develop)
	@BASE_BRANCH=$${BASE:-master}; \
	echo "${YELLOW}Running golangci-lint on files changed since $$BASE_BRANCH...${RESET}"; \
	CHANGED_FILES=$$(git diff --name-only --diff-filter=AM $$BASE_BRANCH...HEAD | grep '\.go$$' | grep '^api/' || echo ""); \
	if [ -z "$$CHANGED_FILES" ]; then \
		echo "${GREEN}No Go files changed since $$BASE_BRANCH${RESET}"; \
	else \
		echo "${CYAN}Changed files: $$CHANGED_FILES${RESET}"; \
		golangci-lint run --timeout=5m --new-from-rev=$$(git merge-base $$BASE_BRANCH HEAD); \
	fi

lint-fix: ## Run golangci-lint with --fix flag on changed files
	@echo "${YELLOW}Running golangci-lint --fix on changed files...${RESET}"
	@CHANGED_FILES=$$(git diff --name-only --diff-filter=AM origin/master...HEAD | grep '\.go$$' | grep '^api/' || echo ""); \
	if [ -z "$$CHANGED_FILES" ]; then \
		echo "${GREEN}No Go files changed since master branch${RESET}"; \
	else \
		echo "${CYAN}Fixing files: $$CHANGED_FILES${RESET}"; \
		golangci-lint run --fix --timeout=5m --new-from-rev=$$(git merge-base origin/master HEAD); \
	fi

lint-install: ## Install golangci-lint
	@echo "${YELLOW}Installing golangci-lint...${RESET}"
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.62.2; \
		echo "${GREEN}golangci-lint installed successfully${RESET}"; \
	else \
		echo "${GREEN}golangci-lint is already installed${RESET}"; \
		golangci-lint version; \
	fi

format: ## Format Go code using gofmt and goimports
	@echo "${YELLOW}Formatting Go code...${RESET}"
	@go fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	else \
		echo "${CYAN}Installing goimports...${RESET}"; \
		go install golang.org/x/tools/cmd/goimports@latest; \
		goimports -w .; \
	fi
	@echo "${GREEN}Code formatted${RESET}"

format-check: ## Check if Go code is properly formatted
	@echo "${YELLOW}Checking Go code formatting...${RESET}"
	@UNFORMATTED=$$(gofmt -l .); \
	if [ -n "$$UNFORMATTED" ]; then \
		echo "${RED}The following files need formatting:${RESET}"; \
		echo "$$UNFORMATTED"; \
		exit 1; \
	else \
		echo "${GREEN}All files are properly formatted${RESET}"; \
	fi

pre-commit: lint-staged format-check ## Run pre-commit checks (lint staged files + format check)
	@echo "${GREEN}✅ Pre-commit checks passed${RESET}"

pre-push: lint-changed test-unit ## Run pre-push checks (lint changed files + unit tests)
	@echo "${GREEN}✅ Pre-push checks passed${RESET}"

check-all: format-check lint test-unit ## Run all code quality checks
	@echo "${GREEN}✅ All checks passed${RESET}"



lint-diff: ## Show diff of what lint-fix would change (dry run)
	@echo "${YELLOW}Showing what lint-fix would change...${RESET}"
	@CHANGED_FILES=$$(git diff --name-only --diff-filter=AM origin/main...HEAD | grep '\.go$$' | grep '^api/' || echo ""); \
	if [ -z "$$CHANGED_FILES" ]; then \
		echo "${GREEN}No Go files changed since main branch${RESET}"; \
	else \
		echo "${CYAN}Files that would be affected: $$CHANGED_FILES${RESET}"; \
		golangci-lint run --fix --diff --timeout=5m --new-from-rev=$$(git merge-base origin/main HEAD); \
	fi

## Git Integration Targets

git-hooks-install: ## Install git hooks for automated linting
	@echo "${YELLOW}Installing git hooks...${RESET}"
	@mkdir -p .git/hooks
	@echo '#!/bin/bash' > .git/hooks/pre-commit
	@echo 'cd api && make pre-commit' >> .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo '#!/bin/bash' > .git/hooks/pre-push
	@echo 'cd api && make pre-push' >> .git/hooks/pre-push
	@chmod +x .git/hooks/pre-push
	@echo "${GREEN}Git hooks installed successfully${RESET}"
	@echo "${CYAN}Pre-commit: runs lint-staged + format-check${RESET}"
	@echo "${CYAN}Pre-push: runs lint-changed + unit tests${RESET}"

git-hooks-remove: ## Remove git hooks
	@echo "${YELLOW}Removing git hooks...${RESET}"
	@rm -f .git/hooks/pre-commit .git/hooks/pre-push
	@echo "${GREEN}Git hooks removed${RESET}"

help: ## Show this help.
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z_-]+:.*?##.*$$/) {printf "    ${YELLOW}%-20s${GREEN}%s${RESET}\n", $$1, $$2} \
		else if (/^## .*$$/) {printf "  ${CYAN}%s${RESET}\n", substr($$1,4)} \
		}' $(MAKEFILE_LIST)	