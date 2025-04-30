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
# Address for the HTTP server to listen on
listenAddress: ":8080"

# Definitions of entity types to discover
entityTypes:
  - name: "behavior"                             # Unique type name
    description: "Defines behavioral directives..." # For documentation
    pathGlob: "../behavior/**/*.bhv"             # Glob pattern relative to config file location
    fileExtensionHint: ".bhv"                    # Optional hint
    requiredFrontMatter:                       # List of required frontmatter keys
      - "title"
      - "priority"
      - "description"
      - "tags"

  - name: "recipe"
    description: "Provides procedural steps..."
    pathGlob: "../cookbook/**/*.rcp"
    fileExtensionHint: ".rcp"
    requiredFrontMatter:
      - "id"
      - "title"
      - "priority"
      - "description"
      - "tags"
```

**Note:** The `pathGlob` within `entityTypes` is interpreted relative to the location of the loaded `config.yaml` file itself.

## Usage

### Server

Start the HTTP server:
```bash
./agentt server start [-c path/to/config.yaml]
```

### CLI

Get summaries:
```bash
./agentt summary [-c path/to/config.yaml]
```

Get specific details:
```bash
./agentt details --id bhv-some-id --id rcp-other-id [-c path/to/config.yaml]
```

Show agent interaction help:
```bash
./agentt llm
```

*(Ensure `./agentt` refers to the built binary)*