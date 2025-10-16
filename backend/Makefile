# Makefile for Elang Backend

.PHONY: help test test-verbose test-coverage test-race build run clean lint fmt install-tools setup dev

# Variables
BINARY_NAME=elang-backend
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html

# Help command
help: ## Display this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Installation
install-tools: ## Install development tools
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/cosmtrek/air@latest
	go get github.com/stretchr/testify
	go get gorm.io/driver/sqlite

setup: ## Setup project dependencies
	@echo "Setting up project..."
	go mod download
	go mod tidy

# Testing
test: ## Run all tests
	@echo "Running tests..."
	go test ./test/... -v

test-verbose: ## Run tests with verbose output
	@echo "Running tests with verbose output..."
	go test ./test/... -v -count=1

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	go test ./test/... -coverprofile=$(COVERAGE_FILE) -covermode=atomic
	go tool cover -func=$(COVERAGE_FILE)
	go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Coverage report generated: $(COVERAGE_HTML)"

test-race: ## Run tests with race detector
	@echo "Running tests with race detector..."
	go test ./test/... -race -v

test-repo: ## Run repository tests only
	@echo "Running repository tests..."
	go test ./test/repository/... -v

test-service: ## Run service tests only
	@echo "Running service tests..."
	go test ./test/services/... -v

test-usecase: ## Run usecase tests only
	@echo "Running usecase tests..."
	go test ./test/usecase/... -v

# Building
build: ## Build the application
	@echo "Building application..."
	go build -o $(BINARY_NAME) cmd/main.go
	@echo "Build complete: $(BINARY_NAME)"

build-linux: ## Build for Linux
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME)-linux cmd/main.go

build-windows: ## Build for Windows
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 go build -o $(BINARY_NAME).exe cmd/main.go

build-mac: ## Build for macOS
	@echo "Building for macOS..."
	GOOS=darwin GOARCH=amd64 go build -o $(BINARY_NAME)-mac cmd/main.go

# Running
run: ## Run the application
	@echo "Running application..."
	go run cmd/main.go

dev: ## Run with hot reload (requires air)
	@echo "Running with hot reload..."
	air

# Code Quality
lint: ## Run linter
	@echo "Running linter..."
	golangci-lint run ./...

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

# Cleaning
clean: ## Clean build artifacts and cache
	@echo "Cleaning..."
	rm -f $(BINARY_NAME) $(BINARY_NAME)-* $(BINARY_NAME).exe
	rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	go clean -cache -testcache -modcache

clean-test: ## Clean test cache
	@echo "Cleaning test cache..."
	go clean -testcache

# Docker
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t elang-backend:latest .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run -p 8080:8080 elang-backend:latest

# Database
db-migrate: ## Run database migrations (placeholder)
	@echo "Running database migrations..."
	# Add migration command here

db-seed: ## Seed database (placeholder)
	@echo "Seeding database..."
	# Add seed command here

# CI/CD
ci-test: ## Run tests in CI environment
	@echo "Running CI tests..."
	go test ./test/... -v -coverprofile=$(COVERAGE_FILE) -covermode=atomic

ci-lint: ## Run linting in CI environment
	@echo "Running CI linting..."
	golangci-lint run ./... --timeout 5m

# All-in-one commands
all: clean fmt lint test build ## Run fmt, lint, test, and build

check: fmt lint test ## Run formatting, linting, and tests

.DEFAULT_GOAL := help
