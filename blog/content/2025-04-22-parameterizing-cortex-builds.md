---
layout: post
title: "Refactoring Cortex: Parameterizing Docker Image Builds"
date: 2025-04-21 22:36:43 -0600
author: "AI Assistant (Gemini)" # Placeholder - Adjust if needed
categories:
  - cortex
  - docker
  - make
  - refactoring
  - ci
  - 100hellos
---

## The Challenge: Hardcoded Image Prefixes

The `cortex` directory, responsible for building Docker images for various programming language environments (originally part of the "100 Hello Worlds" project), had a small limitation: the Docker image prefix was hardcoded as `100hellos/` throughout its Makefiles, utility scripts, and even Dockerfiles.

While this worked perfectly for its original purpose – building and publishing images under the `100hellos` namespace – it presented a challenge when integrating `cortex` into a larger project (`agentt`). We wanted the build system to work intuitively *within* the `cortex` directory itself, ideally producing images prefixed with `cortex/` by default during local development. However, we also needed to retain the ability to build images with the `100hellos/` prefix for CI/CD pipelines that publish to the original Docker Hub repository.

## The Goal: Flexibility and Consistency

Our goal was twofold:

1.  **Local Development:** Allow developers working within the `cortex` directory to run `make <language>` and get an image named `cortex/<language>:local` by default.
2.  **CI/Publishing:** Provide a mechanism to override the prefix, allowing CI jobs to build images as `100hellos/<language>:local` (or `:latest` for publishing).

## The Solution: Introducing `TAG_PATH_ROOT`

We addressed this by introducing a Make variable, `TAG_PATH_ROOT`, to control the image prefix.

1.  **Default Behavior:** In the main `cortex/Makefile`, we set `TAG_PATH_ROOT ?= $(shell basename ${PWD})`. This makes the default value the name of the current directory (`cortex` when running `make` from `/cortex`).
2.  **Override Mechanism:** Users can override this default in several ways:
    *   `make TAG_PATH_ROOT=100hellos <target>` (The explicit way)
    *   `make P=100hellos <target>` (A shorter alias)
    *   `make PREFIX=100hellos <target>` (Another alias)
3.  **Propagation:** This `TAG_PATH_ROOT` variable is now consistently passed down through:
    *   Recursive `make` calls to language subdirectories.
    *   Arguments to the `.utils/build_image.sh` script.
    *   Arguments to the `.utils/functions.sh` script (for dependency resolution).
    *   Docker build arguments (`--build-arg TAG_PATH_ROOT=...`).
4.  **Dockerfile Adaptation:** All relevant `Dockerfile`s were updated:
    *   `ARG TAG_PATH_ROOT` was added at the top.
    *   `FROM 100hellos/...:local` lines were changed to `FROM ${TAG_PATH_ROOT}/...:local`.
5.  **Dependency Fix:** The dependency resolution logic in `.utils/functions.sh` was updated to `grep` for the literal string `${TAG_PATH_ROOT}/` in Dockerfiles, ensuring dependencies are still detected correctly after the parameterization.

## The Outcome

This refactoring provides the desired flexibility:

*   Running `make java` inside `cortex` now correctly builds `cortex/100-java11:local` and then `cortex/java:local`.
*   Running `make P=100hellos java` builds `100hellos/100-java11:local` and `100hellos/java:local`, preserving the original behavior for CI/publishing workflows.

This change allows the `cortex` build system to function naturally within its local context while retaining compatibility with its origins and deployment requirements.

---
