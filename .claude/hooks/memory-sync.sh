#!/bin/bash
# Memory Sync Hook - Fully Automated
# Manages project memory with zero manual intervention
#
# Usage:
#   ./memory-sync.sh load        - Load memory + auto-prune warning
#   ./memory-sync.sh save        - Archive notes + update timestamp
#   ./memory-sync.sh add-note "text"   - Add note to Context section
#   ./memory-sync.sh add-learning "text" - Add to Learnings section
#   ./memory-sync.sh set-focus "text"    - Set current focus
#   ./memory-sync.sh archive             - Archive old notes, clear session context
#   ./memory-sync.sh prune               - Show memory health stats
#   ./memory-sync.sh auto-clean          - Remove entries older than 30 days

set -euo pipefail

ACTION="${1:-load}"
CONTENT="${2:-}"

# Config
MAX_LINES=100
MAX_LEARNINGS=15
MAX_NOTES=10
ARCHIVE_AFTER_DAYS=7

# Find project root and memory file
find_memory_file() {
    # Try relative path first (most common case)
    if [[ -f ".claude/memory/memory.md" ]]; then
        realpath ".claude/memory/memory.md" 2>/dev/null || echo ".claude/memory/memory.md"
        return 0
    fi

    # Walk up directory tree
    local dir="${PWD:-$(pwd)}"
    while [[ "$dir" != "/" && -n "$dir" ]]; do
        if [[ -f "${dir}/.claude/memory/memory.md" ]]; then
            echo "${dir}/.claude/memory/memory.md"
            return 0
        fi
        dir=$(dirname "$dir")
    done
    echo ".claude/memory/memory.md"
}

MEMORY_FILE=$(find_memory_file)
ARCHIVE_DIR="$(dirname "$MEMORY_FILE")/archive"

# Get project name from memory file path
get_project_name() {
    local mem_path="$1"
    # memory.md is at PROJECT/.claude/memory/memory.md
    # So go up 3 levels to get project dir
    local project_dir=$(dirname "$(dirname "$(dirname "$mem_path")")")
    basename "$project_dir"
}

# Helper: count lines in memory file
count_lines() {
    wc -l < "$MEMORY_FILE" 2>/dev/null || echo 0
}

# Helper: archive old dated notes (older than ARCHIVE_AFTER_DAYS)
auto_archive_old_notes() {
    if [[ ! -f "$MEMORY_FILE" ]]; then return; fi
    mkdir -p "$ARCHIVE_DIR"

    local cutoff_date=$(date -d "-${ARCHIVE_AFTER_DAYS} days" "+%Y-%m-%d" 2>/dev/null || date -v-${ARCHIVE_AFTER_DAYS}d "+%Y-%m-%d")
    local archive_file="${ARCHIVE_DIR}/$(date +%Y-%m).md"
    local archived=0
    local temp_file=$(mktemp)

    # Process file: archive old notes, keep recent ones
    while IFS= read -r line; do
        if [[ "$line" =~ ^-\ \*\*([0-9]{4}-[0-9]{2}-[0-9]{2}) ]]; then
            note_date="${BASH_REMATCH[1]}"
            # String comparison works for YYYY-MM-DD format
            if [[ "$note_date" < "$cutoff_date" ]]; then
                echo "$line" >> "$archive_file"
                archived=$((archived + 1))
            else
                echo "$line" >> "$temp_file"
            fi
        else
            echo "$line" >> "$temp_file"
        fi
    done < "$MEMORY_FILE"

    if [[ $archived -gt 0 ]]; then
        mv "$temp_file" "$MEMORY_FILE"
        echo "Auto-archived $archived old notes" >&2
    else
        rm -f "$temp_file"
    fi
}

case "$ACTION" in
    "load")
        if [[ -f "$MEMORY_FILE" ]]; then
            # Auto-archive notes older than threshold
            auto_archive_old_notes

            # Output memory
            PROJECT_NAME=$(get_project_name "$MEMORY_FILE")
            LAST_SESSION=$(grep -oP '(?<=\*\*Last Session:\*\* ).*' "$MEMORY_FILE" 2>/dev/null || echo 'N/A')

            echo "## Project Memory"
            echo ""
            echo "**Project:** $PROJECT_NAME"
            echo "**Last Session:** $LAST_SESSION"

            # Health check
            lines=$(count_lines)
            if [[ $lines -gt $MAX_LINES ]]; then
                echo ""
                echo "⚠️ Memory at $lines lines (max: $MAX_LINES) - auto-cleanup recommended"
            fi
        fi
        ;;

    "save")
        if [[ -f "$MEMORY_FILE" ]]; then
            TIMESTAMP=$(date -Iseconds)
            mkdir -p "$ARCHIVE_DIR"
            ARCHIVE_FILE="${ARCHIVE_DIR}/$(date +%Y-%m).md"

            # Auto-archive context notes before clearing
            note_count=$(grep -c "^- \*\*[0-9]" "$MEMORY_FILE" 2>/dev/null) || note_count=0
            if [[ $note_count -gt 0 ]]; then
                echo "" >> "$ARCHIVE_FILE"
                echo "### Session: $TIMESTAMP" >> "$ARCHIVE_FILE"
                grep "^- \*\*[0-9]" "$MEMORY_FILE" >> "$ARCHIVE_FILE" 2>/dev/null || true
                # Clear context notes
                sed -i '/^- \*\*[0-9]/d' "$MEMORY_FILE"
            fi

            # Update last session timestamp
            if grep -q "^\*\*Last Session:\*\*" "$MEMORY_FILE"; then
                sed -i "s/^\*\*Last Session:\*\*.*/\*\*Last Session:\*\* ${TIMESTAMP}/" "$MEMORY_FILE"
            elif grep -q "## Session History" "$MEMORY_FILE"; then
                sed -i "/## Session History/a \*\*Last Session:\*\* ${TIMESTAMP}" "$MEMORY_FILE"
            fi

            # Auto-cleanup if over limit
            lines=$(count_lines)
            if [[ $lines -gt $MAX_LINES ]]; then
                auto_archive_old_notes
            fi

            echo "Memory saved: ${MEMORY_FILE}" >&2
        fi
        ;;

    "set-focus")
        if [[ -n "$CONTENT" && -f "$MEMORY_FILE" ]]; then
            # Update current focus section
            if grep -q "## Current Focus" "$MEMORY_FILE"; then
                # Replace the line after Current Focus
                sed -i "/## Current Focus/{n;s/.*/- [ ] ${CONTENT}/}" "$MEMORY_FILE"
                echo "Focus updated" >&2
            fi
        fi
        ;;

    "add-note")
        if [[ -n "$CONTENT" && -f "$MEMORY_FILE" ]]; then
            TIMESTAMP=$(date "+%Y-%m-%d %H:%M")
            # Add to Context for Next Session
            if grep -q "## Context for Next Session" "$MEMORY_FILE"; then
                sed -i "/## Context for Next Session/a - **${TIMESTAMP}**: ${CONTENT}" "$MEMORY_FILE"
                echo "Note added to memory" >&2
            else
                echo "" >> "$MEMORY_FILE"
                echo "## Context for Next Session" >> "$MEMORY_FILE"
                echo "- **${TIMESTAMP}**: ${CONTENT}" >> "$MEMORY_FILE"
                echo "Note added to memory" >&2
            fi
        fi
        ;;

    "add-learning")
        if [[ -n "$CONTENT" && -f "$MEMORY_FILE" ]]; then
            # Add to Learnings section
            if grep -q "## Learnings & Gotchas" "$MEMORY_FILE"; then
                sed -i "/## Learnings & Gotchas/a - ${CONTENT}" "$MEMORY_FILE"
                echo "Learning added to memory" >&2
            else
                echo "" >> "$MEMORY_FILE"
                echo "## Learnings & Gotchas" >> "$MEMORY_FILE"
                echo "- ${CONTENT}" >> "$MEMORY_FILE"
                echo "Learning added to memory" >&2
            fi
        fi
        ;;

    "add-decision")
        if [[ -n "$CONTENT" && -f "$MEMORY_FILE" ]]; then
            DATE=$(date "+%Y-%m-%d")
            # Add to Key Decisions table
            if grep -q "## Key Decisions" "$MEMORY_FILE"; then
                # Find the last table row and add after it
                sed -i "/^| [0-9]\{4\}-/a | ${DATE} | ${CONTENT} | - |" "$MEMORY_FILE"
                echo "Decision added to memory" >&2
            fi
        fi
        ;;

    "archive")
        if [[ -f "$MEMORY_FILE" ]]; then
            ARCHIVE_DIR="$(dirname "$MEMORY_FILE")/archive"
            mkdir -p "$ARCHIVE_DIR"
            ARCHIVE_FILE="${ARCHIVE_DIR}/$(date +%Y-%m).md"

            # Extract "Context for Next Session" content and archive it
            if grep -q "## Context for Next Session" "$MEMORY_FILE"; then
                echo "" >> "$ARCHIVE_FILE"
                echo "### $(date '+%Y-%m-%d')" >> "$ARCHIVE_FILE"
                # Get lines between "Context for Next Session" and next "---" or "##"
                sed -n '/## Context for Next Session/,/^---$\|^## /p' "$MEMORY_FILE" | \
                    grep -v "^## Context\|^---$\|^\*Add notes via" >> "$ARCHIVE_FILE" 2>/dev/null || true

                # Clear the context section (keep header, remove notes)
                sed -i '/## Context for Next Session/,/^---$/{/^- \*\*/d}' "$MEMORY_FILE"
                echo "Archived to: $ARCHIVE_FILE" >&2
            fi
        fi
        ;;

    "prune")
        if [[ -f "$MEMORY_FILE" ]]; then
            lines=$(count_lines)
            learnings=$(grep -c "^- \*\*[A-Z]" "$MEMORY_FILE" 2>/dev/null) || learnings=0
            notes=$(grep -c "^- \*\*[0-9]" "$MEMORY_FILE" 2>/dev/null) || notes=0
            echo "Memory: $lines lines (max: $MAX_LINES)"
            echo "Learnings: $learnings (max: $MAX_LEARNINGS)"
            echo "Notes: $notes (max: $MAX_NOTES)"
            if [[ $lines -gt $MAX_LINES ]]; then
                echo "⚠️ Over limit - run 'auto-clean' or session end will auto-archive"
            else
                echo "✓ Healthy"
            fi
        fi
        ;;

    "auto-clean")
        if [[ -f "$MEMORY_FILE" ]]; then
            echo "Running auto-cleanup..." >&2

            # 1. Archive all dated notes
            auto_archive_old_notes

            # 2. If still over limit, trim learnings to MAX_LEARNINGS (keep newest)
            learning_count=$(grep -c "^- \*\*[A-Z]" "$MEMORY_FILE" 2>/dev/null) || learning_count=0
            if [[ $learning_count -gt $MAX_LEARNINGS ]]; then
                mkdir -p "$ARCHIVE_DIR"
                archive_file="${ARCHIVE_DIR}/$(date +%Y-%m).md"
                excess=$((learning_count - MAX_LEARNINGS))

                echo "" >> "$archive_file"
                echo "### Archived Learnings: $(date '+%Y-%m-%d')" >> "$archive_file"

                # Archive oldest learnings (first N matches)
                grep -n "^- \*\*[A-Z]" "$MEMORY_FILE" | head -n "$excess" | while IFS=: read -r num line; do
                    echo "$line" >> "$archive_file"
                    sed -i "${num}d" "$MEMORY_FILE"
                done

                echo "Archived $excess old learnings" >&2
            fi

            echo "Cleanup complete. $(count_lines) lines remaining." >&2
        fi
        ;;

    *)
        echo "Usage: $0 {load|save|add-note|add-learning|add-decision|set-focus|archive|prune|auto-clean} [content]" >&2
        exit 1
        ;;
esac
