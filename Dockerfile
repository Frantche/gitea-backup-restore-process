# ---------- Build stage ----------
FROM golang:1.25-alpine AS builder

WORKDIR /app
# If you fetch private modules, uncomment the next line:
# RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o bin/gitea-backup ./cmd/gitea-backup && \
    go build -o bin/gitea-restore ./cmd/gitea-restore

# ---------- Runtime stage ----------
FROM ubuntu:24.04

ARG DEBIAN_FRONTEND=noninteractive
# Set the Postgres MAJOR you want. 15 matches your server (15.14).
# Change to 16 if you want the newest major.
ARG PG_MAJOR=17

# Base tools + add the official PostgreSQL APT repo (PGDG) for up-to-date clients
RUN apt-get update && apt-get install -y --no-install-recommends \
      ca-certificates curl gnupg lsb-release \
    && install -d -m 0755 /etc/apt/keyrings \
    && curl -fsSL https://www.postgresql.org/media/keys/ACCC4CF8.asc \
         | gpg --dearmor -o /etc/apt/keyrings/postgresql.gpg \
    && echo "deb [signed-by=/etc/apt/keyrings/postgresql.gpg] https://apt.postgresql.org/pub/repos/apt $(. /etc/os-release; echo ${VERSION_CODENAME})-pgdg main" \
         > /etc/apt/sources.list.d/pgdg.list \
    && apt-get update \
    # Install database clients (pg_dump/psql from PGDG in the version you chose)
    && apt-get install -y --no-install-recommends \
         "postgresql-client-${PG_MAJOR}" \
         mysql-client \
         wget \
         jq \
         curl \
    && apt-get purge -y gnupg lsb-release \
    && rm -rf /var/lib/apt/lists/*

# Copy Go binaries from builder stage
COPY --from=builder /app/bin/gitea-backup /usr/local/bin/
COPY --from=builder /app/bin/gitea-restore /usr/local/bin/

# Optional: show the installed pg_dump version at container start
# (handy for debugging images)
# CMD ["bash", "-lc", "pg_dump --version && sleep infinity"]

CMD [ "sleep", "infinity" ]
