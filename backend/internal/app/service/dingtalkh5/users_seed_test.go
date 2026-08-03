package dingtalkh5

import (
	"os"
	"strings"
	"testing"
)

func TestListUsersContextDoesNotRunDefaultUserSeeder(t *testing.T) {
	src, err := os.ReadFile("users.go")
	if err != nil {
		t.Fatalf("read users.go: %v", err)
	}
	body := functionBody(string(src), "func ListUsersContext")
	for _, forbidden := range []string{
		"EnsureSeedContext(ctx)",
		"upsertDefaultPerfUser",
		"defaultUsers()",
	} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("ListUsersContext should not run default user seed work on request path, found %q", forbidden)
		}
	}
}

func TestListUsersContextFiltersByAllowedAccountsInSQL(t *testing.T) {
	usersSrc, err := os.ReadFile("users.go")
	if err != nil {
		t.Fatalf("read users.go: %v", err)
	}
	storeSrc, err := os.ReadFile("user_store.go")
	if err != nil {
		t.Fatalf("read user_store.go: %v", err)
	}
	combined := string(usersSrc) + string(storeSrc)
	for _, snippet := range []string{
		"listPerfUsersByAccountsDB(db, allowed)",
		"func listPerfUsersByAccountsDB(db *gorm.DB, allowed map[string]struct{}) ([]model.DingTalkH5PerfUser, error)",
		"Where(\"`user_mini_openid` IN ? AND `user_status` = 1\", accounts)",
	} {
		if !strings.Contains(combined, snippet) {
			t.Fatalf("ListUsersContext should query only allowed accounts with %q", snippet)
		}
	}
	listBody := functionBody(string(usersSrc), "func listPerfUsersByDataScopeContext")
	if strings.Contains(listBody, "allowedPerfUsers(users, allowed)") {
		t.Fatalf("ListUsersContext should not load all users and filter allowed accounts in memory")
	}
}

func TestUpdateUserContextDoesNotReloadVisibleUsers(t *testing.T) {
	src, err := os.ReadFile("users.go")
	if err != nil {
		t.Fatalf("read users.go: %v", err)
	}
	body := functionBody(string(src), "func UpdateUserContext")
	for _, forbidden := range []string{
		"ListUsersContext(ctx, current)",
		"users, err := ListUsersContext",
	} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("UpdateUserContext should not reload the full visible user list after saving, found %q", forbidden)
		}
	}
	for _, snippet := range []string{
		"if strings.TrimSpace(payload.Password) != \"\" {",
		"updates[\"user_password\"] = next.Password",
	} {
		if !strings.Contains(body, snippet) {
			t.Fatalf("UpdateUserContext should avoid rewriting password unless a new password is submitted, missing %q", snippet)
		}
	}
}

func TestDataScopeUserAccountsUsesUserDeptJoin(t *testing.T) {
	src, err := os.ReadFile("data_scope.go")
	if err != nil {
		t.Fatalf("read data_scope.go: %v", err)
	}
	body := functionBody(string(src), "func userAccountsByDeptIDsContext")
	for _, snippet := range []string{
		"func userAccountsByDeptIDsContext(ctx context.Context, db *gorm.DB, deptIDs []uint) ([]string, error)",
		"Joins(\"JOIN `user_depts`",
		"`user_depts`.`user_dept_user_id` = `users`.`id`",
		"Distinct(\"`users`.`user_mini_openid`\")",
	} {
		if !strings.Contains(body, snippet) {
			t.Fatalf("data-scope user account lookup should use indexed user_depts join, missing %q", snippet)
		}
	}
	for _, forbidden := range []string{
		"`id` IN (SELECT `user_dept_user_id` FROM `user_depts`",
		"WHERE `user_dept_dept_id` IN ?)",
	} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("data-scope user account lookup should not use slow nested IN subquery, found %q", forbidden)
		}
	}
}

func TestDingTalkH5DataScopeUsesMergedExtraPermissions(t *testing.T) {
	src, err := os.ReadFile("data_scope.go")
	if err != nil {
		t.Fatalf("read data_scope.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"extraDataScopeUserAccountsContext(ctx, db, user)",
		"permissionsupport.DataScopeExtrasContext(ctx, db, user.ID, user.RoleID)",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("dingtalk h5 data scope should merge user and multi-role extra permissions with %q", snippet)
		}
	}
	for _, name := range []string{"reviewExtraDataScopeWhereContext", "extraDataScopeUserAccountsContext"} {
		body := functionBody(text, "func "+name)
		if strings.Contains(body, "UserDataScopeExtrasContext") {
			t.Fatalf("%s must not read only user-level data:extra grants", name)
		}
	}
}

func TestPerfUsersHydrateDepartmentFromUserDepts(t *testing.T) {
	src, err := os.ReadFile("user_store.go")
	if err != nil {
		t.Fatalf("read user_store.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"hydratePerfUsersWithUserDeptsDB(db, users)",
		"hydratePerfUserWithUserDeptDB(db, &user)",
		"model.UserDept",
		"model.Department",
		"applyDepartmentPathToPerfUser",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("perf users should prefer user_depts department data, missing %q", snippet)
		}
	}
	for _, name := range []string{"func listPerfUsersDB", "func listPerfUsersByAccountsDB"} {
		body := functionBody(text, name)
		if strings.Contains(body, "hydratePerfUser(&users[index])") {
			t.Fatalf("%s should not only hydrate user_obj metadata; it must also merge user_depts", name)
		}
	}
}

func TestPerfUsersHydratePositionFromUserPositionID(t *testing.T) {
	modelSrc, err := os.ReadFile("../../../model/dingtalk_h5_performance.go")
	if err != nil {
		t.Fatalf("read dingtalk h5 model: %v", err)
	}
	storeSrc, err := os.ReadFile("user_store.go")
	if err != nil {
		t.Fatalf("read user_store.go: %v", err)
	}
	modelText := string(modelSrc)
	storeText := string(storeSrc)
	if !strings.Contains(modelText, "PositionID") || !strings.Contains(modelText, "column:user_position_id") {
		t.Fatalf("dingtalk h5 perf user must map users.user_position_id")
	}
	for _, snippet := range []string{
		"positionNamesByIDDB(db, uniquePositionIDs(users))",
		"model.Position",
		"applyPositionNameToPerfUser",
	} {
		if !strings.Contains(storeText, snippet) {
			t.Fatalf("perf users should hydrate position from positions table, missing %q", snippet)
		}
	}
	if strings.Contains(storeText, "user.Position = strings.TrimSpace(meta.Position)") {
		t.Fatalf("perf users must not use legacy dingtalkH5Performance.position")
	}
}

func TestAuthHydratesPositionAndDepartmentFromAdminTables(t *testing.T) {
	src, err := os.ReadFile("auth.go")
	if err != nil {
		t.Fatalf("read auth.go: %v", err)
	}
	text := string(src)
	for _, name := range []string{"func LoginContext", "func AuthenticateContext"} {
		body := functionBody(text, name)
		if !strings.Contains(body, "hydratePerfUserWithUserDeptDB(db, &user)") {
			t.Fatalf("%s should hydrate user from admin department and position tables", name)
		}
		if strings.Contains(body, "hydratePerfUser(&user)") {
			t.Fatalf("%s must not only hydrate legacy dingtalkH5Performance metadata", name)
		}
	}
}
