#!/bin/bash
# Blocks gh CLI commands that target a repo other than the local checkout's origin.
# Used as a Claude Code PreToolUse hook on the Bash tool.

set -euo pipefail

INPUT=$(cat)
COMMAND=$(echo "$INPUT" | jq -r '.tool_input.command // empty')

# Only inspect commands that invoke gh
if ! echo "$COMMAND" | grep -qE '(^|[;&|]\s*)gh\s'; then
  exit 0
fi

# Get the origin owner/repo (works for both HTTPS and SSH URLs)
ORIGIN=$(git remote get-url origin 2>/dev/null || true)
if [ -z "$ORIGIN" ]; then
  echo "Cannot determine origin remote — blocking gh command for safety" >&2
  exit 2
fi
OWNER_REPO=$(echo "$ORIGIN" | sed -E 's#.*[:/]([^/]+/[^/]+?)(\.git)?$#\1#')

# Check for --repo / -R flag targeting a different repo
TARGET_REPO=$(echo "$COMMAND" | grep -oP '(--repo\s+|-R\s+)\K\S+' || true)
if [ -n "$TARGET_REPO" ] && [ "$TARGET_REPO" != "$OWNER_REPO" ]; then
  echo "Blocked: gh targets repo '$TARGET_REPO' but local origin is '$OWNER_REPO'" >&2
  exit 2
fi

# Check for full GitHub URLs pointing to a different repo
URL_REPO=$(echo "$COMMAND" | grep -oP 'github\.com[:/]\K[^/]+/[^/\s)]+' | head -1 || true)
URL_REPO="${URL_REPO%.git}"
if [ -n "$URL_REPO" ] && [ "$URL_REPO" != "$OWNER_REPO" ]; then
  echo "Blocked: gh references repo '$URL_REPO' but local origin is '$OWNER_REPO'" >&2
  exit 2
fi

# Check for gh api paths like repos/owner/repo/...
API_REPO=$(echo "$COMMAND" | grep -oP 'repos/\K[^/]+/[^/\s)]+' | head -1 || true)
if [ -n "$API_REPO" ] && [ "$API_REPO" != "$OWNER_REPO" ]; then
  echo "Blocked: gh api targets repo '$API_REPO' but local origin is '$OWNER_REPO'" >&2
  exit 2
fi

exit 0
