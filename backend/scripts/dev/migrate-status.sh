#!/usr/bin/env sh

set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)

# golang-migrate exposes "version" rather than a "status" subcommand.
exec "$SCRIPT_DIR/_migrate.sh" version "$@"
