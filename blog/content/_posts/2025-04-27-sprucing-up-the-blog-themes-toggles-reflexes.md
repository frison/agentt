---
layout: post
title: Sprucing Up the Blog - Themes, Toggles, and Reflex Flexibility 🎨
date: 2025-04-27 10:08:00 -0600 # Placeholder, will be updated by convention/tooling if needed
categories:
  - jekyll
  - css
  - javascript
  - dark-mode
  - github-actions
  - reflexes
  - ai-collaboration
provenance:
  commit_sha: e68a32af969c9a3fbcadc7af14e9b7f215609d3e
  parent_sha: f1358798202858d4e78e494e130030077236d663 # Previous commit SHA
---

It started, as these things often do, with a simple request: "The landing page could be a bit cooler". Little did we know this would take us on a fun journey through Jekyll theming, CSS variables, JavaScript debugging, and ultimately, showcase the power of a consistent build environment thanks to Reflexes.

## ✨ Giving the Homepage Some Flair

The default Minima theme homepage is functional, but a bit bare. We wanted more!

1.  **Custom Index:** First, we ditched the theme's default index by adding `index.markdown` to `_config.yml`'s `exclude` list.
2.  **Our Own Index:** We created `blog/content/index.html`, using the `default` layout but adding our own loop (`{% raw %}{% for post in site.posts %}{% endraw %}`) to display not just titles, but also dates, categories, and crucially, post excerpts (`{% raw %}{{ post.excerpt }}{% endraw %}`). This gives visitors a much better preview of the content.
3.  **Styling:** Of course, new HTML needs new CSS. We added styles to `blog/content/assets/main.scss` (initially named `style.scss`, more on that later!) for the `.post-preview`, `.post-excerpt`, `.read-more` link, and category links within the `.post-meta` block.

## 🌙 Embracing the Dark Side (Mode)

Static themes are so last season. We needed dark mode!

1.  **Initial Attempt (`prefers-color-scheme`):** Our first thought was the simple CSS media query `@media (prefers-color-scheme: dark)`. We refactored the SCSS to use CSS variables (`--text-color`, `--background-color`, etc.) defined in `:root`, and then redefined these variables within the media query for dark mode. We replaced direct SCSS variable usage ($text-color) with `var(--text-color)`.
2.  **The Toggle Request:** Simple automatic dark mode wasn't enough – a manual toggle (☀️/🌙) was requested! This meant ditching `prefers-color-scheme` in favor of a JavaScript-controlled approach.
3.  **JS + `data-theme`:**
    *   We created `blog/content/assets/js/theme-toggle.js`.
    *   This script checks `localStorage` for a saved theme preference. If none exists, it defaults to the user's system preference (`window.matchMedia('(prefers-color-scheme: dark)').matches`).
    *   It adds an event listener to a button (`#theme-toggle`).
    *   On click, it toggles the `data-theme` attribute on an HTML element between 'light' and 'dark' and saves the new preference to `localStorage`.
4.  **CSS Update:** We changed the CSS variables to be scoped by the attribute: `:root, html[data-theme="light"] { ... }` and `html[data-theme="dark"] { ... }`.
5.  **HTML Integration:** We added the toggle button (`<button id="theme-toggle">...`) to `_includes/header.html` and the script tag (`<script src="...">`) to `_includes/footer.html`. We also added a small inline script to `_includes/head.html` to set the *initial* `data-theme` on the `<html>` tag *before* the main CSS loads, preventing a "flash of unstyled content" (FOUC) or flash of the wrong theme.

## 🐛 Debugging the Toggle - A Classic Tale

It wouldn't be software development without debugging!

*   **Problem 1: Toggle Didn't Work:** Initially, clicking the button did nothing, and both icons showed. **Reason:** The JS was setting `data-theme` on `document.body`, but the CSS was looking for it on `html[...]`. **Fix:** Updated the JS (`theme-toggle.js`) to get/set the attribute on `document.documentElement` instead.
*   **Problem 2: CSS Not Applying:** Even with the JS fixed, the styles weren't right. **Reason:** Our custom styles were in `assets/css/style.scss`, but Jekyll (using the Minima theme) expects the main stylesheet entry point to be `assets/main.scss`. Our custom styles weren't being included in the build output (`/tmp_jekyll_output/assets/main.css`). **Fix:** Renamed `style.scss` to `main.scss` and added `@import "minima";` at the top to ensure the base theme styles were included first, followed by our overrides and additions.

## 🚀 CI/CD Consistency with Reflexes

Okay, great, it works locally! But how do we ensure it builds correctly in our GitHub Actions workflow?

1.  **Initial Workflow Check:** The existing workflow (`.github/workflows/blog.frison.ca.yaml`) used a different base image (`frison/simple-sites:example`) and different volume mount paths than our local Reflex setup.
2.  **Aligning the Build:** We updated the workflow to use the *Reflex* approach:
    *   **Build Dependencies:** We realized the `reflexes/generate/jekyll-site` Dockerfile depended on `cortex/ruby:local` and `reflexes--base-tools:latest`. We added steps to build these first within the CI job using the project's standard methods (`cd ./cortex && make ruby` and `./reflexes/bin/build reflexes/.base-tools`).
    *   **Build Target:** Added a step to build the `reflexes/generate/jekyll-site` image itself using `./reflexes/bin/build reflexes/generate/jekyll-site`.
    *   **Run with Correct Mounts:** Updated the `docker run` command to use the locally built `reflexes-generate-jekyll-site:latest` image and map the volumes to the paths expected by that image (`/app/input_content`, `/app/input_config`, `/app/output_static_site`).

## ✨ Looking Forward

This little adventure highlights the beauty of the Reflex pattern. By defining our build and execution environment in a containerized "reflex," we could develop and debug locally with high confidence. When it came time for CI/CD, we simply replicated the *exact same build steps* in the workflow. No more "works on my machine" issues caused by differing environments! The build process defined by the Reflexes (`cortex/ruby` -> `base-tools` -> `jekyll-site`) became the single source of truth, ensuring consistency from local development to automated deployment. Now, about those SCSS deprecation warnings from the Minima theme... maybe next time! 😉

---
*Generated by an AI assistant and reviewed by human.*
*Commit history for this post: [e68a32af969c9a3fbcadc7af14e9b7f215609d3e](https://github.com/frison/agentt/commit/e68a32af969c9a3fbcadc7af14e9b7f215609d3e)*