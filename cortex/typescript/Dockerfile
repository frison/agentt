ARG TAG_PATH_ROOT
# syntax=docker/dockerfile:1
# escape=\
FROM ${TAG_PATH_ROOT}/300-node:local

RUN sudo npm install -g typescript

COPY --chown=human:human ./files /hello-world
