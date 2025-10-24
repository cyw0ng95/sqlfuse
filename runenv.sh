#!/usr/bin/env bash

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
podman build -f Containerfile -t sqlsmith-go/dev .
podman run -it --rm \
  -v "$(pwd)":/opt:Z \
  sqlsmith-go/dev \
  bash