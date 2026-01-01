#!/bin/bash
# Run an incremental pgBackRest backup
set -eo pipefail

LOG_FILE="/var/log/pgbackrest/backup.log"

log() {
    echo "$(date -Iseconds) [INCR-BACKUP] $1" | tee -a "$LOG_FILE"
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
    log "Starting incremental backup..."

    # Check if any full backup exists
    local backup_count
    backup_count=$(pgbackrest --stanza=main info --output=json 2>/dev/null | \
        jq -r '.[0].backup | length' 2>/dev/null || echo "0")

    if [ "$backup_count" = "0" ] || [ -z "$backup_count" ]; then
        log "No existing backups found, running full backup instead"
        exec /scripts/backup-full.sh
    fi

    # Run incremental backup
    local start_time
    start_time=$(date +%s)

    if pgbackrest --stanza=main --type=incr backup; then
        local end_time
        end_time=$(date +%s)
        local duration=$((end_time - start_time))

        log "Incremental backup completed successfully in ${duration}s"

        # Get backup info
        local backup_info
        backup_info=$(pgbackrest --stanza=main info --output=json | jq -r '.[0].backup[-1]' 2>/dev/null)

        local backup_label
        backup_label=$(echo "$backup_info" | jq -r '.label' 2>/dev/null)

        local backup_size
        backup_size=$(echo "$backup_info" | jq -r '.info.delta' 2>/dev/null)

        # Convert bytes to human readable
        local size_mb=$((backup_size / 1024 / 1024))

        log "Backup: $backup_label (delta: ${size_mb}MB)"
        notify "SUCCESS" "Incremental backup completed: $backup_label (delta: ${size_mb}MB) in ${duration}s"
    else
        log "ERROR: Incremental backup failed"
        notify "FAILED" "Incremental backup failed at $(date -Iseconds)"
        exit 1
    fi

    log "Incremental backup process completed"
}

main "$@"
