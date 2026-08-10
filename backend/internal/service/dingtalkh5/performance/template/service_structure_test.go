package template

import (
	"os"
	"strings"
	"testing"
)

func TestSanitizeTemplateNormalizesEditableTemplatePayload(t *testing.T) {
	got, err := sanitizeTemplate(TemplateDTO{
		ObjectiveDefaults: []NextObjective{
			{Target: "  提升客户端交付效率  ", Weight: 120},
			{Target: "   ", Weight: 20},
		},
		NextObjectiveDefaults: []NextObjective{
			{ID: " next-keep ", Target: "  完成下月重点需求  ", Weight: -5},
		},
		GradeLevels: []GradeLevel{
			{Label: " 优秀 ", Grade: " A+ ", Coefficient: 1.56},
			{Label: " ", Grade: " ", Coefficient: 1},
		},
		Values: []ValueTemplate{
			{
				Name:       " 团结一心 ",
				Definition: " 主动协作 ",
				Rubric: []ValueRubric{
					{Label: " 卓越 ", Score: 88, Description: " 可作为团队标杆 "},
					{Label: " ", Score: -1},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("sanitizeTemplate returned error: %v", err)
	}
	if len(got.ObjectiveDefaults) != 1 || got.ObjectiveDefaults[0].Target != "提升客户端交付效率" || got.ObjectiveDefaults[0].Weight != 100 || got.ObjectiveDefaults[0].ID == "" {
		t.Fatalf("objective defaults not normalized: %#v", got.ObjectiveDefaults)
	}
	if len(got.NextObjectiveDefaults) != 1 || got.NextObjectiveDefaults[0].ID != "next-keep" || got.NextObjectiveDefaults[0].Weight != 0 {
		t.Fatalf("next objective defaults not normalized: %#v", got.NextObjectiveDefaults)
	}
	if len(got.GradeLevels) != 1 || got.GradeLevels[0].Label != "优秀" || got.GradeLevels[0].Grade != "A+" || got.GradeLevels[0].Coefficient != 1.6 {
		t.Fatalf("grade levels not normalized: %#v", got.GradeLevels)
	}
	if len(got.Values) != 1 || got.Values[0].Name != "团结一心" || got.Values[0].Definition != "主动协作" || got.Values[0].ID == "" {
		t.Fatalf("values not normalized: %#v", got.Values)
	}
	if len(got.Values[0].Rubric) != 1 || got.Values[0].Rubric[0].Label != "卓越" || got.Values[0].Rubric[0].Score != 88 || got.Values[0].Rubric[0].Description != "可作为团队标杆" {
		t.Fatalf("value rubric not normalized: %#v", got.Values[0].Rubric)
	}
}

func TestSanitizeTemplateRejectsEmptyTemplate(t *testing.T) {
	if _, err := sanitizeTemplate(TemplateDTO{}); err == nil {
		t.Fatalf("empty template should be rejected")
	}
}

func TestTemplateLoadAndSaveDoNotRunFullUserSeeder(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	for _, name := range []string{"LoadTemplateContext", "SaveTemplateContext"} {
		body := templateFunctionBody(text, "func "+name)
		if strings.Contains(body, "EnsureSeedContext(ctx)") {
			t.Fatalf("%s should not run full seed context when only template data is needed", name)
		}
	}
	if !strings.Contains(text, "func defaultTemplateRowDB") {
		t.Fatalf("template load/save should use a lightweight default template helper")
	}
	loadBody := templateFunctionBody(text, "func LoadTemplateContext")
	if strings.Contains(loadBody, "ensureDefaultTemplateContext(ctx)") {
		t.Fatalf("LoadTemplateContext should load/create the default template with one query path")
	}
	if !strings.Contains(loadBody, "defaultTemplateRowDB(db)") {
		t.Fatalf("LoadTemplateContext should reuse the one-query default template loader")
	}
}

func templateFunctionBody(src, signature string) string {
	start := strings.Index(src, signature)
	if start < 0 {
		return ""
	}
	body := src[start:]
	depth := 0
	for index, r := range body {
		switch r {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return body[:index+1]
			}
		}
	}
	return body
}
