ARG TAG_PATH_ROOT
# syntax=docker/dockerfile:1
# escape=\
FROM ${TAG_PATH_ROOT}/000-base:local

ENV PATH=/usr/lib/dart/bin:$PATH

COPY --chown=human:human ./artifacts /artifacts
RUN cd /artifacts && make install

COPY --chown=human:human ./files /hello-world
