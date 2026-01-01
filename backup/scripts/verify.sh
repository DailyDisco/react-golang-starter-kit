#!/bin/bash
# Verify pgBackRest backup integrity
set -eo pipefail

LOG_FILE="/var/log/pgbackrest/backup.log"

log() {
    echo "$(date -Iseconds) [VERIFY] $1" | tee -a "$LOG_FILE"
}

notify() {
    local status="$1"
    local message="$2"

    if [ -n "$BACKUP_SLACK_WEBHOOK" ]; then
        curl -s -X POST -H 'Content-type: application/json' \
            --data "{\"text\":\"[${status}] Backup Verification: ${message}\"}" \
            "$BACKUP_SLACK_WEBHOOK" || true
    fi
}

main() {
    log "Starting backup verification..."

    # Check if any backups exist
    local backup_count
    backup_count=$(pgbackrest --stanza=main info --output=json 2>/dev/null | \
        jq -r '.[0].backup | length' 2>/dev/null || echo "0")

    if [ "$backup_count" = "0" ] || [ -z "$backup_count" ]; then
        log "No backups to verify"
        exit 0
    fi

    # Run verification
    local start_time
    start_time=$(date +%s)

    if pgbackrest --stanza=main verify; then
        local end_time
        end_time=$(date +%s)
        local duration=$((end_time - start_time))

        log "Verification completed successfully in ${duration}s"
        notify "SUCCESS" "Backup verification passed in ${duration}s"
    else
        log "ERROR: Verification failed"
        notify "FAILED" "Backup verification failed - check logs"
        exit 1
    fi

    # Show backup summary
    log "Current backup inventory:"
    pgbackrest --stanza=main info | tee -a "$LOG_FILE"

    log "Verification process completed"
}

main "$@"
