.PHONY: build run test clean docker-build docker-run help

# Default target
help:
	@echo "Available commands:"
	@echo "  build        - Build the application"
	@echo "  build-prod   - Build optimized production binary"
	@echo "  run          - Run the application locally"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  fmt          - Format code"
	@echo "  lint         - Lint code with go vet"
	@echo "  security     - Check for security vulnerabilities"
	@echo "  version      - Show Go version and module info"
	@echo "  deps         - Install/update dependencies"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run with Docker Compose"
	@echo "  db-up        - Start database with Docker Compose"
	@echo "  db-down      - Stop database with Docker Compose"

# Build the application
build:
	@echo "Building application..."
	cd auth-service && go build -o ../bin/go-starter-rest .

# Run the application locally
run:
	@echo "Running application..."
	cd auth-service && go run .

# Run tests
test:
	@echo "Running tests..."
	cd auth-service && go test ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf auth-service/go-starter-rest

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t go-starter-rest:latest .

# Run with Docker Compose
docker-run:
	@echo "Starting services with Docker Compose..."
	cd deployment/local && docker-compose up --build

# Start database only
db-up:
	@echo "Starting database..."
	cd deployment/local && docker-compose up -d postgres

# Stop database
db-down:
	@echo "Stopping database..."
	cd deployment/local && docker-compose down

# Install dependencies
deps:
	@echo "Installing dependencies..."
	cd auth-service && go mod tidy
	cd packages/auth && go mod tidy
	cd packages/logger && go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	cd auth-service && go fmt ./...
	cd packages/auth && go fmt ./...
	cd packages/logger && go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	cd auth-service && go vet ./...
	cd packages/auth && go vet ./...
	cd packages/logger && go vet ./...

# Check for security vulnerabilities
security:
	@echo "Checking for security vulnerabilities..."
	cd auth-service && go list -json -deps . | nancy sleuth
	cd packages/auth && go list -json -deps . | nancy sleuth
	cd packages/logger && go list -json -deps . | nancy sleuth

# Build with optimizations for production
build-prod:
	@echo "Building production binary..."
	cd auth-service && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ../bin/go-starter-rest .

# Show Go version
version:
	@echo "Go version:"
	@go version
	@echo "Go modules:"
	@cd auth-service && go mod graph | head -20
