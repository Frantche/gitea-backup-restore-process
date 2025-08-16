#!/bin/bash

# E2E Test Script for PostgreSQL + FTP Configuration
# This script tests the backup and restore process using PostgreSQL database and FTP storage

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
COMPOSE_FILE="${PROJECT_ROOT}/docker-compose/e2e.postgres.ftp.yml"

echo "🧪 Starting PostgreSQL + FTP E2E test for Gitea backup/restore..."

# Cleanup function
cleanup() {
    echo "🧹 Cleaning up..."
    cd "${PROJECT_ROOT}"
    docker compose -f "${COMPOSE_FILE}" down -v --remove-orphans 2>/dev/null || true
}

# Set trap for cleanup
#trap cleanup EXIT

# Change to project root
cd "${PROJECT_ROOT}"

echo "🔨 Building Docker image..."
docker build -t gitea-backup-e2e .

echo "Delete volume before starting E2E test"
docker compose -f "${COMPOSE_FILE}" down
docker volume ls -q | grep '^docker-compose' | xargs -r docker volume rm -f

echo "🚀 Starting services..."
docker compose -f "${COMPOSE_FILE}" up -d

# Check if services are running
echo "📋 Checking service status..."
docker compose -f "${COMPOSE_FILE}" ps

# Test basic connectivity
echo "🌐 Testing service connectivity..."
if curl -f http://localhost:3000/ > /dev/null 2>&1; then
    echo "✅ Gitea is accessible"
else
    echo "❌ Gitea is not accessible"
    docker logs gitea-postgres
    exit 1
fi

# Check that PostgreSQL is working
echo "🔍 Verifying PostgreSQL connection..."
docker exec gitea-db-postgres pg_isready -U gitea -d gitea

# Check that FTP is working
echo "🔍 Verifying FTP connectivity..."
curl ftp://127.0.0.1:21 --user ftpuser:ftppass123 || echo "FTP check skipped"

# Initialize Gitea with a simple admin user
echo "👤 Initializing Gitea admin user..."
docker exec --user git gitea-postgres gitea admin user create --admin --username e2euser --password e2epassword --email e2e@example.com || echo "Admin user might already exist"

# Build and run the E2E test outside of the container
echo "🔧 Building E2E test binary..."
cd tests/e2e
go build -o e2e-test ./e2e.go
cd ../..

# Set environment variables for the E2E test
export GITEA_URL="http://localhost:3000"
export CONTAINER_NAME="gitea-backup-e2e"
export DATA_VOLUME_NAME="docker-compose_gitea-data"
export GITEA_CONTAINER_NAME="gitea-postgres"
export DB_CONTAINER_NAME="gitea-db-postgres"
export DB_VOLUME_NAME="docker-compose_db-postgres-data"

# Run the comprehensive E2E test
echo "🧪 Running comprehensive E2E test..."
if ./tests/e2e/e2e-test; then
    echo "✅ Comprehensive E2E test completed successfully!"
else
    echo "❌ E2E test failed"
    #docker logs gitea-backup-e2e
    #docker logs gitea-postgres
    exit 1
fi

echo "🎉 PostgreSQL + FTP E2E test completed successfully!"