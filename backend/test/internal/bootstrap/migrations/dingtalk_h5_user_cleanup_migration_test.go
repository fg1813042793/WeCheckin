package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDingTalkH5PerfUserCleanupMigrationMigratesBeforeDrop(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_drop_dingtalk_h5_perf_users.sql"))
	if err != nil {
		t.Fatalf("glob dingtalk h5 user cleanup migration: %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("dingtalk h5 user cleanup migration is required")
	}

	src, err := os.ReadFile(matches[len(matches)-1])
	if err != nil {
		t.Fatalf("read dingtalk h5 user cleanup migration: %v", err)
	}
	text := string(src)
	required := []string{
		"INSERT INTO `users`",
		"FROM `dingtalk_h5_perf_users`",
		"ON DUPLICATE KEY UPDATE",
		"JSON_SET",
		"dingtalkH5Performance",
		"CAST(p.`created_at` AS CHAR)",
		"CAST(p.`updated_at` AS CHAR)",
		"DROP TABLE IF EXISTS `dingtalk_h5_perf_users`",
	}
	for _, snippet := range required {
		if !strings.Contains(text, snippet) {
			t.Fatalf("dingtalk h5 user cleanup migration must include %s", snippet)
		}
	}
	insertIndex := strings.Index(text, "INSERT INTO `users`")
	dropIndex := strings.Index(text, "DROP TABLE IF EXISTS `dingtalk_h5_perf_users`")
	if insertIndex < 0 || dropIndex < 0 || dropIndex < insertIndex {
		t.Fatalf("cleanup migration must migrate dingtalk h5 users before dropping old table")
	}
}

func TestAutoMigrateRunsDingTalkH5PerfUserCleanupStep(t *testing.T) {
	src, err := os.ReadFile("migrate.go")
	if err != nil {
		t.Fatalf("read migrate.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		`Name: "cleanup_dingtalk_h5_perf_users"`,
		"cleanupDingTalkH5PerfUsers(db)",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("auto migrate must run dingtalk h5 obsolete user cleanup step %s", snippet)
		}
	}
}
