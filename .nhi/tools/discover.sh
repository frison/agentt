#!/usr/bin/env bash

# discover.sh - NHI Discovery Tool
# Finds and lists NHI directives, principles, and actions in a structured format

set -e

SEARCH_DIR="${1:-.nhi}"
OUTPUT_FORMAT="${2:-json}"
FILTER_TYPE="${3:-all}"  # all, principles, directives, actions

function find_files() {
  case "$FILTER_TYPE" in
    principles)
      find "$SEARCH_DIR" -name "*.nhp" -type f | sort
      ;;
    directives)
      find "$SEARCH_DIR" -name "*.nhd" -type f | sort
      ;;
    actions)
      find "$SEARCH_DIR" -name "*.nha" -type f | sort
      ;;
    all|*)
      find "$SEARCH_DIR" -name "*.nh[pda]" -type f | sort
      ;;
  esac
}

function get_type_from_extension() {
  local file="$1"
  case "$file" in
    *.nhp)
      echo "principle"
      ;;
    *.nhd)
      echo "directive"
      ;;
    *.nha)
      echo "action"
      ;;
    *)
      echo "unknown"
      ;;
  esac
}

function extract_frontmatter() {
  local file="$1"
  # Grab ONLY the first frontmatter section
  # (to avoid parsing example frontmatter in the documentation)
  sed -n '1,/^---$/p' "$file" | sed '1d' | head -n -1
}

function clean_json_string() {
  # Remove control characters and escape quotes
  echo "$1" | tr -d '\000-\037' | sed 's/"/\\"/g'
}

function format_as_json() {
  echo "["
  local first=true

  find_files | while read -r file; do
    local frontmatter
    frontmatter=$(extract_frontmatter "$file")

    # Get file type based on extension
    local type
    type=$(get_type_from_extension "$file")

    # Extract common properties from frontmatter
    local title
    title=$(echo "$frontmatter" | grep -E "^title:" | sed 's/^title: *"\(.*\)"$/\1/')
    title=$(clean_json_string "$title")

    local priority
    priority=$(echo "$frontmatter" | grep -E "^priority:" | sed 's/^priority: *\([0-9]*\)$/\1/')
    priority=${priority:-10}  # Default to lowest priority if not set

    # Extract type-specific properties
    local universal=""
    local disciplines=""
    local scope=""
    local binding=""
    local applies_to=""
    local guided_by=""

    case "$type" in
      principle)
        universal=$(echo "$frontmatter" | grep -E "^universal:" | sed 's/^universal: *\(true\|false\)$/\1/')
        universal=${universal:-false}
        disciplines=$(echo "$frontmatter" | grep -E "^disciplines:" | sed 's/^disciplines: *\[\(.*\)\]$/\1/')
        ;;
      directive)
        scope=$(echo "$frontmatter" | grep -E "^scope:" | sed 's/^scope: *"\(.*\)"$/\1/')
        scope=$(clean_json_string "$scope")
        binding=$(echo "$frontmatter" | grep -E "^binding:" | sed 's/^binding: *\(true\|false\)$/\1/')
        binding=${binding:-false}
        ;;
      action)
        applies_to=$(echo "$frontmatter" | grep -E "^applies_to:" | sed 's/^applies_to: *\[\(.*\)\]$/\1/')
        guided_by=$(echo "$frontmatter" | grep -E "^guided_by:" | sed 's/^guided_by: *\[\(.*\)\]$/\1/')
        ;;
    esac

    local tags
    tags=$(echo "$frontmatter" | grep -E "^tags:" | sed 's/^tags: *\[\(.*\)\]$/\1/')

    if [ "$first" = true ]; then
      first=false
    else
      echo "  ,"
    fi

    # Output JSON for this item
    echo "  {"
    echo "    \"path\": \"$file\","
    echo "    \"type\": \"$type\","
    echo "    \"title\": \"$title\","
    echo "    \"priority\": $priority,"

    # Output type-specific properties
    case "$type" in
      principle)
        echo "    \"universal\": $universal,"
        [ -n "$disciplines" ] && echo "    \"disciplines\": [$disciplines],"
        ;;
      directive)
        [ -n "$scope" ] && echo "    \"scope\": \"$scope\","
        [ -n "$binding" ] && echo "    \"binding\": $binding,"
        ;;
      action)
        [ -n "$applies_to" ] && echo "    \"applies_to\": [$applies_to],"
        [ -n "$guided_by" ] && echo "    \"guided_by\": [$guided_by],"
        ;;
    esac

    # Output tags if present - FIXED without using process substitution with sed -i
    if [ -n "$tags" ]; then
      # No trailing comma needed since this is the last field
      echo "    \"tags\": [$tags]"
    else
      # Remove the trailing comma from the previous line
      # This is a workaround since we can't easily edit the previous line
      # Just end the JSON object properly
      echo -n "    " # Add proper indentation
    fi

    echo "  }"
  done
  echo "]"
}

function format_as_table() {
  printf "%-9s | %-10s | %-30s | %-30s | %s\n" "PRIORITY" "TYPE" "TITLE" "DETAILS" "PATH"
  printf "%s\n" "---------|-----------|-----------------------------|------------------------------|-----------------"

  find_files | while read -r file; do
    local frontmatter
    frontmatter=$(extract_frontmatter "$file")

    # Get file type based on extension
    local type
    type=$(get_type_from_extension "$file")

    # Extract common properties
    local title
    title=$(echo "$frontmatter" | grep -E "^title:" | sed 's/^title: *"\(.*\)"$/\1/')

    local priority
    priority=$(echo "$frontmatter" | grep -E "^priority:" | sed 's/^priority: *\([0-9]*\)$/\1/')
    priority=${priority:-10}

    # Extract type-specific properties for details column
    local details=""

    case "$type" in
      principle)
        local universal
        universal=$(echo "$frontmatter" | grep -E "^universal:" | sed 's/^universal: *\(true\|false\)$/\1/')
        local disciplines
        disciplines=$(echo "$frontmatter" | grep -E "^disciplines:" | sed 's/^disciplines: *\[\(.*\)\]$/\1/')
        details="universal: $universal, disciplines: [$disciplines]"
        ;;
      directive)
        local scope
        scope=$(echo "$frontmatter" | grep -E "^scope:" | sed 's/^scope: *"\(.*\)"$/\1/')
        local binding
        binding=$(echo "$frontmatter" | grep -E "^binding:" | sed 's/^binding: *\(true\|false\)$/\1/')
        details="scope: $scope, binding: $binding"
        ;;
      action)
        local applies_count
        applies_count=$(echo "$frontmatter" | grep -E "^applies_to:" | wc -l)
        details="applies to $applies_count patterns"
        ;;
    esac

    printf "%-9s | %-10s | %-30s | %-30s | %s\n" "$priority" "$type" "${title:0:30}" "${details:0:30}" "$file"
  done
}

case "$OUTPUT_FORMAT" in
  json)
    format_as_json
    ;;
  table)
    format_as_table
    ;;
  *)
    echo "Unknown output format: $OUTPUT_FORMAT"
    echo "Available formats: json, table"
    exit 1
    ;;
esac