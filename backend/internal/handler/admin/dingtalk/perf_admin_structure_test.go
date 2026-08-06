package dingtalk

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDingTalkPerfAdminHandlerStructure(t *testing.T) {
	matches, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob handler files: %v", err)
	}
	var builder strings.Builder
	for _, path := range matches {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		builder.Write(src)
	}
	text := builder.String()
	for _, want := range []string{
		"GetPerfReviews",
		"GetPerfReviewDetail",
		"DeletePerfReview",
		"DeletePerfReviews",
		"GetPerfHistories",
		"DeletePerfHistory",
		"DeletePerfHistories",
		"DingTalkH5PerfReview",
		"DingTalkH5PerfHistory",
		"`deleted_at` = 0",
		"review_no",
		"employee_account",
		"period",
		"status",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("dingtalk performance admin handler missing %q", want)
		}
	}
}

func TestDingTalkPerfAdminFrontendStructure(t *testing.T) {
	routeSrc, err := os.ReadFile("../../../../../../admin/src/router/adminRoutes.ts")
	if err != nil {
		t.Fatalf("read admin routes: %v", err)
	}
	for _, want := range []string{
		`path: 'dingtalk/perf-reviews'`,
		`name: 'DingTalkPerfReviews'`,
		`../views/dingtalk-perf-reviews/index.vue`,
		`path: 'dingtalk/perf-histories'`,
		`name: 'DingTalkPerfHistories'`,
		`../views/dingtalk-perf-histories/index.vue`,
	} {
		if !strings.Contains(string(routeSrc), want) {
			t.Fatalf("admin routes missing dingtalk performance route %q", want)
		}
	}

	reviewSrc, err := os.ReadFile("../../../../../../admin/src/views/dingtalk-perf-reviews/index.vue")
	if err != nil {
		t.Fatalf("read dingtalk performance review page: %v", err)
	}
	for _, want := range []string{
		"绩效考评单",
		"/api/v2/admin/dingtalk/perf-reviews",
		"admin:menu:dingtalk:perf-reviews:detail",
		"admin:menu:dingtalk:perf-reviews:del",
		"ElMessageBox.confirm",
		"admin-toolbar",
		"admin-pagination",
	} {
		if !strings.Contains(string(reviewSrc), want) {
			t.Fatalf("dingtalk performance review page missing %q", want)
		}
	}
	if got := strings.Count(string(reviewSrc), `class="admin-toolbar"`); got < 2 {
		t.Fatalf("dingtalk performance review page should use separated filter/action toolbars, got %d", got)
	}

	historySrc, err := os.ReadFile("../../../../../../admin/src/views/dingtalk-perf-histories/index.vue")
	if err != nil {
		t.Fatalf("read dingtalk performance history page: %v", err)
	}
	for _, want := range []string{
		"绩效流转记录",
		"/api/v2/admin/dingtalk/perf-histories",
		"admin:menu:dingtalk:perf-histories:del",
		"批量删除",
		"ElMessageBox.confirm",
		"reviewNo",
		"byAccount",
		"action",
		"admin-toolbar",
		"admin-pagination",
	} {
		if !strings.Contains(string(historySrc), want) {
			t.Fatalf("dingtalk performance history page missing %q", want)
		}
	}
	if got := strings.Count(string(historySrc), `class="admin-toolbar"`); got < 2 {
		t.Fatalf("dingtalk performance history page should use separated filter/action toolbars, got %d", got)
	}
}
