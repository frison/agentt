---
layout: post
title: "From find to gydnc: The Evolution of Guidance Discovery in .cursor 🔍"
date: 2025-05-23 22:21:38 -0600
categories:
  - guidance-systems
  - discovery-patterns
  - gydnc
  - evolution
  - ai-tooling
---

The `.cursor` directory tells a fascinating story of evolution – not just of code, but of how we think about organizing, discovering, and managing AI guidance. What started as simple file-finding scripts has evolved into `gydnc`, a sophisticated content-addressable guidance system that's now been spun off to its own repository at [github.com/ofthemachine/gydnc](https://github.com/ofthemachine/gydnc). Let's trace this journey from humble `find` commands to a mature guidance ecosystem.

## The Humble Beginnings: discover.sh and File Walking 📁

In the beginning, there was `discover.sh` – a simple Go program that walked directory trees looking for `manifest.yml` files. This was the first attempt at systematic guidance discovery:

```go
// From reflexes/.base-tools/src/basetools/cmd/discover-reflexes/main.go
func findAndParseReflexes(rootDir string) ([]discoverytypes.DiscoveredReflex, error) {
    var discovered []discoverytypes.DiscoveredReflex

    err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
        if !info.IsDir() && info.Name() == manifestFileName {
            reflex, err := parseManifest(path, rootDir)
            if err != nil {
                fmt.Fprintf(os.Stderr, "Error parsing manifest %s: %v\n", path, err)
                return nil // Continue walking
            }
            discovered = append(discovered, reflex)
        }
        return nil
    })

    return discovered, nil
}
```

This approach was straightforward: find files matching a pattern, parse their YAML frontmatter, and return structured data. It worked well for "Reflexes" – the containerized development environments that needed systematic discovery. But as guidance needs grew beyond just development environments, the limitations became apparent.

## The Command Line Era: agentt and Structured Guidance 🛠️

Enter `agentt` – the first proper guidance management tool. This marked a shift from passive discovery to active guidance management. The tool introduced several key concepts:

**Structured File Types:**
- `.bhv` files for behaviors (MUST/SHOULD rules)
- `.rcp` files for recipes (step-by-step procedures)
- ID prefixes (`bhv-`, `rcp-`) to prevent naming collisions

**API-First Design:**
```bash
# The agentt CLI provided both API and command-line access
agentt summary                    # Overview of all guidance
agentt details --id bhv-safety-first  # Detailed content
agentt summary --filter "tier:must quality:safety"  # Filtered results
```

The `agentt` tool represented a major philosophical shift. Instead of simple file discovery, it introduced:

1. **Categorization** - Different types of guidance served different purposes
2. **Filtering** - Advanced query capabilities to find relevant guidance
3. **Dual Interface** - Both API endpoints and CLI commands for different use cases
4. **Efficiency** - Token-conscious design to avoid overwhelming AI systems

As one blog post from the era noted: "Our AI agents, before starting any task, needed to fetch guidance... The existing protocol required three separate API calls: one for mandatory 'MUST' rules, one for suggested 'SHOULD' rules, and one for relevant 'recipes'." The solution was elegant: a summary-first approach that let agents discover what was available, then request only what they needed.

## The find Command Interlude: Ad-Hoc Discovery 🔎

Between structured tools, there was often a return to basics. Simple `find` commands became the go-to for quick guidance discovery:

```bash
# The classic pattern for finding guidance files
find .cursor -name "*.bhv" -o -name "*.rcp" | head -10

# More sophisticated searches
find . -path "*/guidance/*" -name "*.md" | grep -E "(recipe|behavior|rule)"
```

This represented the eternal tension in tooling: sometimes you just need a quick answer, and a well-crafted `find` command is faster than spinning up a complex tool. The find pattern persisted throughout the evolution because it filled the gap when formal tools were either too heavy or didn't exist yet.

## The Content-Addressable Revolution: gydnc 🚀

The most recent evolution is `gydnc` (guidance) – a complete reimagining of how guidance should work. Now spun off to its own repository, gydnc introduces several revolutionary concepts:

**Content-Addressable IDs (CIDs):**
Every piece of guidance gets a cryptographic hash based on its content, making versions immutable and verifiable.

**Unified .g6e Format:**
```yaml
---
title: "Blog Post Creation"
description: "Steps for creating, formatting, and committing a new blog post"
tags:
  - topic:blog
  - process:content_creation
  - tech:markdown
pCID: "sha256-abc123..." # Previous version's CID
author: "cli:user@example.com"
timestamp: "2025-05-23T22:21:38Z"
changeReason: "Updated workflow for gydnc integration"
---
# Recipe: Create New Blog Post

## Prerequisites
- Adherence to the Agent Interaction Framework
- Familiarity with blog writing guidelines
...
```

**Git-Backed Provenance:**
Every change is tracked through Git, with automatic commits and full audit trails.

**Multi-Backend Architecture:**
```yaml
# gydnc.conf
storage:
  default_backend: "localfs"
  backends:
    localfs:
      type: "localfs"
      path: "./guidance"
      git:
        enabled: true
        auto_commit: true
```

**Sophisticated CLI:**
```bash
# Modern gydnc commands
gydnc list --json                              # Overview with full metadata
gydnc get recipes/blog/post-creation          # Retrieve specific guidance
gydnc create must/safety-first --title "..."  # Create new guidance
gydnc hash <alias>                             # Verify content integrity
```

## The Cursor Rules Connection: Full Circle 🔄

Today's `.cursor/rules/gydnc.mdc` file represents the current state of this evolution. It's a meta-guidance document that teaches AI systems how to interact with the gydnc system itself:

```markdown
### 1. Guidance Retrieval Workflow
ALWAYS follow this sequence to ensure you have comprehensive guidance:

1. **BEGIN WITH OVERVIEW:** Start EVERY session by getting a complete overview:
    ```bash
    gydnc list --json
    ```

2. **FETCH DETAILED GUIDANCE:** After identifying relevant guidance:
    ```bash
    gydnc get <entity1> <entity2> <entity3>
    ```
```

This represents the full circle: we started with simple file discovery, evolved through structured tools, and now have guidance *about* guidance – meta-instructions that ensure AI systems can effectively use the guidance system we've built.

## Key Lessons from the Evolution 📚

**1. Simplicity Has Its Place**
The `find` command never went away because sometimes you just need to quickly locate files. Complex systems should complement, not replace, simple tools.

**2. Structure Enables Scale**
As guidance volume grew, the need for categorization, tagging, and filtering became essential. The move from ad-hoc files to structured `.bhv` and `.rcp` formats was crucial.

**3. Content-Addressability Solves Real Problems**
Immutable versions with cryptographic hashes solve the "which version of the guidance was I following?" problem that plagued earlier systems.

**4. Meta-Guidance Is Critical**
The `.cursor/rules/gydnc.mdc` file teaching AI systems how to use gydnc represents a crucial insight: guidance systems need guidance about themselves to be truly effective.

**5. Evolution, Not Revolution**
Each step built on the previous one. Even `gydnc` retains concepts like hierarchical aliases (`must/`, `should/`, `recipes/`) that trace back to the early categorization work in `agentt`.

## Looking Forward: The gydnc Ecosystem 🌟

With `gydnc` now at [github.com/ofthemachine/gydnc](https://github.com/ofthemachine/gydnc), we're seeing the maturation of guidance systems. The roadmap includes:

- **Symbolic version tagging** (`stable`, `beta` channels)
- **Guidance composition** (one guidance file referencing parts of another)
- **Distributed backends** (S3, databases, ledger systems)
- **Rich provenance visualization** (understanding how guidance evolved)

The `.cursor` directory's evolution from simple discovery scripts to sophisticated guidance systems mirrors a broader trend in AI tooling: starting simple, learning from usage patterns, and gradually building more sophisticated abstractions that preserve the power of the simple tools while enabling new capabilities.

From `find . -name "*.yml"` to content-addressable guidance with immutable provenance – it's been quite a journey. And with gydnc's spin-off, this evolution continues in earnest, ready to tackle the next set of challenges in AI guidance management.

---