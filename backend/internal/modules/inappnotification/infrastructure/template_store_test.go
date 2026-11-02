package infrastructure

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"wecheckin/backend/internal/model/notification"
	"wecheckin/backend/internal/modules/inappnotification/application"
)

func TestTemplateStoreOptimisticWritesAndDuplicateName(t *testing.T) {
	for _, tc := range []struct {
		name        string
		affected    int64
		dbErr, want error
	}{
		{name: "saved", affected: 1}, {name: "concurrent update", want: application.ErrTemplateVersionConflict},
		{name: "duplicate", dbErr: &mysqlDriver.MySQLError{Number: 1062}, want: application.ErrTemplateNameExists},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store, script := templateTestStore(t, templateDBStep{contains: "UPDATE `dingtalk_notification_templates`", affected: tc.affected, err: tc.dbErr})
			row := notificationmodel.DingTalkTemplate{ID: 7, Name: "名称", Enabled: false, Version: 4, UpdatedBy: 9, UpdatedAt: 42}
			err := store.UpdateTemplate(context.Background(), &row, 3)
			if !errors.Is(err, tc.want) {
				t.Fatalf("error=%v want=%v", err, tc.want)
			}
			call := script.calls[0]
			if !strings.Contains(call.query, "WHERE id = ? AND version = ?") || !reflect.DeepEqual(call.args[len(call.args)-2:], []any{int64(7), int64(3)}) || strings.Contains(call.query, "created_by") {
				t.Fatalf("unsafe update: %#v", call)
			}
		})
	}
	store, script := templateTestStore(t, templateDBStep{contains: "DELETE FROM `dingtalk_notification_templates`", affected: 0})
	if err := store.DeleteTemplate(context.Background(), 7, 3); !errors.Is(err, application.ErrTemplateVersionConflict) {
		t.Fatalf("delete=%v", err)
	}
	if !strings.Contains(script.calls[0].query, "id = ? AND version = ?") {
		t.Fatalf("unsafe delete=%s", script.calls[0].query)
	}
}

func TestTemplateStoreCreatesDisabledWithAuditFieldsAndReadsMissing(t *testing.T) {
	store, script := templateTestStore(t, templateDBStep{contains: "INSERT INTO `dingtalk_notification_templates`", affected: 1, lastID: 11})
	row := notificationmodel.DingTalkTemplate{Name: "停用模板", Version: 1, CreatedBy: 9, UpdatedBy: 9, CreatedAt: 100, UpdatedAt: 100}
	if err := store.CreateTemplate(context.Background(), &row); err != nil || row.ID != 11 {
		t.Fatalf("create=%#v %v", row, err)
	}
	if !strings.Contains(script.calls[0].query, "`enabled`") || script.calls[0].args[1] != false {
		t.Fatalf("disabled insert=%#v", script.calls[0])
	}
	store, _ = templateTestStore(t, templateDBStep{contains: "FROM `dingtalk_notification_templates`", columns: []string{"id"}})
	if _, err := store.GetTemplate(context.Background(), 99); !errors.Is(err, application.ErrTemplateNotFound) {
		t.Fatalf("get error=%v", err)
	}
}

func TestTemplateStoreEnabledOptionsBoundAndEscapeSearch(t *testing.T) {
	store, script := templateTestStore(t,
		templateDBStep{contains: "count(*)", columns: []string{"count"}, rows: [][]driver.Value{{int64(1)}}},
		templateDBStep{contains: "FROM `dingtalk_notification_templates`", columns: []string{"id", "name", "enabled", "version", "target_kind"}, rows: [][]driver.Value{{int64(1), "模板", true, int64(3), "app_home"}}})
	disabled := false
	rows, total, err := store.ListTemplates(context.Background(), application.DingTalkTemplateQuery{Keyword: "%_!", TargetKind: "workflow_start", Page: 2, PageSize: 5, Enabled: &disabled}, true)
	if err != nil || total != 1 || len(rows) != 1 {
		t.Fatalf("rows=%#v total=%d err=%v", rows, total, err)
	}
	call := script.calls[1]
	if !strings.Contains(call.query, "enabled = ?") || !strings.Contains(call.query, "target_kind = ?") || !strings.Contains(call.query, "LIMIT ? OFFSET ?") || !strings.Contains(call.query, "ESCAPE '!'") {
		t.Fatalf("unbounded options: %#v", call)
	}
	if !reflect.DeepEqual(call.args, []any{true, "workflow_start", "%!%!_!!%", "%!%!_!!%", int64(5), int64(5)}) {
		t.Fatalf("args=%#v", call.args)
	}
}

func TestTemplateStoreWorkflowOptionsUsePublishedSchemaAndManagementName(t *testing.T) {
	store, script := templateTestStore(t,
		templateDBStep{contains: "count(*)", columns: []string{"count"}, rows: [][]driver.Value{{int64(1)}}},
		templateDBStep{contains: "FROM `workflow_definitions`", columns: []string{"id", "definition_name", "definition_display_name"}, rows: [][]driver.Value{{int64(17), "内部名称", "显示名称"}}})
	rows, total, err := store.TemplateWorkflowOptions(context.Background(), application.DingTalkTemplateQuery{ID: 17, Page: 1, PageSize: 10})
	if err != nil || total != 1 || len(rows) != 1 || rows[0].Name != "内部名称" {
		t.Fatalf("rows=%#v error=%v", rows, err)
	}
	for _, call := range script.calls {
		for _, part := range []string{"definition_status = ?", "definition_current_version > 0", "definition_deleted_at = 0", "id = ?"} {
			if !strings.Contains(call.query, part) {
				t.Errorf("query missing %s: %s", part, call.query)
			}
		}
	}
	store, _ = templateTestStore(t, templateDBStep{contains: "FROM `workflow_definitions`", columns: []string{"id", "definition_name", "definition_display_name"}})
	if _, err = store.PublishedTemplateWorkflow(context.Background(), 17); !errors.Is(err, application.ErrTemplateWorkflowUnavailable) {
		t.Fatalf("missing workflow=%v", err)
	}
}

func TestTemplateStoreSchemaFailureIsActionable(t *testing.T) {
	store, _ := templateTestStore(t, templateDBStep{contains: "FROM `dingtalk_notification_templates`", err: &mysqlDriver.MySQLError{Number: 1146, Message: "private SQL details"}})
	if _, err := store.GetTemplate(context.Background(), 1); !errors.Is(err, application.ErrTemplateSchemaMissing) {
		t.Fatalf("schema err=%v", err)
	}
}

type templateDBStep struct {
	contains         string
	columns          []string
	rows             [][]driver.Value
	affected, lastID int64
	err              error
}
type templateDBCall struct {
	query string
	args  []any
}
type templateDBScript struct {
	t     *testing.T
	steps []templateDBStep
	calls []templateDBCall
}
type templateConnector struct{ script *templateDBScript }

func (c templateConnector) Connect(context.Context) (driver.Conn, error) {
	return &templateConn{c.script}, nil
}

func (c templateConnector) Driver() driver.Driver { return templateDriver{} }

type templateDriver struct{}

func (templateDriver) Open(string) (driver.Conn, error) { return nil, errors.New("use connector") }

type templateConn struct{ script *templateDBScript }

func (*templateConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("unexpected prepare")
}

func (*templateConn) Close() error { return nil }

func (*templateConn) Begin() (driver.Tx, error) { return templateTx{}, nil }

type templateTx struct{}

func (templateTx) Commit() error { return nil }

func (templateTx) Rollback() error { return nil }

func (c *templateConn) next(ctx context.Context, q string, args []driver.NamedValue) templateDBStep {
	c.script.t.Helper()
	if _, ok := ctx.Deadline(); !ok {
		c.script.t.Error("query lacks deadline")
	}
	call := templateDBCall{query: q}
	for _, a := range args {
		call.args = append(call.args, a.Value)
	}
	c.script.calls = append(c.script.calls, call)
	if len(c.script.steps) == 0 {
		c.script.t.Fatalf("unexpected query: %s", q)
	}
	step := c.script.steps[0]
	c.script.steps = c.script.steps[1:]
	if !strings.Contains(q, step.contains) {
		c.script.t.Errorf("query %s lacks %s", q, step.contains)
	}
	return step
}

func (c *templateConn) QueryContext(ctx context.Context, q string, args []driver.NamedValue) (driver.Rows, error) {
	s := c.next(ctx, q, args)
	return &templateRows{columns: s.columns, rows: s.rows}, s.err
}

func (c *templateConn) ExecContext(ctx context.Context, q string, args []driver.NamedValue) (driver.Result, error) {
	s := c.next(ctx, q, args)
	return templateResult{s.lastID, s.affected}, s.err
}

type templateResult struct{ id, n int64 }

func (r templateResult) LastInsertId() (int64, error) { return r.id, nil }

func (r templateResult) RowsAffected() (int64, error) { return r.n, nil }

type templateRows struct {
	columns []string
	rows    [][]driver.Value
	pos     int
}

func (r *templateRows) Columns() []string { return r.columns }

func (*templateRows) Close() error { return nil }

func (r *templateRows) Next(dst []driver.Value) error {
	if r.pos == len(r.rows) {
		return io.EOF
	}
	copy(dst, r.rows[r.pos])
	r.pos++
	return nil
}

func templateTestStore(t *testing.T, steps ...templateDBStep) (*GormStore, *templateDBScript) {
	t.Helper()
	script := &templateDBScript{t: t, steps: steps}
	db := sql.OpenDB(templateConnector{script})
	t.Cleanup(func() {
		db.Close()
		if len(script.steps) > 0 {
			t.Errorf("unused steps: %d", len(script.steps))
		}
	})
	orm, err := gorm.Open(mysql.New(mysql.Config{Conn: db, SkipInitializeWithVersion: true}), &gorm.Config{DisableAutomaticPing: true, Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	return NewGormStore(orm), script
}
