# Build stage
FROM golang:1.24.6 AS builder

# Install necessary tools including protoc
RUN apt-get update && apt-get install -y \
    git \
    ca-certificates \
    tzdata \
    wget \
    unzip \
    && rm -rf /var/lib/apt/lists/*

# Install protoc
RUN wget https://github.com/protocolbuffers/protobuf/releases/download/v25.1/protoc-25.1-linux-x86_64.zip \
    && unzip protoc-25.1-linux-x86_64.zip -d /usr/local \
    && rm protoc-25.1-linux-x86_64.zip

# Install Go protobuf plugins
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest \
    && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest \
    && go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest

# Set working directory
WORKDIR /app

# Copy the entire project structure
COPY . .

# Download dependencies
RUN cd services/chat-service && go mod download

# Generate protobuf code
RUN cd services/chat-service && protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    --grpc-gateway_out=. --grpc-gateway_opt=paths=source_relative \
    proto/chat.proto

# Build the application
RUN cd services/chat-service && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/services/chat-service/main .

# Copy the storage migrations directory
COPY --from=builder /app/services/chat-service/storage ./storage

# Expose gRPC port
EXPOSE 8082

# Run the application
CMD ["./main"]
