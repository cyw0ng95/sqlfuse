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

# Build options: support plain docker build (default) or docker buildx
# Enable buildx by setting SQLSMITH_GO_BUILDX=1
# To push the built image (requires a registry in DOCKER_IMAGE), set SQLSMITH_GO_BUILDX_PUSH=1
# Optional env vars:
#  SQLSMITH_GO_PLATFORMS (default: linux/amd64,linux/arm64)
#  SQLSMITH_GO_CACHE_REGISTRY (optional, used for --cache-from/--cache-to when pushing)
#  DOCKER_IMAGE (override default image name)

DOCKER_IMAGE="${DOCKER_IMAGE:-cyw0ng95/sqlsmith-go}"

# Allow callers (CI) to skip building the image if it's already prepared in the environment.
# Set SQLSMITH_GO_SKIP_BUILD=1 to skip the build step and use the existing image.
if [ "${SQLSMITH_GO_SKIP_BUILD:-0}" = "1" ]; then
    echo "SQLSMITH_GO_SKIP_BUILD=1: skipping image build, using existing image: $DOCKER_IMAGE"
else
    if [ "${SQLSMITH_GO_BUILDX:-0}" = "1" ]; then
        # buildx path (single-arch only)
        BUILDER_NAME="${SQLSMITH_GO_BUILDX_BUILDER:-default}"
        CACHE_REGISTRY="${SQLSMITH_GO_CACHE_REGISTRY:-}"

        # Enable buildkit
        export DOCKER_BUILDKIT=1

        # Determine a single local platform to build for (no multi-arch lists)
        arch=$(uname -m || true)
        case "$arch" in
            x86_64|amd64) LOCAL_PLATFORM="linux/amd64" ;;
            aarch64|arm64) LOCAL_PLATFORM="linux/arm64" ;;
            *) LOCAL_PLATFORM="linux/amd64" ;;
        esac

        if [ "${SQLSMITH_GO_BUILDX_PUSH:-0}" = "1" ]; then
            # Push single-platform image to registry. Optionally use cache registry.
            if [ -n "$CACHE_REGISTRY" ]; then
                docker buildx build --builder "$BUILDER_NAME" \
                    --platform "$LOCAL_PLATFORM" \
                    --tag "$DOCKER_IMAGE" \
                    --cache-from=type=registry,ref="$CACHE_REGISTRY" \
                    --cache-to=type=registry,mode=max,ref="$CACHE_REGISTRY" \
                    --push -f Containerfile .
            else
                docker buildx build --builder "$BUILDER_NAME" \
                    --platform "$LOCAL_PLATFORM" \
                    --tag "$DOCKER_IMAGE" \
                    --push -f Containerfile .
            fi
        else
            # Build for local platform and load into local docker
            echo "Building for local platform: $LOCAL_PLATFORM"
            docker buildx build --platform "$LOCAL_PLATFORM" --tag "$DOCKER_IMAGE" --load -f Containerfile .
        fi
    else
        # Traditional docker build (single-platform image)
        docker build -f Containerfile -t "$DOCKER_IMAGE" .
    fi
fi


# Common Docker arguments used for both interactive and non-interactive runs.
# We assemble them into a bash array to avoid duplicating the long list of options.
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