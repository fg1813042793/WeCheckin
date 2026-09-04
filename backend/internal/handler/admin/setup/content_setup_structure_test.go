package setup

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestDebugTokenConfigResponseKeepsStableJSONFields(t *testing.T) {
	encoded, err := json.Marshal(DebugTokenConfigResponse{})
	if err != nil {
		t.Fatal(err)
	}
	text := string(encoded)
	for _, field := range []string{
		`"user_expire_seconds":0`, `"user_expire_str":""`, `"user_prefix":""`,
		`"admin_expire_seconds":0`, `"admin_expire_str":""`, `"admin_prefix":""`,
		`"dingtalk_h5_expire_seconds":0`, `"dingtalk_h5_expire_str":""`, `"dingtalk_h5_prefix":""`,
	} {
		if !strings.Contains(text, field) {
			t.Fatalf("JSON %s missing %s", text, field)
		}
	}
}

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

func TestAdminSetupWritesUseRequestContextAndServiceOwnedCacheInvalidation(t *testing.T) {
	src, err := os.ReadFile("handler.go")
	if err != nil {
		t.Fatalf("read handler.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"setupservice.SetSetupContext(ctx, key, value, \"\", addIP)",
		"setupservice.SetContentSetupContext(ctx, key, value, addIP)",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("admin setup write must include %s", snippet)
		}
	}
	if strings.Contains(text, "tokenutil.InvalidateSetupCache()") {
		t.Fatal("handler must not own token setup cache invalidation")
	}
}
