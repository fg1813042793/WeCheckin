package httpadmin

import (
	"context"
	"testing"

	"github.com/cloudwego/hertz/pkg/route/param"
	"wecheckin/backend/internal/modules/userfeedback/application"
)

type notifyServiceFake struct {
	adminServiceFake
	command application.NotifySubmitterCommand
	calls   int
	context context.Context
}

func (fake *adminServiceFake) NotifySubmitter(context.Context, application.NotifySubmitterCommand) (*application.FeedbackDetail, error) {
	return nil, nil
}
func (fake *notifyServiceFake) NotifySubmitter(ctx context.Context, command application.NotifySubmitterCommand) (*application.FeedbackDetail, error) {
	fake.command, fake.context = command, ctx
	fake.calls++
	return &application.FeedbackDetail{}, nil
}

func TestNotifyUsesAuthenticatedAdminAndBodyOnly(t *testing.T) {
	service := &notifyServiceFake{}
	c := newAdminContext(66)
	c.Params = append(c.Params, param.Param{Key: "id", Value: "7"})
	c.Request.SetRequestURI("/api/v2/admin/user-feedbacks/7/notify?adminId=999&recipientUserId=998")
	c.Request.SetBodyString(`{"content":"请确认","version":3,"requestId":"notify-1"}`)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	NewHandler(service).NotifySubmitter(ctx, c)
	want := application.NotifySubmitterCommand{FeedbackID: 7, AdminID: 66, Content: "请确认", Version: 3, RequestID: "notify-1"}
	if service.calls != 1 || service.command != want || service.context != ctx {
		t.Fatalf("calls=%d command=%#v", service.calls, service.command)
	}
}

func TestNotifyRejectsUnauthenticatedMalformedAndRecipientOverride(t *testing.T) {
	for _, test := range []struct {
		name, id, body string
		adminID        uint
	}{
		{"anonymous", "7", `{"content":"通知","version":1,"requestId":"r1"}`, 0},
		{"invalid id", "0", `{"content":"通知","version":1,"requestId":"r1"}`, 66},
		{"unknown recipient", "7", `{"content":"通知","version":1,"requestId":"r1","recipientUserId":9}`, 66},
		{"unknown admin", "7", `{"content":"通知","version":1,"requestId":"r1","adminId":9}`, 66},
		{"missing version", "7", `{"content":"通知","requestId":"r1"}`, 66},
		{"missing request", "7", `{"content":"通知","version":1}`, 66},
		{"empty content", "7", `{"content":"  ","version":1,"requestId":"r1"}`, 66},
		{"trailing body", "7", `{"content":"通知","version":1,"requestId":"r1"}{}`, 66},
	} {
		t.Run(test.name, func(t *testing.T) {
			service := &notifyServiceFake{}
			c := newAdminContext(test.adminID)
			c.Params = append(c.Params, param.Param{Key: "id", Value: test.id})
			c.Request.SetBodyString(test.body)
			NewHandler(service).NotifySubmitter(context.Background(), c)
			if service.calls != 0 {
				t.Fatal("invalid notification reached application")
			}
		})
	}
}
