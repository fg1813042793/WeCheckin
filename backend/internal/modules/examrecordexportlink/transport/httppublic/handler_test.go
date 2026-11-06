package httppublic

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	exportlinkapp "wecheckin/backend/internal/modules/examrecordexportlink/application"
	"wecheckin/backend/internal/service/dingtalkh5/workflowsummary"
)

type resolverStub struct {
	grant exportlinkapp.Grant
	err   error
	token string
}

func (resolver *resolverStub) Resolve(_ context.Context, token string) (exportlinkapp.Grant, error) {
	resolver.token = token
	return resolver.grant, resolver.err
}

type exporterStub struct {
	examID, recordID uint
	format           workflowsummary.ExportFormat
	result           *workflowsummary.ExportResult
	err              error
}

func (exporter *exporterStub) ExportRecordForLinkContext(_ context.Context, examID, recordID uint, format workflowsummary.ExportFormat) (*workflowsummary.ExportResult, error) {
	exporter.examID, exporter.recordID, exporter.format = examID, recordID, format
	return exporter.result, exporter.err
}

func TestDownloadUsesOnlyTokenBoundRecordAndFormat(t *testing.T) {
	resolver := &resolverStub{grant: exportlinkapp.Grant{ExamID: 5, RecordID: 19, Format: "docx", ExpiresAt: time.Now().Add(time.Hour)}}
	exporter := &exporterStub{result: &workflowsummary.ExportResult{Filename: "answer.docx", ContentType: "application/docx", Body: []byte("document")}}
	c := app.NewContext(0)
	c.Request.SetRequestURI("/api/v2/public/exam-record-exports/download?token=secret&examId=999&recordId=888&format=xlsx")
	NewHandler(resolver, exporter).Download(context.Background(), c)

	if resolver.token != "secret" || exporter.examID != 5 || exporter.recordID != 19 || exporter.format != workflowsummary.ExportFormatDOCX {
		t.Fatalf("resolver token=%q export=%d/%d/%s", resolver.token, exporter.examID, exporter.recordID, exporter.format)
	}
	if c.Response.StatusCode() != 200 || string(c.Response.Body()) != "document" || string(c.Response.Header.Peek("Content-Disposition")) != `attachment; filename="answer.docx"` {
		t.Fatalf("status=%d headers=%s body=%s", c.Response.StatusCode(), c.Response.Header.Header(), c.Response.Body())
	}
}

func TestDownloadReturnsGoneWithoutLeakingErrors(t *testing.T) {
	for _, err := range []error{exportlinkapp.ErrGrantNotFound, exportlinkapp.ErrGrantExpired} {
		c := app.NewContext(0)
		c.Request.SetRequestURI("/api/v2/public/exam-record-exports/download?token=secret")
		NewHandler(&resolverStub{err: err}, &exporterStub{}).Download(context.Background(), c)
		if c.Response.StatusCode() != 410 || !strings.Contains(string(c.Response.Body()), "答卷下载链接已失效") {
			t.Fatalf("status=%d body=%s", c.Response.StatusCode(), c.Response.Body())
		}
	}
	c := app.NewContext(0)
	c.Request.SetRequestURI("/api/v2/public/exam-record-exports/download?token=secret")
	NewHandler(&resolverStub{grant: exportlinkapp.Grant{ExamID: 5, RecordID: 19, Format: "pdf", ExpiresAt: time.Now().Add(time.Hour)}}, &exporterStub{err: errors.New("SQL private")}).Download(context.Background(), c)
	if c.Response.StatusCode() != 500 || strings.Contains(string(c.Response.Body()), "SQL") {
		t.Fatalf("unsafe response status=%d body=%s", c.Response.StatusCode(), c.Response.Body())
	}
}
