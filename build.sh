#!/usr/bin/env bash
set -euo pipefail

# This script only runs when SQLSMITH_GO_CONTAINER_TYPE is set.
if [ -z "${SQLSMITH_GO_CONTAINER_TYPE:-}" ]; then
    echo "ERROR: SQLSMITH_GO_CONTAINER_TYPE is not set. This script can only run when SQLSMITH_GO_CONTAINER_TYPE is defined." >&2
    exit 1
fi

run_tests() {
    echo "-- [INFO] Running tests with coverage..."
    mkdir -p .cache
    go test -coverprofile=.cache/coverage.out ./...
    go tool cover -html=.cache/coverage.out -o .cache/coverage.html
    echo "-- [INFO] Coverage report generated: .cache/coverage.html"
}

build_project() {
    echo "-- [INFO] Starting build..."
    mkdir -p output
    echo "-- [INFO] Running go mod vendor..."
    go mod vendor
    echo "-- [INFO] Building turso_embedded_executor..."
    go build \
        -asan -o output/turso_embedded_executor cmd/executors/turso_embedded.go
    go build \
        -asan -o output/server ./cmd/server/main.go
    echo "-- [INFO] Build complete. Output: output/turso_embedded_executor, output/server"
}

# Parse arguments with getopt
OPTS=$(getopt -o t --long test -n 'build.sh' -- "$@")
eval set -- "$OPTS"
run_test=0

while true; do
  case "$1" in
    -t|--test)
      run_test=1
      shift ;;
    --)
      shift ; break ;;
    *)
      echo "Internal error!" ; exit 1 ;;
  esac
done

if [[ $run_test -eq 1 ]]; then
    run_tests && build_project
    exit 0
fi

build_project