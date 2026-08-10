# Backend Test Layout

`backend/test` stores tests that validate cross-package architecture, route layout, permission wiring, or integration behavior.

Keep package-local unit tests beside their implementation under `backend/internal` when they need unexported helpers, fixtures, or tight package context.

Current convention:

- `backend/test/internal/...`: architecture guards that mirror `backend/internal/...` source directories.
- `backend/test/internal/bootstrap/migrations`: migration SQL and startup migration wiring guards.
- `backend/test/internal/bootstrap/quality`: repository script, docs, UI structure, and seed-data quality guards.
- `backend/internal/...`: package-local unit tests and behavior tests.
