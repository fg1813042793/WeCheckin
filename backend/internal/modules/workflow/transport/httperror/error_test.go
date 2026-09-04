package httperror

import (
	"errors"
	"fmt"
	"testing"

	workflowapp "wecheckin/backend/internal/modules/workflow/application"
	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
)

func TestPublicMessageMapsWrappedWorkflowSentinels(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{name: "required", err: workflowapp.ErrDefinitionRequired, want: "流程定义不能为空"},
		{name: "permission", err: workflowapp.ErrInstanceAccessDenied, want: "无权执行该流程操作"},
		{name: "task actor", err: workflowdomain.ErrTaskActorMismatch, want: "无权执行该流程操作"},
		{name: "state conflict", err: workflowdomain.ErrTaskAlreadyHandled, want: "工作流任务已处理"},
		{name: "wrapped", err: fmt.Errorf("complete task: %w", workflowapp.ErrTaskReturnCommentRequired), want: "退回原因不能为空"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, ok := PublicMessage(test.err)
			if !ok || actual != test.want {
				t.Fatalf("PublicMessage() = (%q, %v), want (%q, true)", actual, ok, test.want)
			}
		})
	}
}

func TestPublicMessageRejectsUnknownErrors(t *testing.T) {
	privateErr := errors.New("SELECT password FROM admins")
	if message, ok := PublicMessage(privateErr); ok || message != "" {
		t.Fatalf("PublicMessage() = (%q, %v), want unknown", message, ok)
	}
}
