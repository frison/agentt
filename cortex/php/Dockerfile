ARG TAG_PATH_ROOT
# syntax=docker/dockerfile:1
# escape=\
#FROM 100hellos/100-java11:local
FROM ${TAG_PATH_ROOT}/000-base:local

RUN sudo \
  apk add --no-cache \
    php

COPY --chown=human:human ./files /hello-world
