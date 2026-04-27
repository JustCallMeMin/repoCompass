#!/usr/bin/env sh

set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
BACKEND_DIR=$(CDPATH= cd -- "$SCRIPT_DIR/../.." && pwd)

if [ "${DATABASE_URL:-}" = "" ]; then
  echo "DATABASE_URL is required." >&2
  echo "Example: export DATABASE_URL='postgres://postgres:postgres@localhost:5432/repocompass?sslmode=disable'" >&2
  exit 1
fi

cleanup() {
  if [ "${POSTGRES_TEST_ROLLBACK:-}" = "1" ]; then
    "$SCRIPT_DIR/migrate-down.sh" -all >/dev/null 2>&1 || true
  fi
}
trap cleanup EXIT

"$SCRIPT_DIR/migrate-up.sh"

cd "$BACKEND_DIR"
go test ./internal/storage/postgres/...
