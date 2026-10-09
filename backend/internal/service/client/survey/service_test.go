package survey

import (
	"testing"

	"wecheckin/backend/internal/model"
)

// 不依赖 DB 的纯函数测试

func TestSplitComma(t *testing.T) {
	tests := []struct {
		in   string
		want []string
	}{
		{"", nil},
		{"1", []string{"1"}},
		{"1,2,3", []string{"1", "2", "3"}},
		{"1,2,", []string{"1", "2"}},
		{",", []string{""}},
	}
	for _, tt := range tests {
		got := splitComma(tt.in)
		if len(got) != len(tt.want) {
			t.Errorf("splitComma(%q) len=%d, want %d", tt.in, len(got), len(tt.want))
			continue
		}
		for i, v := range got {
			if v != tt.want[i] {
				t.Errorf("splitComma(%q)[%d] = %q, want %q", tt.in, i, v, tt.want[i])
			}
		}
	}
}

func TestContains(t *testing.T) {
	if !contains("1,2,3", 2) {
		t.Error("should contain 2")
	}
	if contains("1,2,3", 4) {
		t.Error("should not contain 4")
	}
	if contains("", 1) {
		t.Error("empty should not contain")
	}
}

func TestUserIDToStr(t *testing.T) {
	if userIDToStr(123, true) != "" {
		t.Error("anonymous should be empty")
	}
	if userIDToStr(123, false) != "123" {
		t.Error("non-anonymous should be string id")
	}
	if userIDToStr(0, false) != "0" {
		t.Error("0 should be '0'")
	}
}

func TestParseUAOpenHarmony(t *testing.T) {
	ua := "Mozilla/5.0 (Phone; OpenHarmony 6.1) AppleWebKit/537.36 Chrome/132.0 Mobile MicroMessenger/8.0"
	browser, deviceType, platformType := parseUA(ua)
	if browser != "微信" || deviceType != "移动端" || platformType != "OpenHarmony" {
		t.Fatalf("parseUA OpenHarmony = (%q, %q, %q)", browser, deviceType, platformType)
	}
}

func TestNormalizeResponseDeviceMetadataFillsLegacyPlatform(t *testing.T) {
	response := &model.SurveyResponse{
		Device:       "Mozilla/5.0 (Phone; OpenHarmony 6.1) Mobile MicroMessenger/8.0",
		Browser:      "微信",
		DeviceType:   "移动端",
		PlatformType: "-",
	}
	normalizeResponseDeviceMetadata(response)
	if response.PlatformType != "OpenHarmony" {
		t.Fatalf("legacy platform = %q, want OpenHarmony", response.PlatformType)
	}
}

func TestSurveySubmissionLimitsAcceptsNumericStrings(t *testing.T) {
	deviceLimit, ipLimit := surveySubmissionLimits(`{"deviceLimit":"2","ipLimit":"3"}`)
	if deviceLimit != 2 || ipLimit != 3 {
		t.Fatalf("surveySubmissionLimits numeric strings = (%d,%d), want (2,3)", deviceLimit, ipLimit)
	}
}

func TestSurveySubmissionLimitsIgnoresInvalidValues(t *testing.T) {
	deviceLimit, ipLimit := surveySubmissionLimits(`{"deviceLimit":-1,"ipLimit":"bad"}`)
	if deviceLimit != 0 || ipLimit != 0 {
		t.Fatalf("surveySubmissionLimits invalid values = (%d,%d), want (0,0)", deviceLimit, ipLimit)
	}
}

func TestValidateSurveySchemaForUpdateRejectsEmpty(t *testing.T) {
	if err := validateSurveySchemaForUpdate(""); err == nil || err.Error() != "schema 必填" {
		t.Fatalf("validateSurveySchemaForUpdate empty error = %v, want schema 必填", err)
	}
}

func TestValidateSurveySchemaForUpdateAcceptsValidSchema(t *testing.T) {
	if err := validateSurveySchemaForUpdate(`{"version":"2.0","questions":[]}`); err != nil {
		t.Fatalf("validateSurveySchemaForUpdate valid schema error = %v", err)
	}
}
