package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDingTalkH5MultiCorpMigrationExists(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_dingtalk_h5_multi_corp.sql"))
	if err != nil {
		t.Fatalf("glob dingtalk h5 multi corp migration: %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("dingtalk h5 multi corp migration is required")
	}

	src, err := os.ReadFile(matches[len(matches)-1])
	if err != nil {
		t.Fatalf("read dingtalk h5 multi corp migration: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"CREATE TABLE IF NOT EXISTS `dingtalk_h5_corp_configs`",
		"CREATE TABLE IF NOT EXISTS `dingtalk_h5_user_bindings`",
		"UNIQUE KEY `uk_dt_h5_corp_id` (`corp_id`)",
		"UNIQUE KEY `uk_dt_h5_binding_corp_user` (`corp_id`,`dingtalk_user_id`)",
		"`DINGTALK_H5_CORP_ID`",
		"`DINGTALK_H5_APP_KEY`",
		"`DINGTALK_H5_APP_SECRET`",
		"`user_mini_openid`",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("multi corp migration must include %q", snippet)
		}
	}
}

func TestDingTalkH5MultiCorpMigrationQualifiesUpsertFallbackColumns(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_dingtalk_h5_multi_corp.sql"))
	if err != nil {
		t.Fatalf("glob dingtalk h5 multi corp migration: %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("dingtalk h5 multi corp migration is required")
	}
	src, err := os.ReadFile(matches[len(matches)-1])
	if err != nil {
		t.Fatalf("read dingtalk h5 multi corp migration: %v", err)
	}
	text := string(src)
	for _, forbidden := range []string{
		"VALUES(`corp_name`), `corp_name`)",
		"VALUES(`app_key`), `app_key`)",
		"VALUES(`app_secret`), `app_secret`)",
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("multi corp migration must qualify target fallback columns in INSERT SELECT upsert, found ambiguous %q", forbidden)
		}
	}
}

func TestDingTalkH5MultiCorpMigrationAvoidsLegacyCollationConflicts(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_dingtalk_h5_multi_corp.sql"))
	if err != nil {
		t.Fatalf("glob dingtalk h5 multi corp migration: %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("dingtalk h5 multi corp migration is required")
	}
	src, err := os.ReadFile(matches[len(matches)-1])
	if err != nil {
		t.Fatalf("read dingtalk h5 multi corp migration: %v", err)
	}
	text := string(src)
	for _, want := range []string{
		"DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci",
		"ALTER TABLE `dingtalk_h5_corp_configs` CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci",
		"ALTER TABLE `dingtalk_h5_user_bindings` CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci",
		"cfg.`corp_id` COLLATE utf8mb4_general_ci =",
		"COLLATE utf8mb4_general_ci",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("multi corp migration must avoid legacy collation conflicts, missing %q", want)
		}
	}
}

func TestAutoMigrateIncludesDingTalkH5MultiCorpModels(t *testing.T) {
	src, err := os.ReadFile("migrate.go")
	if err != nil {
		t.Fatalf("read migrate.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"&model.DingTalkH5CorpConfig{}",
		"&model.DingTalkH5UserBinding{}",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("auto migrate must include %q", snippet)
		}
	}
}
