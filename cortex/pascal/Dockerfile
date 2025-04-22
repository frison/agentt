ARG TAG_PATH_ROOT
FROM ${TAG_PATH_ROOT}/050-c:local

# RUN sudo \
#  apk add --no-cache \
#    fpc

COPY --chown=human:human ./artifacts /artifacts
RUN cd /artifacts && make install
ENV PATH=/home/human/.local/bin:$PATH
COPY --chown=human:human ./files /hello-world
