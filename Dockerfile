# ---------- Build stage ----------
FROM golang:1.24-alpine AS builder

WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY go.mod go.sum ./
RUN go mod download

COPY . .
ENV CGO_ENABLED=0
RUN go build -trimpath -ldflags="-s -w" -o bin/gitea-backup ./cmd/gitea-backup && \
    go build -trimpath -ldflags="-s -w" -o bin/gitea-restore ./cmd/gitea-restore

# ---------- PostgreSQL 17 client stage ----------
FROM postgres:17-alpine AS pgtools

# ---------- Runtime stage ----------
FROM alpine:3.22
# Minimal tools + MariaDB CLI + runtime libs required by pg_dump/libpq
RUN apk add --no-cache \
      ca-certificates curl jq \
      mariadb-client \
      lz4-libs \
      zstd-libs \
      krb5-libs \
      openldap \
      libedit

# Copy only what you need from PG17 (keeps size down)
COPY --from=pgtools /usr/local/bin/pg_dump /usr/local/bin/
COPY --from=pgtools /usr/local/bin/pg_restore /usr/local/bin/
COPY --from=pgtools /usr/local/bin/psql /usr/local/bin/

# PG runtime libs/messages
COPY --from=pgtools /usr/local/lib /usr/local/lib
COPY --from=pgtools /usr/local/share/postgresql /usr/local/share/postgresql

# Ensure dynamic linker finds PG libs first
ENV LD_LIBRARY_PATH=/usr/local/lib

# Your Go binaries (built in a previous stage)
COPY --from=builder /app/bin/gitea-backup /usr/local/bin/
COPY --from=builder /app/bin/gitea-restore /usr/local/bin/

CMD ["sleep", "infinity"]