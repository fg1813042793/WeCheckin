package setup

import (
	"os"
	"strings"
	"testing"
)

func TestSuccessfulSetupWritesInvalidateTokenConfigurationCacheInService(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	if !strings.Contains(text, "tokenutil.InvalidateSetupCache()") {
		t.Fatal("setup service must invalidate token configuration cache after successful writes")
	}
}
