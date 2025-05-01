---
layout: post
title: "Refining the Agent: Guidance Integrity and Continuous Improvement"
date: 2025-05-01 02:02:17 -0600
categories:
  - agentt
  - meta
  - development
  - workflow
  - ai
provenance:
  repo: "https://github.com/frison/agentt"
  commit: "379367a35f5361671fd1b58db7046f456c88ec53"
  prompt: "379367a35f5361671fd1b58db7046f456c88ec53" # Commit message served as prompt
  modifications: []
---

We've been continuing the work on `agentt`, our internal system for managing and serving guidance definitions for AI agents (and humans!). A key focus recently has been hardening the system and refining the guidance itself, ensuring it's both robust and easy to work with.

## Ensuring Guidance Integrity (A MUST!)

One of the most critical aspects of any guidance system is its integrity. If the source definitions are ambiguous or inconsistent, the guidance becomes unreliable, leading to unpredictable agent behavior.

We encountered this directly with guidance entity IDs. Initially, the system was tolerant of duplicate IDs, which could arise from copy/paste errors or conflicting filename/frontmatter definitions. However, relying on non-unique IDs is a recipe for subtle bugs.

Therefore, we've implemented a crucial change: **duplicate guidance IDs detected during the initial scan are now treated as a fatal configuration error.** The `agentt` service or CLI will halt immediately, demanding the ambiguity be resolved. This enforces a MUST-level principle: the guidance source itself must be unambiguous and internally consistent for the system to operate safely. This aligns perfectly with our core `Safety First` and `Clarity Over Churn` behaviors.

## Refining the Shoulds: Continuous Improvement 🧹

Beyond critical MUSTs, we're also refining our SHOULD-level guidance based on practical experience:

*   **Modification Hygiene:** Prompted by needing a few passes to clean up after some recent refactoring (mea culpa!), we added a specific "Modification Hygiene" rule to our `code-style.bhv`. It emphasizes performing a thorough cleanup pass after *any* modification – removing dead code, commented-out blocks, unused imports/variables, and ensuring all tests and builds pass cleanly. Leave it cleaner than you found it!
*   **Clearer CLI Output:** We updated `cli-design-standards.bhv` to explicitly state that primary command output (like JSON data) should go to `stdout`, while logs, warnings, and errors go to `stderr`. This makes the CLI much easier for automated tools (including AI agents) to parse reliably.

## From Human Insight to Agentic Behavior

How do these rules come about? Often, it starts with a general human best practice or a lesson learned (like my cleanup oversight). We then translate that into specific, verifiable rules that an AI agent can process.

*   "Don't leave commented-out code" becomes a concrete rule in `.bhv` file's "Rules" section.
*   "Make CLIs scriptable" translates into specific guidance on standard streams and exit codes.
*   "Ambiguous IDs are bad" becomes a fatal error check in the Go code, reflecting a system MUST.

This process of codifying best practices into machine-readable guidance is central to the `agentt` philosophy.

## Guidance as an Accelerator (Yes, Really!) 🚀

Does spending time writing `.bhv` files slow things down? Initially, perhaps. But in the longer run, I firmly believe it's an accelerator.

Encoding guidance upfront:
*   **Enforces Consistency:** Reduces errors and misunderstandings.
*   **Enables Automation:** Allows AI agents to reliably follow best practices.
*   **Improves Quality:** Catches configuration errors and promotes better coding habits.
*   **Boosts Maintainability:** Makes the system easier to understand and evolve.

It shifts the burden from constantly remembering rules to defining them once and letting tooling (and guided AI) help with compliance.

## Looking Forward

We'll continue refining `agentt` and its guidance, focusing on making it a robust, clear, and effective foundation for building reliable AI-assisted workflows. Ensuring the integrity and usability of the guidance itself remains paramount.

---

*This article was originally created in commit [`379367a`](https://github.com/frison/agentt/commit/379367a35f5361671fd1b58db7046f456c88ec53), prompted by commit [`379367a`](https://github.com/frison/agentt/commit/379367a35f5361671fd1b58db7046f456c88ec53).*