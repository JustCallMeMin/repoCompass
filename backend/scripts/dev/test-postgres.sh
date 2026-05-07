#!/usr/bin/env sh

set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
BACKEND_DIR=$(CDPATH= cd -- "$SCRIPT_DIR/../.." && pwd)

if [ "${DATABASE_URL:-}" = "" ]; then
  echo "DATABASE_URL is required." >&2
  echo "Example: export DATABASE_URL='postgres://postgres:postgres@localhost:55432/repocompass?sslmode=disable'" >&2
  exit 1
fi

cleanup() {
  if [ "${POSTGRES_TEST_ROLLBACK:-}" = "1" ]; then
    "$SCRIPT_DIR/migrate-down.sh" -all >/dev/null 2>&1 || true
  fi
}
trap cleanup EXIT

"$SCRIPT_DIR/migrate-up.sh"
"$SCRIPT_DIR/migrate-down.sh" -all
"$SCRIPT_DIR/migrate-up.sh"

cd "$BACKEND_DIR"
go test ./internal/storage/postgres/...

scan_output="$(go run ./cmd/repocompass scan ./testdata/fixtures/local-repositories/good-onboarding-repo --persist)"
printf '%s\n' "$scan_output"

repository_id="$(printf '%s\n' "$scan_output" | awk -F': ' '/Repository ID:/ {print $2; exit}' | tr -d '[:space:]')"
scan_id="$(printf '%s\n' "$scan_output" | awk -F': ' '/Scan ID:/ {print $2; exit}' | tr -d '[:space:]')"

if [ "$repository_id" = "" ] || [ "$scan_id" = "" ]; then
  echo "failed to parse persisted scan IDs from CLI output" >&2
  exit 1
fi

go run ./cmd/repocompass history "$repository_id" --format json >/dev/null
go run ./cmd/repocompass findings "$scan_id" --format json >/dev/null

if DATABASE_URL= go run ./cmd/repocompass history "$repository_id" >/dev/null 2>&1; then
  echo "expected history command to fail without DATABASE_URL" >&2
  exit 1
fi
