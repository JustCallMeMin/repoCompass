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
  PGHOST_VALUE="${PGHOST:-localhost}"
  PGPORT_VALUE="${PGPORT:-5432}"
  PGUSER_VALUE="${PGUSER:-}"
  PGDATABASE_VALUE="${PGDATABASE:-}"
  PGPASSWORD_VALUE="${PGPASSWORD:-}"

  if [ "$PGUSER_VALUE" != "" ] && [ "$PGDATABASE_VALUE" != "" ]; then
    DATABASE_URL="postgres://${PGUSER_VALUE}"

    if [ "$PGPASSWORD_VALUE" != "" ]; then
      DATABASE_URL="${DATABASE_URL}:${PGPASSWORD_VALUE}"
    fi

    DATABASE_URL="${DATABASE_URL}@${PGHOST_VALUE}:${PGPORT_VALUE}/${PGDATABASE_VALUE}?sslmode=disable"
  else
    echo "DATABASE_URL is required." >&2
    echo "Alternatively set PGUSER and PGDATABASE, with optional PGPASSWORD, PGHOST, and PGPORT." >&2
    echo "Example: export DATABASE_URL='postgres://user:pass@localhost:5432/dbname?sslmode=disable'" >&2
    exit 1
  fi
fi

exec go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@"$MIGRATE_VERSION" \
  -path "$MIGRATIONS_DIR" \
  -database "$DATABASE_URL" \
  "$@"
