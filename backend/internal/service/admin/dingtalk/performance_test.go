package dingtalk

import (
	"encoding/json"
	"strings"
	"testing"

	"wecheckin/backend/internal/model"
)

func TestPerfReviewListItemKeepsAdminJSONContract(t *testing.T) {
	item := buildPerfReviewListItem(model.DingTalkH5PerfReview{
		ID: 7, ReviewNo: "P-7", EmployeeAccount: "alice", Status: "manager_review",
		ObjectiveSourceReviewNo: "P-6", ObjectiveSourcePeriod: "2026-Q2",
	})
	encoded, err := json.Marshal(item)
	if err != nil {
		t.Fatal(err)
	}
	text := string(encoded)
	for _, field := range []string{`"reviewNo":"P-7"`, `"employeeAccount":"alice"`, `"statusLabel":"上级评价"`, `"objectiveSourceReviewNo":"P-6"`} {
		if !strings.Contains(text, field) {
			t.Fatalf("JSON %s missing %s", text, field)
		}
	}
}

func TestNormalizePaginationKeepsExistingLimits(t *testing.T) {
	page, pageSize := normalizePagination(0, 101)
	if page != 1 || pageSize != 100 {
		t.Fatalf("pagination = %d/%d", page, pageSize)
	}
}
