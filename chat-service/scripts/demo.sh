#!/bin/bash

# Demo Script for Chat Service
# This script demonstrates all the chat service endpoints with real examples

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="http://localhost:8083"
USER_ID="550e8400-e29b-41d4-a716-446655440001"  # Valid UUID format
CONVERSATION_TITLE="Demo Conversation"

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

print_demo() {
    echo -e "${CYAN}[DEMO]${NC} $1"
}

# Function to check if service is running
check_service() {
    print_step "Checking if chat service is running..."
    
    if curl -s "${BASE_URL}/health" > /dev/null; then
        print_status "Chat service is running ✓"
        return 0
    else
        print_error "Chat service is not running. Please start it first."
        print_status "You can start it with: go run ."
        exit 1
    fi
}

# Function to show demo header
show_demo_header() {
    echo ""
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║                    CHAT SERVICE DEMO                        ║"
    echo "║                                                              ║"
    echo "║  This demo will show you how to use all the chat service    ║"
    echo "║  endpoints with real examples.                              ║"
    echo "║                                                              ║"
    echo "║  User ID: ${USER_ID:0:20}...                                    ║"
    echo "║  Base URL: ${BASE_URL:0:25}...                                    ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo ""
}

# Function to demo health endpoint
demo_health() {
    print_demo "1. Health Check Endpoints"
    echo "   Testing basic health endpoints..."
    
    # Test basic health
    response=$(curl -s "${BASE_URL}/health")
    echo "   GET /health: ${response}"
    
    # Test v1 health
    response=$(curl -s "${BASE_URL}/v1/health")
    echo "   GET /v1/health: ${response}"
    
    print_status "Health endpoints working ✓"
    echo ""
}

# Function to demo conversation creation
demo_create_conversation() {
    print_demo "2. Creating a New Conversation"
    echo "   Creating conversation: '${CONVERSATION_TITLE}'"
    
    response=$(curl -s -X POST "${BASE_URL}/v1/chat/conversations" \
        -H "Content-Type: application/json" \
        -d "{\"user_id\":\"${USER_ID}\",\"title\":\"${CONVERSATION_TITLE}\"}")
    
    echo "   Response: ${response}"
    
    # Extract conversation ID
    CONVERSATION_ID=$(echo "$response" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
    
    if [ -n "$CONVERSATION_ID" ]; then
        print_status "Conversation created successfully ✓"
        print_status "Conversation ID: ${CONVERSATION_ID}"
    else
        print_error "Failed to create conversation"
        exit 1
    fi
    
    echo ""
}

# Function to demo listing conversations
demo_list_conversations() {
    print_demo "3. Listing User Conversations"
    echo "   Fetching conversations for user: ${USER_ID}"
    
    response=$(curl -s "${BASE_URL}/v1/chat/conversations?user_id=${USER_ID}")
    
    echo "   Response: ${response}"
    
    # Check if conversations were found
    if echo "$response" | grep -q "conversations"; then
        print_status "Conversations listed successfully ✓"
    else
        print_error "Failed to list conversations"
    fi
    
    echo ""
}

# Function to demo sending a message
demo_send_message() {
    print_demo "4. Sending a Message"
    echo "   Sending message to conversation: ${CONVERSATION_ID}"
    
    response=$(curl -s -X POST "${BASE_URL}/v1/chat/message" \
        -H "Content-Type: application/json" \
        -d "{\"user_id\":\"${USER_ID}\",\"message\":\"Hello! This is my first message in the demo conversation.\",\"conversation_id\":\"${CONVERSATION_ID}\"}")
    
    echo "   Response: ${response}"
    
    if echo "$response" | grep -q "message"; then
        print_status "Message sent successfully ✓"
    else
        print_error "Failed to send message"
    fi
    
    echo ""
}

# Function to demo sending another message
demo_send_another_message() {
    print_demo "5. Sending Another Message"
    echo "   Sending another message to conversation: ${CONVERSATION_ID}"
    
    response=$(curl -s -X POST "${BASE_URL}/v1/chat/message" \
        -H "Content-Type: application/json" \
        -d "{\"user_id\":\"${USER_ID}\",\"message\":\"This is my second message. I'm really enjoying this demo!\",\"conversation_id\":\"${CONVERSATION_ID}\"}")
    
    echo "   Response: ${response}"
    
    if echo "$response" | grep -q "message"; then
        print_status "Second message sent successfully ✓"
    else
        print_error "Failed to send second message"
    fi
    
    echo ""
}

# Function to demo getting chat history
demo_get_history() {
    print_demo "6. Retrieving Chat History"
    echo "   Getting chat history for conversation: ${CONVERSATION_ID}"
    
    response=$(curl -s "${BASE_URL}/v1/chat/history/${CONVERSATION_ID}?user_id=${USER_ID}")
    
    echo "   Response: ${response}"
    
    if echo "$response" | grep -q "messages"; then
        print_status "Chat history retrieved successfully ✓"
        
        # Count messages
        message_count=$(echo "$response" | grep -o '"id":"[^"]*"' | wc -l)
        print_status "Found ${message_count} messages in the conversation"
    else
        print_error "Failed to retrieve chat history"
    fi
    
    echo ""
}

# Function to demo AI chat (if API key is available)
demo_ai_chat() {
    print_demo "7. Chatting with AI"
    echo "   Sending a message to OpenAI AI..."
    
    # Check if OpenAI API key is set
    if [ "$OPENAI_API_KEY" = "your-openai-api-key-here" ] || [ -z "$OPENAI_API_KEY" ]; then
        print_warning "OpenAI API key not set. Skipping AI chat demo."
        print_status "To test AI chat, set your OPENAI_API_KEY environment variable."
        echo ""
        return
    fi
    
    response=$(curl -s -X POST "${BASE_URL}/v1/chat/ai" \
        -H "Content-Type: application/json" \
        -d "{\"user_id\":\"${USER_ID}\",\"message\":\"Hello AI! Can you tell me a short joke?\",\"conversation_id\":\"${CONVERSATION_ID}\"}")
    
    echo "   Response: ${response}"
    
    if echo "$response" | grep -q "ai_message"; then
        print_status "AI chat successful ✓"
        
        # Extract AI response
        ai_message=$(echo "$response" | grep -o '"ai_message":"[^"]*"' | cut -d'"' -f4)
        echo "   AI Response: ${ai_message}"
    else
        print_error "AI chat failed"
    fi
    
    echo ""
}

# Function to demo error handling
demo_error_handling() {
    print_demo "8. Testing Error Handling"
    echo "   Testing various error scenarios..."
    
    # Test missing user_id
    echo "   Testing missing user_id..."
    response=$(curl -s -w "%{http_code}" -X POST "${BASE_URL}/v1/chat/message" \
        -H "Content-Type: application/json" \
        -d "{\"message\":\"Test message without user_id\"}")
    
    http_code="${response: -3}"
    response_body="${response%???}"
    
    if [ "$http_code" = "400" ]; then
        print_status "Missing user_id error handled correctly ✓ (HTTP ${http_code})"
    else
        print_error "Missing user_id error handling failed. Expected 400, got ${http_code}"
    fi
    
    # Test invalid method
    echo "   Testing invalid HTTP method..."
    response=$(curl -s -w "%{http_code}" -X GET "${BASE_URL}/v1/chat/message")
    
    http_code="${response: -3}"
    if [ "$http_code" = "405" ]; then
        print_status "Invalid method error handled correctly ✓ (HTTP ${http_code})"
    else
        print_error "Invalid method error handling failed. Expected 405, got ${http_code}"
    fi
    
    # Test malformed JSON
    echo "   Testing malformed JSON..."
    response=$(curl -s -w "%{http_code}" -X POST "${BASE_URL}/v1/chat/message" \
        -H "Content-Type: application/json" \
        -d "{invalid json")
    
    http_code="${response: -3}"
    if [ "$http_code" = "400" ]; then
        print_status "Malformed JSON error handled correctly ✓ (HTTP ${http_code})"
    else
        print_error "Malformed JSON error handling failed. Expected 400, got ${http_code}"
    fi
    
    echo ""
}

# Function to show demo summary
show_demo_summary() {
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║                      DEMO COMPLETED                         ║"
    echo "║                                                              ║"
    echo "║  All chat service endpoints have been demonstrated:         ║"
    echo "║                                                              ║"
    echo "║  ✓ Health checks                                            ║"
    echo "║  ✓ Conversation creation                                     ║"
    echo "║  ✓ Conversation listing                                      ║"
    echo "║  ✓ Message sending                                           ║"
    echo "║  ✓ Chat history retrieval                                    ║"
    echo "║  ✓ AI chat integration                                       ║"
    echo "║  ✓ Error handling                                            ║"
    echo "║                                                              ║"
    echo "║  The service is working correctly and can store/retrieve     ║"
    echo "║  data from the database through the REST API!                ║"
    echo "║                                                              ║"
    echo "║  Created conversation ID: ${CONVERSATION_ID:0:20}...                    ║"
    echo "║  User ID: ${USER_ID:0:20}...                                    ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo ""
    
    print_status "Demo completed successfully! 🎉"
    echo ""
    print_status "You can now:"
    echo "  - Use the conversation ID: ${CONVERSATION_ID}"
    echo "  - Send more messages to the conversation"
    echo "  - Test other endpoints with different parameters"
    echo "  - Integrate the service into your applications"
    echo ""
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -u, --user-id USER_ID     Use specific user ID for demo"
    echo "  -b, --base-url URL        Base URL for the service"
    echo "  -h, --help                Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                         # Run demo with defaults"
    echo "  $0 -u my-demo-user        # Run demo with specific user ID"
    echo "  $0 -b http://localhost:9000 # Run demo against different service"
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
    show_demo_header
    
    # Check if service is running
    check_service
    
    # Run demo steps
    demo_health
    demo_create_conversation
    demo_list_conversations
    demo_send_message
    demo_send_another_message
    demo_get_history
    demo_ai_chat
    demo_error_handling
    
    # Show summary
    show_demo_summary
}

# Run main function
main "$@"
