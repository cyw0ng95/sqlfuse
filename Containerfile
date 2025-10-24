FROM debian:trixie-slim

COPY debian.sources /etc/apt/sources.list.d/debian.sources

RUN apt-get update

RUN apt-get install golang ca-certificates clang libclang-rt-dev \
        -y --no-install-recommends && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

RUN go env -w GO111MODULE=on && \
    go env -w GOPROXY=http://goproxy.cn,direct

ENV SQLSMITH_GO_CONTAINER_TYPE=dev
ENV CC=clang
ENV LD_LIBRARY_PATH=/opt/vendor/github.com/tursodatabase/turso-go/libs/linux_amd64/

WORKDIR /opt