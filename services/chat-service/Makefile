.PHONY: help build run test clean proto docker-build docker-run deploy

# Variables
SERVICE_NAME = chat-service
BINARY_NAME = main
DOCKER_IMAGE = $(SERVICE_NAME):latest
GRPC_PORT = 8082
REST_PORT = 8083

# Default target
help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Build the service
build: ## Build the chat service binary
	@echo "Building $(SERVICE_NAME)..."
	@cd $(SERVICE_NAME) && go build -o $(BINARY_NAME) .

# Run the service locally
run: build ## Run the chat service locally
	@echo "Running $(SERVICE_NAME) on port $(GRPC_PORT)..."
	@cd $(SERVICE_NAME) && ./$(BINARY_NAME)

# Run in development mode
dev: ## Run the service in development mode with hot reload
	@echo "Running $(SERVICE_NAME) in development mode..."
	@cd $(SERVICE_NAME) && go run .

# Generate protobuf code
proto: ## Generate protobuf Go code
	@echo "Generating protobuf code..."
	@cd $(SERVICE_NAME) && protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=. --grpc-gateway_opt=paths=source_relative \
		proto/chat.proto

# Run tests
test: ## Run tests
	@echo "Running tests..."
	@cd $(SERVICE_NAME) && go test -v ./...

# Run tests with coverage
test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@cd $(SERVICE_NAME) && go test -v -coverprofile=coverage.out ./...
	@cd $(SERVICE_NAME) && go tool cover -html=coverage.out

# Clean build artifacts
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@cd $(SERVICE_NAME) && rm -f $(BINARY_NAME) coverage.out
	@cd $(SERVICE_NAME) && rm -f proto/*.pb.go proto/*.pb.gw.go

# Install dependencies
deps: ## Install Go dependencies
	@echo "Installing dependencies..."
	@cd $(SERVICE_NAME) && go mod download
	@cd $(SERVICE_NAME) && go mod tidy

# Docker targets
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE) -f $(SERVICE_NAME)/Dockerfile .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run -p $(GRPC_PORT):$(GRPC_PORT) -p $(REST_PORT):$(REST_PORT) \
		-e OPENAI_API_KEY=your-api-key-here \
		-e AUTH_SERVICE_HOST=host.docker.internal \
		-e AUTH_SERVICE_PORT=8081 \
		-e AUTH_SERVICE_TLS=false \
		$(DOCKER_IMAGE)

# Kubernetes/Helm targets
deploy: ## Deploy to Kubernetes using Helm
	@echo "Deploying to Kubernetes..."
	@cd ../deployment/staging && ./deploy-chat.sh

deploy-local: ## Deploy to local Kubernetes
	@echo "Deploying to local Kubernetes..."
	@cd ../deployment/staging && ./deploy-chat.sh -n default

undeploy: ## Remove from Kubernetes
	@echo "Removing from Kubernetes..."
	@helm uninstall chat-service

# Development setup
setup: deps proto ## Setup development environment
	@echo "Development environment setup complete!"

# Format code
fmt: ## Format Go code
	@echo "Formatting Go code..."
	@cd $(SERVICE_NAME) && go fmt ./...

# Lint code
lint: ## Lint Go code
	@echo "Linting Go code..."
	@cd $(SERVICE_NAME) && golangci-lint run

# Security scan
security: ## Run security scan
	@echo "Running security scan..."
	@cd $(SERVICE_NAME) && gosec ./...

# Generate mocks (if using mockgen)
mocks: ## Generate mocks for testing
	@echo "Generating mocks..."
	@cd $(SERVICE_NAME) && mockgen -source=internal/services/openai/client.go -destination=internal/services/openai/mock_client.go
	@cd $(SERVICE_NAME) && mockgen -source=internal/services/chat/service.go -destination=internal/services/chat/mock_service.go
