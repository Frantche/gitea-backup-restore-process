#!/usr/bin/env bash
set -euo pipefail

image="${IMAGE:-testimage:pr}"

user_entry="$(docker run --rm --entrypoint getent "$image" passwd git)"
group_entry="$(docker run --rm --entrypoint getent "$image" group git)"

if [[ "$user_entry" != "git:x:1000:1000:"*":/data/git:"* ]]; then
  echo "expected git user with uid 1000, gid 1000, and home /data/git; got: ${user_entry}" >&2
  exit 1
fi

if [[ "$group_entry" != "git:x:1000:"* ]]; then
  echo "expected git group with gid 1000; got: ${group_entry}" >&2
  exit 1
fi

docker run --rm --tmpfs /data --entrypoint sh "$image" -c \
  'touch /data/restored && chown git:git /data/restored && test "$(stat -c %u:%g /data/restored)" = "1000:1000"'

printf 'Gitea runtime user passed: %s\n' "$user_entry"
