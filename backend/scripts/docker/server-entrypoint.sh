#!/usr/bin/env sh

set -eu

: "${DATABASE_URL:?DATABASE_URL is required}"

MIGRATIONS_PATH="${MIGRATIONS_PATH:-/app/db/migrations}"

echo "Applying database migrations from ${MIGRATIONS_PATH}"
migrate -path "$MIGRATIONS_PATH" -database "$DATABASE_URL" up

exec repocompass-server
