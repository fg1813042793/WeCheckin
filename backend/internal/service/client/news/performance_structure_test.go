package news

import (
	"os"
	"strings"
	"testing"
)

func TestNewsListUsesLightweightColumns(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	required := []string{
		"var ListColumns = []string{",
		"Select(ListColumns)",
	}
	for _, snippet := range required {
		if !strings.Contains(text, snippet) {
			t.Fatalf("news list should select lightweight columns with %q", snippet)
		}
	}

	start := strings.Index(text, "var ListColumns = []string{")
	if start < 0 {
		t.Fatalf("ListColumns declaration missing")
	}
	end := strings.Index(text[start:], "}")
	if end < 0 {
		t.Fatalf("ListColumns declaration is incomplete")
	}
	columnsBlock := text[start : start+end]
	for _, column := range []string{"news_content", "news_forms", "news_obj"} {
		if strings.Contains(columnsBlock, column) {
			t.Fatalf("news list columns should not include heavy column %q", column)
		}
	}
}
