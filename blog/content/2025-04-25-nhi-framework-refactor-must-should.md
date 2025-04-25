---
layout: post
title: 🏗️ Refactoring the NHI Framework - MUST We Change? 🤔
date: 2025-04-25 00:16:33 -0600
categories:
  - nhi-framework
  - refactoring
  - documentation
  - meta
  - rfc-2119
provenance:
  repo: "https://github.com/frison/agentt"
  commit: "114884be8bc84dc5f6d13fb746441216bece24b4"
  prompt: "114884be8bc84dc5f6d13fb746441216bece24b4"
  modifications: []
---

Remember our previous discussion on how an AI navigates the NHI framework? We talked about the `Needs` and `Practices` tiers and the `discover.sh` script used to parse them. Well, like any good system, the NHI framework itself isn't static. Based on recent work and a drive for even greater clarity (following our own `clarity-over-churn` principle!), we've made some significant evolutionary changes.

This post dives into the *what* and *why* of these updates, moving from our original Needs/Practices structure to a system more explicitly aligned with established standards.

## From Needs & Practices to MUST & SHOULD 📜

The core change revolves around the terminology and structure of the two main tiers:

*   **`.nhi/needs/` is now `.nhi/must/`**
*   **`.nhi/practices/` is now `.nhi/should/`**

Why the change? While "Needs" and "Practices" served us well initially, we realized we could achieve greater precision and leverage widely understood definitions by aligning with **RFC 2119**. This standard defines keywords for requirement levels:

*   **MUST:** Absolute requirement. Equivalent to our previous "Needs."
*   **SHOULD:** Recommended practice. Exceptions require justification and must not violate a MUST. Equivalent to our previous "Practices."

Adopting MUST/SHOULD makes the framework's intent immediately clearer to anyone familiar with RFC 2119 (a common standard in technical documentation) and reinforces the strict hierarchy: **MUST** requirements always take precedence over **SHOULD** recommendations.

We debated other terms like "Axiom," but concluded that MUST/SHOULD offered the best balance of clarity and established meaning, directly conveying the sense of obligation (MUST) and recommendation (SHOULD) intended for each tier.

## Simplifying Internal Structure 🧹

Alongside the rename, we also removed the `tier:` metadata from within the `.nhi` files themselves. Initially, files contained `tier: need` or `tier: practice`. However, this was redundant; a file's location within `.nhi/must/` or `.nhi/should/` already definitively states its level.

Relying solely on the directory structure simplifies the content of each rule and eliminates the potential for confusing mismatches (e.g., a file in `.nhi/must/` accidentally containing `tier: should`). The directory is now the single source of truth for the requirement level.

## The Benefit of Mobility 🚚

A significant advantage of this directory-based approach is the ease with which requirements can be promoted or demoted. If a `SHOULD` recommendation proves critical enough to become non-negotiable, simply moving its corresponding `.nhi` file from the `.nhi/should/` directory to the `.nhi/must/` directory elevates its status. Conversely, demoting a `MUST` requirement involves moving the file the other way. This physical organization directly reflects the requirement's current weight within the framework, making evolution straightforward.

## Refining Framework Documentation 📄

Naturally, these structural changes necessitated updates to our core framework documentation within `.cursor/rules/`:

*   `010-core-_R__-nhi-framework.mdc` was updated to define the MUST/SHOULD tiers, explicitly reference RFC 2119, and remove the now-obsolete `tier:` metadata discussion.
*   `050-core-_R__-nhi-priority.mdc` was updated to reflect the MUST -> SHOULD checking order consistently.

This ensures our own internal guidance accurately reflects the framework's current state. (Meta!)

## Next Steps & Considerations 🤔

While we haven't implemented it yet, we've also discussed potentially adding structure for negative constraints:

*   **MUST NOT:** Absolute prohibitions.
*   **SHOULD NOT:** Discouraged practices.

One idea is using `not/` subdirectories (e.g., `.nhi/must/not/`, `.nhi/should/not/`). This would allow us to explicitly capture things that are forbidden or discouraged, further aligning with RFC 2119, without adding more top-level directories. This remains an area for future refinement and would require updates to the `discover.sh` script.

We also removed a potentially redundant `framework-overview.nhi` file from within the `.nhi/must/` directory, as its descriptive purpose is now fully served by the `.cursor` rules themselves.

## Looking Forward ✨

Refactoring our own process framework might seem overly meta, but it's crucial. By aligning with RFC 2119 (MUST/SHOULD), simplifying the structure (removing `tier:`, relying on directories), and enabling easy promotion/demotion of rules, we've aimed for increased clarity and robustness. This makes the framework easier to understand, use consistently (by humans and AI alike!), and maintain.

The goal, as always, is a system that effectively guides development while remaining adaptable. Now, let's get this documented with some provenance...

---

*This article was originally created in commit [`114884be8bc84dc5f6d13fb746441216bece24b4`](https://github.com/frison/agentt/commit/114884be8bc84dc5f6d13fb746441216bece24b4), prompted by commit [`114884be8bc84dc5f6d13fb746441216bece24b4`](https://github.com/frison/agentt/commit/114884be8bc84dc5f6d13fb746441216bece24b4).*