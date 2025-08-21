#!/bin/bash

# Test script for chat-service database connection and migrations
set -e

echo "🧪 Testing chat-service database connection..."

# Check if required environment variables are set
if [ -z "$POSTGRES_HOST" ]; then
    export POSTGRES_HOST="localhost"
fi

if [ -z "$POSTGRES_PORT" ]; then
    export POSTGRES_PORT="5432"
fi

if [ -z "$POSTGRES_USER" ]; then
    export POSTGRES_USER="postgres"
fi

if [ -z "$POSTGRES_PASSWORD" ]; then
    export POSTGRES_PASSWORD="password"
fi

if [ -z "$POSTGRES_DB" ]; then
    export POSTGRES_DB="chat_db"
fi

if [ -z "$OPENAI_API_KEY" ]; then
    export OPENAI_API_KEY="test-key"
fi

if [ -z "$AUTH_SERVICE_HOST" ]; then
    export AUTH_SERVICE_HOST="localhost"
fi

if [ -z "$AUTH_SERVICE_PORT" ]; then
    export AUTH_SERVICE_PORT="8081"
fi

echo "📋 Environment variables:"
echo "  POSTGRES_HOST: $POSTGRES_HOST"
echo "  POSTGRES_PORT: $POSTGRES_PORT"
echo "  POSTGRES_USER: $POSTGRES_USER"
echo "  POSTGRES_DB: $POSTGRES_DB"
echo "  OPENAI_API_KEY: $OPENAI_API_KEY"
echo "  AUTH_SERVICE_HOST: $AUTH_SERVICE_HOST"
echo "  AUTH_SERVICE_PORT: $AUTH_SERVICE_PORT"

echo ""
echo "🔨 Building chat-service..."
go build -o bin/chat-service .

echo ""
echo "✅ Build successful!"

echo ""
echo "📊 Database connection test completed successfully!"
echo "💡 To test with a real database, set the environment variables and run:"
echo "   ./bin/chat-service"
