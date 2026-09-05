package infrastructure

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"
	mysqldialect "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"wecheckin/backend/internal/modules/userfeedback/application"
	"wecheckin/backend/internal/modules/userfeedback/domain"
)

const scriptedDriverName = "userfeedback-scripted"

var (
	scriptedDriverOnce sync.Once
	scriptedDatabases  sync.Map
	scriptedDatabaseID atomic.Uint64
)

type statementRecord struct {
	SQL  string
	Args []driver.NamedValue
}

type queryResult struct {
	Columns []string
	Rows    [][]driver.Value
	Err     error
}

type execResult struct {
	LastInsertID int64
	RowsAffected int64
	Err          error
}

type databaseScript struct {
	mu sync.Mutex

	queries []statementRecord
	execs   []statementRecord

	query func(string, []driver.NamedValue) queryResult
	exec  func(string, []driver.NamedValue) execResult

	beginErr    error
	commitErr   error
	rollbackErr error
	beginCount  int
	commitCount int
	rollbackCnt int
	beginCtx    context.Context
}

func (script *databaseScript) statements() (queries, execs []statementRecord) {
	script.mu.Lock()
	defer script.mu.Unlock()
	return append([]statementRecord(nil), script.queries...), append([]statementRecord(nil), script.execs...)
}

func (script *databaseScript) transactionCounts() (begin, commit, rollback int) {
	script.mu.Lock()
	defer script.mu.Unlock()
	return script.beginCount, script.commitCount, script.rollbackCnt
}

type scriptedDriver struct{}

func (scriptedDriver) Open(name string) (driver.Conn, error) {
	value, ok := scriptedDatabases.Load(name)
	if !ok {
		return nil, fmt.Errorf("unknown scripted database %q", name)
	}
	return &scriptedConn{script: value.(*databaseScript)}, nil
}

type scriptedConn struct {
	script *databaseScript
}

func (*scriptedConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepared statements are not supported by scripted driver")
}

func (*scriptedConn) Close() error { return nil }

func (conn *scriptedConn) Begin() (driver.Tx, error) {
	return conn.BeginTx(context.Background(), driver.TxOptions{})
}

func (*scriptedConn) Ping(context.Context) error { return nil }

func (conn *scriptedConn) BeginTx(ctx context.Context, _ driver.TxOptions) (driver.Tx, error) {
	conn.script.mu.Lock()
	defer conn.script.mu.Unlock()
	conn.script.beginCount++
	conn.script.beginCtx = ctx
	if conn.script.beginErr != nil {
		return nil, conn.script.beginErr
	}
	return &scriptedTx{script: conn.script}, nil
}

func (conn *scriptedConn) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	conn.script.mu.Lock()
	conn.script.execs = append(conn.script.execs, statementRecord{SQL: query, Args: cloneNamedValues(args)})
	callback := conn.script.exec
	conn.script.mu.Unlock()
	result := execResult{RowsAffected: 1}
	if callback != nil {
		result = callback(query, args)
	}
	if result.Err != nil {
		return nil, result.Err
	}
	return scriptedResult{lastInsertID: result.LastInsertID, rowsAffected: result.RowsAffected}, nil
}

func (conn *scriptedConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	conn.script.mu.Lock()
	conn.script.queries = append(conn.script.queries, statementRecord{SQL: query, Args: cloneNamedValues(args)})
	callback := conn.script.query
	conn.script.mu.Unlock()
	result := queryResult{}
	if callback != nil {
		result = callback(query, args)
	}
	if result.Err != nil {
		return nil, result.Err
	}
	return &scriptedRows{columns: result.Columns, rows: result.Rows}, nil
}

type scriptedTx struct {
	script *databaseScript
}

func (tx *scriptedTx) Commit() error {
	tx.script.mu.Lock()
	defer tx.script.mu.Unlock()
	tx.script.commitCount++
	return tx.script.commitErr
}

func (tx *scriptedTx) Rollback() error {
	tx.script.mu.Lock()
	defer tx.script.mu.Unlock()
	tx.script.rollbackCnt++
	return tx.script.rollbackErr
}

type scriptedResult struct {
	lastInsertID int64
	rowsAffected int64
}

func (result scriptedResult) LastInsertId() (int64, error) { return result.lastInsertID, nil }
func (result scriptedResult) RowsAffected() (int64, error) { return result.rowsAffected, nil }

type scriptedRows struct {
	columns []string
	rows    [][]driver.Value
	index   int
}

func (rows *scriptedRows) Columns() []string { return rows.columns }
func (*scriptedRows) Close() error           { return nil }

func (rows *scriptedRows) Next(dest []driver.Value) error {
	if rows.index >= len(rows.rows) {
		return io.EOF
	}
	copy(dest, rows.rows[rows.index])
	rows.index++
	return nil
}

func cloneNamedValues(values []driver.NamedValue) []driver.NamedValue {
	return append([]driver.NamedValue(nil), values...)
}

func openScriptedGormDB(t *testing.T, script *databaseScript) *gorm.DB {
	t.Helper()
	scriptedDriverOnce.Do(func() { sql.Register(scriptedDriverName, scriptedDriver{}) })
	key := fmt.Sprintf("script-%d", scriptedDatabaseID.Add(1))
	scriptedDatabases.Store(key, script)
	db, err := gorm.Open(mysqldialect.New(mysqldialect.Config{
		DriverName:                scriptedDriverName,
		DSN:                       key,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{
		DisableAutomaticPing:   true,
		SkipDefaultTransaction: true,
		Logger:                 logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open scripted gorm database: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get scripted sql database: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
		scriptedDatabases.Delete(key)
	})
	return db
}

func normalizeSQL(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func namedValues(values []driver.NamedValue) []any {
	result := make([]any, 0, len(values))
	for _, value := range values {
		result = append(result, value.Value)
	}
	return result
}

func TestGormStoreImplementsApplicationContracts(t *testing.T) {
	var _ application.Store = NewGormStore(nil)
	var _ application.TransactionStore = NewGormStore(nil)
}

func TestGormStoreReturnsInitializationErrorsInsteadOfPanicking(t *testing.T) {
	stores := []*GormStore{nil, NewGormStore(nil)}
	for _, store := range stores {
		if _, err := store.GetUserOverview(context.Background(), 1); err == nil || !strings.Contains(err.Error(), "database is not initialized") {
			t.Fatalf("GetUserOverview error = %v", err)
		}
		if err := store.InTransaction(context.Background(), func(application.TransactionStore) error { return nil }); err == nil || !strings.Contains(err.Error(), "database is not initialized") {
			t.Fatalf("InTransaction error = %v", err)
		}
	}
}

func TestInTransactionPassesOneTxBackedStoreAndCommits(t *testing.T) {
	script := &databaseScript{}
	store := NewGormStore(openScriptedGormDB(t, script))
	contextKey := struct{}{}
	ctx := context.WithValue(context.Background(), contextKey, "request")
	var callbackStore *GormStore
	var firstPool, secondPool gorm.ConnPool
	err := store.InTransaction(ctx, func(transactionStore application.TransactionStore) error {
		var ok bool
		callbackStore, ok = transactionStore.(*GormStore)
		if !ok {
			t.Fatalf("transaction store type = %T", transactionStore)
		}
		firstDB, cancel, err := callbackStore.contextDB(ctx)
		if err != nil {
			return err
		}
		cancel()
		secondDB, cancel, err := callbackStore.contextDB(ctx)
		if err != nil {
			return err
		}
		cancel()
		firstPool = firstDB.Statement.ConnPool
		secondPool = secondDB.Statement.ConnPool
		return nil
	})
	if err != nil {
		t.Fatalf("InTransaction: %v", err)
	}
	if callbackStore == nil || callbackStore == store {
		t.Fatalf("callback store = %#v, root = %#v", callbackStore, store)
	}
	if _, ok := firstPool.(*sql.Tx); !ok || firstPool != secondPool {
		t.Fatalf("callback queries did not retain one tx pool: %T %p / %T %p", firstPool, firstPool, secondPool, secondPool)
	}
	begin, commit, rollback := script.transactionCounts()
	if begin != 1 || commit != 1 || rollback != 0 {
		t.Fatalf("transaction counts = begin %d commit %d rollback %d", begin, commit, rollback)
	}
	script.mu.Lock()
	beginCtx := script.beginCtx
	script.mu.Unlock()
	if beginCtx == nil || beginCtx.Value(contextKey) != "request" {
		t.Fatalf("begin context did not retain request value")
	}
	if _, ok := beginCtx.Deadline(); !ok {
		t.Fatalf("begin context must be created by database.QueryContext")
	}
}

func TestInTransactionWrapsOnlyCommitFailuresAsOutcomeUnknown(t *testing.T) {
	commitCause := errors.New("commit connection lost")
	script := &databaseScript{commitErr: commitCause}
	store := NewGormStore(openScriptedGormDB(t, script))
	err := store.InTransaction(context.Background(), func(application.TransactionStore) error { return nil })
	if !errors.Is(err, application.ErrTransactionOutcomeUnknown) || !errors.Is(err, commitCause) {
		t.Fatalf("commit error = %v", err)
	}
	if begin, commit, rollback := script.transactionCounts(); begin != 1 || commit != 1 || rollback != 0 {
		t.Fatalf("transaction counts = begin %d commit %d rollback %d", begin, commit, rollback)
	}
}

func TestInTransactionRollsBackCallbackFailureWithoutUnknownMarker(t *testing.T) {
	callbackCause := errors.New("callback failed")
	rollbackCause := errors.New("rollback failed")
	script := &databaseScript{rollbackErr: rollbackCause}
	store := NewGormStore(openScriptedGormDB(t, script))
	err := store.InTransaction(context.Background(), func(application.TransactionStore) error { return callbackCause })
	if !errors.Is(err, callbackCause) || !errors.Is(err, rollbackCause) {
		t.Fatalf("rollback error = %v", err)
	}
	if errors.Is(err, application.ErrTransactionOutcomeUnknown) {
		t.Fatalf("confirmed rollback must not be marked outcome unknown: %v", err)
	}
	if begin, commit, rollback := script.transactionCounts(); begin != 1 || commit != 0 || rollback != 1 {
		t.Fatalf("transaction counts = begin %d commit %d rollback %d", begin, commit, rollback)
	}
}

func TestNextFeedbackNumberCreatesLocksAndAdvancesPastFourDigits(t *testing.T) {
	script := &databaseScript{
		query: func(query string, _ []driver.NamedValue) queryResult {
			if !strings.Contains(normalizeSQL(query), "FOR UPDATE") {
				return queryResult{Err: fmt.Errorf("sequence query lacks FOR UPDATE: %s", query)}
			}
			return queryResult{
				Columns: []string{"sequence_date", "current_value", "updated_at"},
				Rows:    [][]driver.Value{{time.Date(2026, 9, 5, 0, 0, 0, 0, time.UTC), int64(9999), int64(1)}},
			}
		},
		exec: func(string, []driver.NamedValue) execResult { return execResult{RowsAffected: 1} },
	}
	store := NewGormStore(openScriptedGormDB(t, script))
	value, err := store.NextFeedbackNumber(context.Background(), "2026-09-05")
	if err != nil || value != 10000 {
		t.Fatalf("NextFeedbackNumber = %d, %v", value, err)
	}
	queries, execs := script.statements()
	if len(queries) != 1 || len(execs) != 2 {
		t.Fatalf("sequence statements = %d queries, %d execs", len(queries), len(execs))
	}
	if !strings.Contains(normalizeSQL(execs[0].SQL), "INSERT INTO `user_feedback_daily_sequences`") ||
		strings.Contains(normalizeSQL(execs[0].SQL), "ON DUPLICATE KEY") {
		t.Fatalf("missing-row insert SQL = %s", execs[0].SQL)
	}
	if !strings.Contains(normalizeSQL(execs[1].SQL), "UPDATE `user_feedback_daily_sequences`") {
		t.Fatalf("increment SQL = %s", execs[1].SQL)
	}
}

func TestNextFeedbackNumberIgnoresOnlyPrimarySeedConflict(t *testing.T) {
	tests := []struct {
		name         string
		seedErr      error
		wantOriginal bool
	}{
		{
			name:    "bare primary",
			seedErr: mysqlDuplicate(`Duplicate entry '2026-09-05' for key 'PRIMARY'`),
		},
		{
			name:    "qualified quoted primary",
			seedErr: mysqlDuplicate("Duplicate entry '2026-09-05' for key `wecheckin`.`user_feedback_daily_sequences`.`PRIMARY`"),
		},
		{
			name:         "other unique index",
			seedErr:      mysqlDuplicate(`Duplicate entry 'x' for key 'user_feedback_daily_sequences.uk_unexpected'`),
			wantOriginal: true,
		},
		{
			name:         "translated duplicate without index",
			seedErr:      gorm.ErrDuplicatedKey,
			wantOriginal: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			execCount := 0
			script := &databaseScript{
				exec: func(string, []driver.NamedValue) execResult {
					execCount++
					if execCount == 1 {
						return execResult{Err: test.seedErr}
					}
					return execResult{RowsAffected: 1}
				},
				query: func(string, []driver.NamedValue) queryResult {
					return queryResult{
						Columns: []string{"sequence_date", "current_value", "updated_at"},
						Rows:    [][]driver.Value{{time.Date(2026, 9, 5, 0, 0, 0, 0, time.UTC), int64(8), int64(1)}},
					}
				},
			}
			value, err := NewGormStore(openScriptedGormDB(t, script)).NextFeedbackNumber(context.Background(), "2026-09-05")
			if test.wantOriginal {
				if err != test.seedErr || value != 0 {
					t.Fatalf("NextFeedbackNumber = %d, %v, want original %v", value, err, test.seedErr)
				}
				return
			}
			if err != nil || value != 9 {
				t.Fatalf("NextFeedbackNumber = %d, %v", value, err)
			}
			queries, execs := script.statements()
			if len(queries) != 1 || len(execs) != 2 || !strings.Contains(normalizeSQL(queries[0].SQL), "FOR UPDATE") {
				t.Fatalf("sequence coordination SQL = %#v %#v", queries, execs)
			}
			if strings.Contains(normalizeSQL(execs[0].SQL), "ON DUPLICATE KEY") {
				t.Fatalf("sequence seed must be a plain insert: %s", execs[0].SQL)
			}
		})
	}
}

func TestNextFeedbackNumberRejectsInvalidDateAndOverflow(t *testing.T) {
	script := &databaseScript{}
	store := NewGormStore(openScriptedGormDB(t, script))
	for _, value := range []string{"", "2026-9-5", "2026-02-30", "2026-09-05x"} {
		if _, err := store.NextFeedbackNumber(context.Background(), value); err == nil {
			t.Fatalf("date %q unexpectedly accepted", value)
		}
	}
	if _, err := incrementSequence(math.MaxUint64); err == nil {
		t.Fatal("uint64 overflow unexpectedly accepted")
	}
	if value, err := incrementSequence(9999); err != nil || value != 10000 {
		t.Fatalf("incrementSequence(9999) = %d, %v", value, err)
	}
	queries, execs := script.statements()
	if len(queries)+len(execs) != 0 {
		t.Fatalf("invalid dates executed SQL: %#v %#v", queries, execs)
	}
}

func TestCountCreatedByUserOnDateUsesHalfOpenInterval(t *testing.T) {
	script := &databaseScript{query: func(_ string, _ []driver.NamedValue) queryResult {
		return queryResult{Columns: []string{"count"}, Rows: [][]driver.Value{{int64(4)}}}
	}}
	store := NewGormStore(openScriptedGormDB(t, script))
	count, err := store.CountCreatedByUserOnDate(context.Background(), 7, 1000, 2000)
	if err != nil || count != 4 {
		t.Fatalf("CountCreatedByUserOnDate = %d, %v", count, err)
	}
	queries, _ := script.statements()
	if len(queries) != 1 {
		t.Fatalf("query count = %d", len(queries))
	}
	sqlText := normalizeSQL(queries[0].SQL)
	for _, fragment := range []string{"submitter_id = ?", "created_at >= ?", "created_at < ?"} {
		if !strings.Contains(sqlText, fragment) {
			t.Fatalf("count SQL missing %q: %s", fragment, sqlText)
		}
	}
	if got, want := namedValues(queries[0].Args), []any{int64(7), int64(1000), int64(2000)}; !reflect.DeepEqual(got, want) {
		t.Fatalf("count args = %#v, want %#v", got, want)
	}
}

func TestLockFeedbackUsesForUpdateAndMapsNotFound(t *testing.T) {
	script := &databaseScript{query: func(_ string, _ []driver.NamedValue) queryResult {
		return queryResult{Columns: []string{"id"}, Rows: [][]driver.Value{}}
	}}
	store := NewGormStore(openScriptedGormDB(t, script))
	if _, err := store.LockFeedback(context.Background(), 88); !errors.Is(err, application.ErrFeedbackNotFound) {
		t.Fatalf("LockFeedback error = %v", err)
	}
	queries, _ := script.statements()
	if len(queries) != 1 || !strings.Contains(normalizeSQL(queries[0].SQL), "FOR UPDATE") {
		t.Fatalf("lock queries = %#v", queries)
	}
}

func mysqlDuplicate(message string) *mysqlDriver.MySQLError {
	return &mysqlDriver.MySQLError{Number: 1062, Message: message}
}

func feedbackRecord() application.FeedbackRecord {
	return application.FeedbackRecord{
		FeedbackNo: "FB-20260905-0001", SubmitterID: 7, CreateRequestID: "create-1",
		Status: domain.StatusPending, Version: 1, LastActivityAt: 10, CreatedAt: 10, UpdatedAt: 10,
	}
}

func messageRecord() application.MessageRecord {
	return application.MessageRecord{
		FeedbackID: 9, MessageType: domain.MessageTypeSupplement, AuthorType: domain.AuthorTypeUser,
		AuthorID: 7, Content: "more", RequestID: "message-1", CreatedAt: 11,
	}
}

func notificationRecord() application.NotificationOutboxRecord {
	return application.NotificationOutboxRecord{
		IdempotencyKey: "feedback-status:9:20", Channel: "internal", NotificationType: "feedback_status",
		SourceType: "user_feedback", SourceID: "9", RecipientUserID: 7, Title: "title", Content: "content", CreatedAt: 12345,
	}
}

func TestCreateFeedbackClassifiesOnlyExactRequestIndex(t *testing.T) {
	tests := []struct {
		name      string
		cause     error
		duplicate bool
	}{
		{name: "bare index", cause: mysqlDuplicate(`Duplicate entry '7-create-1' for key 'uk_user_feedbacks_submitter_request'`), duplicate: true},
		{name: "single quoted table prefix", cause: mysqlDuplicate(`Duplicate entry '7-create-1' for key 'user_feedbacks.uk_user_feedbacks_submitter_request'`), duplicate: true},
		{name: "backtick schema and table prefix", cause: mysqlDuplicate("Duplicate entry '7-create-1' for key `wecheckin`.`user_feedbacks`.`uk_user_feedbacks_submitter_request`"), duplicate: true},
		{name: "feedback number index", cause: mysqlDuplicate(`Duplicate entry 'FB-20260905-0001' for key 'uk_user_feedbacks_feedback_no'`)},
		{name: "unknown unique index", cause: mysqlDuplicate(`Duplicate entry 'x' for key 'user_feedbacks.uk_unknown'`)},
		{name: "does not contain for key suffix", cause: mysqlDuplicate(`Duplicate entry 'x'`)},
		{name: "translated duplicate without index", cause: gorm.ErrDuplicatedKey},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			script := &databaseScript{exec: func(string, []driver.NamedValue) execResult { return execResult{Err: test.cause} }}
			_, err := NewGormStore(openScriptedGormDB(t, script)).CreateFeedback(context.Background(), feedbackRecord())
			if test.duplicate {
				if !errors.Is(err, application.ErrDuplicateRequest) {
					t.Fatalf("CreateFeedback error = %v", err)
				}
			} else if err != test.cause {
				t.Fatalf("CreateFeedback error = %v, want original %v", err, test.cause)
			}
			assertOnlyPlainInserts(t, script)
		})
	}
}

func TestAppendMessageClassifiesOnlyExactRequestIndex(t *testing.T) {
	tests := []struct {
		name      string
		cause     error
		duplicate bool
	}{
		{name: "request index", cause: mysqlDuplicate(`Duplicate entry '9-user-7-message-1' for key 'user_feedback_messages.uk_user_feedback_messages_request'`), duplicate: true},
		{name: "unknown index", cause: mysqlDuplicate(`Duplicate entry 'x' for key 'uk_unknown'`)},
		{name: "translated duplicate", cause: gorm.ErrDuplicatedKey},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			script := &databaseScript{exec: func(string, []driver.NamedValue) execResult { return execResult{Err: test.cause} }}
			_, err := NewGormStore(openScriptedGormDB(t, script)).AppendMessage(context.Background(), messageRecord())
			if test.duplicate {
				if !errors.Is(err, application.ErrDuplicateRequest) {
					t.Fatalf("AppendMessage error = %v", err)
				}
			} else if err != test.cause {
				t.Fatalf("AppendMessage error = %v, want original %v", err, test.cause)
			}
			assertOnlyPlainInserts(t, script)
		})
	}
}

func TestEnqueueNotificationClassifiesOnlyExactIdempotencyIndex(t *testing.T) {
	tests := []struct {
		name       string
		cause      error
		idempotent bool
	}{
		{name: "idempotency index", cause: mysqlDuplicate(`Duplicate entry 'feedback-status:9:20' for key 'notification_outbox.uk_notification_outbox_idempotency'`), idempotent: true},
		{name: "unknown index", cause: mysqlDuplicate(`Duplicate entry 'x' for key 'uk_unknown'`)},
		{name: "translated duplicate", cause: gorm.ErrDuplicatedKey},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			script := &databaseScript{exec: func(string, []driver.NamedValue) execResult { return execResult{Err: test.cause} }}
			err := NewGormStore(openScriptedGormDB(t, script)).EnqueueNotification(context.Background(), notificationRecord())
			if test.idempotent {
				if err != nil {
					t.Fatalf("EnqueueNotification error = %v", err)
				}
			} else if err != test.cause {
				t.Fatalf("EnqueueNotification error = %v, want original %v", err, test.cause)
			}
			assertOnlyPlainInserts(t, script)
		})
	}
}

func TestPlainInsertZeroRowsReturnsInternalError(t *testing.T) {
	t.Run("feedback", func(t *testing.T) {
		store := NewGormStore(openScriptedGormDB(t, &databaseScript{exec: func(string, []driver.NamedValue) execResult { return execResult{} }}))
		_, err := store.CreateFeedback(context.Background(), feedbackRecord())
		if err == nil || errors.Is(err, application.ErrDuplicateRequest) {
			t.Fatalf("CreateFeedback zero-row error = %v", err)
		}
	})
	t.Run("message", func(t *testing.T) {
		store := NewGormStore(openScriptedGormDB(t, &databaseScript{exec: func(string, []driver.NamedValue) execResult { return execResult{} }}))
		_, err := store.AppendMessage(context.Background(), messageRecord())
		if err == nil || errors.Is(err, application.ErrDuplicateRequest) {
			t.Fatalf("AppendMessage zero-row error = %v", err)
		}
	})
	t.Run("notification", func(t *testing.T) {
		store := NewGormStore(openScriptedGormDB(t, &databaseScript{exec: func(string, []driver.NamedValue) execResult { return execResult{} }}))
		if err := store.EnqueueNotification(context.Background(), notificationRecord()); err == nil {
			t.Fatal("EnqueueNotification accepted zero-row insert")
		}
	})
	t.Run("attachments", func(t *testing.T) {
		store := NewGormStore(openScriptedGormDB(t, &databaseScript{exec: func(string, []driver.NamedValue) execResult { return execResult{} }}))
		err := store.AppendAttachments(context.Background(), []application.AttachmentRecord{{
			FeedbackID: 9, MessageID: 10, ObjectKey: "feedback/a.jpg", OriginalName: "a.jpg",
		}})
		if err == nil {
			t.Fatal("AppendAttachments accepted zero-row insert")
		}
	})
	t.Run("sequence seed", func(t *testing.T) {
		script := &databaseScript{
			exec: func(string, []driver.NamedValue) execResult { return execResult{} },
			query: func(string, []driver.NamedValue) queryResult {
				return queryResult{Columns: []string{"sequence_date", "current_value", "updated_at"}, Rows: [][]driver.Value{{time.Now(), int64(1), int64(1)}}}
			},
		}
		if _, err := NewGormStore(openScriptedGormDB(t, script)).NextFeedbackNumber(context.Background(), "2026-09-05"); err == nil {
			t.Fatal("NextFeedbackNumber accepted zero-row seed insert")
		}
		queries, _ := script.statements()
		if len(queries) != 0 {
			t.Fatalf("zero-row seed continued to locking query: %#v", queries)
		}
	})
}

func assertOnlyPlainInserts(t *testing.T, script *databaseScript) {
	t.Helper()
	_, execs := script.statements()
	if len(execs) != 1 {
		t.Fatalf("insert exec count = %d", len(execs))
	}
	if strings.Contains(normalizeSQL(execs[0].SQL), "ON DUPLICATE KEY") {
		t.Fatalf("insert must not use ON DUPLICATE KEY: %s", execs[0].SQL)
	}
}

func TestAppendAttachmentsSkipsEmptySlice(t *testing.T) {
	script := &databaseScript{}
	store := NewGormStore(openScriptedGormDB(t, script))
	if err := store.AppendAttachments(context.Background(), []application.AttachmentRecord{}); err != nil {
		t.Fatalf("AppendAttachments(empty): %v", err)
	}
	queries, execs := script.statements()
	if len(queries)+len(execs) != 0 {
		t.Fatalf("empty attachments executed SQL: %#v %#v", queries, execs)
	}
}

func TestUpdateSnapshotWritesAllFieldsAndGuardsOldVersion(t *testing.T) {
	script := &databaseScript{exec: func(string, []driver.NamedValue) execResult {
		return execResult{RowsAffected: 0}
	}}
	store := NewGormStore(openScriptedGormDB(t, script))
	handlerID := uint(3)
	resolvedAt, closedAt := int64(120), int64(130)
	err := store.UpdateSnapshot(context.Background(), application.FeedbackSnapshot{
		ID: 9, FeedbackNo: "FB-20260905-0009", SubmitterID: 7, Status: domain.StatusClosed,
		HandlerID: &handlerID, Version: 4, LastActivityAt: 140, ResolvedAt: &resolvedAt,
		ClosedAt: &closedAt, CreatedAt: 100, UpdatedAt: 140,
	}, 3)
	if !errors.Is(err, application.ErrVersionConflict) {
		t.Fatalf("UpdateSnapshot error = %v", err)
	}
	_, execs := script.statements()
	if len(execs) != 1 {
		t.Fatalf("update exec count = %d", len(execs))
	}
	sqlText := normalizeSQL(execs[0].SQL)
	for _, fragment := range []string{
		"`feedback_no`", "`submitter_id`", "`feedback_status`", "`handler_id`", "`version`",
		"`last_activity_at`", "`resolved_at`", "`closed_at`", "`created_at`", "`updated_at`",
		"WHERE id = ? AND version = ?",
	} {
		if !strings.Contains(sqlText, fragment) {
			t.Fatalf("snapshot SQL missing %q: %s", fragment, sqlText)
		}
	}
}

func TestEnqueueNotificationUsesStableRecipientAndPayloadJSON(t *testing.T) {
	row, err := notificationOutboxModel(application.NotificationOutboxRecord{
		IdempotencyKey: "feedback-status:9:20", Channel: "internal", NotificationType: "feedback_status",
		SourceType: "user_feedback", SourceID: "9", RecipientUserID: 7,
		Title: "反馈 FB-9 已解决", Content: "已修复", CreatedAt: 12345,
	})
	if err != nil {
		t.Fatalf("notificationOutboxModel: %v", err)
	}
	if got, want := row.RecipientJSON, `{"notifyAdmin":false,"userIds":[7]}`; got != want {
		t.Fatalf("recipient JSON = %s, want %s", got, want)
	}
	if got, want := row.PayloadJSON, `{"title":"反馈 FB-9 已解决","content":"已修复","sourceType":"user_feedback","sourceId":"9","notificationType":"feedback_status"}`; got != want {
		t.Fatalf("payload JSON = %s, want %s", got, want)
	}
	if row.Status != "pending" || row.Attempts != 0 || row.NextRetryAt != 0 || row.AddTime != 12345 || row.EditTime != 12345 {
		t.Fatalf("outbox state = %#v", row)
	}

	script := &databaseScript{exec: func(string, []driver.NamedValue) execResult {
		return execResult{Err: mysqlDuplicate(`Duplicate entry 'feedback-status:9:20' for key 'uk_notification_outbox_idempotency'`)}
	}}
	store := NewGormStore(openScriptedGormDB(t, script))
	if err := store.EnqueueNotification(context.Background(), notificationRecord()); err != nil {
		t.Fatalf("duplicate EnqueueNotification must succeed: %v", err)
	}
	_, execs := script.statements()
	if len(execs) != 1 || strings.Contains(normalizeSQL(execs[0].SQL), "ON DUPLICATE KEY") {
		t.Fatalf("notification SQL = %#v", execs)
	}
}
