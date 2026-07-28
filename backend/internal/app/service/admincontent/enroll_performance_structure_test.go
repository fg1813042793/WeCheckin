package admincontent

import (
	"os"
	"strings"
	"testing"
)

func TestAdminEnrollListBatchesCountQueries(t *testing.T) {
	src, err := os.ReadFile("enroll.go")
	if err != nil {
		t.Fatalf("read enroll.go: %v", err)
	}
	text := string(src)

	forbidden := []string{
		"db.Raw(\n\t\t\t\"SELECT COUNT(DISTINCT uid)",
		"db.Model(&model.EnrollJoin{}).Where(\"`enroll_join_enroll_id` = ?\", eid).Count(&joinCnt)",
	}
	for _, snippet := range forbidden {
		if strings.Contains(text, snippet) {
			t.Fatalf("admin enroll list must avoid per-row count snippet %q", snippet)
		}
	}

	required := []string{
		"loadEnrollUserCountMapContext(ctx, db, list)",
		"loadEnrollJoinCountMapContext(ctx, db, list)",
	}
	for _, snippet := range required {
		if !strings.Contains(text, snippet) {
			t.Fatalf("admin enroll list must batch load counts with %q", snippet)
		}
	}
}
