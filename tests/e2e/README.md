# End-to-End (E2E) Testing

This directory contains end-to-end tests for the Gitea backup and restore process.

## Overview

The E2E testing infrastructure validates the complete backup and restore workflow:

1. **Environment Setup**: Launches Gitea, MySQL database, and S3-compatible storage (MinIO)
2. **Data Creation**: Creates test repositories and issues
3. **Backup Process**: Performs a complete backup operation
4. **Data Loss Simulation**: Simulates system failure by clearing data
5. **Restore Process**: Restores from backup
6. **Verification**: Validates that data was successfully restored

## Test Scripts

### Local E2E Test (`local_e2e_test.sh`)

A lightweight test that validates the backup/restore binaries work correctly without requiring Docker infrastructure.

**Usage:**
```bash
make test-e2e-local
# or
./tests/e2e/local_e2e_test.sh
```

**What it tests:**
- Binary compilation and execution
- Configuration file parsing
- Basic error handling
- File system operations

### Basic E2E Test (`basic_e2e_test.sh`)

A comprehensive test using Docker Compose to create a full Gitea environment.

**Usage:**
```bash
make test-e2e
# or
./tests/e2e/basic_e2e_test.sh
```

**What it tests:**
- Docker environment setup
- Service connectivity (Gitea, MySQL, MinIO)
- Backup command execution in containerized environment
- Integration between services

### Full E2E Test (`e2e_test.go`)

A complete Go-based test that performs the full backup/restore cycle with data validation.

**Usage:**
```bash
cd tests/e2e
go run e2e_test.go
```

**What it tests:**
- Complete workflow from data creation to restoration
- Gitea API integration
- Backup and restore operations
- Data integrity verification

## Infrastructure

### Docker Compose (`docker-compose.e2e.yml`)

Defines a complete testing environment with:

- **Gitea**: Latest version with MySQL backend
- **MySQL**: Database for Gitea
- **MinIO**: S3-compatible storage for backups
- **Backup Container**: Built from project Dockerfile

### Configuration

- `gitea-config/app.ini`: Gitea configuration for E2E testing
- Environment variables for backup/restore settings

## Running Tests

### Prerequisites

- Docker and Docker Compose
- Go 1.21+
- Make (optional, for convenience)

### Quick Start

```bash
# Run all tests
make test

# Run just E2E tests
make test-e2e-local

# Run full Docker-based E2E tests (requires Docker)
make test-e2e

# Clean up
make clean
```

### Manual Execution

```bash
# Build binaries
make build

# Run local E2E test
./tests/e2e/local_e2e_test.sh

# Run Docker-based E2E test
./tests/e2e/basic_e2e_test.sh
```

## Test Scenarios

### Scenario 1: SQLite Backup/Restore
- Creates SQLite-based Gitea instance
- Performs file-based backup
- Simulates data loss
- Restores from backup

### Scenario 2: MySQL Backup/Restore
- Uses MySQL database
- Performs database dump + file backup
- Validates MySQL restore functionality

### Scenario 3: S3 Storage
- Uses MinIO as S3-compatible storage
- Tests remote backup storage
- Validates download and restore from S3

## Troubleshooting

### Common Issues

1. **Docker build failures**: Ensure Docker has internet access for package downloads
2. **Port conflicts**: Ensure ports 3000, 3306, 9000, 9001 are available
3. **Permission issues**: Ensure test scripts are executable (`chmod +x`)

### Debugging

```bash
# Check Docker logs
docker-compose -f docker-compose.e2e.yml logs

# Check specific service
docker logs gitea-e2e

# Run tests with debug output
BACKUP_LOG_LEVEL=debug ./tests/e2e/basic_e2e_test.sh
```

## Integration with CI

The E2E tests are integrated into the GitHub Actions workflow:

- Local E2E tests run on every PR and push
- Full Docker E2E tests can be enabled for specific branches
- Tests must pass before merging

## Contributing

When adding new E2E tests:

1. Follow the existing naming convention
2. Add appropriate cleanup in test scripts
3. Update this documentation
4. Ensure tests are deterministic and can run in parallel