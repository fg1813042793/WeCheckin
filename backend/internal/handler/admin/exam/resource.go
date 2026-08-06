package exam

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
	formkitadminservice "wecheckin/backend/internal/service/admin/formkitadmin"
	"wecheckin/backend/internal/support/media"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/response"
)

// @Tags PC端-考试管理
// @Summary 上传考试资源
// @Param file formData file true "文件"
// @Param examId formData int true "考试ID"
// @Param resType formData string true "资源类型: bg/header"
// @Success 200 {object} response.Resp
func (h *AdminExamHandler) ResourceUpload(ctx context.Context, c *app.RequestContext) {
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
	examID, _ := strconv.Atoi(string(c.FormValue("examId")))
	resType := string(c.FormValue("resType"))
	if resType != "bg" && resType != "header" {
		response.Fail(c, "无效的资源类型")
		return
	}
	if examID <= 0 {
		response.Fail(c, "无效的考试ID")
		return
	}
	if _, err := h.svc.GetForAdminContext(ctx, uint(examID), admin.ID); err != nil {
		response.Fail(c, "无效的考试ID")
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
	data, err := formkitadminservice.CreateExamResourceContext(ctx, formkitadminservice.ResourceInput{
		OwnerID:  uint(examID),
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

// @Tags PC端-考试管理
// @Summary 查询考试资源列表
// @Param examId query int true "考试ID"
// @Param resType query string false "资源类型: bg/header"
// @Success 200 {object} response.Resp
func (h *AdminExamHandler) ResourceList(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	examID, _ := strconv.Atoi(c.Query("examId"))
	resType := c.Query("resType")
	if examID <= 0 {
		response.JSON(c, []formkitadminservice.ResourceResult{})
		return
	}
	if _, err := h.svc.GetForAdminContext(ctx, uint(examID), admin.ID); err != nil {
		response.JSON(c, []formkitadminservice.ResourceResult{})
		return
	}
	list, err := formkitadminservice.ListExamResourcesContext(ctx, uint(examID), resType)
	if err != nil {
		response.Fail(c, "查询失败")
		return
	}
	response.JSON(c, list)
}

// @Tags PC端-考试管理
// @Summary 删除考试资源
// @Param id formData int true "资源ID"
// @Success 200 {object} response.Resp
func (h *AdminExamHandler) ResourceDelete(ctx context.Context, c *app.RequestContext) {
	adminVal, _ := c.Get("admin")
	admin := adminVal.(*model.Admin)
	id, _ := strconv.Atoi(string(c.FormValue("id")))
	if id <= 0 {
		response.Fail(c, "无效的资源ID")
		return
	}
	res, err := formkitadminservice.DeleteExamResourceForAdminContext(ctx, uint(id), admin.ID)
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
