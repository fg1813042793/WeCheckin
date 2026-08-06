package adminuser

import (
	"os"
	"strings"
	"testing"
)

func TestUserManagementSingleRecordOperationsUseAdminScope(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	required := []string{
		"func userVisibleQueryContext(ctx context.Context, db *gorm.DB, adminID uint) (*gorm.DB, error)",
		"access.UserDataScopeFilterWithDBContext(ctx, db, &admin)",
		"func GetUserByIDForAdminContext(ctx context.Context, id string, adminID uint)",
		"func EditUserForAdminContext(",
		"func DelUserForAdminContext(ctx context.Context, id string, adminID uint) error",
		"func StatusUserForAdminContext(ctx context.Context, id string, status int, reason string, adminID uint) error",
		"func ResetUserPasswordForAdminContext(ctx context.Context, id string, adminID uint) error",
	}
	for _, snippet := range required {
		if !strings.Contains(text, snippet) {
			t.Fatalf("user management must protect single-record operation with %q", snippet)
		}
	}
}
