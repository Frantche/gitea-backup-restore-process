#!/bin/bash

set -e

echo "🧪 Running MySQL + S3 E2E test for Gitea backup/restore"

# Ensure we're in the right directory
cd "$(dirname "$0")/../.."

# Define cleanup function
cleanup() {
    echo "🧹 Cleaning up..."
    docker compose -f docker-compose.e2e.mysql.s3.yml down -v --remove-orphans 2>/dev/null || true
}

# Set trap to cleanup on exit
trap cleanup EXIT

# Build the latest docker image
echo "🔨 Building Docker image..."
docker build -t gitea-backup-e2e .

# Start services
echo "🏃 Starting services..."
docker compose -f docker-compose.e2e.mysql.s3.yml up -d

# Wait for services to be ready
echo "⏳ Waiting for services to initialize..."
sleep 60

# Check if services are running
echo "📋 Checking service status..."
docker compose -f docker-compose.e2e.mysql.s3.yml ps

# Setup MinIO bucket
echo "📦 Setting up MinIO bucket..."
docker exec minio-e2e sh -c "mc alias set local http://localhost:9000 minioadmin minioadmin123 && mc mb local/gitea-backup" || echo "Bucket might already exist"

# Test basic connectivity
echo "🌐 Testing service connectivity..."
if curl -f http://localhost:3000/ > /dev/null 2>&1; then
    echo "✅ Gitea is accessible"
else
    echo "❌ Gitea is not accessible"
    docker logs gitea-mysql
    exit 1
fi

if curl -f http://localhost:9000/minio/health/live > /dev/null 2>&1; then
    echo "✅ MinIO is accessible"
else
    echo "❌ MinIO is not accessible"
    docker logs minio-e2e
    exit 1
fi

# Initialize Gitea with a simple admin user
echo "👤 Initializing Gitea admin user..."
docker exec gitea-mysql gitea admin user create --admin --username e2euser --password e2epassword --email e2e@example.com || echo "Admin user might already exist"

# Build and run the E2E test outside of the container
echo "🔧 Building E2E test binary..."
cd tests/e2e
go build -o e2e-test ./e2e.go
cd ../..

# Set environment variables for the E2E test
export GITEA_URL="http://localhost:3000"
export CONTAINER_NAME="gitea-backup-e2e"
export DATA_VOLUME_NAME="docker-compose.e2e.mysql.s3_gitea-data"
export GITEA_CONTAINER_NAME="gitea-mysql"

# Run the comprehensive E2E test
echo "🧪 Running comprehensive E2E test..."
if ./tests/e2e/e2e-test; then
    echo "✅ Comprehensive E2E test completed successfully!"
else
    echo "❌ E2E test failed"
    docker logs gitea-backup-e2e
    docker logs gitea-mysql
    exit 1
fi

echo "🎉 MySQL + S3 E2E test completed successfully!"