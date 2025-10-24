FROM alpine:3.22

RUN sed -i 's#https\?://dl-cdn.alpinelinux.org/alpine#http://mirrors.tuna.tsinghua.edu.cn/alpine#g' /etc/apk/repositories && \
    apk update

RUN apk add --no-cache go bash

RUN go env -w GO111MODULE=on && \
    go env -w GOPROXY=https://goproxy.cn,direct

WORKDIR /opt