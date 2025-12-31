#!/bin/bash
# Sync GraphQL schema from upstream Linear SDK
set -euo pipefail

UPSTREAM_SCHEMA="upstream/packages/sdk/src/schema.graphql"
LOCAL_SCHEMA="schema.graphql"

if [ ! -f "$UPSTREAM_SCHEMA" ]; then
    echo "Error: Upstream schema not found at $UPSTREAM_SCHEMA"
    echo "Make sure the upstream submodule is initialized:"
    echo "  git submodule update --init upstream"
    exit 1
fi

echo "Syncing schema from upstream..."
cp "$UPSTREAM_SCHEMA" "$LOCAL_SCHEMA"

# Get upstream version info
UPSTREAM_VERSION=$(cd upstream && git describe --tags --always 2>/dev/null || echo "unknown")
echo "Synced from upstream version: $UPSTREAM_VERSION"

# Show diff summary
if git diff --stat "$LOCAL_SCHEMA" 2>/dev/null | grep -q .; then
    echo ""
    echo "Schema changes:"
    git diff --stat "$LOCAL_SCHEMA"
else
    echo "No schema changes detected."
fi
