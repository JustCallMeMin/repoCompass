#!/usr/bin/env sh

set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
BACKEND_DIR=$(CDPATH= cd -- "$SCRIPT_DIR/../.." && pwd)
MIGRATIONS_DIR="$BACKEND_DIR/db/migrations"
MIGRATE_VERSION="v4.19.1"
ENV_FILE="$BACKEND_DIR/.env"

if [ -f "$ENV_FILE" ]; then
  set -a
  # shellcheck disable=SC1090
  . "$ENV_FILE"
  set +a
fi

if [ "${DATABASE_URL:-}" = "" ]; then
  echo "DATABASE_URL is required." >&2
  echo "Set it directly in the shell or in backend/.env." >&2
  echo "Example: export DATABASE_URL='postgres://postgres:postgres@localhost:5432/repocompass?sslmode=disable'" >&2
  exit 1
fi

exec go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@"$MIGRATE_VERSION" \
  -path "$MIGRATIONS_DIR" \
  -database "$DATABASE_URL" \
  "$@"
