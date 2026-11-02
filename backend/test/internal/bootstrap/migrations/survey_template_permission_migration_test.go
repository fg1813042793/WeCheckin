package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSurveyTemplatePermissionMigrationRegistersAndBackfillsPermissions(t *testing.T) {
	path := filepath.Join("..", "..", "migrations", "20261028190000_add_survey_template_page_permissions.sql")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read survey template permission migration: %v", err)
	}
	text := string(source)
	for _, required := range []string{
		"admin:menu:survey:templates",
		"admin:menu:survey:templates:edit",
		"admin:api:survey-template:list",
		"admin:api:survey-template:edit",
		"admin:menu:survey:button:list",
		"admin:menu:survey:edit",
		"admin:api:survey:list",
		"admin:api:survey:edit",
		"survey-template-menu-backfill",
		"ON DUPLICATE KEY UPDATE",
	} {
		if !strings.Contains(text, required) {
			t.Fatalf("survey template permission migration missing %q", required)
		}
	}
}
