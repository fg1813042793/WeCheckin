package admin

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/internal/service"
	"wecheckin-backend/backend/pkg/response"
)

type AdminUserHandler struct{}

func NewAdminUserHandler() *AdminUserHandler { return &AdminUserHandler{} }

func parseUintSlice(s string) []uint {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	ids := make([]uint, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			ids = append(ids, uint(v))
		}
	}
	return ids
}

// @Tags 用户管理
// @Summary 获取用户列表
// @Param page query string false "页码"
// @Param size query string false "每页数量"
// @Param keyword query string false "关键词"
// @Success 200 {object} response.Resp
// @Router /admin/user_list [get]
func (h *AdminUserHandler) GetUserList(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	page, _ := strconv.Atoi(c.Query("page"))
	sizeStr := c.Query("pageSize")
	if sizeStr == "" {
		sizeStr = c.Query("size")
	}
	size, _ := strconv.Atoi(sizeStr)
	keyword := c.Query("keyword")
	sortStr := c.Query("sort")
	list, total, err := service.GetUserList(keyword, sortStr, page, size, admin.ID)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, map[string]interface{}{"list": list, "total": total})
}

// @Tags 用户管理
// @Summary 获取用户详情
// @Param openid query string true "用户OpenID"
// @Success 200 {object} response.Resp
// @Router /admin/user_detail [get]
func (h *AdminUserHandler) GetUserDetail(ctx context.Context, c *app.RequestContext) {
	openID := c.Query("openid")
	data, err := service.GetUserByOpenID(openID)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags 用户管理
// @Summary 根据ID获取用户详情
// @Param id query string true "用户ID"
// @Success 200 {object} response.Resp
// @Router /admin/user_detail_by_id [get]
func (h *AdminUserHandler) GetUserByID(ctx context.Context, c *app.RequestContext) {
	id := c.Query("id")
	data, err := service.GetUserByID(id)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags 用户管理
// @Summary 新增用户
// @Param name formData string true "用户名"
// @Param mobile formData string false "手机号"
// @Param pic formData string false "头像URL"
// @Param forms formData string false "扩展表单数据JSON"
// @Success 200 {object} response.Resp
// @Router /admin/user_add [post]
func (h *AdminUserHandler) AddUser(ctx context.Context, c *app.RequestContext) {
	name := c.PostForm("name")
	mobile := c.PostForm("mobile")
	pic := c.PostForm("pic")
	forms := c.PostForm("forms")
	addIP := c.ClientIP()
	deptIds := parseUintSlice(c.PostForm("deptIds"))
	err := service.AddUser(name, mobile, pic, forms, addIP, deptIds)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.JSON(c, nil)
}

// @Tags 用户管理
// @Summary 编辑用户
// @Param id formData string true "用户ID"
// @Param name formData string false "用户名"
// @Param mobile formData string false "手机号"
// @Param pic formData string false "头像URL"
// @Param forms formData string false "扩展表单数据JSON"
// @Success 200 {object} response.Resp
// @Router /admin/user_edit [post]
func (h *AdminUserHandler) EditUser(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	name := c.PostForm("name")
	mobile := c.PostForm("mobile")
	pic := c.PostForm("pic")
	forms := c.PostForm("forms")
	addIP := c.ClientIP()
	deptIds := parseUintSlice(c.PostForm("deptIds"))
	err := service.EditUser(id, name, mobile, pic, forms, addIP, deptIds)
	if err != nil {
		response.Fail(c, "编辑失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags 用户管理
// @Summary 删除用户
// @Param id formData string true "用户ID"
// @Success 200 {object} response.Resp
// @Router /admin/user_del [post]
func (h *AdminUserHandler) DelUser(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	err := service.DelUser(id)
	if err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}

func (h *AdminUserHandler) DelUsers(ctx context.Context, c *app.RequestContext) {
	idsStr := c.PostForm("ids")
	if idsStr == "" {
		response.Fail(c, "参数错误")
		return
	}
	ids := strings.Split(idsStr, ",")
	if err := service.DelUsers(ids); err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags 用户管理
// @Summary 获取用户表单字段列表
// @Success 200 {object} response.Resp
// @Router /admin/user_form_fields [get]
func (h *AdminUserHandler) GetUserFormFields(ctx context.Context, c *app.RequestContext) {
	list, err := service.GetUserFormFields()
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	if list == nil {
		list = []model.UserFormField{}
	}
	response.JSON(c, list)
}

// @Tags 用户管理
// @Summary 保存用户表单字段配置(全量替换)
// @Param fields formData string true "字段JSON数组"
// @Success 200 {object} response.Resp
// @Router /admin/user_form_field_save [post]
func (h *AdminUserHandler) SaveUserFormFields(ctx context.Context, c *app.RequestContext) {
	fieldsJSON := c.PostForm("fields")
	var fields []model.UserFormField
	if fieldsJSON != "" {
		if err := json.Unmarshal([]byte(fieldsJSON), &fields); err != nil {
			response.Fail(c, "JSON格式错误")
			return
		}
	}
	if err := service.SaveUserFormFields(fields); err != nil {
		response.Fail(c, "保存失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags 用户管理
// @Summary 设置用户状态
// @Param id formData string true "用户ID"
// @Param status formData string true "状态"
// @Success 200 {object} response.Resp
// @Router /admin/user_status [post]
func (h *AdminUserHandler) StatusUser(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	status, _ := strconv.Atoi(c.PostForm("status"))
	reason := c.PostForm("reason")
	err := service.StatusUser(id, status, reason)
	if err != nil {
		response.Fail(c, "操作失败")
		return
	}
	response.JSON(c, nil)
}

func (h *AdminUserHandler) ResetPassword(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	if id == "" {
		response.Fail(c, "参数错误")
		return
	}
	if err := service.ResetUserPassword(id); err != nil {
		response.Fail(c, "重置失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags 用户管理
// @Summary 获取用户数据导出链接
// @Success 200 {object} response.Resp
// @Router /admin/user_data_get [get]
func (h *AdminUserHandler) UserDataGet(ctx context.Context, c *app.RequestContext) {
	data, err := service.GetUserDataURL()
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags 用户管理
// @Summary 导出用户数据
// @Success 200 {object} response.Resp
// @Router /admin/user_data_export [get]
func (h *AdminUserHandler) UserDataExport(ctx context.Context, c *app.RequestContext) {
	data, err := service.ExportUserDataExcel()
	if err != nil {
		response.Fail(c, "导出失败")
		return
	}
	response.JSON(c, data)
}

// @Tags 用户管理
// @Summary 删除用户数据
// @Success 200 {object} response.Resp
// @Router /admin/user_data_del [post]
func (h *AdminUserHandler) UserDataDel(ctx context.Context, c *app.RequestContext) {
	err := service.DeleteUserDataExcel()
	if err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}

func (h *AdminUserHandler) GetOnlineUsers(ctx context.Context, c *app.RequestContext) {
	list, err := service.GetOnlineUsers()
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	if list == nil {
		list = []map[string]interface{}{}
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
	if err := service.ForceOfflineUser(idStr, token); err != nil {
		response.Fail(c, "操作失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags 在线用户
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
		// 也支持 form 数组（兼容老调用）
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
	n, err := service.BatchForceOfflineUser(items)
	if err != nil {
		response.Fail(c, "操作失败")
		return
	}
	response.JSON(c, map[string]int{"count": n})
}
