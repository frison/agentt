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

# --- Setup Temporary Build Directory --- #
BUILD_DIR="/tmp/jekyll_build_$$_${RANDOM}" # Temporary, writable build directory
echo "Setting up temporary build directory: ${BUILD_DIR}"
mkdir -p "${BUILD_DIR}" # Create build dir (assets subdir not needed early anymore)

# --- Tailwind CSS Build (Outputting to /tmp first) ---
echo "Building Tailwind CSS..."
TAILWIND_CONFIG_DIR="/app/tailwind_build_config"
TAILWIND_INPUT="${TAILWIND_CONFIG_DIR}/input.css"
# TAILWIND_OUTPUT_DIR="${BUILD_DIR}/assets" # Old - output to build dir
TAILWIND_TMP_OUTPUT_FILE="/tmp/tailwind.css" # New - output to temp file

# Ensure the Tailwind config dir exists (should be copied by Dockerfile)
if [ ! -d "${TAILWIND_CONFIG_DIR}" ]; then
  echo "Error: Tailwind config directory ${TAILWIND_CONFIG_DIR} not found." >&2
  exit 1
fi

# Ensure node_modules exists (should be created by npm install in Dockerfile)
if [ ! -d "${TAILWIND_CONFIG_DIR}/node_modules" ]; then
  echo "Error: node_modules not found in ${TAILWIND_CONFIG_DIR}. Build image might be incomplete." >&2
  exit 1
fi

echo "Running Tailwind build from ${TAILWIND_CONFIG_DIR}..."
echo "Input: ${TAILWIND_INPUT}"
echo "Output: ${TAILWIND_TMP_OUTPUT_FILE} (temporary)"
echo "Content scanned: /app/input_content/ and /app/tailwind_build_config/ (defined in config)"

# Execute tailwind using npx from the directory where it was installed
(cd "${TAILWIND_CONFIG_DIR}" && npx tailwindcss -c tailwind.config.js -i "${TAILWIND_INPUT}" -o "${TAILWIND_TMP_OUTPUT_FILE}")

# Check if the expected output file was created
if [ ! -f "${TAILWIND_TMP_OUTPUT_FILE}" ]; then
  echo "Error: Tailwind build finished, but temporary output file ${TAILWIND_TMP_OUTPUT_FILE} was not created." >&2
  exit 1
fi
echo "Tailwind CSS build successful. Temporary output at ${TAILWIND_TMP_OUTPUT_FILE}"
# --- End Tailwind CSS Build ---

THEME_DIR="/app/themes/default/blog" # Path to the baked-in theme
# BUILD_DIR="/tmp/jekyll_build_$$_${RANDOM}" # Defined earlier

# echo "Setting up temporary build directory: ${BUILD_DIR}" # Done earlier
# mkdir -p "${BUILD_DIR}" # Done earlier

# Copy theme files to the temporary build directory
# Use the /." pattern to copy the *contents* of the theme dir
cp -R "${THEME_DIR}/." "${BUILD_DIR}/" # Correct: copies contents

# --- Copy Tailwind Theme Files (Layouts, Includes, Assets) into BUILD_DIR --- #
TAILWIND_THEME_SRC="/app/tailwind_build_config"
echo "Copying Tailwind theme files from ${TAILWIND_THEME_SRC} into ${BUILD_DIR}"
# Create target directories if they don't exist (belt and suspenders)
mkdir -p "${BUILD_DIR}/_layouts" "${BUILD_DIR}/_includes" "${BUILD_DIR}/assets/js"
# Copy layouts
if [ -d "${TAILWIND_THEME_SRC}/_layouts" ]; then
  cp -R "${TAILWIND_THEME_SRC}/_layouts/." "${BUILD_DIR}/_layouts/"
fi
# Copy includes
if [ -d "${TAILWIND_THEME_SRC}/_includes" ]; then
  cp -R "${TAILWIND_THEME_SRC}/_includes/." "${BUILD_DIR}/_includes/"
fi
# Copy assets (JS)
if [ -d "${TAILWIND_THEME_SRC}/assets" ]; then
  cp -R "${TAILWIND_THEME_SRC}/assets/." "${BUILD_DIR}/assets/"
fi
# Copy index.html if it exists
if [ -f "${TAILWIND_THEME_SRC}/index.html" ]; then
    cp "${TAILWIND_THEME_SRC}/index.html" "${BUILD_DIR}/"
fi
echo "Tailwind theme files copied."
# --- End Copy Tailwind Theme Files --- #


# Change into the temporary build directory
cd "${BUILD_DIR}"
ls -alh

# --- Configuration Handling (Using Base + User Config) ---
# Configuration will be handled by passing multiple --config flags to Jekyll
BASE_CONFIG="${BUILD_DIR}/_config.yml" # Base config copied from theme files
USER_CONFIG="${CONFIG_SRC_DIR}/_config.yml"

# Prepare the config string for the Jekyll command
CONFIG_STRING="${BASE_CONFIG}"
if [ -f "${USER_CONFIG}" ]; then
  echo "User config found at ${USER_CONFIG}, adding to Jekyll config list."
  CONFIG_STRING="${CONFIG_STRING},${USER_CONFIG}"
else
  echo "No user config found at ${USER_CONFIG}, using base config only."
fi

# --- Content Handling (copy into BUILD_DIR) ---
echo "Copying user content from ${CONTENT_SRC_DIR} to ${BUILD_DIR}"
# Use cp -R to copy contents, OVERWRITING theme/tailwind files if names clash
# This allows user overrides of layouts/includes/index.html etc.
cp -R "${CONTENT_SRC_DIR}/." "${BUILD_DIR}/"

# --- Explicitly remove base theme index markdown to prevent conflict --- #
BASE_INDEX_MD="${BUILD_DIR}/index.markdown"
if [ -f "${BASE_INDEX_MD}" ]; then
  echo "Removing base theme index markdown: ${BASE_INDEX_MD}"
  rm -f "${BASE_INDEX_MD}"
fi
# --- End Removal --- #

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

echo "Using config files: ${CONFIG_STRING}"
JEKYLL_ENV=production bundle exec jekyll build --source "${BUILD_DIR}" --destination "${SITE_DEST_DIR}" --config "${CONFIG_STRING}"

# --- Copy Tailwind CSS from Temp to Final Destination --- #
FINAL_CSS_DIR="${SITE_DEST_DIR}/assets"
FINAL_CSS_FILE="${FINAL_CSS_DIR}/tailwind.css"
echo "Copying Tailwind CSS from ${TAILWIND_TMP_OUTPUT_FILE} to ${FINAL_CSS_FILE}"
mkdir -p "${FINAL_CSS_DIR}"
cp "${TAILWIND_TMP_OUTPUT_FILE}" "${FINAL_CSS_FILE}"
# Optional: remove temp file
# rm -f "${TAILWIND_TMP_OUTPUT_FILE}"
# --- End Copy Tailwind CSS --- #

# Clean up temporary build directory (optional)
# echo "Cleaning up ${BUILD_DIR}"
# rm -rf "${BUILD_DIR}"

echo "Jekyll build complete. Output in ${SITE_DEST_DIR}"

# --- Ownership --- (Handled by entrypoint/runner)

echo "Process complete."
