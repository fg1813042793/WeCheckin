package adminlog

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"wecheckin/backend/internal/support/adminaccess"
	permissionsupport "wecheckin/backend/internal/support/permission"
	"wecheckin/backend/pkg/database"
)

type auditDBFixture struct {
	permissions                                                                []string
	denied                                                                     []string
	missingActor, actorFailure, super, scopeFailure, countFailure, emptyDetail bool
	logQueries                                                                 []string
	logArgs                                                                    [][]driver.NamedValue
	writes                                                                     []string
	writeArgs                                                                  [][]driver.NamedValue
	writeHasDeadline                                                           bool
	scopeAll                                                                   bool
	emptyPage                                                                  bool
	deleteIDs                                                                  []uint
	deleteFailure                                                              bool
	commits, rollbacks                                                         int
}

type auditConnector struct{ fixture *auditDBFixture }

func (c auditConnector) Connect(context.Context) (driver.Conn, error) {
	return &auditConn{fixture: c.fixture}, nil
}
func (c auditConnector) Driver() driver.Driver { return auditDriver{} }

type auditDriver struct{}

func (auditDriver) Open(string) (driver.Conn, error) { return nil, errors.New("use connector") }

type auditConn struct{ fixture *auditDBFixture }

func (*auditConn) Prepare(string) (driver.Stmt, error) { return nil, errors.New("unexpected prepare") }
func (*auditConn) Close() error                        { return nil }
func (c *auditConn) Begin() (driver.Tx, error)         { return auditTx{fixture: c.fixture}, nil }

type auditTx struct{ fixture *auditDBFixture }

func (tx auditTx) Commit() error   { tx.fixture.commits++; return nil }
func (tx auditTx) Rollback() error { tx.fixture.rollbacks++; return nil }

type auditRows struct {
	columns []string
	values  [][]driver.Value
	offset  int
}

func (r *auditRows) Columns() []string { return r.columns }
func (*auditRows) Close() error        { return nil }
func (r *auditRows) Next(dst []driver.Value) error {
	if r.offset == len(r.values) {
		return io.EOF
	}
	copy(dst, r.values[r.offset])
	r.offset++
	return nil
}
func testRows(columns []string, values ...[]driver.Value) driver.Rows {
	return &auditRows{columns: columns, values: values}
}
func testHasArg(args []driver.NamedValue, wanted string) bool {
	for _, arg := range args {
		if fmt.Sprint(arg.Value) == wanted {
			return true
		}
	}
	return false
}

func (c *auditConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	f := c.fixture
	switch {
	case strings.HasPrefix(query, "SELECT DATABASE()"):
		return testRows([]string{"DATABASE()"}, []driver.Value{"audit_test"}), nil
	case strings.Contains(strings.ToLower(query), "information_schema"):
		return testRows([]string{"count(*)"}, []driver.Value{int64(1)}), nil
	case strings.Contains(query, "FROM `users`"):
		if f.actorFailure {
			return nil, errors.New("actor lookup failed")
		}
		cols := []string{"id", "user_status", "user_role_id"}
		if f.missingActor {
			return testRows(cols), nil
		}
		return testRows(cols, []driver.Value{int64(7), int64(1), int64(9)}), nil
	case strings.Contains(query, "FROM user_roles AS ur"):
		return testRows([]string{"user_role_role_id"}, []driver.Value{int64(9)}), nil
	case strings.Contains(query, "FROM `roles`"):
		name := "审计测试"
		if f.super {
			name = adminaccess.ReservedSuperAdminRoleName
		}
		return testRows([]string{"id", "role_name", "role_status", "role_data_scope"}, []driver.Value{int64(9), name, int64(1), int64(3)}), nil
	case strings.Contains(query, "FROM `permission_grants`"):
		cols := []string{"grant_subject_type", "grant_subject_id", "grant_permission_key", "grant_effect"}
		if testHasArg(args, "data:extra") && !testHasArg(args, "data:self") {
			return testRows(cols), nil
		}
		if testHasArg(args, "data:self") {
			if f.scopeFailure {
				return nil, errors.New("scope lookup failed")
			}
			scopeKey := "data:self"
			if f.scopeAll {
				scopeKey = "data:all"
			}
			return testRows(cols, []driver.Value{"role", int64(9), scopeKey, "allow"}), nil
		}
		var rows [][]driver.Value
		subject, id := "role", int64(9)
		keys := f.permissions
		effect := "allow"
		if testHasArg(args, "user") {
			subject, id, keys, effect = "user", 7, f.denied, "deny"
		}
		if testHasArg(args, "admin:login") {
			if subject == "role" {
				keys = []string{"admin:login"}
			} else {
				keys = nil
			}
		}
		for _, key := range keys {
			rows = append(rows, []driver.Value{subject, id, key, effect})
		}
		return testRows(cols, rows...), nil
	case strings.Contains(query, "FROM `logs`"):
		f.logQueries = append(f.logQueries, query)
		f.logArgs = append(f.logArgs, append([]driver.NamedValue(nil), args...))
		if strings.Contains(query, "FOR UPDATE") {
			var rows [][]driver.Value
			for _, id := range f.deleteIDs {
				rows = append(rows, []driver.Value{int64(id)})
			}
			return testRows([]string{"id"}, rows...), nil
		}
		if strings.Contains(query, "count(*)") {
			if f.countFailure {
				return nil, errors.New("count failed")
			}
			return testRows([]string{"count(*)"}, []driver.Value{int64(1205)}), nil
		}
		cols := []string{"id", "log_type", "create_by", "log_admin_name", "log_content", "log_requested_count", "log_affected_count"}
		if f.emptyDetail || f.emptyPage {
			return testRows(cols), nil
		}
		typ, actor := int64(2), int64(7)
		if testHasArg(args, CategoryLogin) {
			typ, actor = 1, 0
		}
		values := []driver.Value{int64(41), typ, actor, "snapshot", "audit entry", int64(3), int64(1)}
		selection := strings.SplitN(query, " FROM ", 2)[0]
		if strings.Contains(selection, "*") || strings.Contains(selection, "`log_target_ids`") {
			cols = append(cols, "log_target_ids")
			values = append(values, []byte(`["41","42","43"]`))
		}
		if strings.Contains(selection, "*") || strings.Contains(selection, "`log_changes`") {
			cols = append(cols, "log_changes")
			values = append(values, []byte(`[{"field":"status","before":0,"after":1}]`))
		}
		if strings.Contains(selection, "*") || strings.Contains(selection, "`log_user_agent`") {
			cols = append(cols, "log_user_agent")
			values = append(values, "detail-only user agent")
		}
		return testRows(cols, values), nil
	default:
		return nil, fmt.Errorf("unexpected SQL: %s", query)
	}
}

func (c *auditConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	c.fixture.writes = append(c.fixture.writes, query)
	_, c.fixture.writeHasDeadline = ctx.Deadline()
	c.fixture.writeArgs = append(c.fixture.writeArgs, append([]driver.NamedValue(nil), args...))
	if strings.HasPrefix(query, "DELETE FROM `logs`") {
		if c.fixture.deleteFailure {
			return nil, errors.New("delete failed")
		}
		return driver.RowsAffected(len(c.fixture.deleteIDs)), nil
	}
	if !strings.HasPrefix(query, "INSERT INTO `logs`") {
		return nil, fmt.Errorf("unexpected write: %s", query)
	}
	return auditResult{}, nil
}

type auditResult struct{}

func (auditResult) LastInsertId() (int64, error) { return 42, nil }
func (auditResult) RowsAffected() (int64, error) { return 1, nil }

func openAuditTestDB(t *testing.T, f *auditDBFixture) *gorm.DB {
	t.Helper()
	conn := sql.OpenDB(auditConnector{fixture: f})
	t.Cleanup(func() { _ = conn.Close() })
	db, err := gorm.Open(mysql.New(mysql.Config{Conn: conn, SkipInitializeWithVersion: true}), &gorm.Config{DisableAutomaticPing: true, Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	previous := database.DB
	database.DB = db
	reset := func() {
		permissionsupport.ResetPermissionTablesReadyCache()
		permissionsupport.InvalidateRuntimePermissionCaches()
		adminaccess.InvalidateAdminAccessCache()
	}
	reset()
	t.Cleanup(func() { database.DB = previous; reset() })
	return db
}

// OpenHandlerTestDB exposes the existing SQL fixture to external handler tests only.
func OpenHandlerTestDB(t *testing.T, failure string, empty bool) {
	openAuditTestDB(t, &auditDBFixture{super: true, countFailure: failure == "count", actorFailure: failure == "actor", emptyPage: empty})
}

func OpenDeleteHandlerTestDB(t *testing.T, ids []uint) {
	openAuditTestDB(t, &auditDBFixture{super: true, deleteIDs: ids})
}
