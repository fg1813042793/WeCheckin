package survey

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	formkitadminservice "wecheckin-backend/backend/internal/app/service/formkitadmin"
	"wecheckin-backend/backend/internal/app/support/media"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/response"
)

// ResourceUpload POST /admin/survey/resource_upload
// @Tags PC端-问卷管理
// @Summary 上传问卷资源（背景图/页眉图）
// @Param file formData file true "文件"
// @Param surveyId formData int true "问卷ID"
// @Param resType formData string true "资源类型: bg/header"
// @Success 200 {object} response.Resp
// @Router /admin/survey/resource_upload [post]
func (h *AdminSurveyHandler) ResourceUpload(ctx context.Context, c *app.RequestContext) {
	h.lazyInit()
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	file, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, "上传失败: "+err.Error())
		return
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" && ext != ".webp" {
		response.Fail(c, "不支持的文件格式，仅允许 jpg/png/gif/webp")
		return
	}
	if file.Size > 20*1024*1024 {
		response.Fail(c, "上传文件过大，最大20MB")
		return
	}
	surveyID, _ := strconv.Atoi(string(c.FormValue("surveyId")))
	resType := string(c.FormValue("resType"))
	if resType != "bg" && resType != "header" {
		response.Fail(c, "无效的资源类型")
		return
	}
	if surveyID <= 0 {
		response.Fail(c, "无效的问卷ID")
		return
	}
	if _, err := h.survey.GetForAdminContext(ctx, uint(surveyID), admin.ID); err != nil {
		response.Fail(c, "无效的问卷ID")
		return
	}

	uploadDir := "./uploads"
	now := time.Now()
	dateDir := now.Format("2006/01/02")
	saveDir := filepath.Join(uploadDir, dateDir)
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		response.Fail(c, "创建目录失败")
		return
	}
	filename := fmt.Sprintf("%d_%s", now.UnixNano(), filepath.Base(file.Filename))
	dst := filepath.Join(saveDir, filename)

	src, err := file.Open()
	if err != nil {
		response.Fail(c, "上传失败")
		return
	}
	defer src.Close()
	out, err := os.Create(dst)
	if err != nil {
		response.Fail(c, "上传失败")
		return
	}
	defer out.Close()
	if _, err := io.Copy(out, src); err != nil {
		response.Fail(c, "上传失败")
		return
	}

	relPath := dateDir + "/" + filename
	absUpload, _ := filepath.Abs(uploadDir)
	relFile := "/uploads/" + relPath

	domain := media.StaticDomain()
	data, err := formkitadminservice.CreateSurveyResourceContext(ctx, formkitadminservice.ResourceInput{
		OwnerID:  uint(surveyID),
		Type:     resType,
		URL:      relFile,
		Filename: filename,
		Path:     filepath.Join(absUpload, relPath),
		Domain:   domain,
		AddTime:  now.UnixMilli(),
	})
	if err != nil {
		_ = os.Remove(dst)
		response.Fail(c, "保存记录失败: "+err.Error())
		return
	}
	response.JSON(c, data)
}

// ResourceList GET /admin/survey/resource_list
// @Tags PC端-问卷管理
// @Summary 查询问卷资源列表
// @Param surveyId query int true "问卷ID"
// @Param resType query string false "资源类型: bg/header，为空则返回全部"
// @Success 200 {object} response.Resp
// @Router /admin/survey/resource_list [get]
func (h *AdminSurveyHandler) ResourceList(ctx context.Context, c *app.RequestContext) {
	h.lazyInit()
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	surveyID, _ := strconv.Atoi(c.Query("surveyId"))
	resType := c.Query("resType")
	if surveyID <= 0 {
		response.Fail(c, "无效的问卷ID")
		return
	}
	if _, err := h.survey.GetForAdminContext(ctx, uint(surveyID), admin.ID); err != nil {
		response.Fail(c, "无效的问卷ID")
		return
	}
	list, err := formkitadminservice.ListSurveyResourcesContext(ctx, uint(surveyID), resType)
	if err != nil {
		response.Fail(c, "查询失败")
		return
	}
	response.JSON(c, list)
}

// ResourceDelete POST /admin/survey/resource_delete
// @Tags PC端-问卷管理
// @Summary 删除问卷资源
// @Param id formData int true "资源ID"
// @Success 200 {object} response.Resp
// @Router /admin/survey/resource_delete [post]
func (h *AdminSurveyHandler) ResourceDelete(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	id, _ := strconv.Atoi(string(c.FormValue("id")))
	if id <= 0 {
		response.Fail(c, "无效的资源ID")
		return
	}
	res, err := formkitadminservice.DeleteSurveyResourceForAdminContext(ctx, uint(id), admin.ID)
	if formkitadminservice.IsNotFound(err) {
		response.Fail(c, "资源不存在")
		return
	}
	if err != nil {
		response.Fail(c, "删除失败: "+err.Error())
		return
	}
	if res.Path != "" {
		_ = os.Remove(res.Path)
	}
	response.JSON(c, nil)
}
