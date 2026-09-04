package main

import (
	"encoding/json"
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

func TestSwaggerPublishesOnlyV2GroupsAndPaths(t *testing.T) {
	mainSource, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}
	for _, line := range strings.Split(string(mainSource), "\n") {
		if strings.Contains(line, "@tag.name") && isLegacySwaggerTag(line) {
			t.Fatalf("main.go must not declare legacy Swagger tag: %s", strings.TrimSpace(line))
		}
	}
	if !strings.Contains(string(mainSource), "H5AppToken") {
		t.Fatal("main.go must declare the H5App Swagger security scheme")
	}

	swaggerSource, err := os.ReadFile("../docs/swagger/swagger.json")
	if err != nil {
		t.Fatalf("read generated Swagger JSON: %v", err)
	}
	var document struct {
		Paths map[string]map[string]struct {
			Tags []string `json:"tags"`
		} `json:"paths"`
		Tags []struct {
			Name string `json:"name"`
		} `json:"tags"`
	}
	if err := json.Unmarshal(swaggerSource, &document); err != nil {
		t.Fatalf("parse generated Swagger JSON: %v", err)
	}
	if len(document.Paths) == 0 {
		t.Fatal("generated Swagger must publish API v2 paths")
	}
	h5AppOperations := 0
	for path, operations := range document.Paths {
		if !strings.HasPrefix(path, "/api/v2/") {
			t.Fatalf("generated Swagger must only publish API v2 paths, found %q", path)
		}
		for method, operation := range operations {
			for _, tag := range operation.Tags {
				if isLegacySwaggerTag(tag) {
					t.Fatalf("generated Swagger operation %s %s must not publish legacy group %q", strings.ToUpper(method), path, tag)
				}
				if isGenericSwaggerTag(tag) {
					t.Fatalf("generated Swagger operation %s %s must use a business category instead of %q", strings.ToUpper(method), path, tag)
				}
			}
			if strings.HasPrefix(path, "/api/v2/dingtalk/h5/") && len(operation.Tags) > 0 {
				h5AppOperations++
				for _, tag := range operation.Tags {
					if !strings.HasPrefix(tag, "API v2-H5App-") {
						t.Fatalf("generated Swagger operation %s %s must use an H5App category, found %q", strings.ToUpper(method), path, tag)
					}
				}
			}
		}
	}
	if h5AppOperations == 0 {
		t.Fatal("generated Swagger must publish H5App operations")
	}
	for _, tag := range document.Tags {
		if isLegacySwaggerTag(tag.Name) {
			t.Fatalf("generated Swagger must not publish legacy group %q", tag.Name)
		}
		if isGenericSwaggerTag(tag.Name) {
			t.Fatalf("generated Swagger must not publish generic group %q", tag.Name)
		}
	}
}

func isLegacySwaggerTag(tag string) bool {
	return strings.Contains(tag, "PC端-") ||
		(strings.Contains(tag, "客户端-") && !strings.Contains(tag, "API v2-客户端-"))
}

func isGenericSwaggerTag(tag string) bool {
	switch tag {
	case "API v2-后台管理", "API v2-客户端", "API v2-公开接口":
		return true
	default:
		return false
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
