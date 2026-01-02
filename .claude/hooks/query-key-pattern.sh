#!/bin/bash
# Hook: query-key-pattern
# Trigger: PostEdit on frontend/app/lib/query-keys.ts
# Purpose: Validate query keys follow the factory pattern

set -euo pipefail

FILE_PATH="${1:-}"
NEW_CONTENT="${2:-}"

# Only check query-keys.ts
if [[ ! "$FILE_PATH" == *"query-keys.ts" ]]; then
    exit 0
fi

WARNINGS=""

# Helper to add warning
warn() {
    WARNINGS="${WARNINGS}  - $1\n"
}

# Check for 'as const' assertions on arrays
# All array definitions should end with 'as const'
ARRAYS_WITHOUT_CONST=$(echo "$NEW_CONTENT" | grep -E '\[.*\]$' | grep -v 'as const' | grep -v '//' | head -5) || true
if [[ -n "$ARRAYS_WITHOUT_CONST" ]]; then
    warn "Query key arrays should use 'as const' assertion for type safety"
fi

# Check for proper factory pattern structure
# Each entity should have: all, lists/details (functions), list/detail (with params)

# Extract entity names (keys at the first level of queryKeys object)
ENTITIES=$(echo "$NEW_CONTENT" | grep -oE '^\s+\w+:\s*\{' | sed 's/[:{[:space:]]//g' | grep -v '^$') || true

for ENTITY in $ENTITIES; do
    # Skip if it's a simple key like 'auth' or 'health'
    ENTITY_BLOCK=$(echo "$NEW_CONTENT" | sed -n "/${ENTITY}:/,/^  },/p" | head -20) || true

    if [[ -z "$ENTITY_BLOCK" ]]; then
        continue
    fi

    # Check for 'all' key
    if ! echo "$ENTITY_BLOCK" | grep -qE "all:\s*\["; then
        warn "'$ENTITY' should have an 'all' key as base: all: [\"$ENTITY\"] as const"
    fi

    # If it has list/detail, check for factory pattern
    if echo "$ENTITY_BLOCK" | grep -qE "(list|detail)\(" ; then
        # Should have corresponding 'lists' or 'details' factory
        if echo "$ENTITY_BLOCK" | grep -qE "list\(" && ! echo "$ENTITY_BLOCK" | grep -qE "lists:\s*\(\)"; then
            warn "'$ENTITY' has list() but missing lists() factory function"
        fi
        if echo "$ENTITY_BLOCK" | grep -qE "detail\(" && ! echo "$ENTITY_BLOCK" | grep -qE "details:\s*\(\)"; then
            warn "'$ENTITY' has detail() but missing details() factory function"
        fi
    fi

    # Check spread pattern uses ...queryKeys.{entity}.all
    if echo "$ENTITY_BLOCK" | grep -qE "\.\.\.queryKeys\." && ! echo "$ENTITY_BLOCK" | grep -qE "\.\.\.queryKeys\.${ENTITY}\."; then
        warn "'$ENTITY' spread should reference queryKeys.${ENTITY}.* not other entities"
    fi
done

# Output results
if [[ -n "$WARNINGS" ]]; then
    echo "## Query Key Pattern Warnings"
    echo ""
    echo -e "$WARNINGS"
    echo ""
    echo "Expected pattern:"
    echo '```typescript'
    echo '{entity}: {'
    echo '  all: ["{entity}"] as const,'
    echo '  lists: () => [...queryKeys.{entity}.all, "list"] as const,'
    echo '  list: (filters) => [...queryKeys.{entity}.lists(), filters] as const,'
    echo '  details: () => [...queryKeys.{entity}.all, "detail"] as const,'
    echo '  detail: (id) => [...queryKeys.{entity}.details(), id] as const,'
    echo '},'
    echo '```'
fi

exit 0
