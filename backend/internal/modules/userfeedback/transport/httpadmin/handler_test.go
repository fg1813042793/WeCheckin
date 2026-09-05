package httpadmin

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route/param"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/modules/userfeedback/application"
	"wecheckin/backend/internal/modules/userfeedback/domain"
)

func TestAdminUserFeedbackListStrictlyParsesFilters(t *testing.T) {
	service := &adminServiceFake{}
	handler := NewHandler(service)
	c := newAdminContext(66)
	c.Request.SetRequestURI("/api/v2/admin/user-feedbacks?keyword=login&status=processing&submitterId=12&handlerId=13&submittedFrom=100&submittedTo=200&page=2&pageSize=50")

	handler.List(context.Background(), c)

	query := service.listQuery
	if service.listCalls != 1 || query.Keyword != "login" || query.Status != domain.StatusProcessing || query.SubmitterID != 12 || query.HandlerID == nil || *query.HandlerID != 13 || query.SubmittedFrom != 100 || query.SubmittedTo != 200 || query.Page != 2 || query.PageSize != 50 {
		t.Fatalf("list calls=%d query=%#v", service.listCalls, query)
	}
}

func TestAdminUserFeedbackListRejectsMalformedFilters(t *testing.T) {
	queries := []string{
		"page=abc",
		"pageSize=1.5",
		"status=unknown",
		"submitterId=abc",
		"submitterId=0",
		"handlerId=abc",
		"handlerId=0",
		"submittedFrom=abc",
		"submittedTo=abc",
	}
	for _, query := range queries {
		t.Run(query, func(t *testing.T) {
			service := &adminServiceFake{}
			handler := NewHandler(service)
			c := newAdminContext(66)
			c.Request.SetRequestURI("/api/v2/admin/user-feedbacks?" + query)

			handler.List(context.Background(), c)

			if service.listCalls != 0 || !adminResponseContains(c, "请求参数不正确") {
				t.Fatalf("list calls=%d response=%s", service.listCalls, c.Response.Body())
			}
		})
	}
}

func TestAdminUserFeedbackDetailStrictlyParsesPositiveID(t *testing.T) {
	for _, id := range []string{"", "0", "-1", "abc"} {
		t.Run("id="+id, func(t *testing.T) {
			service := &adminServiceFake{}
			handler := NewHandler(service)
			c := newAdminContext(66)
			c.Params = append(c.Params, param.Param{Key: "id", Value: id})

			handler.Detail(context.Background(), c)

			if service.detailCalls != 0 || !adminResponseContains(c, "反馈编号不正确") {
				t.Fatalf("detail calls=%d response=%s", service.detailCalls, c.Response.Body())
			}
		})
	}
}

func TestAdminUpdateStatusUsesAuthenticatedAdminAndDefaultsNotifyUser(t *testing.T) {
	service := &adminServiceFake{}
	handler := NewHandler(service)
	c := newAdminContext(66)
	c.Request.SetRequestURI("/api/v2/admin/user-feedbacks/7/status?adminId=999&handlerId=998")
	c.Params = append(c.Params, param.Param{Key: "id", Value: "7"})
	c.Request.SetBodyString(`{"status":"processing","note":"开始处理","version":3,"requestId":"status-1"}`)
	c.Request.Header.SetContentTypeBytes([]byte("application/json"))

	handler.UpdateStatus(context.Background(), c)

	command := service.updateCommand
	if service.updateCalls != 1 || command.AdminID != 66 || command.FeedbackID != 7 || command.Status != domain.StatusProcessing || command.Note != "开始处理" || command.Version != 3 || command.RequestID != "status-1" || !command.NotifyUser {
		t.Fatalf("update calls=%d command=%#v", service.updateCalls, command)
	}
}

func TestAdminUpdateStatusPreservesExplicitFalseNotifyUser(t *testing.T) {
	service := &adminServiceFake{}
	handler := NewHandler(service)
	c := newAdminContext(66)
	c.Params = append(c.Params, param.Param{Key: "id", Value: "7"})
	c.Request.SetBodyString(`{"status":"processing","note":"开始处理","notifyUser":false,"version":3,"requestId":"status-2"}`)
	c.Request.Header.SetContentTypeBytes([]byte("application/json"))

	handler.UpdateStatus(context.Background(), c)

	if service.updateCalls != 1 || service.updateCommand.NotifyUser {
		t.Fatalf("update calls=%d command=%#v", service.updateCalls, service.updateCommand)
	}
}

func TestAdminUpdateStatusRejectsMissingAndUnknownFields(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{name: "missing version", body: `{"status":"processing","requestId":"r1"}`, want: "请提供当前版本号"},
		{name: "missing request id", body: `{"status":"processing","version":1}`, want: "请提供请求标识"},
		{name: "invalid status", body: `{"status":"unknown","version":1,"requestId":"r1"}`, want: "反馈状态无效"},
		{name: "identity in body", body: `{"status":"processing","version":1,"requestId":"r1","adminId":999}`, want: "请求参数不正确"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := &adminServiceFake{}
			handler := NewHandler(service)
			c := newAdminContext(66)
			c.Params = append(c.Params, param.Param{Key: "id", Value: "7"})
			c.Request.SetBodyString(test.body)
			c.Request.Header.SetContentTypeBytes([]byte("application/json"))

			handler.UpdateStatus(context.Background(), c)

			if service.updateCalls != 0 || !adminResponseContains(c, test.want) {
				t.Fatalf("update calls=%d response=%s", service.updateCalls, c.Response.Body())
			}
		})
	}
}

func TestAdminUserFeedbackHandlersRejectMissingAuthenticatedAdmin(t *testing.T) {
	service := &adminServiceFake{}
	handler := NewHandler(service)
	c := app.NewContext(1)

	handler.Overview(context.Background(), c)

	if service.overviewCalls != 0 || !adminResponseContains(c, "未登录或权限失效") {
		t.Fatalf("overview calls=%d response=%s", service.overviewCalls, c.Response.Body())
	}
}

func TestAdminUserFeedbackBusinessErrorsHaveStableMessages(t *testing.T) {
	tests := []struct {
		err  error
		want string
	}{
		{application.ErrInvalidArgument, "请求参数不正确"},
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
			service := &adminServiceFake{updateErr: test.err}
			handler := NewHandler(service)
			c := newAdminContext(66)
			c.Params = append(c.Params, param.Param{Key: "id", Value: "7"})
			c.Request.SetBodyString(`{"status":"processing","version":1,"requestId":"r1"}`)
			c.Request.Header.SetContentTypeBytes([]byte("application/json"))

			handler.UpdateStatus(context.Background(), c)

			if !adminResponseContains(c, test.want) {
				t.Fatalf("response=%s want=%q", c.Response.Body(), test.want)
			}
		})
	}
}

func TestAdminUserFeedbackInternalErrorIsRedacted(t *testing.T) {
	service := &adminServiceFake{listErr: errors.New("select from private_db with token=secret")}
	handler := NewHandler(service)
	c := newAdminContext(66)

	handler.List(context.Background(), c)

	responseBody := string(c.Response.Body())
	if !strings.Contains(responseBody, "反馈操作失败，请稍后重试") || strings.Contains(responseBody, "private_db") || strings.Contains(responseBody, "secret") {
		t.Fatalf("response=%s", responseBody)
	}
}

func newAdminContext(adminID uint) *app.RequestContext {
	c := app.NewContext(1)
	c.Set("admin", &model.Admin{ID: adminID})
	return c
}

func adminResponseContains(c *app.RequestContext, value string) bool {
	return strings.Contains(string(c.Response.Body()), value)
}

type adminServiceFake struct {
	overviewResult application.Overview
	overviewErr    error
	overviewCalls  int
	overviewQuery  application.AdminListQuery

	listResult application.FeedbackList
	listErr    error
	listCalls  int
	listQuery  application.AdminListQuery

	detailResult *application.FeedbackDetail
	detailErr    error
	detailCalls  int
	detailID     uint64

	updateResult  *application.FeedbackDetail
	updateErr     error
	updateCalls   int
	updateCommand application.UpdateStatusCommand
}

func (fake *adminServiceFake) GetAdminOverview(_ context.Context, query application.AdminListQuery) (application.Overview, error) {
	fake.overviewCalls++
	fake.overviewQuery = query
	return fake.overviewResult, fake.overviewErr
}

func (fake *adminServiceFake) ListAdminFeedbacks(_ context.Context, query application.AdminListQuery) (application.FeedbackList, error) {
	fake.listCalls++
	fake.listQuery = query
	return fake.listResult, fake.listErr
}

func (fake *adminServiceFake) GetAdminFeedback(_ context.Context, id uint64) (*application.FeedbackDetail, error) {
	fake.detailCalls++
	fake.detailID = id
	return fake.detailResult, fake.detailErr
}

func (fake *adminServiceFake) UpdateFeedbackStatus(_ context.Context, command application.UpdateStatusCommand) (*application.FeedbackDetail, error) {
	fake.updateCalls++
	fake.updateCommand = command
	return fake.updateResult, fake.updateErr
}
