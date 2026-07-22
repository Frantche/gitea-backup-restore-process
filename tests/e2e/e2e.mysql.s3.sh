#!/bin/bash

# E2E Test Script for MySQL + S3 Configuration
# This script tests the backup and restore process using MySQL database and S3 storage

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
COMPOSE_FILE="${PROJECT_ROOT}/docker-compose/e2e.mysql.s3.yml"

echo "🧪 Starting MySQL + S3 E2E test for Gitea backup/restore..."

# Cleanup function
cleanup() {
    echo "🧹 Cleaning up..."
    cd "${PROJECT_ROOT}"
    docker compose -f "${COMPOSE_FILE}" down -v --remove-orphans 2>/dev/null || true
}

# Set trap for cleanup
trap cleanup EXIT

# Change to project root
cd "${PROJECT_ROOT}"

echo "Delete volume before starting E2E test"
docker compose -f "${COMPOSE_FILE}" down
docker volume ls -q | grep '^docker-compose' | xargs -r docker volume rm -f

echo "🚀 Starting services..."
docker compose -f "${COMPOSE_FILE}" up -d --build

# Check if services are running
echo "📋 Checking service status..."
docker compose -f "${COMPOSE_FILE}" ps

# Setup MinIO bucket
echo "📦 Setting up MinIO bucket..."
docker exec minio-e2e sh -c "mc alias set local http://localhost:9000 minioadmin minioadmin123 && mc mb local/gitea-backup" || echo "Bucket might already exist"

# Test basic connectivity
echo "🌐 Testing service connectivity..."
max_retries=30
retry_interval=5
attempt=1
reachable=false

while [ $attempt -le $max_retries ]; do
    if curl -sf http://localhost:3000/ > /dev/null 2>&1; then
        echo "✅ Gitea is accessible (attempt $attempt/$max_retries)"
        reachable=true
        break
    else
        echo "⏳ Attempt $attempt/$max_retries failed, retrying in ${retry_interval}s..."
        sleep $retry_interval
    fi
    attempt=$((attempt+1))
done

if [ "$reachable" = false ]; then
    echo "❌ Gitea is not accessible after $max_retries attempts"
    docker logs gitea-mysql
    exit 1   # exit with failure
fi

if curl -f http://localhost:9000/minio/health/live > /dev/null 2>&1; then
    echo "✅ MinIO is accessible"
else
    echo "❌ MinIO is not accessible"
    docker logs minio-e2e
    exit 1
fi

# Check that MySQL is working
echo "🔍 Verifying MySQL connection..."
docker exec gitea-db-mysql mysqladmin ping -h 127.0.0.1 -P 3306 -u gitea -pgitea123

# Initialize Gitea with a simple admin user
echo "👤 Initializing Gitea admin user..."
docker exec --user git gitea-mysql gitea admin user create --admin --username e2euser --password e2epassword --email e2e@example.com || echo "Admin user might already exist"

# Build and run the E2E test outside of the container
echo "🔧 Building E2E test binary..."
cd tests/e2e
go build -o e2e-test ./e2e.go
cd ../..

# Set environment variables for the E2E test
export GITEA_URL="http://localhost:3000"
export CONTAINER_NAME="gitea-backup-e2e"
export DATA_VOLUME_NAME="docker-compose_gitea-data"
export GITEA_CONTAINER_NAME="gitea-mysql"
export DB_CONTAINER_NAME="gitea-db-mysql"
export DB_VOLUME_NAME="docker-compose_db-mysql-data"

# Run the comprehensive E2E test
echo "🧪 Running comprehensive E2E test..."
if ./tests/e2e/e2e-test; then
    echo "✅ Comprehensive E2E test completed successfully!"
else
    echo "❌ E2E test failed"
    #docker logs gitea-backup-e2e
    #docker logs gitea-mysql
    exit 1
fi

echo "🎉 MySQL + S3 E2E test completed successfully!"
