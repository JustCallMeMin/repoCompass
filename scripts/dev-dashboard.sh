#!/usr/bin/env sh
set -eu

# log_info prints one local dashboard startup step.
log_info() {
  printf 'INFO %s\n' "$1"
}

log_info "start PostgreSQL"
make db-up

log_info "apply migrations"
make migrate-up

log_info "seed dashboard data"
make db-seed

log_info "start API and dashboard in separate terminals:"
printf 'DEV_HEADER_AUTH=true make server\n'
printf 'cd frontend && npm run dev\n'
