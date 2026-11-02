package adminlog_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"strings"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route/param"
	"wecheckin/backend/internal/handler/admin/mgr"
	"wecheckin/backend/internal/model"
	adminlogservice "wecheckin/backend/internal/service/admin/adminlog"
	"wecheckin/backend/internal/support/auditlog"
	"wecheckin/backend/pkg/logger"
)

func TestAuditHandlersRejectMissingActorWithoutDisclosure(t *testing.T) {
	h := &mgr.AdminMgrHandler{}
	for _, handler := range []app.HandlerFunc{h.GetLogList, h.GetLoginLogList, h.ExportLogs, h.ExportLoginLogs, h.GetLogDetail, h.GetLoginLogDetail} {
		c := app.NewContext(1)
		c.Params = param.Params{{Key: "id", Value: "1"}}
		handler(context.Background(), c)
		var body struct {
			Code int             `json:"code"`
			Data json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(c.Response.Body(), &body); err != nil {
			t.Fatal(err)
		}
		if body.Code == 0 || len(body.Data) != 0 {
			t.Fatalf("unauthenticated disclosure: %s", c.Response.Body())
		}
	}
}

func TestAuditExportRecordsActualReturnedCount(t *testing.T) {
	h := &mgr.AdminMgrHandler{}
	for name, handler := range map[string]app.HandlerFunc{"operation": h.ExportLogs, "login": h.ExportLoginLogs} {
		for _, empty := range []bool{false, true} {
			t.Run(name+map[bool]string{false: " rows", true: " empty"}[empty], func(t *testing.T) {
				adminlogservice.OpenHandlerTestDB(t, "", empty)
				c := app.NewContext(0)
				c.Set("admin", &model.Admin{ID: 7})
				ctx, recorder := auditlog.WithRecorder(context.Background())
				handler(ctx, c)
				var body struct {
					Code int                  `json:"code"`
					Data adminlogservice.Page `json:"data"`
				}
				if err := json.Unmarshal(c.Response.Body(), &body); err != nil {
					t.Fatal(err)
				}
				detail := recorder.Snapshot()
				if body.Code != 0 || detail.TargetType != "audit_log" || detail.AffectedCount != len(body.Data.List) || detail.AffectedCount == int(body.Data.Total) {
					t.Fatalf("export=%s audit=%+v", c.Response.Body(), detail)
				}
			})
		}
	}
}

func TestAuditHandlerLogsInternalErrorsWithoutResponseDisclosure(t *testing.T) {
	h := &mgr.AdminMgrHandler{}
	for _, failure := range []string{"count", "actor"} {
		t.Run(failure, func(t *testing.T) {
			adminlogservice.OpenHandlerTestDB(t, failure, false)
			var logged bytes.Buffer
			oldLogger := logger.Logger
			logger.Logger = log.New(&logged, "", 0)
			t.Cleanup(func() { logger.Logger = oldLogger })
			c := app.NewContext(0)
			c.Set("admin", &model.Admin{ID: 7})
			c.Request.Header.Set("X-Request-ID", "audit-query-test")
			ctx, recorder := auditlog.WithRecorder(context.Background())
			h.ExportLogs(ctx, c)
			body := string(c.Response.Body())
			private := "count failed"
			if failure == "actor" {
				private = "actor lookup failed"
			}
			if !strings.Contains(body, "获取审计日志失败") || strings.Contains(body, private) {
				t.Fatalf("unsafe error response: %s", body)
			}
			for _, want := range []string{"[HTTP_INTERNAL_ERROR]", "operation=admin.auditlog", "request_id=audit-query-test", private} {
				if !strings.Contains(logged.String(), want) {
					t.Errorf("diagnostic missing %q: %s", want, logged.String())
				}
			}
			if d := recorder.Snapshot(); d.TargetType != "" || d.AffectedCount != 0 {
				t.Errorf("failed export reported success: %+v", d)
			}
		})
	}
}

func TestAuditHandlersValidateQueryAndRejectDeletion(t *testing.T) {
	h := &mgr.AdminMgrHandler{}
	c := app.NewContext(0)
	c.Set("admin", &model.Admin{ID: 7})
	c.Request.SetRequestURI("/api/v2/admin/logs?pageSize=1001")
	h.GetLogList(context.Background(), c)
	if !strings.Contains(string(c.Response.Body()), "pageSize") {
		t.Fatalf("query validation: %s", c.Response.Body())
	}
	c = app.NewContext(0)
	h.ClearLog(context.Background(), c)
	if !strings.Contains(string(c.Response.Body()), "只读") || strings.Contains(string(c.Response.Body()), `"code":0`) {
		t.Fatalf("delete: %s", c.Response.Body())
	}
}
