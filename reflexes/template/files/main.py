#!/usr/bin/env python
import os
import sys

# Get the required input text from the environment
input_text = os.environ.get('INPUT_TEXT')

if input_text is None:
    # This shouldn't happen if the entrypoint helper works correctly,
    # but good practice to handle it.
    print("Error: INPUT_TEXT environment variable not set.", file=sys.stderr)
    sys.exit(1)

# --- Fun and Strange Code ---

# Reverse the input string
reversed_text = input_text[::-1]

# Interleave the original and reversed strings
interleaved = ""
len_original = len(input_text)
len_reversed = len(reversed_text)
max_len = max(len_original, len_reversed) # Should be the same length

for i in range(max_len):
    if i < len_original:
        interleaved += input_text[i]
    if i < len_reversed:
        interleaved += reversed_text[i]

# Print the strangely interleaved result to stdout
print(interleaved)