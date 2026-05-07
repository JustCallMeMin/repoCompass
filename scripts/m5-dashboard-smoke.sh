#!/usr/bin/env sh
set -eu

# require_env verifies a required environment variable exists.
require_env() {
  name="$1"
  eval "value=\${$name:-}"
  if [ -z "$value" ]; then
    printf 'ERROR %s is required\n' "$name" >&2
    exit 1
  fi
}

# check_url verifies one dashboard route returns HTML.
check_url() {
  url="$1"
  curl -fsS "$url" >/tmp/repocompass-m5-smoke.html
}

require_env DASHBOARD_URL

check_url "$DASHBOARD_URL/dashboard"
check_url "$DASHBOARD_URL/repositories"

printf 'INFO M5 dashboard smoke completed\n'
