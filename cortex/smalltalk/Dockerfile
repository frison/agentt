ARG TAG_PATH_ROOT
# syntax=docker/dockerfile:1
# escape=\
FROM ${TAG_PATH_ROOT}/050-c:local

RUN sudo \
  apk add --no-cache \
    gawk \
    zip \
    patch

COPY --chown=human:human ./artifacts /artifacts
RUN cd /artifacts && make install

# We install smalltalk into our local user account's home directory
# So we add that to the path.
ENV PATH=/home/human/.local/bin:$PATH
COPY --chown=human:human ./files /hello-world
