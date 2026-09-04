package main

import (
	"os"
	"strings"
	"testing"
)

func TestDictionaryRoutesSeparateAdminAndPublicContracts(t *testing.T) {
	adminRoutes, err := os.ReadFile("../internal/routes/v2/admin/routes.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`admin.POST("/dict/types", aDict.AddDictType)`,
		`admin.PUT("/dict/types/:typeCode", aDict.EditDictType)`,
		`admin.DELETE("/dict/types/:typeCode", aDict.DelDictType)`,
	} {
		if !strings.Contains(string(adminRoutes), want) {
			t.Fatalf("admin dictionary routes missing %q", want)
		}
	}

	clientRoutes, err := os.ReadFile("../internal/routes/v2/client/routes.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"aDict.GetPublicDictTypes", "aDict.GetPublicDictByType"} {
		if !strings.Contains(string(clientRoutes), want) {
			t.Fatalf("public dictionary routes missing %q", want)
		}
	}

	permissions, err := os.ReadFile("../internal/middleware/admin/route_permissions.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`"POST /api/v2/admin/dict/types": "dict:add"`,
		`path: "/api/v2/admin/dict/types/:typeCode", perm: "dict:edit"`,
		`path: "/api/v2/admin/dict/types/:typeCode", perm: "dict:del"`,
	} {
		if !strings.Contains(string(permissions), want) {
			t.Fatalf("dictionary route permissions missing %q", want)
		}
	}
}
