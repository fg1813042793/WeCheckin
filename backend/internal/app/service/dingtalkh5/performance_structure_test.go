package dingtalkh5

import (
	"os"
	"strings"
	"testing"
)

func TestReviewListSupportsPaginationAndBatchHistories(t *testing.T) {
	typesSrc, err := os.ReadFile("types.go")
	if err != nil {
		t.Fatalf("read types.go: %v", err)
	}
	reviewsSrc, err := os.ReadFile("reviews.go")
	if err != nil {
		t.Fatalf("read reviews.go: %v", err)
	}
	handlerSrc, err := os.ReadFile("../../handler/client/dingtalkh5/handler.go")
	if err != nil {
		t.Fatalf("read handler.go: %v", err)
	}
	helpersSrc, err := os.ReadFile("review_helpers.go")
	if err != nil {
		t.Fatalf("read review_helpers.go: %v", err)
	}
	combined := string(typesSrc) + string(reviewsSrc) + string(handlerSrc) + string(helpersSrc)
	for _, snippet := range []string{
		"type ReviewListResponse struct",
		"Page     int",
		"PageSize int",
		"Total    int64",
		"parsePositiveQueryInt(c, \"page\", 1)",
		"parsePositiveQueryInt(c, \"pageSize\", 20)",
		"Limit(filters.PageSize)",
		"Offset((filters.Page - 1) * filters.PageSize)",
		"historiesByReviewIDs(ctx, collectReviewIDs(reviews))",
	} {
		if !strings.Contains(combined, snippet) {
			t.Fatalf("dingtalk h5 review list should include pagination/batch snippet %q", snippet)
		}
	}
	listBody := string(reviewsSrc)
	if start := strings.Index(listBody, "func ListReviewsContext"); start >= 0 {
		listBody = listBody[start:]
		if end := strings.Index(listBody, "\n}\n\nfunc "); end >= 0 {
			listBody = listBody[:end+3]
		}
	}
	if strings.Contains(listBody, "historiesForReview(ctx, review.ID)") {
		t.Fatalf("ListReviewsContext must not query histories one review at a time")
	}
}

func TestDingTalkH5WorkbenchUsesAggregateCounts(t *testing.T) {
	src, err := os.ReadFile("workbench.go")
	if err != nil {
		t.Fatalf("read workbench.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"workbenchStatusCountsContext(ctx, db, user)",
		"workbenchQueueCountContext(ctx, db, user)",
		"Group(\"status\")",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("workbench should use aggregate counts with %q", snippet)
		}
	}
	if strings.Contains(text, "var reviews []model.DingTalkH5PerfReview") {
		t.Fatalf("workbench should not load full review list")
	}
}

func TestBootstrapIncludesPermissionVersionWithoutFullData(t *testing.T) {
	typesSrc, err := os.ReadFile("types.go")
	if err != nil {
		t.Fatalf("read types.go: %v", err)
	}
	reviewsSrc, err := os.ReadFile("reviews.go")
	if err != nil {
		t.Fatalf("read reviews.go: %v", err)
	}
	combined := string(typesSrc) + string(reviewsSrc)
	for _, snippet := range []string{
		"PermissionVersion",
		"`json:\"permissionVersion\"`",
		"permissionVersionForUserContext(ctx, db, user)",
		"MAX(`grant_edit_time`)",
		"`grant_subject_type`",
	} {
		if !strings.Contains(combined, snippet) {
			t.Fatalf("bootstrap should expose a lightweight permission version with %q", snippet)
		}
	}
	bootstrapBody := combined
	if start := strings.Index(bootstrapBody, "func BootstrapContext"); start >= 0 {
		bootstrapBody = bootstrapBody[start:]
		if end := strings.Index(bootstrapBody, "\n}\n\nfunc "); end >= 0 {
			bootstrapBody = bootstrapBody[:end+3]
		}
	}
	for _, forbidden := range []string{
		"ListReviewsContext",
		"WorkbenchStatsContext",
		"TemplateContext",
		"ListUsersContext",
	} {
		if strings.Contains(bootstrapBody, forbidden) {
			t.Fatalf("bootstrap must not load full business data via %s", forbidden)
		}
	}
}
