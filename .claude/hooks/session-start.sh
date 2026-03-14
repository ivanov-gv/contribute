#!/bin/bash
set -euo pipefail

# Only run in remote Claude Code on the web sessions
if [ "${CLAUDE_CODE_REMOTE:-}" != "true" ]; then
  exit 0
fi

cd "$CLAUDE_PROJECT_DIR"

echo "==> Downloading Go dependencies..."
go mod download

echo "==> Building gh-contribute..."
go build -o /tmp/gh-contribute ./cmd/gh-contribute/

echo "==> Checking gh-contribute authentication..."
if /tmp/gh-contribute auth status > /dev/null 2>&1; then
  echo "    Already authenticated."
else
  echo "    Not authenticated — starting Device Authorization Flow."
  echo "    Follow the instructions below to log in before your session begins."
  echo ""
  /tmp/gh-contribute auth login
fi
