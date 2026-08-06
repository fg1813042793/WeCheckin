package dept

import (
	"os"
	"strings"
	"testing"
)

func TestDeptHelpersExposeContextVariants(t *testing.T) {
	src, err := os.ReadFile("dept.go")
	if err != nil {
		t.Fatalf("read dept.go: %v", err)
	}
	text := string(src)
	for _, name := range []string{
		"UserDeptIDContext",
		"UserDeptIDsContext",
		"UserDeptIDsByMiniOpenIDContext",
		"AncestorIDsContext",
		"TopDeptNameContext",
	} {
		if !strings.Contains(text, "func "+name+"(") {
			t.Fatalf("dept helper must expose %s", name)
		}
	}
}

func TestDeptHelpersUseDatabaseWithContext(t *testing.T) {
	src, err := os.ReadFile("dept.go")
	if err != nil {
		t.Fatalf("read dept.go: %v", err)
	}
	if !strings.Contains(string(src), "database.WithContext(ctx)") {
		t.Fatalf("dept helpers should use database.WithContext in context-aware implementations")
	}
}
