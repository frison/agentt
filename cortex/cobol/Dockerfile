ARG TAG_PATH_ROOT
# syntax=docker/dockerfile:1
# escape=\
FROM ${TAG_PATH_ROOT}/050-c:local

RUN sudo \
  apk add --no-cache \
    gmp-dev db-dev

COPY --chown=human:human ./artifacts /artifacts
RUN cd /artifacts && sudo make install

COPY --chown=human:human ./files /hello-world
