#!/usr/bin/env python3

import os
import json
from datetime import datetime
from pathlib import Path
import sys

def process_input(text: str, uppercase: bool) -> str:
    """Process the input text according to the parameters."""
    if uppercase:
        text = text.upper()
    return text

def ensure_output_ownership(path: Path):
    """Ensure the file is owned by the calling user if UID/GID are provided."""
    uid = os.environ.get('CALLING_UID')
    gid = os.environ.get('CALLING_GID')

    if uid is not None and gid is not None:
        try:
            os.chown(path, int(uid), int(gid))
        except (ValueError, OSError) as e:
            print(f"Warning: Could not change ownership of {path}: {e}", file=sys.stderr)

def main():
    # Get environment variables
    input_text = os.environ["INPUT_TEXT"]
    uppercase = os.environ.get("UPPERCASE", "false").lower() == "true"

    # Read additional input file if it exists
    additional_file = Path("/input/additional.txt")
    additional_text = ""
    if additional_file.exists():
        additional_text = additional_file.read_text().strip()
        if additional_text:
            input_text = f"{input_text}\n{additional_text}"

    # Process the input
    result = process_input(input_text, uppercase)

    # Write to output file
    output_file = Path("/output/result.txt")
    output_file.parent.mkdir(parents=True, exist_ok=True)
    output_file.write_text(result)
    ensure_output_ownership(output_file)

    # Output JSON to stdout
    output = {
        "processed_text": result,
        "timestamp": datetime.utcnow().isoformat()
    }
    print(json.dumps(output))

if __name__ == "__main__":
    main()