ARG TAG_PATH_ROOT
# syntax=docker/dockerfile:1
# escape=\
FROM ${TAG_PATH_ROOT}/300-node:local

RUN sudo apk add gcompat gmp && \
  sudo npm install -g purescript bower pulp

COPY --chown=human:human ./files /hello-world
