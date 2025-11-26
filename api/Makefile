.PHONY: all build run clean test help db-migrate db-create db-reset sqlc mocks

# Variables
BINARY_NAME=s3local
BIN_DIR=bin
CMD_DIR=cmd/s3local
DB_PATH=s3local.db
MIGRATIONS_DIR=migrations

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
	@atlas migrate apply --env local --url "sqlite://$(DB_PATH)"
	@echo "Migrations complete"

## db-create: Create a new migration
db-create:
	@echo "Creating new migration..."
	@atlas migrate diff $(name) --env local
	@echo "Migration created"

## db-reset: Reset the database (WARNING: destructive)
db-reset:
	@echo "Resetting database..."
	@rm -f $(DB_PATH)
	@echo "Database reset complete"

## db-setup: Create database and run migrations
db-setup: db-reset db-migrate
	@echo "Database setup complete"

## run-sqlite: Run with SQLite storage backend
run-sqlite: build db-migrate
	@echo "Starting $(BINARY_NAME) with SQLite backend..."
	@STORAGE_BACKEND=sqlite DB_PATH=$(DB_PATH) ./$(BIN_DIR)/$(BINARY_NAME)

## run-memory: Run with in-memory storage backend
run-memory: build
	@echo "Starting $(BINARY_NAME) with in-memory backend..."
	@STORAGE_BACKEND=memory ./$(BIN_DIR)/$(BINARY_NAME)
