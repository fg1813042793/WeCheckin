#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
BACKEND_DIR="${ROOT_DIR}/backend"
GOCACHE_DIR="${ROOT_DIR}/.cache/go-build"

enabled() {
  case "${1:-}" in
    1 | true | TRUE | yes | YES | on | ON) return 0 ;;
    *) return 1 ;;
  esac
}

cleanup() {
  rm -rf "${ROOT_DIR}/.cache" >/dev/null 2>&1 || true
}

trap cleanup EXIT

cd "${ROOT_DIR}"

echo "==> Running deployment config checks"
node scripts/check-deploy-config.mjs

echo "==> Running backend checks"
cd "${BACKEND_DIR}"
GOCACHE="${GOCACHE_DIR}" go test \
  ./cmd \
  ./pkg/tokenutil \
  ./internal/handler \
  ./internal/service \
  ./internal/config \
  ./internal/formkit/... \
  ./test/internal/...

cd "${ROOT_DIR}"

echo "==> Running frontend config checks"
npm --prefix frontend run check:config
npm --prefix frontend run check:request
npm --prefix frontend run check:logs
npm --prefix frontend run check:auth
npm --prefix frontend run check:formkit-logic

echo "==> Running admin request checks"
npm --prefix admin run check:request
npm --prefix admin run check:navigation
npm --prefix admin run check:build-config
npm --prefix admin run check:ui-shell
npm --prefix admin run check:user-list-ui
npm --prefix admin run check:p2-ui
npm --prefix admin run check:icon-runtime

if enabled "${CHECK_FRONTEND_BUILD:-${CHECK_BUILDS:-0}}"; then
  echo "==> Running frontend H5 build"
  npm --prefix frontend run build:h5
fi

if enabled "${CHECK_ADMIN_BUILD:-${CHECK_BUILDS:-0}}"; then
  echo "==> Running admin build"
  npm --prefix admin run build
fi

if enabled "${CHECK_PERFORMANCE:-0}"; then
  echo "==> Running performance baseline checks"
  npm run check:performance
fi

echo "==> Checks passed"
