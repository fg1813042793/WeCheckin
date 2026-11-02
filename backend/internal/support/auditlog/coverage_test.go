package auditlog

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestRegisteredAdminAndH5MutationsHaveAuditClassification(t *testing.T) {
	pattern := regexp.MustCompile(`\b(admin|auth|authed|group)\.(POST|PUT|PATCH|DELETE)\("([^"]+)"`)
	for _, file := range []struct{ path, prefix string }{
		{"admin/routes.go", "/api/v2/admin"}, {"dingtalkh5/routes.go", "/api/v2/dingtalk/h5"},
	} {
		source, err := os.ReadFile(filepath.Join("..", "..", "routes", "v2", file.path))
		if err != nil {
			t.Fatal(err)
		}
		matches := pattern.FindAllStringSubmatch(string(source), -1)
		if len(matches) < 20 {
			t.Fatalf("route inventory incomplete: %s", file.path)
		}
		for _, match := range matches {
			path := file.prefix + match[3]
			if file.path == "admin/routes.go" && match[1] == "auth" {
				path = file.prefix + "/auth" + match[3]
			}
			last := path[strings.LastIndex(path, "/")+1:]
			if last == "validate" || last == "preview" || last == "cron-preview" || last == "parse" || last == "evaluate" || last == "read" || last == "read-all" {
				continue
			}
			op, tracked := Classify(match[2], path)
			if !tracked || op.Module == "other" {
				t.Errorf("missing explicit audit classification for %s %s: %#v", match[2], path, op)
			}
		}
	}
}

func TestCollectorWrapsBothCurrentAndLegacyRoutes(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "routes", "routes.go"))
	if err != nil {
		t.Fatal(err)
	}
	source := string(data)
	collector := strings.Index(source, "h.Use(middleware.AuditLog())")
	for _, route := range []string{"clientv1.Register(h)", "routesv2.Register(h)"} {
		if collector < 0 || strings.Index(source, route) < collector {
			t.Fatalf("audit collector must precede %s", route)
		}
	}
}
