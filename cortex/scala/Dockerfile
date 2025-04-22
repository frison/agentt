ARG TAG_PATH_ROOT
# syntax=docker/dockerfile:1
# escape=\
FROM ${TAG_PATH_ROOT}/100-java11:local

RUN sudo \
  apk add --no-cache \
    gcompat bash

COPY --chown=human:human ./artifacts /artifacts
RUN cd /artifacts && make install

COPY --chown=human:human ./files /hello-world
