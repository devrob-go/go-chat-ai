# Chat Service

A robust chat service built with Go, gRPC, and REST API that provides chat functionality with AI integration using OpenAI.

## Features

- **Real-time Chat**: Send and receive messages with real-time streaming support
- **AI Integration**: Chat with OpenAI models (GPT-3.5-turbo, GPT-4, etc.)
- **Conversation Management**: Create, list, and manage chat conversations
- **Message History**: Retrieve and paginate chat history
- **Database Storage**: Persistent storage using PostgreSQL with automatic migrations
- **Authentication**: JWT-based authentication with auth service integration - **ALL ENDPOINTS PROTECTED**
- **Dual APIs**: Both gRPC and REST API endpoints
- **Rate Limiting**: Configurable rate limiting for API protection
- **TLS Support**: Optional TLS encryption for production deployments
- **Standardized Error Responses**: Consistent JSON error structure across all endpoints

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   REST Client   │    │   gRPC Client   │    │   Web Client    │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
                    ┌─────────────▼─────────────┐
                    │      Chat Service         │
                    │  ┌─────────────────────┐  │
                    │  │   REST Gateway      │  │
                    │  │   (Port 8083)       │  │
                    │  │   + Auth Check      │  │
                    │  └─────────────────────┘  │
                    │  ┌─────────────────────┐  │
                    │  │   gRPC Server       │  │
                    │  │   (Port 8082)       │  │
                    │  │   + Auth Check      │  │
                    │  └─────────────────────┘  │
                    └─────────────┬─────────────┘
                                 │
                    ┌─────────────▼─────────────┐
                    │      Auth Service         │
                    │  ┌─────────────────────┐  │
                    │  │   Token Validation  │  │
                    │  │   + User Context    │  │
                    │  └─────────────────────┘  │
                    └─────────────┬─────────────┘
                                 │
                    ┌─────────────▼─────────────┐
                    │      Chat Service         │
                    │  ┌─────────────────────┐  │
                    │  │   OpenAI Client     │  │
                    │  └─────────────────────┘  │
                    │  ┌─────────────────────┐  │
                    │  │   Storage Layer     │  │
                    │  │   (PostgreSQL)      │  │
                    │  └─────────────────────┘  │
                    └───────────────────────────┘
```

## Quick Start

### Prerequisites

- Go 1.24.6 or later
- PostgreSQL 15 or later
- OpenAI API key
- Docker (optional, for database setup)
- **Auth Service** running and accessible

### 1. Setup Development Environment

```bash
# Clone the repository
git clone <repository-url>
cd go-chat-ai

# Run the setup script
./chat-service/scripts/setup.sh

# Or manually setup
cd chat-service
go mod download
go mod tidy
```

### 2. Configure Environment

Copy the environment template and configure it:

```bash
cd chat-service
cp env.example .env
```

Edit `.env` file and set your configuration:

```bash
# Required: Set your OpenAI API key
OPENAI_API_KEY=your-actual-openai-api-key-here

# Required: Auth service configuration
AUTH_SERVICE_HOST=localhost
AUTH_SERVICE_PORT=8081
AUTH_SERVICE_TLS=false

# Database configuration (adjust if needed)
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=password
POSTGRES_DB=chat_db

# Service ports
APP_PORT=8082      # gRPC port
REST_PORT=8083     # REST port
```

### 3. Start the Service

```bash
cd chat-service
go run .
```

The service will start on:
- **gRPC**: `localhost:8082`
- **REST**: `localhost:8083`

### 4. Test the Endpoints

```bash
# Test all endpoints
./scripts/test-endpoints.sh

# Or test manually with curl
curl http://localhost:8083/health
```

## 🔐 Authentication

**All chat endpoints require authentication via JWT Bearer token.**

### Getting a JWT Token

1. **Sign up** (if you don't have an account):
```bash
curl -X POST http://localhost:8080/v1/auth/signup \
  -H "Content-Type: application/json" \
  -d '{"name": "Your Name", "email": "your@email.com", "password": "yourpassword"}'
```

2. **Sign in** to get your JWT token:
```bash
curl -X POST http://localhost:8080/v1/auth/signin \
  -H "Content-Type: application/json" \
  -d '{"email": "your@email.com", "password": "yourpassword"}'
```

3. **Use the access token** in the Authorization header:
```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8083/v1/chat/conversations
```

### Token Format
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

## 🚀 API Examples

### Create a Conversation
```bash
curl -X POST http://localhost:8083/v1/chat/conversations \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title": "My Chat"}'
```

### Send a Message
```bash
curl -X POST http://localhost:8083/v1/chat/message \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"message": "Hello!", "conversation_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8"}'
```

### Chat with AI
```bash
curl -X POST http://localhost:8083/v1/chat/ai \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"message": "Tell me a joke", "conversation_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8", "model": "gpt-3.5-turbo"}'
```

### Get Chat History
```bash
curl "http://localhost:8083/v1/chat/history/6ba7b810-9dad-11d1-80b4-00c04fd430c8?limit=50" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## API Endpoints

### REST API

#### Health Check (No Authentication Required)
```http
GET /health
GET /v1/health
GET /v1/health/direct
```

#### Chat Endpoints (Authentication Required)

**Send Message**
```http
POST /v1/chat/message
Authorization: Bearer YOUR_JWT_TOKEN
Content-Type: application/json

{
  "message": "Hello, world!",
  "conversation_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
}
```

**Chat with AI**
```http
POST /v1/chat/ai
Authorization: Bearer YOUR_JWT_TOKEN
Content-Type: application/json

{
  "message": "What is the capital of France?",
  "conversation_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
  "model": "gpt-3.5-turbo",
  "temperature": 0.7,
  "max_tokens": 1000
}
```

**Create Conversation**
```http
POST /v1/chat/conversations
Authorization: Bearer YOUR_JWT_TOKEN
Content-Type: application/json

{
  "title": "My First Chat"
}
```

**List Conversations**
```http
GET /v1/chat/conversations?limit=10&offset=0
Authorization: Bearer YOUR_JWT_TOKEN
```

**Get Chat History**
```http
GET /v1/chat/history/6ba7b810-9dad-11d1-80b4-00c04fd430c8?limit=50&offset=0
Authorization: Bearer YOUR_JWT_TOKEN
```

## 📝 Error Response Structure

All endpoints return standardized JSON error responses:

### Success Response
```json
{
  "message": {
    "id": "uuid-here",
    "user_id": "user-uuid-here",
    "content": "Message content",
    "role": "user",
    "created_at": "2025-08-20T15:00:00Z"
  },
  "conversation_id": "conversation-uuid-here",
  "is_ai_response": false
}
```

### Error Response
```json
{
  "error": "ERROR_TYPE",
  "message": "Human readable error message",
  "code": "HTTP_STATUS_CODE",
  "details": {
    "additional": "information"
  }
}
```

### Common Error Types

| Error Type | HTTP Code | Description |
|------------|-----------|-------------|
| `UNAUTHORIZED` | 401 | Invalid or missing JWT token |
| `VALIDATION_ERROR` | 400 | Request validation failed |
| `METHOD_NOT_ALLOWED` | 405 | HTTP method not supported |
| `INTERNAL_ERROR` | 500 | Server internal error |
| `NOT_FOUND` | 404 | Resource not found |

### Error Examples

**Unauthorized (No Token)**
```json
{
  "error": "UNAUTHORIZED",
  "message": "Unauthorized",
  "code": "401"
}
```

**Validation Error**
```json
{
  "error": "VALIDATION_ERROR",
  "message": "Validation error",
  "code": "400",
  "details": {
    "details": "message cannot be empty"
  }
}
```

**Method Not Allowed**
```json
{
  "error": "METHOD_NOT_ALLOWED",
  "message": "Method not allowed",
  "code": "405"
}
```

## UUID Validation

All endpoints that accept IDs require valid UUIDs in the standard format: `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`

**Valid UUID Examples:**
- `550e8400-e29b-41d4-a716-446655440000`
- `6ba7b810-9dad-11d1-80b4-00c04fd430c8`
- `7ba7b810-9dad-11d1-80b4-00c04fd430c8`

**Invalid UUID Examples (will be rejected):**
- `user123` (simple string)
- `invalid-uuid` (wrong format)
- `12345678-1234-1234-1234-123456789abc` (invalid characters)

**UUID Generation:**
```bash
# Generate test UUIDs
./scripts/generate-uuids.sh -t

# Generate a UUID for immediate use
./scripts/generate-uuids.sh -u

# Generate 10 random UUIDs
./scripts/generate-uuids.sh -n 10
```

### gRPC API

The service also provides gRPC endpoints for the same functionality:

- `SendMessage` - Send a message to a conversation
- `StreamMessages` - Stream messages in real-time
- `GetHistory` - Retrieve chat history
- `ChatWithAI` - Chat with OpenAI AI models
- `ListConversations` - List user conversations
- `CreateConversation` - Create a new conversation

**Note**: All gRPC endpoints also require authentication via the `UnaryAuthInterceptor` and `StreamAuthInterceptor`.

## Database Schema

### Conversations Table
```sql
CREATE TABLE conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    title VARCHAR(500) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### Messages Table
```sql
CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    conversation_id UUID NOT NULL,
    content TEXT NOT NULL,
    role VARCHAR(50) NOT NULL CHECK (role IN ('user', 'assistant', 'system')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
);
```

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `APP_ENV` | `development` | Application environment |
| `APP_PORT` | `8082` | gRPC server port |
| `REST_PORT` | `8083` | REST gateway port |
| `OPENAI_API_KEY` | - | **Required** OpenAI API key |
| `AUTH_SERVICE_HOST` | `localhost` | **Required** Auth service host |
| `AUTH_SERVICE_PORT` | `8081` | **Required** Auth service port |
| `AUTH_SERVICE_TLS` | `false` | Use TLS for auth service connection |
| `POSTGRES_HOST` | `localhost` | PostgreSQL host |
| `POSTGRES_PORT` | `5432` | PostgreSQL port |
| `POSTGRES_DB` | `chat_db` | PostgreSQL database name |
| `LOG_LEVEL` | `debug` | Logging level |

### OpenAI Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `OPENAI_MODEL` | `gpt-3.5-turbo` | Default OpenAI model |
| `OPENAI_MAX_TOKENS` | `1000` | Maximum tokens per response |
| `OPENAI_TEMPERATURE` | `0.7` | Response creativity (0-2) |
| `OPENAI_TIMEOUT` | `30` | API timeout in seconds |

## Development

### Project Structure
```
chat-service/
├── cmd/                    # Application entry points
├── configs/               # Configuration management
├── internal/              # Internal packages
│   ├── domain/           # Domain models and interfaces
│   ├── services/         # Business logic services
│   └── transport/        # gRPC handlers and interceptors
├── proto/                # Protocol buffer definitions
├── scripts/              # Utility scripts
├── server/               # Server implementation
├── storage/              # Database layer and migrations
└── utils/                # Utility functions
```

### Running Tests
```bash
# Run all tests
go test -v ./...

# Run tests with coverage
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific test
go test -v ./internal/services/chat
```

### Building
```bash
# Build binary
go build -o main .

# Build with specific flags
go build -ldflags="-s -w" -o main .

# Cross-compile for different platforms
GOOS=linux GOARCH=amd64 go build -o main .
```

### Docker
```bash
# Build image
docker build -t chat-service .

# Run container
docker run -p 8082:8082 -p 8083:8083 \
  -e OPENAI_API_KEY=your-key \
  -e AUTH_SERVICE_HOST=host.docker.internal \
  -e AUTH_SERVICE_PORT=8081 \
  chat-service
```

## Deployment

### Local Development
```bash
# Using Docker Compose
cd deployment/local
docker-compose up -d

# Or manually
./scripts/setup.sh
cd chat-service && go run .
```

### Kubernetes
```bash
# Deploy to staging
cd deployment/staging
./deploy-chat.sh

# Deploy to production
./deploy-chat.sh -e production
```

## Monitoring and Health Checks

### Health Endpoints
- `/health` - Basic health check
- `/v1/health` - Service health status
- `/v1/health/direct` - Direct health check

### Metrics
The service includes Prometheus metrics for:
- Request counts and durations
- Database connection status
- OpenAI API usage
- Error rates

## Security

### Authentication
- **JWT-based authentication** via auth service integration
- **Token validation** on ALL protected endpoints
- **User context injection** in requests (no need to pass user_id)
- **Automatic user isolation** - users can only access their own data

### Rate Limiting
- Configurable rate limiting per endpoint
- Default: 100 requests per minute per IP
- Customizable limits and windows

### TLS Support
- Optional TLS encryption
- Configurable TLS versions (1.2, 1.3)
- Certificate-based authentication

### Data Protection
- **User ID automatically extracted** from JWT token
- **No user_id required** in request bodies
- **Automatic conversation ownership validation**
- **Foreign key constraints** ensure data integrity

## Troubleshooting

### Common Issues

**Authentication Errors**
```bash
# Check if auth service is running
curl http://localhost:8080/health

# Verify JWT token is valid
# Ensure token hasn't expired
# Check Authorization header format
```

**Database Connection Failed**
```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# Check database logs
docker logs chat-postgres

# Verify connection details in .env file
```

**OpenAI API Errors**
```bash
# Verify API key is set
echo $OPENAI_API_KEY

# Check API key format and permissions
# Ensure sufficient credits in OpenAI account
```

**Service Won't Start**
```bash
# Check port availability
netstat -tulpn | grep :8082
netstat -tulpn | grep :8083

# Check logs for specific errors
go run . 2>&1 | grep ERROR
```

### Debug Mode
```bash
# Enable debug logging
export LOG_LEVEL=debug
export LOG_JSON_FORMAT=false

# Start service with verbose output
go run . -v
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For support and questions:
- Create an issue in the repository
- Check the troubleshooting section
- Review the configuration documentation

## Changelog

### v1.1.0 (Latest)
- **All endpoints now require authentication**
- **Removed user_id from request bodies** - automatically extracted from JWT
- **Standardized error response structure** across all endpoints
- **Enhanced validation** for conversations and user ownership
- **Fixed foreign key violations** in message creation
- **Improved UUID validation** and error handling
- **Better gRPC error messages** and validation

### v1.0.0
- Initial release
- Basic chat functionality
- OpenAI integration
- PostgreSQL storage
- gRPC and REST APIs
- Basic authentication support
