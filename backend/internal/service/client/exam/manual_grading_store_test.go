package exam

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"testing"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/support/adminaccess"
	permissionsupport "wecheckin/backend/internal/support/permission"
	"wecheckin/backend/pkg/database"
)

type gradingDBFixture struct {
	record                                    model.ExamRecord
	pending                                   *model.ExamRecord
	failWrite, denyScope, locked, hasDeadline bool
	writes, commits, rollbacks                int
}
type gradingConnector struct{ f *gradingDBFixture }

func (c gradingConnector) Connect(context.Context) (driver.Conn, error) {
	return &gradingConn{c.f}, nil
}
func (gradingConnector) Driver() driver.Driver { return gradingDriver{} }

type gradingDriver struct{}

func (gradingDriver) Open(string) (driver.Conn, error) { return nil, errors.New("use connector") }

type gradingConn struct{ f *gradingDBFixture }

func (*gradingConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("unexpected prepare")
}
func (*gradingConn) Close() error { return nil }
func (c *gradingConn) Begin() (driver.Tx, error) {
	copy := c.f.record
	c.f.pending = &copy
	return gradingTx{c.f}, nil
}

type gradingTx struct{ f *gradingDBFixture }

func (tx gradingTx) Commit() error {
	tx.f.commits++
	if tx.f.pending != nil {
		tx.f.record = *tx.f.pending
	}
	tx.f.pending = nil
	return nil
}
func (tx gradingTx) Rollback() error { tx.f.rollbacks++; tx.f.pending = nil; return nil }

type gradingRows struct {
	columns []string
	values  [][]driver.Value
	offset  int
}

func (r *gradingRows) Columns() []string { return r.columns }
func (*gradingRows) Close() error        { return nil }
func (r *gradingRows) Next(dst []driver.Value) error {
	if r.offset >= len(r.values) {
		return io.EOF
	}
	copy(dst, r.values[r.offset])
	r.offset++
	return nil
}
func gradingRow(columns []string, values ...[]driver.Value) driver.Rows {
	return &gradingRows{columns: columns, values: values}
}

func (c *gradingConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	switch {
	case strings.Contains(query, "FROM `users`"):
		roleID := int64(1)
		if c.f.denyScope {
			roleID = 0
		}
		return gradingRow([]string{"id", "user_status", "user_role_id"}, []driver.Value{int64(8), int64(1), roleID}), nil
	case strings.Contains(query, "FROM `roles`"):
		return gradingRow([]string{"id", "role_name", "role_status", "role_data_scope"}, []driver.Value{int64(1), adminaccess.ReservedSuperAdminRoleName, int64(1), int64(1)}), nil
	case strings.Contains(query, "FROM `exam`"):
		columns := []string{"exam_id", "exam_title", "exam_pass_score"}
		if strings.Contains(query, "1 = 0") || len(args) == 0 || fmt.Sprint(args[0].Value) != fmt.Sprint(c.f.record.ExamID) {
			return gradingRow(columns), nil
		}
		return gradingRow(columns, []driver.Value{int64(c.f.record.ExamID), "测试考试", int64(99)}), nil
	case strings.Contains(query, "FROM `exam_record`"):
		columns := []string{"exam_r_id", "exam_r_exam_id", "exam_r_status", "exam_r_score", "exam_r_total_score", "exam_r_answers", "exam_r_result", "exam_r_review_snapshot", "exam_r_graded_by", "exam_r_graded_at", "exam_r_grading_comment", "exam_r_user_id"}
		owner := fmt.Sprint(c.f.record.ExamID)
		if strings.Contains(query, "exam_r_user_id") {
			owner = c.f.record.UserID
		}
		if len(args) < 2 || fmt.Sprint(args[0].Value) != fmt.Sprint(c.f.record.ID) || fmt.Sprint(args[1].Value) != owner {
			return gradingRow(columns), nil
		}
		c.f.locked = c.f.locked || strings.Contains(query, "FOR UPDATE")
		r := c.f.record
		return gradingRow(columns, []driver.Value{int64(r.ID), int64(r.ExamID), int64(r.Status), int64(r.Score), int64(r.TotalScore), r.Answers, r.Result, r.ReviewSnapshot, int64(r.GradedBy), r.GradedAt, r.GradingComment, r.UserID}), nil
	case strings.HasPrefix(query, "SELECT DATABASE()"):
		return gradingRow([]string{"DATABASE()"}, []driver.Value{"grading_test"}), nil
	case strings.Contains(strings.ToLower(query), "information_schema"):
		return gradingRow([]string{"count(*)"}, []driver.Value{int64(0)}), nil
	default:
		return gradingRow([]string{"id"}), nil
	}
}

func (c *gradingConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !strings.HasPrefix(query, "UPDATE `exam_record`") {
		return nil, fmt.Errorf("unexpected write: %s", query)
	}
	c.f.writes++
	_, c.f.hasDeadline = ctx.Deadline()
	if c.f.failWrite {
		return nil, errors.New("simulated write failure")
	}
	if strings.Contains(query, "exam_r_status = 0") && c.f.record.Status != 0 {
		return driver.RowsAffected(0), nil
	}
	if c.f.pending == nil || (!strings.Contains(query, "exam_r_status = 0") && (!c.f.locked || !strings.Contains(query, "exam_r_status = 1"))) {
		return nil, errors.New("unlocked or unguarded grade write")
	}
	assignments := regexp.MustCompile("`(exam_r_[a-z_]+)`=\\?").FindAllStringSubmatch(strings.Split(query, " WHERE ")[0], -1)
	for i, column := range assignments {
		value := args[i].Value
		switch column[1] {
		case "exam_r_answers":
			c.f.pending.Answers = value.(string)
		case "exam_r_result":
			c.f.pending.Result = value.(string)
		case "exam_r_score":
			c.f.pending.Score = int(value.(int64))
		case "exam_r_total_score":
			c.f.pending.TotalScore = int(value.(int64))
		case "exam_r_pass":
			c.f.pending.Pass = int(value.(int64))
		case "exam_r_status":
			c.f.pending.Status = int(value.(int64))
		case "exam_r_graded_by":
			c.f.pending.GradedBy = uint(value.(int64))
		case "exam_r_graded_at":
			c.f.pending.GradedAt = value.(int64)
		case "exam_r_grading_comment":
			c.f.pending.GradingComment = value.(string)
		default:
			return nil, fmt.Errorf("unexpected grading column %s", column[1])
		}
	}
	return driver.RowsAffected(1), nil
}

func TestSubmittedAnswersCannotReplacePendingOrCompletedGrading(t *testing.T) {
	for _, status := range []int{1, 2} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			record, _ := gradingFixture(t, false)
			record.Status, record.UserID = status, "8"
			f := &gradingDBFixture{record: record}
			openGradingDB(t, f)
			service := NewService()
			if err := service.SaveAnswerContext(context.Background(), 19, 8, `{"essay":"changed"}`); !errors.Is(err, ErrExamRecordSubmitted) {
				t.Fatalf("answer save after submission: %v", err)
			}
			if _, _, _, err := service.PaperSubmissionContext(context.Background(), 19, 8); !errors.Is(err, ErrExamRecordSubmitted) {
				t.Fatalf("resubmit after submission: %v", err)
			}
			if f.writes != 0 {
				t.Fatal("submitted record was modified")
			}
			if err := service.UpdatePaperSubmissionContext(context.Background(), 19, map[string]interface{}{"exam_r_status": 1}); !errors.Is(err, ErrExamRecordSubmitted) {
				t.Fatalf("late submit must be conditional on ongoing state: %v", err)
			}
			if f.record.Result != record.Result || f.record.Status != status {
				t.Fatal("late submit replaced grading")
			}
		})
	}
}

func TestOngoingAnswersRemainEditable(t *testing.T) {
	record, _ := gradingFixture(t, false)
	record.Status, record.UserID = 0, "8"
	f := &gradingDBFixture{record: record}
	openGradingDB(t, f)
	if err := NewService().SaveAnswerContext(context.Background(), 19, 8, `{"essay":"draft"}`); err != nil {
		t.Fatal(err)
	}
	if f.record.Answers != `{"essay":"draft"}` || f.record.Status != 0 || f.commits != 1 {
		t.Fatal("ongoing draft save regressed")
	}
}

func openGradingDB(t *testing.T, f *gradingDBFixture) {
	t.Helper()
	connection := sql.OpenDB(gradingConnector{f})
	connection.SetMaxOpenConns(1)
	db, err := gorm.Open(mysql.New(mysql.Config{Conn: connection, SkipInitializeWithVersion: true}), &gorm.Config{DisableAutomaticPing: true, Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	previous := database.DB
	database.DB = db
	reset := func() {
		adminaccess.InvalidateAdminAccessCache()
		permissionsupport.ResetPermissionTablesReadyCache()
		permissionsupport.InvalidateRuntimePermissionCaches()
	}
	reset()
	t.Cleanup(func() { database.DB = previous; reset(); _ = connection.Close() })
}

func TestGradeRecordTransactionCommitsAndRejectsDuplicate(t *testing.T) {
	record, detail := gradingFixture(t, false)
	f := &gradingDBFixture{record: record}
	openGradingDB(t, f)
	score := 3
	request := GradeRecordRequest{ExpectedVersion: detail.Version, Scores: []ManualQuestionScore{{QuestionID: 3, Score: &score}}, Comment: "评分说明"}
	result, err := NewService().GradeRecordContext(context.Background(), 5, 19, 8, request)
	if err != nil {
		t.Fatal(err)
	}
	if f.commits != 1 || f.rollbacks != 0 || !f.locked || !f.hasDeadline || f.record.Status != 2 || f.record.Score != 7 || result.CanGrade {
		t.Fatalf("transaction not committed correctly: %+v", f)
	}
	if _, err := NewService().GradeRecordContext(context.Background(), 5, 19, 8, request); !errors.Is(err, ErrRecordNotGradable) {
		t.Fatalf("duplicate error=%v", err)
	}
	if f.writes != 1 || f.commits != 1 || f.rollbacks != 1 {
		t.Fatalf("duplicate caused another mutation: %+v", f)
	}
}

func TestGradeRecordTransactionRollsBackAndEnforcesScope(t *testing.T) {
	for _, scenario := range []string{"write failure", "no scope", "wrong exam", "stale version"} {
		t.Run(scenario, func(t *testing.T) {
			record, detail := gradingFixture(t, false)
			f := &gradingDBFixture{record: record, failWrite: scenario == "write failure", denyScope: scenario == "no scope"}
			openGradingDB(t, f)
			score := 3
			request := GradeRecordRequest{ExpectedVersion: detail.Version, Scores: []ManualQuestionScore{{QuestionID: 3, Score: &score}}}
			examID := uint(5)
			if scenario == "wrong exam" {
				examID = 99
			}
			if scenario == "stale version" {
				request.ExpectedVersion = "old"
			}
			result, err := NewService().GradeRecordContext(context.Background(), examID, 19, 8, request)
			if err == nil || result != nil || f.commits != 0 || f.rollbacks != 1 || f.record.Status != 1 || f.record.Result != record.Result {
				t.Fatalf("failed mutation was not rolled back: err=%v fixture=%+v", err, f)
			}
			if scenario != "write failure" && f.writes != 0 {
				t.Fatalf("validation/scope failure wrote data: %+v", f)
			}
			if scenario == "write failure" && f.writes != 1 {
				t.Fatal("write failure test did not reach persistence")
			}
			if scenario == "stale version" && !errors.Is(err, ErrRecordGradingChanged) {
				t.Fatalf("stale failure=%v", err)
			}
			if (scenario == "no scope" || scenario == "wrong exam") && !errors.Is(err, gorm.ErrRecordNotFound) {
				t.Fatalf("scope failure=%v", err)
			}
		})
	}
}
