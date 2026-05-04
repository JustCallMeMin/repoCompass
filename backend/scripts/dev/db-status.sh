#!/usr/bin/env sh

set -eu

if [ "${DATABASE_URL:-}" = "" ]; then
  echo "DATABASE_URL is required." >&2
  echo "Example: export DATABASE_URL='postgres://postgres:postgres@localhost:55432/repocompass?sslmode=disable'" >&2
  exit 1
fi

psql "$DATABASE_URL" -c "SELECT current_database() AS database, current_user AS user, NOW() AS checked_at;"
