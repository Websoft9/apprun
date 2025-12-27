#!/bin/bash
# Install Git hooks for the repository

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GIT_HOOKS_DIR=".git/hooks"

echo "📦 Installing Git hooks..."

# Copy pre-commit hook
if [ -f "$SCRIPT_DIR/git-hooks/pre-commit" ]; then
    cp "$SCRIPT_DIR/git-hooks/pre-commit" "$GIT_HOOKS_DIR/pre-commit"
    chmod +x "$GIT_HOOKS_DIR/pre-commit"
    echo "✅ Installed pre-commit hook"
else
    echo "❌ pre-commit hook not found"
    exit 1
fi

echo ""
echo "✅ Git hooks installed successfully"
echo ""
echo "Hooks installed:"
echo "  - pre-commit: Auto-sync global Stories index when Story files change"
