#!/usr/bin/env sh

set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
BACKEND_DIR=$(CDPATH= cd -- "$SCRIPT_DIR/../.." && pwd)
THRESHOLD="${BACKEND_COVERAGE_THRESHOLD:-45}"
COVERAGE_FILE="${BACKEND_COVERAGE_FILE:-coverage.out}"

cd "$BACKEND_DIR"

go test ./...

packages=$(go list -f '{{if .TestGoFiles}}{{.ImportPath}}{{end}}' ./... | awk 'NF')
go test $packages -coverprofile="$COVERAGE_FILE"

coverage=$(go tool cover -func="$COVERAGE_FILE" | awk '/^total:/ {gsub("%", "", $3); print $3}')

awk -v coverage="$coverage" -v threshold="$THRESHOLD" 'BEGIN {
	if (coverage + 0 < threshold + 0) {
		printf("coverage %.1f%% below threshold %.1f%%\n", coverage, threshold)
		exit 1
	}
	printf("coverage %.1f%% meets threshold %.1f%%\n", coverage, threshold)
}'
