#!/bin/bash

echo "Testing chat service startup..."

# Set environment variables for testing
export APP_ENV=development
export APP_PORT=8082
export REST_PORT=8083
export LOG_LEVEL=debug
export LOG_JSON_FORMAT=false
export TLS_ENABLED=false
export AUTH_SERVICE_HOST=localhost
export AUTH_SERVICE_PORT=8081
export AUTH_SERVICE_TLS=false
export OPENAI_API_KEY=test-key
export POSTGRES_USER=postgres
export POSTGRES_PASSWORD=password
export POSTGRES_DB=chat_db
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5432

echo "Environment variables set:"
echo "APP_ENV: $APP_ENV"
echo "APP_PORT: $APP_PORT"
echo "REST_PORT: $REST_PORT"
echo "AUTH_SERVICE_HOST: $AUTH_SERVICE_HOST"
echo "AUTH_SERVICE_PORT: $AUTH_SERVICE_PORT"

echo "Building chat service..."
go build -o bin/chat-service main.go

if [ $? -eq 0 ]; then
    echo "Build successful!"
    echo "Testing startup (will timeout after 10 seconds)..."
    
    # Start the service in background and capture PID
    timeout 10s ./bin/chat-service &
    SERVICE_PID=$!
    
    # Wait a moment for startup
    sleep 2
    
    # Check if process is still running
    if kill -0 $SERVICE_PID 2>/dev/null; then
        echo "Service started successfully! PID: $SERVICE_PID"
        echo "Stopping service..."
        kill $SERVICE_PID
        wait $SERVICE_PID 2>/dev/null
        echo "Service stopped successfully!"
        echo "✅ Startup test PASSED - No protobuf registry conflicts detected"
    else
        echo "❌ Service failed to start or crashed immediately"
        echo "Startup test FAILED"
        exit 1
    fi
else
    echo "❌ Build failed!"
    exit 1
fi
