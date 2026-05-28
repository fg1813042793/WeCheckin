package api

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin-backend/backend/internal/service"
	"wecheckin-backend/backend/pkg/response"
)

type FavHandler struct{}

func NewFavHandler() *FavHandler { return &FavHandler{} }

// @Tags 收藏
// @Summary 更新收藏
// @Param title formData string true "标题"
// @Param oid formData string true "对象ID"
// @Param typ formData string true "类型"
// @Param path formData string false "路径"
// @Param user_id formData string false "用户ID"
// @Success 200 {object} response.Resp
// @Router /fav/update [post]
func (h *FavHandler) UpdateFav(ctx context.Context, c *app.RequestContext) {
	title := c.PostForm("title")
	oid := c.PostForm("oid")
	typ := c.PostForm("typ")
	path := c.PostForm("path")
	userID := c.PostForm("user_id")
	addIP := c.ClientIP()
	err := service.UpdateFav(userID, title, typ, oid, path, addIP)
	if err != nil {
		response.Fail(c, "操作失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags 收藏
// @Summary 删除收藏
// @Param oid formData string true "对象ID"
// @Param user_id formData string false "用户ID"
// @Success 200 {object} response.Resp
// @Router /fav/del [post]
func (h *FavHandler) DelFav(ctx context.Context, c *app.RequestContext) {
	oid := c.PostForm("oid")
	userID := c.PostForm("user_id")
	err := service.DelFav(userID, oid)
	if err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags 收藏
// @Summary 是否已收藏
// @Param oid query string true "对象ID"
// @Param typ query string true "类型"
// @Param user_id query string false "用户ID"
// @Success 200 {object} response.Resp
// @Router /fav/is_fav [get]
func (h *FavHandler) IsFav(ctx context.Context, c *app.RequestContext) {
	oid := c.Query("oid")
	userID := c.Query("user_id")
	data, err := service.IsFav(userID, oid)
	if err != nil {
		response.Fail(c, "查询失败")
		return
	}
	response.JSON(c, data)
}

// @Tags 收藏
// @Summary 获取我的收藏列表
// @Param typ query string false "类型"
// @Param user_id query string false "用户ID"
// @Success 200 {object} response.Resp
// @Router /fav/my_list [get]
func (h *FavHandler) GetMyFavList(ctx context.Context, c *app.RequestContext) {
	userID := c.Query("user_id")
	data, err := service.GetMyFavList(userID)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}
