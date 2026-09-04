package httpadmin

import (
	"context"
	"strings"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"

	"wecheckin/backend/internal/modules/scheduledtask/application"
)

func TestRespondMapsSystemTaskReadOnlyError(t *testing.T) {
	c := app.NewContext(0)
	respond(context.Background(), c, nil, application.ErrSystemTaskReadOnly)
	body := string(c.Response.Body())
	if !strings.Contains(body, "系统定时任务不可编辑") || strings.Contains(body, application.ErrSystemTaskReadOnly.Error()) {
		t.Fatalf("response = %s", body)
	}
}
