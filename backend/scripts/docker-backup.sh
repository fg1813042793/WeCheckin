#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
BACKUP_DIR="${BACKUP_DIR:-${BACKEND_DIR}/backups}"
TIMESTAMP="$(date +%Y%m%d-%H%M%S)"
SQL_FILE="${BACKUP_DIR}/wecheckin-${TIMESTAMP}.sql"
UPLOADS_FILE="${BACKUP_DIR}/uploads-${TIMESTAMP}.tar.gz"

mkdir -p "${BACKUP_DIR}"
cd "${BACKEND_DIR}"

if ! docker compose ps mysql >/dev/null 2>&1; then
  echo "MySQL container is not available. Start it with: docker compose up -d mysql"
  exit 1
fi

docker compose exec -T mysql sh -c 'mysqldump --single-transaction --routines --triggers -u"$MYSQL_USER" -p"$MYSQL_PASSWORD" "$MYSQL_DATABASE"' > "${SQL_FILE}"
echo "Database backup written to ${SQL_FILE}"

if [ -d "${BACKEND_DIR}/uploads" ]; then
  tar -czf "${UPLOADS_FILE}" -C "${BACKEND_DIR}" uploads
  echo "Uploads backup written to ${UPLOADS_FILE}"
fi
