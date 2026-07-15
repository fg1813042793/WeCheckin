#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

if [ $# -ne 1 ]; then
  echo "Usage: $0 /path/to/wecheckin-backup.sql"
  exit 1
fi

SQL_FILE="$1"
if [ ! -f "${SQL_FILE}" ]; then
  echo "Backup file not found: ${SQL_FILE}"
  exit 1
fi

echo "This will restore ${SQL_FILE} into the Docker Compose MySQL database."
echo "Existing data in the target database may be overwritten."
printf "TYPE RESTORE TO CONTINUE: "
read -r CONFIRM
if [ "${CONFIRM}" != "RESTORE" ]; then
  echo "Restore cancelled."
  exit 1
fi

cd "${BACKEND_DIR}"
docker compose exec -T mysql sh -c 'mysql -u"$MYSQL_USER" -p"$MYSQL_PASSWORD" "$MYSQL_DATABASE"' < "${SQL_FILE}"
echo "Database restore completed from ${SQL_FILE}"
