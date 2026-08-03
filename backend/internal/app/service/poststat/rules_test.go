package poststat

import (
	"strings"
	"testing"

	"wecheckin/backend/internal/app/formkit/report"
)

func TestParseRulesKeepsOnlyPostStatRules(t *testing.T) {
	settings := `{"logicRules":[{"id":"1","action":"postStat","notifyAdmin":true},{"id":"2","action":"show"}]}`

	rules := parseRules(settings)

	if len(rules) != 1 {
		t.Fatalf("expected one postStat rule, got %#v", rules)
	}
	if rules[0].ID != "1" || !rules[0].NotifyAdmin {
		t.Fatalf("unexpected parsed rule: %#v", rules[0])
	}
}

func TestParseRulesSupportsStringEncodedRules(t *testing.T) {
	settings := `{"logicRules":"[{\"id\":\"string-rule\",\"action\":\"postStat\",\"webhookType\":\"dingtalk\"}]"}`

	rules := parseRules(settings)

	if len(rules) != 1 {
		t.Fatalf("expected one postStat rule from string payload, got %#v", rules)
	}
	if rules[0].ID != "string-rule" || rules[0].WebhookType != "dingtalk" {
		t.Fatalf("unexpected parsed rule: %#v", rules[0])
	}
}

func TestBuildResultTextAggregatesAndSortsDistribution(t *testing.T) {
	text := buildResultText([]report.FieldStat{
		{Title: "A", Type: "radio", Dist: map[string]int{"选项A": 2, "选项B": 1}},
		{Title: "B", Type: "checkbox", Dist: map[string]int{"选项A": 1}},
		{Title: "说明", Type: "description", Dist: map[string]int{"忽略": 100}},
	}, "value")

	if !strings.Contains(text, "统计类型（选项值）") {
		t.Fatalf("expected value mode label, got %q", text)
	}
	if !strings.Contains(text, "选项A 个数 3 比重 75.0%") {
		t.Fatalf("expected aggregated option A result, got %q", text)
	}
	if strings.Contains(text, "忽略") {
		t.Fatalf("description fields should be ignored, got %q", text)
	}
}
