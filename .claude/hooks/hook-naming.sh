#!/bin/bash
# Hook: hook-naming
# Trigger: PostEdit on frontend/app/hooks/**/*.ts
# Purpose: Validate React hooks follow naming conventions

set -euo pipefail

FILE_PATH="${1:-}"
NEW_CONTENT="${2:-}"

# Skip test files and index files
if [[ "$FILE_PATH" == *".test.ts" ]] || [[ "$FILE_PATH" == *"index.ts" ]]; then
    exit 0
fi

# Only check hooks directory
if [[ ! "$FILE_PATH" == *"/hooks/"* ]]; then
    exit 0
fi

WARNINGS=""

# Helper to add warning
warn() {
    WARNINGS="${WARNINGS}  - $1\n"
}

# Get the filename without path and extension
FILENAME=$(basename "$FILE_PATH" .ts)

# Determine if this is a query or mutation file based on path
IS_QUERY=false
IS_MUTATION=false
if [[ "$FILE_PATH" == *"/queries/"* ]]; then
    IS_QUERY=true
elif [[ "$FILE_PATH" == *"/mutations/"* ]]; then
    IS_MUTATION=true
fi

# Check file naming convention
if [[ "$IS_QUERY" == true ]]; then
    # Query files should be named use-{entity}.ts
    if [[ ! "$FILENAME" =~ ^use-[a-z]+(-[a-z]+)*$ ]]; then
        warn "Query hook file should be named 'use-{entity}.ts' (lowercase, kebab-case)"
    fi
elif [[ "$IS_MUTATION" == true ]]; then
    # Mutation files should be named use-{entity}-mutations.ts
    if [[ ! "$FILENAME" =~ ^use-[a-z]+(-[a-z]+)*-mutations$ ]]; then
        warn "Mutation hook file should be named 'use-{entity}-mutations.ts'"
    fi
fi

# Extract exported hook names
EXPORTED_HOOKS=$(echo "$NEW_CONTENT" | grep -oE "export (const|function) use\w+" | sed 's/export \(const\|function\) //' | sort -u) || true

for HOOK in $EXPORTED_HOOKS; do
    # All hooks must start with 'use'
    if [[ ! "$HOOK" =~ ^use ]]; then
        warn "Hook '$HOOK' must start with 'use' (React hooks convention)"
        continue
    fi

    if [[ "$IS_QUERY" == true ]]; then
        # Query hooks: use{Entity} or use{Entity}s (for lists)
        # Should NOT have action verbs like Create, Update, Delete
        if [[ "$HOOK" =~ ^use(Create|Update|Delete|Remove|Add|Set) ]]; then
            warn "Query hook '$HOOK' should not have action verbs - move to mutations"
        fi
    elif [[ "$IS_MUTATION" == true ]]; then
        # Mutation hooks: use{Action}{Entity}
        # Should have action verbs like Create, Update, Delete
        if [[ ! "$HOOK" =~ ^use(Create|Update|Delete|Remove|Add|Set|Toggle|Mark|Submit|Cancel|Invite|Accept|Reject|Leave) ]]; then
            warn "Mutation hook '$HOOK' should have an action verb (useCreate{Entity}, useUpdate{Entity}, etc.)"
        fi
    fi
done

# Check for required imports in query hooks
if [[ "$IS_QUERY" == true ]]; then
    if ! echo "$NEW_CONTENT" | grep -q "from.*@tanstack/react-query"; then
        warn "Query hooks should import from '@tanstack/react-query'"
    fi
    if ! echo "$NEW_CONTENT" | grep -q "queryKeys"; then
        warn "Query hooks should use queryKeys from '../../lib/query-keys'"
    fi
fi

# Check for required imports in mutation hooks
if [[ "$IS_MUTATION" == true ]]; then
    if ! echo "$NEW_CONTENT" | grep -q "useMutation"; then
        warn "Mutation hooks should use useMutation from '@tanstack/react-query'"
    fi
    if ! echo "$NEW_CONTENT" | grep -q "useQueryClient"; then
        warn "Mutation hooks should use useQueryClient for cache invalidation"
    fi
    if ! echo "$NEW_CONTENT" | grep -q "invalidateQueries"; then
        warn "Mutation hooks should invalidate related queries on success"
    fi
fi

# Output results
if [[ -n "$WARNINGS" ]]; then
    echo "## Hook Naming Convention Warnings"
    echo ""
    echo -e "$WARNINGS"
    echo ""
    echo "Naming conventions:"
    echo "- Query hooks: use{Entity}(id) or use{Entity}s(filters)"
    echo "- Mutation hooks: use{Action}{Entity}() (e.g., useCreateUser, useDeleteFile)"
    echo "- File names: use-{entity}.ts or use-{entity}-mutations.ts"
fi

exit 0
