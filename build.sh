#!/usr/bin/env bash
set -euo pipefail

# This script only runs when SQLSMITH_GO_CONTAINER_TYPE is set.
if [ -z "${SQLSMITH_GO_CONTAINER_TYPE:-}" ]; then
    echo "ERROR: SQLSMITH_GO_CONTAINER_TYPE is not set. This script can only run when SQLSMITH_GO_CONTAINER_TYPE is defined." >&2
    exit 1
fi

mkdir -p output

go mod vendor

go build -v \
    -asan -o output/turso_embedded_executor cmd/executors/turso_embedded.go