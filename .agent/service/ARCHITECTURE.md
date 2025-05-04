# agentt Service Architecture

This document provides a high-level overview of the Go service located in the `.agent/service` directory.

## Directory Structure

```
.agent/service/
├── Makefile          # Build automation script.
├── README.md         # Service-specific README.
├── agentt            # Compiled binary (output of make).
├── cmd/              # Command definitions and main entry points.
│   ├── cli/          # Implementation for CLI subcommands (e.g., summary, details).
│   ├── server/       # Implementation for the `server` command.
│   ├── *.go          # Cobra command definitions.
│   └── *.txt         # Embedded help text files.
├── config.yaml       # Default configuration file.
├── go.mod            # Go module definition (module name: agentt).
├── go.sum            # Go module checksums.
├── internal/         # Core application logic (not intended for external import).
│   ├── config/       # Configuration loading and validation.
│   ├── content/      # Data structures for guidance items (Item, ItemSummary) and ID generation.
│   ├── discovery/    # File discovery, parsing, and watching logic.
│   ├── server/       # HTTP server implementation (handlers, middleware).
│   └── store/        # In-memory storage for guidance items.
└── main.go           # Main application entry point, likely executes Cobra root command.
```

## Key Components

*   **Configuration (`internal/config`, `config.yaml`):** Defines entity types and server settings. Loaded using standard precedence.
*   **Discovery (`internal/discovery`):** Scans the filesystem based on `config.yaml` globs, parses `.bhv` and `.rcp` files using `internal/content` definitions.
*   **Content Model (`internal/content`):** Defines the `Item` and `ItemSummary` structs. Contains logic for generating canonical entity IDs (`GetItemID`).
*   **Storage (`internal/store`):** Provides an in-memory `GuidanceStore` to hold parsed `Item` objects, indexed by `SourcePath` and canonical ID.
*   **Commands (`cmd/`, `main.go`):** Uses the Cobra library to define CLI commands (`server`, `summary`, `details`, `llm`). `main.go` executes the root command.
*   **Server (`internal/server`, `cmd/server.go`):** Implements the HTTP API endpoints (`/summary`, `/details`, etc.) using the `GuidanceStore`.
*   **Build (`Makefile`):** Handles formatting, linting (optional), and building the `agentt` binary.

## Module Path

The Go module name is `agentt` as defined in `go.mod`. Internal packages should be imported relative to this, e.g., `agentt/internal/content`.

### Testing

Unit and integration tests are crucial for ensuring the service's correctness and stability. Tests should be co-located with the code they test (e.g., `foo_test.go` alongside `foo.go`).

Test files that require filesystem content (like mock configuration files, behaviors, or recipes) should place these assets in a `testdata` directory adjacent to the test file. Helper functions within the tests (e.g., using `t.TempDir()` and `filepath.WalkDir`) should be used to copy these assets into a temporary location for each test run, ensuring test isolation and repeatable results.