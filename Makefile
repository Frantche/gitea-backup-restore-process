.PHONY: help build test test-unit test-integration test-e2e-mysql-s3 test-e2e-local test-e2e-postgres-s3 test-e2e-mysql-ftp test-e2e-postgres-ftp clean

help: ## Display this help message
	@echo "Available targets:"
	@awk 'BEGIN {FS = ": ## "} /^[a-zA-Z0-9_-]+: ## / {printf "  %-25s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the backup and restore binaries
	@echo "🔨 Building binaries..."
	@go build -o bin/gitea-backup ./cmd/gitea-backup
	@go build -o bin/gitea-restore ./cmd/gitea-restore
	@echo "✅ Build completed"

test: ## Run all tests
test: test-unit test-integration test-e2e-local

test-unit: ## Run unit tests
	@echo "🧪 Running unit tests..."
	@go test -v ./... -short
	@echo "✅ Unit tests completed"

test-integration: ## Run integration tests
	@echo "🧪 Running integration tests..."
	@go test -v ./tests/integration/...
	@echo "✅ Integration tests completed"

test-e2e-local: ## Run local E2E tests
	$(MAKE) build
	@echo "🧪 Running local E2E tests..."
	@./tests/e2e/local_e2e_test.sh
	@echo "✅ Local E2E tests completed"

test-e2e-mysql-s3: ## Run E2E tests with MySQL + S3
	$(MAKE) build
	@echo "🧪 Running E2E tests (MySQL + S3)..."
	@./tests/e2e/e2e.mysql.s3.sh
	@echo "✅ MySQL + S3 E2E tests completed"

test-e2e-postgres-s3: ## Run E2E tests with PostgreSQL + S3
	$(MAKE) build
	@echo "🧪 Running E2E tests (PostgreSQL + S3)..."
	@./tests/e2e/e2e.postgres.s3.sh
	@echo "✅ PostgreSQL + S3 E2E tests completed"

test-e2e-mysql-ftp: ## Run E2E tests with MySQL + FTP
	$(MAKE) build
	@echo "🧪 Running E2E tests (MySQL + FTP)..."
	@./tests/e2e/e2e.mysql.ftp.sh
	@echo "✅ MySQL + FTP E2E tests completed"

test-e2e-postgres-ftp: ## Run E2E tests with PostgreSQL + FTP
	$(MAKE) build
	@echo "🧪 Running E2E tests (PostgreSQL + FTP)..."
	@./tests/e2e/e2e.postgres.ftp.sh
	@echo "✅ PostgreSQL + FTP E2E tests completed"

test-e2e-all: ## Run all E2E test combinations
test-e2e-all: test-e2e-mysql-s3 test-e2e-postgres-s3 test-e2e-mysql-ftp test-e2e-postgres-ftp

clean: ## Clean build artifacts and test data
	@echo "🧹 Cleaning up..."
	@rm -f bin/gitea-backup bin/gitea-restore
	@rm -rf /tmp/gitea-e2e-test
	@docker compose -f docker-compose.e2e.mysql.s3.yml down -v --remove-orphans 2>/dev/null || true
	@docker compose -f docker-compose.e2e.postgres.s3.yml down -v --remove-orphans 2>/dev/null || true
	@docker compose -f docker-compose.e2e.mysql.ftp.yml down -v --remove-orphans 2>/dev/null || true
	@docker compose -f docker-compose.e2e.postgres.ftp.yml down -v --remove-orphans 2>/dev/null || true
	@echo "✅ Cleanup completed"