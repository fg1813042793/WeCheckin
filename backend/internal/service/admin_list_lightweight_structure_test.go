package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAdminContentListsSelectLightweightColumns(t *testing.T) {
	checks := []struct {
		file     string
		columns  string
		selectBy string
		forbid   []string
	}{
		{
			file:     filepath.Join("client", "survey", "service.go"),
			columns:  "var adminSurveyListColumns = []string{",
			selectBy: "Select(adminSurveyListColumns)",
			forbid:   []string{"survey_schema", "survey_settings"},
		},
		{
			file:     filepath.Join("client", "exam", "service.go"),
			columns:  "var adminExamListColumns = []string{",
			selectBy: "Select(adminExamListColumns)",
			forbid:   []string{"exam_schema", "exam_settings"},
		},
		{
			file:     filepath.Join("admin", "admincontent", "enroll.go"),
			columns:  "var adminEnrollListColumns = []string{",
			selectBy: "Select(adminEnrollListColumns)",
			forbid:   []string{"enroll_forms", "enroll_join_forms", "enroll_user_list"},
		},
		{
			file:     filepath.Join("client", "event", "admin.go"),
			columns:  "var adminEventListColumns = []string{",
			selectBy: "Select(adminEventListColumns)",
			forbid:   []string{"event_forms", "event_score_fields"},
		},
	}
	for _, check := range checks {
		src, err := os.ReadFile(check.file)
		if err != nil {
			t.Fatalf("read %s: %v", check.file, err)
		}
		text := string(src)
		for _, required := range []string{check.columns, check.selectBy} {
			if !strings.Contains(text, required) {
				t.Fatalf("%s should use lightweight admin list columns with %q", check.file, required)
			}
		}
		start := strings.Index(text, check.columns)
		end := strings.Index(text[start:], "}")
		if start < 0 || end < 0 {
			t.Fatalf("%s lightweight columns declaration is incomplete", check.file)
		}
		block := text[start : start+end]
		for _, forbidden := range check.forbid {
			if strings.Contains(block, forbidden) {
				t.Fatalf("%s lightweight columns should not include %q", check.file, forbidden)
			}
		}
	}
}
