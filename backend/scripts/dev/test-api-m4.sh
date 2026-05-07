#!/usr/bin/env sh
set -eu

# log_info writes one smoke-test info line.
log_info() {
  printf 'INFO %s\n' "$1"
}

# log_error writes one smoke-test error line.
log_error() {
  printf 'ERROR %s\n' "$1" >&2
}

# require_env verifies that one required variable exists.
require_env() {
  name="$1"
  eval "value=\${$name:-}"
  if [ -z "$value" ]; then
    log_error "$name is required"
    exit 1
  fi
}

# request sends an API request and fails on non-2xx responses.
request() {
  method="$1"
  path="$2"
  body="${3:-}"
  if [ -n "$body" ]; then
    curl -fsS -X "$method" "$BASE_URL$path" \
      -H 'Content-Type: application/json' \
      -H "X-User-Id: ${API_USER_ID:-mock_user}" \
      -H "X-Organization-Id: ${API_ORG_ID:-00000000-0000-0000-0000-000000000000}" \
      --data "$body"
  else
    curl -fsS -X "$method" "$BASE_URL$path" \
      -H "X-User-Id: ${API_USER_ID:-mock_user}" \
      -H "X-Organization-Id: ${API_ORG_ID:-00000000-0000-0000-0000-000000000000}"
  fi
}

require_env BASE_URL

log_info "checking health"
request GET /api/v1/health >/tmp/repocompass-health.json

log_info "checking repositories"
request GET /api/v1/repositories >/tmp/repocompass-repositories.json

log_info "checking direct scan"
request POST /api/v1/scans "{\"source_type\":\"local\",\"path\":\"${SCAN_PATH:-./testdata/fixtures/local-repositories/good-onboarding-repo}\"}" >/tmp/repocompass-scan.json

log_info "checking session"
request GET /api/v1/auth/session >/tmp/repocompass-session.json

log_info "M4 API smoke completed"
