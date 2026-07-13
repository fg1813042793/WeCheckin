# Project Health Optimization Design

## Goal

Improve the project baseline so local development is less surprising and the existing formkit test suite becomes a useful signal again.

## Scope

This pass is intentionally small. It covers:

- Fixing the reproducible `location` formkit validation test failure.
- Aligning backend startup guidance with the current config path and server port.
- Updating the README so it reflects the current Go, Vue, admin, and uni-app structure.
- Tightening `.gitignore` for local caches, backend logs, runtime uploads, and built binaries.
- Verifying the formkit packages after the change.

This pass does not restructure backend modules, rewrite frontend UI, change database schema design, or modify the existing uncommitted survey and logic-engine business changes.

## Current Observations

- The backend entrypoint is `backend/cmd/main.go`, with runtime config under `backend/config/`.
- The default backend config currently uses port `8083`.
- The README still describes older stack details, including Vue 2 and Go 1.19-era guidance.
- `backend/start.sh` checks for `backend/config.yaml` and prints `8080`, which does not match the current layout.
- `go test ./backend/internal/app/formkit/...` currently fails at `TestLocation_Validate`: the test expects a non-map value to fail, while `LocationQuestion.Validate` accepts a non-empty string.
- The repository has generated/runtime artifacts that should be ignored for future changes, including Go build cache, backend logs, uploads, and built binaries.

## Design

### Location Validation

Use the test's stricter contract as the desired behavior for this pass: `location` answers should be structured objects. A required location answer must include at least one non-empty location field from the accepted object shape. A plain string should fail validation instead of being treated as a valid location.

`FormatValue` can continue formatting strings defensively for legacy display paths, because display tolerance does not need to imply write-time validation tolerance.

### Startup and Documentation

Update `backend/start.sh` so its config check points at `config/config.yaml` from inside the backend directory, and its startup message reports `http://localhost:8083`.

Update README to describe the current repository shape:

- Backend: Go + Hertz + GORM + MySQL + Redis.
- Admin: Vue 3 + Vite + Element Plus.
- Frontend: uni-app + Vue 3.
- Backend run command from the repository root or backend directory.
- Admin and frontend package scripts as they actually exist.

### Repository Hygiene

Extend `.gitignore` to exclude local/generated artifacts that should not be part of normal source changes:

- `.cache/`
- backend logs, including compressed rotated logs
- backend uploads/runtime export files
- backend built binaries such as `backend/bin/`, `backend/cmd/cmd`, and `cmd.exe`

Existing tracked files are not removed in this pass. The ignore changes are for future cleanliness only.

## Testing

Run:

```bash
GOCACHE=$PWD/.cache/go-build go test ./backend/internal/app/formkit/...
```

Expected result: all formkit packages pass. If this creates `.cache/`, remove it after verification so the working tree stays clean.

## Risks

- If any frontend caller still submits location answers as plain strings, stricter validation may reject those submissions. The current test suite already encodes the stricter object-only behavior, so this pass follows that contract.
- Ignore rules do not untrack files that are already tracked by Git. A separate cleanup pass can remove tracked runtime artifacts if desired.
