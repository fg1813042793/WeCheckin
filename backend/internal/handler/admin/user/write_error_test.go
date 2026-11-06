package user

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin/backend/internal/model"
	adminuserservice "wecheckin/backend/internal/service/admin/adminuser"
	"wecheckin/backend/internal/support/permissiondelegation"
)

func TestUserWriteErrorPreservesBusinessReasonAndHidesInternalError(t *testing.T) {
	for _, test := range []struct {
		err  error
		want string
	}{
		{adminuserservice.ErrUserNameTaken, "用户名已存在，请更换后重试"},
		{permissiondelegation.ErrRoleExceeded, "所选角色或目标用户现有角色不存在、已停用或超出当前管理员可授权范围"},
		{errors.New("private_db password=SECRET"), "操作失败，请稍后重试"},
	} {
		c := app.NewContext(0)
		respondUserWriteError(context.Background(), c, "admin.user.edit", test.err)
		assertUserWriteErrorResponse(t, c, test.want)
	}
}

func TestEditUserRejectsMissingNameWithActionableMessage(t *testing.T) {
	c := app.NewContext(0)
	c.Set("admin", &model.Admin{ID: 7})
	c.Request.SetMethod("PUT")
	c.Request.Header.SetContentTypeBytes([]byte("application/x-www-form-urlencoded"))
	c.Request.SetBodyString(url.Values{"id": {"178"}, "name": {"  "}}.Encode())
	NewAdminUserHandler().EditUser(context.Background(), c)
	assertUserWriteErrorResponse(t, c, "请输入用户名")
}

func assertUserWriteErrorResponse(t *testing.T, c *app.RequestContext, want string) {
	t.Helper()
	var body struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(c.Response.Body(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Code != 1 || body.Msg != want || strings.Contains(body.Msg, "SECRET") {
		t.Fatalf("response=%s", c.Response.Body())
	}
}
