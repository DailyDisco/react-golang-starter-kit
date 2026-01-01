#!/bin/bash
# Point-in-Time Recovery using pgBackRest
# Usage: ./restore-pitr.sh <target-time>
#   target-time: ISO 8601 timestamp (e.g., "2025-01-15 14:30:00")
set -eo pipefail

TARGET_TIME="${1:-}"
LOG_FILE="/var/log/pgbackrest/restore.log"

log() {
    echo "$(date -Iseconds) [PITR] $1" | tee -a "$LOG_FILE"
}

error() {
    echo "$(date -Iseconds) [ERROR] $1" | tee -a "$LOG_FILE" >&2
}

show_usage() {
    echo "Usage: $0 <target-time>"
    echo ""
    echo "Arguments:"
    echo "  target-time    Required. Timestamp to recover to."
    echo "                 Format: ISO 8601 (YYYY-MM-DD HH:MM:SS)"
    echo ""
    echo "Examples:"
    echo "  $0 \"2025-01-15 14:30:00\"     # Recover to specific time"
    echo "  $0 \"2025-01-15 14:30:00+00\"  # With timezone"
    echo ""
    echo "Available backups:"
    pgbackrest --stanza=main info 2>/dev/null || echo "  (Unable to list backups)"
    echo ""
    echo "Note: Recovery can only go back to the oldest available backup."
    echo "WAL archives between backups enable point-in-time recovery."
}

validate_timestamp() {
    local ts="$1"

    # Try to parse the timestamp
    if ! date -d "$ts" > /dev/null 2>&1; then
        error "Invalid timestamp format: $ts"
        echo "Expected format: YYYY-MM-DD HH:MM:SS"
        exit 1
    fi

    # Check if timestamp is in the future
    local target_epoch
    target_epoch=$(date -d "$ts" +%s)
    local current_epoch
    current_epoch=$(date +%s)

    if [ "$target_epoch" -gt "$current_epoch" ]; then
        error "Target time is in the future!"
        exit 1
    fi

    log "Target time validated: $ts"
}

confirm_restore() {
    echo ""
    echo "WARNING: Point-in-Time Recovery"
    echo ""
    echo "This will restore the database to: $TARGET_TIME"
    echo "All changes AFTER this time will be LOST."
    echo ""

    read -p "Are you sure you want to proceed? (yes/no): " confirm

    if [ "$confirm" != "yes" ]; then
        log "PITR cancelled by user"
        exit 0
    fi
}

main() {
    # Show help if requested or no arguments
    if [ -z "$TARGET_TIME" ] || [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
        show_usage
        exit 0
    fi

    log "Starting Point-in-Time Recovery..."
    log "Target time: $TARGET_TIME"

    # Validate timestamp
    validate_timestamp "$TARGET_TIME"

    # Check if backups exist
    local backup_count
    backup_count=$(pgbackrest --stanza=main info --output=json 2>/dev/null | \
        jq -r '.[0].backup | length' 2>/dev/null || echo "0")

    if [ "$backup_count" = "0" ] || [ -z "$backup_count" ]; then
        error "No backups available for PITR"
        exit 1
    fi

    # Confirm with user (skip if non-interactive)
    if [ -t 0 ]; then
        confirm_restore
    else
        log "Running in non-interactive mode, proceeding with PITR..."
    fi

    # Build PITR command
    local restore_cmd="pgbackrest --stanza=main restore"
    restore_cmd="$restore_cmd --delta"
    restore_cmd="$restore_cmd --type=time"
    restore_cmd="$restore_cmd --target=\"$TARGET_TIME\""
    restore_cmd="$restore_cmd --target-action=promote"

    # Execute PITR
    local start_time
    start_time=$(date +%s)

    log "Executing: $restore_cmd"

    if eval "$restore_cmd"; then
        local end_time
        end_time=$(date +%s)
        local duration=$((end_time - start_time))

        log "PITR completed successfully in ${duration}s"
        echo ""
        echo "POINT-IN-TIME RECOVERY COMPLETE"
        echo ""
        echo "Database restored to: $TARGET_TIME"
        echo ""
        echo "Next steps:"
        echo "1. Start PostgreSQL: make prod (or docker compose up -d postgres)"
        echo "2. PostgreSQL will replay WAL to reach target time"
        echo "3. Verify data integrity"
        echo "4. Create a new backup immediately after verification"
    else
        error "PITR failed!"
        exit 1
    fi
}

main "$@"
