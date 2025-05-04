# AI Plan: Refactor Agentt Configuration Handling

**Goal:** Update the `agentt` tool to use a more robust and flexible configuration structure (v4.2), supporting multiple backends and using config-relative paths.

**Target Configuration Format (v4.2):**

```yaml
# Network address for the optional HTTP server component
listenAddress: ":8080"

# Defines metadata ABOUT the entity types used in the system.
# This definition is shared across all configured backends below.
entityTypes:
  - name: "behavior"
    description: "Defines rules, constraints, or preferred practices for agent operation."
    # List of fields REQUIRED in the data structure for each entity of this type,
    # as provided by any backend.
    requiredFields:
      - "id"          # Unique identifier for the entity (Mandatory for all entities)
      - "title"
      - "tier"        # e.g., "must", "should" (Meaningful for behaviors)
      - "priority"    # e.g., 1 (highest) to 100 (lowest)
      - "tags"        # List of relevant keywords
      - "description" # Short summary of the entity's purpose

  - name: "recipe"
    description: "Provides step-by-step instructions or procedures for specific tasks."
    requiredFields:
      - "id"          # Unique identifier for the entity (Mandatory for all entities)
      - "title"
      # Note: 'tier' is intentionally omitted here as it's not semantically relevant for recipes
      - "priority"
      - "tags"
      - "description"

# Configures the backends used to load entity definition files.
# Backends are processed in the order they appear in this list.
# The system may issue warnings if entities with the same ID are found across
# different backends, but will attempt to load all entities.
backends:
  - # --- First Backend Configuration ---
    # Unique identifier for this backend instance (optional but recommended)
    # name: "primary_local"
    type: localfs
    # Settings specific to this 'localfs' backend instance:
    # Base directory for resolving the 'entityLocations' globs below.
    # *** Path is relative to the directory containing this config file. ***
    rootDir: "." # Example: "." means the same directory as the config file.
    # Maps entity type names to glob patterns for *this* backend.
    entityLocations:
      behavior: ".agent/behavior/**/*.bhv" # Resolved relative to config file dir + rootDir
      recipe: ".agent/cookbook/**/*.rcp"   # Resolved relative to config file dir + rootDir

  # - # --- Second Backend Configuration (Example) ---
  #   name: "shared_behaviors_fs"
  #   type: localfs
  #   rootDir: "../shared_guidance_repo" # Example: Goes up one level from config file dir
  #   entityLocations:
  #     behavior: "behaviors/**/*.bhv"

  # - # --- Third Backend Configuration (Hypothetical Example) ---
  #   name: "central_database"
  #   type: database
  #   connectionString: "..."
  #   behaviorQuery: "SELECT id, title, tier, priority, tags, description, body FROM agent_behaviors WHERE ..."
  #   recipeQuery: "SELECT id, title, priority, tags, description, body FROM agent_recipes WHERE ..."
```

**Phase 1: Implement Configuration Parsing**

1.  **Define Go Structs:** Create/Update Go structs in the `internal/config` package to accurately represent the target YAML structure shown above. This includes:
    *   Top-level `Config` struct with `ListenAddress`, `EntityTypes` (slice), and `Backends` (slice).
    *   `EntityType` struct with `Name`, `Description`, `RequiredFields` (slice of strings).
    *   `Backend` struct (potentially an interface or using generics/embedding if types differ significantly, but initially focus on `localfs`).
    *   `LocalFSBackendSettings` struct containing `RootDir` (string) and `EntityLocations` (map[string]string). Embed or include this in the `Backend` struct for `type: localfs`.

2.  **Update YAML Parsing:** Modify the `LoadConfig` function (and potentially `FindAndLoadConfig`) in `internal/config/config.go`:
    *   Use a suitable YAML parsing library (e.g., `gopkg.in/yaml.v3`) to unmarshal the config file contents into the new Go structs.
    *   Implement *general* validation logic within `LoadConfig`:
        *   Ensure required top-level fields (`EntityTypes`, `Backends`) are present.
        *   Validate `entityTypes[*].requiredFields` includes `"id"`.
        *   Ensure `entityTypes[*].Name` is present and unique.
        *   Ensure `backends[*].Type` is present.
    *   **Note:** Backend-specific validation (e.g., checking `rootDir` and `entityLocations` for `localfs`, or validating `entityLocations` keys against known `entityTypes`) will be moved to the backend instantiation logic in Phase 2.

**Phase 2: Update Backend Logic**

1.  **Modify Backend Interface/Implementation:** Refactor the guidance backend interface (likely in `internal/guidance/backend`) and the `localfs` implementation:
    *   The backend initialization (`NewLocalFSBackend` or similar) should accept its specific configuration settings (derived from one entry in the `Backends` list in the config) and the *path of the loaded configuration file*.
    *   **Perform backend-specific validation** within the initialization function (e.g., `NewLocalFSBackend` must validate `rootDir`, `entityLocations`, and ensure `entityLocations` keys match defined `entityTypes`).
    *   The logic for finding files (e.g., in `loadFiles` or similar methods) must now resolve paths correctly based on the config file's directory and the validated `rootDir`.
    *   Implement file path resolution: Config Dir + `rootDir` + `entityLocations[*]` glob -> `filepath.Glob`.

2.  **Adapt Core Logic:** Update the main command logic (e.g., in `cmd/summary.go`, `cmd/details.go`, potentially `server` logic) to handle the list of backends:
    *   Iterate through the configured `backends` in order.
    *   Instantiate each backend (which includes its specific validation) using its settings and the config file path.
    *   Load entities from each backend.
    *   Implement logic to merge results and detect/warn about duplicate entity IDs found across different backends.

**Phase 3: Testing and Refinement**

1.  **Update Unit Tests:** Adjust existing unit tests for `internal/config` and backend loading to reflect the new structure and path resolution logic. Add tests for multiple backends and ID conflict warnings.
2.  **Manual Testing:** Test the `agentt summary`, `agentt details`, and potentially `agentt server start` commands with various config files, `rootDir` values, and file structures to ensure correct behavior.

**Considerations:**

*   **Error Handling:** Ensure robust error handling during config parsing, path resolution, and file loading.
*   **Logging:** Add informative logging, especially for path resolution steps and duplicate ID warnings.