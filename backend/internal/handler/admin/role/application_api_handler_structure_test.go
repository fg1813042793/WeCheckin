package role

import (
	"os"
	"strings"
	"testing"
)

func TestRoleHandlerAcceptsApplicationAPIKeys(t *testing.T) {
	src, err := os.ReadFile("handler.go")
	if err != nil {
		t.Fatalf("read handler.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		`clientAPIPermissionKeys := parsePermissionKeys(c.PostForm("clientApiPermissionKeys"))`,
		`dingtalkH5APIPermissionKeys := parsePermissionKeys(c.PostForm("dingtalkH5ApiPermissionKeys"))`,
		"clientAPIPermissionKeys, dingtalkH5APIPermissionKeys",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("role handler must accept application API permission keys with %s", snippet)
		}
	}
}
