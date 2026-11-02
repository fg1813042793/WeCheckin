package adminlog

import (
	"context"
	"errors"
	"strings"
	"testing"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/support/auditlog"
)

func TestSecurityPermissionsDoNotDiscloseThroughListDetailOrExport(t *testing.T) {
	for _, tc := range []struct {
		name, category string
		keys, denied   []string
		export         bool
	}{
		{name: "ordinary login list", category: CategoryLogin, keys: []string{"admin:api:log:list"}},
		{name: "ordinary login export", category: CategoryLogin, keys: []string{"admin:api:log:list", "admin:api:log:export"}, export: true},
		{name: "security export needs read", category: CategoryLogin, keys: []string{"admin:api:log:security-export"}, export: true},
		{name: "security export needs export", category: CategoryLogin, keys: []string{"admin:api:log:security"}, export: true},
		{name: "operations export needs read", category: CategoryOperation, keys: []string{"admin:api:log:export"}, export: true},
		{name: "operations export needs export", category: CategoryOperation, keys: []string{"admin:api:log:list"}, export: true},
		{name: "explicit user deny", category: CategoryLogin, keys: []string{"admin:api:log:security"}, denied: []string{"admin:api:log:security"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := &auditDBFixture{permissions: tc.keys, denied: tc.denied}
			openAuditTestDB(t, f)
			_, err := listContext(context.Background(), tc.category, Query{Page: 1, PageSize: 20}, 7, tc.export)
			if !errors.Is(err, ErrForbidden) || len(f.logQueries) != 0 {
				t.Fatalf("disclosure: err=%v queries=%v", err, f.logQueries)
			}
			if !tc.export {
				_, err = DetailContext(context.Background(), tc.category, 41, 7)
				if !errors.Is(err, ErrForbidden) || len(f.logQueries) != 0 {
					t.Fatalf("detail disclosure: %v %v", err, f.logQueries)
				}
			}
		})
	}
}

func TestAuditPagesExcludeCurrentEventFromCountAndRowsOnly(t *testing.T) {
	f := &auditDBFixture{super: true}
	openAuditTestDB(t, f)
	ctx, _ := auditlog.WithRecorder(context.Background())
	const eventID = "audit-current-export"
	auditlog.Update(ctx, func(d *auditlog.Details) { d.EventID = eventID })
	for _, category := range []string{CategoryOperation, CategoryLogin} {
		for _, export := range []bool{false, true} {
			before := len(f.logQueries)
			if _, err := listContext(ctx, category, Query{Page: 1, PageSize: 20}, 7, export); err != nil {
				t.Fatal(err)
			}
			if len(f.logQueries)-before != 2 {
				t.Fatal("expected count and page queries")
			}
			for i := before; i < len(f.logQueries); i++ {
				if !strings.Contains(f.logQueries[i], "(`log_event_id` <> ? OR `log_event_id` IS NULL)") || !testHasArg(f.logArgs[i], eventID) {
					t.Errorf("current event not excluded while retaining legacy NULL: %s %#v", f.logQueries[i], f.logArgs[i])
				}
			}
		}
		if _, err := DetailContext(ctx, category, 41, 7); err != nil {
			t.Fatal(err)
		}
		last := len(f.logQueries) - 1
		if strings.Contains(f.logQueries[last], "`log_event_id` <>") || testHasArg(f.logArgs[last], eventID) {
			t.Error("detail must not exclude the current event")
		}
	}
	for _, export := range []bool{false, true} {
		before := len(f.logQueries)
		if _, err := listContext(context.Background(), CategoryOperation, Query{Page: 1, PageSize: 20}, 7, export); err != nil {
			t.Fatal(err)
		}
		for _, query := range f.logQueries[before:] {
			if strings.Contains(query, "`log_event_id` <>") {
				t.Error("missing recorder must not introduce an event exclusion")
			}
		}
	}
}

func TestAuditQueryRequiresExistingActorAndSuccessfulScope(t *testing.T) {
	for name, f := range map[string]*auditDBFixture{
		"missing": {missingActor: true}, "lookup error": {actorFailure: true},
		"scope error": {scopeFailure: true, permissions: []string{"admin:api:log:list"}},
	} {
		t.Run(name, func(t *testing.T) {
			openAuditTestDB(t, f)
			_, err := ListContext(context.Background(), CategoryOperation, Query{Page: 1, PageSize: 20}, 7)
			if err == nil || len(f.logQueries) != 0 {
				t.Fatalf("query did not fail closed: %v %v", err, f.logQueries)
			}
		})
	}
}

func TestOperationQueriesRetainScopeAndCategory(t *testing.T) {
	f := &auditDBFixture{permissions: []string{"admin:api:log:list", "admin:api:log:export"}}
	openAuditTestDB(t, f)
	q := Query{Page: 1, PageSize: 20, Search: "%admin_"}
	page, err := ListContext(context.Background(), CategoryOperation, q, 7)
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 1205 || len(page.List) != 1 || page.List[0].Outcome != "unknown" {
		t.Fatalf("page: %+v", page)
	}
	if _, err = DetailContext(context.Background(), CategoryOperation, 41, 7); err != nil {
		t.Fatal(err)
	}
	q.PageSize = 1000
	if _, err = ExportContext(context.Background(), CategoryOperation, q, 7); err != nil {
		t.Fatal(err)
	}
	for i, sql := range f.logQueries {
		if !strings.Contains(sql, "`log_category` = ?") || !strings.Contains(sql, "`log_type` <> 1") || !strings.Contains(sql, "`create_by` = ?") || !testHasArg(f.logArgs[i], CategoryOperation) || !testHasArg(f.logArgs[i], "7") {
			t.Fatalf("unscoped operation SQL: %s %#v", sql, f.logArgs[i])
		}
	}
	if !strings.Contains(f.logQueries[len(f.logQueries)-1], "LIMIT ?") || !testHasArg(f.logArgs[len(f.logArgs)-1], "1000") {
		t.Fatal("export not bounded")
	}
	if !testHasArg(f.logArgs[0], "%!%admin!_%") {
		t.Fatalf("search not literal: %#v", f.logArgs[0])
	}
}

func TestSecurityReaderSeesAnonymousAttemptsWithoutActorScope(t *testing.T) {
	f := &auditDBFixture{permissions: []string{"admin:api:log:security", "admin:api:log:security-export"}, scopeFailure: true}
	openAuditTestDB(t, f)
	page, err := ExportContext(context.Background(), CategoryLogin, Query{Page: 1, PageSize: 1000}, 7)
	if err != nil || page.Total != 1205 || len(page.List) != 1 || page.List[0].AdminID != 0 || page.List[0].Category != CategoryLogin {
		t.Fatalf("security page: %+v %v", page, err)
	}
	if _, err := DetailContext(context.Background(), CategoryLogin, 41, 7); err != nil {
		t.Fatal(err)
	}
	for _, sql := range f.logQueries {
		if strings.Contains(strings.SplitN(sql, " WHERE ", 2)[1], "create_by") || !strings.Contains(sql, "`log_type` = 1") {
			t.Fatalf("security SQL: %s", sql)
		}
	}
}

func TestAuditCountFailureReturnsNoPartialPage(t *testing.T) {
	f := &auditDBFixture{super: true, countFailure: true}
	openAuditTestDB(t, f)
	page, err := ListContext(context.Background(), CategoryOperation, Query{Page: 1, PageSize: 20}, 7)
	if err == nil || page.List != nil || len(f.logQueries) != 1 {
		t.Fatalf("partial page: %+v %v %v", page, err, f.logQueries)
	}
}

func TestOrdinaryAllScopeExcludesAnonymousOperationsAndAllLoginRows(t *testing.T) {
	f := &auditDBFixture{permissions: []string{"admin:api:log:list", "admin:api:log:export"}, scopeAll: true}
	openAuditTestDB(t, f)
	for _, source := range []string{"admin", "client", "h5app"} {
		if _, err := ExportContext(context.Background(), CategoryOperation, Query{Page: 1, PageSize: 1000, Source: source}, 7); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := DetailContext(context.Background(), CategoryOperation, 41, 7); err != nil {
		t.Fatal(err)
	}
	for i, query := range f.logQueries {
		if !strings.Contains(query, "`create_by` > 0") || !testHasArg(f.logArgs[i], CategoryOperation) || !strings.Contains(query, "`log_type` <> 1") {
			t.Fatalf("all-scope disclosure: %s %#v", query, f.logArgs[i])
		}
	}
	before := len(f.logQueries)
	if _, err := DetailContext(context.Background(), CategoryLogin, 41, 7); !errors.Is(err, ErrForbidden) || len(f.logQueries) != before {
		t.Fatalf("login disclosure: %v", err)
	}
}

func TestAuditDetailDoesNotDistinguishInvisibleFromMissing(t *testing.T) {
	f := &auditDBFixture{permissions: []string{"admin:api:log:list"}, emptyDetail: true}
	openAuditTestDB(t, f)
	_, err := DetailContext(context.Background(), CategoryOperation, 41, 7)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("got %v", err)
	}
}

func TestSuperAdminBypassAndAppendStorage(t *testing.T) {
	f := &auditDBFixture{super: true}
	db := openAuditTestDB(t, f)
	if _, err := ExportContext(context.Background(), CategoryLogin, Query{Page: 1, PageSize: 1000}, 7); err != nil {
		t.Fatal(err)
	}
	record := model.Log{Category: CategoryOperation, Source: "admin", AdminID: 7, AdminName: "actor snapshot"}
	if err := CreateWithDBContext(context.Background(), db, &record); err != nil {
		t.Fatal(err)
	}
	if record.EventID == "" || len(f.writes) != 1 || !strings.Contains(f.writes[0], "`log_event_id`") || !strings.Contains(f.writes[0], "`log_changes`") {
		t.Fatalf("storage: %+v %v", record, f.writes)
	}
	for _, sql := range f.logQueries {
		if strings.Contains(strings.SplitN(sql, " WHERE ", 2)[1], "create_by") {
			t.Fatalf("superadmin scoped: %s", sql)
		}
	}
}

func TestAuditPagesOmitHeavyColumnsAndDetailsRetainThem(t *testing.T) {
	f := &auditDBFixture{super: true}
	openAuditTestDB(t, f)
	for _, category := range []string{CategoryOperation, CategoryLogin} {
		for _, export := range []bool{false, true} {
			page, err := listContext(context.Background(), category, Query{Page: 1, PageSize: 20}, 7, export)
			if err != nil {
				t.Fatal(err)
			}
			query := f.logQueries[len(f.logQueries)-1]
			selection := strings.SplitN(query, " FROM ", 2)[0]
			if strings.Contains(selection, "*") || strings.Contains(selection, "log_changes") || strings.Contains(selection, "log_user_agent") || strings.Contains(selection, "log_target_ids") {
				t.Errorf("heavy page selection: %s", selection)
			}
			if len(page.List) != 1 || string(page.List[0].Changes) != "[]" || page.List[0].UserAgent != "" || string(page.List[0].TargetIDs) != "[]" {
				t.Errorf("page should omit detail payload: %+v", page.List)
			}
			if len(page.List) == 1 && (page.List[0].RequestedCount != 3 || page.List[0].AffectedCount != 1) {
				t.Errorf("page conflated requested and affected counts: %+v", page.List[0])
			}
		}
		detail, err := DetailContext(context.Background(), category, 41, 7)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(detail.Changes), `"field":"status"`) || detail.UserAgent != "detail-only user agent" {
			t.Errorf("detail lost payload: %+v", detail)
		}
		if string(detail.TargetIDs) != `["41","42","43"]` || detail.RequestedCount != 3 || detail.AffectedCount != 1 {
			t.Errorf("detail lost requested IDs or conflated counts: %+v", detail)
		}
	}
}
