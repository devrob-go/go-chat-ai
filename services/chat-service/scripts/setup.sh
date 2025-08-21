#!/bin/bash

# Setup Script for Chat Service
# This script helps set up the development environment

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

# Function to check prerequisites
check_prerequisites() {
    print_step "Checking prerequisites..."
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go 1.24.6 or later."
        exit 1
    fi
    
    # Check Go version
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    REQUIRED_VERSION="1.24.6"
    
    if [[ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]]; then
        print_error "Go version $GO_VERSION is too old. Required: $REQUIRED_VERSION or later."
        exit 1
    fi
    
    print_status "Go version $GO_VERSION is compatible."
    
    # Check if Docker is installed (for database)
    if ! command -v docker &> /dev/null; then
        print_warning "Docker is not installed. You'll need to set up PostgreSQL manually."
        DOCKER_AVAILABLE=false
    else
        print_status "Docker is available."
        DOCKER_AVAILABLE=true
    fi
    
    # Check if curl is installed
    if ! command -v curl &> /dev/null; then
        print_error "curl is not installed. Please install curl to test endpoints."
        exit 1
    fi
    
    print_status "Prerequisites check completed."
}

# Function to setup database with Docker
setup_database_docker() {
    if [ "$DOCKER_AVAILABLE" = false ]; then
        print_warning "Skipping Docker database setup."
        return
    fi
    
    print_step "Setting up PostgreSQL database with Docker..."
    
    # Check if PostgreSQL container is already running
    if docker ps | grep -q "postgres"; then
        print_status "PostgreSQL container is already running."
        return
    fi
    
    # Start PostgreSQL container
    docker run -d \
        --name chat-postgres \
        -e POSTGRES_USER=postgres \
        -e POSTGRES_PASSWORD=password \
        -e POSTGRES_DB=chat_db \
        -p 5432:5432 \
        postgres:15-alpine
    
    print_status "PostgreSQL container started."
    print_status "Waiting for database to be ready..."
    
    # Wait for database to be ready
    sleep 10
    
    # Test database connection
    if docker exec chat-postgres pg_isready -U postgres > /dev/null 2>&1; then
        print_status "Database is ready."
    else
        print_error "Database failed to start properly."
        exit 1
    fi
}

# Function to setup database manually
setup_database_manual() {
    print_step "Setting up database manually..."
    
    print_status "Please ensure you have PostgreSQL running with the following configuration:"
    echo "  Host: localhost"
    echo "  Port: 5432"
    echo "  User: postgres"
    echo "  Password: password"
    echo "  Database: chat_db"
    echo ""
    
    read -p "Press Enter when PostgreSQL is ready..."
    
    # Test database connection
    if command -v psql &> /dev/null; then
        if PGPASSWORD=password psql -h localhost -U postgres -d chat_db -c "SELECT 1;" > /dev/null 2>&1; then
            print_status "Database connection successful."
        else
            print_error "Database connection failed. Please check your PostgreSQL setup."
            exit 1
        fi
    else
        print_warning "psql not found. Please verify database connection manually."
    fi
}

# Function to install Go dependencies
install_dependencies() {
    print_step "Installing Go dependencies..."
    
    cd chat-service
    
    # Download dependencies
    go mod download
    
    # Tidy up modules
    go mod tidy
    
    cd ..
    
    print_status "Dependencies installed successfully."
}

# Function to generate protobuf code
generate_proto() {
    print_step "Generating protobuf code..."
    
    cd chat-service
    
    # Check if proto files exist
    if [[ ! -f "proto/chat.proto" ]]; then
        print_error "Proto files not found. Please ensure proto/chat.proto exists."
        exit 1
    fi
    
    # Check if protoc is available
    if ! command -v protoc &> /dev/null; then
        print_warning "protoc is not available. Protobuf generation will be skipped."
        print_status "You can install protoc from: https://grpc.io/docs/protoc-installation/"
        cd ..
        return
    fi
    
    # Generate Go code from protobuf
    protoc --go_out=. --go_opt=paths=source_relative \
        --go-grpc_out=. --go-grpc_opt=paths=source_relative \
        --grpc-gateway_out=. --grpc-gateway_opt=paths=source_relative \
        proto/chat.proto
    
    cd ..
    
    print_status "Protobuf code generated successfully."
}

# Function to create environment file
create_env_file() {
    print_step "Creating environment file..."
    
    cd chat-service
    
    if [ -f ".env" ]; then
        print_status ".env file already exists."
    else
        print_status "Creating .env file from template..."
        cp env.example .env
        print_warning "Please edit .env file and set your OpenAI API key."
    fi
    
    cd ..
}

# Function to build the service
build_service() {
    print_step "Building chat service..."
    
    cd chat-service
    
    # Build the service
    go build -o main .
    
    if [[ -f "main" ]]; then
        print_status "Build successful! Binary created: main"
    else
        print_error "Build failed. Binary not created."
        exit 1
    fi
    
    cd ..
}

# Function to run tests
run_tests() {
    print_step "Running tests..."
    
    cd chat-service
    
    # Run tests
    if go test -v ./...; then
        print_status "All tests passed!"
    else
        print_warning "Some tests failed. This may indicate issues that need attention."
    fi
    
    cd ..
}

# Function to show next steps
show_next_steps() {
    echo ""
    print_status "Setup completed successfully! 🎉"
    echo ""
    print_status "Next steps:"
    echo "1. Edit chat-service/.env file and set your OpenAI API key"
    echo "2. Start the service: cd chat-service && go run ."
    echo "3. Test the endpoints: ./scripts/test-endpoints.sh"
    echo "4. The service will be available at:"
    echo "   - gRPC: localhost:8082"
    echo "   - REST: localhost:8083"
    echo ""
    print_status "Database migrations will run automatically when the service starts."
    echo ""
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -d, --docker              Use Docker for database setup (default)"
    echo "  -m, --manual              Manual database setup"
    echo "  -t, --test                Run tests after setup"
    echo "  -h, --help                Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                         # Setup with Docker database"
    echo "  $0 -m                      # Setup with manual database"
    echo "  $0 -t                      # Setup and run tests"
}

# Parse command line arguments
USE_DOCKER=true
RUN_TESTS=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -d|--docker)
            USE_DOCKER=true
            shift
            ;;
        -m|--manual)
            USE_DOCKER=false
            shift
            ;;
        -t|--test)
            RUN_TESTS=true
            shift
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Main execution
main() {
    print_status "Starting chat service setup..."
    echo ""
    
    check_prerequisites
    install_dependencies
    generate_proto
    create_env_file
    
    if [ "$USE_DOCKER" = true ]; then
        setup_database_docker
    else
        setup_database_manual
    fi
    
    build_service
    
    if [ "$RUN_TESTS" = true ]; then
        run_tests
    fi
    
    show_next_steps
}

# Run main function
main "$@"
