# Behavior Domain Directory

This directory contains the core components of the project's **behavioral domain**. It structures requirements, guidelines, and operational procedures into distinct tiers, defining *how* work MUST or SHOULD be performed.

## Structure and Purpose

*   **`.agent/behavior/must/`**: Contains MUST-level behaviour definitions (`.bhv` files).
*   **`.agent/behavior/should/`**: Contains SHOULD-level behaviour recommendations (`.bhv` files).
*   **`.agent/behavior/FORMAT.md`**: Defines the required format for all `.bhv` files.
*   **`.agent/behavior/bin/`**: Contains supporting scripts for the behavioral domain, primarily the discovery tool.

## File Format

All `.bhv` files within the `must/` and `should/` subdirectories MUST adhere to the structure defined in `.agent/behavior/FORMAT.md`.

## Discovery

The primary mechanism for discovering relevant behavioral directives is the script located at `.agent/behavior/bin/discover.sh`. It should be invoked specifying the target tier (must/should) and desired output format (e.g., json):

```bash
# Discover MUST directives
.agent/behavior/bin/discover.sh .agent/behavior/must json

# Discover SHOULD directives
.agent/behavior/bin/discover.sh .agent/behavior/should json
```
*(Note: Specific agent rules, like those in `.cursor/rules/`, may enforce the execution order and usage of this script.)*

## Relation to Other Domains

*   **Cookbook:** While this domain defines *how* work MUST/SHOULD be done, the Cookbook (see `.agent/COOKBOOK.md`) provides specific procedures for *accomplishing* tasks.
*   **Reflexes:** Reflexes (potentially defined elsewhere) represent executable capabilities that might be governed by behavioral directives.