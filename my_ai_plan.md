# Plan: Enhance agentt Guidance Loading and Filtering

## 0. Initial Refactoring & Stabilization (Completed)

### 0.1. Summary

Before embarking on major new features, several initial improvements and fixes were implemented to stabilize the existing `agentt` CLI and align it with best practices:

*   **Standardized Logging:** Migrated from standard `log` package + custom helpers to the standard `log/slog` package. Implemented standard verbosity levels (`WARN`, `INFO`, `DEBUG`) controlled by flags (`-q`, `-v`, `-vv`). Documented in `bhv-logging-levels`.
*   **Consistent ID Handling:**
    *   Removed implicit ID generation logic (e.g., using `title` or prefixes like `bhv-`).
    *   Enforced requirement for an explicit `id` field in the frontmatter of all guidance files (`.bhv`, `.rcp`).
    *   Updated `agentt summary` to only output summaries for entities possessing an explicit `id`.
    *   Updated `agentt details` to correctly parse `--id` flags (instead of positional arguments) and use the explicit IDs provided by `summary`.
*   **Concise `details` Output:** Refactored `agentt details` to output a more focused JSON structure, removing fields redundant with the `summary` output (e.g., `description`, `tags` are no longer duplicated).
*   **CLI Help Text:** Updated `agentt llm` help text to reflect the use of non-prefixed IDs.

### 0.2. Outcome

The `agentt` CLI now has consistent ID handling, standard logging behavior, and clearer output, providing a more stable foundation for future enhancements.

## 1. Introduction

### 1.1. Goals

This plan outlines the steps to implement two key enhancements for the `agentt` guidance system:

1.  **Pluggable Guidance Storage Backend:** Refactor the guidance loading mechanism to support different sources (starting with the existing local filesystem) through a common interface.
2.  **Advanced Filtering Mechanism:** Implement enhanced filtering capabilities for `agentt` CLI commands (`summary`, `details`) based on entity metadata (initially focusing on tags).

### 1.2. Guiding Behaviors

The following existing behaviors are particularly relevant to this work:

*   **`id: separation-of-concerns` (MUST):** Critical for Phase 1, ensuring loading logic is separated from storage specifics.
*   **`id: clarity-over-churn` (MUST):** Changes should prioritize improving clarity, usability, and architecture. New interfaces/flags must be clear.
*   **`id: safety-first` (MUST):** Refactoring must not compromise existing functionality. Robust testing required.
*   **`id: shell-safety` (MUST):** Any new CLI flags introduced must be designed for non-interactive use.
*   **`id: logging-levels` (SHOULD):** The standardized logging levels should be maintained and utilized appropriately.
*   **`id: ai-interaction` (SHOULD):** Consider how AI agents will utilize new capabilities. Interfaces should be clear for programmatic use.

## 2. Phase 1: Pluggable Guidance Storage Backend

### 2.1. Goals

*   Define a Go `interface` abstracting the retrieval of guidance entities (summaries and details).
*   Refactor `agentt`'s core logic (CLI commands, potentially background service) to rely on this interface.
*   Implement an initial concrete backend replicating the current local filesystem scanning logic.
*   Update configuration handling to allow specifying which backend to use.

### 2.2. Detailed Steps

1.  **Define `GuidanceBackend` Interface (Go):**
    *   Location: `internal/guidance/backend` (or similar).
    *   Methods (Initial):
        *   `GetSummary() ([]Summary, error)`: (`Summary` struct: `ID`, `Type`, `Tier`, `Tags`, `Description`).
        *   `GetDetails(ids []string) ([]Entity, error)`: (`Entity` struct: `ID`, `Type`, `Tier`, full `Body`, generalized `ResourceLocator`, `Metadata` map, `LastUpdated`, etc.).
        *   `Initialize(config map[string]interface{}) error`.
    *   **Note:** As part of defining the `Entity` structure returned by `GetDetails`, the existing `SourcePath` field (currently tied to files) **MUST** be revisited and generalized (e.g., renamed to `ResourceLocator` or `OriginURI`, potentially with a different type) to accommodate non-filesystem backends.

2.  **Refactor Core Logic:**
    *   Identify logic in `cmd/` package using the `Store` directly for loading.
    *   Modify to instantiate and use the `GuidanceBackend` interface.
    *   Ensure error handling.
    *   *Adherence:* `id: separation-of-concerns`.

3.  **Implement `LocalFilesystemBackend`:**
    *   Create struct implementing `GuidanceBackend` in `internal/guidance/backend/localfs`.
    *   Port existing `Store` loading/reading logic (currently partly in `discovery` and `store` packages) into this backend.
    *   `Initialize` takes root dir, globs, etc.
    *   *Adherence:* `id: safety-first`.

4.  **Update Configuration:**
    *   Modify `config.yaml` structure (e.g., under `guidance.backend`).
    *   Add `type` (e.g., `localfs`) and backend-specific `settings`.
    *   Update config loading logic (factory pattern?).
    *   Define default behavior (use `localfs`).

### 2.3. Test Plan

1.  **Unit Tests:** `LocalFilesystemBackend` isolation (mock fs?), `GetSummary`, `GetDetails`, `Initialize`, error handling.
2.  **Integration Tests:** `agentt summary`/`details` with default/explicit `localfs`, different dirs, non-existent IDs, config loading.
3.  **Regression Tests:** Ensure previous functionality remains.

## 3. Phase 2: Advanced Filtering Mechanism

### 3.1. Goals

*   Enhance `agentt summary` and `agentt details` with a CLI flag for filtering entities based on metadata using a query language.
*   Support flexible filtering logic including boolean operations and tag matching.
*   Apply filtering *after* retrieving entities from the backend.

### 3.2. Detailed Steps

1.  **Define Filtering Logic & CLI Flag:**
    *   Implement a single, powerful filtering flag:
        *   `--filter '<query_string>'`: Filter entities based on a metadata query string.
    *   Define the `<query_string>` syntax. Initial proposal:
        *   Terms: Match against tags (e.g., `tag:core`, `tag:git`, `priority:100`, `type:behavior`, `tier:must`). Assume case-insensitive matching for tag values unless specified otherwise.
        *   Boolean Operators: Support `AND` (implicit default), `OR`, `NOT` (or `-` prefix).
        *   Grouping: Support parentheses `()` for controlling operator precedence.
    *   *Consideration:* This query language provides a flexible foundation. Features like key existence checks (`key:*`), value wildcards, or filtering on other metadata can be added later if needed. Standard boolean precedence (NOT > AND > OR) should apply, respecting parentheses.

2.  **Implement Filtering Logic:**
    *   Create function(s) (e.g., in `internal/guidance/filtering`) that take a slice of `Summary` or `Entity` objects and the filter query string.
    *   Implement a parser for the defined query string syntax (handling terms, operators, parentheses).
    *   Implement the metadata matching logic based on the parsed query tree.
    *   Ensure the logic correctly handles combinations and precedence.

3.  **Integrate into CLI Commands:**
    *   Modify `cmd/summary.go` and `cmd/details.go`:
        *   Add the `--filter` CLI flag.
        *   Parse the flag value.
        *   After retrieving summaries/details via `GuidanceBackend`, pass results and the query string to the filtering logic.
        *   Use filtered slice for output.
    *   *Adherence:* Maintain separation.

### 3.3. Test Plan

1.  **Unit Tests:** Filtering logic and query parser isolation.
    *   Test parsing of various query strings (terms, operators, parentheses).
    *   Test matching logic with mock entities and different metadata/tag combinations.
    *   Test boolean operators and precedence rules.
    *   Test edge cases (empty query, non-existent keys/tags, entities with no metadata, malformed queries), case sensitivity.

2.  **Integration Tests:** Use test files with known tags/metadata.
    *   Test `agentt summary --filter '<query>'` and `agentt details --filter '<query>'` with various valid query combinations (simple terms, AND, OR, NOT, parentheses).
    *   Verify output contains only matching entities.
    *   Test with malformed queries (expect errors).

3.  **Usability Tests (Manual):** Check help output (`--help`), ensure `--filter` flag is documented clearly with query syntax examples.

## 4. Documentation Updates

*   Update `README.md` and command help (`--help`) to reflect:
    *   New backend configuration options (Phase 1).
    *   The new `--filter` flag and its query language syntax (Phase 2).
*   Update relevant guidance (e.g., behaviors, recipes) if interfaces or commands change significantly.

## 5. Configuration Refactoring (Post Phase 1)

### 5.1. Goals

*   Improve configuration clarity and extensibility by separating logical entity type definitions from backend-specific location details.
*   Ensure the configuration structure logically supports multiple backend types.
*   Solidify the requirement for a unique `id` for all entity types.

### 5.2. Rationale

The previous configuration mixed logical `entityType` definitions (like `requiredFrontMatter`) with location-specific details (`pathGlob`, `fileExtensionHint`) needed only by the `localfs` backend. This refactoring addresses that by moving location details into the backend's specific configuration section.

### 5.3. Revised Configuration Structure

```yaml
# config.yaml (Conceptual Structure)

entityTypes:
  - name: "behavior"
    description: "..."
    requiredFields: # Renamed from requiredFrontMatter
      - "id"          # Now explicitly required for all types
      - "title"
      - "priority"
      - "description"
      - "tags"

  - name: "recipe"
    description: "..."
    requiredFields:
      - "id"
      - "title"
      - "priority"
      - "description"
      - "tags"

# ... other logical entity types ...

backend:
  type: localfs
  rootDir: "." # Base directory for globs
  requireExplicitID: true # Example backend-specific option
  entityLocations: # Location details specific to localfs backend
    behavior: # Key matches entityType name
      pathGlob: ".agent/behavior/**/*.bhv" # Path relative to rootDir
      # fileExtensionHint removed as redundant with pathGlob
    recipe:   # Key matches entityType name
      pathGlob: ".agent/cookbook/**/*.rcp"
    # ... location info for other types if used by localfs ...
```

### 5.4. Implementation Steps

1.  **Modify `internal/config/config.go`:**
    *   Rename `requiredFrontMatter` -> `requiredFields` in `EntityTypeDefinition`.
    *   Ensure `id` is always included in `requiredFields` for all defined types.
    *   Remove `PathGlob` and `FileExtensionHint` fields from `EntityTypeDefinition`.
    *   Introduce structs for backend config parsing (`EntityLocationConfig` {`PathGlob`}, `LocalFsBackendConfig` {`RootDir`, `RequireExplicitID`, `EntityLocations map[string]EntityLocationConfig`}).
    *   Update config loading logic to handle the nested backend structure.
2.  **Modify `internal/guidance/backend/localfs/localfs.go`:**
    *   Update `Initialize` / `extractConfig` to accept and parse the new `LocalFsBackendConfig`.
    *   Use `entityLocations[typeName].PathGlob` for file scanning.
    *   Fetch `requiredFields` from the top-level `EntityTypeDefinition` (passed into `Initialize` or retrieved via a config accessor) for validation in `parseAndValidateFile`.
    *   Respect the `requireExplicitID` flag when handling missing `id` in frontmatter.
3.  **Modify `internal/guidance/backend/localfs/localfs_test.go`:**
    *   Update `getDefaultEntityDefs` helper (remove path/hint, ensure `id` required).
    *   Update `configMap` in tests to match the new nested backend structure.
    *   Ensure test entity files include an `id` field.
4.  **Update `.agent/service/config.yaml`:** Apply the structural changes, add `id` requirement, remove hints, add `entityLocations` map and `requireExplicitID` flag.

### 5.5. Outcome

A cleaner, more extensible configuration system where logical entity definitions are decoupled from backend storage specifics, and the requirement for entity IDs is explicit.