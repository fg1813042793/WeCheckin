package infrastructure

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	userfeedbackmodel "wecheckin/backend/internal/model/userfeedback"
	"wecheckin/backend/internal/modules/userfeedback/application"
	"wecheckin/backend/internal/modules/userfeedback/domain"
)

func feedbackDriverRow(id uint64) []driver.Value {
	return []driver.Value{
		int64(id), fmt.Sprintf("FB-20260905-%04d", id), int64(7), fmt.Sprintf("create-%d", id), "processing",
		int64(3), int64(2), int64(2000 + id), nil, nil, int64(1000 + id), int64(2000 + id),
	}
}

func feedbackColumns() []string {
	return []string{
		"id", "feedback_no", "submitter_id", "create_request_id", "feedback_status", "handler_id", "version",
		"last_activity_at", "resolved_at", "closed_at", "created_at", "updated_at",
	}
}

func listQueryScript(size int) *databaseScript {
	return &databaseScript{query: func(query string, _ []driver.NamedValue) queryResult {
		sqlText := normalizeSQL(query)
		switch {
		case strings.Contains(strings.ToLower(sqlText), "count(*)") && strings.Contains(sqlText, "FROM `user_feedbacks`"):
			return queryResult{Columns: []string{"count"}, Rows: [][]driver.Value{{int64(size)}}}
		case strings.Contains(sqlText, "FROM `user_feedbacks`"):
			rows := make([][]driver.Value, 0, size)
			for index := 1; index <= size; index++ {
				rows = append(rows, feedbackDriverRow(uint64(index)))
			}
			return queryResult{Columns: feedbackColumns(), Rows: rows}
		case strings.Contains(sqlText, "FROM `user_feedback_messages`"):
			rows := make([][]driver.Value, 0, size)
			for index := 1; index <= size; index++ {
				rows = append(rows, []driver.Value{int64(index), fmt.Sprintf("  第%d条\n  初始   内容  ", index)})
			}
			return queryResult{Columns: []string{"feedback_id", "content"}, Rows: rows}
		case strings.Contains(sqlText, "FROM `user_feedback_attachments`"):
			rows := make([][]driver.Value, 0, size)
			for index := 1; index <= size; index++ {
				rows = append(rows, []driver.Value{int64(index), int64(index + 1)})
			}
			return queryResult{Columns: []string{"feedback_id", "image_count"}, Rows: rows}
		case strings.Contains(sqlText, "FROM `users`"):
			return queryResult{Columns: []string{"id", "user_name"}, Rows: [][]driver.Value{{int64(3), "管理员"}, {int64(7), "提交人"}}}
		default:
			return queryResult{Err: fmt.Errorf("unexpected query: %s", query)}
		}
	}}
}

func TestCompactSummaryIsUTF8SafeAndCollapsesWhitespace(t *testing.T) {
	if got, want := compactSummary(" \n 用户\t反馈  内容 \r\n "), "用户 反馈 内容"; got != want {
		t.Fatalf("compactSummary = %q, want %q", got, want)
	}
	input := strings.Repeat("界", 101)
	got := compactSummary(input)
	if len([]rune(got)) != 100 || !strings.HasSuffix(got, "界") {
		t.Fatalf("UTF-8 summary length/content = %d, %q", len([]rune(got)), got)
	}
}

func TestDetailMapperKeepsObjectKeysAndPassesContextToURLBuilder(t *testing.T) {
	ctxKey := struct{}{}
	ctx := context.WithValue(context.Background(), ctxKey, "request-context")
	handlerID := uint(3)
	feedback := userfeedbackmodel.Feedback{
		ID: 9, FeedbackNo: "FB-9", SubmitterID: 7, Status: "processing", HandlerID: &handlerID,
		Version: 2, LastActivityAt: 20, CreatedAt: 10, UpdatedAt: 20,
	}
	messages := []userfeedbackmodel.Message{
		{ID: 21, FeedbackID: 9, MessageType: "initial", AuthorType: "user", AuthorID: 7, Content: "初始", CreatedAt: 10},
		{ID: 22, FeedbackID: 9, MessageType: "supplement", AuthorType: "user", AuthorID: 7, Content: "补充", CreatedAt: 20},
	}
	attachments := []userfeedbackmodel.Attachment{
		{ID: 31, FeedbackID: 9, MessageID: 21, ObjectKey: "feedback/a.jpg", OriginalName: "a.jpg", SortOrder: 0},
		{ID: 32, FeedbackID: 9, MessageID: 22, ObjectKey: "feedback/b.png", OriginalName: "b.png", SortOrder: 0},
	}
	var paths []string
	detail := buildFeedbackDetailWithURL(ctx, feedback, messages, attachments, map[uint]string{3: "管理员", 7: "提交人"}, func(gotCtx context.Context, path string) string {
		if gotCtx.Value(ctxKey) != "request-context" {
			t.Fatalf("URL builder context value missing")
		}
		paths = append(paths, path)
		return "https://static.example" + path
	})
	if detail == nil || len(detail.Messages) != 2 || len(detail.Messages[0].Attachments) != 1 || len(detail.Messages[1].Attachments) != 1 {
		t.Fatalf("detail mapping = %#v", detail)
	}
	if got := detail.Messages[0].Attachments[0]; got.ObjectKey != "feedback/a.jpg" || got.URL != "https://static.example/feedback/a.jpg" {
		t.Fatalf("first attachment = %#v", got)
	}
	if got, want := paths, []string{"/feedback/a.jpg", "/feedback/b.png"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("URL paths = %#v, want %#v", got, want)
	}
	if detail.SubmitterName != "提交人" || detail.HandlerName != "管理员" || detail.Messages[0].AuthorName != "提交人" {
		t.Fatalf("display names = %#v", detail)
	}
}

func replayQueryScript() *databaseScript {
	return &databaseScript{query: func(query string, _ []driver.NamedValue) queryResult {
		sqlText := normalizeSQL(query)
		switch {
		case strings.Contains(sqlText, "FROM `user_feedback_messages`") && strings.Contains(sqlText, "`request_id` = ?") && strings.Contains(sqlText, "LIMIT ?"):
			return queryResult{Columns: []string{"feedback_id"}, Rows: [][]driver.Value{{int64(9)}}}
		case strings.Contains(sqlText, "FROM `user_feedbacks`"):
			return queryResult{Columns: feedbackColumns(), Rows: [][]driver.Value{feedbackDriverRow(9)}}
		case strings.Contains(sqlText, "FROM `user_feedback_messages`"):
			return queryResult{Columns: []string{"id", "feedback_id", "message_type", "author_type", "author_id", "content", "from_status", "to_status", "request_id", "created_at"}, Rows: [][]driver.Value{
				{int64(21), int64(9), "initial", "user", int64(7), "初始", "", "", "create-9", int64(10)},
				{int64(22), int64(9), "supplement", "user", int64(7), "补充", "", "", "supplement-1", int64(20)},
			}}
		case strings.Contains(sqlText, "FROM `user_feedback_attachments`"):
			return queryResult{Columns: []string{"id", "feedback_id", "message_id", "storage_provider", "object_key", "original_name", "content_type", "size_bytes", "sort_order", "created_at"}, Rows: [][]driver.Value{
				{int64(31), int64(9), int64(21), "local", "feedback/a.jpg", "a.jpg", "image/jpeg", int64(10), int64(0), int64(10)},
				{int64(32), int64(9), int64(22), "local", "feedback/b.png", "b.png", "image/png", int64(20), int64(0), int64(20)},
			}}
		case strings.Contains(sqlText, "FROM `users`"):
			return queryResult{Columns: []string{"id", "user_name"}, Rows: [][]driver.Value{{int64(3), "管理员"}, {int64(7), "提交人"}}}
		default:
			return queryResult{Err: fmt.Errorf("unexpected replay query: %s", query)}
		}
	}}
}

func TestReplayQueriesReturnCompleteDetailWithEveryObjectKey(t *testing.T) {
	tests := []struct {
		name string
		find func(*GormStore) (*application.FeedbackDetail, bool, error)
	}{
		{name: "create", find: func(store *GormStore) (*application.FeedbackDetail, bool, error) {
			return store.FindCreateReplay(context.Background(), application.CreateReplayKey{SubmitterID: 7, RequestID: "create-9"})
		}},
		{name: "message", find: func(store *GormStore) (*application.FeedbackDetail, bool, error) {
			return store.FindMessageReplay(context.Background(), application.MessageReplayKey{FeedbackID: 9, AuthorType: domain.AuthorTypeUser, AuthorID: 7, RequestID: "supplement-1"})
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			script := replayQueryScript()
			detail, found, err := test.find(NewGormStore(openScriptedGormDB(t, script)))
			if err != nil || !found || detail == nil {
				t.Fatalf("find replay = %#v, %v, %v", detail, found, err)
			}
			var keys []string
			for _, message := range detail.Messages {
				for _, attachment := range message.Attachments {
					keys = append(keys, attachment.ObjectKey)
				}
			}
			if want := []string{"feedback/a.jpg", "feedback/b.png"}; !reflect.DeepEqual(keys, want) {
				t.Fatalf("replay object keys = %#v, want %#v", keys, want)
			}
		})
	}
}

func TestReplayQueriesReturnNilFalseWhenMissing(t *testing.T) {
	tests := []struct {
		name string
		find func(*GormStore) (*application.FeedbackDetail, bool, error)
	}{
		{name: "create", find: func(store *GormStore) (*application.FeedbackDetail, bool, error) {
			return store.FindCreateReplay(context.Background(), application.CreateReplayKey{SubmitterID: 7, RequestID: "missing"})
		}},
		{name: "message", find: func(store *GormStore) (*application.FeedbackDetail, bool, error) {
			return store.FindMessageReplay(context.Background(), application.MessageReplayKey{FeedbackID: 9, AuthorType: domain.AuthorTypeUser, AuthorID: 7, RequestID: "missing"})
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			script := &databaseScript{query: func(_ string, _ []driver.NamedValue) queryResult {
				return queryResult{Columns: []string{"id", "feedback_id"}}
			}}
			detail, found, err := test.find(NewGormStore(openScriptedGormDB(t, script)))
			if detail != nil || found || err != nil {
				t.Fatalf("missing replay = %#v, %v, %v", detail, found, err)
			}
		})
	}
}

func TestGetUserFeedbackConstrainsIDAndOwner(t *testing.T) {
	script := &databaseScript{query: func(_ string, _ []driver.NamedValue) queryResult {
		return queryResult{Columns: []string{"id"}}
	}}
	store := NewGormStore(openScriptedGormDB(t, script))
	if _, err := store.GetUserFeedback(context.Background(), 9, 7); !errors.Is(err, application.ErrFeedbackNotFound) {
		t.Fatalf("GetUserFeedback error = %v", err)
	}
	queries, _ := script.statements()
	if len(queries) != 1 {
		t.Fatalf("detail query count = %d", len(queries))
	}
	sqlText := normalizeSQL(queries[0].SQL)
	if !strings.Contains(sqlText, "id = ?") || !strings.Contains(sqlText, "submitter_id = ?") {
		t.Fatalf("owned detail SQL = %s", sqlText)
	}
}

func TestOverviewQueriesEachExecuteOneConditionalAggregation(t *testing.T) {
	newScript := func() *databaseScript {
		return &databaseScript{query: func(_ string, _ []driver.NamedValue) queryResult {
			return queryResult{Columns: []string{"pending", "processing", "resolved", "closed"}, Rows: [][]driver.Value{{int64(1), int64(2), int64(3), int64(4)}}}
		}}
	}
	t.Run("user", func(t *testing.T) {
		script := newScript()
		got, err := NewGormStore(openScriptedGormDB(t, script)).GetUserOverview(context.Background(), 7)
		if err != nil || got != (application.Overview{Pending: 1, Processing: 2, Resolved: 3, Closed: 4}) {
			t.Fatalf("GetUserOverview = %#v, %v", got, err)
		}
		queries, _ := script.statements()
		assertConditionalOverviewSQL(t, queries, true)
		if strings.Contains(normalizeSQL(queries[0].SQL), "ORDER BY") || strings.Contains(normalizeSQL(queries[0].SQL), "LIMIT") {
			t.Fatalf("overview loaded list records: %s", queries[0].SQL)
		}
	})
	t.Run("admin ignores status card filter", func(t *testing.T) {
		script := newScript()
		handlerID := uint(3)
		query := application.AdminListQuery{
			Keyword: " 研发%_! ", SubmitterID: 7, HandlerID: &handlerID, Status: domain.StatusPending,
			SubmittedFrom: 1000, SubmittedTo: 2000, Page: 1, PageSize: 20,
		}
		if _, err := NewGormStore(openScriptedGormDB(t, script)).GetAdminOverview(context.Background(), query); err != nil {
			t.Fatalf("GetAdminOverview: %v", err)
		}
		queries, _ := script.statements()
		assertConditionalOverviewSQL(t, queries, false)
		sqlText := normalizeSQL(queries[0].SQL)
		for _, fragment := range []string{"submitter_id", "handler_id", "created_at >=", "created_at <=", "feedback_no LIKE", "user_feedback_messages", "users"} {
			if !strings.Contains(sqlText, fragment) {
				t.Fatalf("admin overview missing shared filter %q: %s", fragment, sqlText)
			}
		}
		if strings.Contains(sqlText, "WHERE feedback_status =") || strings.Contains(sqlText, "AND feedback_status =") {
			t.Fatalf("admin overview status filter polluted all cards: %s", sqlText)
		}
		if !containsNamedValue(queries[0].Args, "%研发!%!_!!%") {
			t.Fatalf("escaped keyword missing from %#v", namedValues(queries[0].Args))
		}
	})
}

func assertConditionalOverviewSQL(t *testing.T, queries []statementRecord, user bool) {
	t.Helper()
	if len(queries) != 1 {
		t.Fatalf("overview query count = %d", len(queries))
	}
	sqlText := normalizeSQL(queries[0].SQL)
	for _, fragment := range []string{
		"COALESCE(SUM(feedback_status = 'pending'), 0)",
		"COALESCE(SUM(feedback_status = 'processing'), 0)",
		"COALESCE(SUM(feedback_status = 'resolved'), 0)",
		"COALESCE(SUM(feedback_status = 'closed'), 0)",
	} {
		if !strings.Contains(sqlText, fragment) {
			t.Fatalf("overview SQL missing %q: %s", fragment, sqlText)
		}
	}
	if user && !strings.Contains(sqlText, "submitter_id") {
		t.Fatalf("user overview lacks owner constraint: %s", sqlText)
	}
}

func containsNamedValue(values []driver.NamedValue, want any) bool {
	for _, value := range values {
		if reflect.DeepEqual(value.Value, want) {
			return true
		}
	}
	return false
}

func TestAdminListUsesFixedBatchQueriesAndAllFilters(t *testing.T) {
	for _, size := range []int{2, 25} {
		t.Run(fmt.Sprintf("rows_%d", size), func(t *testing.T) {
			script := listQueryScript(size)
			handlerID := uint(3)
			query := application.AdminListQuery{
				Keyword: " 研发%_! ", SubmitterID: 7, HandlerID: &handlerID, Status: domain.StatusProcessing,
				SubmittedFrom: 1000, SubmittedTo: 3000, Page: 2, PageSize: 25,
			}
			result, err := NewGormStore(openScriptedGormDB(t, script)).ListAdminFeedbacks(context.Background(), query)
			if err != nil {
				t.Fatalf("ListAdminFeedbacks: %v", err)
			}
			if len(result.List) != size || result.Total != int64(size) || result.Page != 2 || result.PageSize != 25 {
				t.Fatalf("list result = %#v", result)
			}
			if result.List[0].SubmitterName != "提交人" || result.List[0].HandlerName != "管理员" ||
				result.List[0].Summary != "第1条 初始 内容" || result.List[0].ImageCount != 2 {
				t.Fatalf("enriched first row = %#v", result.List[0])
			}
			queries, _ := script.statements()
			if len(queries) != 5 {
				t.Fatalf("list query count for %d rows = %d, want fixed 5", size, len(queries))
			}
			listSQL := normalizeSQL(queries[1].SQL)
			for _, fragment := range []string{
				"feedback_status", "submitter_id", "handler_id", "created_at >=", "created_at <=",
				"feedback_no LIKE", "user_feedback_messages", "users", "ORDER BY last_activity_at DESC,id DESC", "LIMIT ? OFFSET ?",
			} {
				if !strings.Contains(listSQL, fragment) {
					t.Fatalf("admin list SQL missing %q: %s", fragment, listSQL)
				}
			}
			if !containsNamedValue(queries[1].Args, "%研发!%!_!!%") {
				t.Fatalf("escaped keyword missing from %#v", namedValues(queries[1].Args))
			}
		})
	}
}

func TestUserListIsOwnedAndEmptyResultIsNonNil(t *testing.T) {
	script := &databaseScript{query: func(query string, _ []driver.NamedValue) queryResult {
		if strings.Contains(strings.ToLower(normalizeSQL(query)), "count(*)") {
			return queryResult{Columns: []string{"count"}, Rows: [][]driver.Value{{int64(0)}}}
		}
		return queryResult{Columns: feedbackColumns()}
	}}
	result, err := NewGormStore(openScriptedGormDB(t, script)).ListUserFeedbacks(context.Background(), 7, application.UserListQuery{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("ListUserFeedbacks: %v", err)
	}
	if result.List == nil || len(result.List) != 0 {
		t.Fatalf("empty list = %#v", result.List)
	}
	queries, _ := script.statements()
	if len(queries) != 2 {
		t.Fatalf("empty list query count = %d", len(queries))
	}
	for _, record := range queries {
		if !strings.Contains(normalizeSQL(record.SQL), "submitter_id") {
			t.Fatalf("user list query lacks owner: %s", record.SQL)
		}
	}
}
