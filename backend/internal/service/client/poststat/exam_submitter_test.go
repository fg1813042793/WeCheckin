package poststat

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
	notificationoutboxapp "wecheckin/backend/internal/modules/notificationoutbox/application"
	"wecheckin/backend/pkg/database"
)

const submitterExamSchema = `{"version":"2.0","questions":[{"id":"phone","type":"phone","title":"手机"},{"id":"name","type":"name","title":"姓名"},{"id":"email","type":"email","title":"邮箱"},{"id":"student","type":"studentId","title":"学号"},{"id":"employee","type":"employeeId","title":"工号"},{"id":"essay","type":"textarea","title":"姓名"}]}`

func TestExamPostStatNotificationUsesPersonalAnswers(t *testing.T) {
	for _, tc := range []struct {
		name, answers, account, want string
		userID                       uint
	}{
		{name: "anonymous name precedes phone", answers: `{"name":"  测试考生  ","phone":"13800000000"}`, want: "测试考生"},
		{name: "answered name precedes account", answers: `{"name":"测试考生"}`, account: "登录账号", userID: 7, want: "测试考生"},
		{name: "blank name falls back to phone", answers: `{"name":"  ","phone":"13800000000"}`, want: "13800000000"},
		{name: "email", answers: `{"email":"exam@example.com"}`, want: "exam@example.com"},
		{name: "student ID", answers: `{"student":"S001"}`, want: "S001"},
		{name: "employee ID", answers: `{"employee":"E001"}`, want: "E001"},
		{name: "account fallback", answers: `{}`, account: "登录账号", userID: 7, want: "登录账号"},
		{name: "missing account fallback", answers: `{}`, userID: 7, want: "用户7"},
		{name: "anonymous fallback", answers: `{}`, want: "匿名用户"},
		{name: "ordinary answer is not identity", answers: `{"essay":"普通答案"}`, want: "匿名用户"},
		{name: "malformed answers", answers: `{invalid`, want: "匿名用户"},
		{name: "structured name is not identity", answers: `{"name":{"value":"名字"},"phone":[]}`, want: "匿名用户"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := &examSubmitterFixture{schema: submitterExamSchema, answers: tc.answers, account: tc.account}
			dispatcher := openExamSubmitterFixture(t, fixture)
			if err := ProcessExamSubmissionContext(context.Background(), 5, 22, tc.userID, 0, 0, 1); err != nil {
				t.Fatal(err)
			}
			if len(dispatcher.requests) != 2 {
				t.Fatalf("expected webhook and internal notifications, got %d", len(dispatcher.requests))
			}
			for _, request := range dispatcher.requests {
				payload := request.Payload.(notificationoutboxapp.MessagePayload)
				if payload.Content != "提交人："+tc.want+"；待人工批改" {
					t.Fatalf("%s notification = %q, want submitter %q", request.Channel, payload.Content, tc.want)
				}
			}
		})
	}
}

func TestExamPostStatDoesNotEnqueueWhenRecordCannotBeRead(t *testing.T) {
	fixture := &examSubmitterFixture{schema: submitterExamSchema, recordError: errors.New("record unavailable")}
	dispatcher := openExamSubmitterFixture(t, fixture)
	err := ProcessExamSubmissionContext(context.Background(), 5, 22, 0, 0, 0, 0)
	if !errors.Is(err, fixture.recordError) || len(dispatcher.requests) != 0 {
		t.Fatalf("record read failure should prevent incorrect notification: err=%v requests=%d", err, len(dispatcher.requests))
	}
}

func TestSurveyAndExamSharePersonalSubmitterPriority(t *testing.T) {
	ctx := context.Background()
	if got := resolveSubmitterContext(ctx, submitterExamSchema, `{"name":" 测试考生 ","phone":"13800000000"}`, "昵称", 0); got != "测试考生" {
		t.Fatalf("survey submitter = %q", got)
	}
	if got := resolveSubmitterContext(ctx, submitterExamSchema, `{}`, "昵称", 0); got != "昵称" {
		t.Fatalf("survey nickname fallback = %q", got)
	}
	for _, schema := range []string{"", "{invalid", `{"version":"2.0","questions":[]}`} {
		if got := resolveExamSubmitterContext(ctx, 0, schema, `{"name":"不可猜测的答案"}`); got != "匿名用户" {
			t.Fatalf("missing personal schema must remain anonymous, got %q", got)
		}
	}
}

// A local SQL driver exercises the actual service and GORM projections without
// production data, a running database, or sending any notification.
type examSubmitterFixture struct {
	schema, answers, account string
	recordError              error
}

func openExamSubmitterFixture(t *testing.T, fixture *examSubmitterFixture) *notificationDispatcherStub {
	t.Helper()
	connection := sql.OpenDB(examSubmitterConnector{fixture})
	db, err := gorm.Open(mysql.New(mysql.Config{Conn: connection, SkipInitializeWithVersion: true}), &gorm.Config{DisableAutomaticPing: true, Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	previousDB, previousDispatcher := database.DB, configuredNotificationDispatcher()
	database.DB = db
	dispatcher := &notificationDispatcherStub{}
	ConfigureNotificationDispatcher(dispatcher)
	t.Cleanup(func() {
		database.DB = previousDB
		ConfigureNotificationDispatcher(previousDispatcher)
		_ = connection.Close()
	})
	return dispatcher
}

type examSubmitterConnector struct{ fixture *examSubmitterFixture }

func (c examSubmitterConnector) Connect(context.Context) (driver.Conn, error) {
	return &examSubmitterConn{c.fixture}, nil
}
func (examSubmitterConnector) Driver() driver.Driver { return examSubmitterDriver{} }

type examSubmitterDriver struct{}

func (examSubmitterDriver) Open(string) (driver.Conn, error) { return nil, errors.New("use connector") }

type examSubmitterConn struct{ fixture *examSubmitterFixture }

func (*examSubmitterConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("unexpected prepare")
}
func (*examSubmitterConn) Begin() (driver.Tx, error) { return nil, errors.New("unexpected write") }
func (*examSubmitterConn) Close() error              { return nil }
func (c *examSubmitterConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if strings.Contains(query, "FROM `exam`") {
		columns := []string{"exam_id", "exam_title", "exam_settings"}
		values := []driver.Value{int64(5), "用户信息考试", `{"logicRules":[{"action":"postStat","notifyChannel":"both","notifyAdmin":true,"webhookType":"dingtalk","webhookUrl":"https://example.com/hook","messageTemplate":"提交人：{submitter}；{result}"}]}`}
		if strings.Contains(query, "`exam_schema`") {
			columns, values = append(columns, "exam_schema"), append(values, c.fixture.schema)
		}
		return &examSubmitterRows{columns: columns, values: values}, nil
	}
	if strings.Contains(query, "FROM `exam_record`") {
		if strings.Contains(query, "count(*)") {
			return &examSubmitterRows{columns: []string{"count(*)"}, values: []driver.Value{int64(1)}}, nil
		}
		if !strings.Contains(query, "`exam_r_id` = ? AND `exam_r_exam_id` = ?") || len(args) < 2 || args[0].Value != int64(22) || args[1].Value != int64(5) {
			return nil, fmt.Errorf("answer read must be scoped to current record and exam: %s", query)
		}
		if c.fixture.recordError != nil {
			return nil, c.fixture.recordError
		}
		return &examSubmitterRows{columns: []string{"exam_r_answers"}, values: []driver.Value{c.fixture.answers}}, nil
	}
	if strings.Contains(query, "FROM `users`") {
		if !strings.Contains(query, "`user_name`") {
			return nil, errors.New("user name must be read from user_name column")
		}
		return &examSubmitterRows{columns: []string{"id", "user_name"}, values: []driver.Value{int64(7), c.fixture.account}}, nil
	}
	return nil, fmt.Errorf("unexpected query: %s", query)
}

type examSubmitterRows struct {
	columns []string
	values  []driver.Value
	read    bool
}

func (r *examSubmitterRows) Columns() []string { return r.columns }
func (*examSubmitterRows) Close() error        { return nil }
func (r *examSubmitterRows) Next(dest []driver.Value) error {
	if r.read {
		return io.EOF
	}
	r.read = true
	copy(dest, r.values)
	return nil
}
