#!/bin/bash
# Run a full pgBackRest backup
set -eo pipefail

LOG_FILE="/var/log/pgbackrest/backup.log"

log() {
    echo "$(date -Iseconds) [FULL-BACKUP] $1" | tee -a "$LOG_FILE"
}

notify() {
    local status="$1"
    local message="$2"

    if [ -n "$BACKUP_SLACK_WEBHOOK" ]; then
        curl -s -X POST -H 'Content-type: application/json' \
            --data "{\"text\":\"[${status}] PostgreSQL Backup: ${message}\"}" \
            "$BACKUP_SLACK_WEBHOOK" || true
    fi
}

main() {
    log "Starting full backup..."

    # Ensure stanza exists
    if ! pgbackrest --stanza=main info > /dev/null 2>&1; then
        log "Stanza not found, attempting to create..."
        pgbackrest --stanza=main stanza-create || {
            log "ERROR: Failed to create stanza"
            notify "FAILED" "Full backup failed - stanza creation error"
            exit 1
        }
    fi

    # Run full backup
    local start_time
    start_time=$(date +%s)

    if pgbackrest --stanza=main --type=full backup; then
        local end_time
        end_time=$(date +%s)
        local duration=$((end_time - start_time))

        log "Full backup completed successfully in ${duration}s"

        # Get backup info
        local backup_info
        backup_info=$(pgbackrest --stanza=main info --output=json | jq -r '.[0].backup[-1]' 2>/dev/null)

        local backup_label
        backup_label=$(echo "$backup_info" | jq -r '.label' 2>/dev/null)

        local backup_size
        backup_size=$(echo "$backup_info" | jq -r '.info.size' 2>/dev/null)

        # Convert bytes to human readable
        local size_mb=$((backup_size / 1024 / 1024))

        log "Backup: $backup_label (${size_mb}MB)"
        notify "SUCCESS" "Full backup completed: $backup_label (${size_mb}MB) in ${duration}s"
    else
        log "ERROR: Full backup failed"
        notify "FAILED" "Full backup failed at $(date -Iseconds)"
        exit 1
    fi

    # Run retention cleanup
    log "Running retention cleanup..."
    pgbackrest --stanza=main expire || log "WARNING: Retention cleanup had issues"

    log "Full backup process completed"
}

main "$@"
