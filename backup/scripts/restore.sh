#!/bin/bash
# Restore PostgreSQL from pgBackRest backup
# Usage: ./restore.sh [backup-label]
#   backup-label: Optional specific backup to restore (default: latest)
set -eo pipefail

BACKUP_LABEL="${1:-}"
LOG_FILE="/var/log/pgbackrest/restore.log"

log() {
    echo "$(date -Iseconds) [RESTORE] $1" | tee -a "$LOG_FILE"
}

error() {
    echo "$(date -Iseconds) [ERROR] $1" | tee -a "$LOG_FILE" >&2
}

show_usage() {
    echo "Usage: $0 [backup-label]"
    echo ""
    echo "Arguments:"
    echo "  backup-label    Optional. Specific backup to restore."
    echo "                  If not provided, restores the latest backup."
    echo ""
    echo "Examples:"
    echo "  $0                           # Restore latest backup"
    echo "  $0 20250115-030000F          # Restore specific full backup"
    echo "  $0 20250116-030000I          # Restore specific incremental"
    echo ""
    echo "Available backups:"
    pgbackrest --stanza=main info 2>/dev/null || echo "  (Unable to list backups)"
}

confirm_restore() {
    echo ""
    echo "WARNING: This will restore the PostgreSQL database."
    echo "The current data directory will be replaced."
    echo ""

    if [ -n "$BACKUP_LABEL" ]; then
        echo "Restoring backup: $BACKUP_LABEL"
    else
        echo "Restoring: LATEST backup"
    fi

    echo ""
    read -p "Are you sure you want to proceed? (yes/no): " confirm

    if [ "$confirm" != "yes" ]; then
        log "Restore cancelled by user"
        exit 0
    fi
}

main() {
    log "Starting restore process..."

    # Show help if requested
    if [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
        show_usage
        exit 0
    fi

    # Check if backups exist
    local backup_count
    backup_count=$(pgbackrest --stanza=main info --output=json 2>/dev/null | \
        jq -r '.[0].backup | length' 2>/dev/null || echo "0")

    if [ "$backup_count" = "0" ] || [ -z "$backup_count" ]; then
        error "No backups available to restore"
        exit 1
    fi

    # Confirm with user (skip if non-interactive)
    if [ -t 0 ]; then
        confirm_restore
    else
        log "Running in non-interactive mode, proceeding with restore..."
    fi

    # Build restore command
    local restore_cmd="pgbackrest --stanza=main restore --delta"

    if [ -n "$BACKUP_LABEL" ]; then
        restore_cmd="$restore_cmd --set=$BACKUP_LABEL"
        log "Restoring backup: $BACKUP_LABEL"
    else
        log "Restoring latest backup..."
    fi

    # Execute restore
    local start_time
    start_time=$(date +%s)

    log "Executing: $restore_cmd"

    if eval "$restore_cmd"; then
        local end_time
        end_time=$(date +%s)
        local duration=$((end_time - start_time))

        log "Restore completed successfully in ${duration}s"
        echo ""
        echo "RESTORE COMPLETE"
        echo ""
        echo "Next steps:"
        echo "1. Start PostgreSQL: make prod (or docker compose up -d postgres)"
        echo "2. Verify data integrity"
        echo "3. Resume normal operations"
    else
        error "Restore failed!"
        exit 1
    fi
}

main "$@"
