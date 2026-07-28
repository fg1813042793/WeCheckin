#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

cd "$ROOT_DIR"
node scripts/check-quality-gates.mjs
GOCACHE="$ROOT_DIR/.cache/go-build" go test ./backend/...

cd "$ROOT_DIR/admin"
npm run check:all

cd "$ROOT_DIR/frontend"
npm run check:all
