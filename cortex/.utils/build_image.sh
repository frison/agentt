#!/usr/bin/env bash
# Usage:
#   build_image.sh <tag_path_root> <language> [docker_build_args...]

# Example:
#   build_image.sh myproject visual-basic --platform=linux/amd64

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "$SCRIPT_DIR/functions.sh"
cd "$SCRIPT_DIR/.."

build_image() {
  local tag_path_root=$1
  local language=$2
  shift 2 # Shift past tag_path_root and language
  local build_args=$*
  local dockerfile="$language/Dockerfile"
  local image_name="$tag_path_root/$language:local"
  echo "Building image: $image_name with args: $build_args"
  docker build --build-arg TAG_PATH_ROOT="$tag_path_root" $build_args -t "$image_name" -f "$dockerfile" "$language"
}

tag_path_root=$1
language=$2
shift 2 # Shift past tag_path_root and language
build_args=$*

echo "Building dependencies for: $language ($tag_path_root) with docker args \"$build_args\""
# Pass tag_path_root to recursive calls
for dependency in $(find_dependencies "$tag_path_root" "$language" | sort -u); do
  echo "Building dependency: $dependency"
  build_image "$tag_path_root" "$dependency" $build_args
done

# Build the main image
build_image "$tag_path_root" "$language" $build_args
