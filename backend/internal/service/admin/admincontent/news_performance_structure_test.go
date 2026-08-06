package admincontent

import (
	"os"
	"strings"
	"testing"
)

func TestAdminNewsListUsesLightweightColumns(t *testing.T) {
	src, err := os.ReadFile("news.go")
	if err != nil {
		t.Fatalf("read news.go: %v", err)
	}
	if !strings.Contains(string(src), "Select(newsservice.ListColumns)") {
		t.Fatalf("admin news list should reuse lightweight news list columns")
	}
}
