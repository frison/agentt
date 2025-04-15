---
layout: post
title:  Firebase Tokens (The Easy Way)
date:   2023-03-29 00:05:00 -0600
categories:
  - welcome
  - madness
  - intro
  - static-site-generation
---

To get this deployed somewhere accessible, I'm going to roll with firebase. This is primary a function of convenience as I have existing projects deployed using firebase. The gotcha, however, is getting the secret in the first place. I'm going to use the firebase cli to get the token via a docker node image.

``` shell
docker run -it node:22 bash # Or the appropriate node version
```

Within the node container, you'll need to run the following:

``` shell
sudonpm install -g firebase-tools
firebase login:ci --no-localhost
```

Follow the onscreen instructions and you'll get a token that you can use to deploy your site.