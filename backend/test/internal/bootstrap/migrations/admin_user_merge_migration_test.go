package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAdminUserMergeMigrationCopiesAdminsIntoUsers(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_merge_admins_into_users.sql"))
	if err != nil {
		t.Fatalf("glob admin user merge migration: %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("admin user merge migration is required")
	}

	src, err := os.ReadFile(matches[len(matches)-1])
	if err != nil {
		t.Fatalf("read admin user merge migration: %v", err)
	}
	text := string(src)
	required := []string{
		"`user_account`",
		"`user_admin_enabled`",
		"`user_admin_type`",
		"`user_role_id`",
		"`user_admin_desc`",
		"`user_admin_token`",
		"INSERT INTO `users`",
		"FROM `admins`",
		"ON DUPLICATE KEY UPDATE",
		"INSERT INTO `user_depts`",
		"FROM `admin_depts`",
		"UPDATE `logs`",
	}
	for _, snippet := range required {
		if !strings.Contains(text, snippet) {
			t.Fatalf("admin user merge migration must include %s", snippet)
		}
	}

	insertIndex := strings.Index(text, "INSERT INTO `users`")
	deptInsertIndex := strings.Index(text, "INSERT INTO `user_depts`")
	if insertIndex < 0 || deptInsertIndex < insertIndex {
		t.Fatalf("admin user merge migration must copy admins before merging admin_depts into user_depts")
	}
}

func TestAutoMigrateRunsAdminUserMergeStep(t *testing.T) {
	src, err := os.ReadFile("migrate.go")
	if err != nil {
		t.Fatalf("read migrate.go: %v", err)
	}
	text := string(src)
	required := []string{
		"Name: \"merge_admins_into_users\"",
		"mergeAdminsIntoUsers(db)",
		"user_admin_enabled",
	}
	for _, snippet := range required {
		if !strings.Contains(text, snippet) {
			t.Fatalf("startup migration must include %s", snippet)
		}
	}
}

func TestAdminUserMergeSanitizesLegacyZeroTimestamps(t *testing.T) {
	src, err := os.ReadFile("migrate.go")
	if err != nil {
		t.Fatalf("read migrate.go: %v", err)
	}
	text := string(src)
	requiredSource := []string{
		"createdAt := migrationTimeOrNow(item.CreatedAt",
		"updatedAt := migrationTimeOrNow(item.UpdatedAt",
		"\"updated_at\":            updatedAt",
		"CreatedAt:      createdAt",
		"UpdatedAt:      updatedAt",
	}
	for _, snippet := range requiredSource {
		if !strings.Contains(text, snippet) {
			t.Fatalf("startup admin merge must sanitize legacy zero timestamps with %s", snippet)
		}
	}
	forbiddenSource := []string{
		"\"updated_at\":            item.UpdatedAt",
		"CreatedAt:      item.CreatedAt",
		"UpdatedAt:      item.UpdatedAt",
	}
	for _, snippet := range forbiddenSource {
		if strings.Contains(text, snippet) {
			t.Fatalf("startup admin merge must not write legacy zero timestamp directly: %s", snippet)
		}
	}

	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_merge_admins_into_users.sql"))
	if err != nil {
		t.Fatalf("glob admin user merge migration: %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("admin user merge migration is required")
	}
	migration, err := os.ReadFile(matches[len(matches)-1])
	if err != nil {
		t.Fatalf("read admin user merge migration: %v", err)
	}
	migrationText := string(migration)
	requiredSQL := []string{
		"CAST(`created_at` AS CHAR)",
		"CAST(`updated_at` AS CHAR)",
		"NOW(3) ELSE `created_at`",
		"NOW(3) ELSE `updated_at`",
	}
	for _, snippet := range requiredSQL {
		if !strings.Contains(migrationText, snippet) {
			t.Fatalf("sql admin merge migration must sanitize legacy zero timestamps with %s", snippet)
		}
	}
	if strings.Contains(migrationText, "`admin_add_ip`, `admin_edit_ip`, `created_at`, `updated_at` FROM `admins`") {
		t.Fatalf("sql admin merge migration must not copy legacy zero timestamps directly")
	}
}

func TestAdminUserMergeMigrationUsesCollationSafeStringComparisons(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_merge_admins_into_users.sql"))
	if err != nil {
		t.Fatalf("glob admin user merge migration: %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("admin user merge migration is required")
	}
	migration, err := os.ReadFile(matches[len(matches)-1])
	if err != nil {
		t.Fatalf("read admin user merge migration: %v", err)
	}
	text := string(migration)
	required := []string{
		"CAST(u.`user_mini_openid` AS CHAR CHARACTER SET utf8mb4) COLLATE utf8mb4_unicode_ci",
		"CAST(CONCAT(''admin:'', a.`id`) AS CHAR CHARACTER SET utf8mb4) COLLATE utf8mb4_unicode_ci",
		"CAST(l.`log_admin_id` AS CHAR CHARACTER SET utf8mb4) COLLATE utf8mb4_unicode_ci",
		"CAST(m.`legacy_admin_id` AS CHAR CHARACTER SET utf8mb4) COLLATE utf8mb4_unicode_ci",
		"CAST(existing.`user_id` AS CHAR CHARACTER SET utf8mb4) COLLATE utf8mb4_unicode_ci",
	}
	for _, snippet := range required {
		if !strings.Contains(text, snippet) {
			t.Fatalf("admin user merge migration must avoid mixed collations with %s", snippet)
		}
	}
}
