#!/usr/bin/env sh

set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
BACKEND_DIR=$(CDPATH= cd -- "$SCRIPT_DIR/../.." && pwd)

if [ "${DATABASE_URL:-}" = "" ]; then
  echo "DATABASE_URL is required." >&2
  echo "Example: export DATABASE_URL='postgres://postgres:postgres@localhost:55432/repocompass?sslmode=disable'" >&2
  exit 1
fi

cd "$BACKEND_DIR"
go run ./cmd/repocompass scan ./testdata/fixtures/local-repositories/missing-readme-repo --persist
go run ./cmd/repocompass scan ./testdata/fixtures/local-repositories/good-onboarding-repo --persist
