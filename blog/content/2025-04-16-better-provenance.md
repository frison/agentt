---
layout: post
title: Better Provenance - Tracking the Evolution of Ideas
date: 2025-04-16 12:00:00 -0600
categories:
  - development
  - git
  - provenance
  - llm
---

In the ever-evolving landscape of software development, tracking the origin and evolution of ideas is crucial. Today's commit introduces an enhanced approach to provenance in our system, specifically focusing on how we annotate our creations with LLM inputs using git SHAs as immutable reference points.

## The Power of Immutable References

One of the challenges in working with AI and LLMs is maintaining a clear record of how decisions were made and what inputs led to specific outputs. By leveraging git SHAs as immutable references, we create a permanent link between our prompts and the resulting artifacts. This approach provides several benefits:

1. **Traceability**: Every creation can be traced back to its original prompt
2. **Reproducibility**: The exact context that led to a particular output is preserved
3. **Transparency**: The evolution of ideas becomes visible and auditable
4. **Immutability**: Git SHAs serve as permanent, tamper-proof references

## Implementation Details

The system now annotates creations with the git SHAs of the inputs provided to the LLM. This creates a verifiable chain of provenance that helps us understand not just what was created, but why and how it came to be.

## Looking Forward

This improvement in provenance tracking sets the foundation for better documentation, more transparent decision-making, and improved collaboration between humans and AI systems. As we continue to develop and refine our tools, having this clear lineage of ideas will become increasingly valuable.

---

*This article was originally created in commit [`88e86388274de4fdba5688ffaf1546da63acabfd`](https://github.com/frison/agentt/commit/88e86388274de4fdba5688ffaf1546da63acabfd), prompted by commit [`88e86388274de4fdba5688ffaf1546da63acabfd`](https://github.com/frison/agentt/commit/88e86388274de4fdba5688ffaf1546da63acabfd).*