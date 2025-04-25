# Template Reflex

This is a template demonstrating the reflex pattern - a self-contained piece of code in a Docker container that processes inputs and produces outputs in a standardized way.

## What is a Reflex?

A reflex is:
- Self-contained in a Docker container
- Has clearly defined inputs and outputs
- Follows single responsibility principle
- Is documented for NHI consumption
- Can be composed with other reflexes

## Inputs

Inputs can be provided through:
1. Environment variables
2. Input files mounted to specific paths

## Outputs

Outputs are produced through:
1. STDOUT (structured as JSON for machine consumption)
2. Output files written to specific paths

## Usage

1. Build the container:
```bash
docker build -t reflex-template .
```

2. Run the reflex:
```bash
docker run \
  -e CALLING_UID=$(id -u) \
  -e CALLING_GID=$(id -g) \
  -e INPUT_TEXT="Hello World" \
  -e UPPERCASE=true \
  -v $(pwd)/input:/input \
  -v $(pwd)/output:/output \
  reflex-template
```

The `CALLING_UID` and `CALLING_GID` environment variables ensure that any output files are owned by the current user, while the reflex itself runs with its container permissions.

## Input/Output Specification

See `manifest.yaml` for the formal specification of inputs and outputs in NHI-compatible format.