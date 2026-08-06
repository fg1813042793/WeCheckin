package role

import (
	"os"
	"strings"
	"testing"
)

func TestAdminRoleHandlerAcceptsApplicationMenuPermissionKeys(t *testing.T) {
	src, err := os.ReadFile("handler.go")
	if err != nil {
		t.Fatalf("read handler.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		`c.PostForm("clientMenuKeys")`,
		`c.PostForm("dingtalkH5MenuKeys")`,
		"parsePermissionKeys",
		"clientMenuKeys, dingtalkH5MenuKeys",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("role handler must accept app menu permission keys with %s", snippet)
		}
	}
}
