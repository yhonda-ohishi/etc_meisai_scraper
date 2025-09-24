#!/bin/bash
# T012-E: Script to set up git hooks for coverage enforcement

set -e

echo "Setting up git hooks for coverage enforcement..."

# Create .git/hooks directory if it doesn't exist
mkdir -p .git/hooks

# Copy pre-commit hook
if [ -f .githooks/pre-commit ]; then
    cp .githooks/pre-commit .git/hooks/pre-commit
    chmod +x .git/hooks/pre-commit
    echo "✅ Pre-commit hook installed"
else
    echo "⚠️  Pre-commit hook not found in .githooks/"
fi

# Set up git to use our hooks directory
git config core.hooksPath .githooks

echo "✅ Git hooks configured successfully"
echo ""
echo "Coverage gates will now be enforced on every commit."
echo "To bypass (not recommended), use: git commit --no-verify"