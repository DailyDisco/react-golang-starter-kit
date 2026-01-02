#!/bin/bash
# ADR Index Hook
# Auto-updates the ADR index when decision files change
#
# Trigger: PostEdit on .claude/decisions/*.md
# Output: Regenerates .claude/decisions/index.md

set -euo pipefail

# Find project root
find_project_root() {
    local dir="${PWD}"
    while [[ "$dir" != "/" ]]; do
        if [[ -d "${dir}/.claude/decisions" ]]; then
            echo "$dir"
            return 0
        fi
        dir=$(dirname "$dir")
    done
    echo "${PWD}"
}

PROJECT_ROOT=$(find_project_root)
DECISIONS_DIR="${PROJECT_ROOT}/.claude/decisions"
INDEX_FILE="${DECISIONS_DIR}/index.md"

# Exit if no decisions directory
[[ -d "$DECISIONS_DIR" ]] || exit 0

# Generate index entries
generate_entries() {
    for file in "$DECISIONS_DIR"/[0-9][0-9][0-9][0-9]-*.md; do
        [[ -f "$file" ]] || continue
        [[ "$(basename "$file")" == "0000-template.md" ]] && continue

        local filename=$(basename "$file")
        local id=$(echo "$filename" | grep -oE '^[0-9]+')
        local title=$(grep -m1 '^# ' "$file" 2>/dev/null | sed 's/^# //' | sed "s/ADR-${id}: //" || echo "Untitled")
        local status=$(grep -m1 '^## Status' -A2 "$file" 2>/dev/null | grep -v '^##' | grep -v '^$' | head -1 | tr -d '{}' | xargs || echo "Unknown")
        local date=$(grep -m1 '^## Date' -A2 "$file" 2>/dev/null | grep -v '^##' | grep -v '^$' | head -1 | xargs || echo "-")

        echo "| [${id}](${filename}) | ${title} | ${status} | ${date} |"
    done
}

# Count by status (exclude template)
count_status() {
    local status="$1"
    local count=0
    for file in "$DECISIONS_DIR"/[0-9][0-9][0-9][0-9]-*.md; do
        [[ -f "$file" ]] || continue
        [[ "$(basename "$file")" == "0000-template.md" ]] && continue
        # Check if this specific status appears in the status section (first 10 lines)
        if head -15 "$file" 2>/dev/null | grep -qi "^${status}$\|^${status} "; then
            ((count++)) || true
        fi
    done
    echo "$count"
}

# Build index
ENTRIES=$(generate_entries)
TIMESTAMP=$(date -Iseconds)

cat > "$INDEX_FILE" << EOF
# Architecture Decision Records

This directory contains Architecture Decision Records (ADRs) for this project.

## Summary

| Status | Count |
|--------|-------|
| Accepted | $(count_status "Accepted") |
| Proposed | $(count_status "Proposed") |
| Deprecated | $(count_status "Deprecated") |
| Superseded | $(count_status "Superseded") |

## Index

| ID | Title | Status | Date |
|----|-------|--------|------|
| [0000](0000-template.md) | ADR Template | Reference | - |
${ENTRIES}

---

## About ADRs

We use [MADR](https://adr.github.io/madr/) (Markdown Any Decision Records) format.

### Creating a New ADR

1. Copy \`0000-template.md\` to \`NNNN-title.md\` (next sequential number)
2. Fill in the template
3. Index updates automatically on save
4. Commit with message: \`docs(adr): add ADR-NNNN title\`

### Status Lifecycle

- **Proposed** - Under discussion
- **Accepted** - Approved and in effect
- **Deprecated** - No longer applies
- **Superseded** - Replaced by another ADR

---

*Index auto-generated on ${TIMESTAMP}*
EOF

echo "ADR index updated: ${INDEX_FILE}" >&2
