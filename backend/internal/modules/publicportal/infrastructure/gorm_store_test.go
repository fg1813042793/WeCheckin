package infrastructure

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	mysql "github.com/go-sql-driver/mysql"

	pm "wecheckin/backend/internal/model/publicportal"
	"wecheckin/backend/internal/modules/publicportal/application"
)

func sqlContains(t *testing.T, parts ...string) func(string, []driver.NamedValue) {
	return func(q string, _ []driver.NamedValue) {
		t.Helper()
		for _, part := range parts {
			if !strings.Contains(q, part) {
				t.Errorf("SQL missing %q: %s", part, q)
			}
		}
	}
}
func TestSchemaMissingAndDefaults(t *testing.T) {
	store, _ := openScript(t, scriptStep{query: true, err: &mysql.MySQLError{Number: 1146, Message: "secret database table missing"}})
	if _, err := store.Settings(context.Background()); !errors.Is(err, application.ErrSchemaMissing) {
		t.Fatalf("schema=%v", err)
	}
	store, _ = openScript(t, scriptStep{query: true, columns: []string{"id"}})
	got, err := store.Settings(context.Background())
	if err != nil || got.ID != 1 || got.Enabled || got.Title != "公开问卷与考试" || got.Version != 1 || got.PublicBaseURL != "" {
		t.Fatalf("defaults=%+v %v", got, err)
	}
}
func TestTransactionLocksSingletonAndCASPreservesFalse(t *testing.T) {
	store, script := openScript(t,
		scriptStep{check: sqlContains(t, "INSERT INTO `public_portal_settings`", "ON DUPLICATE KEY UPDATE"), affected: 1},
		scriptStep{query: true, check: sqlContains(t, "FOR UPDATE"), columns: []string{"id", "version"}, rows: [][]driver.Value{{int64(1), int64(3)}}},
		scriptStep{check: func(q string, args []driver.NamedValue) {
			sqlContains(t, "UPDATE `public_portal_settings`", "`enabled`=?", "`show_ended`=?", "version = ?")(q, args)
			found := false
			for _, a := range args {
				if a.Value == false {
					found = true
				}
			}
			if !found {
				t.Fatal("false flags omitted")
			}
		}, affected: 1},
	)
	err := store.Transaction(context.Background(), func(tx application.Store) error {
		settings, e := tx.LockSettings(context.Background())
		if e != nil {
			return e
		}
		settings.Version++
		settings.Enabled = false
		return tx.SaveSettings(context.Background(), settings, 3)
	})
	if err != nil || script.begins != 1 || script.commits != 1 || script.rollbacks != 0 {
		t.Fatalf("tx=%+v %v", script, err)
	}
}
func TestCASConflictAndWriteErrorRollback(t *testing.T) {
	for _, writeErr := range []error{nil, &mysql.MySQLError{Number: 1054, Message: "unknown column"}, errors.New("driver failed")} {
		store, script := openScript(t, scriptStep{check: sqlContains(t, "UPDATE `public_portal_items`", "version = ?"), affected: 0, err: writeErr})
		err := store.Transaction(context.Background(), func(tx application.Store) error {
			return tx.UpdateItem(context.Background(), pm.Item{ID: 4, Version: 2}, 1)
		})
		if err == nil || script.commits != 0 || script.rollbacks != 1 {
			t.Fatalf("rollback=%+v err=%v", script, err)
		}
		if writeErr == nil && !errors.Is(err, application.ErrVersionConflict) {
			t.Fatalf("conflict=%v", err)
		}
	}
}
func TestSourcesUseProjectionAndExtractOnlyFlags(t *testing.T) {
	for _, kind := range []application.Kind{application.Survey, application.Exam} {
		prefix := string(kind)
		store, _ := openScript(t, scriptStep{query: true, check: func(q string, _ []driver.NamedValue) {
			if strings.Contains(q, "schema") || strings.Contains(q, "paper_id") || strings.Contains(q, "SELECT *") {
				t.Fatalf("unsafe projection: %s", q)
			}
			if !strings.Contains(q, prefix+"_settings") {
				t.Fatal("settings not selected")
			}
		}, columns: []string{prefix + "_id", prefix + "_title", prefix + "_settings", prefix + "_status", prefix + "_visibility"}, rows: [][]driver.Value{{int64(9), "管理名称", `{"displayTitle":"公开标题","password":"secret","loginRequired":true,"fillTemplate":"ef1","schema":"private"}`, int64(1), int64(0)}}})
		rows, err := store.Sources(context.Background(), kind, []uint{9}, 0)
		if err != nil || len(rows) != 1 {
			t.Fatalf("rows=%+v %v", rows, err)
		}
		got := rows[0]
		if got.DisplayTitle != "公开标题" || !got.PasswordRequired || !got.SettingsValid || got.LoginRequired != (kind == application.Exam) {
			t.Fatalf("source=%+v", got)
		}
	}
}
func TestSourceSettingsMalformedFailClosed(t *testing.T) {
	for _, raw := range []string{"{bad", `{"password":{}}`, `{"displayTitle":[]}`} {
		s := parseSourceSettings(application.Source{Kind: application.Exam}, raw)
		if s.SettingsValid {
			t.Fatalf("accepted malformed settings: %s", raw)
		}
	}
	s := parseSourceSettings(application.Source{Kind: application.Survey}, "")
	if !s.SettingsValid || s.LoginRequired || s.PasswordRequired {
		t.Fatalf("defaults=%+v", s)
	}
}

func TestScopeDeniedReadsOnlySourceIDOutsideScope(t *testing.T) {
	store, script := openScript(t)
	readIDs, readScoped := false, false
	script.query = func(q string, args []driver.NamedValue) scriptStep {
		switch {
		case strings.Contains(q, "FROM `survey`"):
			if strings.HasPrefix(q, "SELECT `survey_id` ") {
				readIDs = true
				return scriptStep{columns: []string{"survey_id"}, rows: [][]driver.Value{{int64(9)}}}
			}
			if !strings.Contains(q, "1 = 0") {
				t.Errorf("missing fail-closed source scope: %s", q)
			}
			if strings.Contains(q, "schema") {
				t.Error("selected schema")
			}
			readScoped = true
			return scriptStep{columns: []string{"survey_id"}}
		case strings.Contains(q, "FROM `users`"):
			return scriptStep{columns: []string{"id", "user_role_id"}, rows: [][]driver.Value{{int64(7), int64(0)}}}
		case strings.Contains(q, "DATABASE()"):
			return scriptStep{columns: []string{"database"}, rows: [][]driver.Value{{"portal_test"}}}
		case strings.Contains(strings.ToLower(q), "information_schema"):
			return scriptStep{columns: []string{"count"}, rows: [][]driver.Value{{int64(0)}}}
		default:
			t.Errorf("unexpected scope SQL: %s", q)
			return scriptStep{err: errors.New("unexpected SQL")}
		}
	}
	rows, err := store.Sources(context.Background(), application.Survey, []uint{9}, 7)
	if err != nil || len(rows) != 1 || !readIDs || !readScoped {
		t.Fatalf("rows=%+v err=%v", rows, err)
	}
	if rows[0] != (application.Source{Kind: application.Survey, ID: 9}) {
		t.Fatalf("unscoped data leaked: %+v", rows[0])
	}
}
func TestOptionsScopePublicationAndPaginationUseBoundSQL(t *testing.T) {
	store, script := openScript(t,
		scriptStep{query: true, columns: []string{"id", "user_role_id"}, rows: [][]driver.Value{{int64(7), int64(830001)}}},
		scriptStep{query: true, columns: []string{"id", "role_name", "role_status"}, rows: [][]driver.Value{{int64(830001), "超级管理员", int64(1)}}},
		scriptStep{query: true, check: sqlContains(t, "exam_status = ? AND exam_visibility = ?", "exam_title LIKE ?"), columns: []string{"count"}, rows: [][]driver.Value{{int64(3)}}},
		scriptStep{query: true, check: func(q string, args []driver.NamedValue) {
			sqlContains(t, "exam_status = ? AND exam_visibility = ?", "ORDER BY exam_id ASC", "LIMIT ? OFFSET ?")(q, args)
			if strings.Contains(q, "schema") || strings.Contains(q, "SELECT *") || strings.Contains(q, "OR 1=1") {
				t.Errorf("unsafe candidate SQL: %s", q)
			}
		}, columns: []string{"exam_id", "exam_title", "exam_settings", "exam_status", "exam_visibility"}, rows: [][]driver.Value{{int64(9), "Exam", `{"displayTitle":"Public","password":"secret"}`, int64(1), int64(0)}}},
	)
	_ = script
	got, err := store.Options(context.Background(), 7, application.OptionsQuery{Kind: application.Exam, Keyword: "' OR 1=1 --", Page: 2, PageSize: 1})
	if err != nil || got.Total != 3 || len(got.List) != 1 || !got.List[0].CanManage || got.List[0].DisplayTitle != "Public" {
		t.Fatalf("options=%+v err=%v", got, err)
	}
}
func TestItemWritesRequireTransactionAndPreserveAuditMilliseconds(t *testing.T) {
	now := int64(1793678000123)
	store, script := openScript(t,
		scriptStep{check: func(q string, args []driver.NamedValue) {
			sqlContains(t, "INSERT INTO `public_portal_items`", "`created_at`", "`updated_at`")(q, args)
			count := 0
			for _, arg := range args {
				if arg.Value == now {
					count++
				}
			}
			if count != 2 {
				t.Fatalf("timestamps not milliseconds: %+v", args)
			}
		}, affected: 1},
		scriptStep{check: sqlContains(t, "DELETE FROM `public_portal_items`", "id = ? AND version = ?"), affected: 1},
	)
	item := pm.Item{Kind: "survey", SourceID: 10, Version: 1, CreatedAt: now, UpdatedAt: now}
	if err := store.InsertItem(context.Background(), &item); !errors.Is(err, errTransactionRequired) {
		t.Fatalf("untransactional write: %v", err)
	}
	err := store.Transaction(context.Background(), func(tx application.Store) error {
		if err := tx.InsertItem(context.Background(), &item); err != nil {
			return err
		}
		if item.ID != 12 {
			t.Fatalf("generated id=%d", item.ID)
		}
		return tx.DeleteItem(context.Background(), item.ID, 1)
	})
	if err != nil || script.commits != 1 {
		t.Fatalf("write=%v script=%+v", err, script)
	}
}
func TestDuplicateSourceOnlyMapsItsOwnUniqueIndex(t *testing.T) {
	for _, index := range []string{"uk_public_portal_source", "PRIMARY"} {
		original := &mysql.MySQLError{Number: 1062, Message: "Duplicate entry for key '" + index + "'"}
		store, script := openScript(t, scriptStep{err: original})
		err := store.Transaction(context.Background(), func(tx application.Store) error {
			return tx.InsertItem(context.Background(), &pm.Item{Kind: "survey", SourceID: 10})
		})
		if index == "uk_public_portal_source" {
			if !errors.Is(err, application.ErrDuplicateSource) {
				t.Fatalf("duplicate=%v", err)
			}
		} else if !errors.Is(err, original) {
			t.Fatalf("masked unknown DB error: %v", err)
		}
		if script.rollbacks != 1 {
			t.Fatal("duplicate did not roll back")
		}
	}
}
func TestItemsUseCurrentLockingReadAfterSingletonAndStableOrder(t *testing.T) {
	store, script := openScript(t, scriptStep{query: true, check: sqlContains(t, "ORDER BY sort ASC,id ASC", "LIMIT ?", "FOR UPDATE"), columns: []string{"id"}})
	err := store.Transaction(context.Background(), func(tx application.Store) error { _, err := tx.Items(context.Background()); return err })
	if err != nil || script.commits != 1 {
		t.Fatalf("locked read=%v", err)
	}
}
func TestActorRejectsInactiveAndCancellationPropagates(t *testing.T) {
	store, _ := openScript(t, scriptStep{query: true, check: sqlContains(t, "user_status = 1"), columns: []string{"id"}})
	if _, err := store.Actor(context.Background(), 7); !errors.Is(err, application.ErrForbidden) {
		t.Fatalf("inactive actor: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	store, _ = openScript(t)
	if _, err := store.Settings(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation=%v", err)
	}
	if _, err := NewGormStore(nil).Settings(context.Background()); err == nil {
		t.Fatal("nil database accepted")
	}
}

func TestMutationSourceExistenceUsesCurrentRead(t *testing.T) {
	store, _ := openScript(t,
		scriptStep{query: true, check: sqlContains(t, "SELECT `survey_id`", "FOR SHARE"), columns: []string{"survey_id"}, rows: [][]driver.Value{{int64(9)}}},
		scriptStep{query: true, columns: []string{"id", "user_role_id"}, rows: [][]driver.Value{{int64(7), int64(830002)}}},
		scriptStep{query: true, columns: []string{"id", "role_name", "role_status"}, rows: [][]driver.Value{{int64(830002), "超级管理员", int64(1)}}},
		scriptStep{query: true, check: sqlContains(t, "FOR SHARE"), columns: []string{"survey_id", "survey_status", "survey_visibility"}, rows: [][]driver.Value{{int64(9), int64(1), int64(0)}}},
	)
	err := store.Transaction(context.Background(), func(tx application.Store) error {
		_, err := tx.Sources(context.Background(), application.Survey, []uint{9}, 7)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestOptionsSearchesManagementAndStringDisplayNamesWithinScope(t *testing.T) {
	for _, kind := range []application.Kind{application.Survey, application.Exam} {
		for _, keyword := range []string{"管理名称", "显示名称"} {
			t.Run(string(kind)+keyword, func(t *testing.T) {
				store, script := openScript(t)
				prefix := string(kind)
				filtered := 0
				script.query = func(q string, args []driver.NamedValue) scriptStep {
					switch {
					case strings.Contains(q, "FROM `"+prefix+"`"):
						filtered++
						sqlContains(t, "1 = 0", prefix+"_status = ? AND "+prefix+"_visibility = ?", "("+prefix+"_title LIKE ? OR", "CASE WHEN JSON_VALID("+prefix+"_settings)", "JSON_TYPE(JSON_EXTRACT(", "'$.displayTitle'", "= 'STRING'", "JSON_UNQUOTE(")(q, args)
						matches := 0
						for _, arg := range args {
							if arg.Value == "%"+keyword+"%" {
								matches++
							}
						}
						if matches != 2 {
							t.Errorf("expected both names parameterized: %+v", args)
						}
						if strings.Contains(q, "count(*)") {
							return scriptStep{columns: []string{"count"}, rows: [][]driver.Value{{int64(0)}}}
						}
						return scriptStep{columns: []string{prefix + "_id"}}
					case strings.Contains(q, "FROM `users`"):
						return scriptStep{columns: []string{"id", "user_role_id"}, rows: [][]driver.Value{{int64(7), int64(0)}}}
					case strings.Contains(q, "DATABASE()"):
						return scriptStep{columns: []string{"database"}, rows: [][]driver.Value{{"portal_test"}}}
					case strings.Contains(strings.ToLower(q), "information_schema"):
						return scriptStep{columns: []string{"count"}, rows: [][]driver.Value{{int64(0)}}}
					default:
						t.Errorf("unexpected query: %s", q)
						return scriptStep{err: errors.New("unexpected SQL")}
					}
				}
				got, err := store.Options(context.Background(), 7, application.OptionsQuery{Kind: kind, Keyword: keyword, Page: 1, PageSize: 20})
				if err != nil || got.Total != 0 || len(got.List) != 0 || filtered != 2 {
					t.Fatalf("options=%+v err=%v", got, err)
				}
			})
		}
	}
}
func TestSourceDescriptionIsPlainTextWithoutExecutableContent(t *testing.T) {
	for _, kind := range []application.Kind{application.Survey, application.Exam} {
		prefix := string(kind)
		store, _ := openScript(t, scriptStep{query: true, columns: []string{prefix + "_id", prefix + "_desc", prefix + "_status"}, rows: [][]driver.Value{{int64(1), `<p>欢迎 &amp; 参与<strong>问卷</strong></p><script>alert('secret')</script><style>.secret{}</style><template>hidden</template><p>下一行<br>说明</p>`, int64(1)}}})
		rows, err := store.Sources(context.Background(), kind, []uint{1}, 0)
		if err != nil || len(rows) != 1 || rows[0].Description != "欢迎 & 参与问卷 下一行 说明" {
			t.Fatalf("description=%+v err=%v", rows, err)
		}
	}
}

func TestSourceProjectionReadsCategory(t *testing.T) {
	for _, kind := range []application.Kind{application.Survey, application.Exam} {
		prefix := string(kind)
		store, _ := openScript(t, scriptStep{query: true, check: sqlContains(t, "`"+prefix+"_category`"), columns: []string{prefix + "_id", prefix + "_category", prefix + "_status"}, rows: [][]driver.Value{{int64(1), " 面试题 ", int64(1)}}})
		rows, err := store.Sources(context.Background(), kind, []uint{1}, 0)
		raw, _ := json.Marshal(rows)
		if err != nil || len(rows) != 1 || !strings.Contains(string(raw), `"Category":"面试题"`) {
			t.Fatalf("category projection = %s err=%v", raw, err)
		}
	}
}
