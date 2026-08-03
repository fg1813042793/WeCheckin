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

func TestReviewListDefaultsToSkippingHistories(t *testing.T) {
	handlerSrc, err := os.ReadFile("../../handler/client/dingtalkh5/handler.go")
	if err != nil {
		t.Fatalf("read handler.go: %v", err)
	}
	text := string(handlerSrc)
	if !strings.Contains(text, `SkipHistory:     parseBoolQuery(c, "skipHistory", true)`) {
		t.Fatalf("review list should skip histories by default")
	}
	if !strings.Contains(text, `!parseBoolQuery(c, "includeHistory", false)`) {
		t.Fatalf("review list should load histories only when includeHistory is explicitly enabled")
	}
}

func TestReviewListHydratesParticipantNamesForFrontend(t *testing.T) {
	typesSrc, err := os.ReadFile("types.go")
	if err != nil {
		t.Fatalf("read types.go: %v", err)
	}
	reviewsSrc, err := os.ReadFile("reviews.go")
	if err != nil {
		t.Fatalf("read reviews.go: %v", err)
	}
	helpersSrc, err := os.ReadFile("review_helpers.go")
	if err != nil {
		t.Fatalf("read review_helpers.go: %v", err)
	}
	combined := string(typesSrc) + string(reviewsSrc) + string(helpersSrc)
	for _, snippet := range []string{
		"EmployeeName",
		"`json:\"employeeName\"`",
		"ManagerName",
		"`json:\"managerName\"`",
		"HRBPName",
		"`json:\"hrbpName\"`",
		"HRBPReviewerName",
		"`json:\"hrbpReviewerName\"`",
		"func collectReviewParticipantAccounts(reviews []model.DingTalkH5PerfReview) []string",
		"participants, err := usersByAccounts(ctx, collectReviewParticipantAccounts(reviews))",
		"reviewDTOWithUsers(review, historiesByID[review.ID], participants)",
		"func reviewDTOWithUsers(review model.DingTalkH5PerfReview, histories []model.DingTalkH5PerfHistory, users map[string]*model.DingTalkH5PerfUser) ReviewDTO",
		"result.EmployeeName = reviewUserName(users, review.EmployeeAccount)",
		"result.ManagerName = reviewUserName(users, review.ManagerAccount)",
		"result.HRBPName = reviewUserName(users, review.HRBPAccount)",
		"result.HRBPReviewerName = reviewUserName(users, review.HRBPReviewerAccount)",
	} {
		if !strings.Contains(combined, snippet) {
			t.Fatalf("review list should hydrate participant names for frontend with %q", snippet)
		}
	}
}

func TestReviewParticipantNameLookupDoesNotHydrateDepartmentsOrPositions(t *testing.T) {
	src, err := os.ReadFile("review_helpers.go")
	if err != nil {
		t.Fatalf("read review_helpers.go: %v", err)
	}
	body := functionBody(string(src), "func usersByAccounts")
	if strings.Contains(body, "hydratePerfUsersWithUserDeptsDB") {
		t.Fatalf("review participant name lookup should not hydrate departments or positions")
	}
	if !strings.Contains(body, "Select(\"`user_mini_openid`, `user_name`\")") {
		t.Fatalf("review participant name lookup should select only account and name")
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
	snapshotSrc, err := os.ReadFile("permission_snapshot.go")
	if err != nil {
		t.Fatalf("read permission_snapshot.go: %v", err)
	}
	combined := string(typesSrc) + string(reviewsSrc) + string(snapshotSrc)
	for _, snippet := range []string{
		"PermissionVersion",
		"`json:\"permissionVersion\"`",
		"snapshot.version",
		"permissionVersionFallback(user)",
		"`grant_edit_time`",
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
