#!/bin/bash
# Health check for pgBackRest backup service
set -e

# Maximum age of last successful backup (26 hours = 1 day + 2h buffer)
MAX_BACKUP_AGE_HOURS=26

# Check PostgreSQL connection
check_postgres() {
    if ! pg_isready -h "$PGHOST" -p "${PGPORT:-5432}" -U "$PGUSER" > /dev/null 2>&1; then
        echo "UNHEALTHY: Cannot connect to PostgreSQL"
        return 1
    fi
    return 0
}

# Check S3 repository accessibility
check_s3() {
    if ! pgbackrest --stanza=main repo-ls / > /dev/null 2>&1; then
        echo "UNHEALTHY: Cannot access S3 repository"
        return 1
    fi
    return 0
}

# Check last backup age
check_backup_age() {
    # Get info about backups
    local info
    info=$(pgbackrest --stanza=main info --output=json 2>/dev/null) || {
        # No backups yet is OK for initial setup (first 26 hours)
        echo "WARNING: No backup info available yet"
        return 0
    }

    # Parse last backup timestamp using jq
    local last_backup
    last_backup=$(echo "$info" | jq -r '.[0].backup[-1].timestamp.stop // empty' 2>/dev/null)

    if [ -z "$last_backup" ]; then
        # Check if this is a fresh install (stanza exists but no backups)
        local stanza_status
        stanza_status=$(echo "$info" | jq -r '.[0].status.message // "error"' 2>/dev/null)

        if [ "$stanza_status" = "ok" ]; then
            echo "WARNING: Stanza OK but no backups yet"
            return 0
        fi
        return 0
    fi

    # Calculate age in hours
    local backup_epoch
    backup_epoch=$(date -d "$last_backup" +%s 2>/dev/null) || {
        echo "WARNING: Cannot parse backup timestamp"
        return 0
    }

    local current_epoch
    current_epoch=$(date +%s)

    local age_hours
    age_hours=$(( (current_epoch - backup_epoch) / 3600 ))

    if [ "$age_hours" -gt "$MAX_BACKUP_AGE_HOURS" ]; then
        echo "UNHEALTHY: Last backup is ${age_hours}h old (max: ${MAX_BACKUP_AGE_HOURS}h)"
        return 1
    fi

    return 0
}

# Run all checks
main() {
    check_postgres || exit 1
    check_s3 || exit 1
    check_backup_age || exit 1

    echo "HEALTHY"
    exit 0
}

main
