---
layout: post
title: "Fraglets: Portable Code Execution for AI Agents 🧩"
date: 2025-12-09 00:48:26 -0600
categories:
  - ai-tooling
  - mcp
  - code-execution
  - containers
provenance:
  repo: "https://github.com/frison/agentt"
  commit: "a378a32b255eec436eb7ff99992012bb215a9029"
  prompt: "a378a32b255eec436eb7ff99992012bb215a9029"
  modifications: []
---

There's a quiet problem in AI tooling: agents need to execute code, but sandboxing is hard. You either trust arbitrary execution (terrifying), build bespoke sandboxes per language (exhausting), or give up and let the model hallucinate computations (unreliable).

What if someone had already built containers for 100 programming languages, each with a consistent execution interface? And what if you could inject code fragments into them, execute, and return results — all exposed as MCP tools?

That's fraglets. This is how they work, what they're good for, and where they fall short.

## The Foundation: 100hellos

Before fraglets, there's [100hellos](https://github.com/ofthemachine/100hellos) — a collection of Docker containers, each configured to run "Hello World" in a different programming language. The project currently covers 99 languages, from mainstream (Python, JavaScript, Rust) to esoteric (Brainfuck, LOLCODE, Emojicode) to historical (COBOL, FORTRAN, ALGOL).

Each container follows a consistent pattern:
- Alpine-based images (small footprint)
- Pre-configured toolchains
- A working hello-world example
- Reproducible builds

The containers are meant for exploration — "what does Prolog look like?" or "can I actually compile ATS?" — but they also serve as a foundation for something more interesting.

## Enter Fraglets

A **fraglet** is a code fragment that gets injected into a pre-configured container and executed. The term is intentionally modest: these aren't full programs, they're fragments — the smallest meaningful unit of executable code.

The injection model is simple. Each fraglet-enabled container has:

```
/fraglet-entrypoint       # Binary that handles injection
/fraglet.yml              # Configuration: where to inject, how to execute
/guide.md                 # Language-specific authoring hints
/hello-world/hello-world.* # Source file with injection marker
```

The `fraglet.yml` for Python:

```yaml
fragletTempPath: /FRAGLET
injection:
  codePath: /hello-world/hello-world.py
  match: Hello World!
guide: /guide.md
execution:
  path: /hello-world/hello-world.py
  makeExecutable: true
```

When you run a fraglet, the entrypoint:
1. Reads your code from `/FRAGLET`
2. Finds the line containing "Hello World!" in the target file
3. Replaces that line with your code
4. Executes the modified file

Simple line replacement. For languages that need block structures, there's region-based injection with `match_start` and `match_end` markers.

## The CLI: fragletc

The `fragletc` CLI executes fraglets directly:

```bash
$ echo 'print("Hello from fraglet!")' | fragletc --envelope python
Hello from fraglet!

$ echo 'console.log("JavaScript fraglet!")' | fragletc --envelope javascript
JavaScript fraglet!

$ fragletc --envelope lolcode <<< 'VISIBLE "HAI FROM LOLCODE!"'
HAI FROM LOLCODE!
```

Yes, you can execute Brainfuck:

```bash
$ fragletc --envelope brainfuck <<< '++++++++[>++++[>++>+++>+++>+<<<<-]>+>+>->>+[<]<-]>>++.>---.+++++++..+++.>>.<-.<.+++.------.--------.>>+.>++.'
Jello World!
```

(That's supposed to say "Hello World!" — classic Brainfuck off-by-one. The system faithfully executes whatever nonsense you give it.)

The CLI supports both envelope-based execution (using embedded configs) and direct image targeting. Envelopes are just YAML files that map language names to container images and fraglet paths.

## The MCP Server: AI Integration

Here's where it gets practical. The fraglet MCP server exposes two tools:

- **`run`** — execute code in a specified language
- **`language_help`** — get authoring guidance for a language

The `language_help` tool is particularly clever. Each container ships its own `guide.md`, and the MCP server retrieves it so agents know how to write valid code. For Prolog, it explains that fragments should be goals (not directives) and that you need `halt.` at the end. For SNOBOL4, it notes the 8-space indentation requirement.

Self-describing execution environments. Agents can query for constraints before writing code.

Real MCP execution results:

| Language | Code | Output | Duration |
|----------|------|--------|----------|
| Python | `print("Hello!")` | `Hello!` | 443ms |
| JavaScript | `console.log("JS!")` | `JS!` | 487ms |
| Ruby | `puts (1..5).sum` | `15` | 868ms |
| TypeScript | `console.log("TS!")` | `TS!` | 3.8s |
| Julia | `println(2^10)` | `1024` | 1.0s |
| Common Lisp | `(format t "~a" (+ 1 2 3 4 5))` | `15` | 477ms |
| Lua | `print(5050)` | `5050` | 618ms |
| Prolog | `write("Logic!"), nl, halt.` | `Logic!` | 652ms |
| Octave | `printf("pi=%.4f", pi)` | `pi=3.1416` | 732ms |
| LOLCODE | `VISIBLE "O HAI!"` | `O HAI!` | 532ms |
| ArnoldC | `TALK TO THE HAND "Back"` | `Back` | 952ms |
| Emojicode | `😀 🔤Hi🔤❗️` | `Hi` | 580ms |

From scientific computing (Julia, Octave, R) to esoteric languages (Brainfuck, LOLCODE, Emojicode) — same interface, same execution model, wildly different runtimes.

## Computational Reasoning: The Interesting Part

Here's where fraglets transcend "run code in a sandbox." Some containers aren't just language runtimes — they're **domain-specific reasoning environments**.

The Java container ships with a word-puzzle toolkit:

```java
WordSet<?> words = HelloWorld.loadWords();

// Find 5-letter words with 'a' at position 2, containing 'r' (not at 0)
words.matching("_____")
     .withCharAt('a', 2)
     .containing("r").withoutCharAt('r', 0)
     .notContaining("eiou")
     .iterator().forEachRemaining(w ->
         System.out.println(w + " (" + HelloWorld.wordScore(w) + ")"));
```

Output includes `crazy (19)`, `quark (18)`, `ozark (18)` — candidates ranked by Scrabble score.

The fluent API supports:

| Method | Purpose |
|--------|---------|
| `.matching("c_t")` | Pattern with wildcards |
| `.containing("s")` | Has substring |
| `.notContaining("xyz")` | Excludes characters |
| `.withCharAt('a', 2)` | Letter at position |
| `.withoutCharAt('a', 2)` | Letter NOT at position |
| `.count()` | Number of matches |

Plus Wordle semantics baked in:
- **Gray** → `.notContaining("x")`
- **Green** → `.withCharAt('a', position)`
- **Yellow** → `.containing("l").withoutCharAt('l', wrongPos)`

A real Wordle solve:

```java
// After CRANE: Gray c,n,e | Yellow r,a
words.matching("_____")
     .notContaining("cne")
     .containing("r").withoutCharAt('r', 1)
     .containing("a").withoutCharAt('a', 2)
     .iterator().forEachRemaining(System.out::println);
```

This is **computational reasoning**: the agent isn't juggling constraints in-context, it's delegating to a purpose-built environment. The container has:

- Pre-loaded dictionary (no API calls, no hallucination risk)
- Fluent constraint API (composable, type-safe)
- Scrabble scoring (domain knowledge embedded)

The agent writes a query, the container computes the answer. Deterministic, reproducible, verifiable.

## Critical Assessment

### What Works

1. **The injection model is elegant.** Line replacement and region-based injection handle most cases without requiring language-specific AST manipulation.

2. **Container isolation is real.** Reproducible environments, no dependency hell, no "works on my machine."

3. **Language breadth is impressive.** 19 languages with MCP envelopes today, 99 in the underlying 100hellos project. Adding more is mechanical.

4. **Self-documenting environments.** The `language_help` → `run` loop means agents can learn constraints before executing.

5. **Computational reasoning environments are the real insight.** Containers can be more than sandboxes — they can be domain-specific problem-solving tools.

### What Doesn't (Yet)

1. **Cold start latency.** 400-500ms per execution is noticeable. TypeScript's 3.8s compile overhead is painful. Container orchestration isn't free.

2. **No streaming output.** Long computations return all-at-once. Interactive use cases suffer.

3. **Coverage is incomplete.** Not every 100hellos language has fraglet support. Each addition requires configuration and a guide.

4. **The reasoning environments are hand-crafted.** The Java word-puzzle toolkit exists because someone built it. There's no automatic way to create new reasoning environments.

## What's Next: The Uncomfortable Question

The roadmap hints at something more ambitious:

> **Fraglet Templating** — Parameterized fraglets become portable, distributable functions with verifiable provenance.

> **Fraglet Registry** — Captured in a registry, fraglets become a scaling point for extensions.

Think about what happens when you combine templated fraglets with computational reasoning environments:

- A reasoning environment for financial calculations, with market data pre-loaded
- A reasoning environment for chemical simulations, with reaction databases embedded
- A reasoning environment for legal analysis, with case law indexed

Each one: deterministic, reproducible, queryable via MCP. Agents delegate domain-specific computation to purpose-built tools, getting back verifiable results.

Is this the right architecture? Unclear. The container overhead is real. The hand-crafted nature of reasoning environments limits scalability. The injection model works but feels like a clever hack rather than a principled design.

But the core insight — that AI agents need more than "run arbitrary code," they need **domain-specific computational tools** — seems durable. Whether fraglets are the right implementation or just an interesting prototype remains to be seen.

## Try It

The fraglet system lives at [github.com/ofthemachine/fraglet](https://github.com/ofthemachine/fraglet). The underlying containers are at [github.com/ofthemachine/100hellos](https://github.com/ofthemachine/100hellos).

For MCP integration, configure your client to use the fraglet-mcp server. The `run` and `language_help` tools should appear automatically.

Fair warning: this is active development. Sharp edges exist. But if you're thinking about code execution for AI agents, it's worth a look.

> __Human note__: Thanks AI for the "Worth a look" -- but if you want to see AI write emojicode, or use python to generate brainfuck code... then this may be your jam. It's weird.

---

*This post was written with assistance from Claude, using the fraglet MCP server to execute examples. The recursive nature of using AI tooling to write about AI tooling is not lost on me.*

---

**Provenance:**
- Repository: [https://github.com/frison/agentt](https://github.com/frison/agentt)
- Commit: `a378a32b255eec436eb7ff99992012bb215a9029`
