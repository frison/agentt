---
layout: post
title: "When in Rome: Agents MUST Understand the Map Before Moving"
date: 2025-05-01 02:11:50 -0600
categories:
  - agentt
  - meta
  - development
  - workflow
  - ai
  - architecture
  - context
# Provenance block will be added in the next step
---

When we task an AI agent with modifying a complex software project, we're essentially asking it to perform delicate surgery in an unfamiliar operating room, or perhaps more aptly, navigate a bustling, unfamiliar city. Just as a driver needs to understand the map, the one-way streets, and the local traffic patterns *before* pulling into traffic, an agent **MUST** understand the context before making changes.

This isn't just good advice; it's a core safety and effectiveness requirement we encode in our `agentt` guidance system.

## The Critical MUST: Understand Context First

One of our foundational **MUST** behaviors, enshrined in `.agent/behavior/must/understand-context-first.bhv`, mandates that an agent *must* actively seek out and process relevant context before modifying *anything*. This context is the agent's map and rulebook for the specific area it's working in.

What does this map include?
*   **Structural Metadata:** Build files (`Makefile`, `go.mod`), architecture documents (`ARCHITECTURE.md`), IaC definitions.
*   **Code Structure:** Directory layouts, package organization, key file locations.
*   **Existing Patterns:** How related code is written, used, tested, and deployed.
*   **Specific Guidance:** Relevant `.bhv` or `.rcp` files defining constraints or procedures for the task at hand.

Why is this a **MUST**?
Modifying code without context is like driving blindfolded. It risks regressions, inconsistencies, security vulnerabilities, and architectural decay. An agent blindly following instructions without understanding the surrounding environment cannot operate safely or effectively. It *must* consult the map and rules first.

## Refining the Maps and Tools

While the `understand-context-first` behavior is paramount, we also continually refine the tools and other guidance that *support* this:

*   **Clearer Tools (CLI Output):** We improved the `agentt` CLI to separate primary data (`stdout`) from operational logs (`stderr`), making it easier for agents (or humans) to cleanly read the map data (like summaries or details) provided by the tool.
*   **Better Cartography Practices (Modification Hygiene):** We added a "Modification Hygiene" rule to `code-style.bhv` based on our own recent cleanup needs. This ensures that when *we* update the maps (the guidance files or the codebase itself), we do so cleanly – removing commented-out drafts, unused elements, and verifying everything still works. This keeps the maps accurate and reliable for the agents using them.

## Reliable Maps Accelerate the Journey

This focus on context and codified guidance might seem like overhead, but it's truly an accelerator. Providing clear, reliable maps (`ARCHITECTURE.md`, well-structured code, specific `.bhv` and `.rcp` files via `agentt`) allows agents to:
*   **Operate Safely:** By understanding constraints and potential side effects *before* acting.
*   **Act Consistently:** By following established patterns and procedures.
*   **Integrate Correctly:** By understanding how their changes fit into the larger system.
*   **Reduce Errors:** By leveraging codified best practices and avoiding known pitfalls.

Investing in building and maintaining these reliable maps, and enforcing the MUST requirement for agents to consult them, is crucial for enabling safe, effective, and ultimately faster AI-assisted development.

We'll keep refining these maps and ensuring our agents know: when in Rome, understand the map before you move.