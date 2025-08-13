#!/bin/bash

# E2E Test Script for PostgreSQL + S3 Configuration
# This script tests the backup and restore process using PostgreSQL database and S3 storage

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
COMPOSE_FILE="${PROJECT_ROOT}/docker-compose.e2e.postgres.yml"

echo "🧪 Starting E2E test with PostgreSQL + S3..."

# Cleanup function
cleanup() {
    echo "🧹 Cleaning up..."
    cd "${PROJECT_ROOT}"
    docker compose -f "${COMPOSE_FILE}" down -v --remove-orphans 2>/dev/null || true
    docker system prune -f --volumes 2>/dev/null || true
}

# Set trap for cleanup
trap cleanup EXIT

# Change to project root
cd "${PROJECT_ROOT}"

echo "📦 Building services..."
docker compose -f "${COMPOSE_FILE}" build --no-cache

echo "🚀 Starting services..."
docker compose -f "${COMPOSE_FILE}" up -d

echo "⏳ Waiting for services to be healthy..."
# Wait for all services to be healthy
timeout 300 bash -c '
    while true; do
        if docker compose -f "'"${COMPOSE_FILE}"'" ps | grep -q "unhealthy\|starting"; then
            echo "Services still starting..."
            sleep 10
            continue
        fi
        if ! docker compose -f "'"${COMPOSE_FILE}"'" ps | grep -q "(healthy)"; then
            echo "Some services not healthy yet..."
            sleep 10
            continue
        fi
        break
    done
'

echo "✅ All services are healthy"

# Check that PostgreSQL is working
echo "🔍 Verifying PostgreSQL connection..."
docker exec gitea-db-e2e-postgres pg_isready -U gitea -d gitea

# Check that MinIO is working
echo "🔍 Verifying MinIO S3 connectivity..."
docker exec gitea-backup-e2e-postgres curl -f http://minio:9000/minio/health/live

echo "🧪 Running E2E test..."
# Run the E2E test in the backup container with PostgreSQL environment
docker exec gitea-backup-e2e-postgres sh -c '
    export GITEA_URL="http://gitea-e2e-postgres:3000"
    export CONTAINER_NAME="gitea-backup-e2e-postgres"
    export DATA_VOLUME_NAME="gitea-backup-restore-process_gitea-data-postgres"
    export GITEA_CONTAINER_NAME="gitea-e2e-postgres"
    cd /tests/e2e && go run e2e_test.go
'

echo "✅ PostgreSQL + S3 E2E test completed successfully!"