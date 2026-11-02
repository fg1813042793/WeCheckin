package workflowsummary

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"strings"
	"testing"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"wecheckin/backend/internal/model"
)

type nameDBFixture struct {
	names    map[uint]string
	queries  int
	err      error
	deadline bool
}
type nameConnector struct{ fixture *nameDBFixture }

func (c nameConnector) Connect(context.Context) (driver.Conn, error) {
	return &nameConn{c.fixture}, nil
}
func (c nameConnector) Driver() driver.Driver { return nameDriver{} }

type nameDriver struct{}

func (nameDriver) Open(string) (driver.Conn, error) { return nil, errors.New("use connector") }

type nameConn struct{ fixture *nameDBFixture }

func (*nameConn) Prepare(string) (driver.Stmt, error) { return nil, errors.New("unexpected prepare") }
func (*nameConn) Close() error                        { return nil }
func (*nameConn) Begin() (driver.Tx, error)           { return nil, errors.New("unexpected transaction") }
func (c *nameConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	c.fixture.queries++
	_, c.fixture.deadline = ctx.Deadline()
	if c.fixture.err != nil {
		return nil, c.fixture.err
	}
	if !strings.Contains(query, "workflow_definitions") || !strings.Contains(query, "definition_deleted_at") || strings.Contains(query, "definition_status") || strings.Contains(query, "definition_display_name") {
		return nil, errors.New("expected a current management-name-only lookup, including disabled definitions")
	}
	rows := &nameRows{}
	for _, arg := range args {
		id, ok := arg.Value.(int64)
		if ok && id > 0 {
			if name, exists := c.fixture.names[uint(id)]; exists {
				rows.values = append(rows.values, []driver.Value{id, name})
			}
		}
	}
	return rows, nil
}

type nameRows struct {
	values [][]driver.Value
	offset int
}

func (*nameRows) Columns() []string { return []string{"id", "definition_name"} }
func (*nameRows) Close() error      { return nil }
func (r *nameRows) Next(dest []driver.Value) error {
	if r.offset == len(r.values) {
		return io.EOF
	}
	copy(dest, r.values[r.offset])
	r.offset++
	return nil
}
func nameTestDB(t *testing.T, fixture *nameDBFixture) *gorm.DB {
	t.Helper()
	connection := sql.OpenDB(nameConnector{fixture})
	t.Cleanup(func() { _ = connection.Close() })
	db, err := gorm.Open(mysql.New(mysql.Config{Conn: connection, SkipInitializeWithVersion: true}), &gorm.Config{DisableAutomaticPing: true, Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func TestExportReadsRenamedManagementNameAndKeepsHistoricalForm(t *testing.T) {
	for _, content := range []ExportContent{ExportContentFull, ExportContentForm} {
		for _, format := range []ExportFormat{ExportFormatXLSX, ExportFormatPDF, ExportFormatDOCX} {
			runtime, ids := contentExportFixture(1)
			fixture := &nameDBFixture{names: map[uint]string{1: "MH-通用绩效考评单"}}
			service := NewService(nameTestDB(t, fixture), runtime)
			service.access = summaryAccessStub{}
			for _, currentName := range []string{"MH-通用绩效考评单", "MH-应届实习生绩效考评单"} {
				fixture.names[1] = currentName
				result, err := service.Export(context.Background(), &model.DingTalkH5PerfUser{ID: 7}, ExportRequest{InstanceIDs: ids, Format: format, Content: content})
				if err != nil {
					t.Fatal(err)
				}
				body := result.Body
				if format == ExportFormatXLSX {
					entries := readZipEntries(t, body)
					if !strings.Contains(entries["xl/workbook.xml"], "张三-"+currentName) {
						t.Errorf("sheet still uses old name: %s", entries["xl/workbook.xml"])
					}
					body = []byte(entries["xl/worksheets/sheet1.xml"])
				}
				text := exportedText(t, body, format)
				if !strings.Contains(text, "张三-"+currentName) || strings.Contains(text, "历史考评表") || !strings.Contains(text, "历史字段甲") || !strings.Contains(text, "历史值甲") {
					t.Errorf("%s/%s export uses old name or changed schema: %s", content, format, text)
				}
				if runtime.details[ids[0]].Instance.DefinitionName != "历史考评表" || runtime.details[ids[0]].Instance.DefinitionVersion != 2 {
					t.Fatal("historical instance was modified")
				}
			}
			if fixture.queries != 2 || !fixture.deadline {
				t.Errorf("metadata should refresh once per export with a deadline: %+v", fixture)
			}
		}
	}
}

func TestExportManagementNameFallbackFailureAndAuthorization(t *testing.T) {
	runtime, ids := contentExportFixture(2)
	fixture := &nameDBFixture{names: map[uint]string{1: "当前管理名称"}}
	service := NewService(nameTestDB(t, fixture), runtime)
	service.access = summaryAccessStub{}
	result, err := service.Export(context.Background(), &model.DingTalkH5PerfUser{ID: 7}, ExportRequest{InstanceIDs: ids, Format: ExportFormatXLSX, Content: ExportContentForm})
	if err != nil {
		t.Fatal(err)
	}
	workbook := readZipEntries(t, result.Body)["xl/workbook.xml"]
	if !strings.Contains(workbook, "张三-当前管理名称") || !strings.Contains(workbook, "张三-历史考评表") || fixture.queries != 1 {
		t.Fatalf("batch name lookup/fallback: %s, queries %d", workbook, fixture.queries)
	}
	fixture.err = errors.New("name lookup failed")
	result, err = service.Export(context.Background(), &model.DingTalkH5PerfUser{ID: 7}, ExportRequest{InstanceIDs: ids, Format: ExportFormatXLSX})
	if !errors.Is(err, fixture.err) || result != nil {
		t.Fatal("DB errors must not silently export stale names")
	}
	queries := fixture.queries
	runtime.list.List = runtime.list.List[:1]
	_, err = service.Export(context.Background(), &model.DingTalkH5PerfUser{ID: 7}, ExportRequest{InstanceIDs: ids, Format: ExportFormatXLSX})
	if !errors.Is(err, ErrSummaryAccessDenied) || fixture.queries != queries {
		t.Fatal("names must only load after authorizing the entire selection")
	}
}
