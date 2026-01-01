#!/bin/bash
# pgBackRest Backup Container Entrypoint
# Handles configuration, stanza setup, and starts cron scheduler
set -e

log() {
    echo "$(date -Iseconds) [BACKUP] $1"
}

error() {
    echo "$(date -Iseconds) [ERROR] $1" >&2
}

# ============================================
# Validate Required Environment Variables
# ============================================
validate_env() {
    local required_vars=(
        "BACKUP_S3_BUCKET"
        "AWS_ACCESS_KEY_ID"
        "AWS_SECRET_ACCESS_KEY"
        "BACKUP_ENCRYPTION_KEY"
        "PGHOST"
        "PGUSER"
        "PGPASSWORD"
        "PGDATABASE"
    )

    local missing=()
    for var in "${required_vars[@]}"; do
        if [ -z "${!var}" ]; then
            missing+=("$var")
        fi
    done

    if [ ${#missing[@]} -gt 0 ]; then
        error "Missing required environment variables: ${missing[*]}"
        exit 1
    fi
}

# ============================================
# Generate pgBackRest Configuration
# ============================================
generate_config() {
    log "Generating pgBackRest configuration..."

    # Use envsubst to replace environment variables in template
    envsubst < /etc/pgbackrest/pgbackrest.conf.template > /etc/pgbackrest/pgbackrest.conf

    # Also copy to shared config volume for postgres container
    if [ -d /shared-config ]; then
        cp /etc/pgbackrest/pgbackrest.conf /shared-config/pgbackrest.conf
        log "Configuration copied to shared volume"
    fi

    chmod 640 /etc/pgbackrest/pgbackrest.conf
    log "Configuration generated successfully"
}

# ============================================
# Wait for PostgreSQL
# ============================================
wait_for_postgres() {
    log "Waiting for PostgreSQL at ${PGHOST}:${PGPORT:-5432}..."

    local retries=30
    local count=0

    while [ $count -lt $retries ]; do
        if pg_isready -h "$PGHOST" -p "${PGPORT:-5432}" -U "$PGUSER" -d "$PGDATABASE" > /dev/null 2>&1; then
            log "PostgreSQL is ready"
            return 0
        fi
        count=$((count + 1))
        log "Waiting for PostgreSQL... ($count/$retries)"
        sleep 2
    done

    error "PostgreSQL not available after $retries attempts"
    exit 1
}

# ============================================
# Test S3 Connection
# ============================================
test_s3_connection() {
    log "Testing S3 connection to bucket: ${BACKUP_S3_BUCKET}..."

    # pgBackRest will create the path if it doesn't exist
    # Just verify we can reach S3 by checking the repo
    if pgbackrest --stanza=main repo-ls / > /dev/null 2>&1; then
        log "S3 connection successful"
    else
        log "S3 repository not yet initialized (this is normal for first run)"
    fi
}

# ============================================
# Initialize Stanza
# ============================================
init_stanza() {
    log "Checking stanza 'main'..."

    # Check if stanza already exists
    if pgbackrest --stanza=main info > /dev/null 2>&1; then
        log "Stanza 'main' already exists"
        pgbackrest --stanza=main info
    else
        log "Creating stanza 'main'..."

        # Create the stanza
        if pgbackrest --stanza=main stanza-create; then
            log "Stanza 'main' created successfully"
        else
            error "Failed to create stanza. This may be normal if archive_mode is not yet enabled on PostgreSQL."
            log "Will retry stanza creation on first backup"
        fi
    fi
}

# ============================================
# Main
# ============================================
main() {
    log "Starting pgBackRest backup service..."
    log "Version: $(pgbackrest version)"

    # Step 1: Validate environment
    validate_env

    # Step 2: Generate configuration
    generate_config

    # Step 3: Wait for PostgreSQL
    wait_for_postgres

    # Step 4: Test S3 connection
    test_s3_connection

    # Step 5: Initialize stanza (may fail on first run, that's OK)
    init_stanza || true

    # Display schedule info
    log "Backup schedule:"
    log "  - Full backup: Sunday 03:00 UTC"
    log "  - Incremental backup: Daily 03:00 UTC (Mon-Sat)"
    log "  - Verification: Monday 04:00 UTC"
    log ""
    log "Backup service initialized. Starting cron scheduler..."

    # Execute CMD (crond)
    exec "$@"
}

main "$@"
