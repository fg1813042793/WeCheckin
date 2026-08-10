package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMishiOrgSeedMigrationExists(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_seed_mishi_org_users.sql"))
	if err != nil {
		t.Fatalf("glob mishi org seed migration: %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("mishi org user seed migration is required")
	}

	src, err := os.ReadFile(matches[len(matches)-1])
	if err != nil {
		t.Fatalf("read mishi org seed migration: %v", err)
	}
	text := string(src)
	required := []string{
		"米视科技",
		"M/H业务",
		"M/H业务线",
		"Java开发一组",
		"安卓开发二组",
		"tmp_mishi_user_seed",
		"INSERT INTO `users`",
		"INSERT INTO `user_depts`",
		"MD5('123456')",
		"COLLATE=utf8mb4_general_ci",
	}
	for _, snippet := range required {
		if !strings.Contains(text, snippet) {
			t.Fatalf("mishi org seed migration must include %s", snippet)
		}
	}
	for _, userID := range []string{
		"lip",
		"arthur",
		"foster",
		"cube",
		"hrbp",
		"nick",
		"rock",
		"sherif",
		"paul",
		"neil",
		"david",
		"monica",
		"lucky",
		"betty",
		"cherry",
		"amy",
	} {
		if !strings.Contains(text, "'"+userID+"'") {
			t.Fatalf("mishi org seed migration must include user %s", userID)
		}
	}
}
