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
	"wecheckin-backend/backend/internal/app/support/media"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
	"wecheckin-backend/backend/pkg/response"
)

// @Tags PC端-考试管理
// @Summary 上传考试资源
// @Param file formData file true "文件"
// @Param examId formData int true "考试ID"
// @Param resType formData string true "资源类型: bg/header"
// @Success 200 {object} response.Resp
// @Router /admin/exam/resource_upload [post]
func (h *AdminExamHandler) ResourceUpload(_ context.Context, c *app.RequestContext) {
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
	res := model.ExamResource{
		ExamID:   uint(examID),
		Type:     resType,
		URL:      relFile,
		Filename: filename,
		Path:     filepath.Join(absUpload, relPath),
		Domain:   domain,
		AddTime:  now.UnixMilli(),
	}
	if err := database.DB.Create(&res).Error; err != nil {
		response.Fail(c, "保存记录失败: "+err.Error())
		return
	}
	response.JSON(c, map[string]any{
		"id":       res.ID,
		"url":      relFile,
		"filename": filename,
		"path":     filepath.Join(absUpload, relPath),
		"domain":   domain,
		"type":     resType,
	})
}

// @Tags PC端-考试管理
// @Summary 查询考试资源列表
// @Param examId query int true "考试ID"
// @Param resType query string false "资源类型: bg/header"
// @Success 200 {object} response.Resp
// @Router /admin/exam/resource_list [get]
func (h *AdminExamHandler) ResourceList(_ context.Context, c *app.RequestContext) {
	examID, _ := strconv.Atoi(c.Query("examId"))
	resType := c.Query("resType")
	if examID <= 0 {
		response.JSON(c, []model.ExamResource{})
		return
	}
	query := database.DB.Where("`exam_res_exam_id` = ?", examID)
	if resType != "" {
		query = query.Where("`exam_res_type` = ?", resType)
	}
	var list []model.ExamResource
	if err := query.Order("`exam_res_add_time` DESC").Find(&list).Error; err != nil {
		response.Fail(c, "查询失败")
		return
	}
	response.JSON(c, list)
}

// @Tags PC端-考试管理
// @Summary 删除考试资源
// @Param id formData int true "资源ID"
// @Success 200 {object} response.Resp
// @Router /admin/exam/resource_delete [post]
func (h *AdminExamHandler) ResourceDelete(_ context.Context, c *app.RequestContext) {
	id, _ := strconv.Atoi(string(c.FormValue("id")))
	if id <= 0 {
		response.Fail(c, "无效的资源ID")
		return
	}
	var res model.ExamResource
	if err := database.DB.First(&res, id).Error; err != nil {
		response.Fail(c, "资源不存在")
		return
	}
	if res.Path != "" {
		os.Remove(res.Path)
	}
	if err := database.DB.Delete(&res).Error; err != nil {
		response.Fail(c, "删除失败: "+err.Error())
		return
	}
	response.JSON(c, nil)
}
