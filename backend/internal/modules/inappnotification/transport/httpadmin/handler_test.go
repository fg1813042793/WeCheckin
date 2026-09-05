package httpadmin

import (
	"context"
	"strings"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route/param"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/modules/inappnotification/application"
	"wecheckin/backend/internal/support/notificationstyle"
)

func TestSendUsesManualSourceAndRecipientRuleFromRequest(t *testing.T) {
	service := &serviceStub{sendResult: application.SendResult{SourceID: "manual-1", SentCount: 2}}
	handler := NewHandler(service)
	c := newAdminContext(66)
	c.Request.SetBodyString(`{"requestId":"manual-1","title":"系统通知","content":"内容","scope":"departments","departmentIds":[3,5],"notificationType":"approval_result_returned"}`)

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
	if service.sendInput.NotificationType != "" {
		t.Fatalf("ordinary send must ignore notification type, got %q", service.sendInput.NotificationType)
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

func TestListUsesRecordFiltersAndMarkReadUsesAuthenticatedAdminUser(t *testing.T) {
	service := &serviceStub{listResult: application.NotificationList{Total: 1}}
	handler := NewHandler(service)
	listContext := newAdminContext(66)
	listContext.Request.SetRequestURI("/api/v2/admin/in-app-notifications?page=2&pageSize=15&title=%E7%B3%BB%E7%BB%9F&recipientName=%E5%BC%A0%E4%B8%89&sourceType=workflow&type=task_arrived&isRead=0&addTimeFrom=100&addTimeTo=200")

	handler.List(context.Background(), listContext)

	if service.listQuery.Title != "系统" || service.listQuery.RecipientName != "张三" ||
		service.listQuery.SourceType != "workflow" || service.listQuery.Type != "task_arrived" ||
		service.listQuery.IsRead == nil || *service.listQuery.IsRead != 0 ||
		service.listQuery.AddTimeFrom != 100 || service.listQuery.AddTimeTo != 200 ||
		service.listQuery.Page != 2 || service.listQuery.PageSize != 15 {
		t.Fatalf("list query = %#v", service.listQuery)
	}
	readContext := newAdminContext(66)
	readContext.Params = append(readContext.Params, param.Param{Key: "id", Value: "7"})
	handler.MarkRead(context.Background(), readContext)
	if service.markReadUserID != 66 || service.markReadID != 7 {
		t.Fatalf("mark read user=%d id=%d", service.markReadUserID, service.markReadID)
	}
}

func TestDeleteRecordUsesPathIDAndAuthenticatedAdmin(t *testing.T) {
	service := &serviceStub{}
	handler := NewHandler(service)
	c := newAdminContext(66)
	c.Params = append(c.Params, param.Param{Key: "id", Value: "7"})

	handler.DeleteRecord(context.Background(), c)

	if service.deleteRecordAdminID != 66 || service.deleteRecordID != 7 {
		t.Fatalf("delete record admin=%d id=%d", service.deleteRecordAdminID, service.deleteRecordID)
	}
	if !strings.Contains(string(c.Response.Body()), `"id":7`) {
		t.Fatalf("response = %s", c.Response.Body())
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

func TestNotificationStylesCanBeReadAndSaved(t *testing.T) {
	service := &serviceStub{styles: notificationstyle.DefaultConfig()}
	handler := NewHandler(service)
	readContext := newAdminContext(66)

	handler.NotificationStyles(context.Background(), readContext)

	if service.styleReadCalls != 1 || !strings.Contains(string(readContext.Response.Body()), `"task_arrived"`) {
		t.Fatalf("style read calls=%d response=%s", service.styleReadCalls, readContext.Response.Body())
	}
	writeContext := newAdminContext(66)
	writeContext.Request.SetBodyString(`{"version":1,"styles":[{"type":"task_arrived","label":"新待办","icon":"bell","tone":"danger"}]}`)

	handler.SaveNotificationStyles(context.Background(), writeContext)

	if service.styleSaveCalls != 1 || len(service.savedStyles.Styles) != 1 || service.savedStyles.Styles[0].Label != "新待办" {
		t.Fatalf("saved styles = %#v", service.savedStyles)
	}
}

func TestSendInAppStyleTestUsesSelectedTypeAndMarksTitleAsTest(t *testing.T) {
	service := &serviceStub{sendResult: application.SendResult{SentCount: 1}}
	handler := NewHandler(service)
	c := newAdminContext(66)
	c.Request.SetBodyString(`{"notificationType":"approval_result_returned","title":"退回通知","content":"测试内容","userIds":[3]}`)

	handler.SendInAppStyleTest(context.Background(), c)

	if service.sendInput.NotificationType != notificationstyle.TypeApprovalResultReturned {
		t.Fatalf("style test notification type = %q", service.sendInput.NotificationType)
	}
	if service.sendInput.Scope != application.ScopeUsers || len(service.sendInput.UserIDs) != 1 {
		t.Fatalf("style test recipients = %#v", service.sendInput)
	}
	if service.sendInput.Title != "[样式测试] 退回通知" {
		t.Fatalf("style test title = %q", service.sendInput.Title)
	}
}

func TestSendDingTalkStyleTestUsesSelectedUsersAndMarksTitleAsTest(t *testing.T) {
	service := &serviceStub{dingTalkSendResult: application.SendResult{SentCount: 2}}
	handler := NewHandler(service)
	c := newAdminContext(66)
	c.Request.SetBodyString(`{"notificationType":"task_reminder","title":"催办通知","content":"测试内容","userIds":[3,5]}`)

	handler.SendDingTalkStyleTest(context.Background(), c)

	if service.dingTalkSendCalls != 1 {
		t.Fatalf("DingTalk style test send calls = %d", service.dingTalkSendCalls)
	}
	if service.dingTalkSendInput.NotificationType != notificationstyle.TypeTaskReminder {
		t.Fatalf("DingTalk style test notification type = %q", service.dingTalkSendInput.NotificationType)
	}
	if service.dingTalkSendInput.Scope != application.ScopeUsers || len(service.dingTalkSendInput.UserIDs) != 2 {
		t.Fatalf("DingTalk style test recipients = %#v", service.dingTalkSendInput)
	}
	if service.dingTalkSendInput.Title != "[样式测试] 催办通知" {
		t.Fatalf("DingTalk style test title = %q", service.dingTalkSendInput.Title)
	}
}

func TestSendStyleTestRejectsEmptyOriginalTitle(t *testing.T) {
	service := &serviceStub{}
	handler := NewHandler(service)
	c := newAdminContext(66)
	c.Request.SetBodyString(`{"notificationType":"task_arrived","title":" ","content":"测试内容","userIds":[3]}`)

	handler.SendInAppStyleTest(context.Background(), c)

	if service.sendCalls != 0 || !strings.Contains(string(c.Response.Body()), "请输入通知标题") {
		t.Fatalf("send calls=%d response=%s", service.sendCalls, c.Response.Body())
	}
}

type serviceStub struct {
	sendInput           application.SendInput
	sendResult          application.SendResult
	sendCalls           int
	dingTalkSendInput   application.SendInput
	dingTalkSendResult  application.SendResult
	dingTalkSendCalls   int
	listResult          application.NotificationList
	listQuery           application.NotificationRecordQuery
	markReadUserID      uint
	markReadID          uint
	deleteRecordAdminID uint
	deleteRecordID      uint
	styles              notificationstyle.Config
	savedStyles         notificationstyle.Config
	styleReadCalls      int
	styleSaveCalls      int
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

func (stub *serviceStub) ListRecords(_ context.Context, query application.NotificationRecordQuery) (application.NotificationList, error) {
	stub.listQuery = query
	return stub.listResult, nil
}

func (stub *serviceStub) DeleteRecord(_ context.Context, adminID, notificationID uint) error {
	stub.deleteRecordAdminID = adminID
	stub.deleteRecordID = notificationID
	return nil
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

func (stub *serviceStub) NotificationStyles(context.Context) (notificationstyle.Config, error) {
	stub.styleReadCalls++
	return stub.styles, nil
}

func (stub *serviceStub) SaveNotificationStyles(_ context.Context, config notificationstyle.Config) (notificationstyle.Config, error) {
	stub.styleSaveCalls++
	stub.savedStyles = config
	return config, nil
}

func newAdminContext(adminID uint) *app.RequestContext {
	c := app.NewContext(1)
	c.Set("admin", &model.Admin{ID: adminID})
	c.Request.Header.SetContentTypeBytes([]byte("application/json"))
	return c
}
