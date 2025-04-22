ARG TAG_PATH_ROOT
# syntax=docker/dockerfile:1
# escape=\
FROM ${TAG_PATH_ROOT}/000-base:local

RUN sudo \
  apk add --no-cache \
    ocaml

COPY --chown=human:human ./files /hello-world
