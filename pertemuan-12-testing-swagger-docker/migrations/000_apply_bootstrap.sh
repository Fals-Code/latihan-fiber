#!/bin/sh
# PostgreSQL official image loads this script before root-level 005_*.sql.
# The bootstrap/ directory is NOT recursively executed by the entrypoint.
# Safe for both sourced (non-executable) and executable .sh files.
(
    set -eu
    export PGPASSWORD="${POSTGRES_PASSWORD:-}"
    for file in /docker-entrypoint-initdb.d/bootstrap/*.sql; do
        [ -f "$file" ] || continue
        psql --no-psqlrc -v ON_ERROR_STOP=1 \
            -U "${POSTGRES_USER:-postgres}" \
            -d "${POSTGRES_DB:-postgres}" -f "$file"
    done
)
