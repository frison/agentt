#!/usr/bin/env sh

set -euo pipefail

# --- Environment Variables (Expected from nhi-entrypoint-helper) ---
# The nhi-entrypoint-helper has already validated that the corresponding
# input paths exist and output paths exist and are writable.
# It has exported environment variables based on manifest keys:
# - INPUT_CONTENT=/app/input_content
# - INPUT_CONFIG=/app/input_config
# - OUTPUT_STATIC_SITE=/app/output_static_site

echo "Using derived paths from environment:"
echo "  Content Source: ${INPUT_CONTENT}"   # Use var without _DIR
echo "  Config Source:  ${INPUT_CONFIG}"    # Use var without _DIR
echo "  Site Dest:      ${OUTPUT_STATIC_SITE}" # Use var without _DIR

# Use the variables directly
CONTENT_SRC_DIR="${INPUT_CONTENT}"     # Use var without _DIR
CONFIG_SRC_DIR="${INPUT_CONFIG}"      # Use var without _DIR
SITE_DEST_DIR="${OUTPUT_STATIC_SITE}"   # Use var without _DIR

# --- Build Logic --- #

THEME_DIR="/app/themes/default/blog" # Path to the baked-in theme
BUILD_DIR="/tmp/jekyll_build_$$_${RANDOM}" # Temporary, writable build directory

echo "Setting up temporary build directory: ${BUILD_DIR}"
mkdir -p "${BUILD_DIR}"

# Copy theme files to the temporary build directory
# Use the /." pattern to copy the *contents* of the theme dir
# cp -R "${THEME_DIR}" "${BUILD_DIR}/" # Incorrect: copies dir itself
cp -R "${THEME_DIR}/." "${BUILD_DIR}/" # Correct: copies contents

# Change into the temporary build directory
cd "${BUILD_DIR}"
ls -alh
# --- Configuration Handling (in BUILD_DIR) ---
if [ ! -f "_config.yml" ]; then
  echo "Error: Theme base _config.yml not found in ${BUILD_DIR} after copy." >&2
  exit 1
fi
cp _config.yml _config.temp.yml
if [ -f "${CONFIG_SRC_DIR}/_config.yml" ]; then
  echo -e "\n# User Config from ${CONFIG_SRC_DIR}/_config.yml:" >> _config.temp.yml
  cat "${CONFIG_SRC_DIR}/_config.yml" >> _config.temp.yml
else
  echo "Warning: No user _config.yml found in ${CONFIG_SRC_DIR}" >&2
fi

# --- Content Handling (copy into BUILD_DIR) ---
echo "Copying user content from ${CONTENT_SRC_DIR} to ${BUILD_DIR}"
# Use -T with cp to copy contents, overwriting theme files if names clash
cp -R "${CONTENT_SRC_DIR}/." "${BUILD_DIR}/"

# Specific handling for about page if present (legacy)
if [ -f "${BUILD_DIR}/about.md" ]; then
    mv "${BUILD_DIR}/about.md" "${BUILD_DIR}/about.markdown"
fi

# --- Build Step (Source is BUILD_DIR) ---
echo "Running Jekyll build..."
# Ensure SITE_DEST_DIR exists (Helper validates, but good practice)
mkdir -p "${SITE_DEST_DIR}"

# Set BUNDLE_GEMFILE to point to where bundle install was run
export BUNDLE_GEMFILE="/app/Gemfile"

JEKYLL_ENV=production bundle exec jekyll build --source "${BUILD_DIR}" --destination "${SITE_DEST_DIR}" --config "${BUILD_DIR}/_config.temp.yml"

# Clean up temporary build directory (optional)
# echo "Cleaning up ${BUILD_DIR}"
# rm -rf "${BUILD_DIR}"

echo "Jekyll build complete. Output in ${SITE_DEST_DIR}"

# --- Ownership --- (Handled by entrypoint/runner)

echo "Process complete."
