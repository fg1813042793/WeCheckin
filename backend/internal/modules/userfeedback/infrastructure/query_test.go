package infrastructure

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"math"
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
				rows = append(rows, []driver.Value{int64(index), int64(index + 1), fmt.Sprintf("feedback/%d-first.jpg", index)})
			}
			return queryResult{Columns: []string{"feedback_id", "image_count", "first_object_key"}, Rows: rows}
		case strings.Contains(sqlText, "FROM `users`"):
			return queryResult{Columns: []string{"id", "user_name"}, Rows: [][]driver.Value{{int64(3), "管理员"}, {int64(7), "提交人"}}}
		default:
			return queryResult{Err: fmt.Errorf("unexpected query: %s", query)}
		}
	}}
}

type feedbackKeywordFixture struct {
	initial       string
	supplement    string
	status        string
	submitterName string
}

func timelineKeywordScript(t *testing.T, keyword string, admin bool, fixture feedbackKeywordFixture) *databaseScript {
	t.Helper()
	return &databaseScript{query: func(query string, args []driver.NamedValue) queryResult {
		sqlText := normalizeSQL(query)
		switch {
		case strings.Contains(sqlText, "FROM `user_feedbacks`"):
			matched := scriptedKeywordMatches(t, sqlText, args, keyword, admin, fixture)
			if strings.Contains(strings.ToLower(sqlText), "count(*)") {
				total := int64(0)
				if matched {
					total = 1
				}
				return queryResult{Columns: []string{"count"}, Rows: [][]driver.Value{{total}}}
			}
			if matched {
				return queryResult{Columns: feedbackColumns(), Rows: [][]driver.Value{feedbackDriverRow(1)}}
			}
			return queryResult{Columns: feedbackColumns()}
		case strings.Contains(sqlText, "FROM `user_feedback_messages`"):
			if !strings.Contains(sqlText, "message_type = ?") || !containsNamedValue(args, string(domain.MessageTypeInitial)) {
				return queryResult{Err: fmt.Errorf("summary query must remain initial-only: %s %#v", sqlText, namedValues(args))}
			}
			return queryResult{
				Columns: []string{"feedback_id", "content"},
				Rows:    [][]driver.Value{{int64(1), fixture.initial}},
			}
		case strings.Contains(sqlText, "FROM `user_feedback_attachments`"):
			return queryResult{Columns: []string{"feedback_id", "image_count", "first_object_key"}}
		case strings.Contains(sqlText, "FROM `users`"):
			return queryResult{Columns: []string{"id", "user_name"}, Rows: [][]driver.Value{
				{int64(3), "管理员"}, {int64(7), fixture.submitterName},
			}}
		default:
			return queryResult{Err: fmt.Errorf("unexpected timeline keyword query: %s", query)}
		}
	}}
}

func scriptedKeywordMatches(t *testing.T, sqlText string, args []driver.NamedValue, keyword string, admin bool, fixture feedbackKeywordFixture) bool {
	t.Helper()
	expectedPattern := "%" + keyword + "%"
	wantPatternCount := 2
	if admin {
		wantPatternCount = 3
	}
	if got := countNamedValue(args, expectedPattern); got != wantPatternCount {
		t.Fatalf("keyword pattern count = %d, want %d in %#v", got, wantPatternCount, namedValues(args))
	}
	if strings.Contains(sqlText, keyword) {
		t.Fatalf("keyword was interpolated into SQL instead of parameterized: %s", sqlText)
	}

	contents := []string{fixture.initial, fixture.supplement, fixture.status}
	if strings.Contains(sqlText, "keyword_message.message_type") {
		contents = contents[:1]
	}
	for _, content := range contents {
		if strings.Contains(content, keyword) {
			return true
		}
	}
	return admin && strings.Contains(fixture.submitterName, keyword)
}

func countNamedValue(values []driver.NamedValue, want any) int {
	count := 0
	for _, value := range values {
		if reflect.DeepEqual(value.Value, want) {
			count++
		}
	}
	return count
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
	if detail.ImageCount != 2 || detail.FirstImageURL != "https://static.example/feedback/a.jpg" {
		t.Fatalf("detail image summary = count %d, first %q", detail.ImageCount, detail.FirstImageURL)
	}
	if detail.SubmitterName != "提交人" || detail.HandlerName != "管理员" || detail.Messages[0].AuthorName != "提交人" {
		t.Fatalf("display names = %#v", detail)
	}
}

func TestDetailMapperImageSummaryUsesSortOrderThenID(t *testing.T) {
	feedback := userfeedbackmodel.Feedback{ID: 9, FeedbackNo: "FB-9", SubmitterID: 7, Status: "pending"}
	attachments := []userfeedbackmodel.Attachment{
		{ID: 30, FeedbackID: 9, MessageID: 21, ObjectKey: "feedback/later.jpg", SortOrder: 1},
		{ID: 50, FeedbackID: 9, MessageID: 21, ObjectKey: "feedback/same-sort-later.jpg", SortOrder: 0},
		{ID: 40, FeedbackID: 9, MessageID: 21, ObjectKey: "feedback/first.jpg", SortOrder: 0},
	}
	var paths []string
	detail := buildFeedbackDetailWithURL(context.Background(), feedback, nil, attachments, nil, func(_ context.Context, path string) string {
		paths = append(paths, path)
		return "https://static.example" + path
	})
	if detail.ImageCount != 3 {
		t.Fatalf("detail image count = %d, want 3", detail.ImageCount)
	}
	if got, want := detail.FirstImageURL, "https://static.example/feedback/first.jpg"; got != want {
		t.Fatalf("detail first image URL = %q, want %q", got, want)
	}
	if want := []string{"/feedback/later.jpg", "/feedback/same-sort-later.jpg", "/feedback/first.jpg"}; !reflect.DeepEqual(paths, want) {
		t.Fatalf("URL builder paths = %#v, want %#v", paths, want)
	}
}

func TestDetailMapperImageSummaryIsEmptyWithoutAttachments(t *testing.T) {
	detail := buildFeedbackDetailWithURL(
		context.Background(),
		userfeedbackmodel.Feedback{ID: 9, FeedbackNo: "FB-9", SubmitterID: 7, Status: "pending"},
		nil,
		nil,
		nil,
		func(_ context.Context, path string) string {
			t.Fatalf("URL builder called without attachments: %q", path)
			return ""
		},
	)
	if detail.ImageCount != 0 || detail.FirstImageURL != "" {
		t.Fatalf("empty detail image summary = count %d, first %q", detail.ImageCount, detail.FirstImageURL)
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

func TestListKeywordMatchesSupplementAndStatusTimelineMessages(t *testing.T) {
	fixture := feedbackKeywordFixture{
		initial:       "仅用作列表摘要",
		supplement:    "用户补充了登录戳中详情",
		status:        "状态流转记录含审核驳回原因",
		submitterName: "当前提交人",
	}
	keywords := []struct {
		name    string
		keyword string
	}{
		{name: "supplement", keyword: "登录戳中"},
		{name: "status", keyword: "审核驳回"},
	}
	for _, admin := range []bool{false, true} {
		audience := "user"
		if admin {
			audience = "admin"
		}
		for _, keyword := range keywords {
			t.Run(audience+"_"+keyword.name, func(t *testing.T) {
				script := timelineKeywordScript(t, keyword.keyword, admin, fixture)
				store := NewGormStore(openScriptedGormDB(t, script))
				var (
					result application.FeedbackList
					err    error
				)
				if admin {
					result, err = store.ListAdminFeedbacks(context.Background(), application.AdminListQuery{
						Keyword: keyword.keyword, Page: 1, PageSize: 20,
					})
				} else {
					result, err = store.ListUserFeedbacks(context.Background(), 7, application.UserListQuery{
						Keyword: keyword.keyword, Page: 1, PageSize: 20,
					})
				}
				if err != nil {
					t.Fatalf("list timeline keyword: %v", err)
				}
				if result.Total != 1 || len(result.List) != 1 {
					t.Fatalf("timeline keyword result = %#v, want one match", result)
				}
				if got := result.List[0].Summary; got != fixture.initial {
					t.Fatalf("summary = %q, want initial content %q", got, fixture.initial)
				}
			})
		}
	}
}

func TestAdminListKeywordAlsoMatchesCurrentSubmitterName(t *testing.T) {
	fixture := feedbackKeywordFixture{
		initial:       "与关键词无关的初始内容",
		supplement:    "与关键词无关的补充内容",
		status:        "与关键词无关的状态内容",
		submitterName: "产品经理李明",
	}
	script := timelineKeywordScript(t, "李明", true, fixture)
	result, err := NewGormStore(openScriptedGormDB(t, script)).ListAdminFeedbacks(context.Background(), application.AdminListQuery{
		Keyword: "李明", Page: 1, PageSize: 20,
	})
	if err != nil {
		t.Fatalf("ListAdminFeedbacks by submitter name: %v", err)
	}
	if result.Total != 1 || len(result.List) != 1 || result.List[0].SubmitterName != fixture.submitterName {
		t.Fatalf("submitter-name keyword result = %#v", result)
	}
}

func TestListFeedbacksRejectsOffsetOverflowBeforeSQL(t *testing.T) {
	tests := []struct {
		name string
		call func(*GormStore) error
	}{
		{
			name: "user",
			call: func(store *GormStore) error {
				_, err := store.ListUserFeedbacks(context.Background(), 7, application.UserListQuery{Page: math.MaxInt, PageSize: 2})
				return err
			},
		},
		{
			name: "admin",
			call: func(store *GormStore) error {
				_, err := store.ListAdminFeedbacks(context.Background(), application.AdminListQuery{Page: math.MaxInt, PageSize: 2})
				return err
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			script := &databaseScript{query: func(query string, _ []driver.NamedValue) queryResult {
				if strings.Contains(strings.ToLower(normalizeSQL(query)), "count(*)") {
					return queryResult{Columns: []string{"count"}, Rows: [][]driver.Value{{int64(0)}}}
				}
				return queryResult{Columns: feedbackColumns()}
			}}
			err := test.call(NewGormStore(openScriptedGormDB(t, script)))
			if !errors.Is(err, application.ErrInvalidArgument) {
				t.Fatalf("overflow page error = %v, want ErrInvalidArgument", err)
			}
			queries, execs := script.statements()
			if len(queries) != 0 || len(execs) != 0 {
				t.Fatalf("overflow page executed SQL: queries=%#v execs=%#v", queries, execs)
			}
		})
	}
}

func TestListFeedbackOffsetSafeBoundaries(t *testing.T) {
	newEmptyScript := func() *databaseScript {
		return &databaseScript{query: func(query string, _ []driver.NamedValue) queryResult {
			if strings.Contains(strings.ToLower(normalizeSQL(query)), "count(*)") {
				return queryResult{Columns: []string{"count"}, Rows: [][]driver.Value{{int64(0)}}}
			}
			return queryResult{Columns: feedbackColumns()}
		}}
	}

	t.Run("first page", func(t *testing.T) {
		script := newEmptyScript()
		result, err := NewGormStore(openScriptedGormDB(t, script)).ListUserFeedbacks(context.Background(), 7, application.UserListQuery{Page: 1, PageSize: 20})
		if err != nil || result.Page != 1 || result.PageSize != 20 {
			t.Fatalf("first page result = %#v, %v", result, err)
		}
		queries, _ := script.statements()
		if len(queries) != 2 || strings.Contains(normalizeSQL(queries[1].SQL), "OFFSET") {
			t.Fatalf("first page SQL = %#v", queries)
		}
	})

	t.Run("maximum safe offset", func(t *testing.T) {
		const pageSize = application.MaxPageSize
		page := math.MaxInt/pageSize + 1
		wantOffset := (page - 1) * pageSize
		script := newEmptyScript()
		result, err := NewGormStore(openScriptedGormDB(t, script)).ListUserFeedbacks(context.Background(), 7, application.UserListQuery{Page: page, PageSize: pageSize})
		if err != nil || result.Page != page || result.PageSize != pageSize {
			t.Fatalf("maximum safe page result = %#v, %v", result, err)
		}
		queries, _ := script.statements()
		if len(queries) != 2 || !strings.Contains(normalizeSQL(queries[1].SQL), "OFFSET ?") {
			t.Fatalf("maximum safe page SQL = %#v", queries)
		}
		if !containsIntegerNamedValue(queries[1].Args, wantOffset) {
			t.Fatalf("maximum safe offset %d missing from %#v", wantOffset, namedValues(queries[1].Args))
		}
	})

	page, pageSize := queryPage(math.MaxInt, 2)
	if page != math.MaxInt || pageSize != 2 {
		t.Fatalf("queryPage remapped a large positive page to %d/%d", page, pageSize)
	}
}

func containsIntegerNamedValue(values []driver.NamedValue, want int) bool {
	for _, value := range values {
		switch got := value.Value.(type) {
		case int:
			if got == want {
				return true
			}
		case int64:
			if got == int64(want) {
				return true
			}
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
				result.List[0].Summary != "第1条 初始 内容" || result.List[0].ImageCount != 2 ||
				!strings.HasSuffix(result.List[0].FirstImageURL, "/feedback/1-first.jpg") {
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
			imageSQL := normalizeSQL(queries[3].SQL)
			for _, fragment := range []string{
				"COUNT(*) AS image_count", "first_attachment.object_key", "ORDER BY first_attachment.sort_order ASC, first_attachment.id ASC", "LIMIT 1",
			} {
				if !strings.Contains(imageSQL, fragment) {
					t.Fatalf("image summary SQL missing %q: %s", fragment, imageSQL)
				}
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
