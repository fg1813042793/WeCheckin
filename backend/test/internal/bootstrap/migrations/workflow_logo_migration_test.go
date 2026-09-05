package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkflowLogoMigrationAddsDefinitionLogoColumn(t *testing.T) {
	path := filepath.Join("..", "..", "migrations", "20260901140000_add_workflow_definition_logo.sql")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read workflow logo migration: %v", err)
	}
	for _, snippet := range []string{
		"INFORMATION_SCHEMA.COLUMNS",
		"COLUMN_NAME = 'definition_logo_url'",
		"PREPARE stmt FROM @ddl",
		"ALTER TABLE `workflow_definitions`",
		"ADD COLUMN `definition_logo_url` varchar(500)",
	} {
		if !strings.Contains(string(source), snippet) {
			t.Fatalf("workflow logo migration must include %q", snippet)
		}
	}
}
