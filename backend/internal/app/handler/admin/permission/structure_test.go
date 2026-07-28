package permission

import (
	"os"
	"strings"
	"testing"
)

func TestAdminPermissionHandlerExposesRESTResource(t *testing.T) {
	src, err := os.ReadFile("handler.go")
	if err != nil {
		t.Fatalf("read handler.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"type AdminPermissionHandler struct",
		"NewAdminPermissionHandler",
		"GetPermissionTree",
		"GetPermissionList",
		"AddPermission",
		"EditPermission",
		"DelPermission",
		"adminpermission.SaveRequest",
		`c.Query("types")`,
		`TreeContext(ctx, c.Query("platform"), c.Query("types"))`,
		`ListContext(ctx, c.Query("platform"), c.Query("types"))`,
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("admin permission handler must expose %s", snippet)
		}
	}
}
