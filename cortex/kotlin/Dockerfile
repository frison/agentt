ARG TAG_PATH_ROOT
# syntax=docker/dockerfile:1
# escape=\
FROM ${TAG_PATH_ROOT}/100-java11:local

RUN sudo \
  apk add --no-cache \
    bash zip

# There countless reasons why you should never do this for a production
# image, or on your own machine. I'm less concerned about it inside
# a Dockerfile for this purpose
RUN curl -s "https://get.sdkman.io" | bash
RUN bash -c "source /home/human/.sdkman/bin/sdkman-init.sh && sdk install kotlin"

COPY --chown=human:human ./files /hello-world
