package model

import (
	"os"
	"strings"
	"testing"
)

func TestModelLayerDoesNotDefineLegacyMenuTable(t *testing.T) {
	src, err := os.ReadFile("rbac.go")
	if err != nil {
		t.Fatalf("read rbac.go: %v", err)
	}
	text := string(src)
	for _, forbidden := range []string{
		"type Menu struct",
		"column:menu_name",
		"column:menu_parent_id",
		"column:menu_path",
		"column:menu_perms",
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("model layer must not define legacy menus table snippet %s", forbidden)
		}
	}
}
