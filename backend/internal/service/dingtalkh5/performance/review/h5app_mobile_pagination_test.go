package review

import (
	"os"
	"strings"
	"testing"
)

func readH5PerformancePage(t *testing.T, name string) string {
	t.Helper()
	src, err := os.ReadFile("../../../../../../h5app/src/pages/performance/" + name + ".vue")
	if err != nil {
		t.Fatalf("read h5app performance page %s: %v", name, err)
	}
	return string(src)
}

func TestH5AppMobileReviewListsUseServerPaginationAndLoadMore(t *testing.T) {
	shared, err := os.ReadFile("../../../../../../h5app/src/pages/performance/components/mobilePagination.ts")
	if err != nil {
		t.Fatalf("read h5app mobile pagination helper: %v", err)
	}
	for _, snippet := range []string{
		"export const MOBILE_REVIEW_PAGE_SIZE = 20",
		"function mobileReviewListParams",
		"function showMobileLoadMore",
		"function showMobileNoMore",
	} {
		if !strings.Contains(string(shared), snippet) {
			t.Fatalf("mobile pagination helper should include %q", snippet)
		}
	}

	cases := []struct {
		page     string
		state    string
		loadMore string
	}{
		{page: "summary", state: "summaryMobilePagination", loadMore: "loadMoreSummaryRows"},
		{page: "history", state: "historyMobilePagination", loadMore: "loadMoreHistoryRows"},
		{page: "review", state: "managerMobilePagination", loadMore: "loadMoreManagerRows"},
		{page: "hrbp", state: "hrbpMobilePagination", loadMore: "loadMoreHrbpRows"},
		{page: "mine", state: "mineMobilePagination", loadMore: "loadMoreMineRows"},
	}
	for _, tt := range cases {
		t.Run(tt.page, func(t *testing.T) {
			src := readH5PerformancePage(t, tt.page)
			for _, snippet := range []string{
				"createMobilePaginationState()",
				tt.state,
				"mobileReviewListParams(isMobilePage.value, " + tt.state + ")",
				"updateMobilePaginationTotal(" + tt.state,
				tt.loadMore,
				"showMobileLoadMore(" + tt.state,
				"showMobileNoMore(" + tt.state,
				`class="mobile-list-pagination"`,
			} {
				if !strings.Contains(src, snippet) {
					t.Fatalf("%s mobile pagination should include %q", tt.page, snippet)
				}
			}
		})
	}
}
