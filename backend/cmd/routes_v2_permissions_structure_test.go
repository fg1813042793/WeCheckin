package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestV2AdminRoutesExposePermissionResource(t *testing.T) {
	src, err := os.ReadFile("routes_v2.go")
	if err != nil {
		t.Fatalf("read routes_v2.go: %v", err)
	}
	text := string(src)
	for _, want := range []string{
		"adminpermission.NewAdminPermissionHandler()",
		`admin.GET("/permissions/tree", aPermission.GetPermissionTree)`,
		`admin.GET("/permissions", aPermission.GetPermissionList)`,
		`admin.POST("/permissions", aPermission.AddPermission)`,
		`admin.PUT("/permissions/:key", aPermission.EditPermission)`,
		`admin.DELETE("/permissions/:key", withFormParam("key", "key", aPermission.DelPermission))`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("routes_v2.go missing permission resource route %s", want)
		}
	}
	if strings.Contains(text, `admin.PUT("/permissions/:key", withFormParam("key", "key", aPermission.EditPermission))`) {
		t.Fatalf("permission edit route must not inject route key into form key because the form key is editable")
	}
}

func TestBackendNoLongerRegistersLegacyAdminPermissionRoutes(t *testing.T) {
	routesSrc, err := os.ReadFile("routes.go")
	if err != nil {
		t.Fatalf("read routes.go: %v", err)
	}
	if strings.Contains(string(routesSrc), "registerAdminRoutes(h)") {
		t.Fatalf("backend must not register legacy /admin permission routes after v2 permission migration")
	}
	if _, err := os.Stat("routes_admin.go"); err == nil {
		t.Fatalf("legacy routes_admin.go should be removed after /api/v2/admin takes over")
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat routes_admin.go: %v", err)
	}
}

func TestAdminHandlersDoNotDocumentLegacyAdminRoutes(t *testing.T) {
	root := filepath.Join("..", "internal", "app", "handler", "admin")
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.Contains(string(src), "@Router /admin") {
			t.Fatalf("%s must document /api/v2/admin routes instead of legacy /admin routes", path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk admin handlers: %v", err)
	}
}
