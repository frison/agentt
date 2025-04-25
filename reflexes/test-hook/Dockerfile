ARG BASE_IMAGE_TAG=latest
FROM reflexes/base:${BASE_IMAGE_TAG}

# Copy hook scripts into the expected location
COPY hooks/ /hooks/

# Ensure hooks are executable
RUN chmod +x /hooks/build.sh && \
    chmod +x /hooks/entrypoint.sh

# Optionally override the default CMD from the base image
# CMD ["some-other-command"]