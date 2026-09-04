package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDictionaryTypeMigrationPreservesHistoricalItems(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_split_dictionary_types.sql"))
	if err != nil {
		t.Fatalf("glob dictionary type migration: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected one dictionary type migration, got %d", len(matches))
	}
	source, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read dictionary type migration: %v", err)
	}
	text := string(source)
	for _, want := range []string{
		"CREATE TABLE IF NOT EXISTS `sys_dict_types`",
		"PRIMARY KEY (`dict_type_code`)",
		"INSERT INTO `sys_dict_types`",
		"FROM `sys_dicts`",
		"ON DUPLICATE KEY UPDATE",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("dictionary type migration missing %q", want)
		}
	}
	if strings.Contains(strings.ToUpper(text), "DELETE FROM `SYS_DICTS`") {
		t.Fatal("dictionary type migration must preserve historical dictionary rows")
	}
}

func TestDictionaryTypeCollationRepairMigrationSupportsAlreadyInitializedDatabases(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_normalize_dictionary_type_collation.sql"))
	if err != nil {
		t.Fatalf("glob dictionary type collation migration: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected one dictionary type collation repair migration, got %d", len(matches))
	}
	source, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read dictionary type collation migration: %v", err)
	}
	text := string(source)
	for _, want := range []string{
		"ALTER TABLE `sys_dict_types`",
		"CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("dictionary type collation repair migration missing %q", want)
		}
	}
}
