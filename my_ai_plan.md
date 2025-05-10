# AI Plan: Add Guidance CRUD Operations to Agentt CLI

**Request Date:** 2025-05-04
**Status:** Planned

## 1. Goal
Enhance the `agentt` CLI to support creating and updating guidance entities (behaviors, recipes) directly, while respecting the multi-backend architecture and incorporating necessary safety mechanisms. Deletion via the CLI will not be supported.

## 2. Target Commands
- `agentt create <type> --backend <backend-name|index> [flags...]`
- `agentt update <entity-id> [flags...]`
- (Consider) `agentt edit <entity-id>` (Opens entity body in `$EDITOR`)

## 3. Core Requirements

*   **Entity-to-Backend Mapping:** The guidance loading mechanism must be extended to retain information about which backend each loaded entity originates from. This is crucial for targeting updates.
*   **Backend Writability:** The backend configuration (`config.yaml`) needs a way to specify if a backend instance is writable by the CLI (e.g., `writable: true/false` flag for each backend entry). The backend implementation must respect this flag.
*   **Safety Mechanisms:**
    *   Mandatory user confirmation prompts for `update` operations.
    *   Clear error handling for attempts to modify non-writable backends.
    *   Acknowledge that AI-initiated calls to these commands will likely trigger external confirmation prompts due to inherent AI safety protocols.
*   **Data Validation:** Use the `entityTypes` definitions from `config.yaml` to validate metadata provided during `create` and `update` operations.
*   **Targeting:**
    *   `create`: Must specify the target *writable* backend if multiple writable backends exist.
    *   `update`: Must operate on the specific backend where the entity resides, failing if that backend is not writable.

## 4. Phase 1: Backend Infrastructure Enhancements

1.  **Modify Guidance Loading:** Update the `internal/guidance` loading logic (including backend interfaces and implementations like `localfs`):
    *   When entities are loaded, store a reference (e.g., name or index) to their originating backend alongside the entity data/summary.
    *   This mapping needs to be accessible by the CLI command logic.
2.  **Add `writable` Config Flag:**
    *   Update the `internal/config` structs to include an optional `writable` boolean field for backend definitions (defaulting to `false`).
    *   Ensure `LoadConfig` parses this flag.
3.  **Update Backend Instantiation:** Backend constructors should accept and store their `writable` status.

## 5. Phase 2: `localfs` Create/Update Implementation

1.  **Define `WritableBackend` Interface:** Create an interface (e.g., in `internal/guidance/backend`) with methods like:
    *   `CreateEntity(entityData map[string]interface{}, body string) error`
    *   `UpdateEntity(entityID string, updatedData map[string]interface{}, updatedBody string) error`
2.  **Implement Interface for `localfs`:**
    *   Modify the `localfs` backend implementation to satisfy the `WritableBackend` interface.
    *   Implement the logic to create and modify (both frontmatter and body) the underlying entity files (`.bhv`, `.rcp`) within the backend's configured `rootDir` and `entityLocations`.
    *   Ensure these methods check the backend's `writable` status before proceeding.
3.  **Develop CLI Commands:**
    *   Implement `create`, `update` commands in `cmd/`.
    *   **`create`:**
        *   Requires entity `type` (behavior/recipe).
        *   Requires `--backend` flag if multiple writable backends exist.
        *   Needs flags for required metadata (`--id`, `--title`, etc.) and potentially optional ones (`--tag`, `--priority`).
        *   Needs mechanism for body input (e.g., `--body-from-file <path>`, `--body-stdin`, or piped input).
        *   Performs validation against `entityTypes`.
        *   Calls `CreateEntity` on the target backend.
    *   **`update`:**
        *   Requires `entity-id`.
        *   Determines the source backend using the mapping from Phase 1.
        *   Checks if the source backend is writable.
        *   Accepts flags for metadata fields to update.
        *   Accepts body input mechanism.
        *   Requires confirmation prompt.
        *   Performs validation.
        *   Calls `UpdateEntity`.
    *   **(Optional) `edit`:**
        *   Requires `entity-id`.
        *   Determines source backend and file path.
        *   Checks writability.
        *   Launches `$EDITOR` on the entity file.

## 6. Phase 3: Handling Non-Writable Backends

1.  **Implement User Feedback:** Ensure commands provide clear, informative messages when an operation targets a non-writable backend (e.g., "Cannot update entity '<id>' because it belongs to read-only backend '<name>'. Please modify the source directly.").
2.  **Refine Error Handling:** Ensure consistent error handling across different backend types (even if only `localfs` is writable initially).

## 7. Phase 4: Testing & Refinement

1.  **Unit Tests:** Add tests for:
    *   Backend mapping logic.
    *   `WritableBackend` implementation for `localfs`.
    *   CLI command argument parsing and validation.
    *   Safety confirmation prompts (mocking user input).
    *   Correct backend targeting and writability checks.
2.  **Integration/Manual Testing:** Test the commands with various `config.yaml` setups (multiple backends, mixed writability), different entity types, edge cases, and error conditions.

## 8. Considerations

*   **AI Safety Integration:** Explicitly document that AI usage of these commands will require user confirmation via external mechanisms. The CLI itself doesn't bypass these.
*   **Complex Metadata:** How to handle updates to list-based metadata like `tags` (add, remove, replace?). Flags might become complex; `edit` command might be simpler.
*   **File Formatting:** Ensure created/updated files maintain consistent formatting (e.g., YAML frontmatter separation).
*   **Atomicity:** Acknowledge potential lack of atomicity for file system operations.
*   **Future Backends:** Design the `WritableBackend` interface with potential future implementations (e.g., database, Git API) in mind.