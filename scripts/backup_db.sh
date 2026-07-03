#!/bin/sh
set -e

BACKUP_DIR="/backups"
mkdir -p "$BACKUP_DIR"

TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_FILE="$BACKUP_DIR/tivri_backup_$TIMESTAMP.sql"

pg_dump -h db -U "$POSTGRES_USER" -d "$POSTGRES_DB" -F c -b -v -f "$BACKUP_FILE"
