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

mkdir -p .cache/go

podman build -f Containerfile -t sqlsmith-go/dev .
podman run -it --rm \
  -p 8080:8080 \
  -v "$(pwd)":/opt:Z \
  -v "$(pwd)/.cache/go":/root/go:Z \
  sqlsmith-go/dev \
  bash