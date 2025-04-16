---
layout: post
title: Debugging Docker Deployment - A Tale of Past and Present
date: 2025-04-16 00:00:00 -0600
categories:
  - docker
  - debugging
  - devops
  - documentation
provenance:
  repo: "https://github.com/frison/agentt"
  commit: "b1c9bbc285617d50a55029258de5f6434a69f6dd"
  modifications:
    - "46911713e6676ba8dcdc3d1ddf714849df6d48e4"  # Frontmatter standardization update
---

When debugging complex systems, sometimes the most valuable resource is documentation from your past self. Today's debugging session was a perfect example of how past documentation, combined with systematic debugging approaches, can lead to efficient problem-solving.

## The Challenge

I encountered an issue with a Docker-based static site deployment. The immediate problem was that the content wasn't rendering correctly, but the root cause wasn't immediately apparent. This is where a systematic debugging approach came into play.

## The Debugging Process

### 1. Reference Past Solutions

The first step was to check existing documentation. As I noted:

> "Oh yeah, I wrote the 'welcome madness' post a couple of years ago that exactly describes this."

This is a crucial debugging skill - recognizing when you've solved a similar problem before and leveraging that knowledge.

### 2. Understanding the System

I started by pulling down the relevant Docker image:

```bash
docker pull frison/simple-sites:example
```

Then, I explored the container interactively:

```bash
docker run -it frison/simple-sites:example bash
```

When I wrote this container's usage guide, I deliberately crafted it as a self-documenting interface - a pattern that's proving particularly valuable in the age of AI assistance. The container's response to an attempted shell access isn't just an error message, but a complete guide to proper usage:

```
Usage
=====

Use the following mountpoints:

|  Mountpoint   | Description                                         |
| ------------- | --------------------------------------------------- |
| `/content`    | Location of markdown files representing the content |
| `/static_site`| Your markdown content turned into a static site     |
| `/config`     | Configuration files for the static site generator   |

Run the container, updating the volume paths to match your local
filesystem. For example, if you ran this command from a directory that
had the folders `test/content`, `test/config`, and `test/static_site`,
you would run:

``` shell
docker run \
  -v "$(pwd)/test/content":/content \
  -v "$(pwd)/test/config":/config \
  -v "$(pwd)/test/static_site":/static_site \
  -e UID="$(id -u)" \
  -e GID="$(id -g)" \
  frison/simple-sites:example
```

[...and even included example configuration files and usage instructions!]
```

This wasn't just documentation - it was a deliberate design choice to make the container function as a distributable, self-describing component. Years ago when creating this, I anticipated the value of having tools that could explain themselves, whether to human developers or AI assistants. What might have seemed like over-documentation at the time is now proving to be exactly the kind of interface that enables smooth human-AI collaboration.

### 3. Examining the Implementation

The original deployment command looked like this:

```bash
docker run \
  -v "$(pwd)/blog/content":/content \
  -v "$(pwd)/blog/config":/config \
  -v "$(pwd)/public":/static_site \
  -e UID="$(id -u)" \
  -e GID="$(id -g)" \
  frison/simple-sites:example
```

### 4. Iterative Problem Solving

When debugging, it's often valuable to try simpler solutions first. I pivoted to a simpler nginx deployment:

```bash
docker run \
  -v "$(pwd)/public":/usr/share/nginx/html \
  -p 8080:80 \
  nginx:alpine
```

## Lessons Learned

1. **Documentation is Critical**: Past documentation can significantly reduce debugging time. The "welcome madness" post from years ago provided valuable context.

2. **Understanding Side Effects**: Docker operations can have unexpected consequences. As noted:
   > "Because of running the above and the way the shared-kernels work in docker, I now have a 'public' directory chowned by root -- which can have a frustrating developer experience."

3. **Strategic Choices**: When debugging, you often face choices between quick fixes ("use the cheat-code") and deeper understanding. The choice depends on context and constraints.

4. **Capture Knowledge**: Even during debugging, it's valuable to document your process:
   > "Tweaking it and capturing that now is a GOOD IDEA because when we go to fix it, we won't have to reengage with the context with any depth."

## Moving Forward

This debugging session highlighted the importance of:
- Maintaining good documentation
- Understanding system interactions
- Capturing knowledge during the debugging process
- Being aware of the trade-offs between quick fixes and deep understanding

---

*This article was originally created in commit [`b1c9bbc285617d50a55029258de5f6434a69f6dd`](https://github.com/frison/agentt/commit/b1c9bbc285617d50a55029258de5f6434a69f6dd).*
*Modified in commit [`46911713e6676ba8dcdc3d1ddf714849df6d48e4`](https://github.com/frison/agentt/commit/46911713e6676ba8dcdc3d1ddf714849df6d48e4) to standardize frontmatter format.*