# Build stage for manifest handler
FROM 100hellos/golang:latest as tool-builder
# Set working directory to the module root
WORKDIR /build

# Copy the entire consolidated module source
COPY src/basetools .

# Run the build script from within the specific cmd directory using sudo
RUN sudo cmd/manifest/build.sh
RUN sudo cmd/nhi-entrypoint-helper/build.sh

# Final stage
FROM scratch

# Copy manifest handler from its build location
COPY --from=tool-builder /build/cmd/manifest/manifest /usr/local/bin/manifest
COPY --from=tool-builder /build/cmd/nhi-entrypoint-helper/nhi-entrypoint-helper /usr/local/bin/nhi-entrypoint-helper

# Copy files (scripts, etc.) into the root
COPY files /

# This container is not "runnable" by itself, it is intended to be used as
# collection of tools for other containers.