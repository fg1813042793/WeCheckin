package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenericOAWorkflowMigrationsContainRuntimeDataAndPermissions(t *testing.T) {
	formMatches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_workflow_form_data.sql"))
	if err != nil || len(formMatches) != 1 {
		t.Fatalf("workflow form data migration = %v, err = %v", formMatches, err)
	}
	formSQL, err := os.ReadFile(formMatches[0])
	if err != nil {
		t.Fatalf("read workflow form migration: %v", err)
	}
	if !strings.Contains(string(formSQL), "workflow_process_instances") || !strings.Contains(string(formSQL), "form_data_json") {
		t.Fatal("workflow form migration must add form_data_json to process instances")
	}

	permissionMatches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_generic_oa_workflow_permissions.sql"))
	if err != nil || len(permissionMatches) != 1 {
		t.Fatalf("generic OA permission migration = %v, err = %v", permissionMatches, err)
	}
	permissionSQL, err := os.ReadFile(permissionMatches[0])
	if err != nil {
		t.Fatalf("read generic OA permission migration: %v", err)
	}
	for _, required := range []string{
		"client:api-category:workflow",
		"client:api:workflow:view",
		"client:api:workflow:start",
		"client:api:workflow:handle",
		"client:api:workflow:withdraw",
		"admin:api:workflow:instance:cancel",
	} {
		if !strings.Contains(string(permissionSQL), required) {
			t.Fatalf("generic OA permission migration missing %s", required)
		}
	}
	if strings.Contains(strings.ToLower(string(permissionSQL)), "permission_grants") {
		t.Fatal("generic OA permission migration must not auto-grant permissions")
	}
}
