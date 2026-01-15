#!/bin/sh
set -e

if [ -z "$DATABASE_URL" ]; then
  echo "DATABASE_URL is not set" >&2
  exit 1
fi

goose -dir /app/db/migrations postgres "$DATABASE_URL" up

exec /app/server
