# syntax=docker/dockerfile:1
# escape=\
FROM 100hellos/100-java11:local

RUN sudo \
  apk add --no-cache \
    gcompat bash

COPY --chown=human:human ./artifacts /artifacts
RUN cd /artifacts && make install

COPY --chown=human:human ./files /hello-world
