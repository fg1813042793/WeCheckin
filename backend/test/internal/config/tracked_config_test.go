package config_test

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestTrackedConfigsDoNotContainSecrets(t *testing.T) {
	backendRoot := backendRootFromCaller(t)
	for _, name := range []string{"config.yaml", "config.prod.yaml"} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(backendRoot, "config", name)
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read tracked config %s: %v", name, err)
			}
			var document any
			if err := yaml.Unmarshal(source, &document); err != nil {
				t.Fatalf("parse tracked config %s: %v", name, err)
			}
			assertSafeTrackedConfigValue(t, name, "", document)
		})
	}
}

func backendRootFromCaller(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test source path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..", ".."))
}

func assertSafeTrackedConfigValue(t *testing.T, filename, keyPath string, value any) {
	t.Helper()
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			path := key
			if keyPath != "" {
				path = keyPath + "." + key
			}
			assertSafeTrackedConfigValue(t, filename, path, child)
		}
	case []any:
		for index, child := range typed {
			assertSafeTrackedConfigValue(t, filename, fmt.Sprintf("%s[%d]", keyPath, index), child)
		}
	case string:
		assertSafeTrackedConfigScalar(t, filename, keyPath, typed)
	}
}

func assertSafeTrackedConfigScalar(t *testing.T, filename, keyPath, value string) {
	t.Helper()
	key := strings.ToLower(keyPath)
	leafKey := key
	if index := strings.LastIndex(leafKey, "."); index >= 0 {
		leafKey = leafKey[index+1:]
	}
	trimmed := strings.TrimSpace(value)
	for _, sensitive := range []string{"password", "secret", "access_key"} {
		if strings.Contains(leafKey, sensitive) && trimmed != "" && trimmed != "CHANGE_ME" {
			t.Errorf("%s contains a tracked credential at %s", filename, keyPath)
			return
		}
	}
	if (leafKey == "token" || strings.HasSuffix(leafKey, "_token")) && trimmed != "" && trimmed != "CHANGE_ME" {
		t.Errorf("%s contains a tracked credential at %s", filename, keyPath)
		return
	}
	if key != "database.host" && key != "redis.host" {
		return
	}
	if trimmed == "" || trimmed == "localhost" || trimmed == "0.0.0.0" {
		return
	}
	if ip := net.ParseIP(trimmed); ip != nil && ip.IsLoopback() {
		return
	}
	t.Errorf("%s contains a non-local service host at %s", filename, keyPath)
}
