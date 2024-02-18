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
	curl -s https://packagecloud.io/install/repositories/golang-migrate/migrate/script.deb.sh | sudo bash
	sudo apt install migrate
	sudo apt install inotify-tools

migrate-up: ## Run database migrations
	@POSTGRESQL_URL=$$($(GODOTENV_PATH) -f .env sh -c 'echo postgres://$$POSTGRES_USER:$$POSTGRES_PASSWORD@localhost:$$POSTGRES_PORT/$$POSTGRES_DB?sslmode=disable'); \
	migrate -database $$POSTGRESQL_URL -path migrations up

migrate-down: ## Rollback the last database migration executed
	@POSTGRESQL_URL=$$($(GODOTENV_PATH) -f .env sh -c 'echo postgres://$$POSTGRES_USER:$$POSTGRES_PASSWORD@localhost:$$POSTGRES_PORT/$$POSTGRES_DB?sslmode=disable'); \
	migrate -database $$POSTGRESQL_URL -path migrations down 1

migrate-drop: ## Drop everything
	@POSTGRESQL_URL=$$($(GODOTENV_PATH) -f .env sh -c 'echo postgres://$$POSTGRES_USER:$$POSTGRES_PASSWORD@localhost:$$POSTGRES_PORT/$$POSTGRES_DB?sslmode=disable'); \
	migrate -database $$POSTGRESQL_URL -path migrations drop

run: ## run ./cmd/web
	$(GODOTENV_PATH) -f .env go run ./cmd/web

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
	go build -o bin/main ./cmd/web

clean: ## Remove build related file
	rm -fr ./bin

create-migration: ## Create a new database migration file with the provided file name
	@read -p "Enter the migration file name: " FILE_NAME; \
	if [ -z "$$FILE_NAME" ]; then \
		echo "Migration file name is required."; \
	else \
		migrate create -ext sql -dir migrations -seq $$FILE_NAME; \
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