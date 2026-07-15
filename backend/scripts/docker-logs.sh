#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
SERVICE="${1:-backend}"
TAIL="${TAIL:-200}"

cd "${BACKEND_DIR}"

echo "Following ${SERVICE} logs, tail=${TAIL}. Press Ctrl+C to stop."
docker compose logs --tail "${TAIL}" -f "${SERVICE}"
