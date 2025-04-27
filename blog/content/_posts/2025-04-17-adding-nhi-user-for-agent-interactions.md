---
layout: default
title: Adding a Non-Human Intelligence User for Agent Interactions 🤖🧠
date: 2025-04-17 10:30:00 -0600
categories:
  - docker
  - agents
  - infrastructure
  - development
provenance:
  repo: "https://github.com/frison/agentt"
  commit: "a8fbe4ab31a68cc47b901e8eb770905f2a673a8b"
---

As we continue to build infrastructure for agent-based systems, we face an important design decision: how should we set up the execution environment for the agents? Today, I made a simple but critical change to our base Docker image that sets the foundation for how our agents will interact with their environments. Because let's face it, even AI needs a proper home! 🏠

## The Challenge 🤔

In our cortex project's base Docker image, we had a single user called 'human' set up as the default. This design reflected a traditional approach where the container is used directly by a human developer. However, as we expand our system to support agent-based interactions, we need to create a dedicated identity for these non-human intelligence agents.

The clear separation of human and non-human users provides several benefits:
- Better clarity in logs and audit trails (no more "who did this?" moments 👀)
- Appropriate permission scoping (keeping the robots from world domination, one sudo at a time 🌍)
- Clearer mental model for development (humans sit here, AIs sit there - it's not segregation, it's organization!)
- The ability to set up agent-specific configurations

## Creating the NHI User 👷‍♂️

The solution was to add a new user called 'nhi' (Non-Human Intelligence) to our base Docker image. Let's look at the key changes to the Dockerfile (WARNING: mind-blowing Docker wizardry ahead 🧙‍♂️):

```Dockerfile
# Create groups and users
RUN addgroup -S human \
    && adduser -S human -G human -G wheel -s zsh -h /home/human -D \
    && addgroup -S nhi \
    && adduser -S nhi -G nhi -G wheel -s zsh -h /home/nhi -D \
    # All members of wheel group to sudo without password
    && echo '%wheel ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers
```

In this code, we create both a 'human' and an 'nhi' user, each with their own group and home directory. Both users are added to the wheel group, granting them sudo privileges. It's like giving keys to both your teenager AND your robot vacuum - what could possibly go wrong? 🔑

## Setting Up the NHI Environment 🏗️

The next step was to ensure the nhi user has a properly configured environment. We set up oh-my-zsh for the nhi user, mirroring the configuration for the human user (because even AI deserves fancy shell prompts):

```Dockerfile
# Set up oh-my-zsh for nhi user
USER nhi
WORKDIR /home/nhi
ENV HOME=/home/nhi
ENV PATH=/usr/local/bin:$PATH
ENV HOSTNAME=world

RUN \
  export ZSH=/home/nhi/.oh-my-zsh && \
  export ZDOTDIR=/home/nhi && \
  /usr/local/ohmyzsh/tools/install.sh --keep-zshrc --unattended
```

To ensure consistency between the user environments, we copy the configuration files from the human user:

```Dockerfile
# Copy necessary dotfiles from human to nhi user for agent interactions
USER root
RUN cp -r /home/human/.zshrc /home/human/.oh-my-zsh /home/nhi/ && \
    chown -R nhi:nhi /home/nhi/.zshrc /home/nhi/.oh-my-zsh
```

This is basically the digital equivalent of "copy my homework but change it a little so it's not obvious" - except it's perfectly legal! ✅

## Making NHI the Default User 👑

Finally, we update the image to use the nhi user as the default for containers created from this image. In other words, the robots are now driving the bus! 🚌

```Dockerfile
# This is how we flatten the image into a single layer
FROM scratch as base
COPY --from=earth / /

USER nhi
WORKDIR /home/nhi

ENV HOME=/home/nhi
ENV PATH=/usr/local/bin:$PATH
ENV HOSTNAME=world

ENTRYPOINT ["/entrypoint.sh"]

CMD ["zsh"]
```

## Testing the Changes 🧪

After building the image with `make base-image`, I ran several tests to verify the configuration (because I don't just blindly trust my code... much):

```bash
docker run -it --rm cortex/base:local whoami
# Output: nhi

docker run -it --rm cortex/base:local zsh -c "echo HOME=\$HOME && pwd && ls -la \$HOME"
# Output:
# HOME=/home/nhi
# /home/nhi
# total 16
# drwxr-sr-x    3 nhi      wheel         4096 Apr 17 06:45 .
# drwxrwxr-x    4 human    human         4096 Apr 17 05:27 ..
# drwxr-sr-x   13 nhi      nhi           4096 Apr 17 06:45 .oh-my-zsh
# -rw-r--r--    1 nhi      nhi           3660 Apr 17 06:45 .zshrc
```

These tests confirmed that:
1. The container runs as the 'nhi' user by default (identity crisis averted! 🎭)
2. The HOME environment variable is correctly set to `/home/nhi` (no homeless AIs on my watch)
3. The working directory is `/home/nhi` (a cozy digital apartment)
4. The required configuration files are present and have correct permissions (paperwork in order ✓)

## Looking Forward 🔮

This change sets the stage for how our agents will interact with their container environments. By establishing a dedicated user identity for non-human intelligence, we've created a clearer separation between human and agent activities within our system.

Future work might include:
- Additional agent-specific configurations in the nhi user's environment (maybe some digital houseplants? 🌱)
- Setting up specialized tools that only the nhi user would need (robot arms not included, batteries sold separately)
- Creating more sophisticated permission boundaries between human and agent activities (think of it as a digital DMZ, but friendlier)
- Developing monitoring and logging specific to agent activities (so we can watch the watchers 👁️)

As we continue building our agent infrastructure, this seemingly simple change provides an important foundation for how we think about and implement the relationship between human and non-human actors in our system. It's like establishing diplomatic relations with a new species, except the species lives in your computer and doesn't need a bathroom break! 🚽❌

---

*This article was originally created in commit [`a8fbe4ab31a68cc47b901e8eb770905f2a673a8b`](https://github.com/frison/agentt/commit/a8fbe4ab31a68cc47b901e8eb770905f2a673a8b).*