package httperror

import (
	"errors"
	"fmt"
	"testing"

	"github.com/cloudwego/hertz/pkg/protocol/consts"

	workflowapp "wecheckin/backend/internal/modules/workflow/application"
	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
)

func TestFormRevisionHTTPStatus(t *testing.T) {
	tests := []struct {
		err  error
		want int
	}{
		{err: workflowapp.ErrCompletedRevisionNoReviewPath, want: consts.StatusBadRequest},
		{err: workflowapp.ErrCompletedRevisionNotAllowed, want: consts.StatusForbidden},
		{err: workflowapp.ErrCompletedRevisionActive, want: consts.StatusConflict},
		{err: workflowdomain.ErrRevisionTaskNotFound, want: consts.StatusNotFound},
		{err: fmt.Errorf("wrapped: %w", workflowdomain.ErrRevisionTaskAlreadyHandled), want: consts.StatusConflict},
	}
	for _, test := range tests {
		if actual := HTTPStatus(test.err); actual != test.want {
			t.Fatalf("HTTPStatus(%v) = %d, want %d", test.err, actual, test.want)
		}
	}
}

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
