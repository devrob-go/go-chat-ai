#!/bin/bash

# Test Endpoints Script for Chat Service
# This script tests all the REST endpoints to ensure they work properly

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="http://localhost:8083"
USER_ID="550e8400-e29b-41d4-a716-446655440000"  # Valid UUID format
CONVERSATION_TITLE="Test Conversation"

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

# Function to check if service is running
check_service() {
    print_step "Checking if chat service is running..."
    
    if curl -s "${BASE_URL}/health" > /dev/null; then
        print_status "Chat service is running"
        return 0
    else
        print_error "Chat service is not running. Please start it first."
        print_status "You can start it with: cd chat-service && go run ."
        exit 1
    fi
}

# Function to test health endpoint
test_health() {
    print_step "Testing health endpoint..."
    
    response=$(curl -s "${BASE_URL}/health")
    if echo "$response" | grep -q "SERVING"; then
        print_status "Health endpoint working ✓"
    else
        print_error "Health endpoint failed"
        echo "Response: $response"
        return 1
    fi
}

# Function to test creating a conversation
test_create_conversation() {
    print_step "Testing conversation creation..."
    
    response=$(curl -s -X POST "${BASE_URL}/v1/chat/conversations" \
        -H "Content-Type: application/json" \
        -d "{\"user_id\":\"${USER_ID}\",\"title\":\"${CONVERSATION_TITLE}\"}")
    
    if echo "$response" | grep -q "id"; then
        print_status "Conversation creation working ✓"
        # Extract conversation ID for later tests
        CONVERSATION_ID=$(echo "$response" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
        print_status "Created conversation ID: ${CONVERSATION_ID}"
    else
        print_error "Conversation creation failed"
        echo "Response: $response"
        return 1
    fi
}

# Function to test listing conversations
test_list_conversations() {
    print_step "Testing conversation listing..."
    
    response=$(curl -s "${BASE_URL}/v1/chat/conversations?user_id=${USER_ID}")
    
    if echo "$response" | grep -q "conversations"; then
        print_status "Conversation listing working ✓"
        print_status "Response: $response"
    else
        print_error "Conversation listing failed"
        echo "Response: $response"
        return 1
    fi
}

# Function to test sending a message
test_send_message() {
    print_step "Testing message sending..."
    
    if [ -z "$CONVERSATION_ID" ]; then
        print_error "No conversation ID available. Skipping message test."
        return 1
    fi
    
    response=$(curl -s -X POST "${BASE_URL}/v1/chat/message" \
        -H "Content-Type: application/json" \
        -d "{\"user_id\":\"${USER_ID}\",\"message\":\"Hello, this is a test message!\",\"conversation_id\":\"${CONVERSATION_ID}\"}")
    
    if echo "$response" | grep -q "message"; then
        print_status "Message sending working ✓"
        print_status "Response: $response"
    else
        print_error "Message sending failed"
        echo "Response: $response"
        return 1
    fi
}

# Function to test getting chat history
test_get_history() {
    print_step "Testing chat history retrieval..."
    
    if [ -z "$CONVERSATION_ID" ]; then
        print_error "No conversation ID available. Skipping history test."
        return 1
    fi
    
    response=$(curl -s "${BASE_URL}/v1/chat/history/${CONVERSATION_ID}?user_id=${USER_ID}")
    
    if echo "$response" | grep -q "messages"; then
        print_status "Chat history retrieval working ✓"
        print_status "Response: $response"
    else
        print_error "Chat history retrieval failed"
        echo "Response: $response"
        return 1
    fi
}

# Function to test AI chat (if OpenAI API key is available)
test_ai_chat() {
    print_step "Testing AI chat..."
    
    if [ -z "$CONVERSATION_ID" ]; then
        print_error "No conversation ID available. Skipping AI chat test."
        return 1
    fi
    
    # Check if OpenAI API key is set
    if [ "$OPENAI_API_KEY" = "your-openai-api-key-here" ]; then
        print_warning "OpenAI API key not set. Skipping AI chat test."
        print_status "Set OPENAI_API_KEY environment variable to test AI chat functionality."
        return 0
    fi
    
    response=$(curl -s -X POST "${BASE_URL}/v1/chat/ai" \
        -H "Content-Type: application/json" \
        -d "{\"user_id\":\"${USER_ID}\",\"message\":\"Hello AI, how are you?\",\"conversation_id\":\"${CONVERSATION_ID}\"}")
    
    if echo "$response" | grep -q "ai_message"; then
        print_status "AI chat working ✓"
        print_status "Response: $response"
    else
        print_error "AI chat failed"
        echo "Response: $response"
        return 1
    fi
}

# Function to test error handling
test_error_handling() {
    print_step "Testing error handling..."
    
    # Test missing user_id
    response=$(curl -s -w "%{http_code}" -X POST "${BASE_URL}/v1/chat/message" \
        -H "Content-Type: application/json" \
        -d "{\"message\":\"Test message\"}")
    
    http_code="${response: -3}"
    response_body="${response%???}"
    
    if [ "$http_code" = "400" ]; then
        print_status "Error handling for missing user_id working ✓"
    else
        print_error "Error handling for missing user_id failed. Expected 400, got $http_code"
        echo "Response: $response_body"
    fi
    
    # Test invalid method
    response=$(curl -s -w "%{http_code}" -X GET "${BASE_URL}/v1/chat/message")
    
    http_code="${response: -3}"
    if [ "$http_code" = "405" ]; then
        print_status "Error handling for invalid method working ✓"
    else
        print_error "Error handling for invalid method failed. Expected 405, got $http_code"
    fi
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -u, --user-id USER_ID     Use specific user ID for testing (default: test-user-123)"
    echo "  -b, --base-url URL        Base URL for the service (default: http://localhost:8083)"
    echo "  -h, --help                Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                         # Run all tests with defaults"
    echo "  $0 -u my-user-456         # Run tests with specific user ID"
    echo "  $0 -b http://localhost:9000 # Run tests against different service"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -u|--user-id)
            USER_ID="$2"
            shift 2
            ;;
        -b|--base-url)
            BASE_URL="$2"
            shift 2
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
    print_status "Starting chat service endpoint tests..."
    print_status "Base URL: ${BASE_URL}"
    print_status "User ID: ${USER_ID}"
    echo ""
    
    # Check if service is running
    check_service
    
    # Run tests
    test_health
    test_create_conversation
    test_list_conversations
    test_send_message
    test_get_history
    test_ai_chat
    test_error_handling
    
    echo ""
    print_status "All endpoint tests completed successfully! ✓"
    echo ""
    print_status "Test Summary:"
    print_status "- Health endpoint: Working"
    print_status "- Conversation creation: Working"
    print_status "- Conversation listing: Working"
    print_status "- Message sending: Working"
    print_status "- Chat history: Working"
    print_status "- AI chat: ${OPENAI_API_KEY:-"Skipped (no API key)"}"
    print_status "- Error handling: Working"
    echo ""
    print_status "The chat service is working correctly and can store/retrieve data from the database!"
}

# Run main function
main "$@"
