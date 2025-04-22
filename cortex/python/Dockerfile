ARG TAG_PATH_ROOT
# syntax=docker/dockerfile:1
# escape=\
FROM ${TAG_PATH_ROOT}/000-base:local

RUN sudo \
  apk add \
    python3-dev 

COPY --chown=human:human ./files /hello-world
