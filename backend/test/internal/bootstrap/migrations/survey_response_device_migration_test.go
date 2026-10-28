package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSurveyResponseDeviceMigrationPreservesFullUserAgent(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*_expand_survey_response_device.sql"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("survey response device migration = %v, err = %v", matches, err)
	}
	source, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read survey response device migration: %v", err)
	}
	text := string(source)
	for _, snippet := range []string{
		"INFORMATION_SCHEMA.COLUMNS",
		"TABLE_NAME = 'survey_response'",
		"COLUMN_NAME = 'survey_resp_device'",
		"MODIFY COLUMN `survey_resp_device` TEXT",
		"PREPARE",
		"DEALLOCATE PREPARE",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("survey response device migration missing %q", snippet)
		}
	}

	modelSource, err := os.ReadFile(filepath.Join("..", "model", "assessment", "survey.go"))
	if err != nil {
		t.Fatalf("read survey response model: %v", err)
	}
	if !strings.Contains(string(modelSource), "gorm:\"column:survey_resp_device;type:text\"") {
		t.Fatal("SurveyResponse.Device must use a text column for the complete User-Agent")
	}
}
