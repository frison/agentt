---
layout: post
title: "100 Languages Walk Into a Shebang... 🐚"
date: 2026-01-17 00:30:00 -0600
categories:
  - 100hellos
  - fraglets
  - containers
  - absurdity
---

Have you heard of "Time Until Hello World"? It's a developer experience metric for programming languages. I didn't know about it before I started on this project, but holy shit am I intimately familiar with it now.

Even after 25 years hacking away on various things, I naively thought "A hundred of these should be easy, it's just 'Hello World!'" — and that put me down a multi-year journey of pain, enlightenment, and questionable life choices.

## The Pitch

What if you could run Java like a script? C like a script? Brainfuck like a script?

```java
#!/usr/bin/env -S fragletc --vein=java

public static void main(String[] args) {
    int i = 0;
    for (String arg : args) {
        System.out.println(i++ + ": " + arg);
    }
}
```

Save it. `chmod +x`. Run it. Watch Java compile and execute your "script" in a containerized sandbox. Because we live in the future and the future is weird.

## The Absurdity Scales

99 languages. Each one containerized. Each one shebang-able. From the respectable to the unhinged:

```rust
#!/usr/bin/env -S fragletc --vein=rust
fn main() {
    println!("Rust like a script!");
}
```

```csharp
#!/usr/bin/env -S fragletc --vein=csharp
Console.WriteLine("C# like a script!");
```

```c
#!/usr/bin/env -S fragletc --vein=c
#include <stdio.h>
int main() {
    printf("C like a script!\n");
    return 0;
}
```

```lolcode
#!/usr/bin/env -S fragletc --vein=lolcode
VISIBLE "HAI FROM LOLCODE!"
```

Yes, that LOLCODE example works. Yes, I have tested it. Yes, I question my choices regularly.

## Veins: A Metaphor I Can't Shake

I needed a name for "the injection point where your code gets slotted into a container." I landed on **vein**. As in, the code flows through the vein into the container's execution context.

Is it a perfect metaphor? No. Is it psychologically sticky and slightly unsettling? Absolutely. That's the sweet spot.

Within a single vein, you can have different **modes** — different injection configurations for different use cases:

```java
#!/usr/bin/env -S fragletc --vein=java:wordalytica

WordSet<?> words = HelloWorld.loadWords();
int count = words.endingWith("ing").count();
System.out.println("Words ending with 'ing': " + count);
```

Same container, different injection point. The `wordalytica` mode gives you a fluent word-puzzle API with a pre-loaded dictionary. Because of course it does.

## The Brainfuck Test

Any system claiming to support "100 languages" needs to prove it can handle the weird ones:

```brainfuck
#!/usr/bin/env -S fragletc --vein=brainfuck
++++++++[>++++[>++>+++>+++>+<<<<-]>+>+>->>+[<]<-]>>++.>---.+++++++..+++.>>.<-.<.+++.------.--------.>>+.>++.
```

Output: `Jello World!`

That's supposed to say "Hello World!" — classic Brainfuck off-by-one. The system faithfully executes whatever nonsense you give it. Garbage in, `Jello` out.

## The Tool: `fragletc`

Grab it, run it, question reality:

```bash
# From stdin
echo 'print("Hello!")' | fragletc --vein=python

# From a file
fragletc script.py

# As a shebang (just run the file)
./my-java-script.java arg1 arg2

# Heredoc for the shell wizards
fragletc --vein=lolcode - << 'EOF'
VISIBLE "LOLCODE VIA HEREDOC!"
EOF
```

It handles the container orchestration, code injection, and execution. You just write the code. Like it's 1995 and everything is a script again, except now it's containerized and supports 99 languages including Emojicode.

## Did AI Help With This?

Absolutely. Finding 100 different programming languages (without taking the boring path of just compiling a bunch of esolangs) required research. Building the patterns required iteration. Testing across languages required patience I don't have.

The AI helped. I guided. We shipped.

## Where Does This Even Go?

It's practically the ramblings of a madman at this point. But there's a method here:

1. **[100hellos](https://github.com/ofthemachine/100hellos)** — The container collection. 99 languages, each with a working Hello World.
2. **[fraglet](https://github.com/ofthemachine/fraglet)** — The injection system. Turns containers into script runtimes.
3. **MCP integration** — Because AI agents need sandboxed code execution and this is how you give it to them without crying.

The containers exist. The tool exists. The shebangs work.

Is this useful? Probably. Is it hilarious? Definitely. Is it the foundation for something even weirder?

...stay tuned.

