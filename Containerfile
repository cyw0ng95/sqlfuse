FROM debian:trixie-slim

COPY debian.sources /etc/apt/sources.list.d/debian.sources

RUN apt-get update

RUN apt-get install golang ca-certificates clang libclang-rt-dev \
        -y --no-install-recommends && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Install curl and xz-utils to fetch and extract Node.js tarball
RUN apt-get update && \
    apt-get install -y --no-install-recommends curl xz-utils ca-certificates && \
    rm -rf /var/lib/apt/lists/*

RUN go env -w GO111MODULE=on && \
    go env -w GOPROXY=http://goproxy.cn,direct

# Install Node.js (prebuilt binary)
ENV NODE_VERSION=24.1.0
ENV NODE_DISTRO=node-v${NODE_VERSION}-linux-x64
ENV NODE_TARBALL=https://mirrors.tuna.tsinghua.edu.cn/nodejs-release/v${NODE_VERSION}/${NODE_DISTRO}.tar.xz

RUN set -eux; \
    curl -fsSL "$NODE_TARBALL" -o /tmp/${NODE_DISTRO}.tar.xz; \
    tar -xJf /tmp/${NODE_DISTRO}.tar.xz -C /usr/local; \
    # Create a stable symlink and ensure executable path
    ln -sfn /usr/local/${NODE_DISTRO} /usr/local/node; \
    chmod -R a+rX /usr/local/${NODE_DISTRO}; \
    rm -f /tmp/${NODE_DISTRO}.tar.xz

ENV PATH=/usr/local/node/bin:${PATH}
ENV NODE_HOME=/usr/local/node

# Configure npm to use the TUNA / npmmirror registry mirror
# Prefer the registry URL that is used by the mirror service.
RUN npm config set registry https://registry.npmmirror.com/ --location=global && \
    npm install -g pnpm@latest-10 && \
    pnpm config set registry https://registry.npmmirror.com/

ENV NUXT_TELEMETRY_DISABLED=1

ENV SQLSMITH_GO_CONTAINER_TYPE=dev
ENV CC=clang
ENV LD_LIBRARY_PATH=/opt/vendor/github.com/tursodatabase/turso-go/libs/linux_amd64/

WORKDIR /opt