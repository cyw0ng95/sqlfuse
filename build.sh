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
    # Add -v flag when running in GitHub Actions for verbose output
    if [ "${GITHUB_ACTIONS:-false}" = "true" ]; then
        go test -v -coverprofile=.cache/coverage.out ./...
    else
        go test -coverprofile=.cache/coverage.out ./...
    fi
    go tool cover -html=.cache/coverage.out -o .cache/coverage.html
    echo "-- [INFO] Coverage report generated: .cache/coverage.html"
}

build_view() {
    echo "-- [INFO] Building view (frontend)..."
    if [[ -d view ]]; then
        (cd view && pnpm install)
        (cd view && pnpm run build)
        echo "-- [INFO] View build complete. Output: view/dist"
    else
        echo "-- [WARN] view directory not found; skipping frontend build"
    fi
}

build_project() {
    echo "-- [INFO] Starting build..."
    mkdir -p output
    echo "-- [INFO] Running go mod vendor..."
    go mod vendor
    
    # Add -v flag when running in GitHub Actions for verbose output
    local build_flags=()
    if [ "${GITHUB_ACTIONS:-false}" = "true" ]; then
        build_flags=(-v)
    fi
    
    echo "-- [INFO] Building turso_embedded_executor..."
    go build "${build_flags[@]}" -o output/turso_embedded_executor cmd/executors/turso_embedded/main.go
    echo "-- [INFO] Building go_sqlite3_embedded_executor..."
    go build "${build_flags[@]}" -o output/go_sqlite3_embedded_executor cmd/executors/go_sqlite3_embedded/main.go
    go build "${build_flags[@]}" -o output/server ./cmd/server
    echo "-- [INFO] Build complete. Output: output/turso_embedded_executor, output/go_sqlite3_embedded_executor, output/server"
    
    # Build the frontend view
    build_view
}

# Function to start frontend dev server and the Go server, with cleanup trap
start_services() {
    echo "-- [INFO] Starting frontend dev server..."
    frontend_pid=""

    if [[ -d view ]]; then
        # install deps first
        (cd view && pnpm install)
        # start dev server in background and capture its pid
        (cd view && pnpm run dev) &
        frontend_pid=$!
        mkdir -p output
        echo "$frontend_pid" > output/frontend.pid
        echo "-- [INFO] Frontend dev server started (PID $frontend_pid)"
    else
        echo "-- [WARN] view directory not found; skipping frontend dev server"
    fi

    # Ensure frontend is cleaned up when this script exits or is interrupted
    cleanup() {
        if [[ -n "${frontend_pid:-}" ]]; then
            echo "-- [INFO] Stopping frontend (PID $frontend_pid)"
            kill "$frontend_pid" 2>/dev/null || true
            wait "$frontend_pid" 2>/dev/null || true
            rm -f output/frontend.pid || true
        fi
    }
    trap cleanup EXIT INT TERM

    echo "-- [INFO] Starting server..."
    ./output/server
}

# Parse arguments with getopt
OPTS=$(getopt -o tr --long test,run -n 'build.sh' -- "$@")
eval set -- "$OPTS"
run_test=0
run_server=0

while true; do
  case "$1" in
    -t|--test)
      run_test=1
      shift ;;
    -r|--run)
      run_server=1
      shift ;;
    --)
      shift ; break ;;
    *)
      echo "Internal error!" ; exit 1 ;;
  esac
done

if [[ $run_test -eq 1 ]]; then
    run_tests
    build_project
else
    build_project
fi

# If requested, start the built server in the background and write pid/log
if [[ $run_server -eq 1 ]]; then
    start_services
fi