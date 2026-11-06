package adminlog

import (
	"context"
	"errors"
	"strings"
	"testing"

	"wecheckin/backend/internal/support/auditlog"
)

func TestDeleteLogsRequiresReadAndSeparateDeletePermissions(t *testing.T) {
	for _, tc := range []struct {
		category     string
		keys, denied []string
	}{
		{CategoryOperation, []string{"admin:api:log:list"}, nil},
		{CategoryOperation, []string{"admin:api:log:delete"}, nil},
		{CategoryOperation, []string{"admin:api:log:list", "admin:api:log:del"}, nil},
		{CategoryLogin, []string{"admin:api:log:security", "admin:api:log:delete"}, nil},
		{CategoryLogin, []string{"admin:api:log:security-delete"}, nil},
		{CategoryOperation, []string{"admin:api:log:list", "admin:api:log:delete"}, []string{"admin:api:log:delete"}},
	} {
		f := &auditDBFixture{permissions: tc.keys, denied: tc.denied, deleteIDs: []uint{41}}
		openAuditTestDB(t, f)
		_, err := DeleteForAdminContext(context.Background(), tc.category, []uint{41}, 7)
		if !errors.Is(err, ErrForbidden) || len(f.writes) != 0 || len(f.logQueries) != 0 {
			t.Fatalf("delete allowed without permissions: err=%v writes=%v queries=%v", err, f.writes, f.logQueries)
		}
	}
}

func TestDeleteLogsIsScopedAtomicAndAudited(t *testing.T) {
	for _, category := range []string{CategoryOperation, CategoryLogin} {
		t.Run(category, func(t *testing.T) {
			keys := []string{"admin:api:log:list", "admin:api:log:delete"}
			if category == CategoryLogin {
				keys = []string{"admin:api:log:security", "admin:api:log:security-delete"}
			}
			f := &auditDBFixture{permissions: keys, deleteIDs: []uint{41, 42}}
			openAuditTestDB(t, f)
			ctx, recorder := auditlog.WithRecorder(context.Background())
			auditlog.Update(ctx, func(d *auditlog.Details) { d.EventID = "current-delete" })
			result, err := DeleteForAdminContext(ctx, category, []uint{41, 42}, 7)
			if err != nil || result.DeletedCount != 2 || f.commits != 1 || f.rollbacks != 0 || len(f.writes) != 1 {
				t.Fatalf("result=%+v err=%v fixture=%+v", result, err, f)
			}
			for i, sql := range append(append([]string{}, f.logQueries...), f.writes...) {
				if !strings.Contains(sql, "`log_category`") || !strings.Contains(sql, "`logs`.`id` IN") {
					t.Fatalf("unscoped SQL: %s", sql)
				}
				if category == CategoryOperation && !strings.Contains(sql, "`create_by`") {
					t.Fatalf("missing data scope: %s", sql)
				}
				if i == 0 && (!strings.Contains(sql, "FOR UPDATE") || !testHasArg(f.logArgs[0], "current-delete")) {
					t.Fatalf("missing lock/event exclusion: %s", sql)
				}
			}
			d := recorder.Snapshot()
			if d.AffectedCount != 2 || len(d.TargetIDs) != 2 || d.TargetType != "audit_log" {
				t.Fatalf("audit=%+v", d)
			}
		})
	}
}

func TestDeleteLogsRejectsIncompleteTargetsAndRollsBackFailure(t *testing.T) {
	for _, failedWrite := range []bool{false, true} {
		f := &auditDBFixture{super: true, deleteIDs: []uint{41}, deleteFailure: failedWrite}
		if failedWrite {
			f.deleteIDs = []uint{41, 42}
		}
		openAuditTestDB(t, f)
		ctx, recorder := auditlog.WithRecorder(context.Background())
		result, err := DeleteForAdminContext(ctx, CategoryOperation, []uint{41, 42}, 7)
		if err == nil || result.DeletedCount != 0 || f.commits != 0 || f.rollbacks != 1 || recorder.Snapshot().AffectedCount != 0 {
			t.Fatalf("partial delete: result=%+v err=%v fixture=%+v", result, err, f)
		}
		if !failedWrite && (!errors.Is(err, ErrNotFound) || len(f.writes) != 0) {
			t.Fatalf("incomplete selection wrote rows: %v", f.writes)
		}
	}
}

func TestDeleteLogsRejectsEmptyInvalidAndOversizedBatches(t *testing.T) {
	for _, ids := range [][]uint{nil, {}, {0}, {1, 1}, make([]uint, 101)} {
		f := &auditDBFixture{super: true}
		openAuditTestDB(t, f)
		_, err := DeleteForAdminContext(context.Background(), CategoryOperation, ids, 7)
		var validation *ValidationError
		if !errors.As(err, &validation) || len(f.writes) != 0 {
			t.Fatalf("invalid IDs accepted: %v", err)
		}
	}
}
