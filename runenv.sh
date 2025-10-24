#!/usr/bin/env bash

podman build -f Containerfile -t sqlsmith-go/dev .
podman run -it --rm \
  -v "$(pwd)":/opt:Z \
  sqlsmith-go/dev \
  bash