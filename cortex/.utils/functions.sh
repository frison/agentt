#!/usr/bin/env bash

find_dependencies() {
  local tag_path_root=$1
  local dir=$2
  local depth=${3:-0}

  # Don't recurse more than 1 level
  if [[ $depth -gt 1 ]]; then
    return
  fi

  # Don't include the filename in the output
  # Only print the matching parts of the line
  # Use extended regexes
  # Search for literal '${TAG_PATH_ROOT}/<dependency>:local' pattern in files
  # We no longer need escaped_tag_path_root for this specific grep
  for my_dependency in $(grep -h -o -E '\$\{TAG_PATH_ROOT\}/(.*):local' "$dir"/* 2>/dev/null | grep -v '/$' | cut -d/ -f2 | cut -d: -f1 | sort -u); do
    echo "$my_dependency"
    # Recurse into the dependency up to our maximum depth, passing tag_path_root
    find_dependencies "$tag_path_root" "$my_dependency" $((depth + 1))
  done
}

published_languages() {
  find . -maxdepth 1 -type d | sed 's|./||' | grep -vxFf .no-publish
}

log() {
  echo "[$(date)] $1"
}

ci_error() {
  >&2 echo "::error::$@"
  exit 1
}

ci_pass() {
  echo "✅ $@"
}