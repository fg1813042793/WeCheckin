package admin

import (
	"os"
	"strings"
	"testing"
)

func TestAdminV2RoutesDisableBrowserCachingBeforeAuthentication(t *testing.T) {
	source, err := os.ReadFile("routes.go")
	if err != nil {
		t.Fatalf("read routes.go: %v", err)
	}

	want := `h.Group("/api/v2/admin", adminmw.NoStore(), adminmw.AdminAuth(), adminmw.AdminPerm())`
	if !strings.Contains(string(source), want) {
		t.Fatalf("admin v2 route group must install no-store middleware before auth: missing %q", want)
	}
}
