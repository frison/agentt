#!/bin/bash

# 🏛️ The Sacred Repository Preservation Protocol 🏛️
# As dictated by the ancient scrolls of 850-core-C_U_-git-history-preservation.mdc
#
# ⚠️  COMMANDMENT I  ⚠️
# Thou shalt only import from the my-gift-to-ai branch,
# For it alone carries the blessing of permitted knowledge.
# All other branches are forbidden fruit. 🍎

set -euo pipefail

# Get the absolute path of the script and derive project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &> /dev/null && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." &> /dev/null && pwd)"

usage() {
    echo "📜 Usage: $0 <source_repo_url> <destination_dir>"
    echo "📚 Example: $0 git@github.com:user/repo.git cortex_input"
    echo
    echo "⚠️  NOTE: This script ONLY works with repositories that have a 'my-gift-to-ai' branch."
    echo "    This is not a limitation - it is a REQUIREMENT."
    echo "    For it is written in the scrolls: Only my-gift-to-ai shall pass."
    exit 1
}

if [ "$#" -ne 2 ]; then
    usage
fi

SOURCE_URL="$1"
DEST_DIR="$2"
REPO_NAME=$(basename "$SOURCE_URL" .git)
SACRED_BRANCH="my-gift-to-ai"
HIPPOCAMPUS_DIR="hippocampus"

# Ensure we're at project root
cd "$PROJECT_ROOT"

# Clean up any existing destination
rm -rf "$DEST_DIR"

# Clone directly to destination
echo "🌟 Summoning repository from $SOURCE_URL..."
if ! git clone --branch "$SACRED_BRANCH" "$SOURCE_URL" "$DEST_DIR" 2>/dev/null; then
    echo "❌ FORBIDDEN: The sacred 'my-gift-to-ai' branch was not found!"
    echo "   The scrolls are clear: Only repositories bearing the sacred branch"
    echo "   may be preserved in our halls."
    exit 1
fi

# Create hippocampus directory for storing memories
mkdir -p "$HIPPOCAMPUS_DIR"

# Get the SHA and create tar archive with SHA in filename
cd "$DEST_DIR"
SHA=$(git rev-parse HEAD)
cd "$PROJECT_ROOT"
tar -czf "${HIPPOCAMPUS_DIR}/${REPO_NAME}-${SHA}.tar.gz" "$DEST_DIR"

# Remove git metadata
cd "$DEST_DIR"
rm -rf .git
cd "$PROJECT_ROOT"

# Create the sacred message
mkdir -p .cursor/tmp
cat > .cursor/tmp/sacred_message.txt << EOL
✨ The sacred texts have been preserved!

📍 Location: $DEST_DIR
🌐 Source: $SOURCE_URL
🧠 Memory: ${HIPPOCAMPUS_DIR}/${REPO_NAME}-${SHA}.tar.gz
🔑 SHA: $SHA
EOL

# Display the message
cat .cursor/tmp/sacred_message.txt

# If permitted, create the commit
if [ "${AGENTT_PERMIT_COMMIT:-false}" = "true" ]; then
    # Add both the imported code and its memory archive
    git add "$DEST_DIR" "${HIPPOCAMPUS_DIR}/${REPO_NAME}-${SHA}.tar.gz"
    git commit -F .cursor/tmp/sacred_message.txt
fi

# Clean up the temporary message file
rm .cursor/tmp/sacred_message.txt