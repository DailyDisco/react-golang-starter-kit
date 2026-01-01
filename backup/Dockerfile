# pgBackRest Backup Container for PostgreSQL 17
# Industry-standard backup solution with S3 integration

FROM postgres:17-alpine

# Install pgBackRest and dependencies
RUN apk add --no-cache \
    pgbackrest \
    dcron \
    bash \
    curl \
    jq \
    gettext \
    tzdata \
    openssh-client

# Set timezone
ENV TZ=UTC

# Create required directories
RUN mkdir -p \
    /etc/pgbackrest \
    /var/lib/pgbackrest \
    /var/log/pgbackrest \
    /var/spool/pgbackrest \
    /scripts \
    /var/spool/cron/crontabs

# Copy scripts
COPY scripts/ /scripts/
RUN chmod +x /scripts/*.sh

# Copy crontab
COPY crontabs/backup /var/spool/cron/crontabs/root
RUN chmod 0600 /var/spool/cron/crontabs/root

# Copy pgbackrest config template
COPY pgbackrest.conf.template /etc/pgbackrest/pgbackrest.conf.template

# Set ownership
RUN chown -R postgres:postgres /var/lib/pgbackrest /var/log/pgbackrest /var/spool/pgbackrest

# Health check
HEALTHCHECK --interval=5m --timeout=30s --start-period=2m --retries=3 \
    CMD /scripts/healthcheck.sh

# Entrypoint handles config generation and stanza setup
ENTRYPOINT ["/scripts/entrypoint.sh"]

# Default command: run crond in foreground
CMD ["crond", "-f", "-l", "2"]
