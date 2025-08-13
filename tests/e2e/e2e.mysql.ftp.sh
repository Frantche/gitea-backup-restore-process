#!/bin/bash

# E2E Test Script for MySQL + FTP Configuration
# This script tests the backup and restore process using MySQL database and FTP storage

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
COMPOSE_FILE="${PROJECT_ROOT}/docker-compose.e2e.mysql.ftp.yml"

echo "🧪 Starting E2E test with MySQL + FTP..."

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

# Check that MySQL is working
echo "🔍 Verifying MySQL connection..."
docker exec gitea-db-e2e-ftp mysqladmin ping -h localhost -u gitea -pgitea123

# Check that FTP is working
echo "🔍 Verifying FTP connectivity..."
docker exec gitea-backup-e2e-ftp sh -c 'echo "test connection" | nc -w 5 ftp-server 21 | head -1 | grep -q "220"'

echo "🧪 Running E2E test..."
# Run the E2E test in the backup container with FTP environment
docker exec gitea-backup-e2e-ftp sh -c '
    export GITEA_URL="http://gitea-e2e-ftp:3000"
    export CONTAINER_NAME="gitea-backup-e2e-ftp"
    export DATA_VOLUME_NAME="gitea-backup-restore-process_gitea-data-ftp"
    export GITEA_CONTAINER_NAME="gitea-e2e-ftp"
    cd /tests/e2e && go run e2e_test.go
'

echo "✅ MySQL + FTP E2E test completed successfully!"