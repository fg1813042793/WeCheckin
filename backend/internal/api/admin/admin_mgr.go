package admin

import (
	"context"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/internal/service"
	"wecheckin-backend/backend/pkg/response"
)

type AdminMgrHandler struct{}

func NewAdminMgrHandler() *AdminMgrHandler { return &AdminMgrHandler{} }

// @Tags 管理员管理
// @Summary 管理员登录
// @Param name formData string true "用户名"
// @Param password formData string true "密码"
// @Success 200 {object} response.Resp
// @Router /admin/login [post]
func (h *AdminMgrHandler) AdminLogin(ctx context.Context, c *app.RequestContext) {
	name := c.PostForm("name")
	password := c.PostForm("password")
	if password == "" {
		password = c.PostForm("pwd")
	}
	addIP := c.ClientIP()
	device := string(c.UserAgent())
	data, err := service.AdminLogin(name, password, addIP, device)
	if err != nil {
		response.Fail(c, "登录失败")
		return
	}
	response.JSON(c, data)
}

// @Tags 管理员管理
// @Summary 获取管理员列表
// @Success 200 {object} response.Resp
// @Router /admin/mgr_list [get]
func (h *AdminMgrHandler) GetMgrList(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.Query("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	if pageSize < 1 {
		pageSize = 20
	}
	data, err := service.GetMgrList(admin.ID, keyword, page, pageSize)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags 管理员管理
// @Summary 新增管理员
// @Param name formData string true "用户名"
// @Param password formData string true "密码"
// @Param desc formData string false "描述"
// @Param phone formData string false "手机号"
// @Success 200 {object} response.Resp
// @Router /admin/mgr_insert [post]
func (h *AdminMgrHandler) InsertMgr(ctx context.Context, c *app.RequestContext) {
	name := c.PostForm("name")
	password := c.PostForm("password")
	desc := c.PostForm("desc")
	phone := c.PostForm("phone")
	roleID, _ := strconv.Atoi(c.PostForm("roleId"))
	deptIds := parseUintSlice(c.PostForm("deptIds"))
	addIP := c.ClientIP()
	typ := 2 // default admin type
	err := service.InsertMgr(name, password, desc, phone, addIP, typ, uint(roleID), deptIds)
	if err != nil {
		response.Fail(c, "新增失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags 管理员管理
// @Summary 删除管理员
// @Param id formData string true "管理员ID"
// @Success 200 {object} response.Resp
// @Router /admin/mgr_del [post]
func (h *AdminMgrHandler) DelMgr(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	err := service.DelMgr(id)
	if err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}

func (h *AdminMgrHandler) DelMgrs(ctx context.Context, c *app.RequestContext) {
	idsStr := c.PostForm("ids")
	if idsStr == "" {
		response.Fail(c, "参数错误")
		return
	}
	ids := strings.Split(idsStr, ",")
	if err := service.DelMgrs(ids); err != nil {
		response.Fail(c, "删除失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags 管理员管理
// @Summary 获取管理员详情
// @Param id query string true "管理员ID"
// @Success 200 {object} response.Resp
// @Router /admin/mgr_detail [get]
func (h *AdminMgrHandler) GetMgrDetail(ctx context.Context, c *app.RequestContext) {
	id := c.Query("id")
	data, err := service.GetMgrDetail(id)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, data)
}

// @Tags 管理员管理
// @Summary 编辑管理员
// @Param id formData string true "管理员ID"
// @Param name formData string false "用户名"
// @Param desc formData string false "描述"
// @Param phone formData string false "手机号"
// @Success 200 {object} response.Resp
// @Router /admin/mgr_edit [post]
func (h *AdminMgrHandler) EditMgr(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	name := c.PostForm("name")
	desc := c.PostForm("desc")
	pic := c.PostForm("pic")
	phone := c.PostForm("phone")
	password := c.PostForm("password")
	roleID, _ := strconv.Atoi(c.PostForm("roleId"))
	deptIds := parseUintSlice(c.PostForm("deptIds"))
	addIP := c.ClientIP()
	err := service.EditMgr(id, name, desc, pic, phone, password, addIP, uint(roleID), deptIds)
	if err != nil {
		response.Fail(c, "编辑失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags 管理员管理
// @Summary 设置管理员状态
// @Param id formData string true "管理员ID"
// @Param status formData string true "状态"
// @Success 200 {object} response.Resp
// @Router /admin/mgr_status [post]
func (h *AdminMgrHandler) StatusMgr(ctx context.Context, c *app.RequestContext) {
	id := c.PostForm("id")
	status, _ := strconv.Atoi(c.PostForm("status"))
	err := service.StatusMgr(id, status)
	if err != nil {
		response.Fail(c, "操作失败")
		return
	}
	response.JSON(c, nil)
}

// @Tags 管理员管理
// @Summary 修改管理员密码
// @Param id formData string true "管理员ID"
// @Param password formData string true "新密码"
// @Success 200 {object} response.Resp
// @Router /admin/mgr_pwd [post]
func (h *AdminMgrHandler) PwdMgr(ctx context.Context, c *app.RequestContext) {
	adminVal, exists := c.Get("admin")
	if !exists {
		response.Fail(c, "未登录")
		return
	}
	admin := adminVal.(*model.Admin)

	oldPassword := c.PostForm("oldPassword")
	password := c.PostForm("password")
	err := service.PwdMgr(strconv.Itoa(int(admin.ID)), oldPassword, password)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.JSON(c, nil)
}

// @Tags 管理员管理
// @Summary 获取操作日志列表
// @Success 200 {object} response.Resp
// @Router /admin/log_list [get]
func (h *AdminMgrHandler) GetLogList(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	keyword := c.Query("search")
	page, _ := strconv.Atoi(c.Query("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	if pageSize < 1 {
		pageSize = 20
	}
	list, total, err := service.GetLogList(keyword, page, pageSize, admin.ID)
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.JSON(c, map[string]interface{}{"list": list, "total": total})
}

// @Tags 管理员管理
// @Summary 清除操作日志
// @Success 200 {object} response.Resp
// @Router /admin/log_clear [get]
func (h *AdminMgrHandler) ClearLog(ctx context.Context, c *app.RequestContext) {
	err := service.ClearLog()
	if err != nil {
		response.Fail(c, "操作失败")
		return
	}
	response.JSON(c, nil)
}

func (h *AdminMgrHandler) AdminLogout(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	service.AdminLogout(admin.ID)
	response.JSON(c, nil)
}

func (h *AdminMgrHandler) GetOnlineAdmins(ctx context.Context, c *app.RequestContext) {
	list, err := service.GetOnlineAdmins()
	if err != nil {
		response.Fail(c, "获取失败")
		return
	}
	if list == nil {
		list = []map[string]interface{}{}
	}
	response.JSON(c, list)
}

func (h *AdminMgrHandler) ForceOfflineAdmin(ctx context.Context, c *app.RequestContext) {
	idStr := c.PostForm("id")
	if idStr == "" {
		response.Fail(c, "参数错误")
		return
	}
	if err := service.ForceOfflineAdmin(idStr); err != nil {
		response.Fail(c, "操作失败")
		return
	}
	response.JSON(c, nil)
}
