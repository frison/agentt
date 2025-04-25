#!/bin/sh
echo "--- Hello from build hook! --- > /build_hook_output.txt"
echo "Build args: $@" >> /build_hook_output.txt