package config

import (
	"os"
	"path/filepath"
	"strings"
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

func withTempWorkingDir(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
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
oss:
  type: "aliyun"
  aliyun:
    access_key_id: "ak-file"
    access_key_secret: "sk-file"
    endpoint: "oss-cn-hangzhou.aliyuncs.com"
    bucket: "bucket-file"
  local:
    path: "./local-uploads"
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
	if cfg.OSS.Type != "aliyun" {
		t.Fatalf("oss type = %q", cfg.OSS.Type)
	}
	if cfg.OSS.Aliyun.Bucket != "bucket-file" {
		t.Fatalf("oss aliyun bucket = %q", cfg.OSS.Aliyun.Bucket)
	}
}

func TestLoadConfigDefaultsToBackendPort(t *testing.T) {
	withTempWorkingDir(t)

	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	if cfg.Server.Port != "8083" {
		t.Fatalf("default server port = %q, want 8083", cfg.Server.Port)
	}
}

func TestLoadConfigDefaultCORSMethodsIncludePatch(t *testing.T) {
	withTempWorkingDir(t)

	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	for _, method := range cfg.CORS.AllowMethods {
		if strings.EqualFold(method, "PATCH") {
			return
		}
	}
	t.Fatalf("default CORS methods = %#v, want PATCH", cfg.CORS.AllowMethods)
}

func TestLoadConfigAllowsEnvironmentOverrides(t *testing.T) {
	withTempConfigDir(t, testConfigYAML)
	t.Setenv("WECHECKIN_SERVER_PORT", "18083")
	t.Setenv("WECHECKIN_SERVER_TIMEOUT", "45")
	t.Setenv("WECHECKIN_SERVER_READ_TIMEOUT_SECONDS", "12")
	t.Setenv("WECHECKIN_SERVER_WRITE_TIMEOUT_SECONDS", "34")
	t.Setenv("WECHECKIN_SERVER_IDLE_TIMEOUT_SECONDS", "56")
	t.Setenv("WECHECKIN_DATABASE_HOST", "db.env")
	t.Setenv("WECHECKIN_DATABASE_PORT", "3310")
	t.Setenv("WECHECKIN_DATABASE_PASSWORD", "from-env")
	t.Setenv("WECHECKIN_DATABASE_CONNECT_TIMEOUT_SECONDS", "9")
	t.Setenv("WECHECKIN_DATABASE_READ_TIMEOUT_SECONDS", "19")
	t.Setenv("WECHECKIN_DATABASE_WRITE_TIMEOUT_SECONDS", "29")
	t.Setenv("WECHECKIN_DATABASE_MAX_IDLE_CONNS", "8")
	t.Setenv("WECHECKIN_DATABASE_MAX_OPEN_CONNS", "64")
	t.Setenv("WECHECKIN_REDIS_DB", "5")
	t.Setenv("WECHECKIN_REDIS_PASSWORD", "redis-env")
	t.Setenv("WECHECKIN_LOG_COMPRESS", "false")
	t.Setenv("WECHECKIN_OSS_TYPE", "local")
	t.Setenv("WECHECKIN_OSS_LOCAL_PATH", "./env-uploads")
	t.Setenv("WECHECKIN_CORS_ALLOW_ORIGINS", "http://one.local,http://two.local")
	t.Setenv("WECHECKIN_CORS_ALLOW_CREDENTIALS", "true")
	t.Setenv("WECHECKIN_TOKEN_ADMIN_EXPIRE", "48h")

	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	if cfg.Server.Port != "18083" {
		t.Fatalf("server port = %q", cfg.Server.Port)
	}
	if cfg.Server.TimeoutSec != 45 || cfg.Server.ReadTimeoutSec != 12 || cfg.Server.WriteTimeoutSec != 34 || cfg.Server.IdleTimeoutSec != 56 {
		t.Fatalf("server timeouts = %#v", cfg.Server)
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
	if cfg.Database.ConnectTimeoutSec != 9 || cfg.Database.ReadTimeoutSec != 19 || cfg.Database.WriteTimeoutSec != 29 {
		t.Fatalf("database timeouts = %#v", cfg.Database)
	}
	if cfg.Database.MaxIdleConns != 8 || cfg.Database.MaxOpenConns != 64 {
		t.Fatalf("database pool limits = %#v", cfg.Database)
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
	if cfg.OSS.Type != "local" || cfg.OSS.Local.Path != "./env-uploads" {
		t.Fatalf("oss config = %#v", cfg.OSS)
	}
	wantOrigins := []string{"http://one.local", "http://two.local"}
	if len(cfg.CORS.AllowOrigins) != len(wantOrigins) || cfg.CORS.AllowOrigins[0] != wantOrigins[0] || cfg.CORS.AllowOrigins[1] != wantOrigins[1] {
		t.Fatalf("cors origins = %#v", cfg.CORS.AllowOrigins)
	}
	if !cfg.CORS.AllowCredentials {
		t.Fatal("cors allow credentials = false")
	}
	if cfg.Token.Admin.Expire != "48h" {
		t.Fatalf("admin token expire = %q", cfg.Token.Admin.Expire)
	}
}

func TestLoadConfigNormalizesServerTimeoutDefaults(t *testing.T) {
	withTempWorkingDir(t)

	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	if cfg.Server.Host != "0.0.0.0" {
		t.Fatalf("default server host = %q, want 0.0.0.0", cfg.Server.Host)
	}
	if cfg.Server.TimeoutSec != 30 || cfg.Server.ReadTimeoutSec != 30 || cfg.Server.WriteTimeoutSec != 60 || cfg.Server.IdleTimeoutSec != 120 {
		t.Fatalf("default server timeouts = %#v", cfg.Server)
	}
	if cfg.CORS.AllowCredentials {
		t.Fatal("default CORS credentials must be disabled")
	}
	if cfg.Database.ConnectTimeoutSec != 10 || cfg.Database.ReadTimeoutSec != 30 || cfg.Database.WriteTimeoutSec != 30 {
		t.Fatalf("default database timeouts = %#v", cfg.Database)
	}
	if cfg.Database.MaxIdleConns != 10 || cfg.Database.MaxOpenConns != 100 || cfg.Database.ConnMaxLifetimeMin != 60 || cfg.Database.ConnMaxIdleTimeMin != 10 {
		t.Fatalf("default database pool = %#v", cfg.Database)
	}
}

func TestLoadConfigRejectsInvalidValuesWithoutReplacingGlobalConfig(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantKey string
	}{
		{name: "port is not numeric", yaml: "server:\n  port: nope\n", wantKey: "server.port"},
		{name: "port is out of range", yaml: "server:\n  port: 70000\n", wantKey: "server.port"},
		{name: "timeout is not positive", yaml: "server:\n  timeout: 0\n", wantKey: "server.timeout"},
		{name: "database host is empty", yaml: "database:\n  host: \"\"\n", wantKey: "database.host"},
		{name: "database port is invalid", yaml: "database:\n  port: 0\n", wantKey: "database.port"},
		{name: "redis db is invalid", yaml: "redis:\n  db: 256\n", wantKey: "redis.db"},
		{
			name:    "wildcard cors cannot use credentials",
			yaml:    "cors:\n  allow_origins: [\"*\"]\n  allow_credentials: true\n",
			wantKey: "cors.allow_credentials",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			withTempConfigDir(t, test.yaml)
			previous := &Config{Server: ServerConfig{Port: "19999"}}
			Cfg = previous

			_, err := LoadConfig("")
			if err == nil {
				t.Fatalf("LoadConfig() error = nil, want %s validation error", test.wantKey)
			}
			if !strings.Contains(err.Error(), test.wantKey) {
				t.Fatalf("LoadConfig() error = %q, want key %q", err, test.wantKey)
			}
			if Cfg != previous {
				t.Fatal("invalid config replaced the active global config")
			}
		})
	}
}
