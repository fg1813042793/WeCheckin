package setup

import (
	"os"
	"strings"
	"testing"
)

func TestAdminSetupProvidesAuthenticatedContentRead(t *testing.T) {
	src, err := os.ReadFile("handler.go")
	if err != nil {
		t.Fatalf("read handler.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"func (h *AdminSetupHandler) GetContentSetup",
		"setupservice.GetSetupContext(ctx, key)",
		"response.JSON(c, setup.Value)",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("admin content setup read must include %s", snippet)
		}
	}
}
