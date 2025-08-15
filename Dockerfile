# Multi-stage build for Go application
FROM golang:1.24-alpine AS builder

# Install ca-certificates
RUN apk add --no-cache ca-certificates git

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o bin/gitea-backup ./cmd/gitea-backup && \
    go build -o bin/gitea-restore ./cmd/gitea-restore

# Final runtime image - using Ubuntu for better package availability
FROM ubuntu:22.04

# Install necessary database clients and tools
RUN apt-get update && apt-get install -y \
    mysql-client \
    postgresql-client \
    wget \
    curl \
    jq \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Copy Go binaries from builder stage
COPY --from=builder /app/bin/gitea-backup /usr/local/bin/
COPY --from=builder /app/bin/gitea-restore /usr/local/bin/

CMD [ "sleep", "infinity" ]