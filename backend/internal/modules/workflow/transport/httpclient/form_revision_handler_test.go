package httpclient

import (
	"context"
	"strings"
	"testing"

	"github.com/cloudwego/hertz/pkg/route/param"

	workflowapp "wecheckin/backend/internal/modules/workflow/application"
)

func TestCreateCompletedFormRevisionUsesAuthenticatedActor(t *testing.T) {
	stub := &runtimeServiceStub{}
	handler := NewRuntimeHandler(stub)
	c := newUserContext(42)
	c.Params = append(c.Params, param.Param{Key: "id", Value: "instance-1"})
	c.Request.SetBodyString(`{"sourceNodeId":"manager","expectedRevision":3,"formData":{"remark":"x"},"reason":"修正","requesterId":"forged-user","instanceId":"forged-instance"}`)

	handler.CreateCompletedFormRevision(context.Background(), c)

	if stub.createRevisionRequest.RequesterID != "42" || stub.createRevisionRequest.InstanceID != "instance-1" {
		t.Fatalf("request=%#v", stub.createRevisionRequest)
	}
	if stub.createRevisionRequest.SourceNodeID != "manager" || stub.createRevisionRequest.ExpectedRevision != 3 {
		t.Fatalf("request body not forwarded: %#v", stub.createRevisionRequest)
	}
}

func TestCompleteFormRevisionTaskRejectsInvalidAction(t *testing.T) {
	stub := &runtimeServiceStub{}
	handler := NewRuntimeHandler(stub)
	c := newUserContext(42)
	c.Params = append(c.Params, param.Param{Key: "id", Value: "revision-task-1"})
	c.Request.SetBodyString(`{"action":"return","comment":"invalid"}`)

	handler.CompleteFormRevisionTask(context.Background(), c)

	if c.Response.StatusCode() != 400 || stub.completeRevisionTaskCalls != 0 {
		t.Fatalf("status=%d calls=%d body=%s", c.Response.StatusCode(), stub.completeRevisionTaskCalls, c.Response.Body())
	}
}

func TestCancelCompletedFormRevisionMapsActiveTaskError(t *testing.T) {
	stub := &runtimeServiceStub{cancelRevisionErr: workflowapp.ErrCompletedRevisionCannotCancel}
	handler := NewRuntimeHandler(stub)
	c := newUserContext(42)
	c.Params = append(c.Params, param.Param{Key: "id", Value: "revision-1"})

	handler.CancelCompletedFormRevision(context.Background(), c)

	if c.Response.StatusCode() != 400 || !strings.Contains(string(c.Response.Body()), "已有确认任务被处理") {
		t.Fatalf("status=%d body=%s", c.Response.StatusCode(), c.Response.Body())
	}
}
