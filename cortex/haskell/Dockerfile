ARG TAG_PATH_ROOT
FROM ${TAG_PATH_ROOT}/050-c:local

RUN sudo \
  apk add --no-cache \
    ghc

COPY --chown=human:human ./files /hello-world
