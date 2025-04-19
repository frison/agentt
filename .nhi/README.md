# NHI (Non-Human Intelligence) Framework

This directory contains the NHI framework - a structured approach to guide and govern how non-human intelligences (AI systems) interact with this codebase.

## Directory Structure

```
.nhi/
├── principles/          # Core values and reasoning frameworks
│   ├── universal/       # Always-applicable principles
│   └── disciplines/     # Domain-specific principles
├── directives/          # Structure and organization patterns
│   ├── meta/            # Meta-directives about directives themselves
│   ├── core/            # Fundamental operational directives
│   ├── domain/          # Domain-specific directives
│   └── contexts/        # Context-sensitive directives
├── actions/             # Specific implementation instructions
│   ├── common/          # Generally applicable actions
│   └── specific/        # Task-specific actions
├── tools/               # Tools for working with NHI files
│   └── discover.sh      # Discovery script for NHI guidance
└── tmp/                 # Temporary storage for NHI operations
```

## Three-Tier Structure

The NHI framework uses a three-tier structure to separate different types of guidance:

1. **Principles** (`.nhp` files): Core values and reasoning frameworks
   - Abstract, enduring guidelines that rarely change
   - Explain "why" certain patterns are preferred
   - Set guardrails for technical decision-making

2. **Directives** (`.nhd` files): Structure and organization patterns
   - Define the format, organization, and discovery mechanisms
   - Establish conventions for code organization
   - Maintain consistent patterns across the codebase

3. **Actions** (`.nha` files): Specific implementation instructions
   - Task-oriented and focused on implementation
   - Clear "if situation X, then do Y" guidance
   - Directly executable by AI systems

This separation helps minimize cognitive load by clearly distinguishing between the "why" (principles), the "what" (directives), and the "how" (actions).

## Using the NHI Framework

To discover relevant guidance:

```bash
# List all NHI files
.nhi/tools/discover.sh

# List files in a specific format
.nhi/tools/discover.sh .nhi table

# Filter by type
.nhi/tools/discover.sh .nhi json principles

# Search in a specific directory
.nhi/tools/discover.sh .nhi/principles/universal
```

When working with the codebase, AI systems should:

1. Start with relevant principles to understand core values
2. Follow applicable directives for structure and organization
3. Apply specific actions for implementation details
4. Prioritize based on priority values (lower numbers = higher priority)
5. Always apply universal principles

## File Formats

### Principles (`.nhp`)
```yaml
---
title: "Principle Name"
priority: 1-10 (1 = highest)
universal: true|false
disciplines: ["area1", "area2"]
---
```

### Directives (`.nhd`)
```yaml
---
title: "Directive Name"
priority: 1-10
scope: "meta|global|domain|context"
binding: true|false
---
```

### Actions (`.nha`)
```yaml
---
title: "Action Name"
priority: 1-10
applies_to: ["glob/patterns/*"]
guided_by: ["principles/that/apply.nhp"]
---
```

## What are Directives?

Directives are structured guidance documents for AI systems. Each directive:

1. Has a clear scope and priority
2. Provides explicit rules and guidance
3. Includes examples to illustrate correct patterns
4. Specifies exceptions where rules may not apply

Directives use the `.nhd` file extension and follow a consistent format defined in `.nhi/directives/meta/001-directive-format.nhd`.

## Using Directives

To discover relevant directives:

```bash
# List all directives in JSON format
.nhi/tools/discover.sh

# List all directives in tabular format
.nhi/tools/discover.sh .nhi/directives table

# List directives in a specific subdirectory
.nhi/tools/discover.sh .nhi/directives/core
```

When working with the codebase, AI systems should:

1. Find relevant directives for the current task
2. Follow high-priority and binding directives
3. Consider lower-priority directives as guidance
4. Respect the specified exceptions

## Contributing

To add or modify directives:

1. Follow the format specified in `.nhi/directives/meta/001-directive-format.nhd`
2. Use the appropriate subdirectory based on directive scope
3. Assign a meaningful priority (1-10, with 1 being highest)
4. Include clear examples and exceptions
5. Mark as `binding: true` only if the directive must always be followed