# Config Environment Override Design

## Goal

Make backend configuration safer and easier to run across machines by allowing environment variables to override file-based defaults, while preserving the existing YAML configuration flow.

## Scope

This pass covers:

- Adding explicit environment variable overrides for backend server, database, Redis, CORS, log, and token settings that are already modeled in `backend/internal/config/config.go`.
- Adding tests for environment override behavior in the config package.
- Adding a safe example configuration file for reference.
- Updating README guidance so sensitive values can be supplied through environment variables.

This pass does not:

- Remove or rewrite the existing `backend/config/config.yaml`, `config.dev.yaml`, or `config.local.yaml` files.
- Change database initialization, migrations, seed behavior, or `DROP TABLE` behavior.
- Change Docker networking or introduce a new secrets manager.
- Change frontend request configuration.

## Current Observations

- `LoadConfig(env string)` currently reads `config.yaml`, then optionally merges `config.<env>.yaml`.
- Config defaults are defined through Viper, but there is no environment variable binding.
- Local config files contain machine-specific database and Redis credentials.
- `Config` currently models server, database, CORS, log, Redis, and token settings. Other YAML keys such as `oss`, `storage`, and `security` are present in files but are not part of the `Config` struct.

## Design

### Environment Variables

Add a `WECHECKIN_` environment variable namespace. File config remains the baseline; environment variables override after files are read and before unmarshalling.

Supported variables:

- `WECHECKIN_SERVER_PORT`
- `WECHECKIN_SERVER_HOST`
- `WECHECKIN_SERVER_MODE`
- `WECHECKIN_DATABASE_HOST`
- `WECHECKIN_DATABASE_PORT`
- `WECHECKIN_DATABASE_USER`
- `WECHECKIN_DATABASE_PASSWORD`
- `WECHECKIN_DATABASE_DBNAME`
- `WECHECKIN_REDIS_HOST`
- `WECHECKIN_REDIS_PORT`
- `WECHECKIN_REDIS_PASSWORD`
- `WECHECKIN_REDIS_DB`
- `WECHECKIN_LOG_DIR`
- `WECHECKIN_LOG_LEVEL`
- `WECHECKIN_LOG_MAX_AGE`
- `WECHECKIN_LOG_COMPRESS`
- `WECHECKIN_TOKEN_USER_EXPIRE`
- `WECHECKIN_TOKEN_USER_REDIS_PREFIX`
- `WECHECKIN_TOKEN_ADMIN_EXPIRE`
- `WECHECKIN_TOKEN_ADMIN_REDIS_PREFIX`
- `WECHECKIN_CORS_ALLOW_ORIGINS`
- `WECHECKIN_CORS_ALLOW_METHODS`
- `WECHECKIN_CORS_ALLOW_HEADERS`

Comma-separated values are acceptable for CORS list settings.

### Config Package Tests

Add tests in `backend/internal/config/config_test.go` that use an isolated temporary directory. The tests should verify:

- File values still load when no environment overrides are present.
- Environment variables override YAML values for strings, integers, booleans, and list settings.

Tests must use `t.Setenv` so they do not leak process environment changes.

### Example Config

Add `backend/config/config.example.yaml` with safe placeholder values:

- local hostnames and default ports
- placeholder passwords
- non-secret OSS/security placeholder values

This file is documentation and a starting point, not a replacement for existing local config files.

### Documentation

Update README to describe:

- YAML files provide defaults.
- `WECHECKIN_` environment variables override YAML values.
- Sensitive values such as database and Redis passwords should be supplied through environment variables when possible.
- A short example command showing `WECHECKIN_DATABASE_PASSWORD=... go run cmd/main.go`.

## Testing

Run:

```bash
GOCACHE=$PWD/.cache/go-build go test ./backend/internal/config ./backend/internal/app/formkit/...
```

Expected result: config package tests and existing formkit tests pass.

If `.cache/` is created, remove it after verification.

## Risks

- Viper instances are global by default. The implementation must reset or isolate Viper state enough that tests are deterministic.
- Environment variable parsing for list values must remain predictable. Use comma-separated values and document that convention.
- Existing local YAML files are left in place to avoid breaking current development startup, so this pass improves override safety without fully removing committed secrets.
