ARG TAG_PATH_ROOT
# syntax=docker/dockerfile:1
# escape=\
FROM ${TAG_PATH_ROOT}/000-base:local

RUN sudo \
  apk add --no-cache \
    apache2 && \
  sudo rm -rf /var/www/localhost/htdocs

COPY --chown=human:human ./files /hello-world

RUN sudo \
  ln -s /hello-world/www /var/www/localhost/htdocs