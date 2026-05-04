#!/usr/bin/env sh

set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)

if [ "${DATABASE_URL:-}" = "" ]; then
  echo "DATABASE_URL is required." >&2
  echo "Example: export DATABASE_URL='postgres://postgres:postgres@localhost:5432/repocompass?sslmode=disable'" >&2
  exit 1
fi

"$SCRIPT_DIR/migrate-down.sh" -all
"$SCRIPT_DIR/migrate-up.sh"
