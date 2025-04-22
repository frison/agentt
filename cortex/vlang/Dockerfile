ARG TAG_PATH_ROOT
# syntax=docker/dockerfile:1
# escape=\
FROM ${TAG_PATH_ROOT}/050-c:local

RUN sudo \
  apk add --no-cache \
    build-base

RUN sudo \
  apk add vlang --repository=https://dl-cdn.alpinelinux.org/alpine/edge/testing

COPY --chown=human:human ./files /hello-world
