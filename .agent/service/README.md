# Agent Guidance Service (`agentt`)

This service provides:
1.  An HTTP API for discovering and retrieving agent guidance definitions (behaviors, recipes).
2.  A command-line interface (`agentt`) for the same purpose, suitable for direct agent interaction or scripting.

## Configuration

The `agentt` service and CLI commands require a configuration file (`config.yaml` or `agentt.yaml` by default) to locate the guidance definition files (behaviors, recipes).

The path to the configuration file is determined using the following order of precedence:

1.  **`--config` / `-c` Command-Line Flag:**
    ```bash
    agentt summary --config path/to/your/config.yaml
    agentt server start -c path/to/your/config.yaml
    ```
    This flag explicitly provides the path and overrides all other methods.

2.  **`AGENTT_CONFIG` Environment Variable:**
    ```bash
    export AGENTT_CONFIG="path/to/your/config.yaml"
    agentt summary
    ```
    If the flag is not provided, the tool checks this environment variable.

3.  **Default Search Paths:**
    If neither the flag nor the environment variable is set, `agentt` searches for a configuration file in the following locations relative to the **current working directory**:
    *   `./config.yaml`
    *   `./agentt.yaml`
    *   `./.agent/service/config.yaml`
    *   `./.agentt/config.yaml`

    The first path found in this list will be used.

If no configuration file is found through any of these methods, the command will exit with an error.

### Configuration File Format (`config.yaml`)

The configuration file uses YAML format:

```yaml
# Address for the HTTP server to listen on (if running server)
listenAddress: ":8080"

# Definitions of all known entity types and their required fields
entityTypes:
  - name: "behavior"
    description: "Defines rules, constraints, or preferred practices for agent operation."
    requiredFields:
      - "id"          # Now explicitly required for all types
      - "title"
      - "tier"
      - "priority"
      - "tags"
      - "description"

  - name: "recipe"
    description: "Provides step-by-step instructions or procedures for specific tasks."
    requiredFields:
      - "id"
      - "title"
      - "priority"
      - "tags"
      - "description"

# Configuration for the selected guidance backend
backend:
  type: localfs # Specifies the backend type (currently only "localfs" supported)
  settings:
    # Settings specific to the "localfs" backend:
    rootDir: "." # Base directory for resolving globs below.
                 # Relative paths resolved against project root (where agentt runs).
                 # Defaults to "." if omitted.
    requireExplicitID: true # If true, files without an 'id' field in frontmatter are ignored.
                            # Defaults to true.
    entityLocations: # Maps entity type name (from entityTypes above) to glob patterns
      behavior:      # Key must match an entityType name
        pathGlob: ".agent/behavior/**/*.bhv" # Glob pattern relative to rootDir
      recipe:        # Key must match an entityType name
        pathGlob: ".agent/cookbook/**/*.rcp"
      # ... add entries for other entity types if needed ...
```

**Note:** The `pathGlob` within `entityTypes` is interpreted relative to the location of the loaded `config.yaml` file itself.

## Usage

### Server

Start the HTTP server:
```bash
./agentt server start [-c path/to/config.yaml]
```

The server logs basic information about each incoming request (method, path, status, duration, source) to standard output.

#### API Endpoints

*   `GET /health`: Returns `200 OK` if the server is running.
*   `GET /entityTypes`: Returns a JSON array of configured entity type definitions from `config.yaml`.
*   `GET /summary`: Returns a JSON array of summaries for all valid guidance entities. Each summary includes a prefixed ID (`bhv-`, `rcp-`), type, tier (for behaviors), tags, and description.
*   `POST /details`: Expects a JSON body `{"ids": ["prefixed-id-1", "prefixed-id-2"]}`. Returns a JSON array containing the full details for the requested valid prefixed IDs.
*   `GET /llm.txt`: Returns the embedded agent interaction protocol text.

### CLI

Get summaries (outputs JSON):
```bash
./agentt summary [-c path/to/config.yaml]
```

Get specific details (outputs JSON):
```bash
./agentt details --id bhv-some-id --id rcp-other-id [-c path/to/config.yaml]
```

#### Filtering CLI Output

Both `summary` and `details` commands support a `--filter` flag to filter the results based on metadata *after* they have been retrieved from the backend.

```bash
# Get summaries for all 'must' tier behaviors
./agentt summary --filter 'tier:must AND type:behavior'

# Get details for specific IDs, but only show those tagged 'core'
./agentt details --id bhv-safety-first --id sep-of-concerns --filter 'tag:core'
```

**Filter Query Syntax (Basic):**
*   Terms: Match against fields (e.g., `tag:core`, `type:behavior`, `tier:must`). Tag values are case-insensitive.
*   (Note: Full support for boolean operators like `AND`, `OR`, `NOT`, and grouping `()` is planned but the current parser implementation is basic and may only support simple `key:value` terms.)

Show agent interaction help (outputs text):
```bash
./agentt llm | cat
```

*(Ensure `./agentt` refers to the built binary)*