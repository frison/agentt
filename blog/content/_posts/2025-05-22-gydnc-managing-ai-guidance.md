---
layout: post
title: "gydnc: Managing AI Guidance with Git - A Human/AI Knowledge Bridge 🌉"
date: 2025-05-21 23:30:00 -0600
categories:
  - ai-tools
  - development
  - knowledge-management
provenance:
  repo: "https://github.com/frison/agentt"
  commit: "808ee58"
  prompt: "808ee58"
  modifications: []
---

In the rapid evolution of AI tooling, maintaining control over how AI behaves is becoming increasingly critical. Whether you're guiding an AI to follow specific coding conventions, adhere to your project architecture, or respect company policies, there's a growing need for structured guidance that doesn't get lost in translation. This is where `gydnc` enters the scene.

## What is gydnc? 🤔

`gydnc` (pronounced "guidance") is a lightweight command-line tool designed to manage structured guidance for AI agents. Born from the frustration of having AI guidance scattered across configuration files, documents, and conversations, it provides a Git-friendly way to create, organize, retrieve, and version-control guidance documents.

The core concept is simple: store guidance as human-readable Markdown files with YAML frontmatter, organize them hierarchically, and provide a clean CLI interface for interacting with them.

```bash
# Clone the repository and build gydnc
git clone git@github.com:frison/agentt.git && cd agentt/gydnc && make build

# Move the binary to your PATH
mv gydnc /usr/local/bin/  # or somewhere else on your PATH

# Initialize gydnc in a Git repository (IMPORTANT: use Git for version control!)
mkdir my-guidance && cd my-guidance
git init  # Create a Git repository first
gydnc init .  # Initialize gydnc in this Git repository

# Add this to your .bashrc, .zshrc, or similar shell configuration
export GYDNC_CONFIG="/path/to/your/my-guidance/.gydnc/config.yml"
```

## The Human/AI Boundary Problem 🧠🤖

One of the persistent challenges in AI tooling is the "boundary problem" — how do we effectively transfer knowledge and constraints from humans to AI systems? Traditional approaches often involve:

1. Long, unstructured prompts that are hard to maintain and version
2. Configuration files that lack the expressiveness of natural language
3. Ad-hoc systems that don't scale as your guidance needs grow

`gydnc` addresses this by creating a structured, version-controlled, and discoverable repository of guidance that both humans and AI can understand. The human writes guidance in Markdown, and the AI can programmatically access it through the CLI.

## Git Integration: Version Control for AI Guidance 📜

A critical feature of `gydnc` is its seamless integration with Git. By initializing `gydnc` in a Git repository, you gain several key benefits:

- **Version History**: Track changes to guidance over time
- **Collaboration**: Allow multiple team members to contribute and review guidance
- **Branching**: Experiment with guidance changes in separate branches
- **Rollback**: Easily revert to previous versions if needed
- **Auditing**: See who modified guidance and when

This Git-based approach makes `gydnc` particularly valuable for teams that need to maintain consistent AI guidance across multiple projects or developers.

## How It Works: The Core Workflow 🛠️

The `gydnc` workflow is designed to be simple yet powerful:

### 1. Creating Guidance

Guidance entities in `gydnc` follow a hierarchical organization with clear categories:

```bash
# Create critical behavior guidance
gydnc create --title "Safety First" \
    --description "Guidelines for ensuring code safety" \
    --tags quality:safety,scope:universal \
    must/safety-first
```

The system supports several logical hierarchies:
- `must/` for critical behaviors
- `should/` for recommended practices
- `recipes/domain/name` for procedural guidance
- `process/domain/name` for workflow processes

### 2. Discovering Guidance

Finding relevant guidance is straightforward with tag-based filtering. The tool supports everything from simple tag matches to advanced pattern matching and exclusions:

```bash
# List all guidance
gydnc list

# Get JSON output (perfect for AI consumption)
gydnc list --json

# Filter by tags
gydnc list --filter-tags "quality:safety" --json

# Advanced filtering (get everything EXCEPT safety guidance - proceed with caution!)
gydnc list --filter-tags "quality:* -quality:safety" --json
```

### 3. Retrieving Guidance

Once you've identified the guidance you need, retrieving it is easy:

```bash
# Get a specific guidance entity
gydnc get must/safety-first

# Get multiple guidance entities at once
gydnc get must/safety-first recipes/git/commit-creation
```

### 4. Updating Guidance

As your best practices and requirements evolve, you can update existing guidance entities:

```bash
# Update the title or description of existing guidance
gydnc update must/safety-first --title "Updated Safety Guidelines" --description "New description"

# Add or remove specific tags
gydnc update must/safety-first --add-tag "scope:critical,quality:essential" --remove-tag "scope:universal"

# Update the content body by piping new content
cat updated_content.md | gydnc update must/safety-first
```

## Integration Testing: The Human/AI Boundary in Practice 🧪

One of the most interesting aspects of `gydnc` is its integration test framework, which demonstrates a human/AI boundary in practice. The tests use a declarative approach where:

1. **Arrange**: Set up the test environment with sample files and configurations
2. **Act**: Execute a script that runs `gydnc` commands
3. **Assert**: Verify the outputs and filesystem changes against expected values

This declarative approach makes the tests themselves excellent examples of the human/AI boundary — they're human-readable, AI-interpretable, and serve as living documentation of the expected behavior.

For example, a simple test looks like this:

```bash
# act.sh script
./gydnc init .
./gydnc --config .gydnc/config.yml list
```

With assertions in YAML:

```yaml
# assert.yml
exit_code: 0
stdout:
  - match_type: ORDERED_LINES
    content: |
      # REGEX: ^Created guidance store: /.*\.gydnc$
      # REGEX: ^Created tag_ontology.md: /.*\.gydnc/tag_ontology.md$
      # REGEX: ^Created configuration file: /.*\.gydnc/config.yml$
      # REGEX: ^gydnc initialized successfully in /.*$
      # REGEX: ^\s*export GYDNC_CONFIG="/[^"]+"\s*$
filesystem:
  - path: ".gydnc/config.yml"
    exists: true
  - path: ".gydnc/tag_ontology.md"
    exists: true
```

This approach not only ensures the tool works as expected but also documents its behavior in a way that's accessible to both humans and machines.

## Migration from Cursor Rules: A Real-World Use Case 📝

One compelling use case that demonstrates `gydnc`'s effectiveness was the migration of guidance from cursor rules to the new format. The tool includes a detailed migration recipe (a recipe **in** gydnc) that outlines the process of converting from `.cursor/rules/*.mdc` or `.agent/{behaviors,recipes}/*` files to gydnc using only the cli.

The migration involves:
1. Source analysis (identifying file format and metadata)
2. Format conversion (adapting YAML frontmatter)
3. Tag structure alignment (using `category:value` style)
4. Path structure planning (creating logical hierarchies)
5. Content adaptation (updating references and formatting)

This process not only preserves the original guidance but enhances it with better organization and discoverability.

## Self-Dogfooding: The gydnc Interaction Framework 🔄

One of the most interesting aspects of `gydnc` is that it uses its own format to define how AI systems should interact with it. This self-dogfooding approach demonstrates the power of the tool. Here's the actual guidance framework that tells AI systems how to use gydnc effectively:

```markdown
# gydnc-interaction-framework
# Guidance Agent Interaction Framework

## Intent
Ensure effective guidance retrieval and creation through the gydnc CLI tool, adapting to evolving user requests throughout a session.

## Rules

### 1. Guidance Retrieval Workflow
ALWAYS follow this sequence to ensure you have comprehensive guidance:

1. **BEGIN WITH OVERVIEW:** Start EVERY session by getting a complete overview of available guidance:
    ```bash
    # CRITICAL: Get overview of ALL available guidance entities
    gydnc list --json
    ```
    This step is NON-OPTIONAL. You must begin by understanding what guidance is available.

2. **FETCH DETAILED GUIDANCE:** After identifying relevant guidance from the overview, retrieve full details:
    ```bash
    # Get complete guidance content for multiple entities in one command
    gydnc get <entity1> <entity2> <entity3>
    ```
    Do NOT use the --json flag with 'get' commands, as the default output provides the complete guidance text.

3. **PREFER BATCH RETRIEVAL:** Always fetch multiple relevant guidance entities in a single command rather than separate commands.

4. **RE-FETCH AS REQUESTS EVOLVE:** When the user's request changes direction or introduces new requirements, IMMEDIATELY fetch additional relevant guidance:
    ```bash
    # Example: When user asks about a new topic (e.g., "write a blog post")
    gydnc list --json
    gydnc get <relevant-blog-writing-guidance>
    ```
    It is CRITICAL to adapt and fetch new guidance as the conversation progresses.

### 2. Guidance Creation Workflow
When creating new guidance entities:

1. **CAPTURE EXISTING CONTENT:** Use stdin piping to create guidance from existing content:
    ```bash
    # Pipe content directly into a new guidance entity
    cat existing-content.md | gydnc create --title "Title" --description "Description" --tags "tech:git,lang:go,repo:agentt" <alias>
    ```
    This is the PREFERRED method for creating guidance with existing content.

2. **COMPLETE METADATA:** Always include comprehensive metadata when creating:
    ```bash
    # Create with full metadata
    gydnc create --title "Comprehensive Title" --description "Detailed description" --tags "lang:go,repo:backend,core:must" <alias>
    ```

3. **USE HIERARCHICAL ORGANIZATION:** Organize guidance in logical hierarchies:
    ```bash
    # Create entity in appropriate category structure
    gydnc create development/backend/api-design/<entity-name>
    ```

### 3. Guidance Update Workflow
When updating existing guidance:

1. **TARGETED UPDATES:** Make specific updates to metadata or content:
    ```bash
    # Update title, description, or add/remove specific tags
    gydnc update <alias> --title "Updated Title" --description "New description"
    gydnc update <alias> --tags-add "quality:critical" --tags-remove "scope:universal"
    ```

2. **CONTENT UPDATES:** Update the body content with meaningful changes:
    ```bash
    # Update content from a file
    gydnc update <alias> --body-from-file updated_content.md

    # Document the reason for updates (useful for Git commit messages)
    gydnc update <alias> --reason "Updated to reflect new requirements"
    ```
```

This guidance document shows how AI systems should interact with the `gydnc` tool, emphasizing the importance of first listing available guidance, then retrieving relevant entities, and adapting to changing requirements throughout a conversation. It's a perfect example of using guidance to manage guidance—meta-guidance that improves AI interactions.

## Looking Forward 🚀

`gydnc` is still in its early stages, but it represents a promising approach to the human/AI boundary problem. Future developments might include:

1. **Multiple Backend Support**: Beyond the current filesystem backend, supporting distributed storage
2. **Full Content-Addressable IDs**: Moving towards content-based identifiers for robust provenance
3. **Structured Ontology Management**: Enhanced tools for managing tag ontologies and relationships
4. **Integration with AI Platforms**: Direct integration with popular AI tools and frameworks

The real power of `gydnc` lies in its simplicity and adaptability. By focusing on human-readable files, Git integration, and a clean CLI, it creates a knowledge bridge between humans and AI that can evolve with your needs.

If you're managing AI guidance in any form, `gydnc` might be worth a look. Check it out at [github.com/frison/agentt/blob/main/gydnc/README.md](https://github.com/frison/agentt/blob/main/gydnc/README.md) for full documentation and usage instructions.

---

*This article was originally created in commit [`808ee58`](https://github.com/frison/agentt/commit/808ee58).*