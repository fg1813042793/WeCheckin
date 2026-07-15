# Config Environment Override Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add tested `WECHECKIN_` environment variable overrides for backend YAML configuration and document safer local configuration practices.

**Architecture:** Keep the existing `LoadConfig(env string)` API and YAML merge order. Internally use an isolated Viper instance per load, bind explicit `WECHECKIN_` environment variables to existing config keys, and keep the package-level `Cfg` assignment for current callers.

**Tech Stack:** Go 1.24, Viper, YAML config, Go testing package, Markdown docs.

---

### Task 1: Write Config Override Tests

**Files:**
- Create: `backend/internal/config/config_test.go`
- Modify: none

- [ ] **Step 1: Add test helpers and YAML baseline test**

Create `backend/internal/config/config_test.go` with tests that write a temporary `config.yaml`, run `LoadConfig("")` inside that temp directory, and verify YAML values load without environment overrides.

```go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func withTempConfigDir(t *testing.T, yaml string) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(yaml), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	oldwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldwd); err != nil {
			t.Fatalf("restore wd: %v", err)
		}
	})
}

const testConfigYAML = `
server:
  port: "8083"
  host: "127.0.0.1"
  mode: "debug"
cors:
  allow_origins:
    - "http://localhost:3000"
  allow_methods:
    - "GET"
  allow_headers:
    - "Authorization"
database:
  host: "db.local"
  port: 3307
  user: "root"
  password: "from-file"
  dbname: "go_wecheckin"
log:
  dir: "./logs"
  level: "info"
  max_age: 30
  compress: true
redis:
  host: "redis.local"
  port: 6379
  password: "redis-file"
  db: 3
token:
  user:
    expire: "168h"
    redis_prefix: "user_token:"
  admin:
    expire: "24h"
    redis_prefix: "admin_token:"
`

func TestLoadConfigUsesYAMLValues(t *testing.T) {
	withTempConfigDir(t, testConfigYAML)

	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	if cfg.Server.Port != "8083" {
		t.Fatalf("server port = %q", cfg.Server.Port)
	}
	if cfg.Database.Password != "from-file" {
		t.Fatalf("database password = %q", cfg.Database.Password)
	}
	if cfg.Redis.DB != 3 {
		t.Fatalf("redis db = %d", cfg.Redis.DB)
	}
	if got := cfg.CORS.AllowOrigins; len(got) != 1 || got[0] != "http://localhost:3000" {
		t.Fatalf("cors origins = %#v", got)
	}
}
```

- [ ] **Step 2: Add environment override test**

Append a test that sets representative string, int, bool, and list overrides.

```go
func TestLoadConfigAllowsEnvironmentOverrides(t *testing.T) {
	withTempConfigDir(t, testConfigYAML)
	t.Setenv("WECHECKIN_SERVER_PORT", "18083")
	t.Setenv("WECHECKIN_DATABASE_HOST", "db.env")
	t.Setenv("WECHECKIN_DATABASE_PORT", "3310")
	t.Setenv("WECHECKIN_DATABASE_PASSWORD", "from-env")
	t.Setenv("WECHECKIN_REDIS_DB", "5")
	t.Setenv("WECHECKIN_REDIS_PASSWORD", "redis-env")
	t.Setenv("WECHECKIN_LOG_COMPRESS", "false")
	t.Setenv("WECHECKIN_CORS_ALLOW_ORIGINS", "http://one.local,http://two.local")
	t.Setenv("WECHECKIN_TOKEN_ADMIN_EXPIRE", "48h")

	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	if cfg.Server.Port != "18083" {
		t.Fatalf("server port = %q", cfg.Server.Port)
	}
	if cfg.Database.Host != "db.env" {
		t.Fatalf("database host = %q", cfg.Database.Host)
	}
	if cfg.Database.Port != 3310 {
		t.Fatalf("database port = %d", cfg.Database.Port)
	}
	if cfg.Database.Password != "from-env" {
		t.Fatalf("database password = %q", cfg.Database.Password)
	}
	if cfg.Redis.DB != 5 {
		t.Fatalf("redis db = %d", cfg.Redis.DB)
	}
	if cfg.Redis.Password != "redis-env" {
		t.Fatalf("redis password = %q", cfg.Redis.Password)
	}
	if cfg.Log.Compress {
		t.Fatalf("log compress = true")
	}
	wantOrigins := []string{"http://one.local", "http://two.local"}
	if len(cfg.CORS.AllowOrigins) != len(wantOrigins) || cfg.CORS.AllowOrigins[0] != wantOrigins[0] || cfg.CORS.AllowOrigins[1] != wantOrigins[1] {
		t.Fatalf("cors origins = %#v", cfg.CORS.AllowOrigins)
	}
	if cfg.Token.Admin.Expire != "48h" {
		t.Fatalf("admin token expire = %q", cfg.Token.Admin.Expire)
	}
}
```

- [ ] **Step 3: Run tests and verify RED**

Run:

```bash
GOCACHE=$PWD/.cache/go-build go test ./backend/internal/config -count=1
```

Expected: `TestLoadConfigAllowsEnvironmentOverrides` fails because environment overrides are not implemented yet.

### Task 2: Implement Environment Overrides

**Files:**
- Modify: `backend/internal/config/config.go`
- Test: `backend/internal/config/config_test.go`

- [ ] **Step 1: Refactor LoadConfig to use a local Viper instance**

Change `LoadConfig` to create `v := viper.New()` and replace package-level `viper.` calls inside the function with `v.` calls.

- [ ] **Step 2: Add explicit environment bindings**

Add helper functions in `config.go`:

```go
func setDefaults(v *viper.Viper) {
	// move existing defaults here
}

func bindEnv(v *viper.Viper) {
	bindings := map[string]string{
		"server.port": "WECHECKIN_SERVER_PORT",
		"server.host": "WECHECKIN_SERVER_HOST",
		"server.mode": "WECHECKIN_SERVER_MODE",
		"database.host": "WECHECKIN_DATABASE_HOST",
		"database.port": "WECHECKIN_DATABASE_PORT",
		"database.user": "WECHECKIN_DATABASE_USER",
		"database.password": "WECHECKIN_DATABASE_PASSWORD",
		"database.dbname": "WECHECKIN_DATABASE_DBNAME",
		"redis.host": "WECHECKIN_REDIS_HOST",
		"redis.port": "WECHECKIN_REDIS_PORT",
		"redis.password": "WECHECKIN_REDIS_PASSWORD",
		"redis.db": "WECHECKIN_REDIS_DB",
		"log.dir": "WECHECKIN_LOG_DIR",
		"log.level": "WECHECKIN_LOG_LEVEL",
		"log.max_age": "WECHECKIN_LOG_MAX_AGE",
		"log.compress": "WECHECKIN_LOG_COMPRESS",
		"token.user.expire": "WECHECKIN_TOKEN_USER_EXPIRE",
		"token.user.redis_prefix": "WECHECKIN_TOKEN_USER_REDIS_PREFIX",
		"token.admin.expire": "WECHECKIN_TOKEN_ADMIN_EXPIRE",
		"token.admin.redis_prefix": "WECHECKIN_TOKEN_ADMIN_REDIS_PREFIX",
		"cors.allow_origins": "WECHECKIN_CORS_ALLOW_ORIGINS",
		"cors.allow_methods": "WECHECKIN_CORS_ALLOW_METHODS",
		"cors.allow_headers": "WECHECKIN_CORS_ALLOW_HEADERS",
	}
	for key, env := range bindings {
		if err := v.BindEnv(key, env); err != nil {
			log.Printf("Warning: bind env %s: %v", env, err)
		}
	}
}
```

Call `setDefaults(v)` before reading config files and `bindEnv(v)` before unmarshalling.

- [ ] **Step 3: Run config tests and verify GREEN**

Run:

```bash
GOCACHE=$PWD/.cache/go-build go test ./backend/internal/config -count=1
```

Expected: PASS.

### Task 3: Add Safe Example Config

**Files:**
- Create: `backend/config/config.example.yaml`

- [ ] **Step 1: Add example YAML**

Create `backend/config/config.example.yaml` with local defaults and placeholders:

```yaml
server:
  port: "8083"
  host: "0.0.0.0"
  mode: "debug"
  timeout: 30

cors:
  allow_origins:
    - "*"
  allow_methods:
    - "GET"
    - "POST"
    - "PUT"
    - "DELETE"
    - "OPTIONS"
  allow_headers:
    - "Origin"
    - "Content-Type"
    - "Accept"
    - "Authorization"

database:
  host: "localhost"
  port: 3306
  user: "wecheckin"
  password: "change-me"
  dbname: "wecheckin"
  charset: "utf8mb4"
  parse_time: true
  loc: "Local"

log:
  dir: "./logs"
  level: "info"
  max_age: 30
  compress: true

redis:
  host: "localhost"
  port: 6379
  password: "change-me"
  db: 0

token:
  user:
    expire: "168h"
    redis_prefix: "user_token:"
  admin:
    expire: "24h"
    redis_prefix: "admin_token:"

oss:
  type: "local"
  local:
    path: "./uploads"

security:
  app_id: "change-me"
  app_secret: "change-me"
```

### Task 4: Update README Config Guidance

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Document environment override convention**

In the configuration section, add that `WECHECKIN_` environment variables override YAML values.

- [ ] **Step 2: Add examples for sensitive values**

Add example:

```bash
cd backend
WECHECKIN_DATABASE_PASSWORD='your-db-password' \
WECHECKIN_REDIS_PASSWORD='your-redis-password' \
go run cmd/main.go
```

- [ ] **Step 3: Mention example config**

Mention `backend/config/config.example.yaml` as the safe reference file for new environments.

### Task 5: Verify and Clean Up

**Files:**
- Read: working tree status

- [ ] **Step 1: Run full scoped verification**

Run:

```bash
GOCACHE=$PWD/.cache/go-build go test ./backend/internal/config ./backend/internal/app/formkit/...
```

Expected: all listed packages pass.

- [ ] **Step 2: Remove Go cache directory**

Run:

```bash
rm -rf .cache
```

Expected: `.cache/` is absent.

- [ ] **Step 3: Review final diff**

Run:

```bash
git diff -- backend/internal/config/config.go backend/internal/config/config_test.go backend/config/config.example.yaml README.md docs/superpowers/plans/2026-07-13-config-env-override.md
```

Expected: only scoped configuration override changes appear.
