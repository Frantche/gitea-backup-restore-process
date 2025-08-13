#!/bin/bash

set -e

echo "🚀 Starting Gitea Backup/Restore E2E Tests"

# Define cleanup function
cleanup() {
    echo "🧹 Cleaning up..."
    docker-compose -f docker-compose.e2e.yml down -v --remove-orphans || true
    docker volume prune -f || true
}

# Set trap to cleanup on exit
trap cleanup EXIT

# Ensure we're in the right directory
cd "$(dirname "$0")/../.."

# Build the latest docker image
echo "🔨 Building Docker image..."
docker build -t gitea-backup-e2e .

# Start services
echo "🏃 Starting services..."
docker-compose -f docker-compose.e2e.yml up -d

# Wait for services to be ready
echo "⏳ Waiting for services to start..."
sleep 30

# Create MinIO bucket
echo "📦 Setting up MinIO bucket..."
docker exec minio-e2e mc alias set local http://localhost:9000 minioadmin minioadmin123 || true
docker exec minio-e2e mc mb local/gitea-backup || true

# Build and run E2E test
echo "🧪 Building E2E test..."
cd tests/e2e
go mod init e2e-test 2>/dev/null || true
go mod tidy
go build -o e2e_test e2e_test.go

echo "🚀 Running E2E test..."
./e2e_test

echo "✅ E2E tests completed successfully!"