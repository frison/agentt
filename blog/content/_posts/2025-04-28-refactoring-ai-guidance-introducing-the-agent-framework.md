---
layout: post
title: Refactoring Our AI Guidance - Introducing the .agent Framework 🤖
date: 2025-04-28 01:12:49 -0600 # Derived from system time & formatted to CST/CDT per Time Awareness Directive
categories:
  - ai
  - meta
  - refactoring
  - agent-framework
  - development-process
---

As our project evolves, so does our collaboration with AI assistants. We recently undertook a significant refactoring of how we provide guidance to these non-human team members. Our previous system, a mix of `.nhi` files and `.cursor` rules scattered about, became increasingly confusing. As one developer noted, *"I'm thinking as a human, looking in my IDE, without a clear boundary around what's AI meta/instructions..."* We needed a change, prioritizing **"clarity over churn"**. This post details the journey to our new, unified `.agent` framework!

## Why Refactor? 🤔

Several key motivations drove this overhaul:

1.  **Clarity & Consistency:** The primary goal was a clear, understandable, and consistent structure for both humans and AI. The old mix of formats and locations created ambiguity.
2.  **Human & AI Readability:** We needed a structure intuitive for humans browsing the repository, visually grouping AI-related configuration together.
3.  **Agent-Agnostic Definitions:** We wanted to define core guidance domains (like behavior rules and procedural recipes) in a way that wasn't tied *only* to Cursor, potentially allowing other tools or agents to leverage them in the future. The user noted, *"keeping instructions for these domains outside of the .cursor directory is preferable. This also makes it more clear their intention may be used for other agents..."*
4.  **Robustness & Maintainability:** The framework required reliable discovery mechanisms and stable ways to link related guidance (e.g., between behaviors and recipes).

## Introducing the `.agent` Directory Structure ✨

To address these points, we consolidated project-specific AI guidance under a new top-level `.agent/` directory. This immediately provides that clear boundary we were looking for.

```
.agent/
├── README.md         # Overview of the .agent directory
├── behavior/         # Defines MUST/SHOULD behavioral rules
│   ├── README.md     # Behavior domain definition
│   ├── FORMAT.md     # .bhv file format spec
│   ├── must/         # MUST rule files (.bhv)
│   ├── should/       # SHOULD rule files (.bhv)
│   └── bin/
│       └── discover.sh # Behavior discovery script
└── cookbook/         # Defines procedural recipes
    ├── README.md     # Cookbook domain definition
    ├── meta/         # Recipe category example
    ├── git/          # Recipe category example
    ├── blog.frison.ca/ # Recipe category example
    ├── *.rcp         # Recipe files
    └── bin/
        └── discover.sh # Cookbook discovery script
```

This structure separates concerns logically into distinct domains.

## The Behavior Domain (.bhv) 🚦

Formerly known as `.nhi`, the `behavior` domain defines the high-level rules governing *how* work should be done.

*   **Purpose:** Contains core principles, constraints, and recommended practices.
*   **Tiers:** Separated into `must/` (absolute requirements) and `should/` (recommendations), mirroring RFC 2119 concepts.
*   **Format:** Uses the `.bhv` file extension.
*   **Discovery:** A dedicated script (`.agent/behavior/bin/discover.sh`) allows reliable discovery of rules.

This ensures foundational principles and core practices are clearly defined and easily discoverable.

## The Cookbook Domain (.rcp) 🍳

While behavior defines *how* to operate, the `cookbook` domain provides step-by-step procedural recipes for *accomplishing* specific tasks.

*   **Purpose:** Offers clear, repeatable instructions for common workflows (like creating this blog post!).
*   **Format:** Uses the `.rcp` file extension with a standardized V2 YAML frontmatter. This includes a stable `id` field, crucial for robustly linking recipes to other guidance or even other recipes. (User: *"Ohhh interesting, the id's could be references in other rules -- and that would work well for you?"* AI: Yes, crucial for stable linking. User: *"Alright, sold!"*)
*   **Discovery:** Also features a dedicated script (`.agent/cookbook/bin/discover.sh`) using `awk` and `yq` for reliable discovery.

This provides actionable, tested procedures for recurring tasks.

## Centralized Enforcement 👮‍♀️

How do we ensure this new framework is consistently used? We consolidated the enforcement logic into a single, primary Cursor rule: `.cursor/rules/agent-interaction-framework.mdc`.

This rule mandates a strict interaction sequence for the AI:
1.  Consult the framework overview (`.agent/README.md`).
2.  Check **Behavior** rules (MUST then SHOULD) using the discovery script.
3.  Check **Cookbook** for relevant recipes using its discovery script.
4.  Only *then* proceed with the task, guided by the discovered directives and recipes.

This ensures that all actions taken by the AI align with our defined behaviors and leverage established procedures when available.

## Looking Forward 🚀

This refactoring provides a much clearer, more consistent, robust, and maintainable framework for guiding AI collaboration. By separating behavioral rules from procedural recipes and consolidating them under `.agent`, we've improved the developer experience and laid a stronger foundation for future integrations. We believe this structured approach will lead to more predictable, reliable, and effective AI assistance as we continue to build. Who knows, maybe a `reflex` domain for executable agent capabilities is next? 😉