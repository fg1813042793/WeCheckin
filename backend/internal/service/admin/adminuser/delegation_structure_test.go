package adminuser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAdminUserCreationCarriesDelegatingAdminIdentity(t *testing.T) {
	handler, err := os.ReadFile(filepath.Join("..", "..", "..", "handler", "admin", "user", "handler.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, snippet := range []string{
		`adminVal, _ := c.Get("admin")`,
		"AddUserWithManagerAndAdminAccessForAdminContext(ctx",
		"admin.ID",
	} {
		if !strings.Contains(string(handler), snippet) {
			t.Fatalf("user creation must pass current admin with %q", snippet)
		}
	}
}

