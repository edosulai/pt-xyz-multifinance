FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install required packages
RUN apk add --no-cache make gcc musl-dev protobuf protobuf-dev git

# Install protoc plugins and ensure they're in PATH
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28 && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2 && \
    go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.15.0 && \
    go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.15.0 && \
    mv /go/bin/protoc-gen-go* /usr/local/bin/ && \
    mv /go/bin/protoc-gen-grpc* /usr/local/bin/ && \
    mv /go/bin/protoc-gen-openapiv2 /usr/local/bin/

# Set up third party protos
RUN mkdir -p third_party/google/api && \
    cd third_party/google/api && \
    wget https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto && \
    wget https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application
COPY . .

# Create necessary directories and set permissions
RUN mkdir -p proto/gen/openapiv2 && \
    chmod -R 755 proto/gen

# Generate proto files
RUN make proto

# Build the application
RUN make build

FROM alpine:latest

WORKDIR /app

# Install necessary runtime packages
RUN apk add --no-cache curl postgresql-client

# Create logs directory with proper permissions
RUN mkdir -p /app/logs && chmod 777 /app/logs

# Install golang-migrate
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.16.2/migrate.linux-amd64.tar.gz | tar xvz && \
    mv migrate /usr/local/bin/migrate && \
    chmod +x /usr/local/bin/migrate

# Copy binary and configuration files from builder
COPY --from=builder /app/bin/server /app/main
COPY --from=builder /app/migrations /app/migrations
COPY --from=builder /app/configs /app/configs
COPY --from=builder /app/proto/gen /app/proto/gen

# Copy entrypoint script
COPY docker-entrypoint.sh /app/
RUN chmod +x /app/docker-entrypoint.sh

EXPOSE 8080 9090

ENTRYPOINT ["/app/docker-entrypoint.sh"]
