# Go Chat AI - Microservices Architecture

A production-ready microservices platform built with Go, featuring authentication, user management, and AI-powered chat functionality. The platform follows clean architecture principles with comprehensive test coverage and Kubernetes deployment support.

## 🏗️ Architecture Overview

This is a **microservices monorepo** built with Go modules and Go workspaces, featuring:

- **Auth Service**: Complete authentication and user management with gRPC + REST APIs
- **Chat Service**: AI-powered chat functionality with OpenAI integration
- **Shared Packages**: Reusable authentication middleware and structured logging
- **Kubernetes Deployment**: Helm charts for staging and production environments
- **Local Development**: Docker Compose setup for easy local development

## 🚀 Services

### Auth Service (`auth-service/`)
- **Ports**: gRPC 8080, REST 8081
- **Features**: User registration, authentication, JWT token management
- **Database**: PostgreSQL with automatic migrations
- **Architecture**: Clean architecture with domain, service, and transport layers

### Chat Service (`chat-service/`)
- **Ports**: gRPC 8082, REST 8083
- **Features**: AI chat conversations, OpenAI integration, conversation management
- **Database**: PostgreSQL with conversation and message storage
- **Security**: JWT authentication with auth service integration

### Shared Packages (`packages/`)
- **`auth/`**: JWT middleware and authentication utilities
- **`logger/`**: Structured logging with correlation IDs and multiple output formats

## 🏛️ Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              Client Applications                            │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │   REST      │  │   gRPC      │  │   Web       │  │   Mobile/CLI        │ │
│  │   Client    │  │   Client    │  │   Client    │  │   Client            │ │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────────────┘ │
└─────────────────────┬───────────────────────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────────────────────┐
│                              API Gateway Layer                              │
│  ┌────────────────────────────────────────────────────────────────────────┐ │
│  │  Auth Service (Ports: gRPC 8080, REST 8081)                            │ │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐    │ │
│  │  │   gRPC      │  │   REST      │  │   Auth      │  │   Health    │    │ │
│  │  │   Server    │  │   Gateway   │  │   Server    │  │   Server    │    │ │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘    │ │
│  └────────────────────────────────────────────────────────────────────────┘ │
│  ┌────────────────────────────────────────────────────────────────────────┐ │
│  │  Chat Service (Ports: gRPC 8082, REST 8083)                            │ │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐    │ │
│  │  │   gRPC      │  │   REST      │  │   Chat      │  │   OpenAI    │    │ │
│  │  │   Server    │  │   Gateway   │  │   Server    │  │   Client    │    │ │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘    │ │
│  └────────────────────────────────────────────────────────────────────────┘ │
└─────────────────────┬───────────────────────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────────────────────┐
│                              Service Layer                                  │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │   Auth      │  │   User      │  │   Chat      │  │   OpenAI    │         │
│  │   Service   │  │   Service   │  │   Service   │  │   Service   │         │
│  │ • SignUp    │  │ • GetUsers  │  │ • Chat      │  │ • GPT-3.5   │         │
│  │ • SignIn    │  │ • Pagination│  │ • History   │  │ • GPT-4     │         │
│  │ • SignOut   │  │ • Validation│  │ • Stream    │  │ • Streaming │         │
│  │ • Refresh   │  └─────────────┘  └─────────────┘  └─────────────┘         │
│  │ • Revoke    │                                                            │
│  └─────────────┘                                                            │
└─────────────────────┬───────────────────────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────────────────────┐
│                              Domain Layer                                   │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │    User     │  │   Auth      │  │   Chat      │  │ Validation  │         │
│  │ • Models    │  │ • Tokens    │  │ • Messages  │  │ • Rules     │         │
│  │ • Credentials│  │ • JWT      │  │ • Conversations│ • Errors    │         │
│  │ • Validation│  │ • Security  │  │ • AI Models │  │ • Messages  │         │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘         │
└─────────────────────┬───────────────────────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────────────────────┐
│                          Infrastructure Layer                               │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │ PostgreSQL  │  │   Redis     │  │   Logger    │  │   OpenAI    │         │
│  │ • Migrations│  │ • Caching   │  │ • Structured│  │ • API       │         │
│  │ • Connection│  │ • Sessions  │  │ • Correlation│  │ • Rate Limit│        │
│  │ • Pool      │  │ • Rate Limit│  │ • Levels    │  │ • Models    │         │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘         │
└─────────────────────────────────────────────────────────────────────────────┘
```

## 🚀 Quick Start

### Prerequisites

- Go 1.24.6+
- Docker & Docker Compose
- Protocol Buffers compiler (`protoc`)
- PostgreSQL 15+
- OpenAI API key (for chat service)

### 1. Clone and Setup

```bash
git clone git@github.com:devrob-go/go-chat-ai.git
cd go-chat-ai

# Install dependencies for all modules
make deps
```

### 2. Generate Protocol Buffers

```bash
# Generate auth service protobuf code
cd auth-service
make proto

# Generate chat service protobuf code
cd ../chat-service
make proto
```

### 3. Configure Environment

```bash
# Auth service
cd auth-service
cp env.example .env
# Edit .env with your database and JWT settings

# Chat service
cd ../chat-service
cp env.example .env
# Edit .env with your OpenAI API key and database settings
```

### 4. Start Local Development Stack

```bash
# Start all services with Docker Compose
make docker-run

# Or start only specific services
cd deployment/local
docker-compose up -d postgres  # Database only
docker-compose up -d           # All services
```

## 🛠️ Development

### Available Commands

```bash
# Root level commands
make help          # Show all available commands
make deps          # Install/update dependencies for all modules
make fmt           # Format code across all modules
make lint          # Lint code across all modules
make test          # Run tests across all modules
make security      # Check for security vulnerabilities

# Service-specific commands
cd auth-service
make dev           # Run with hot reload
make build         # Build binary
make test          # Run tests
make proto         # Generate protobuf code

cd chat-service
make dev           # Run with hot reload
make build         # Build binary
make test          # Run tests
make proto         # Generate protobuf code
```

### Project Structure

```
go-chat-ai/
├── auth-service/              # Authentication and user management service
│   ├── proto/                # Protocol Buffer definitions
│   ├── server/               # gRPC server implementation
│   ├── services/             # Business logic layer
│   ├── storage/              # Database operations and migrations
│   ├── models/               # Data models
│   ├── utils/                # Utility functions
│   ├── config/               # Configuration management
│   └── client/               # Example gRPC client
├── chat-service/             # AI chat service
│   ├── proto/                # Chat service protobuf definitions
│   ├── internal/             # Service implementation
│   ├── storage/              # Chat storage and migrations
│   └── scripts/              # Setup and testing scripts
├── packages/                  # Shared packages
│   ├── auth/                 # JWT middleware and auth utilities
│   └── logger/               # Structured logging package
├── deployment/               # Deployment configurations
│   ├── local/                # Local development with Docker Compose
│   └── staging/              # Kubernetes deployment with Helm
├── go.work                   # Go workspace configuration
└── Makefile                  # Root level build commands
```

## 🔐 Authentication

All services use JWT-based authentication:

- **Access Tokens**: Short-lived tokens for API access
- **Refresh Tokens**: Long-lived tokens for token renewal
- **Token Revocation**: Secure token invalidation
- **Role-Based Access**: User roles and permissions

### Protected Endpoints

- **Chat Service**: ALL endpoints require valid JWT token
- **Auth Service**: Most endpoints are public (signup/signin), user management requires admin role

## 🗄️ Database

### PostgreSQL Schema

- **Users Table**: User accounts and credentials
- **User Tokens**: JWT token storage and management
- **Conversations**: Chat conversation metadata
- **Messages**: Individual chat messages with AI responses

### Migrations

Automatic database migrations run on service startup:
- User management tables
- Chat conversation and message tables
- Indexes and constraints

## 🚀 Deployment

### Local Development

```bash
cd deployment/local
docker-compose up -d
```

### Kubernetes Deployment

```bash
# Deploy auth service
cd deployment/staging
./deploy.sh

# Deploy chat service
./deploy-chat.sh

# Deploy both services
./deploy.sh && ./deploy-chat.sh
```

### Production Considerations

- **TLS/SSL**: Configure certificates for production
- **Rate Limiting**: Implement API rate limiting
- **Monitoring**: Health checks and metrics
- **Scaling**: Horizontal pod autoscaling
- **Security**: Network policies and RBAC

## 🧪 Testing

### Test Coverage

```bash
# Run all tests
make test

# Run specific service tests
cd auth-service && go test ./...
cd chat-service && go test ./...

# Run with coverage
go test -coverpkg=./... -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Test Structure

- **Unit Tests**: Individual function testing
- **Integration Tests**: Service layer testing
- **Storage Tests**: Database operation testing
- **Mock Testing**: External dependency mocking

## 🔧 Configuration

### Environment Variables

#### Auth Service
- `APP_PORT`: gRPC server port (default: 8080)
- `REST_PORT`: REST gateway port (default: 8081)
- `POSTGRES_*`: Database connection settings
- `JWT_*_SECRET`: JWT signing secrets

#### Chat Service
- `OPENAI_API_KEY`: Required OpenAI API key
- `CHAT_GRPC_PORT`: gRPC server port (default: 8082)
- `CHAT_REST_PORT`: REST gateway port (default: 8083)
- `POSTGRES_*`: Database connection settings

## 📚 API Documentation

### Auth Service APIs

#### gRPC (Port 8080)
- **AuthService**: User authentication and management
- **HealthService**: Service health monitoring

#### REST (Port 8081)
- `POST /v1/auth/signup` - User registration
- `POST /v1/auth/signin` - User login
- `POST /v1/auth/signout` - User logout
- `POST /v1/auth/refresh` - Token refresh
- `GET /v1/users` - List users (admin only)

### Chat Service APIs

#### gRPC (Port 8082)
- **ChatService**: Chat conversations and AI interactions

#### REST (Port 8083)
- `POST /v1/chat/conversations` - Create conversation
- `GET /v1/chat/conversations` - List conversations
- `POST /v1/chat/conversations/{id}/messages` - Send message
- `GET /v1/chat/conversations/{id}/messages` - Get message history

## 🤝 Contributing

1. Follow Go coding standards and best practices
2. Add comprehensive tests for new functionality
3. Maintain or improve test coverage
4. Update documentation and examples
5. Use conventional commit messages
6. Test edge cases and error scenarios

## 📄 License

This project is open-source and available under the MIT license. See LICENSE for more details.

## 🔗 Related Projects

- **Auth Service**: [README](auth-service/README.md)
- **Chat Service**: [README](chat-service/README.md)
- **Deployment**: [README](deployment/README.md)
- **Shared Packages**: [Auth](packages/auth/), [Logger](packages/logger/)

