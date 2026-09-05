package infrastructure

import (
	"database/sql"
	"reflect"
	"strings"
	"testing"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/modules/inappnotification/application"
)

func TestApplyNotificationRecordFiltersIncludesAllRecordFields(t *testing.T) {
	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn: &sql.DB{}, SkipInitializeWithVersion: true,
	}), &gorm.Config{DryRun: true, DisableAutomaticPing: true})
	if err != nil {
		t.Fatalf("open dry-run database: %v", err)
	}
	readStatus := 0
	query := application.NotificationRecordQuery{
		Title: " 系统%_!通知 ", SourceType: "workflow", Type: "task_arrived",
		IsRead: &readStatus, AddTimeFrom: 100, AddTimeTo: 200,
	}
	var rows []model.Notify
	statement := applyNotificationRecordFilters(db.Model(&model.Notify{}), query, []string{"7", "9"}).Find(&rows).Statement
	sqlText := statement.SQL.String()

	for _, fragment := range []string{
		"notify_admin_deleted_at = ?", "notify_title LIKE ?", "ESCAPE '!'", "notify_source_type = ?", "notify_type = ?",
		"notify_is_read = ?", "notify_add_time >= ?", "notify_add_time <= ?", "notify_user_id IN",
	} {
		if !strings.Contains(sqlText, fragment) {
			t.Fatalf("record query missing %q: %s", fragment, sqlText)
		}
	}
	want := []interface{}{int64(0), "%系统!%!_!!通知%", "workflow", "task_arrived", 0, int64(100), int64(200), "7", "9"}
	if !reflect.DeepEqual(statement.Vars, want) {
		t.Fatalf("query vars = %#v, want %#v", statement.Vars, want)
	}
}
