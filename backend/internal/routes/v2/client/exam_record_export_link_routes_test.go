package client

import (
	"os"
	"strings"
	"testing"
)

func TestPublicExamRecordExportLinkRoute(t *testing.T) {
	data, err := os.ReadFile("routes.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(data)
	for _, snippet := range []string{
		`h.GET("/api/v2/public/exam-record-exports/download", examRecordExportLink.Download)`,
		`examrecordexportlinkhttp.NewHandler(`,
	} {
		if !strings.Contains(source, snippet) {
			t.Fatalf("public exam export route missing %q", snippet)
		}
	}
}
