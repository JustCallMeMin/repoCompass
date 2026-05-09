#!/usr/bin/env bash

set -euo pipefail

DEMO_DIR="${REPOCOMPASS_DEMO_DIR:-/tmp/repocompass-demo}"
CLONE_TIMEOUT_SECONDS="${REPOCOMPASS_DEMO_CLONE_TIMEOUT_SECONDS:-180}"

log() {
    printf '[INFO] %s\n' "$*"
}

fail() {
    printf '[ERROR] %s\n' "$*" >&2
    exit 1
}

require_command() {
    command -v "$1" >/dev/null 2>&1 || fail "Required command not found: $1"
}

run_with_timeout() {
    if command -v timeout >/dev/null 2>&1; then
        timeout "$CLONE_TIMEOUT_SECONDS" "$@"
    else
        "$@"
    fi
}

clone_if_missing() {
    repo_url="$1"
    target_dir="$2"

    if [ -d "$target_dir/.git" ]; then
        log "Skipping existing repository: $target_dir"
        return 0
    fi

    if [ -e "$target_dir" ]; then
        fail "Target exists but is not a git repository: $target_dir"
    fi

    log "Cloning $repo_url into $target_dir"
    run_with_timeout git clone --depth 1 "$repo_url" "$target_dir" || fail "Clone failed: $repo_url"
}

require_command git

log "Preparing demo fixtures in $DEMO_DIR"
mkdir -p "$DEMO_DIR" || fail "Cannot create demo directory: $DEMO_DIR"

clone_if_missing https://github.com/kubernetes/kubernetes "$DEMO_DIR/kubernetes"
clone_if_missing https://github.com/expressjs/express "$DEMO_DIR/express"
clone_if_missing https://github.com/pallets/flask "$DEMO_DIR/flask"

log "Demo fixtures prepared at $DEMO_DIR"
