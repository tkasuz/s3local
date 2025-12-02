.PHONY: all build run clean test help db-migrate db-create db-reset sqlc mocks

# Variables
BINARY_NAME=s3local
BIN_DIR=bin
CMD_DIR=cmd/s3local
DB_PATH=s3local.db
MIGRATIONS_DIR=internal/db/migrations

# Default target
all: build

## help: Show this help message
help:
	@echo 'Usage:'
	@echo '  make [target]'
	@echo ''
	@echo 'Available targets:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## build: Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/$(BINARY_NAME) ./$(CMD_DIR)
	@echo "Build complete: $(BIN_DIR)/$(BINARY_NAME)"

## run: Run the application
run: build
	@echo "Starting $(BINARY_NAME)..."
	@./$(BIN_DIR)/$(BINARY_NAME)

## clean: Remove build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BIN_DIR)
	@go clean
	@echo "Clean complete"

## test: Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

## test-cover: Run tests with coverage
test-cover:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## tidy: Tidy go modules
tidy:
	@echo "Tidying go modules..."
	@go mod tidy
	@echo "Tidy complete"

## fmt: Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Format complete"

## lint: Run linters
lint:
	@echo "Running linters..."
	@golangci-lint run
	@echo "Lint complete"

## dev: Run in development mode with hot reload (requires air)
dev:
	@command -v air > /dev/null || (echo "air not found, install with: go install github.com/air-verse/air@latest" && exit 1)
	@air

## sqlc: Generate Go code from SQL using sqlc
sqlc:
	@echo "Generating code from SQL..."
	@go tool sqlc generate
	@echo "SQL code generation complete"

## mocks: Generate mocks for testing using mockery
mocks:
	@echo "Generating mocks..."
	@mockery
	@echo "Mock generation complete"

## db-migrate: Run database migrations
db-migrate:
	@echo "Running database migrations..."
	@migrate -path $(MIGRATIONS_DIR) -database "sqlite3://$(DB_PATH)" up
	@echo "Migrations complete"

## db-migrate-down: Rollback last migration
db-migrate-down:
	@echo "Rolling back last migration..."
	@migrate -path $(MIGRATIONS_DIR) -database "sqlite3://$(DB_PATH)" down 1
	@echo "Rollback complete"

## db-create: Create a new migration (usage: make db-create name=migration_name)
db-create:
	@if [ -z "$(name)" ]; then echo "Error: name is required. Usage: make db-create name=migration_name"; exit 1; fi
	@echo "Creating new migration: $(name)..."
	@migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)
	@echo "Migration created"

## db-reset: Reset the database (WARNING: destructive)
db-reset:
	@echo "Resetting database..."
	@rm -f $(DB_PATH)
	@echo "Database reset complete"

## db-setup: Create database and run migrations
db-setup: db-reset db-migrate
	@echo "Database setup complete"
