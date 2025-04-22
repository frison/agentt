ARG TAG_PATH_ROOT
# syntax=docker/dockerfile:1
# escape=\
FROM ${TAG_PATH_ROOT}/100-java11:local

COPY --chown=human:human ./artifacts /artifacts
RUN cd /artifacts && make install
ENV OBERON_BIN=/artifacts/oberonc/bin

COPY --chown=human:human ./files /hello-world
