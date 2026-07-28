package main

import (
	"os"
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
		`admin.PUT("/permissions/:key", withFormParam("key", "key", aPermission.EditPermission))`,
		`admin.DELETE("/permissions/:key", withFormParam("key", "key", aPermission.DelPermission))`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("routes_v2.go missing permission resource route %s", want)
		}
	}
}
