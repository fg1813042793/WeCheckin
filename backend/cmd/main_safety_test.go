package main

import (
	"os"
	"strings"
	"testing"
)

func TestPublicDebugTokenRouteIsNotRegistered(t *testing.T) {
	src, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}

	if strings.Contains(string(src), `"/test/debug_token"`) {
		t.Fatalf("public debug token route must not be registered")
	}
}

func TestMainDelegatesRouteRegistration(t *testing.T) {
	src, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}

	text := string(src)
	if !strings.Contains(text, "registerRoutes(h)") {
		t.Fatalf("main.go must delegate route registration to registerRoutes")
	}
	for _, snippet := range []string{`adminGroup.GET`, `h.POST("/upload"`} {
		if strings.Contains(text, snippet) {
			t.Fatalf("main.go must not contain route registration snippet %s", snippet)
		}
	}
}

func TestMainDoesNotRunDatabaseMaintenanceOnServiceStartup(t *testing.T) {
	src, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}

	text := string(src)
	for _, snippet := range []string{
		"bootstrap.InitBusiness",
		"bootstrap.RunMaintenance",
		"autoMigrate",
		"seedMenus",
		"seedSetups",
	} {
		if strings.Contains(text, snippet) {
			t.Fatalf("main.go must not run database maintenance during service startup: %s", snippet)
		}
	}
}

func TestStartScriptRunsBackendPackage(t *testing.T) {
	src, err := os.ReadFile("../start.sh")
	if err != nil {
		t.Fatalf("read start.sh: %v", err)
	}

	text := string(src)
	if strings.Contains(text, "go run cmd/main.go") {
		t.Fatalf("start.sh must not run a single Go file; use package mode so routes.go is compiled")
	}
	if !strings.Contains(text, "go run ./cmd") {
		t.Fatalf("start.sh must run the backend command package")
	}
}

func TestBackendPortReferencesUse8083(t *testing.T) {
	files := map[string]struct {
		required    []string
		requiredAny [][]string
		forbidden   []string
	}{
		"main.go": {
			required:  []string{"localhost:8083"},
			forbidden: []string{"localhost:8080"},
		},
		"../internal/support/media/static.go": {
			required:  []string{"http://localhost:8083"},
			forbidden: []string{"http://localhost:8080"},
		},
		"../Dockerfile": {
			required:  []string{"EXPOSE 8083"},
			forbidden: []string{"EXPOSE 8080"},
		},
		"../docker-compose.yml": {
			required: []string{`"8083:8083"`},
			requiredAny: [][]string{
				{"WECHECKIN_SERVER_PORT=8083", "WECHECKIN_SERVER_PORT=${WECHECKIN_SERVER_PORT:-8083}"},
			},
			forbidden: []string{`"8080:8080"`},
		},
		"../nginx.conf": {
			required:  []string{"location /api/", "proxy_pass http://backend:8083;", "location = /health", "location = /ready"},
			forbidden: []string{"backend:8080", "proxy_pass http://backend:8083/;", "\n}\n\n# API 服务器配置\nserver"},
		},
		"../docs/swagger/docs.go": {
			required:  []string{"localhost:8083"},
			forbidden: []string{"localhost:8080"},
		},
		"../docs/swagger/swagger.json": {
			required:  []string{`"host": "localhost:8083"`},
			forbidden: []string{`"host": "localhost:8080"`},
		},
		"../docs/swagger/swagger.yaml": {
			required:  []string{"host: localhost:8083"},
			forbidden: []string{"host: localhost:8080"},
		},
	}

	for path, checks := range files {
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		text := string(src)
		for _, snippet := range checks.required {
			if !strings.Contains(text, snippet) {
				t.Fatalf("%s must contain %q", path, snippet)
			}
		}
		for _, candidates := range checks.requiredAny {
			matched := false
			for _, snippet := range candidates {
				if strings.Contains(text, snippet) {
					matched = true
					break
				}
			}
			if !matched {
				t.Fatalf("%s must contain one of %q", path, candidates)
			}
		}
		for _, snippet := range checks.forbidden {
			if strings.Contains(text, snippet) {
				t.Fatalf("%s must not contain %q", path, snippet)
			}
		}
	}
}

func TestHealthAndReadinessRoutesAreRegistered(t *testing.T) {
	src, err := os.ReadFile("../internal/routes/common/health.go")
	if err != nil {
		t.Fatalf("read internal/routes/common/health.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{`h.GET("/health"`, `h.GET("/ready"`, "database.WithContext"} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("internal/routes/common/health.go must contain %q", snippet)
		}
	}
	if strings.Contains(text, "database.DB") {
		t.Fatalf("internal/routes/common/health.go must use database.WithContext instead of database.DB")
	}
}
