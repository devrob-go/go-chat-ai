#!/bin/bash

# Test Build Script for Chat Service
# This script tests if the chat service can be built successfully

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
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

# Function to check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
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
    
    # Check if protoc is installed
    if ! command -v protoc &> /dev/null; then
        print_warning "protoc is not installed. Protobuf generation will be skipped."
        SKIP_PROTO=true
    else
        print_status "protoc is available."
        SKIP_PROTO=false
    fi
    
    print_status "Prerequisites check completed."
}

# Function to install Go dependencies
install_dependencies() {
    print_status "Installing Go dependencies..."
    
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
    if [[ "$SKIP_PROTO" == "true" ]]; then
        print_warning "Skipping protobuf generation (protoc not available)."
        return
    fi
    
    # Check if proto files exist
    if [[ ! -f "proto/chat.proto" ]]; then
        print_error "Proto files not found. Please ensure proto/chat.proto exists."
        exit 1
    fi

    # Generate proto files
    print_status "Generating protobuf code..."
    protoc --go_out=. --go_opt=paths=source_relative \
        --go-grpc_out=. --go-grpc_opt=paths=source_relative \
        --grpc-gateway_out=. --grpc-gateway_opt=paths=source_relative \
        proto/chat.proto
    
    cd ..
    
    print_status "Protobuf code generated successfully."
}

# Function to build the service
build_service() {
    print_status "Building chat service..."
    
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

# Function to run basic tests
run_tests() {
    print_status "Running basic tests..."
    
    cd chat-service
    
    # Run tests
    if go test -v ./...; then
        print_status "All tests passed!"
    else
        print_warning "Some tests failed. This may indicate issues that need attention."
    fi
    
    cd ..
}

# Function to clean up
cleanup() {
    print_status "Cleaning up build artifacts..."
    
    cd chat-service
    
    # Remove binary
    if [[ -f "main" ]]; then
        rm main
        print_status "Build artifacts cleaned up."
    fi
    
    cd ..
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -c, --cleanup              Clean up build artifacts after testing"
    echo "  -t, --test                 Run tests after building"
    echo "  -h, --help                 Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                         # Build only"
    echo "  $0 -t                      # Build and test"
    echo "  $0 -c                      # Build and cleanup"
    echo "  $0 -t -c                   # Build, test, and cleanup"
}

# Parse command line arguments
RUN_TESTS=false
CLEANUP=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -t|--test)
            RUN_TESTS=true
            shift
            ;;
        -c|--cleanup)
            CLEANUP=true
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
    print_status "Starting chat service build test..."
    
    check_prerequisites
    install_dependencies
    generate_proto
    build_service
    
    if [[ "$RUN_TESTS" == "true" ]]; then
        run_tests
    fi
    
    if [[ "$CLEANUP" == "true" ]]; then
        cleanup
    fi
    
    print_status "Build test completed successfully!"
    
    echo ""
    echo "Next steps:"
    echo "1. The service can be built successfully"
    echo "2. Consider running: make test (for comprehensive testing)"
    echo "3. Consider running: make docker-build (for Docker image)"
    echo "4. The service is ready for development and deployment"
}

# Run main function
main "$@"
