#!/usr/bin/env bash
set -euo pipefail

# If SQLSMITH_GO_CONTAINER_TYPE is set, skip running inside a container.
if [ -n "${SQLSMITH_GO_CONTAINER_TYPE:-}" ]; then
    echo "SQLSMITH_GO_CONTAINER_TYPE is set; skipping container build/run"
    # If sourced, return; if executed, exit.
    if [ "${BASH_SOURCE[0]}" != "${0}" ]; then
        return 0
    else
        exit 0
    fi
fi

mkdir -p .cache/go

docker build -f Containerfile -t sqlsmith-go/dev .

# Common Docker arguments used for both interactive and non-interactive runs.
# We assemble them into a bash array to avoid duplicating the long list of options.
DOCKER_IMAGE="sqlsmith-go/dev"
DOCKER_COMMON_ARGS=(--rm 
    -p 8080:8080
    -p 3000:3000
    -v "$(pwd)":/opt:Z
    -v "$(pwd)/.cache/go":/root/go:Z
)

# If a command is provided to this script, forward it to the container and run it
# non-interactively. Otherwise, open an interactive bash shell inside the container.
if [ "$#" -gt 0 ]; then
    docker run "${DOCKER_COMMON_ARGS[@]}" "$DOCKER_IMAGE" "$@"
else
    # Add -it for interactive shells.
    docker run -it "${DOCKER_COMMON_ARGS[@]}" "$DOCKER_IMAGE" bash
fi