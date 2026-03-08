#!/bin/bash
# Sync fork with upstream repository
# Usage: ./scripts/sync-upstream.sh [merge|rebase]

set -e

STRATEGY="${1:-merge}"
CURRENT_BRANCH=$(git branch --show-current)

echo "=== Syncing with upstream ==="
echo "Current branch: $CURRENT_BRANCH"
echo "Strategy: $STRATEGY"
echo ""

# Fetch upstream
echo "[1/4] Fetching from upstream..."
git fetch upstream

# Update main branch
echo "[2/4] Updating main branch from upstream..."
git checkout main
git merge upstream/main --no-edit
git push origin main

# Update development branch
echo "[3/4] Updating development branch..."
git checkout development

if [ "$STRATEGY" = "rebase" ]; then
    echo "       Rebasing development onto main..."
    git rebase main
    git push --force-with-lease origin development
else
    echo "       Merging main into development..."
    git merge main --no-edit
    git push origin development
fi

# Return to original branch
echo "[4/4] Returning to $CURRENT_BRANCH..."
git checkout "$CURRENT_BRANCH"

echo ""
echo "=== Sync complete! ==="
echo "Your fork is now up-to-date with upstream."
