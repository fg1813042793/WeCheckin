package user

import (
	"context"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	onlineservice "wecheckin-backend/backend/internal/app/service/online"
	"wecheckin-backend/backend/pkg/response"
)

func (h *AdminUserHandler) GetOnlineUsers(ctx context.Context, c *app.RequestContext) {
	list, err := onlineservice.GetOnlineUsersContext(ctx)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	if list == nil {
		list = []onlineservice.UserSession{}
	}
	response.JSON(c, list)
}

func (h *AdminUserHandler) ForceOfflineUser(ctx context.Context, c *app.RequestContext) {
	idStr := c.PostForm("id")
	token := c.PostForm("token")
	if idStr == "" {
		response.Fail(c, "参数错误")
		return
	}
	if err := onlineservice.ForceOfflineUserContext(ctx, idStr, token); err != nil {
		response.Fail(c, "操作失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags PC端-在线用户
// @Summary 批量强制下线
// @Param items body []object{idStr,token} true "items: [{idStr,token}, ...]"
// @Success 200 {object} response.Resp
// @Router /admin/user/batch_force_offline [post]
func (h *AdminUserHandler) BatchForceOfflineUser(ctx context.Context, c *app.RequestContext) {
	var items []struct {
		IDStr string `json:"idStr"`
		Token string `json:"token"`
	}
	if err := c.BindAndValidate(&items); err != nil || len(items) == 0 {
		ids := c.PostForm("ids")
		tokens := c.PostForm("tokens")
		if ids != "" && tokens != "" {
			for i, id := range strings.Split(ids, ",") {
				ts := strings.Split(tokens, ",")
				if i < len(ts) {
					items = append(items, struct {
						IDStr string `json:"idStr"`
						Token string `json:"token"`
					}{id, ts[i]})
				}
			}
		}
	}
	if len(items) == 0 {
		response.Fail(c, "参数错误")
		return
	}
	n, err := onlineservice.BatchForceOfflineUserContext(ctx, items)
	if err != nil {
		response.Fail(c, "操作失败")
		return
	}
	response.JSON(c, map[string]int{"count": n})
}
