package exam

import (
	"encoding/json"
	"strings"
	"testing"

	"wecheckin/backend/internal/model"
)

func TestNormalizeExamDisplayTitleSettings(t *testing.T) {
	for _, raw := range []string{"", "{}", `{"autoSave":true}`, `{"displayTitle":null}`, `{"displayTitle":"   "}`} {
		result, err := NormalizeDisplayTitleSettings(raw)
		if err != nil || !json.Valid([]byte(result)) {
			t.Fatalf("legacy settings %q: %q %v", raw, result, err)
		}
	}
	raw := `{"displayTitle":"  考生名称  ","unrelated":9007199254740993,"logicRules":"[]"}`
	result, err := NormalizeDisplayTitleSettings(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, `"displayTitle":"考生名称"`) || !strings.Contains(result, "9007199254740993") {
		t.Fatalf("normalization lost data: %s", result)
	}
	for _, invalid := range []string{`{"displayTitle":123}`, `{"displayTitle":{}}`, `{"displayTitle":true}`, `[]`, `broken`, `{"displayTitle":"` + strings.Repeat("考", 101) + `"}`} {
		if _, err := NormalizeDisplayTitleSettings(invalid); err == nil {
			t.Fatalf("accepted invalid settings: %s", invalid)
		}
	}
}

func TestNormalizeExamDownloadSettings(t *testing.T) {
	raw := `{"answerDownloadFormat":" DOCX ","answerDownloadExpireDays":30,"answerDownloadBaseUrl":" https://api.example.com:8086/ ","publicUrl":" https://exam.example.com/ef1/5 ","logicRules":"[]"}`
	result, err := NormalizeDisplayTitleSettings(raw)
	if err != nil {
		t.Fatal(err)
	}
	var settings map[string]interface{}
	if err := json.Unmarshal([]byte(result), &settings); err != nil {
		t.Fatal(err)
	}
	if settings["answerDownloadFormat"] != "docx" || settings["answerDownloadExpireDays"] != float64(30) || settings["answerDownloadBaseUrl"] != "https://api.example.com:8086" || settings["logicRules"] != "[]" {
		t.Fatalf("normalized settings = %#v", settings)
	}

	for _, invalid := range []string{
		`{"answerDownloadFormat":"zip"}`,
		`{"answerDownloadFormat":1}`,
		`{"answerDownloadExpireDays":0}`,
		`{"answerDownloadExpireDays":31}`,
		`{"answerDownloadExpireDays":"7"}`,
		`{"answerDownloadBaseUrl":"https://api.example.com/base"}`,
		`{"answerDownloadBaseUrl":"https://api.example.com?token=x"}`,
		`{"answerDownloadBaseUrl":"https://user:pass@api.example.com"}`,
	} {
		if _, err := NormalizeDisplayTitleSettings(invalid); err == nil {
			t.Fatalf("accepted invalid download settings: %s", invalid)
		}
	}
}

func TestCopyKeepsDisplayNameSeparateFromManagementName(t *testing.T) {
	copied := buildExamCopy(model.Exam{Title: "内部管理名", Settings: `{"displayTitle":"考生名称"}`}, 1, 2, 1000)
	var settings map[string]interface{}
	if err := json.Unmarshal([]byte(copied.Settings), &settings); err != nil {
		t.Fatal(err)
	}
	if copied.Title != "内部管理名 (副本)" || settings["displayTitle"] != "考生名称" {
		t.Fatalf("copy conflated names: title=%q settings=%v", copied.Title, settings)
	}
}
