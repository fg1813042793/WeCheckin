package adminuser

import (
	"os"
	"strings"
	"testing"
)

func TestAdminUserPageSynchronizesApplicationPermissionsBeforeSave(t *testing.T) {
	src, err := os.ReadFile("../../../../../admin/src/views/user/index.vue")
	if err != nil {
		t.Fatalf("read user page: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"function syncUserApplicationPermissionSelections",
		"syncUserApplicationPermissionSelections('allow')",
		"const allowClientMenuCheckedKeys = computed(() => allowClientMenuKeys.value)",
		"const denyClientMenuCheckedKeys = computed(() => denyClientMenuKeys.value)",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("admin user page must preserve permission selections with %q", snippet)
		}
	}
}
