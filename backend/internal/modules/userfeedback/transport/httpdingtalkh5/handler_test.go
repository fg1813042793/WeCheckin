package httpdingtalkh5

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"net/textproto"
	"strings"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route/param"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/modules/userfeedback/application"
	"wecheckin/backend/internal/modules/userfeedback/domain"
	"wecheckin/backend/internal/support/dingtalkh5session"
)

func TestUserFeedbackHandlersUseOnlyAuthenticatedUserIdentity(t *testing.T) {
	service := &serviceFake{}
	handler := NewHandler(service)

	listContext := newUserContext(42)
	listContext.Request.SetRequestURI("/api/v2/dingtalk/h5/user-feedbacks?userId=999&submitterId=998&page=2&pageSize=30&status=processing&keyword=login")
	handler.List(context.Background(), listContext)
	if service.listCalls != 1 || service.listUserID != 42 {
		t.Fatalf("list calls=%d user=%d", service.listCalls, service.listUserID)
	}
	if service.userQuery.Page != 2 || service.userQuery.PageSize != 30 || service.userQuery.Status != domain.StatusProcessing || service.userQuery.Keyword != "login" {
		t.Fatalf("list query = %#v", service.userQuery)
	}

	createContext := newMultipartUserContext(t, 42, map[string]string{
		"content":     "登录页异常",
		"requestId":   "request-create",
		"submitterId": "997",
		"userId":      "996",
	}, []multipartTestFile{
		{name: "..\\secret.png", contentType: "image/png", content: []byte("first")},
		{name: "screen.png", contentType: "image/png", content: []byte("second")},
	})
	createContext.Request.SetRequestURI("/api/v2/dingtalk/h5/user-feedbacks?submitterId=995&userId=994")
	handler.Create(context.Background(), createContext)
	if service.createCalls != 1 || service.createCommand.SubmitterID != 42 {
		t.Fatalf("create calls=%d command=%#v", service.createCalls, service.createCommand)
	}
	if service.createCommand.Content != "登录页异常" || service.createCommand.RequestID != "request-create" || len(service.createCommand.Attachments) != 2 {
		t.Fatalf("create command = %#v", service.createCommand)
	}
	if strings.ContainsAny(service.createCommand.Attachments[0].OriginalName, `/\\`) {
		t.Fatalf("unsafe original filename = %q", service.createCommand.Attachments[0].OriginalName)
	}
	if service.createCommand.Attachments[0].ContentType != "image/png" || string(service.createCommand.Attachments[1].Content) != "second" {
		t.Fatalf("attachments = %#v", service.createCommand.Attachments)
	}

	supplementContext := newMultipartUserContext(t, 42, map[string]string{
		"content":     "补充截图",
		"requestId":   "request-supplement",
		"version":     "3",
		"submitterId": "993",
	}, nil)
	supplementContext.Params = append(supplementContext.Params, param.Param{Key: "id", Value: "7"})
	handler.Supplement(context.Background(), supplementContext)
	if service.supplementCalls != 1 || service.supplementCommand.FeedbackID != 7 || service.supplementCommand.SubmitterID != 42 || service.supplementCommand.Version != 3 {
		t.Fatalf("supplement calls=%d command=%#v", service.supplementCalls, service.supplementCommand)
	}
}

func TestUserFeedbackHandlersRejectMissingAuthenticatedUser(t *testing.T) {
	service := &serviceFake{}
	handler := NewHandler(service)
	c := app.NewContext(1)

	handler.Overview(context.Background(), c)

	if service.overviewCalls != 0 || !responseContains(c, "未登录或权限失效") {
		t.Fatalf("overview calls=%d response=%s", service.overviewCalls, c.Response.Body())
	}
}

func TestUserFeedbackHandlerStrictlyParsesQueryAndPathIDs(t *testing.T) {
	invalidQueries := []string{
		"page=abc",
		"pageSize=1.5",
		"status=unknown",
	}
	for _, query := range invalidQueries {
		t.Run(query, func(t *testing.T) {
			service := &serviceFake{}
			handler := NewHandler(service)
			c := newUserContext(42)
			c.Request.SetRequestURI("/api/v2/dingtalk/h5/user-feedbacks?" + query)

			handler.List(context.Background(), c)

			if service.listCalls != 0 || !responseContains(c, "请求参数不正确") {
				t.Fatalf("list calls=%d response=%s", service.listCalls, c.Response.Body())
			}
		})
	}

	for _, id := range []string{"", "0", "-1", "abc"} {
		t.Run("id="+id, func(t *testing.T) {
			service := &serviceFake{}
			handler := NewHandler(service)
			c := newUserContext(42)
			c.Params = append(c.Params, param.Param{Key: "id", Value: id})

			handler.Detail(context.Background(), c)

			if service.detailCalls != 0 || !responseContains(c, "反馈编号不正确") {
				t.Fatalf("detail calls=%d response=%s", service.detailCalls, c.Response.Body())
			}
		})
	}
}

func TestUserFeedbackMultipartBodyLimitRunsBeforeParsing(t *testing.T) {
	service := &serviceFake{}
	handler := NewHandler(service)
	c := newUserContext(42)
	c.Request.Header.SetContentTypeBytes([]byte("application/json"))
	c.Request.Header.SetContentLength(MaxMultipartBodyBytes + 1)
	c.Request.SetBodyString("not multipart")
	c.Request.Header.SetContentLength(MaxMultipartBodyBytes + 1)

	handler.Create(context.Background(), c)

	if service.createCalls != 0 || !responseContains(c, "反馈请求总大小不能超过 31MB") {
		t.Fatalf("create calls=%d response=%s", service.createCalls, c.Response.Body())
	}
}

func TestUserFeedbackMultipartRequiresMultipartFormData(t *testing.T) {
	service := &serviceFake{}
	handler := NewHandler(service)
	c := newUserContext(42)
	c.Request.Header.SetContentTypeBytes([]byte("application/json"))
	c.Request.SetBodyString(`{"content":"x"}`)

	handler.Create(context.Background(), c)

	if service.createCalls != 0 || !responseContains(c, "请使用 multipart/form-data 提交反馈") {
		t.Fatalf("create calls=%d response=%s", service.createCalls, c.Response.Body())
	}
}

func TestUserFeedbackMultipartRejectsFileLimitsBeforeService(t *testing.T) {
	t.Run("too many images", func(t *testing.T) {
		files := make([]multipartTestFile, application.MaxImagesPerMessage+1)
		for i := range files {
			files[i] = multipartTestFile{name: fmt.Sprintf("%d.png", i), contentType: "image/png", content: []byte("x")}
		}
		service := &serviceFake{}
		handler := NewHandler(service)
		c := newMultipartUserContext(t, 42, map[string]string{"content": "x", "requestId": "many"}, files)

		handler.Create(context.Background(), c)

		if service.createCalls != 0 || !responseContains(c, "每次最多上传 6 张图片") {
			t.Fatalf("create calls=%d response=%s", service.createCalls, c.Response.Body())
		}
	})

	t.Run("forged large header", func(t *testing.T) {
		service := &serviceFake{}
		handler := NewHandler(service)
		c := newMultipartUserContext(t, 42, map[string]string{"content": "x", "requestId": "large-header"}, []multipartTestFile{{name: "a.png", contentType: "image/png", content: []byte("x")}})
		form, err := c.MultipartForm()
		if err != nil {
			t.Fatal(err)
		}
		form.File["images"][0].Size = application.MaxImageSizeBytes + 1

		handler.Create(context.Background(), c)

		if service.createCalls != 0 || !responseContains(c, "单张图片不能超过 10MB") {
			t.Fatalf("create calls=%d response=%s", service.createCalls, c.Response.Body())
		}
	})

	t.Run("forged small header", func(t *testing.T) {
		service := &serviceFake{}
		handler := NewHandler(service)
		content := bytes.Repeat([]byte{'x'}, application.MaxImageSizeBytes+1)
		c := newMultipartUserContext(t, 42, map[string]string{"content": "x", "requestId": "small-header"}, []multipartTestFile{{name: "a.png", contentType: "image/png", content: content}})
		form, err := c.MultipartForm()
		if err != nil {
			t.Fatal(err)
		}
		form.File["images"][0].Size = 1

		handler.Create(context.Background(), c)

		if service.createCalls != 0 || !responseContains(c, "单张图片不能超过 10MB") {
			t.Fatalf("create calls=%d response=%s", service.createCalls, c.Response.Body())
		}
	})

	t.Run("empty image", func(t *testing.T) {
		service := &serviceFake{}
		handler := NewHandler(service)
		c := newMultipartUserContext(t, 42, map[string]string{"content": "x", "requestId": "empty"}, []multipartTestFile{{name: "a.png", contentType: "image/png"}})

		handler.Create(context.Background(), c)

		if service.createCalls != 0 || !responseContains(c, "图片文件不能为空") {
			t.Fatalf("create calls=%d response=%s", service.createCalls, c.Response.Body())
		}
	})
}

func TestUserFeedbackNotFoundResponseDoesNotRevealOwnership(t *testing.T) {
	service := &serviceFake{detailErr: application.ErrFeedbackNotFound}
	handler := NewHandler(service)
	c := newUserContext(42)
	c.Params = append(c.Params, param.Param{Key: "id", Value: "7"})

	handler.Detail(context.Background(), c)

	if service.detailUserID != 42 || !responseContains(c, "反馈不存在") || responseContains(c, "无权") {
		t.Fatalf("detail user=%d response=%s", service.detailUserID, c.Response.Body())
	}
}

func TestUserFeedbackBusinessErrorsHaveStableMessages(t *testing.T) {
	tests := []struct {
		err  error
		want string
	}{
		{application.ErrInvalidArgument, "请求参数不正确"},
		{application.ErrAttachmentTypeNotAllowed, "仅支持 JPG、PNG、WebP 图片"},
		{application.ErrAttachmentTooLarge, "单张图片不能超过 10MB"},
		{application.ErrAttachmentLimitExceeded, "图片数量超过限制"},
		{application.ErrDailyLimitExceeded, "今日反馈数量已达上限"},
		{application.ErrFeedbackNotFound, "反馈不存在"},
		{application.ErrSupplementNotAllowed, "当前状态不允许补充反馈"},
		{application.ErrTransitionNotAllowed, "不允许执行该状态变更"},
		{application.ErrTransitionNoteRequired, "当前状态变更需要填写处理说明"},
		{application.ErrVersionConflict, "反馈已更新，请刷新后重试"},
		{application.ErrDuplicateRequest, "请勿重复提交"},
		{application.ErrTransactionOutcomeUnknown, "操作结果暂未确认，请稍后刷新"},
	}
	for _, test := range tests {
		t.Run(test.want, func(t *testing.T) {
			service := &serviceFake{listErr: test.err}
			handler := NewHandler(service)
			c := newUserContext(42)

			handler.List(context.Background(), c)

			if !responseContains(c, test.want) {
				t.Fatalf("response=%s want=%q", c.Response.Body(), test.want)
			}
		})
	}
}

func TestUserFeedbackInternalErrorIsRedacted(t *testing.T) {
	service := &serviceFake{listErr: errors.New("dial mysql private-host:3306: password=secret")}
	handler := NewHandler(service)
	c := newUserContext(42)

	handler.List(context.Background(), c)

	responseBody := string(c.Response.Body())
	if !strings.Contains(responseBody, "反馈操作失败，请稍后重试") || strings.Contains(responseBody, "private-host") || strings.Contains(responseBody, "secret") {
		t.Fatalf("response=%s", responseBody)
	}
}

type multipartTestFile struct {
	name        string
	contentType string
	content     []byte
}

func newMultipartUserContext(t *testing.T, userID uint, fields map[string]string, files []multipartTestFile) *app.RequestContext {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatal(err)
		}
	}
	for _, file := range files {
		header := make(textproto.MIMEHeader)
		header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="images"; filename="%s"`, file.name))
		header.Set("Content-Type", file.contentType)
		part, err := writer.CreatePart(header)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := part.Write(file.content); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	c := newUserContext(userID)
	c.Request.SetBody(body.Bytes())
	c.Request.Header.SetContentTypeBytes([]byte(writer.FormDataContentType()))
	return c
}

func newUserContext(userID uint) *app.RequestContext {
	c := app.NewContext(1)
	dingtalkh5session.SetAuth(c, &model.DingTalkH5PerfUser{ID: userID}, "token")
	return c
}

func responseContains(c *app.RequestContext, value string) bool {
	return strings.Contains(string(c.Response.Body()), value)
}

type serviceFake struct {
	overviewResult application.Overview
	overviewErr    error
	overviewCalls  int
	overviewUserID uint

	listResult application.FeedbackList
	listErr    error
	listCalls  int
	listUserID uint
	userQuery  application.UserListQuery

	createResult  *application.FeedbackDetail
	createErr     error
	createCalls   int
	createCommand application.CreateCommand

	detailResult *application.FeedbackDetail
	detailErr    error
	detailCalls  int
	detailID     uint64
	detailUserID uint

	supplementResult  *application.FeedbackDetail
	supplementErr     error
	supplementCalls   int
	supplementCommand application.SupplementCommand
}

func (fake *serviceFake) GetUserOverview(_ context.Context, userID uint) (application.Overview, error) {
	fake.overviewCalls++
	fake.overviewUserID = userID
	return fake.overviewResult, fake.overviewErr
}

func (fake *serviceFake) ListUserFeedbacks(_ context.Context, userID uint, query application.UserListQuery) (application.FeedbackList, error) {
	fake.listCalls++
	fake.listUserID = userID
	fake.userQuery = query
	return fake.listResult, fake.listErr
}

func (fake *serviceFake) CreateFeedback(_ context.Context, command application.CreateCommand) (*application.FeedbackDetail, error) {
	fake.createCalls++
	fake.createCommand = command
	return fake.createResult, fake.createErr
}

func (fake *serviceFake) GetUserFeedback(_ context.Context, id uint64, userID uint) (*application.FeedbackDetail, error) {
	fake.detailCalls++
	fake.detailID = id
	fake.detailUserID = userID
	return fake.detailResult, fake.detailErr
}

func (fake *serviceFake) SupplementFeedback(_ context.Context, command application.SupplementCommand) (*application.FeedbackDetail, error) {
	fake.supplementCalls++
	fake.supplementCommand = command
	return fake.supplementResult, fake.supplementErr
}
