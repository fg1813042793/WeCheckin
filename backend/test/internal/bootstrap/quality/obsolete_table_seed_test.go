package quality_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSeedScriptsDoNotReferenceObsoleteUserTables(t *testing.T) {
	files := []string{
		filepath.Join("..", "..", "seed.sql"),
		filepath.Join("..", "..", "scripts", "test_data.sql"),
	}
	for _, file := range files {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		text := string(src)
		forbidden := []string{
			"TRUNCATE TABLE `admins`",
			"DELETE FROM admins",
			"INSERT INTO `admins`",
			"INSERT INTO `admin_depts`",
			"INSERT INTO `dingtalk_h5_perf_users`",
		}
		for _, snippet := range forbidden {
			if strings.Contains(text, snippet) {
				t.Fatalf("%s must not reference obsolete table with %s", file, snippet)
			}
		}
		for _, snippet := range []string{
			"INSERT INTO `users`",
			"`user_account`",
			"`user_admin_enabled`",
		} {
			if !strings.Contains(text, snippet) {
				t.Fatalf("%s must seed admin accounts through users with %s", file, snippet)
			}
		}
	}
}
