#!/bin/bash

set -e

echo "🧪 Running basic E2E test for Gitea backup/restore"

# Ensure we're in the right directory
cd "$(dirname "$0")/../.."

# Define cleanup function
cleanup() {
    echo "🧹 Cleaning up..."
    docker-compose -f docker-compose.e2e.yml down -v --remove-orphans 2>/dev/null || true
}

# Set trap to cleanup on exit
trap cleanup EXIT

# Build the latest docker image
echo "🔨 Building Docker image..."
docker build -t gitea-backup-e2e .

# Start services
echo "🏃 Starting services..."
docker-compose -f docker-compose.e2e.yml up -d

# Wait for services to be ready
echo "⏳ Waiting for services to initialize..."
sleep 60

# Check if services are running
echo "📋 Checking service status..."
docker-compose -f docker-compose.e2e.yml ps

# Setup MinIO bucket
echo "📦 Setting up MinIO bucket..."
docker exec minio-e2e sh -c "mc alias set local http://localhost:9000 minioadmin minioadmin123 && mc mb local/gitea-backup" || echo "Bucket might already exist"

# Test basic connectivity
echo "🌐 Testing service connectivity..."
if curl -f http://localhost:3000/ > /dev/null 2>&1; then
    echo "✅ Gitea is accessible"
else
    echo "❌ Gitea is not accessible"
    docker logs gitea-e2e
    exit 1
fi

if curl -f http://localhost:9000/minio/health/live > /dev/null 2>&1; then
    echo "✅ MinIO is accessible"
else
    echo "❌ MinIO is not accessible"
    docker logs minio-e2e
    exit 1
fi

# Test backup functionality (basic test)
echo "💾 Testing backup functionality..."
docker exec gitea-backup-e2e sh -c "ls -la /data && echo 'Gitea data directory:' && ls -la /data/gitea || echo 'No gitea directory yet'"

# Initialize Gitea with a simple admin user
echo "👤 Initializing Gitea admin user..."
docker exec gitea-e2e gitea admin user create --admin --username admin --password admin123 --email admin@example.com || echo "Admin user might already exist"

# Test a simple backup operation
echo "💾 Performing test backup..."
if docker exec gitea-backup-e2e gitea-backup; then
    echo "✅ Backup command executed successfully"
else
    echo "❌ Backup command failed"
    docker logs gitea-backup-e2e
    exit 1
fi

echo "✅ Basic E2E test completed successfully!"
echo "🎉 All services are working and backup command can be executed"