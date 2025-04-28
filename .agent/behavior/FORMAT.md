---
title: "Agent Behaviours Format (`.bhv` files)"
priority: 102
description: "Defines the standard structure and format for all behavioral MUST Requirement and SHOULD Recommendation files."
tags: ["meta", "format", "structure", "must", "should", "behavior", "agent"]
---

# Agent Behaviours Format (`.bhv` files)

## Intent
This document defines the standard structure, format, and metadata for all agent behaviour files (`must/*.bhv` and `should/*.bhv`) within the `.agent/behavior/` directory. This ensures consistency, discoverability, and appropriate application of guidance across the system.

## Rules
- All behaviour files MUST use the `.bhv` extension.
- Filenames SHOULD use a concise, kebab-case name describing the content (e.g., `build-process.bhv`, `shell-safety.bhv`).
- All behaviour files MUST begin with YAML frontmatter bounded by triple dashes (`---`).
- Frontmatter MUST include these fields:
  - `title`: "Clear and Concise Title" (String)
  - `priority`: NNN (Numeric value, lower = higher importance)
  - `description`: "A short (1-2 sentence) explanation." (String)
  - `tags`: `["list", "of", "relevant", "tags"]` (Array of strings)
- Frontmatter fields MUST appear in the order specified above.
- The main body MUST follow the frontmatter.
- Body MUST be structured Markdown with at least an "Intent" or "Core Statement" section explaining the purpose.
- Body SHOULD include relevant sections like "Rules", "Actions", "Applications", "Examples", "Common Mistakes", or "Exceptions" as appropriate for the content.

## Examples

### MUST Behaviour Example (`must/*.bhv`)
```yaml
---
title: "Example Core MUST Behaviour"
priority: 1
description: "This is a fundamental, non-negotiable behaviour for the system."
tags: ["core", "example", "fundamental"]
---

# Example Core MUST Behaviour

## Core Statement
All systems must adhere to this fundamental principle X.

## Rationale
Why principle X is critical...

## Applications
- How X applies in scenario A...
- How X applies in scenario B...
```

### SHOULD Behaviour Example (`should/*.bhv`)
```yaml
---
title: "Example Implementation Recommendation"
priority: 201
description: "Standard procedure for implementing feature Y according to our needs."
tags: ["implementation", "workflow", "feature-y"]
---

# Example Implementation Recommendation

## Intent
Provide the standard steps for implementing Feature Y.

## Actions
- Step 1: Do this...
- Step 2: Then do that...

## Rules
- Always check condition Z before proceeding.

## Examples
```bash
# Code example demonstrating the practice
example_command --option
```

## Exceptions
- When integrating with legacy system Q, use alternative procedure Z.