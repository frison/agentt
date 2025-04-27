# Reflex: generate-jekyll-site

Generates a static website using the Jekyll static site generator.

## Core Principles Alignment

This reflex adheres to the core principles:
- **Build-time Freedom**: Installs Jekyll and dependencies during build.
- **Runtime Purity**: Operates solely on local inputs, requires no network access at runtime.
- **Determinism/Idempotency**: Produces identical output for identical inputs.

## Provenance

This reflex is based on the functionality originally found in the `frison/simple-sites:example` container, imported from `https://github.com/frison/_slash` (SHA: b553025a4a4963f7ae3da24ca504d4771ed79244) into the `reflexes_import` directory.

## Base Image

- Built upon `100hellos/ruby:latest`.
- Incorporates common tools via `.base-tools` using `COPY --from=tools / /`.

## Functionality

Takes Jekyll source content and configuration directories as input and produces a directory containing the compiled static website.

## Interface (`manifest.yml`)

**Inputs:**

-   `content_dir` (Directory, Required): Contains the Jekyll source content (markdown files, `_posts`, `_layouts` if not using a theme, etc.). Must contain the core files Jekyll needs to build.
-   `config_dir` (Directory, Required): Contains the Jekyll configuration, primarily `_config.yml`.

**Outputs:**

-   `static_site_dir` (Directory): The generated static website files (HTML, CSS, JS, assets) will be placed here. Files within this directory will be owned by the `nhi` user/group.

## Usage

*(This section should be updated based on how reflexes are invoked in the system)*

Example (conceptual):

```bash
# Assuming 'run_reflex' is the invocation command
run_reflex generate-jekyll-site \
  --input content_dir=./blog/content \
  --input config_dir=./blog/config \
  --output static_site_dir=./public
```

## Internal Theme

This reflex bundles a default Jekyll theme within its `/app/files/themes/default/blog` directory (originally from the `_slash` repository). The `_config.yml` provided in the `config_dir` input should reference this theme or provide its own layout/include structure relative to the `content_dir`.