package httpadmin

import (
	"context"
	"strings"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route/param"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/modules/inappnotification/application"
)

func TestSendUsesManualSourceAndRecipientRuleFromRequest(t *testing.T) {
	service := &serviceStub{sendResult: application.SendResult{SourceID: "manual-1", SentCount: 2}}
	handler := NewHandler(service)
	c := newAdminContext(66)
	c.Request.SetBodyString(`{"requestId":"manual-1","title":"系统通知","content":"内容","scope":"departments","departmentIds":[3,5]}`)

	handler.Send(context.Background(), c)

	if service.sendCalls != 1 {
		t.Fatalf("send calls = %d", service.sendCalls)
	}
	if service.sendInput.SourceType != application.SourceAdminManual || service.sendInput.SourceID != "manual-1" {
		t.Fatalf("send source = %q/%q", service.sendInput.SourceType, service.sendInput.SourceID)
	}
	if service.sendInput.Scope != application.ScopeDepartments || len(service.sendInput.DepartmentIDs) != 2 {
		t.Fatalf("send recipients = %#v", service.sendInput)
	}
	if !strings.Contains(string(c.Response.Body()), `"sentCount":2`) {
		t.Fatalf("response = %s", c.Response.Body())
	}
}

func TestSendGeneratesSourceIDWhenRequestIDIsEmpty(t *testing.T) {
	service := &serviceStub{}
	handler := NewHandler(service)
	handler.newSourceID = func() (string, error) { return "generated-1", nil }
	c := newAdminContext(66)
	c.Request.SetBodyString(`{"title":"系统通知","content":"内容","scope":"all"}`)

	handler.Send(context.Background(), c)

	if service.sendInput.SourceID != "generated-1" {
		t.Fatalf("generated source ID = %q", service.sendInput.SourceID)
	}
}

func TestSendDingTalkUsesManualDingTalkSource(t *testing.T) {
	service := &serviceStub{dingTalkSendResult: application.SendResult{SourceID: "ding-1", SentCount: 2, FailedCount: 1}}
	handler := NewHandler(service)
	c := newAdminContext(66)
	c.Request.SetBodyString(`{"requestId":"ding-1","title":"钉钉通知","content":"内容","scope":"users","userIds":[3,5]}`)

	handler.SendDingTalk(context.Background(), c)

	if service.dingTalkSendCalls != 1 {
		t.Fatalf("dingTalk send calls = %d", service.dingTalkSendCalls)
	}
	if service.dingTalkSendInput.SourceType != application.SourceAdminManualDingTalk || service.dingTalkSendInput.SourceID != "ding-1" {
		t.Fatalf("dingTalk send source = %q/%q", service.dingTalkSendInput.SourceType, service.dingTalkSendInput.SourceID)
	}
	if service.dingTalkSendInput.Scope != application.ScopeUsers || len(service.dingTalkSendInput.UserIDs) != 2 {
		t.Fatalf("dingTalk send recipients = %#v", service.dingTalkSendInput)
	}
	if !strings.Contains(string(c.Response.Body()), `"sentCount":2`) || !strings.Contains(string(c.Response.Body()), `"failedCount":1`) {
		t.Fatalf("response = %s", c.Response.Body())
	}
}

func TestListAndMarkReadUseAuthenticatedAdminUser(t *testing.T) {
	service := &serviceStub{listResult: application.NotificationList{Total: 1}}
	handler := NewHandler(service)
	listContext := newAdminContext(66)
	listContext.Request.SetRequestURI("/api/v2/admin/in-app-notifications?page=2&pageSize=15&userId=999")

	handler.List(context.Background(), listContext)

	if service.listUserID != 66 || service.listPage != 2 || service.listPageSize != 15 {
		t.Fatalf("list query user=%d page=%d pageSize=%d", service.listUserID, service.listPage, service.listPageSize)
	}
	readContext := newAdminContext(66)
	readContext.Params = append(readContext.Params, param.Param{Key: "id", Value: "7"})
	handler.MarkRead(context.Background(), readContext)
	if service.markReadUserID != 66 || service.markReadID != 7 {
		t.Fatalf("mark read user=%d id=%d", service.markReadUserID, service.markReadID)
	}
}

func TestInboxHandlersRejectUnauthenticatedRequests(t *testing.T) {
	service := &serviceStub{}
	handler := NewHandler(service)
	c := app.NewContext(1)
	c.Request.SetBodyString(`{"title":"系统通知","content":"内容","scope":"all"}`)

	handler.Send(context.Background(), c)

	if service.sendCalls != 0 || !strings.Contains(string(c.Response.Body()), "未登录") {
		t.Fatalf("send calls=%d response=%s", service.sendCalls, c.Response.Body())
	}
}

type serviceStub struct {
	sendInput          application.SendInput
	sendResult         application.SendResult
	sendCalls          int
	dingTalkSendInput  application.SendInput
	dingTalkSendResult application.SendResult
	dingTalkSendCalls  int
	listResult         application.NotificationList
	listUserID         uint
	listPage           int
	listPageSize       int
	markReadUserID     uint
	markReadID         uint
}

func (stub *serviceStub) Send(_ context.Context, input application.SendInput) (application.SendResult, error) {
	stub.sendCalls++
	stub.sendInput = input
	return stub.sendResult, nil
}

func (stub *serviceStub) SendDingTalk(_ context.Context, input application.SendInput) (application.SendResult, error) {
	stub.dingTalkSendCalls++
	stub.dingTalkSendInput = input
	return stub.dingTalkSendResult, nil
}

func (stub *serviceStub) List(_ context.Context, userID uint, page, pageSize int) (application.NotificationList, error) {
	stub.listUserID = userID
	stub.listPage = page
	stub.listPageSize = pageSize
	return stub.listResult, nil
}

func (stub *serviceStub) UnreadCount(context.Context, uint) (int64, error) { return 0, nil }

func (stub *serviceStub) MarkRead(_ context.Context, userID, notificationID uint) error {
	stub.markReadUserID = userID
	stub.markReadID = notificationID
	return nil
}

func (stub *serviceStub) MarkAllRead(context.Context, uint) error { return nil }

func (stub *serviceStub) RecipientOptions(context.Context) (application.RecipientOptions, error) {
	return application.RecipientOptions{}, nil
}

func newAdminContext(adminID uint) *app.RequestContext {
	c := app.NewContext(1)
	c.Set("admin", &model.Admin{ID: adminID})
	c.Request.Header.SetContentTypeBytes([]byte("application/json"))
	return c
}
