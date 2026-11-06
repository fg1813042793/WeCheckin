package infrastructure

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"testing"

	mysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// A database/sql driver exercises actual GORM SQL and transaction boundaries
// without a server, credentials, or an added mock dependency.
type scriptStep struct {
	query    bool
	check    func(string, []driver.NamedValue)
	columns  []string
	rows     [][]driver.Value
	affected int64
	err      error
}
type sqlScript struct {
	query                      func(string, []driver.NamedValue) scriptStep
	t                          *testing.T
	steps                      []scriptStep
	begins, commits, rollbacks int
}

var scripts sync.Map
var driverOnce sync.Once
var nextScript atomic.Uint64

type scriptDriver struct{}

func (scriptDriver) Open(name string) (driver.Conn, error) {
	s, ok := scripts.Load(name)
	if !ok {
		return nil, errors.New("unknown script")
	}
	return &scriptConn{s.(*sqlScript)}, nil
}

type scriptConn struct{ s *sqlScript }

func (c *scriptConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare unsupported")
}
func (c *scriptConn) Close() error              { return nil }
func (c *scriptConn) Begin() (driver.Tx, error) { c.s.begins++; return &scriptTx{c.s}, nil }
func (c *scriptConn) BeginTx(ctx context.Context, _ driver.TxOptions) (driver.Tx, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return c.Begin()
}
func (c *scriptConn) take(ctx context.Context, query bool, sql string, args []driver.NamedValue) scriptStep {
	if err := ctx.Err(); err != nil {
		return scriptStep{err: err}
	}
	if query && c.s.query != nil {
		return c.s.query(sql, args)
	}
	if len(c.s.steps) == 0 {
		c.s.t.Errorf("unexpected SQL: %s", sql)
		return scriptStep{err: errors.New("unexpected SQL")}
	}
	step := c.s.steps[0]
	c.s.steps = c.s.steps[1:]
	if step.query != query {
		c.s.t.Errorf("SQL operation mismatch: %s", sql)
	}
	if step.check != nil {
		step.check(sql, args)
	}
	return step
}
func (c *scriptConn) ExecContext(ctx context.Context, q string, args []driver.NamedValue) (driver.Result, error) {
	s := c.take(ctx, false, q, args)
	return scriptResult(s.affected), s.err
}
func (c *scriptConn) QueryContext(ctx context.Context, q string, args []driver.NamedValue) (driver.Rows, error) {
	s := c.take(ctx, true, q, args)
	return &scriptRows{columns: s.columns, rows: s.rows}, s.err
}

type scriptTx struct{ s *sqlScript }

func (t *scriptTx) Commit() error   { t.s.commits++; return nil }
func (t *scriptTx) Rollback() error { t.s.rollbacks++; return nil }

type scriptResult int64

func (s scriptResult) LastInsertId() (int64, error) { return 12, nil }
func (s scriptResult) RowsAffected() (int64, error) { return int64(s), nil }

type scriptRows struct {
	columns []string
	rows    [][]driver.Value
	i       int
}

func (r *scriptRows) Columns() []string { return r.columns }
func (r *scriptRows) Close() error      { return nil }
func (r *scriptRows) Next(out []driver.Value) error {
	if r.i == len(r.rows) {
		return io.EOF
	}
	copy(out, r.rows[r.i])
	r.i++
	return nil
}
func openScript(t *testing.T, steps ...scriptStep) (*GormStore, *sqlScript) {
	t.Helper()
	driverOnce.Do(func() { sql.Register("publicportal-script", scriptDriver{}) })
	s := &sqlScript{t: t, steps: steps}
	id := fmt.Sprint(nextScript.Add(1))
	scripts.Store(id, s)
	db, err := gorm.Open(mysql.New(mysql.Config{DriverName: "publicportal-script", DSN: id, SkipInitializeWithVersion: true}), &gorm.Config{DisableAutomaticPing: true, SkipDefaultTransaction: true, Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
		scripts.Delete(id)
		if len(s.steps) > 0 {
			t.Errorf("%d SQL statements not executed", len(s.steps))
		}
	})
	return NewGormStore(db), s
}
