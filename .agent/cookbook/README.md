# Cookbook Domain (Practices / How-Tos)

## Purpose

The Cookbook contains concrete, step-by-step instructions, guides, or recipes for completing specific tasks within the context of a particular project or repository. It focuses on the "how-to" for common or complex procedures.

## Location & Structure

Cookbook recipes reside in the `.agent/cookbook/` directory. They are organized into subdirectories based on category or domain, for example:

```
.agent/cookbook/
├── meta/       # Recipes about the cookbook or rules themselves
├── blog.frison.ca/ # Recipes specific to the blog
├── git/        # Recipes related to Git workflows
└── ...         # Other categories
```

## Format

Recipes are defined in files with a `.rcp` extension. They MUST contain YAML frontmatter (defined below) followed by the procedural steps in Markdown.

```yaml
---
# === Identity & Discovery ===
id: unique-recipe-id      # REQUIRED: Unique, stable, kebab-case identifier.
title: "Human Readable Title"  # REQUIRED: Title for display/search.
priority: 200               # REQUIRED: Numeric priority.
description: "Short summary." # REQUIRED: Explains the goal.
tags: ["keyword", "example"] # REQUIRED: Search tags.

# === Context & Applicability ===
domain: category              # Optional: Primary domain (e.g., blog, git).
applies_to_paths: []        # Optional: Glob patterns.
tools_required: []           # Optional: List of external tools.

# === Relationships ===
related_guidance:           # Optional: Links relative to .agent dir.
  behavior: []
# related_reflexes: []
calls_recipes: []           # Optional: List of recipe IDs this recipe invokes.
---

# Recipe Title (matches frontmatter)
...(Markdown steps)...
```

## Discovery

Recipes are discovered by searching the `.agent/cookbook/` directory for `.rcp` files. The standard command (executed from the project root) is:

```bash
# Discover script (preferred):
.agent/cookbook/bin/discover.sh

# Manual alternative (basic):
# find .agent/cookbook/ -name "*.rcp" -type f | cat
```
*(Note: Specific agent rules enforce using the `discover.sh` script.)*

## Relation to Behavior Domain

While the Behavior domain (`.agent/behavior/`) defines *how* work MUST or SHOULD be done generally, Cookbook recipes provide specific procedures for *accomplishing* tasks within those behavioral boundaries.