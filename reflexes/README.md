# Reflexes

Self-contained, deterministic units of functionality in Docker containers.

## Core Principles

- **Build-time Freedom**: Reflexes can download, install, or compile dependencies during build
- **Runtime Purity**: Once built, reflexes must:
  - Be completely deterministic
  - Have no external dependencies
  - Operate only on their inputs
  - Be idempotent
  - Run entirely locally

## Base Images (`100hellos/*`)

All reflexes are built on top of the language-specific `100hellos` base images (e.g., `100hellos/python:latest`, `100hellos/golang:latest`), which are maintained and published from the `cortex` directory. These images provide:
- Standardized, secure base environments tailored for specific languages/runtimes.
- Common utilities and libraries.
- Consistent versioning and updates.
- Built-in best practices for containerization.
- Clear provenance and history.

Always use the most appropriate `100hellos` base image for your reflex's language/runtime needs.

## Common Toolchain (`.base-tools`)

To ensure consistency and reduce duplication, a common set of essential NHI reflex tools (static binaries, required scripts, etc.) is provided in a separate utility image named `.base-tools`.

- **Purpose:** Acts as a source for common tools needed by reflexes.
- **Usage:** Tools from `.base-tools` are copied into the reflex's build context, typically using a multi-stage Docker build.
- **Benefit:** Keeps language base images focused, avoids tool duplication across reflexes, and provides a single point of update for common utilities.

**Mechanism:**
Tools are copied from the `.base-tools` image into the final image stage, which is based on a `100hellos` language-specific image.

```dockerfile
# Example: Using .base-tools in a Python reflex build
FROM .base-tools AS tools

FROM 100hellos/python:latest

USER nhi
WORKDIR /app

# Copy the entire toolset from .base-tools overlaying the root filesystem
COPY --from=tools / /

# Copy reflex application code
COPY files/ .

# Default ENTRYPOINT uses the nhi-entrypoint-helper to provide
# usage instructions based on manifest.yml
ENTRYPOINT ["/usr/local/bin/nhi-entrypoint-helper", "python", "main.py"]
```

## Container Structure

### Directory Layout
Every reflex should follow this structure:
```
reflex-name/
├── files/            # All source files
│   └── main.*        # Main entry point (language specific)
├── Dockerfile        # Container definition
├── manifest.yml      # Input/output specification
└── README.md         # Documentation
```

### Container Patterns

Reflexes typically use a language-specific `100hellos` base image. Multi-stage builds are recommended, especially for compiled languages, to keep images small and incorporate tools from `.base-tools`.

**Example (Compiled Language - e.g., Go):**
```dockerfile
# Tool staging (using .base-tools)
FROM .base-tools AS tools

# Build stage (using a standard Go builder image)
FROM 100hellos/golang:latest AS builder
# Note: Using the 100hellos version ensures consistent Go environment
WORKDIR /build
COPY files/ .
# Copy any tools needed *during* build from .base-tools if necessary
# Example: COPY --from=tools /path/to/build_tool /usr/local/bin/
# Or, if the build tool is included in the standard overlay:
# COPY --from=tools / /
RUN CGO_ENABLED=0 go build -o reflex_app

# Final runtime stage (using a minimal 100hellos base or appropriate language base)
FROM 100hellos/base:latest # Or a more specific minimal base if available
USER nhi
WORKDIR /app

# Copy the compiled application from the builder stage
COPY --from=builder /build/reflex_app .

# Copy the entire runtime toolset from .base-tools overlaying the root filesystem
COPY --from=tools / /

# Set entrypoint using the standard helper
# The helper will execute /app/reflex_app if args are valid
ENTRYPOINT ["/usr/local/bin/nhi-entrypoint-helper", "/app/reflex_app"]
```

**Example (Interpreted Language - e.g., Python):**
```dockerfile
# Tool staging (using .base-tools)
FROM .base-tools AS tools

# Final runtime stage (using the appropriate 100hellos language image)
FROM 100hellos/python:latest
USER nhi
WORKDIR /app

# Copy application code
COPY files/ .

# Copy the entire runtime toolset from .base-tools overlaying the root filesystem
COPY --from=tools / /

# Set entrypoint using the standard helper
# The helper will execute python main.py if args are valid
ENTRYPOINT ["/usr/local/bin/nhi-entrypoint-helper", "python", "main.py"]
```
*Note: The `100hellos` base images should already be configured to run as the `nhi` user.*

### Manifest Format
The `manifest.yml` should be formatted for both NHI and human consumption:

```yaml
# Human-readable section
name: example-reflex
version: "1.0"
description: |
  A clear description of what this reflex does.
  Can span multiple lines for clarity.

# NHI-compatible specification
inputs:
  environment:
    EXAMPLE_VAR:
      type: string
      description: "NHI-parseable description"
      required: true

outputs:
  stdout:
    type: json
    schema:
      $schema: "http://json-schema.org/draft-07/schema#"
      type: object
      properties:
        result:
          type: string
          description: "NHI-parseable description"
```

### Best Practices
1. Source Organization:
   - All source files in `files/` directory
   - Clear main entry point
   - Flat file structure when possible

2. Build Process:
   - Multi-stage builds for compiled languages
   - Minimize layers
   - Leverage build caching
   - Document build requirements

3. Runtime:
   - Use appropriate `100hellos` base image.
   - Include only necessary artifacts (app code, copied tools).
   - Run as 'nhi' user (provided by base image).
   - Use the standard `nhi-entrypoint-helper` for the `ENTRYPOINT` unless there's a strong reason otherwise.
   - Handle errors gracefully

## Directory Structure

```
reflexes/
├── transform/      # Data transformation reflexes
│   ├── format/     # Format conversions (json->yaml, markdown->html, etc)
│   ├── text/       # Text processing operations
│   └── struct/     # Data structure transformations
├── compute/        # Computation-focused reflexes
│   ├── math/       # Mathematical operations
│   ├── stats/      # Statistical computations
│   └── graph/      # Graph processing
├── validate/       # Validation reflexes
│   ├── schema/     # Schema validation
│   ├── syntax/     # Syntax checking
│   └── lint/       # Linting operations
├── analyze/        # Analysis reflexes
│   ├── metrics/    # Metric computation
│   ├── extract/    # Information extraction
│   └── classify/   # Classification operations
├── template/       # Template and example reflexes
└── test/          # Test reflexes and validation suite
```

## Guidelines

1. Each reflex must be:
   - Self-contained (one directory with all needed files)
   - Deterministic (same input = same output)
   - Idempotent (multiple runs = same result)
   - Local-only during runtime
   - Based on an appropriate `100hellos` image, incorporating tools from `.base-tools` as needed.

2. Build phase can:
   - Download dependencies (in builder stages)
   - Pull in models or data
   - Compile code
   - Prepare resources
   - Run as non-root 'nhi' user (when feasible with the chosen minimal base)

3. Runtime phase must:
   - Use only local resources
   - Operate only on provided inputs
   - Produce consistent outputs
   - Not require network access
   - Run as 'nhi' user (provided by base image)

4. Documentation should specify:
   - Build-time requirements (including any builder images)
   - Runtime guarantees
   - Input/output specifications
   - Resource requirements
   - Base `100hellos` image used and version
   - Confirmation that tools are overlaid from `.base-tools` using `COPY --from=tools / /`.
   - Verification that the `nhi-entrypoint-helper` displays correct usage based on `manifest.yml`.

5. Testing should verify:
   - Deterministic behavior
   - Idempotency
   - No runtime external dependencies
   - Correct ownership of file outputs (should be `nhi` user)
   - Compatibility with the base `100hellos` image version
   - Correct function with overlaid tools from `.base-tools`.