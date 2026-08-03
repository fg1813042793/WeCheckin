package adminuser

import (
	"os"
	"strings"
	"testing"
)

func TestUserListAvoidsPerRowDeptQueries(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)

	forbidden := []string{
		"deptsupport.UserDeptIDsContext(ctx, u.ID)",
	}
	for _, snippet := range forbidden {
		if strings.Contains(text, snippet) {
			t.Fatalf("user list must avoid per-row query snippet %q", snippet)
		}
	}

	required := []string{
		"loadUserDeptIDMapContext(ctx, db, list)",
	}
	for _, snippet := range required {
		if !strings.Contains(text, snippet) {
			t.Fatalf("user list must batch load related data with %q", snippet)
		}
	}
}

func TestUserListSelectsOnlyListColumns(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)

	required := []string{
		"var userListColumns = []string{",
		"Select(userListColumns)",
	}
	for _, snippet := range required {
		if !strings.Contains(text, snippet) {
			t.Fatalf("user list should explicitly select lightweight list columns with %q", snippet)
		}
	}

	forbiddenColumns := []string{
		"user_forms",
		"user_obj",
		"user_password",
	}
	start := strings.Index(text, "var userListColumns = []string{")
	if start < 0 {
		t.Fatalf("userListColumns declaration missing")
	}
	end := strings.Index(text[start:], "}")
	if end < 0 {
		t.Fatalf("userListColumns declaration is incomplete")
	}
	columnsBlock := text[start : start+end]
	for _, column := range forbiddenColumns {
		if strings.Contains(columnsBlock, column) {
			t.Fatalf("user list columns should not include heavy/sensitive column %q", column)
		}
	}
}

func TestUserListDefaultOrderMatchesIndex(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	if !strings.Contains(text, "orderClause = \"`user_add_time` DESC, `id` DESC\"") {
		t.Fatalf("user list default order should match idx_users_add_time_id")
	}
}

func TestUserListSearchCoversIndexedVisibleFields(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"`user_name` LIKE ?",
		"`user_mobile` LIKE ?",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("user list keyword search should include indexed field %q", snippet)
		}
	}
}

func TestUserListCachesUnfilteredTotalCount(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	required := []string{
		"useTotalCountCache := keyword == \"\" && !hasDataScopeFilter",
		"getUserTotalCountCache(now)",
		"setUserTotalCountCache(total, now)",
		"invalidateUserTotalCountCache()",
	}
	for _, snippet := range required {
		if !strings.Contains(text, snippet) {
			t.Fatalf("user list should cache unfiltered total count with %q", snippet)
		}
	}
}

func TestUserListUsesRequestContextForAvatarStaticDomain(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	if !strings.Contains(text, "media.FullURLWithStaticDomainContext(ctx, u.Pic)") {
		t.Fatalf("user list avatars should use request context and cached static domain")
	}
	if strings.Contains(text, "media.FullURLWithStaticDomain(u.Pic)") {
		t.Fatalf("user list avatars should not call static domain lookup without request context")
	}
}

func TestUserPositionFieldsAreReturnedWithoutPerRowQueries(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)

	required := []string{
		"PositionID",
		"PositionName",
		"user_position_id",
		"loadPositionNameMapContext(ctx, db, list)",
	}
	for _, snippet := range required {
		if !strings.Contains(text, snippet) {
			t.Fatalf("user service must expose and batch load position data with %q", snippet)
		}
	}

	forbidden := []string{
		"positionNameForIDContext(ctx, u.PositionID)",
	}
	for _, snippet := range forbidden {
		if strings.Contains(text, snippet) {
			t.Fatalf("user list must avoid per-row position query snippet %q", snippet)
		}
	}
}
