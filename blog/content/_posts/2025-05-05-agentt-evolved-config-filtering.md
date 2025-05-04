---
layout: post
title: "Agentt Evolved: Multi-Source Guidance, Precision Filtering, and AI Efficiency 🚀"
date: 2025-05-05 10:30:00 -0600 # Placeholder time, adjust as needed
categories:
  - agentt
  - ai-development
  - configuration
  - cli
draft: true # Start as draft
provenance:
  repo: "https://github.com/your_repo/agentt" # Adjust repo URL if needed
  commit: "1239727da589afe0981320fd7e4d58bfa89c80aa"
  prompt: "1239727da589afe0981320fd7e4d58bfa89c80aa" # Often the same as commit
  modifications: []
---

As AI assistants become more integrated into our development workflows, managing the guidance that shapes their behavior becomes increasingly critical. Simple lists of rules or instructions quickly become unwieldy. The Agentt project focuses on providing robust tooling for this exact challenge, and we're excited to roll out some significant enhancements designed to make guidance management more flexible, powerful, and efficient – especially for complex team and enterprise environments (and frankly, to make your AI assistant's life easier 😉).

## The Growing Pains of Guidance

Initially, having a single place for agent behaviors and procedural recipes works well. But what happens when:

*   Your organization wants baseline safety rules applied *everywhere*?
*   Your platform team has specific infrastructure conventions?
*   Your application team needs project-specific coding recipes?
*   Your AI assistant needs to find *just* the relevant security rules without wading through UI guidelines?

Fetching *everything* all the time becomes inefficient and noisy. Fetching only by specific ID requires knowing exactly what you need beforehand. We needed something better.

## Layer Up: Multi-Backend Configuration 🍰

To address the need for composable guidance, Agentt now boasts a revamped configuration structure that supports **multiple guidance backends**.

Instead of pointing to a single directory, your `config.yaml` now defines a list of `backends`, each specifying a source for behaviors and recipes. The most common type is `localfs`, which points to a directory on your local filesystem.

```yaml
# .agent/service/config.yaml (Example)
entityTypes:
  # ... (behavior, recipe definitions) ...

backends:
  - type: localfs
    # Org-level guidance (read-only, perhaps?)
    rootDir: "../../../shared/company-guidance" # Relative to config file
    entityLocations:
      behavior: "must/security/*.bhv"
      # Org recipes...

  - type: localfs
    # Team-level conventions
    rootDir: "../team-conventions" # Relative to config file
    entityLocations:
      behavior: "should/api/*.bhv"
      recipe: "deployment/*.rcp"

  - type: localfs
    # Project-specific guidance (in this repo)
    rootDir: "." # Relative to config file (i.e., .agent/service)
    entityLocations:
      behavior: ".agent/behaviors/**/*.bhv" # Local overrides/additions
      recipe: ".agent/recipes/**/*.rcp"
```

**Why is this awesome?**

*   **Layering:** Define base organizational rules (security, compliance) in one shared source, layer team conventions on top, and keep project-specific recipes close to the code they apply to. Agentt loads them all, warning about ID conflicts but making everything available.
*   **Portability:** The `rootDir` for each `localfs` backend is relative to the location of the `config.yaml` file itself, making configurations easier to share and manage across different environments.
*   **Scalability:** Easily manage guidance across large organizations with diverse needs.

## Cut Through the Noise: Advanced Filtering with `--filter` ✂️

Loading guidance from multiple sources is great, but fetching *everything* via `agentt summary` can be costly, especially when you only need a specific subset. In a moderately sized project, the output of `agentt summary` might be hundreds or even thousands of tokens!

Enter the new `--filter` flag, available on the `summary`, `details`, and the brand-new `ids` commands. This lets you move beyond simple ID lookups and perform powerful queries directly against guidance metadata, retrieving only what's relevant.

**Syntax Highlights:**

*   **Key-Value:** `tier:must`, `type:recipe`
*   **Negation (Prefix):** `-tag:scope:legacy` (exclude legacy items)
*   **Existence Check:** `priority:*` (find items where priority *is set*)
*   **Keywords:** `NOT tag:deprecated`, `tier:should AND tag:scope:core` (AND is also implicit between terms)
*   **Tag Wildcards (`*`):** Match patterns within tags!
    *   `tag:tech:*` (Prefix: any tech tag)
    *   `tag:*:git` (Suffix: any git tag)
    *   `tag:*log*` (Substring: tags containing "log")

**Examples:**

```bash
# Show summaries of all high-priority MUST behaviors
agentt summary --filter "tier:must priority:*"

# Get details for core recipes, excluding anything tagged 'experimental'
agentt details --filter "type:recipe tag:scope:core NOT tag:experimental"

# Find behaviors related to 'safety' or 'security' using wildcards
agentt summary --filter "type:behavior tag:*safe* OR tag:*security*" # (OR coming soon!)
# For now, you might use:
agentt summary --filter "type:behavior tag:*safe*"
agentt summary --filter "type:behavior tag:*security*"

# Find non-recipe guidance tagged 'meta'
agentt summary --filter "tag:scope:meta -type:recipe"
```

**Why does this help *me* (your friendly AI assistant)?** Instead of fetching and parsing nearly 1000 tokens (based on current repo stats) from `agentt summary` just to find the specific rules you asked for ("Use MUST safety rules"), I can construct a precise filter query like `--filter "tier:must quality:safety"`. This directly gives me the relevant summaries (if using `summary`) or allows me to proceed efficiently with `details` or `ids`.

This filtering capability is also key for **iterative context enrichment**. As our conversation evolves (e.g., we start discussing Git commits), I can use a targeted filter (`--filter "tag:tech:git type:recipe"`) to fetch *only* the newly relevant Git recipes and add them to our working context without flooding it with unrelated information.

## Peak Efficiency: `agentt ids --filter "..."` ✨

While `--filter` on `summary` and `details` reduces noise, sometimes all I need is the *list of IDs* matching specific criteria. If you ask me to "Apply all MUST-tier guidance related to quality:safety", I don't need the full summaries first, just the list of IDs to fetch details for.

Fetching full summaries (~971 tokens in our example) just to extract two IDs (`safety-first`, `shell-safety`) felt... wasteful (think of the poor tokens!). So, we introduced the `ids` command:

```bash
# Get just the IDs for MUST-tier guidance tagged 'quality:safety'
agentt ids --filter "tier:must quality:safety"

# Output: (approx. 48 tokens)
# ["safety-first", "shell-safety"]
```

This command performs the same filtering but outputs *only* a clean JSON array of matching IDs, costing only ~48 tokens in this example.

**The Big Win (Token Math!):**

*   **Old Way:** `summary` (971 tokens) + `details --id id1 --id id2` (530 tokens) = **~1501 tokens** (plus parsing)
*   **New Way:** `ids --filter "..."` (48 tokens) + `details --id id1 --id id2` (530 tokens) = **~578 tokens**

That's roughly a **60% reduction** in token usage for this common workflow! This targeted optimization makes AI interaction significantly leaner, faster, and cheaper.

## Bringing It All Together

These enhancements represent a major step forward for Agentt:

*   **Scalable Guidance:** Manage rules across your org with multi-source backends.
*   **Precision Control:** Find exactly what you need with advanced filtering.
*   **AI/Automation Efficiency:** Reduce overhead and token usage with targeted queries and the `ids` command.
*   **Improved Maintainability:** A cleaner, more portable configuration structure.

We think these features will significantly improve the management and application of agent guidance. Give them a try and let us know what you think! What other filtering capabilities would make your life easier?

---

*This article was originally created in commit [`1239727da589afe0981320fd7e4d58bfa89c80aa`](https://github.com/frison/agentt/commit/1239727da589afe0981320fd7e4d58bfa89c80aa), prompted by commit [`1239727da589afe0981320fd7e4d58bfa89c80aa`](https://github.com/frison/agentt/commit/1239727da589afe0981320fd7e4d58bfa89c80aa).*