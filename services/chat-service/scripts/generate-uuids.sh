#!/bin/bash

# Generate UUIDs Script for Chat Service
# This script generates valid UUIDs for testing purposes

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_uuid() {
    echo -e "${CYAN}[UUID]${NC} $1"
}

print_usage() {
    echo -e "${BLUE}[USAGE]${NC} $1"
}

# Function to generate a single UUID
generate_uuid() {
    if command -v uuidgen &> /dev/null; then
        uuidgen
    elif command -v python3 &> /dev/null; then
        python3 -c "import uuid; print(str(uuid.uuid4()))"
    elif command -v python &> /dev/null; then
        python -c "import uuid; print(str(uuid.uuid4()))"
    else
        # Fallback: generate a simple UUID-like string
        printf "%08x-%04x-%04x-%04x-%012x\n" \
            $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM$RANDOM
    fi
}

# Function to generate multiple UUIDs
generate_multiple_uuids() {
    local count=${1:-5}
    print_status "Generating $count UUIDs..."
    echo ""
    
    for i in $(seq 1 $count); do
        uuid=$(generate_uuid)
        print_uuid "$uuid"
    done
    echo ""
}

# Function to generate UUIDs for specific purposes
generate_test_uuids() {
    print_status "Generating UUIDs for testing..."
    echo ""
    
    print_status "User IDs:"
    print_uuid "550e8400-e29b-41d4-a716-446655440000"  # Test User 1
    print_uuid "550e8400-e29b-41d4-a716-446655440001"  # Test User 2
    print_uuid "550e8400-e29b-41d4-a716-446655440002"  # Test User 3
    echo ""
    
    print_status "Conversation IDs:"
    print_uuid "6ba7b810-9dad-11d1-80b4-00c04fd430c8"  # Test Conversation 1
    print_uuid "6ba7b811-9dad-11d1-80b4-00c04fd430c8"  # Test Conversation 2
    print_uuid "6ba7b812-9dad-11d1-80b4-00c04fd430c8"  # Test Conversation 3
    echo ""
    
    print_status "Message IDs:"
    print_uuid "7ba7b810-9dad-11d1-80b4-00c04fd430c8"  # Test Message 1
    print_uuid "7ba7b811-9dad-11d1-80b4-00c04fd430c8"  # Test Message 2
    print_uuid "7ba7b812-9dad-11d1-80b4-00c04fd430c8"  # Test Message 3
    echo ""
}

# Function to generate a UUID for immediate use
generate_for_use() {
    local uuid=$(generate_uuid)
    print_status "Generated UUID for immediate use:"
    print_uuid "$uuid"
    echo ""
    print_usage "Copy this UUID to use in your API calls"
    echo ""
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -n, --number N        Generate N UUIDs (default: 5)"
    echo "  -t, --test            Generate predefined test UUIDs"
    echo "  -u, --use             Generate one UUID for immediate use"
    echo "  -h, --help            Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                     # Generate 5 random UUIDs"
    echo "  $0 -n 10              # Generate 10 random UUIDs"
    echo "  $0 -t                  # Generate test UUIDs"
    echo "  $0 -u                  # Generate one UUID for use"
    echo ""
    echo "UUIDs can be used for:"
    echo "  - user_id in API calls"
    echo "  - conversation_id in API calls"
    echo "  - Testing and development"
    echo "  - Database seeding"
}

# Parse command line arguments
GENERATE_COUNT=5
GENERATE_TEST=false
GENERATE_USE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -n|--number)
            GENERATE_COUNT="$2"
            shift 2
            ;;
        -t|--test)
            GENERATE_TEST=true
            shift
            ;;
        -u|--use)
            GENERATE_USE=true
            shift
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        *)
            print_status "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Main execution
main() {
    if [ "$GENERATE_TEST" = true ]; then
        generate_test_uuids
    elif [ "$GENERATE_USE" = true ]; then
        generate_for_use
    else
        generate_multiple_uuids "$GENERATE_COUNT"
    fi
    
    print_status "UUID generation completed!"
    echo ""
    print_usage "Use these UUIDs in your API calls:"
    print_usage "curl -X POST http://localhost:8083/v1/chat/conversations \\"
    print_usage "  -H \"Content-Type: application/json\" \\"
    print_usage "  -d '{\"user_id\":\"<UUID>\",\"title\":\"My Chat\"}'"
}

# Run main function
main "$@"
