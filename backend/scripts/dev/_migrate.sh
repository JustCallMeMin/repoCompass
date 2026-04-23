#!/usr/bin/env sh

set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
BACKEND_DIR=$(CDPATH= cd -- "$SCRIPT_DIR/../.." && pwd)
MIGRATIONS_DIR="$BACKEND_DIR/db/migrations"
MIGRATE_VERSION="v4.19.1"

if [ "${DATABASE_URL:-}" = "" ]; then
  echo "DATABASE_URL is required." >&2
  echo "Example: export DATABASE_URL='postgres://user:pass@localhost:5432/dbname?sslmode=disable'" >&2
  exit 1
fi

exec go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@"$MIGRATE_VERSION" \
  -path "$MIGRATIONS_DIR" \
  -database "$DATABASE_URL" \
  "$@"
