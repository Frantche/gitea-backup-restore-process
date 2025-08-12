# Multi-stage build for Go application
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o bin/gitea-backup ./cmd/gitea-backup && \
    go build -o bin/gitea-restore ./cmd/gitea-restore

# Final runtime image
FROM registry.access.redhat.com/ubi9

# Install necessary database clients
RUN dnf install wget -y \
  && wget https://dev.mysql.com/get/mysql80-community-release-el9-5.noarch.rpm \
  && dnf install ./mysql80-community-release-el9-5.noarch.rpm -y \
  && dnf install mysql -y

# Install PostgreSQL client
RUN dnf install -y https://download.postgresql.org/pub/repos/yum/reporpms/EL-9-x86_64/pgdg-redhat-repo-latest.noarch.rpm \
  && dnf install postgresql -y

# Copy Go binaries from builder stage
COPY --from=builder /app/bin/gitea-backup /usr/local/bin/
COPY --from=builder /app/bin/gitea-restore /usr/local/bin/

# Create default directories
RUN mkdir -p /data /tmp/backup /tmp/restore

CMD [ "sleep", "infinity" ]