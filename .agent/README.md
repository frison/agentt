# Agent Guidance Directory (`.agent/`)

This directory centralizes project-specific guidance, rules, and procedures intended for use by AI assistants or humans following defined processes.

## Purpose

The goal is to clearly separate AI/process-related meta-information from the primary functional code or content of the repository.

## Domains

This directory houses distinct domains of guidance:

*   **`.agent/behavior/`**: Defines the **behavioral domain** - how work MUST or SHOULD be performed. Contains MUST/SHOULD directives (`.bhv` files) and its own README detailing structure and discovery.
*   **`.agent/cookbook/`**: Contains the **procedural domain** - concrete recipes or how-tos (`.rcp` files) for specific tasks within the repository. See `.agent/cookbook/README.md` for details.
*   **(Future domains)**: Other types of agent guidance could be added here (e.g., `personas/`, `reflexes/` if moved here later).

## Interaction

AI assistants interacting with this repository should:

1.  Consult `.cursor/rules/agent-interaction-framework.mdc` as the primary entry point.
2.  Refer to the specific README file for each domain (`.agent/behavior/README.md`, `.agent/cookbook/README.md`) to understand its structure and discovery mechanism.
3.  Follow the enforcement rules within `.cursor/rules/` when interacting with each domain.