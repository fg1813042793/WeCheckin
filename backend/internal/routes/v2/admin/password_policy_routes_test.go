package admin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPasswordPolicyRoutesPermissionsAndSwagger(t *testing.T) {
	files := map[string]string{
		"routes":     "routes.go",
		"middleware": filepath.Join("..", "..", "..", "middleware", "admin", "route_permissions.go"),
		"swagger":    filepath.Join("..", "swagger", "password_policy.go"),
	}
	contents := make(map[string]string, len(files))
	for name, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		contents[name] = string(data)
	}
	for _, snippet := range []string{`admin.GET("/password-policy", aSetup.GetPasswordPolicy)`, `admin.PUT("/password-policy", aSetup.SavePasswordPolicy)`} {
		if !strings.Contains(contents["routes"], snippet) {
			t.Fatalf("password policy routes missing %q", snippet)
		}
	}
	for _, snippet := range []string{
		`"GET /api/v2/admin/password-policy":`, `"PUT /api/v2/admin/password-policy":`,
		`"setup:list"`, `"setup:edit"`,
	} {
		if !strings.Contains(contents["middleware"], snippet) {
			t.Fatalf("password policy middleware missing %q", snippet)
		}
	}
	for _, snippet := range []string{"AdminPasswordPolicyGet", "AdminPasswordPolicyPut", "/api/v2/admin/password-policy"} {
		if !strings.Contains(contents["swagger"], snippet) {
			t.Fatalf("password policy swagger missing %q", snippet)
		}
	}
}
