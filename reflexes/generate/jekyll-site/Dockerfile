# Bring in the common toolchain
FROM reflexes--base-tools:latest AS tools

# Use the standard NHI Ruby base image
FROM cortex/ruby:local AS base
USER root

# Install build dependencies (Ruby, C, Node)
RUN apk add --no-cache gcc g++ musl-dev make nodejs npm


# Copy common tools from .base-tools image
COPY --from=tools / /
# Set working directory
WORKDIR /app

# Copy reflex-specific files (scripts, theme, Gemfile)
COPY files/Gemfile /app/

# === Install Node.js dependencies for Tailwind ===
# Copy the entire tailwind config directory into the image
COPY files/tailwind /app/tailwind_build_config
# Set build config directory as workdir
WORKDIR /app/tailwind_build_config
# Install npm dependencies (including devDependencies like Tailwind and typography plugin)
RUN npm install
# Go back to the main app directory for subsequent steps
WORKDIR /app
# =================================================

# Install Jekyll and dependencies from Gemfile
# Ensure bundler is installed and use it
RUN gem install bundler && \
    cd /app && bundle install

# We do it this way to keep the above layers from changing
# and requiring another "gem install bundler" when
# unnecessary
COPY files/ /app/

# Copy the reflex manifest to the expected location for the helper
COPY manifest.yml /manifest.yml


# Ensure the main processing script is executable (MUST be done before switching user)
RUN chmod +x /app/process.sh

# Switch to the standard non-root user for execution
WORKDIR /app

# Use the standard NHI entrypoint helper
ENTRYPOINT ["/usr/local/bin/nhi-entrypoint-helper"]

# Default command for the helper to execute
CMD ["/app/process.sh"]

# Ensure the main processing script is executable
# RUN chmod +x /app/process.sh # Moved earlier