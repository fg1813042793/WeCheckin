package exam

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAdminExamListBatchLoadsSubmittedResponseCounts(t *testing.T) {
	checks := []struct {
		path      string
		required  []string
		forbidden []string
	}{
		{
			path: "service.go",
			required: []string{
				"func (s *Service) ResponseCountsContext(ctx context.Context, examIDs []uint) (map[uint]int, error)",
				"Where(\"`exam_r_exam_id` IN ? AND `exam_r_status` >= ?\", examIDs, 1)",
				"Group(\"`exam_r_exam_id`\")",
			},
		},
		{
			path: filepath.Join("..", "..", "..", "handler", "admin", "exam", "handler.go"),
			required: []string{
				"countMap, err := h.svc.ResponseCountsContext(ctx, ids)",
				"newExamListItem(item, countMap[item.ID])",
			},
			forbidden: []string{"ResponseCountsContext(ctx, []uint{item.ID})"},
		},
		{
			path: filepath.Join("..", "..", "..", "handler", "admin", "exam", "dto.go"),
			required: []string{
				"ResponseCount int    `json:\"responseCount\"`",
				"func newExamListItem(exam model.Exam, responseCount int) examListItem",
				"ResponseCount: responseCount",
			},
		},
	}

	for _, check := range checks {
		source, err := os.ReadFile(check.path)
		if err != nil {
			t.Fatalf("read %s: %v", check.path, err)
		}
		text := string(source)
		for _, snippet := range check.required {
			if !strings.Contains(text, snippet) {
				t.Fatalf("%s must include %q", check.path, snippet)
			}
		}
		for _, snippet := range check.forbidden {
			if strings.Contains(text, snippet) {
				t.Fatalf("%s must not include per-exam count query %q", check.path, snippet)
			}
		}
	}
}
