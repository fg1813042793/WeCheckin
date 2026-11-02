package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/ut"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/support/auditlog"
	"wecheckin/backend/pkg/database"
)

func TestAuditSnapshotDiffPrivateAndUnchanged(t *testing.T) {
	before := snapshotData{values: map[string]interface{}{"user.1.user_status": "1", "user.1.user_name": "same"}, private: map[string]string{"settings.SECRET": "old-secret", "settings.SAME": "unchanged-secret"}}
	after := snapshotData{values: map[string]interface{}{"user.1.user_status": "0", "user.1.user_name": "same"}, private: map[string]string{"settings.SECRET": "new-secret", "settings.SAME": "unchanged-secret"}}
	got, err := snapshotDiff(before, after)
	if err != nil || len(got) != 2 {
		t.Fatalf("diff=%#v err=%v", got, err)
	}
	want := []auditlog.Change{{Field: "settings.SECRET", Before: "[已隐藏]", After: "[已变更]"}, {Field: "user.1.user_status", Before: "1", After: "0"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("diff=%#v", got)
	}
	raw, _ := json.Marshal(got)
	if strings.Contains(string(raw), "secret") || strings.Contains(string(raw), "same") {
		t.Fatal("private or unchanged value escaped")
	}
}

func TestAuditSnapshotPriorityRoutes(t *testing.T) {
	cases := []struct{ method, route, family string }{
		{"PUT", "/users/:id", "user"}, {"PUT", "/managers/:id", "user"},
		{"PATCH", "/users/:id/status", "user"}, {"DELETE", "/users", "user"},
		{"PUT", "/roles/:id", "role"}, {"DELETE", "/roles", "role"},
		{"PUT", "/permissions/:key", "permission"},
		{"PUT", "/workflow-org-approver-assignments", "org_approver"},
		{"PUT", "/settings", "settings"}, {"PUT", "/settings/content", "settings"},
		{"PUT", "/dingtalk/settings", "settings"},
		{"PUT", "/surveys/:id", "survey"}, {"PATCH", "/surveys/:id/status", "survey"},
		{"PUT", "/exams/:id", "exam"}, {"PATCH", "/exams/:id/status", "exam"},
		{"PUT", "/workflow-definitions/:id", "workflow_definition"},
		{"POST", "/workflow-definitions/:id/publish", "workflow_definition"},
		{"POST", "/workflow-definitions/:id/versions/:version/rollback", "workflow_definition"},
		{"PUT", "/departments/:id", "department"}, {"PUT", "/positions/:id", "position"},
		{"PUT", "/dict/items/:id", "dict_item"}, {"PUT", "/dict/types/:typeCode", "dict_type"},
		{"DELETE", "/dict/types/:typeCode/items", "dict_item"},
	}
	for _, tc := range cases {
		t.Run(tc.method+tc.route, func(t *testing.T) {
			spec, ok := snapshotRoute(tc.method, "/api/v2/admin"+tc.route)
			if !ok || spec.family != tc.family {
				t.Fatalf("spec=%#v ok=%v", spec, ok)
			}
		})
	}
	for _, route := range []string{"/api/v2/admin/users/:id/unknown", "/api/v2/admin/auth/login", "/api/v2/admin/survey-responses/:id", "/api/v2/admin/exams/:id/records/:recordId", "/api/v2/me", "/api/v2/admin/workflow-definitions/:id/validate"} {
		if _, ok := snapshotRoute("POST", route); ok {
			t.Fatalf("unexpected snapshot: %s", route)
		}
	}
	if _, ok := snapshotRoute("GET", "/api/v2/admin/users/:id"); ok {
		t.Fatal("read route mapped")
	}
}

func runSnapshotRequest(t *testing.T, response string, status int, denied bool, read snapshotReader) (auditlog.Details, int) {
	t.Helper()
	h := server.New()
	var recorder *auditlog.Recorder
	called := 0
	h.Use(func(ctx context.Context, c *app.RequestContext) {
		ctx, recorder = auditlog.WithRecorder(ctx)
		c.Set("admin", &model.Admin{ID: 9})
		if denied {
			c.String(403, `{"code":1}`)
			c.Abort()
			return
		}
		c.Next(ctx)
	})
	h.PUT("/api/v2/admin/users/:id", auditSnapshotWithReader(read), func(ctx context.Context, c *app.RequestContext) {
		called++
		c.String(status, response)
	})
	result := ut.PerformRequest(h.Engine, "PUT", "/api/v2/admin/users/7", nil).Result()
	if !denied && (result.StatusCode() != status || string(result.Body()) != response) {
		t.Fatalf("collector changed business response: status=%d body=%s", result.StatusCode(), result.Body())
	}
	return recorder.Snapshot(), called
}

func TestAuditSnapshotSuccessAndFailureGating(t *testing.T) {
	for _, tc := range []struct {
		name, response         string
		status, reads, changes int
		denied                 bool
	}{
		{"success", `{"code":0}`, 200, 2, 1, false},
		{"business failure", `{"code":1}`, 200, 1, 0, false},
		{"missing code", `{}`, 200, 1, 0, false},
		{"null code", `{"code":null}`, 200, 1, 0, false},
		{"string code", `{"code":"0"}`, 200, 1, 0, false},
		{"http failure", `{"code":0}`, 500, 1, 0, false},
		{"malformed", `{"code":0`, 200, 1, 0, false},
		{"denied", `{"code":0}`, 200, 0, 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reads := 0
			d, called := runSnapshotRequest(t, tc.response, tc.status, tc.denied, func(_ context.Context, p snapshotPlan) (snapshotData, error) {
				reads++
				if !reflect.DeepEqual(p.ids, []string{"7"}) {
					t.Fatalf("ids=%v", p.ids)
				}
				return snapshotData{values: map[string]interface{}{"user.7.user_status": reads}, name: "private-before", count: 1}, nil
			})
			if reads != tc.reads || len(d.Changes) != tc.changes {
				t.Fatalf("reads=%d details=%#v", reads, d)
			}
			if !tc.denied && called != 1 {
				t.Fatal("handler not called exactly once")
			}
			if tc.changes == 0 && d.TargetName != "" {
				t.Fatal("before data attached to failed request")
			}
			if tc.changes > 0 && (d.SnapshotStatus != "recorded" || d.SnapshotReason != "") {
				t.Fatalf("successful comparison status=%q reason=%q", d.SnapshotStatus, d.SnapshotReason)
			}
			if tc.changes == 0 && (d.SnapshotStatus != "" || d.SnapshotReason != "") {
				t.Fatal("collector finalized a failed request instead of leaving it to the parent")
			}
		})
	}
}

func TestAuditSnapshotReadFailureDoesNotFailMutationOrLeak(t *testing.T) {
	for _, tc := range []struct {
		state  string
		failAt int
		err    error
	}{
		{"before_read_failed", 1, errors.New("SQL private-secret")},
		{"after_read_failed", 2, errors.New("SQL private-secret")},
		{"before_limit", 1, errSnapshotLimit},
		{"after_limit", 2, errSnapshotLimit},
	} {
		t.Run(tc.state, func(t *testing.T) {
			reads := 0
			d, called := runSnapshotRequest(t, `{"code":0}`, 200, false, func(_ context.Context, _ snapshotPlan) (snapshotData, error) {
				reads++
				if reads == tc.failAt {
					return snapshotData{}, tc.err
				}
				return snapshotData{values: map[string]interface{}{"user.7.user_name": "private-before"}}, nil
			})
			if called != 1 || reads != tc.failAt || len(d.Changes) != 0 || d.SnapshotStatus != "unavailable" || d.SnapshotReason != tc.state || d.ReasonCode != "" || d.Reason != "" {
				t.Fatalf("details=%#v called=%d", d, called)
			}
			raw, _ := json.Marshal(d)
			if strings.Contains(string(raw), "private") || strings.Contains(string(raw), "SQL") {
				t.Fatal("read error or before state leaked")
			}
		})
	}
}

func TestAuditSnapshotDiffLimitDiscardsWholeDiff(t *testing.T) {
	before := snapshotData{private: map[string]string{"settings.ONE": ""}}
	after := snapshotData{private: map[string]string{"settings.ONE": "secret"}, values: map[string]interface{}{"large": strings.Repeat("x", snapshotMaxDiffBytes+1)}}
	if changes, err := snapshotDiff(before, after); err == nil || len(changes) != 0 {
		t.Fatal("oversized diff must be discarded")
	}
}

func TestAuditSnapshotQueriesUseRealModelColumnsAndBoundedSQL(t *testing.T) {
	db, err := gorm.Open(mysql.New(mysql.Config{DSN: "unused:unused@tcp(localhost:1)/unused", SkipInitializeWithVersion: true}), &gorm.Config{DryRun: true, DisableAutomaticPing: true})
	if err != nil {
		t.Fatal(err)
	}
	models := []interface{}{&model.User{}, &model.UserRole{}, &model.UserDept{}, &model.UserReportingRelation{}, &model.Role{}, &model.Permission{}, &model.PermissionGrant{}, &model.Department{}, &model.Position{}, &model.SysDict{}, &model.SysDictType{}, &model.Survey{}, &model.Exam{}, &model.WorkflowDefinition{}, &model.WorkflowOrgApproverAssignment{}, &model.Setup{}, &model.DingTalkH5CorpConfig{}}
	tables := map[string]*schema.Schema{}
	for _, m := range models {
		s, err := schema.Parse(m, &sync.Map{}, schema.NamingStrategy{})
		if err != nil {
			t.Fatal(err)
		}
		tables[s.Table] = s
	}
	for _, family := range []string{"user", "role", "permission", "department", "position", "dict_item", "dict_type", "survey", "exam", "workflow_definition", "settings", "org_approver"} {
		p := snapshotPlan{spec: snapshotResource(family), ids: []string{"7"}, subjectType: "department", identity: "leader"}
		if family == "settings" {
			p.spec.mode = "dingtalk"
			p.corpConfigs = true
			p.ids = []string{"DINGTALK_H5_APP_SECRET"}
		}
		for _, q := range snapshotQueries(p) {
			t.Run(family+"/"+q.table, func(t *testing.T) {
				s := tables[q.table]
				if s == nil {
					t.Fatalf("unknown model table %s", q.table)
				}
				cols := append(append([]string{}, q.columns...), q.key)
				if q.privateColumn != "" {
					cols = append(cols, q.privateColumn)
				}
				for col := range q.where {
					cols = append(cols, col)
				}
				for _, col := range cols {
					if s.FieldsByDBName[col] == nil {
						t.Fatalf("non-model column %s.%s", q.table, col)
					}
				}
				var rows []map[string]interface{}
				stmt := snapshotSelect(db, q).Find(&rows).Statement
				sql := stmt.SQL.String()
				if strings.Contains(sql, "*") || !strings.Contains(sql, "LIMIT ?") || (q.table != "dingtalk_h5_corp_configs" && !strings.Contains(sql, "WHERE")) {
					t.Fatalf("unsafe SQL: %s", sql)
				}
				for _, forbidden := range []string{"user_password", "user_admin_token", "user_mobile", "user_forms", "survey_schema", "survey_settings", "exam_schema", "exam_settings", "definition_draft_json", "definition_bpmn_xml", "content_hash"} {
					if strings.Contains(sql, forbidden) {
						t.Fatalf("private column %s selected: %s", forbidden, sql)
					}
				}
				if q.privateColumn != "" && !strings.Contains(sql, "CASE WHEN OCTET_LENGTH(") {
					t.Fatal("private value must be length bounded in SQL")
				}
			})
		}
	}
}

func TestAuditSnapshotRelationsIgnoreSurrogateIDsAndRedactScope(t *testing.T) {
	p := snapshotPlan{spec: snapshotResource("role"), ids: []string{"7"}}
	q := snapshotQueries(p)[1]
	rows := []map[string]interface{}{{"id": 33, "grant_subject_id": 7, "grant_permission_key": "data:custom", "grant_effect": "allow", "grant_status": 1, "snapshot_private": `{"deptIds":[3,2,3],"userIds":[8],"password":"secret","nested":{"answer":"private"}}`}}
	a := snapshotData{values: map[string]interface{}{}, private: map[string]string{}}
	if err := snapshotRows(&a, p, q, rows, false); err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(a.values)
	if strings.Contains(string(raw), "secret") || strings.Contains(string(raw), "private") || strings.Contains(string(raw), `"id"`) {
		t.Fatalf("leaked relation: %s", raw)
	}
	rows[0]["id"] = 99
	rows[0]["snapshot_private"] = `{"userIds":[8],"deptIds":[2,3]}`
	b := snapshotData{values: map[string]interface{}{}, private: map[string]string{}}
	if err := snapshotRows(&b, p, q, rows, false); err != nil {
		t.Fatal(err)
	}
	if changes, err := snapshotDiff(a, b); err != nil || len(changes) != 0 {
		t.Fatalf("unchanged grants diff=%#v err=%v", changes, err)
	}
	rows[0]["grant_effect"] = "deny"
	if err := snapshotRows(&b, p, q, rows, false); err != nil {
		t.Fatal(err)
	}
	if changes, _ := snapshotDiff(a, b); len(changes) != 1 {
		t.Fatal("grant effect change missing")
	}
}

func TestAuditSnapshotDingTalkAllPrivateValuesAreMarkers(t *testing.T) {
	p := snapshotPlan{spec: snapshotResource("settings"), ids: []string{"DINGTALK_H5_APP_SECRET"}, corpConfigs: true}
	p.spec.mode = "dingtalk"
	queries := snapshotQueries(p)
	if len(queries) != 2 || queries[1].table != "dingtalk_h5_corp_configs" {
		t.Fatalf("multi-corp config missing: %#v", queries)
	}
	q := queries[1]
	rows := []map[string]interface{}{{"id": 7, "corp_id": "private-corp", "app_key": "private-appkey", "snapshot_private": "private-secret"}}
	a := snapshotData{values: map[string]interface{}{}, private: map[string]string{}}
	if err := snapshotRows(&a, p, q, rows, false); err != nil {
		t.Fatal(err)
	}
	changes, err := snapshotDiff(snapshotData{}, a)
	if err != nil || len(changes) == 0 {
		t.Fatalf("changes=%v err=%v", changes, err)
	}
	raw, _ := json.Marshal(changes)
	if strings.Contains(string(raw), "private-") {
		t.Fatalf("private config escaped: %s", raw)
	}
	for _, change := range changes {
		if change.Before != "[已隐藏]" || change.After != "[已变更]" {
			t.Fatalf("config contains non-marker: %#v", change)
		}
	}
}

func TestAuditSnapshotPermissionRenameKeepsStableRowIdentity(t *testing.T) {
	h := server.New()
	var recorder *auditlog.Recorder
	reads := 0
	h.Use(func(ctx context.Context, c *app.RequestContext) {
		ctx, recorder = auditlog.WithRecorder(ctx)
		c.Set("admin", &model.Admin{ID: 9})
		c.Next(ctx)
	})
	h.PUT("/api/v2/admin/permissions/:key", auditSnapshotWithReader(func(_ context.Context, p snapshotPlan) (snapshotData, error) {
		reads++
		key := "old"
		if reads == 2 {
			key = "new"
		}
		if !reflect.DeepEqual(p.ids, []string{key}) {
			t.Fatalf("rename query ids=%v want=%s", p.ids, key)
		}
		d := snapshotData{values: map[string]interface{}{}, private: map[string]string{}}
		q := snapshotQueries(p)[0]
		err := snapshotRows(&d, p, q, []map[string]interface{}{{"id": 7, "permission_key": key, "permission_name": "Same"}}, true)
		return d, err
	}), func(_ context.Context, c *app.RequestContext) { c.String(200, `{"code":0}`) })
	ut.PerformRequest(h.Engine, "PUT", "/api/v2/admin/permissions/old", &ut.Body{Body: strings.NewReader("key=new"), Len: 7}, ut.Header{Key: "Content-Type", Value: "application/x-www-form-urlencoded"})
	d := recorder.Snapshot()
	if len(d.Changes) != 1 || d.Changes[0].Field != "permission.7.permission_key" || d.Changes[0].Before != "old" || d.Changes[0].After != "new" {
		t.Fatalf("rename diff=%#v", d.Changes)
	}
}

func TestAuditSnapshotBodySelectorsAndLimits(t *testing.T) {
	for _, tc := range []struct {
		route, body, contentType, state string
		wantIDs                         []string
	}{
		{"/users", "ids=1,2,2", "application/x-www-form-urlencoded", "", []string{"1", "2"}},
		{"/users", "ids=1 OR 1=1", "application/x-www-form-urlencoded", "invalid_target", nil},
		{"/users", "ids=" + strings.Repeat("1,", snapshotMaxTargets) + "2", "application/x-www-form-urlencoded", "target_limit", nil},
		{"/settings", "key=ARBITRARY_SECRET&value=secret", "application/x-www-form-urlencoded", "unsupported_setup_key", nil},
		{"/settings", "key=STATIC_DOMAIN&value=secret", "multipart/form-data; boundary=x", "request_unavailable", nil},
		{"/settings", "key=STATIC_DOMAIN&value=" + strings.Repeat("x", snapshotMaxBodyBytes), "application/x-www-form-urlencoded", "request_unavailable", nil},
		{"/workflow-org-approver-assignments", `{"departmentId":7,"identityCode":"leader","userIds":[2]}`, "application/json", "", []string{"7"}},
	} {
		t.Run(tc.route+tc.state, func(t *testing.T) {
			c := app.NewContext(0)
			c.Request.Header.Set("Content-Type", tc.contentType)
			c.Request.SetBodyString(tc.body)
			method := "PUT"
			if tc.route == "/users" {
				method = "DELETE"
			}
			spec, ok := snapshotRoute(method, "/api/v2/admin"+tc.route)
			if !ok {
				t.Fatal("route missing")
			}
			p, state := snapshotRequestPlan(c, spec)
			if state != tc.state || (state == "" && !reflect.DeepEqual(p.ids, tc.wantIDs)) {
				t.Fatalf("state=%s ids=%v", state, p.ids)
			}
			body, _ := c.Body()
			if string(body) != tc.body {
				t.Fatal("request mutated")
			}
		})
	}
}

func TestAuditSnapshotNilDBContinuesAndReportsSafeState(t *testing.T) {
	previous := database.DB
	database.DB = nil
	t.Cleanup(func() { database.DB = previous })
	d, called := runSnapshotRequest(t, `{"code":0}`, 200, false, readAuditSnapshot)
	if called != 1 || d.SnapshotStatus != "unavailable" || d.SnapshotReason != "before_read_failed" || d.ReasonCode != "" || d.Reason != "" || len(d.Changes) != 0 {
		t.Fatalf("%+v", d)
	}
}

func TestAuditSnapshotRowAndRelationLimits(t *testing.T) {
	p := snapshotPlan{spec: snapshotResource("user"), ids: []string{"7"}}
	for _, q := range snapshotQueries(p) {
		d := snapshotData{values: map[string]interface{}{}, private: map[string]string{}}
		rows := make([]map[string]interface{}, q.limit+1)
		if !errors.Is(snapshotRows(&d, p, q, rows, q.table == "users"), errSnapshotLimit) {
			t.Fatalf("%s has no row limit", q.table)
		}
	}
	if _, err := snapshotScope(fmt.Sprintf(`{"deptIds":[%s]}`, strings.TrimSuffix(strings.Repeat("1,", 101), ","))); !errors.Is(err, errSnapshotLimit) {
		t.Fatal("scope ids unbounded")
	}
}

func TestAuditSnapshotSkipsUntrustedEntryAndPreservesHandlerContext(t *testing.T) {
	for _, authenticated := range []bool{false, true} {
		h := server.New()
		var recorder *auditlog.Recorder
		reads := 0
		h.Use(func(ctx context.Context, c *app.RequestContext) {
			ctx, recorder = auditlog.WithRecorder(ctx)
			if authenticated {
				c.Set("admin", &model.Admin{ID: 9})
			}
			c.Next(ctx)
		})
		h.PUT("/api/v2/admin/users/:id", auditSnapshotWithReader(func(_ context.Context, _ snapshotPlan) (snapshotData, error) {
			reads++
			return snapshotData{values: map[string]interface{}{"user.7.user_status": reads}}, nil
		}), func(ctx context.Context, c *app.RequestContext) {
			auditlog.Update(ctx, func(d *auditlog.Details) {
				d.TargetID, d.TargetType, d.TargetName = "handler-id", "handler-type", "handler-name"
				d.ReasonCode, d.Reason, d.Outcome, d.AffectedCount = "handler-reason", "reason", "accepted", 4
				d.Changes = []auditlog.Change{{Field: "user.7.user_status", Before: "handler-before", After: "handler-after"}}
			})
			c.String(200, `{"code":0}`)
		})
		ut.PerformRequest(h.Engine, "PUT", "/api/v2/admin/users/7", nil)
		d := recorder.Snapshot()
		if (!authenticated && reads != 0) || (authenticated && reads != 2) {
			t.Fatalf("auth=%v reads=%d", authenticated, reads)
		}
		if d.TargetID != "handler-id" || d.TargetName != "handler-name" || d.TargetType != "handler-type" || d.ReasonCode != "handler-reason" || d.Reason != "reason" || d.Outcome != "accepted" || d.AffectedCount != 4 || len(d.Changes) != 1 || d.Changes[0].After != "handler-after" {
			t.Fatalf("handler context overwritten: %+v", d)
		}
	}
}

func TestAuditSnapshotDingTalkScopeQueriesOnlyRequestedConfiguration(t *testing.T) {
	for _, scope := range []string{"login", "app", "corp", "all"} {
		c := app.NewContext(0)
		c.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		c.Request.SetBodyString("scope=" + scope)
		spec, _ := snapshotRoute("PUT", "/api/v2/admin/dingtalk/settings")
		p, state := snapshotRequestPlan(c, spec)
		if state != "" {
			t.Fatal(state)
		}
		queries := snapshotQueries(p)
		if scope == "app" || scope == "login" {
			if len(queries) != 1 || strings.Contains(strings.Join(p.ids, ","), "APP_SECRET") || strings.Contains(strings.Join(p.ids, ","), "APP_KEY") {
				t.Fatalf("unrelated corp config selected for %s: %+v", scope, p)
			}
		} else if len(queries) != 2 {
			t.Fatalf("corp config missing for %s", scope)
		}
	}
}

func TestAuditSnapshotPrivateOversizeIsNotMistakenForUnchanged(t *testing.T) {
	p := snapshotPlan{spec: snapshotResource("settings"), ids: []string{"DINGTALK_H5_APP_SECRET"}}
	q := snapshotQueries(p)[0]
	d := snapshotData{values: map[string]interface{}{}, private: map[string]string{}}
	rows := []map[string]interface{}{{"setup_key": "DINGTALK_H5_APP_SECRET", "snapshot_private": nil, "snapshot_oversize": 1}}
	if err := snapshotRows(&d, p, q, rows, true); !errors.Is(err, errSnapshotLimit) {
		t.Fatalf("oversize error=%v", err)
	}
	if len(d.private) != 0 || len(d.values) != 0 {
		t.Fatal("oversize projected as a value")
	}
}

func TestAuditSnapshotSettingsPublicValuesAndPrivateMarkers(t *testing.T) {
	for _, tc := range []struct {
		key           string
		before, after interface{}
		want          auditlog.Change
	}{
		{"ADMIN_LOGIN_CAPTCHA_ENABLED", []byte("0"), []byte("true"), auditlog.Change{Before: false, After: true}},
		{"DINGTALK_H5_LOGIN_CAPTCHA_ENABLED", true, false, auditlog.Change{Before: true, After: false}},
		{"USER_SINGLE_LOGIN", int64(0), int64(1), auditlog.Change{Before: false, After: true}},
		{"TOKEN_ADMIN_EXPIRE", []byte("60m"), "120m", auditlog.Change{Before: "1h0m0s", After: "2h0m0s"}},
		{"DINGTALK_H5_NOTIFY_MODE", "agent", []byte("robot"), auditlog.Change{Before: "agent", After: "robot"}},
		{"DINGTALK_H5_APP_SECRET", "private-old", "private-new", auditlog.Change{Before: "[已隐藏]", After: "[已变更]"}},
		{"STATIC_DOMAIN", "private-old", "private-new", auditlog.Change{Before: "[已隐藏]", After: "[已变更]"}},
	} {
		t.Run(tc.key, func(t *testing.T) {
			p := snapshotPlan{spec: snapshotResource("settings"), ids: []string{tc.key}}
			q := snapshotQueries(p)[0]
			read := func(value interface{}) snapshotData {
				data := snapshotData{values: map[string]interface{}{}, private: map[string]string{}}
				if err := snapshotRows(&data, p, q, []map[string]interface{}{{"setup_key": tc.key, "snapshot_private": value}}, true); err != nil {
					t.Fatal(err)
				}
				return data
			}
			got, err := snapshotDiff(read(tc.before), read(tc.after))
			tc.want.Field = "settings." + tc.key
			if err != nil || !reflect.DeepEqual(got, []auditlog.Change{tc.want}) {
				t.Fatalf("changes=%#v err=%v want=%#v", got, err, tc.want)
			}
		})
	}
}

func TestAuditSnapshotInvalidPublicValuesStayPrivate(t *testing.T) {
	for _, tc := range []struct {
		name, key string
		value     interface{}
	}{
		{"secret string", "ADMIN_LOGIN_CAPTCHA_ENABLED", "private-secret"},
		{"object", "USER_SINGLE_LOGIN", map[string]interface{}{"enabled": true, "secret": "private-secret"}},
		{"array", "USER_SINGLE_LOGIN", []interface{}{true}},
		{"invalid number", "USER_SINGLE_LOGIN", 2},
		{"negative duration", "TOKEN_USER_EXPIRE", "-1h"},
		{"zero duration", "TOKEN_USER_EXPIRE", "0s"},
		{"invalid enum", "DINGTALK_H5_NOTIFY_MODE", "private-secret"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := snapshotPlan{spec: snapshotResource("settings"), ids: []string{tc.key}}
			data := snapshotData{values: map[string]interface{}{}, private: map[string]string{}}
			if err := snapshotRows(&data, p, snapshotQueries(p)[0], []map[string]interface{}{{"setup_key": tc.key, "snapshot_private": tc.value}}, true); err != nil {
				t.Fatal(err)
			}
			field := "settings." + tc.key
			if _, exists := data.private[field]; !exists || len(data.values) != 0 {
				t.Fatal("invalid public value was not kept private")
			}
			public := snapshotData{values: map[string]interface{}{field: nil}}
			for _, snapshots := range [][2]snapshotData{{public, data}, {data, public}} {
				changes, err := snapshotDiff(snapshots[0], snapshots[1])
				want := []auditlog.Change{{Field: field, Before: "[已隐藏]", After: "[已变更]"}}
				if err != nil || !reflect.DeepEqual(changes, want) {
					t.Fatalf("private/public transition=%#v err=%v", changes, err)
				}
			}
		})
	}
}

func TestAuditSnapshotDingTalkMixedPublicFlagsAndSecrets(t *testing.T) {
	p := snapshotPlan{spec: snapshotResource("settings"), ids: []string{"DINGTALK_H5_NOTIFY_ENABLED", "DINGTALK_H5_APP_SECRET"}, corpConfigs: true}
	p.spec.mode = "dingtalk"
	queries := snapshotQueries(p)
	read := func(enabled int64, mode, secret string) snapshotData {
		data := snapshotData{values: map[string]interface{}{}, private: map[string]string{}}
		settings := []map[string]interface{}{
			{"setup_key": "DINGTALK_H5_NOTIFY_ENABLED", "snapshot_private": enabled},
			{"setup_key": "DINGTALK_H5_APP_SECRET", "snapshot_private": secret},
		}
		corps := []map[string]interface{}{{"id": 7, "enabled": enabled, "notify_enabled": enabled, "notify_mode": mode, "app_url": secret, "snapshot_private": secret}}
		if err := snapshotRows(&data, p, queries[0], settings, true); err != nil {
			t.Fatal(err)
		}
		if err := snapshotRows(&data, p, queries[1], corps, false); err != nil {
			t.Fatal(err)
		}
		return data
	}
	changes, err := snapshotDiff(read(0, "agent", "private-old"), read(1, "agent_fallback", "private-new"))
	want := []auditlog.Change{
		{Field: "settings.DINGTALK_H5_APP_SECRET", Before: "[已隐藏]", After: "[已变更]"},
		{Field: "settings.DINGTALK_H5_NOTIFY_ENABLED", Before: false, After: true},
		{Field: "settings.corp.7.app_secret", Before: "[已隐藏]", After: "[已变更]"},
		{Field: "settings.corp.7.app_url", Before: "[已隐藏]", After: "[已变更]"},
		{Field: "settings.corp.7.enabled", Before: false, After: true},
		{Field: "settings.corp.7.notify_enabled", Before: false, After: true},
		{Field: "settings.corp.7.notify_mode", Before: "agent", After: "agent_fallback"},
	}
	if err != nil || !reflect.DeepEqual(changes, want) {
		t.Fatalf("changes=%#v err=%v", changes, err)
	}
}

func TestAuditSnapshotDingTalkInvalidFlagsStayPrivate(t *testing.T) {
	p := snapshotPlan{spec: snapshotResource("settings"), corpConfigs: true}
	q := snapshotQueries(p)[1]
	read := func(enabled, notifyEnabled, notifyMode interface{}) snapshotData {
		data := snapshotData{values: map[string]interface{}{}, private: map[string]string{}}
		row := map[string]interface{}{"id": 7, "enabled": enabled, "notify_enabled": notifyEnabled, "notify_mode": notifyMode}
		if err := snapshotRows(&data, p, q, []map[string]interface{}{row}, false); err != nil {
			t.Fatal(err)
		}
		return data
	}
	before := read(false, 0, "agent")
	after := read(map[string]interface{}{"secret": "private-secret"}, []interface{}{true}, "private-secret")
	changes, err := snapshotDiff(before, after)
	want := []auditlog.Change{
		{Field: "settings.corp.7.enabled", Before: "[已隐藏]", After: "[已变更]"},
		{Field: "settings.corp.7.notify_enabled", Before: "[已隐藏]", After: "[已变更]"},
		{Field: "settings.corp.7.notify_mode", Before: "[已隐藏]", After: "[已变更]"},
	}
	if err != nil || !reflect.DeepEqual(changes, want) {
		t.Fatalf("invalid corp flag changes=%#v err=%v", changes, err)
	}
}

func TestAuditSnapshotComparisonStatus(t *testing.T) {
	for _, tc := range []struct {
		name, status, reason string
		before, after        interface{}
		wantChanges          int
	}{
		{"captcha toggled", "recorded", "", "false", "true", 1},
		{"normalized unchanged", "unchanged", "no_collected_changes", "1", "true", 0},
		{"private unchanged", "unchanged", "no_collected_changes", "private-secret", "private-secret", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := server.New()
			var recorder *auditlog.Recorder
			h.Use(func(ctx context.Context, c *app.RequestContext) {
				ctx, recorder = auditlog.WithRecorder(ctx)
				auditlog.Update(ctx, func(d *auditlog.Details) {
					d.SnapshotStatus, d.SnapshotReason = "pending", "stale_reason"
				})
				c.Set("admin", &model.Admin{ID: 9})
				c.Next(ctx)
			})
			reads := 0
			h.PUT("/api/v2/admin/settings", auditSnapshotWithReader(func(_ context.Context, p snapshotPlan) (snapshotData, error) {
				reads++
				if p.spec.family != "settings" || !reflect.DeepEqual(p.ids, []string{"ADMIN_LOGIN_CAPTCHA_ENABLED"}) {
					t.Fatalf("settings request plan=%#v", p)
				}
				value := tc.before
				if reads == 2 {
					value = tc.after
				}
				data := snapshotData{values: map[string]interface{}{}, private: map[string]string{}}
				err := snapshotRows(&data, p, snapshotQueries(p)[0], []map[string]interface{}{{"setup_key": p.ids[0], "snapshot_private": value}}, true)
				return data, err
			}), func(ctx context.Context, c *app.RequestContext) {
				auditlog.Update(ctx, func(d *auditlog.Details) {
					d.Outcome, d.ReasonCode, d.Reason = "success", "saved", "业务保存成功"
				})
				c.String(200, `{"code":0}`)
			})
			body := "key=ADMIN_LOGIN_CAPTCHA_ENABLED&value=true"
			result := ut.PerformRequest(h.Engine, "PUT", "/api/v2/admin/settings", &ut.Body{Body: strings.NewReader(body), Len: len(body)}, ut.Header{Key: "Content-Type", Value: "application/x-www-form-urlencoded"}).Result()
			d := recorder.Snapshot()
			if result.StatusCode() != 200 || string(result.Body()) != `{"code":0}` || d.Outcome != "success" || d.ReasonCode != "saved" || d.Reason != "业务保存成功" {
				t.Fatal("snapshot collection changed the business response or reason")
			}
			if reads != 2 || d.SnapshotStatus != tc.status || d.SnapshotReason != tc.reason || len(d.Changes) != tc.wantChanges {
				t.Fatalf("reads=%d details=%#v", reads, d)
			}
			if tc.wantChanges == 1 && (d.Changes[0].Before != false || d.Changes[0].After != true) {
				t.Fatalf("captcha boolean change=%#v", d.Changes)
			}
		})
	}
}

func TestAuditSnapshotUnknownSettingsRemainUncollected(t *testing.T) {
	for _, response := range []string{`{"code":0}`, `{"code":1}`} {
		t.Run(response, func(t *testing.T) {
			h := server.New()
			var recorder *auditlog.Recorder
			h.Use(func(ctx context.Context, c *app.RequestContext) {
				ctx, recorder = auditlog.WithRecorder(ctx)
				c.Set("admin", &model.Admin{ID: 9})
				c.Next(ctx)
			})
			body := "key=UNKNOWN_ENABLED&value=private-secret"
			called := 0
			h.PUT("/api/v2/admin/settings", auditSnapshotWithReader(func(_ context.Context, _ snapshotPlan) (snapshotData, error) {
				t.Fatal("unsupported setting must not trigger a database read")
				return snapshotData{}, nil
			}), func(ctx context.Context, c *app.RequestContext) {
				called++
				if got, err := c.Body(); err != nil || string(got) != body {
					t.Fatal("collector changed the request body")
				}
				auditlog.SetReason(ctx, "saved", "业务说明")
				c.String(200, response)
			})
			result := ut.PerformRequest(h.Engine, "PUT", "/api/v2/admin/settings", &ut.Body{Body: strings.NewReader(body), Len: len(body)}, ut.Header{Key: "Content-Type", Value: "application/x-www-form-urlencoded"}).Result()
			d := recorder.Snapshot()
			if called != 1 || result.StatusCode() != 200 || string(result.Body()) != response || len(d.Changes) != 0 || d.ReasonCode != "saved" || d.Reason != "业务说明" {
				t.Fatalf("unsupported key changed business handling: %#v", d)
			}
			if response == `{"code":0}` {
				if d.SnapshotStatus != "not_configured" || d.SnapshotReason != "unsupported_setup_key" {
					t.Fatalf("unsupported key status=%q reason=%q", d.SnapshotStatus, d.SnapshotReason)
				}
			} else if d.SnapshotStatus != "" || d.SnapshotReason != "" {
				t.Fatal("collector finalized failed operation")
			}
		})
	}
}

func TestAuditSnapshotDiffLimitReportsUnavailable(t *testing.T) {
	reads := 0
	d, called := runSnapshotRequest(t, `{"code":0}`, 200, false, func(_ context.Context, _ snapshotPlan) (snapshotData, error) {
		reads++
		data := snapshotData{values: map[string]interface{}{}}
		if reads == 2 {
			for i := 0; i <= snapshotMaxChanges; i++ {
				data.values[fmt.Sprintf("user.%d.user_status", i)] = "1"
			}
		}
		return data, nil
	})
	if called != 1 || reads != 2 || d.SnapshotStatus != "unavailable" || d.SnapshotReason != "diff_limit" || len(d.Changes) != 0 || d.ReasonCode != "" || d.Reason != "" {
		t.Fatalf("diff limit details=%#v", d)
	}
}

func TestAuditSnapshotUnavailablePreservesBusinessReason(t *testing.T) {
	for _, tc := range []struct{ state, status string }{
		{"unsupported_setup_key", "not_configured"},
		{"request_unavailable", "unavailable"},
		{"invalid_target", "unavailable"},
		{"target_limit", "unavailable"},
		{"before_read_failed", "unavailable"},
		{"after_read_failed", "unavailable"},
		{"before_limit", "unavailable"},
		{"after_limit", "unavailable"},
		{"diff_limit", "unavailable"},
	} {
		t.Run(tc.state, func(t *testing.T) {
			ctx, recorder := auditlog.WithRecorder(context.Background())
			auditlog.Update(ctx, func(d *auditlog.Details) {
				d.Outcome, d.ReasonCode, d.Reason = "success", "business_reason", "业务说明"
			})
			c := app.NewContext(0)
			snapshotUnavailable(ctx, c, tc.state)
			d := recorder.Snapshot()
			if d.SnapshotStatus != tc.status || d.SnapshotReason != tc.state || d.Outcome != "success" || d.ReasonCode != "business_reason" || d.Reason != "业务说明" || len(d.Changes) != 0 {
				t.Fatalf("details=%#v", d)
			}
		})
	}
}
