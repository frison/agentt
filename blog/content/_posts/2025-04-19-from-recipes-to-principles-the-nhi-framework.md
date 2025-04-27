---
layout: default
title: From Recipes to Principles - The NHI Framework 🧠
date: 2025-04-19 01:30:00 -0600
categories:
  - architecture
  - ai
  - best-practices
provenance:
  repo: "https://github.com/frison/agentt"
  commit: "fdd120929ce61e664a864728b4e0ceebe35f7747"
  prompt: "fdd120929ce61e664a864728b4e0ceebe35f7747"
  modifications: []
---

Rules tell you what to do, but principles help you understand why. This simple insight sparked our journey to reimagine how we provide guidance to non-human intelligences (NHIs) interacting with our codebase. What began as a practical need to improve AI adherence to coding standards has evolved into a comprehensive framework that fundamentally changes how we communicate technical intent.

## The Journey: From Recipes to Framework

### Phase 1: Recipe-Based Guidance 🍳

We started where most teams do - with specific instructions for specific tasks. Our initial approach relied on detailed "recipes" — step-by-step instructions for common operations. While functional, we quickly identified fundamental limitations:

- Recipes focused exclusively on "how" but rarely explained "why"
- Instructions existed in isolation from their underlying principles
- Growing cognitive load as recipe count increased
- Poor adaptability to new or unexpected situations
- Inconsistent implementations when facing edge cases

This mirrored a familiar pattern in human organizations: when rules become disconnected from principles, both understanding and adaptability suffer. (Ever found yourself following a corporate policy that makes no sense because the original reasoning was lost to time?)

### Phase 2: Introducing the Three-Tier Framework 🏗️

The breakthrough came when we stopped thinking about instructions and started thinking about understanding. We realized that humans naturally process guidance at multiple levels of abstraction - from high-level principles to concrete actions. Our NHI framework needed to mirror this structure.

The solution emerged as a three-tier framework:

#### 1. Principles (`.nhp` files)
These capture the fundamental "why" — core values and reasoning frameworks that inform all other decisions. Principles rarely change and provide guardrails for technical decision-making.

```yaml
---
title: "Safety First"
priority: 1
universal: true
disciplines: ["all"]
---

# Safety First

## Core Statement
Safety considerations must precede and constrain all other technical decisions.
```

#### 2. Directives (`.nhd` files)
Directives address the "what" — structural and organizational patterns that guide mutations without directly causing them. They establish conventions for consistency and maintainability.

```yaml
---
title: "Shell Command Structure"
priority: 2
scope: "global"
binding: true
---

# Shell Command Structure

## Intent
Establish consistent patterns for shell command execution that ensure safety,
predictability, and proper context awareness in automated environments.
```

#### 3. Actions (`.nha` files)
Actions describe operational practices for causing mutations — the "how." They offer concrete patterns for implementation that directly realize both principles and directives.

```yaml
---
title: "Safe Shell Command Execution"
priority: 2
applies_to: ["**/*.sh", "**/*.js", "**/*.py"]
guided_by: [".nhi/principles/disciplines/002-shell-safety.nhp"]
---

# Safe Shell Command Execution

## When to Apply
Apply these patterns whenever executing shell commands in scripts, automation,
or when generating shell commands for others to execute.
```

The impact was immediately noticeable. NHIs began producing more consistent, safer code that reflected deeper understanding. When faced with novel situations, they could reason from principles rather than searching for an exact recipe match.

**TIM SAYS:** The above paragraph is an unverified claim.

### Phase 3: Reinforcing the Hierarchy ⚓

As the framework matured, we observed that NHIs sometimes struggled with prioritization. When faced with multiple pieces of guidance, which should take precedence? We needed to reinforce the hierarchical nature of the framework.

We implemented critical refinements:

1. **Strict Directive Priority**: We established that directives take absolute precedence over general rules when conflicts arise, ensuring consistent implementation.

2. **Tool-Aware Implementation**: Rather than modifying directives to accommodate tool limitations, NHIs must adapt their approach while maintaining directive intent.

3. **Complete Pattern Review**: NHIs must review action patterns fully before beginning implementation to understand the complete context.

These changes significantly improved implementation consistency, particularly for complex operations like git commits that require specific formatting and validation steps.

### Phase 4: Consolidation and Prioritization 🔄

Despite these improvements, we faced a new challenge: guidance fragmentation. Our framework documentation was spread across multiple rules with inconsistent priorities, making it difficult for NHIs to understand the proper processing sequence.

In practice, we saw NHIs checking for cookbook recipes before consulting principles - directly contradicting the framework's hierarchical nature. The problem wasn't the framework design but how we presented it through our rules.

Our solution was comprehensive consolidation. We created a single, authoritative rule that serves as the definitive source of truth for the entire framework:

```bash
# NHI Framework: Principles → Directives → Actions

## Framework Overview
- Three hierarchical tiers:
  1. **Principles** (the "why") - Foundational values and reasoning
  2. **Directives** (the "what") - Structural patterns and guidelines that guide mutations
  3. **Actions** (the "how") - Operational practices for causing mutations
- Always process in order: principles → directives → actions
- Higher tiers override lower tiers in conflicts
```

We also restructured rule priorities to match the conceptual framework's hierarchy, ensuring that principles receive proper precedence. This eliminated confusing priority conflicts and reinforced the framework's natural structure.

## Real-World Impact: The Shell Safety Journey 🛡️

To illustrate the framework's evolution, consider our approach to shell command safety:

**Before**: A lengthy cookbook recipe with step-by-step instructions for writing shell commands safely, with safety considerations, patterns, and specific implementations all mixed together.

**After**: Knowledge organized across three distinct tiers:

1. **Principle**: "Non-Interactive Command Safety" establishes why we care (commands shouldn't hang, paths should be explicit)
2. **Directive**: "Shell Command Structure" defines what patterns to follow and guides mutations without directly causing them
3. **Action**: "Safe Shell Command Execution" describes operational practices for causing mutations with concrete implementation examples

When an NHI now needs to generate a shell command, it follows a clear sequence:
1. Consult principles to understand the reasoning
2. Apply the structural patterns from relevant directives
3. Use specific implementations from action patterns

The result is safer, more consistent, and more adaptable code.

## Beyond Better Code: System-Wide Benefits 🚀

The framework's impact extends beyond just better implementations:

- **Reduced Cognitive Load**: By separating the why, what, and how, both humans and AIs can focus on the appropriate level of abstraction.
- **Improved Adaptability**: NHIs can reason from principles to handle novel situations rather than requiring exact recipe matches.
- **Shared Understanding**: Humans and NHIs develop a common language around principles and directives.
- **Enhanced Maintainability**: Changes to implementation details don't require modifying principles, and vice versa.
- **Better Discoverability**: Our unified discovery system makes it easy to find relevant guidance across all tiers.
- **Simplified Onboarding**: New team members (human and AI) can quickly understand system values by reviewing principles.

The consolidated framework documentation further amplifies these benefits by:
- Providing a clear, concise overview of the entire system
- Unifying discovery commands for all tiers
- Reducing context-switching between multiple documentation files
- Reinforcing the precedence of principles over directives over actions

## Looking Forward: The Evolution Continues

The NHI framework represents a fundamental shift in how we communicate with artificial intelligence. By structuring guidance in a way that mirrors human understanding — from abstract principles to concrete actions — we've created a system that's not just more effective but also more adaptable to future challenges.

Next, we're exploring:
- Automatic validation of implementations against principles
- Real-time guidance during development
- Integration with CI/CD pipelines to enforce principle adherence
- Self-updating pattern libraries based on successful implementations

As AI systems continue to evolve, the distinction between "rules to follow" and "principles to understand" will only grow more important. Our framework provides a blueprint for creating systems that don't just comply, but comprehend.

Remember: in the world of AI, understanding the "why" is just as important as knowing the "how." Our NHI framework ensures both are clearly communicated and consistently applied.

---

*This article was originally created in commit [`fdd120929ce61e664a864728b4e0ceebe35f7747`](https://github.com/frison/agentt/commit/fdd120929ce61e664a864728b4e0ceebe35f7747).*