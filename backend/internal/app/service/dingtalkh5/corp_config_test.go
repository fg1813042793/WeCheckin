package dingtalkh5

import (
	"os"
	"strings"
	"testing"
)

func TestSaveCorpConfigsReplacesPersistedList(t *testing.T) {
	src, err := os.ReadFile("corp_config.go")
	if err != nil {
		t.Fatalf("read corp_config.go: %v", err)
	}
	text := string(src)
	for _, want := range []string{
		"func deleteOmittedDingTalkH5CorpConfigsDB",
		"`corp_id` NOT IN ?",
		"deleteOmittedDingTalkH5CorpConfigsDB(tx, savedCorpIDs)",
		"db.Transaction(func(tx *gorm.DB) error",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("saving corp configs should replace omitted configs with %q", want)
		}
	}
}

func TestSaveCorpConfigsAllowsClearingAllConfigs(t *testing.T) {
	src, err := os.ReadFile("corp_config.go")
	if err != nil {
		t.Fatalf("read corp_config.go: %v", err)
	}
	text := string(src)
	for _, want := range []string{
		`return db.Where("1 = 1").Delete(&model.DingTalkH5CorpConfig{}).Error`,
		"return clearLegacyDingTalkH5CorpConfigContext(ctx)",
		"func clearLegacyDingTalkH5CorpConfigContext(ctx context.Context) error",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("saving an empty corp config list should clear all configs with %q", want)
		}
	}
}
