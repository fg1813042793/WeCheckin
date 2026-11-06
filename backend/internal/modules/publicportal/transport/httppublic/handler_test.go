package httppublic

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"

	"wecheckin/backend/internal/modules/publicportal/application"
)

type serviceFake struct {
	calls int
	err   error
}

func (f *serviceFake) List(context.Context) (application.PublicResponse, error) {
	f.calls++
	return application.PublicResponse{Enabled: false, Items: []application.PublicItem{}}, f.err
}
func TestAnonymousListIgnoresPreviewAndNeverCaches(t *testing.T) {
	f := &serviceFake{}
	c := app.NewContext(0)
	c.Request.SetRequestURI("/?preview=1")
	NewHandler(f).List(context.Background(), c)
	if f.calls != 1 || !strings.Contains(string(c.Response.Body()), `"enabled":false`) || string(c.Response.Header.Peek("Cache-Control")) != "no-store" {
		t.Fatalf("response=%s", c.Response.Body())
	}
}
func TestPublicErrorsNeverLeakSchemaOrInternalDetails(t *testing.T) {
	for _, err := range []error{application.ErrSchemaMissing, errors.New("SQL secret")} {
		c := app.NewContext(0)
		NewHandler(&serviceFake{err: err}).List(context.Background(), c)
		body := string(c.Response.Body())
		if !strings.Contains(body, "公共页面暂不可用") || strings.Contains(body, "backend/") || strings.Contains(body, "SQL") {
			t.Fatalf("unsafe response: %s", body)
		}
	}
}
