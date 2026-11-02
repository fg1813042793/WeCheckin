package middleware

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/ut"
	"strings"
	"testing"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/support/auditlog"
	"wecheckin/backend/pkg/response"
)

func TestAuditSnapshotStatusIsIndependentOfBusinessOutcome(t *testing.T) {
	for _, tt := range []struct {
		name, method, path, status, reason, want string
		failure                                  bool
		changes                                  []auditlog.Change
	}{
		{name: "login", method: "POST", path: "/api/v2/admin/auth/login", want: "not_applicable"},
		{name: "new object", method: "POST", path: "/api/v2/admin/roles", want: "not_applicable"},
		{name: "unsupported mutation", method: "PUT", path: "/api/v2/admin/roles/1", want: "not_configured"},
		{name: "unchanged", method: "PUT", path: "/api/v2/admin/roles/1", status: "unchanged", reason: "no_collected_changes", want: "unchanged"},
		{name: "read failure", method: "PUT", path: "/api/v2/admin/roles/1", status: "unavailable", reason: "before_read_failed", want: "unavailable"},
		{name: "failed operation", method: "PUT", path: "/api/v2/admin/roles/1", status: "recorded", failure: true, want: "operation_failed", changes: []auditlog.Change{{Field: "role.1.role_status", Before: 1, After: 0}}},
		{name: "safe change", method: "PUT", path: "/api/v2/admin/roles/1", want: "recorded", changes: []auditlog.Change{{Field: "role.1.role_status", Before: 1, After: 0}}},
		{name: "encoding limit", method: "PUT", path: "/api/v2/admin/roles/1", want: "unavailable", changes: []auditlog.Change{{Field: "role.1.role_name", Before: "old", After: strings.Repeat("x", 513)}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			writer := &auditMemoryWriter{}
			h := server.New()
			h.Use(auditLogWithWriter(writer, func(*model.Log, string) {}))
			h.Handle(tt.method, tt.path, func(ctx context.Context, c *app.RequestContext) {
				auditlog.Update(ctx, func(d *auditlog.Details) {
					d.SnapshotStatus = tt.status
					d.SnapshotReason = tt.reason
					d.Changes = tt.changes
				})
				if tt.failure {
					response.Fail(c, "操作失败")
				} else {
					response.JSON(c, nil)
				}
			})
			ut.PerformRequest(h.Engine, tt.method, tt.path, nil)
			if writer.events[0].SnapshotStatus != "pending" {
				t.Fatalf("begin status=%s", writer.events[0].SnapshotStatus)
			}
			got := writer.events[len(writer.events)-1]
			if got.SnapshotStatus != tt.want {
				t.Fatalf("status=%s want=%s", got.SnapshotStatus, tt.want)
			}
			if !tt.failure && got.Outcome != "success" {
				t.Fatalf("snapshot changed business outcome: %s", got.Outcome)
			}
			if tt.name == "read failure" && (got.SnapshotReason != "before_read_failed" || got.ReasonCode != "" || got.Reason != "") {
				t.Fatalf("snapshot error mixed into operation: %#v", got)
			}
			if tt.name == "encoding limit" && got.SnapshotReason != "encoding_limit" {
				t.Fatalf("reason=%s", got.SnapshotReason)
			}
			if tt.failure && string(got.Changes) != "[]" {
				t.Fatal("failed operation retained changes")
			}
		})
	}
}
