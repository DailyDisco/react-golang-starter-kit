# PostgreSQL 17 with pgBackRest for Production Backups
# Extends official postgres:17-alpine with backup capabilities

FROM postgres:17-alpine

# Install pgBackRest
RUN apk add --no-cache \
    pgbackrest \
    bash

# Create pgBackRest directories
RUN mkdir -p \
    /etc/pgbackrest \
    /var/lib/pgbackrest \
    /var/log/pgbackrest \
    /var/spool/pgbackrest \
    && chown -R postgres:postgres \
    /etc/pgbackrest \
    /var/lib/pgbackrest \
    /var/log/pgbackrest \
    /var/spool/pgbackrest

# Default entrypoint from postgres image
