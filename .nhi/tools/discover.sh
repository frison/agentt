#!/usr/bin/env bash

# 🛸 NHI Artifact Discovery Protocol 🛸
# Scans specified sectors (.nhi directories) for Need/Practice artifacts (.nhi files)
# Parses YAML frontmatter and transmits findings as structured JSON data.
#
# Based on the principles encoded in .nhi/practices/directive-format.nhi

# --- Configuration & Safety ---
set -euo pipefail

# --- Argument Parsing & Validation ---
SEARCH_DIR="${1:-.nhi}"   # Default scan target: .nhi directory
OUTPUT_FORMAT="${2:-json}" # Default transmission format: json
# Note: 'table' format is deprecated, script now primarily outputs JSON.
# FILTER_TYPE="${3:-all}" # Filtering by type (.nhn/.nhp) is deprecated, use 'tier' from frontmatter.

if [ ! -d "$SEARCH_DIR" ]; then
  echo "❌ Error: Search directory '$SEARCH_DIR' not found." >&2
  exit 1
fi

# --- Core Logic Functions ---

# Finds all .nhi artifacts within the designated search directory.
find_artifacts() {
  find "$SEARCH_DIR" -name "*.nhi" -type f | sort
}

# Extracts the YAML frontmatter block (between the first pair of '---')
# Args: $1 - file path
# Output: YAML frontmatter content
extract_frontmatter() {
  local file="$1"
  # Use awk for robust block extraction between the first pair of '---'
  awk 'BEGIN{p=0} /^---$/{if(p==0){p=1;next} if(p==1){p=2;exit}} p==1{print}' "$file"
}

# Extracts a specific value from YAML frontmatter using basic grep/sed.
# Handles simple key: value lines and basic tags: [...] arrays.
# Args: $1 - frontmatter content, $2 - key name
# Output: Extracted value (cleaned)
get_yaml_value() {
  local frontmatter="$1"
  local key="$2"
  local value

  # Use grep to find the line, then sed to extract the value
  # Handles strings in quotes, numbers, booleans, and simple bracketed arrays
  value=$(echo "$frontmatter" | grep -E "^${key}:" | sed -E \
    -e "s/^${key}: +\"([^\"]*)\"$/\1/" \
    -e "s/^${key}: +([^ ]*)$/\1/" \
    -e "s/^${key}: +\[(.*)\]$/\1/")

  # Basic cleanup: remove potential leading/trailing whitespace (sed leaves it)
  value=$(echo "$value" | sed -e 's/^[[:space:]]*//' -e 's/[[:space:]]*$//')
  echo "$value"
}

# Escapes a string for safe inclusion in a JSON string value.
# Args: $1 - string to escape
# Output: JSON-escaped string
escape_json_string() {
  # Replace backslash, quote, newline, carriage return, tab
  echo "$1" | sed -e 's/\\/\\\\/g' -e 's/"/\\"/g' -e 's/\n/\\n/g' -e 's/\r/\\r/g' -e 's/\t/\\t/g'
}

# Formats the list of tags into a JSON array string.
# Args: $1 - comma-separated tag string (from get_yaml_value)
# Output: JSON array string like ["tag1", "tag2"]
format_tags_json() {
  local tags_str="$1"
  if [ -z "$tags_str" ]; then
    echo "[]"
    return
  fi
  echo "$tags_str" | sed -e 's/"//g' -e "s/,[[:space:]]*/","/g" -e "s/.*/[&]/" -e "s/\([^,\"[]\+\)/\"\1\"/g"
}

# --- Output Formatting Functions ---

# Outputs discovered artifacts as a JSON array.
format_as_json() {
  local first=true
  echo "[" # Start JSON array

  while read -r artifact_path; do
    local frontmatter
    frontmatter=$(extract_frontmatter "$artifact_path")

    # Skip if frontmatter is empty
    if [ -z "$frontmatter" ]; then
        # echo "# Warning: Skipping $artifact_path - empty frontmatter." >&2
        continue
    fi

    # Extract metadata using the YAML helper
    local title priority tier description tags
    title=$(get_yaml_value "$frontmatter" "title")
    priority=$(get_yaml_value "$frontmatter" "priority")
    tier=$(get_yaml_value "$frontmatter" "tier")
    description=$(get_yaml_value "$frontmatter" "description")
    tags=$(get_yaml_value "$frontmatter" "tags")

    # Basic validation - ensure essential fields are present
    if [ -z "$title" ] || [ -z "$priority" ] || [ -z "$tier" ]; then
      echo "# Warning: Skipping $artifact_path - missing required frontmatter (title, priority, tier)." >&2
      continue
    fi

    # Escape strings for JSON
    local escaped_title escaped_description escaped_path escaped_tier
    escaped_title=$(escape_json_string "$title")
    escaped_description=$(escape_json_string "$description")
    escaped_path=$(escape_json_string "$artifact_path")
    escaped_tier=$(escape_json_string "$tier")
    local formatted_tags_json=$(format_tags_json "$tags")

    # Add comma separator if not the first element
    if [ "$first" = true ]; then
      first=false
    else
      echo "," # Comma precedes the next object
    fi

    # Construct and print JSON object for the artifact
    cat <<-JSON_EOF
  {
    "path": "${escaped_path}",
    "tier": "${escaped_tier}",
    "title": "${escaped_title}",
    "priority": ${priority:-999},
    "description": "${escaped_description}",
    "tags": ${formatted_tags_json}
  }
JSON_EOF
  # Using cat <<- removes leading tabs, ensuring proper JSON format.
  # Default priority to 999 if somehow empty after check.

  done < <(find_artifacts)

  echo # Add a newline if any objects were printed
  echo "]" # End JSON array
}

# Deprecated table format - kept for reference or potential future use.
format_as_table() {
  echo "# Warning: Table output format is deprecated. Use JSON." >&2
  printf "%-9s | %-10s | %-30s | %s\n" "PRIORITY" "TIER" "TITLE" "PATH"
  printf "%s\n" "---------|-----------|-----------------------------|-----------------"

  while read -r artifact_path; do
     local frontmatter
     frontmatter=$(extract_frontmatter "$artifact_path")
     if [ -z "$frontmatter" ]; then continue; fi

     local title priority tier
     title=$(get_yaml_value "$frontmatter" "title")
     priority=$(get_yaml_value "$frontmatter" "priority")
     tier=$(get_yaml_value "$frontmatter" "tier")

     if [ -z "$title" ] || [ -z "$priority" ] || [ -z "$tier" ]; then continue; fi

     printf "%-9s | %-10s | %-30s | %s\n" "${priority:-999}" "${tier}" "${title:0:30}" "$artifact_path"
  done < <(find_artifacts)
}

# --- Main Execution ---

# Transmit findings based on requested format
echo "# 📡 Scanning Sector: $SEARCH_DIR for NHI artifacts..." >&2
case "$OUTPUT_FORMAT" in
  json)
    format_as_json
    ;;
  table)
    format_as_table # Keep for potential compatibility, but warn.
    ;;
  *)
    echo "❌ Error: Unknown output format '$OUTPUT_FORMAT'. Use 'json' or 'table'." >&2
    exit 1
    ;;
esac

echo "# ✅ Scan complete." >&2