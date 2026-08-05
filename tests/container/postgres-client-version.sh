#!/usr/bin/env bash
set -euo pipefail

image="${IMAGE:-testimage:pr}"
expected_major="${EXPECTED_PG_MAJOR:-18}"

version="$(docker run --rm --entrypoint pg_dump "$image" --version)"
if [[ "$version" != "pg_dump (PostgreSQL) ${expected_major}."* ]]; then
  echo "expected pg_dump major ${expected_major}, got: ${version}" >&2
  exit 1
fi

printf 'PostgreSQL client version passed: %s\n' "$version"
