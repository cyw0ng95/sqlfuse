#!/usr/bin/env bash
set -euo pipefail

# This script only runs when SQLSMITH_GO_CONTAINER_TYPE is set.
if [ -z "${SQLSMITH_GO_CONTAINER_TYPE:-}" ]; then
    echo "ERROR: SQLSMITH_GO_CONTAINER_TYPE is not set. This script can only run when SQLSMITH_GO_CONTAINER_TYPE is defined." >&2
    exit 1
fi

run_tests() {
    echo "Running tests with coverage..."
    mkdir -p .cache
    go test -coverprofile=.cache/coverage.out ./...
    go tool cover -html=.cache/coverage.out -o .cache/coverage.html
    echo "Coverage report generated: .cache/coverage.html"
}

build_project() {
    mkdir -p output
    go mod vendor
    go build -v \
        -asan -o output/turso_embedded_executor cmd/executors/turso_embedded.go
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