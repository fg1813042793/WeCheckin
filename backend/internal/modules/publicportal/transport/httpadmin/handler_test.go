package httpadmin

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route/param"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/modules/publicportal/application"
)

type serviceFake struct {
	Service
	calls int
	actor uint
	input application.ItemInput
	query application.OptionsQuery
	err   error
}

func (f *serviceFake) CreateItem(_ context.Context, id uint, in application.ItemInput) (application.AdminItem, error) {
	f.calls++
	f.actor = id
	f.input = in
	return application.AdminItem{}, f.err
}
func (f *serviceFake) Get(context.Context, uint) (application.AdminResponse, error) {
	f.calls++
	return application.AdminResponse{}, f.err
}
func (f *serviceFake) Options(_ context.Context, id uint, q application.OptionsQuery) (application.OptionsResponse, error) {
	f.calls++
	f.actor = id
	f.query = q
	return application.OptionsResponse{}, f.err
}
func (f *serviceFake) DeleteItem(context.Context, uint, uint64, uint64) (application.DeleteResponse, error) {
	f.calls++
	return application.DeleteResponse{}, f.err
}
func (f *serviceFake) UpdateItem(context.Context, uint, uint64, application.ItemUpdate) (application.AdminItem, error) {
	f.calls++
	return application.AdminItem{}, f.err
}
func (f *serviceFake) SaveSettings(context.Context, uint, application.Settings) (application.Settings, error) {
	f.calls++
	return application.Settings{}, f.err
}
func (f *serviceFake) Preview(context.Context, uint) (application.PublicResponse, error) {
	f.calls++
	return application.PublicResponse{Enabled: true, Items: []application.PublicItem{}}, f.err
}
func adminContext() *app.RequestContext {
	c := app.NewContext(0)
	c.Set("admin", &model.Admin{ID: 7})
	return c
}
func code(t *testing.T, c *app.RequestContext) int {
	t.Helper()
	var v struct {
		Code int `json:"code"`
	}
	if err := json.Unmarshal(c.Response.Body(), &v); err != nil {
		t.Fatalf("invalid response: %s", c.Response.Body())
	}
	return v.Code
}
func TestStrictBodyDoesNotReachService(t *testing.T) {
	for _, body := range []string{"", "null", "[]", `{"kind":"survey","sourceId":1,"adminId":999}`, `{"kind":"survey","sourceId":1} {}`, `{"kind":"survey","sourceId":-1}`, `{"kind":"survey","sourceId":1,"version":1}`} {
		f := &serviceFake{}
		c := adminContext()
		c.Request.SetBodyString(body)
		NewHandler(f).CreateItem(context.Background(), c)
		if f.calls != 0 || code(t, c) == 0 {
			t.Fatalf("accepted %q: calls=%d resp=%s", body, f.calls, c.Response.Body())
		}
	}
}
func TestCreateUsesAuthenticatedIdentity(t *testing.T) {
	f := &serviceFake{}
	c := adminContext()
	c.Request.SetRequestURI("/?adminId=999")
	c.Request.SetBodyString(`{"kind":"survey","sourceId":12,"enabled":true,"sort":1}`)
	NewHandler(f).CreateItem(context.Background(), c)
	if f.calls != 1 || f.actor != 7 || f.input.SourceID != 12 || code(t, c) != 0 {
		t.Fatalf("fake=%+v response=%s", f, c.Response.Body())
	}
}
func TestStrictPositivePathVersionAndPagination(t *testing.T) {
	for _, id := range []string{"0", "-1", "+1", " 1", "1x", "18446744073709551616"} {
		f := &serviceFake{}
		c := adminContext()
		c.Params = param.Params{{Key: "id", Value: id}}
		c.Request.SetRequestURI("/?version=1")
		NewHandler(f).DeleteItem(context.Background(), c)
		if f.calls != 0 || code(t, c) == 0 {
			t.Fatalf("id accepted=%q", id)
		}
	}
	for _, query := range []string{"version=0", "version=-1", "version=abc", "version=1&version=2", ""} {
		f := &serviceFake{}
		c := adminContext()
		c.Params = param.Params{{Key: "id", Value: "1"}}
		c.Request.SetRequestURI("/?" + query)
		NewHandler(f).DeleteItem(context.Background(), c)
		if f.calls != 0 || code(t, c) == 0 {
			t.Fatalf("version accepted=%q", query)
		}
	}
	for _, query := range []string{"kind=bad", "kind=survey&page=0", "kind=survey&page=+1", "kind=survey&pageSize=101", "kind=survey&pageSize=1.1", "kind=survey&page=1&page=2"} {
		f := &serviceFake{}
		c := adminContext()
		c.Request.SetRequestURI("/?" + query)
		NewHandler(f).Options(context.Background(), c)
		if f.calls != 0 || code(t, c) == 0 {
			t.Fatalf("query accepted=%q", query)
		}
	}
	f := &serviceFake{}
	c := adminContext()
	c.Request.SetRequestURI("/?kind=survey")
	NewHandler(f).Options(context.Background(), c)
	if f.calls != 1 || f.query.Page != 1 || f.query.PageSize != 20 {
		t.Fatalf("defaults=%+v", f)
	}
}
func TestErrorsAreSafeAndAdminRequired(t *testing.T) {
	for _, err := range []error{application.ErrVersionConflict, application.ErrSchemaMissing, errors.New("secret SQL SELECT credentials")} {
		f := &serviceFake{err: err}
		c := adminContext()
		NewHandler(f).Get(context.Background(), c)
		if code(t, c) == 0 || strings.Contains(string(c.Response.Body()), "credentials") {
			t.Fatalf("unsafe response=%s", c.Response.Body())
		}
		if err == application.ErrSchemaMissing && !strings.Contains(string(c.Response.Body()), "backend/init.sh") {
			t.Fatal("missing actionable message")
		}
	}
	f := &serviceFake{}
	c := app.NewContext(0)
	NewHandler(f).Get(context.Background(), c)
	if f.calls != 0 || code(t, c) == 0 {
		t.Fatal("anonymous admin")
	}
	c = adminContext()
	NewHandler(nil).Get(context.Background(), c)
	if code(t, c) == 0 {
		t.Fatal("nil service accepted")
	}
}
func TestPreviewNoStore(t *testing.T) {
	c := adminContext()
	NewHandler(&serviceFake{}).Preview(context.Background(), c)
	if string(c.Response.Header.Peek("Cache-Control")) != "no-store" || code(t, c) != 0 {
		t.Fatalf("preview=%s", c.Response.Body())
	}
}
