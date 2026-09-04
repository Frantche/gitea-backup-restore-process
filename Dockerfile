# ---------- Build stage ----------
FROM golang:1.27-alpine AS builder

WORKDIR /app
# If you fetch private modules, uncomment the next line:
# RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o bin/gitea-backup ./cmd/gitea-backup && \
    go build -o bin/gitea-restore ./cmd/gitea-restore

# ---------- Runtime stage ----------
FROM ubuntu:26.04

ARG DEBIAN_FRONTEND=noninteractive
# Keep the client at the newest PostgreSQL server major supported by this image.
# Newer pg_dump clients can dump older supported servers, but older clients
# refuse to dump newer servers.
ARG PG_MAJOR=18

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
    && pg_dump --version | grep -Eq "^pg_dump \(PostgreSQL\) ${PG_MAJOR}\." \
    && apt-get purge -y gnupg lsb-release \
    && rm -rf /var/lib/apt/lists/*

RUN groupmod --new-name git ubuntu \
    && usermod --login git --home /data/git ubuntu

# Copy Go binaries from builder stage
COPY --from=builder /app/bin/gitea-backup /usr/local/bin/
COPY --from=builder /app/bin/gitea-restore /usr/local/bin/

# Optional: show the installed pg_dump version at container start
# (handy for debugging images)
# CMD ["bash", "-lc", "pg_dump --version && sleep infinity"]

CMD [ "sleep", "infinity" ]
