package adminuser

import (
	"database/sql"
	"strings"
	"testing"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"wecheckin/backend/internal/model"
)

func openUserListFilterDryRunDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn: &sql.DB{}, SkipInitializeWithVersion: true,
	}), &gorm.Config{DryRun: true, DisableAutomaticPing: true})
	if err != nil {
		t.Fatalf("open dry-run database: %v", err)
	}
	return db
}

func TestApplyUserListFiltersIncludesVisibleFieldsAndDirectManager(t *testing.T) {
	db := openUserListFilterDryRunDB(t)
	status := 2
	filter := UserListFilter{
		Keyword: " Shawn ", Mobile: "138", Status: &status,
		PositionID: 7, ManagerUserID: 9,
	}
	var rows []model.User
	statement := applyUserListFilters(db.Model(&model.User{}), filter, 1700000000000, true).Find(&rows).Statement
	sqlText := normalizeFilterSQL(statement.SQL.String())

	for _, fragment := range []string{
		"user_name LIKE ?", "user_mobile LIKE ?", "user_status = ?", "user_position_id = ?",
		"user_reporting_relations", "employee_user_id = users.id", "manager_user_id = ?",
		"relation_type = ?", "relation_status = ?", "effective_from", "effective_to",
	} {
		if !strings.Contains(sqlText, fragment) {
			t.Fatalf("user list filter query missing %q: %s", fragment, sqlText)
		}
	}

	for _, value := range []interface{}{"%Shawn%", "%138%", status, uint(7), uint(9), int64(1700000000000)} {
		if !containsQueryVar(statement.Vars, value) {
			t.Fatalf("user list filter vars %#v missing %#v", statement.Vars, value)
		}
	}
}

func TestApplyUserListFiltersMatchesRelationAndLegacyRoles(t *testing.T) {
	db := openUserListFilterDryRunDB(t)
	var rows []model.User
	statement := applyUserListFilters(db.Model(&model.User{}), UserListFilter{RoleID: 6}, 0, true).Find(&rows).Statement
	sqlText := normalizeFilterSQL(statement.SQL.String())

	for _, fragment := range []string{"user_role_id = ?", "user_roles", "user_role_user_id = users.id", "user_role_role_id = ?", "user_role_status = ?"} {
		if !strings.Contains(sqlText, fragment) {
			t.Fatalf("user role filter query missing %q: %s", fragment, sqlText)
		}
	}

	legacyStatement := applyUserListFilters(db.Model(&model.User{}), UserListFilter{RoleID: 6}, 0, false).Find(&rows).Statement
	legacySQL := normalizeFilterSQL(legacyStatement.SQL.String())
	if !strings.Contains(legacySQL, "user_role_id = ?") || strings.Contains(legacySQL, "user_roles") {
		t.Fatalf("legacy role filter query = %s", legacySQL)
	}
}

func TestUserListFilterHasAny(t *testing.T) {
	if (UserListFilter{}).HasAny() {
		t.Fatal("empty user list filter must not disable the total cache")
	}
	status := 0
	if !(UserListFilter{Status: &status}).HasAny() {
		t.Fatal("pending status must be treated as an active filter")
	}
}

func containsQueryVar(values []interface{}, expected interface{}) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func normalizeFilterSQL(value string) string {
	return strings.ReplaceAll(strings.Join(strings.Fields(value), " "), "`", "")
}
