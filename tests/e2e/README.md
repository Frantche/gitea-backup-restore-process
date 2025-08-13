# End-to-End (E2E) Testing

This directory contains end-to-end tests for the Gitea backup and restore process.

## Overview

The E2E testing infrastructure validates the complete backup and restore workflow across multiple database and storage combinations:

1. **Environment Setup**: Launches Gitea with various database backends and storage options
2. **Data Creation**: Creates test repositories and issues
3. **Backup Process**: Performs a complete backup operation
4. **Data Loss Simulation**: Simulates system failure by clearing data
5. **Restore Process**: Restores from backup
6. **Verification**: Validates that data was successfully restored

## Supported Test Configurations

### Database Backends
- **MySQL 8.0**: Default relational database
- **PostgreSQL 15**: Alternative relational database

### Storage Backends
- **S3 (MinIO)**: S3-compatible object storage
- **FTP**: Traditional file transfer protocol

### Test Matrix

| Configuration | Database | Storage | Test Command |
|---------------|----------|---------|--------------|
| Default       | MySQL    | S3      | `make test-e2e` |
| PostgreSQL    | PostgreSQL | S3    | `make test-e2e-postgres` |
| FTP           | MySQL    | FTP     | `make test-e2e-ftp` |
| Full Alt      | PostgreSQL | FTP   | `make test-e2e-postgres-ftp` |

## Test Infrastructure Components

### Gitea Bootstrap Script (`gitea_bootstrap.sh`)

A reusable script for initializing Gitea instances with test data across all E2E test configurations.

**Usage:**
```bash
./gitea_bootstrap.sh <gitea_url> <user> <password> <repo_name> <issue_title> [issue_body]
```

**Example:**
```bash
./gitea_bootstrap.sh http://localhost:3000 e2euser e2epassword test-repo "Bug Report" "Test issue description"
```

**What it does:**
- Creates a new repository with auto-initialization (README.md)
- Creates a test issue within the repository
- Uses Gitea REST API for consistent data creation
- Provides standardized test data across all E2E test scenarios
- Ensures reproducible test environments

**Dependencies:**
- `curl`: For HTTP API calls
- `jq`: For JSON processing

The bootstrap script is automatically used by the Go-based E2E test (`e2e.go`) to create consistent test data across all database and storage combinations.

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

### MySQL + S3 E2E Test (`e2e.mysql.s3.sh`)

A comprehensive test using Docker Compose to create a full Gitea environment with MySQL + S3.

**Usage:**
```bash
make test-e2e
# or
./tests/e2e/e2e.mysql.s3.sh
```

### PostgreSQL + S3 E2E Test (`e2e.postgres.s3.sh`)

Tests PostgreSQL database backend with S3 storage.

**Usage:**
```bash
make test-e2e-postgres
# or
./tests/e2e/e2e.postgres.s3.sh
```

### MySQL + FTP E2E Test (`e2e.mysql.ftp.sh`)

Tests FTP storage backend with MySQL database.

**Usage:**
```bash
make test-e2e-ftp
# or
./tests/e2e/e2e.mysql.ftp.sh
```

### PostgreSQL + FTP E2E Test (`e2e.postgres.ftp.sh`)

Tests the combination of PostgreSQL database with FTP storage.

**Usage:**
```bash
make test-e2e-postgres-ftp
# or
./tests/e2e/e2e.postgres.ftp.sh
```

### All E2E Tests

Run all test combinations:

```bash
make test-e2e-all
```

### Full E2E Test (`e2e.go`)

A complete Go-based test that performs the full backup/restore cycle with data validation. This test is configurable and used by all the Docker-based test scripts.

**Configuration via environment variables:**
- `GITEA_URL`: Gitea service URL (default: http://localhost:3000)
- `CONTAINER_NAME`: Backup container name  
- `DATA_VOLUME_NAME`: Gitea data volume name
- `GITEA_CONTAINER_NAME`: Gitea container name

**What it tests:**
- Complete workflow from data creation to restoration
- Gitea API integration
- Backup and restore operations
- Data integrity verification

## Infrastructure

### Docker Compose Files

Multiple compose files for different test scenarios:

- `docker-compose.e2e.mysql.s3.yml`: MySQL + S3 (default)
- `docker-compose.e2e.postgres.s3.yml`: PostgreSQL + S3
- `docker-compose.e2e.mysql.ftp.yml`: MySQL + FTP
- `docker-compose.e2e.postgres.ftp.yml`: PostgreSQL + FTP

Each defines a complete testing environment with:

- **Gitea**: Latest version with appropriate database backend
- **Database**: MySQL 8.0 or PostgreSQL 15
- **Storage**: MinIO (S3) or vsftpd (FTP)
- **Backup Container**: Built from project Dockerfile

### Configuration

- `gitea-config/app.ini`: Gitea configuration for E2E testing
- Environment variables for backup/restore settings specific to each test scenario

## Running Tests

### Prerequisites

- Docker and Docker Compose
- Go 1.21+
- Make (optional, for convenience)

### Quick Start

```bash
# Run all tests
make test

# Run just local E2E tests
make test-e2e-local

# Run specific configuration tests
make test-e2e              # MySQL + S3
make test-e2e-postgres     # PostgreSQL + S3  
make test-e2e-ftp          # MySQL + FTP
make test-e2e-postgres-ftp # PostgreSQL + FTP

# Run all E2E test combinations
make test-e2e-all

# Clean up
make clean
```

### Manual Execution

```bash
# Build binaries
make build

# Run specific test
./tests/e2e/e2e.postgres.s3.sh
./tests/e2e/e2e.mysql.ftp.sh
./tests/e2e/e2e.postgres.ftp.sh
```

## Test Scenarios

### Scenario 1: Default (MySQL + S3)
- Uses MySQL 8.0 database
- Uses MinIO S3-compatible storage
- Tests standard deployment pattern

### Scenario 2: PostgreSQL + S3
- Uses PostgreSQL 15 database
- Uses MinIO S3-compatible storage
- Tests PostgreSQL backup/restore functionality

### Scenario 3: MySQL + FTP
- Uses MySQL 8.0 database
- Uses vsftpd FTP server
- Tests FTP upload/download functionality

### Scenario 4: PostgreSQL + FTP
- Uses PostgreSQL 15 database
- Uses vsftpd FTP server
- Tests full alternative stack

### All Scenarios Include:
- Database-specific backup methods (mysqldump/pg_dump)
- Storage-specific upload/download
- Data integrity verification
- Proper cleanup and error handling

## CI/CD Integration

### GitHub Actions Compatibility

The E2E tests are integrated with GitHub Actions and use Docker Compose v2 (`docker compose` command) which is the default in modern GitHub Actions runners. All test scripts have been updated to use the modern syntax.

### Sequential vs Parallel Execution

**Recommended: Sequential Execution**
By default, E2E tests run sequentially to avoid port conflicts and resource contention:

```bash
make test-e2e-all  # Runs all combinations sequentially
```

**Optional: Parallel Execution** 
For advanced users, a parallel testing workflow is available via manual trigger:

- Navigate to GitHub Actions → "Parallel E2E Tests (Optional)"
- Set `run_parallel` to `true` to run tests in parallel
- Each test configuration uses isolated Docker projects to avoid conflicts

**Note**: Parallel execution requires careful port management and may be less reliable than sequential execution.

## Troubleshooting

### Common Issues

**Docker Compose Command Not Found**
- Error: `docker-compose: command not found`
- Solution: The tests use `docker compose` (Docker Compose v2). Ensure you have Docker Desktop with Compose v2 installed.

**Port Conflicts in Parallel Execution**
- Error: `Port already in use`
- Solution: Use sequential execution (`make test-e2e-all`) or implement unique port allocation for each test.

**Service Health Check Failures**
- Error: Services not becoming healthy
- Solution: Increase timeout values or check Docker logs:

```bash
docker compose -f docker-compose.e2e.postgres.yml logs
```

### Debugging

```bash
# Check Docker logs for specific configuration
docker compose -f docker-compose.e2e.postgres.s3.yml logs

# Check specific service
docker logs gitea-e2e-postgres

# Run tests with debug output
BACKUP_LOG_LEVEL=debug ./tests/e2e/e2e.postgres.s3.sh
```

### Service-Specific Debugging

```bash
# PostgreSQL connection test
docker exec gitea-db-e2e-postgres pg_isready -U gitea -d gitea

# FTP connection test  
docker exec gitea-backup-e2e-ftp nc -w 5 ftp-server 21

# MinIO S3 health check
docker exec gitea-backup-e2e curl -f http://minio:9000/minio/health/live
```

## Integration with CI

The E2E tests are integrated into the GitHub Actions workflow:

- Local E2E tests run on every PR and push
- Full Docker E2E tests run for all database and storage combinations
- Tests must pass before merging
- Matrix testing ensures all database and storage combinations work
- Sequential execution prevents resource conflicts and ensures reliability

## Contributing

When adding new E2E tests:

1. Follow the existing naming convention (`e2e.{db-type}.{target-type}.sh`)
2. Add appropriate cleanup in test scripts (use trap for cleanup functions)
3. Update this documentation
4. Ensure tests are deterministic and can run in parallel
5. Add new docker-compose files for new service combinations following `docker-compose.e2e.{db-type}.{target-type}.yml` pattern
6. Update Makefile with new test targets