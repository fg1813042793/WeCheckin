#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

cd "$ROOT_DIR/backend"

unformatted="$(gofmt -l cmd internal pkg test)"
if [[ -n "$unformatted" ]]; then
  printf 'Backend Go files are not formatted:\n%s\n' "$unformatted" >&2
  exit 1
fi

echo "Backend gofmt check passed."
