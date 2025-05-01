# Plan: Enhance agentt Guidance Loading and Filtering

## 1. Introduction

### 1.1. Goals

This plan outlines the steps to implement two key enhancements for the `agentt` guidance system:

1.  **Pluggable Guidance Storage Backend:** Refactor the guidance loading mechanism to support different sources (starting with the existing local filesystem) through a common interface.
2.  **Advanced Tagging / Filtering Mechanism:** Implement enhanced filtering capabilities for `agentt` CLI commands (`summary`, `details`) based on entity tags.

### 1.2. Guiding Behaviors

The following existing behaviors are particularly relevant to this work:

*   **`bhv-Separation of Concerns` (MUST):** This is critical for Phase 1, ensuring the guidance loading logic is cleanly separated from the specifics of how/where guidance is stored.
*   **`bhv-Clarity Over Churn` (MUST):** Both phases involve refactoring or adding features. Changes should prioritize improving the clarity, usability, and architecture of the system, even if it requires modifying existing code. The new filtering CLI flags must be clear and unambiguous.
*   **`bhv-Safety First` (MUST):** Refactoring must not compromise the existing functionality or integrity of the guidance system. Robust testing is required. New features must handle errors gracefully.
*   **`bhv-Non-Interactive Command Safety` (MUST):** Any new CLI flags introduced for filtering must be designed for non-interactive use.
*   **`bhv-AI Interaction Guidelines Practice` (SHOULD):** Consider how AI agents will utilize the new filtering capabilities. The interface (CLI flags) should be clear and predictable for programmatic use.

## 2. Phase 1: Pluggable Guidance Storage Backend

### 2.1. Goals

*   Define a Go `interface` abstracting the retrieval of guidance entities (summaries and details).
*   Refactor `agentt`'s core logic (CLI commands, potentially background service if applicable) to rely on this interface instead of directly accessing the filesystem.
*   Implement an initial concrete backend that replicates the current local filesystem scanning logic.
*   Update configuration handling to allow specifying which backend to use (defaulting to the local filesystem).

### 2.2. Detailed Steps

1.  **Define `GuidanceBackend` Interface (Go):**
    *   Location: Define within an appropriate package, e.g., `internal/guidance/backend`.
    *   Methods (Initial):
        *   `GetSummary() ([]Summary, error)`: Returns summaries for all entities managed by the backend. (`Summary` struct likely needs `ID`, `Type`, `Tier`, `Tags`, `Description`).
        *   `GetDetails(ids []string) ([]Entity, error)`: Returns full details for specified entity IDs. (`Entity` struct needs full frontmatter, body, source path, etc.).
        *   `Initialize(config map[string]interface{}) error`: Method to pass backend-specific configuration.

2.  **Refactor Core Logic:**
    *   Identify all places in the `agentt` codebase (primarily CLI command implementations like `cmd/summary.go`, `cmd/details.go`) that currently perform filesystem scanning or file reading for `.bhv` and `.rcp` files.
    *   Modify this logic to:
        *   Instantiate a `GuidanceBackend` based on configuration.
        *   Call the interface methods (`GetSummary`, `GetDetails`) to retrieve guidance data.
    *   Ensure error handling from the backend interface is propagated correctly.
    *   *Adherence:* `bhv-Separation of Concerns` is paramount here. The core command logic should know *nothing* about *how* summaries/details are retrieved, only that they conform to the interface.

3.  **Implement `LocalFilesystemBackend`:**
    *   Create a new struct implementing the `GuidanceBackend` interface.
    *   Location: e.g., `internal/guidance/backend/localfs`.
    *   Port the *existing* filesystem scanning and file reading logic into the methods of this struct (`GetSummary`, `GetDetails`).
    *   The `Initialize` method would likely take configuration specifying the root directory, globs for behaviors/recipes, etc. (similar to current config).
    *   *Adherence:* Ensure this implementation faithfully replicates the current functionality to avoid regressions (`bhv-Safety First`).

4.  **Update Configuration:**
    *   Modify the `agentt` configuration structure (e.g., `config.yaml`).
    *   Add a section to specify the guidance backend, e.g.:
        ```yaml
        guidance:
          backend:
            type: localfs # Default if omitted?
            # Backend-specific settings nested here
            settings:
              # Settings for localfs backend (paths, globs etc.)
              # Example:
              # behavior_glob: ".agent/behavior/**/*.bhv"
              # recipe_glob: ".agent/cookbook/**/*.rcp"
        ```
    *   Update configuration loading logic to parse this section and instantiate the correct backend type via a factory pattern or similar.
    *   *Consideration:* Define clear behavior if the `backend` section or `type` is missing (default to `localfs`).

### 2.3. Test Plan

1.  **Unit Tests:**
    *   Test the `LocalFilesystemBackend` implementation in isolation. Mock filesystem interactions if necessary.
    *   Verify `GetSummary` correctly scans and parses frontmatter from test files.
    *   Verify `GetDetails` correctly reads and returns full content for specified IDs.
    *   Verify `Initialize` correctly processes configuration settings.
    *   Test error handling (e.g., invalid paths, malformed files).

2.  **Integration Tests:**
    *   Test the `agentt summary` command:
        *   Run `agentt summary` with the default configuration (implicitly using `localfs`). Verify output matches the expected summaries from the test guidance files.
        *   Run `agentt summary` with configuration explicitly specifying `type: localfs`. Verify identical output.
        *   Test with configuration pointing to different directories.
    *   Test the `agentt details` command:
        *   Run `agentt details --id <id1> --id <id2>` with the default/explicit `localfs` backend. Verify the correct full entity details are returned.
        *   Test requesting non-existent IDs.
        *   Test requesting a mix of valid and invalid IDs.
    *   Test configuration loading:
        *   Verify correct backend instantiation based on config `type`.
        *   Verify backend-specific settings are passed correctly via `Initialize`.
        *   Test handling of invalid backend types or missing configuration.

3.  **Regression Tests:**
    *   Ensure all existing tests for `summary` and `details` commands still pass after the refactoring.

## 3. Phase 2: Advanced Tagging / Filtering Mechanism

### 3.1. Goals

*   Enhance `agentt summary` and `agentt details` commands with new CLI flags for filtering entities based on tags.
*   Support filtering logic like including entities with *any* of a set of tags, excluding entities with *any* of a set of tags, and requiring entities to have *all* of a set of tags.
*   Apply filtering *after* retrieving entities from the backend but *before* presenting output.

### 3.2. Detailed Steps

1.  **Finalize Filtering Logic & CLI Flags:**
    *   Define the exact CLI flags. Proposal (Initial - Simple Tags):
        *   `--include-tags <tag1,tag2,...>`: Include entities matching *any* of these tags.
        *   `--exclude-tags <tag3,tag4,...>`: Exclude entities matching *any* of these tags.
        *   `--require-all-tags <tag5,tag6,...>`: Only include entities matching *all* of these tags.
    *   *Consideration:* Evaluate using key:value tags (e.g., `key:value`) for potentially more precise filtering, which could optimize AI interactions (token count, relevance). This adds complexity vs. simple tags. Final design should balance filtering effectiveness (especially for AI consumption) and implementation effort. CLI flags might adapt (e.g., `--include-tag key:value --include-tag otherkey:value2`).
    *   Clarify precedence/interaction:
        *   Exclusion likely takes precedence over inclusion.
        *   Entities must pass *both* inclusion/exclusion checks *and* the `require-all` check if specified.

2.  **Implement Filtering Logic:**
    *   Create a new function or set of functions (e.g., in `internal/guidance/filtering`) that takes a slice of `Summary` or `Entity` objects and the filter parameters (parsed from CLI flags) and returns the filtered slice.
    *   Implement the agreed-upon tag matching logic (case-insensitive matching for tags?).
    *   Ensure the logic correctly handles combinations of include, exclude, and require-all filters.

3.  **Integrate into CLI Commands:**
    *   Modify `cmd/summary.go` and `cmd/details.go` (or equivalent command implementations):
        *   Add the new CLI flags using a library like `cobra` or `spf13/pflag`.
        *   Parse the flag values.
        *   After retrieving the full set of summaries/details from the `GuidanceBackend` interface (`GetSummary` or `GetDetails`), pass the results and the parsed filter parameters to the filtering logic function.
        *   Use the *filtered* slice for generating the final command output (JSON).
    *   *Adherence:* Maintain separation - command logic orchestrates fetching and filtering; filtering logic is self-contained.

### 3.3. Test Plan

1.  **Unit Tests:**
    *   Test the core filtering logic function(s) in isolation.
    *   Create mock `Summary`/`Entity` slices with various tag combinations.
    *   Test each filter flag (`--include-tags`, `--exclude-tags`, `--require-all-tags`) independently with single and multiple tags.
    *   Test combinations of flags to verify precedence and interaction logic.
    *   Test edge cases: empty tag lists, non-existent tags, entities with no tags.
    *   Test case sensitivity of tag matching (if applicable).

2.  **Integration Tests:**
    *   Use a set of test guidance files with known tags.
    *   Test `agentt summary` with various filter flag combinations:
        *   No flags (should return all).
        *   `--include-tags`: Verify only entities with *any* of the specified tags are returned.
        *   `--exclude-tags`: Verify entities with *any* of the specified tags are *not* returned.
        *   `--require-all-tags`: Verify only entities with *all* specified tags are returned.
        *   Combinations: e.g., `include=a exclude=b`, `include=a require=b`, `exclude=a require=b`, `include=a exclude=b require=c`.
        *   Verify output format (JSON) remains valid.
    *   Repeat integration tests for `agentt details`, ensuring filtering applies correctly *after* fetching the details for initially requested IDs (or perhaps filtering applies *before* fetching details if `details` is modified to accept filters directly?). *Decision Point:* Does filtering apply *before* or *after* the ID selection in `details`? Applying *after* seems simpler: fetch details for specified IDs, then filter that list. Applying *before* might be more efficient but couples `details` more tightly to filtering. Assume applying *after* for now.

3.  **Usability Tests (Manual):**
    *   Manually run commands with different filter combinations to ensure the behavior feels intuitive and matches the documentation.
    *   Check help output (`agentt summary --help`, `agentt details --help`) to ensure flags are documented clearly.

## 4. Documentation Updates

*   Update `README.md` or relevant documentation to explain the pluggable backend architecture (even if only `localfs` exists initially).
*   Document the new configuration format for specifying backends.
*   Document the new CLI flags for tag filtering in `agentt summary --help` and `agentt details --help`.
*   Update any user guides or examples that use these commands.
*   Consider adding a new `behavior` or `recipe` related to CLI design standards if deemed necessary during Phase 2.