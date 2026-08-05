# Gitea Backup & Restore Process

A simple and efficient backup and restore solution for Gitea written in Go.

## Features

- **Database Support**: MySQL, PostgreSQL, and SQLite3
- **Storage Backends**: S3-compatible storage and FTP
- **File Backup**: Repositories, avatars, and configuration files
- **Retention Management**: Automatic cleanup of old backups
- **Restore History**: Prevents duplicate restores
- **Docker Support**: Ready-to-use Docker container
- **Simple Configuration**: YAML-based configuration via environment variables

## Quick Start

### Using Docker

```bash
docker run --rm \
  -e BACKUP_ENABLE=true \
  -e BACKUP_METHODE=s3 \
  -e ENDPOINT_URL=https://s3.amazonaws.com \
  -e AWS_ACCESS_KEY_ID=your_access_key \
  -e AWS_SECRET_ACCESS_KEY=your_secret_key \
  -e BUCKET=your_bucket \
  -v /path/to/gitea:/data \
  ghcr.io/frantche/gitea-backup-restore-process:latest \
  gitea-backup
```

### Restore with Docker

Restores are intentionally explicit: provide the exact backup archive name with
`BACKUP_FILENAME`, mount the same Gitea data directory, and keep the database
reachable from the restore container. Stop the Gitea web container first, but
keep its database running.

```bash
docker run --rm \
  --network your_gitea_network \
  -e BACKUP_ENABLE=true \
  -e BACKUP_METHODE=s3 \
  -e ENDPOINT_URL=https://s3.amazonaws.com \
  -e AWS_ACCESS_KEY_ID=your_access_key \
  -e AWS_SECRET_ACCESS_KEY=your_secret_key \
  -e BUCKET=your_bucket \
  -e BACKUP_FILENAME=gitea-backup-YYYY-MM-DD-HH-MM-SS.zip \
  -v /path/to/gitea:/data \
  ghcr.io/frantche/gitea-backup-restore-process:latest \
  gitea-restore
```

The restore command downloads and extracts the archive, replaces the restored
repositories/avatar directories, restores the database with strict SQL error
handling, and normalizes restored file ownership for the Gitea user. By default
that user is `git`; set `GITEA_USER` if your container uses another user or a
numeric owner such as `1000:1000`.

### Using Binaries

1. Download the latest release
2. Set environment variables
3. Run the backup or restore command

```bash
# Backup
./gitea-backup

# Restore
./gitea-restore
```

## Configuration

All configuration is done through environment variables:

### Required Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `BACKUP_ENABLE` | Enable backup process | `true` |
| `BACKUP_METHODE` | Storage backend (s3/ftp) | `s3` |

### S3 Configuration

| Variable | Description | Example |
|----------|-------------|---------|
| `ENDPOINT_URL` | S3 endpoint URL | `https://s3.amazonaws.com` |
| `AWS_ACCESS_KEY_ID` | AWS access key | `AKIAIOSFODNN7EXAMPLE` |
| `AWS_SECRET_ACCESS_KEY` | AWS secret key | `wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY` |
| `BUCKET` | S3 bucket name | `my-gitea-backups` |
| `S3_LOG_DEBUG` | Enable S3 debug logging | `true` or `false` (default: `false`) |
| `S3_MULTIPART_ENABLED` | Enable multipart S3 upload and download with 10 MiB parts | `true` or `false` (default: `false`) |

### FTP Configuration

| Variable | Description | Example |
|----------|-------------|---------|
| `BACKUP_FTP_HOST` | FTP server host | `ftp.example.com:21` |
| `BACKUP_FTP_USER` | FTP username | `backup_user` |
| `BACKUP_FTP_PASSWORD` | FTP password | `secure_password` |
| `BACKUP_FTP_DIR` | FTP directory (optional) | `/backups` |

### Optional Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `BACKUP_PREFIX` | `gitea-backup` | Backup file prefix |
| `BACKUP_MAX_RETENTION` | `5` | Maximum number of backups to keep |
| `APP_INI_PATH` | `/data/gitea/conf/app.ini` | Path to Gitea configuration |
| `BACKUP_TMP_FOLDER` | `/tmp/backup` | Temporary backup folder |
| `RESTORE_TMP_FOLDER` | `/tmp/restore` | Temporary restore folder |
| `GITEA_USER` | `git` | User or `uid:gid` that should own restored repositories and avatars |
| `LOG_LEVEL` | `info` | Set to `debug` for additional operational details. Credentials are never logged. |

## Database Support

### SQLite3
Automatically detects SQLite databases and performs file-based backups.

### MySQL
Uses `mysqldump` and `mysql` commands. Requires MySQL client tools.

### PostgreSQL
Uses `pg_dump` and `psql` commands. The container image ships PostgreSQL client
18 so it can back up PostgreSQL 18 servers as well as older supported majors.

## Development

### Building from Source

```bash
# Clone the repository
git clone https://github.com/Frantche/gitea-backup-restore-process.git
cd gitea-backup-restore-process

# Build binaries
go build -o bin/gitea-backup ./cmd/gitea-backup
go build -o bin/gitea-restore ./cmd/gitea-restore
```

### Running Tests

```bash
go test ./...
```

### End-to-End Testing

The project includes comprehensive E2E tests that validate the complete backup and restore workflow:

```bash
# Run all tests including E2E
make test

# Run only E2E tests
make test-e2e-local

# Run integration tests
make test-integration

# Run full Docker-based E2E tests (requires Docker)
make test-e2e
```

See [tests/e2e/README.md](tests/e2e/README.md) for detailed information about the E2E testing infrastructure.

### Building Docker Image

```bash
docker build -t gitea-backup .
```

## Migration from Python Version

This Go implementation is a drop-in replacement for the Python version with the following improvements:

- **Performance**: Significantly faster execution
- **Dependencies**: Reduced external dependencies
- **SQLite Support**: Added native SQLite3 support
- **Better Error Handling**: More robust error handling and logging
- **Memory Efficiency**: Lower memory footprint

The configuration remains compatible with the Python version.

## License

This project is open source and available under the [MIT License](LICENSE).

## Contributing


Contributions are welcome! Please feel free to submit a Pull Request.
