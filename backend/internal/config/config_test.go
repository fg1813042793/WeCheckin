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
	t.Setenv("WECHECKIN_DATABASE_HOST", "db.env")
	t.Setenv("WECHECKIN_DATABASE_PORT", "3310")
	t.Setenv("WECHECKIN_DATABASE_PASSWORD", "from-env")
	t.Setenv("WECHECKIN_REDIS_DB", "5")
	t.Setenv("WECHECKIN_REDIS_PASSWORD", "redis-env")
	t.Setenv("WECHECKIN_LOG_COMPRESS", "false")
	t.Setenv("WECHECKIN_OSS_TYPE", "local")
	t.Setenv("WECHECKIN_OSS_LOCAL_PATH", "./env-uploads")
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
	if cfg.OSS.Type != "local" || cfg.OSS.Local.Path != "./env-uploads" {
		t.Fatalf("oss config = %#v", cfg.OSS)
	}
	wantOrigins := []string{"http://one.local", "http://two.local"}
	if len(cfg.CORS.AllowOrigins) != len(wantOrigins) || cfg.CORS.AllowOrigins[0] != wantOrigins[0] || cfg.CORS.AllowOrigins[1] != wantOrigins[1] {
		t.Fatalf("cors origins = %#v", cfg.CORS.AllowOrigins)
	}
	if cfg.Token.Admin.Expire != "48h" {
		t.Fatalf("admin token expire = %q", cfg.Token.Admin.Expire)
	}
}
