package bootstrap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSurveyExamIndexMigrationExists(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_add_survey_exam_indexes.sql"))
	if err != nil {
		t.Fatalf("glob survey exam index migration: %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("survey exam index migration is required")
	}

	src, err := os.ReadFile(matches[len(matches)-1])
	if err != nil {
		t.Fatalf("read survey exam index migration: %v", err)
	}
	text := string(src)
	required := []string{
		"idx_surveys_status_order_id",
		"idx_surveys_category_status_order",
		"idx_survey_resp_survey_id_id",
		"idx_survey_resp_survey_status_add",
		"idx_survey_resp_survey_device",
		"idx_survey_resp_survey_ip",
		"idx_survey_questions_category_type_time",
		"idx_exams_status_order_id",
		"idx_exams_category_status_order",
		"idx_exam_records_exam_id_id",
		"idx_exam_records_exam_status_submit",
		"idx_exam_records_exam_device_status",
		"idx_exam_records_exam_ip_status",
		"idx_exam_questions_category_type_time",
	}
	for _, snippet := range required {
		if !strings.Contains(text, snippet) {
			t.Fatalf("survey exam index migration must include %s", snippet)
		}
	}
}
