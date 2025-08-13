.PHONY: help build test test-unit test-e2e test-e2e-local clean

help: ## Display this help message
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the backup and restore binaries
	@echo "🔨 Building binaries..."
	@go build -o bin/gitea-backup ./cmd/gitea-backup
	@go build -o bin/gitea-restore ./cmd/gitea-restore
	@echo "✅ Build completed"

test: test-unit test-integration test-e2e-local ## Run all tests

test-unit: ## Run unit tests
	@echo "🧪 Running unit tests..."
	@go test -v ./... -short
	@echo "✅ Unit tests completed"

test-integration: ## Run integration tests
	@echo "🧪 Running integration tests..."
	@go test -v ./tests/integration/...
	@echo "✅ Integration tests completed"

test-e2e-local: build ## Run local E2E tests
	@echo "🧪 Running local E2E tests..."
	@./tests/e2e/local_e2e_test.sh
	@echo "✅ Local E2E tests completed"

test-e2e: build ## Run full E2E tests with Docker
	@echo "🧪 Running full E2E tests..."
	@./tests/e2e/basic_e2e_test.sh
	@echo "✅ Full E2E tests completed"

clean: ## Clean build artifacts and test data
	@echo "🧹 Cleaning up..."
	@rm -f bin/gitea-backup bin/gitea-restore
	@rm -rf /tmp/gitea-e2e-test
	@docker-compose -f docker-compose.e2e.yml down -v --remove-orphans 2>/dev/null || true
	@echo "✅ Cleanup completed"