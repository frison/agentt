#!/usr/bin/env bash

# 🍳 Cookbook Recipe Discovery Protocol 🍳
# Scans specified cookbook sectors (.agent/cookbook directories) for recipes (.rcp files)
# Extracts YAML frontmatter with awk, parses with yq, and transmits findings as structured JSON data.

# --- Configuration & Safety ---
set -euo pipefail

# --- Dependencies ---
# Check if yq is available
if ! command -v yq &> /dev/null; then
    echo "❌ Error: yq (YAML processor) is required but not found. Please install yq." >&2
    exit 1
fi
# Check if awk is available (standard tool, but good practice)
if ! command -v awk &> /dev/null; then
    echo "❌ Error: awk is required but not found." >&2
    exit 1
fi


# --- Argument Parsing & Validation ---
SEARCH_DIR="${1:-.agent/cookbook}" # Default scan target: .agent/cookbook directory
OUTPUT_FORMAT="${2:-json}"        # Default transmission format: json (only json supported)

if [ ! -d "$SEARCH_DIR" ]; then
  echo "❌ Error: Search directory '$SEARCH_DIR' not found." >&2
  exit 1
fi

# --- Core Logic Functions ---

# Finds all .rcp recipe files within the designated search directory.
find_artifacts() {
  find "$SEARCH_DIR" -name "*.rcp" -type f | sort
}

# Extracts the YAML frontmatter block (between the first pair of '---')
# Args: $1 - file path
# Output: YAML frontmatter content
extract_frontmatter() {
  local file="$1"
  # Use awk for robust block extraction between the first pair of '---'
  awk 'BEGIN{p=0} /^---$/{if(p==0){p=1;next} if(p==1){p=2;exit}} p==1{print}' "$file"
}

# Escapes a string for safe inclusion in a JSON string value.
escape_json_string() {
  # Use jq for robust JSON escaping if available, otherwise basic sed
  if command -v jq &> /dev/null; then
      # Read raw input (-R), slurp all input (-s), output as JSON string (-a)
      jq -Rsa <<< "$1"
  else
      # Basic sed escaping (less robust for complex cases)
      echo "$1" | sed -e 's/\\/\\\\/g' -e 's/"/\\"/g' -e 's/\n/\\n/g' -e 's/\r/\\r/g' -e 's/\t/\\t/g'
  fi
}

# --- Output Formatting Functions ---

# Outputs discovered recipes as a JSON array using awk + yq.
format_as_json() {
  local first=true
  echo "[" # Start JSON array

  while IFS= read -r artifact_path; do
    # Extract frontmatter using awk
    local frontmatter
    frontmatter=$(extract_frontmatter "$artifact_path")

    # Skip if extracted frontmatter is empty
    if [ -z "$frontmatter" ]; then
        # echo "# Warning: Skipping $artifact_path - No frontmatter found." >&2
        continue
    fi

    # Attempt to convert the extracted frontmatter YAML to compact JSON using yq
    local frontmatter_json
    frontmatter_json=$(echo "$frontmatter" | yq -c '.' 2>/dev/null || echo "null")

    # Skip if yq failed or returned null/empty object
    if [ "$frontmatter_json" == "null" ] || [ -z "$frontmatter_json" ] || [ "$frontmatter_json" == "{}" ]; then
      echo "# Warning: Skipping $artifact_path - Could not parse extracted frontmatter with yq." >&2
      continue
    fi

    # Basic validation - check for required fields
    local missing_fields=false
    # Use jq for reliable validation if available
    if command -v jq &> /dev/null; then
        # Check for presence of required keys and that tags is a non-empty array
        if ! echo "$frontmatter_json" | jq -e 'has("id") and has("title") and has("priority") and has("description") and (.tags | type == "array" and length > 0)' > /dev/null; then
            missing_fields=true
        fi
    else
        # Basic check if jq not available (less reliable)
        if ! echo "$frontmatter_json" | grep -q '"id":' || \
           ! echo "$frontmatter_json" | grep -q '"title":' || \
           ! echo "$frontmatter_json" | grep -q '"priority":' || \
           ! echo "$frontmatter_json" | grep -q '"description":' || \
           ! echo "$frontmatter_json" | grep -q '"tags":\[.*\]'; then # Basic check for tags array, might miss empty array
             missing_fields=true
        fi
    fi

    if [ "$missing_fields" = true ]; then
      echo "# Warning: Skipping $artifact_path - Missing required frontmatter fields (id, title, priority, description, tags). Parsed: $frontmatter_json" >&2
      continue
    fi

    # Add comma separator if not the first element
    if [ "$first" = true ]; then first=false; else echo ","; fi

    # Combine frontmatter JSON with the artifact path
    escaped_path=$(escape_json_string "$artifact_path")
    # Use jq to merge if available for robustness
    if command -v jq &> /dev/null; then
        # Add path to the existing JSON object
        echo "$frontmatter_json" | jq --arg path "$artifact_path" '. + {path: $path}'
    else
        # Less robust merging without jq - prepend path field
        # Remove leading { and add path field, then add { back
        processed_json=$(echo "$frontmatter_json" | sed 's/^{\(.*\)}$/\1/')
        printf '{"path": %s, %s}' "$escaped_path" "$processed_json"
    fi

done < <(find_artifacts)

  echo # Add a newline if any objects were printed
  echo "]" # End JSON array
}

# --- Main Execution ---

# Transmit findings based on requested format
echo "# 🍳 Scanning Sector: $SEARCH_DIR for cookbook recipes (using awk+yq)..." >&2
case "$OUTPUT_FORMAT" in
  json)
    format_as_json
    ;;
  table)
     echo "# Error: Table output format is not supported." >&2
     exit 1
     ;;
  *)
    echo "❌ Error: Unknown output format '$OUTPUT_FORMAT'. Use 'json'." >&2
    exit 1
    ;;
esac

echo "# ✅ Scan complete." >&2