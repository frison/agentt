# Stage 1: Get tools from the base-tools image
# Replace 'reflexes:base-tools' with the actual image name/tag if different
FROM reflexes:base-tools AS tools

# Stage 2: Build/runtime environment (example using Python)
# Choose a minimal base image suitable for the reflex language
FROM python:3.9-alpine AS final

# Set up the user and working directory (ensure nhi user exists or use root)
# Alpine images don't have adduser by default, common pattern is to run as root
# or use a fuller base image if non-root is strictly required.
USER root
WORKDIR /app

# Copy the toolset from the tools stage, overlaying the root filesystem
COPY --from=tools / /

# Copy the mandatory manifest file to the root
COPY manifest.yml /manifest.yml

# Install language-specific dependencies first for better caching
COPY files/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application code (assuming it's in a files/ directory)
# Adjust if the template code structure is different
COPY files/ .

# Ensure the main script is executable (if applicable)
# Adjust 'main.py' if the primary script file has a different name
RUN chmod +x main.py

# Default ENTRYPOINT uses the nhi-entrypoint-helper to provide
# usage instructions based on manifest.yml and run the main script
# Adjust 'python' and 'main.py' based on the actual reflex language/entrypoint
ENTRYPOINT ["/usr/local/bin/nhi-entrypoint-helper", "python", "main.py"]