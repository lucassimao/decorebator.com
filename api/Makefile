# Load environment variables from the .env file
include .env

GREEN  		  := $(shell tput -Txterm setaf 2)
YELLOW 		  := $(shell tput -Txterm setaf 3)
WHITE  		  := $(shell tput -Txterm setaf 7)
CYAN   		  := $(shell tput -Txterm setaf 6)
RESET  		  := $(shell tput -Txterm sgr0)
GODOTENV_PATH := ~/go/bin/godotenv
DIR_TO_WATCH  := $(PWD)

all: help

setup: ## Install all dev and build dependencies
	go install github.com/joho/godotenv/cmd/godotenv@latest
	go install github.com/go-delve/delve/cmd/dlv@latest
	sudo apt-get install -y inotify-tools migrate

migrate-up: ## Run database migrations
	@POSTGRESQL_URL=$$($(GODOTENV_PATH) -f .env sh -c 'echo postgres://$$POSTGRES_USER:$$POSTGRES_PASSWORD@localhost:$$POSTGRES_PORT/$$POSTGRES_DB?sslmode=disable'); \
	migrate -database $$POSTGRESQL_URL -path cmd/migrate/migrations up

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
	go test -v -short ./tests/integration/...

test-clean: ## Clean up test environment (stop containers, remove volumes)
	@echo "Cleaning up test environment..."
	./scripts/run-tests.sh clean

migrate-up-test: ## Run database migrations for test database
	@POSTGRESQL_URL=$$($(GODOTENV_PATH) -f .env.test sh -c 'echo $$TEST_DATABASE_URL'); \
	migrate -database $$POSTGRESQL_URL -path cmd/migrate/migrations up

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

create-migration: ## Create a new database migration file with the provided file name
	@read -p "Enter the migration file name: " FILE_NAME; \
	if [ -z "$$FILE_NAME" ]; then \
		echo "Migration file name is required."; \
	else \
		migrate create -ext sql -dir cmd/migrate/migrations -seq $$FILE_NAME; \
	fi	

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