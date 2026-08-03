package menu

import (
	"os"
	"strings"
	"testing"
)

func TestAdminMenuHandlerOnlyExposesCurrentUserV2Endpoints(t *testing.T) {
	src, err := os.ReadFile("handler.go")
	if err != nil {
		t.Fatalf("read handler.go: %v", err)
	}
	text := string(src)
	for _, forbidden := range []string{
		"func (h *AdminMenuHandler) GetMenuTree",
		"func (h *AdminMenuHandler) GetMenuList",
		"func (h *AdminMenuHandler) AddMenu",
		"func (h *AdminMenuHandler) EditMenu",
		"func (h *AdminMenuHandler) DelMenu",
		"legacyPermissionType",
		"legacyPermissionKey",
		"@Router ",
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("admin menu handler must not expose legacy menu management snippet %s", forbidden)
		}
	}
	for _, want := range []string{
		"func (h *AdminMenuHandler) GetAdminMenus",
		"func (h *AdminMenuHandler) GetAdminPerms",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("admin menu handler must keep current v2 endpoint snippet %s", want)
		}
	}
}
