package upload

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"wecheckin/backend/internal/support/media"
	"wecheckin/backend/internal/support/storage"
	"wecheckin/backend/pkg/response"
)

const maxUploadSize = 20 * 1024 * 1024

var imageUploadExtensions = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
}

var videoUploadExtensions = map[string]bool{
	".mp4": true, ".mov": true, ".avi": true, ".wmv": true, ".flv": true, ".mkv": true,
}

type Handler struct{}

func NewHandler() *Handler { return &Handler{} }

func allowedUploadExtension(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return imageUploadExtensions[ext] || videoUploadExtensions[ext]
}

func validateUploadContent(filename string, content []byte) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	detected := http.DetectContentType(content)
	if imageUploadExtensions[ext] {
		return strings.HasPrefix(detected, "image/")
	}
	if videoUploadExtensions[ext] {
		return strings.HasPrefix(detected, "video/") || detected == "application/octet-stream"
	}
	return false
}

// Upload POST /api/v2/admin/uploads
// @Tags PC端-文件上传
// @Summary 上传后台图片或视频
// @Accept multipart/form-data
// @Param file formData file true "文件"
// @Success 200 {object} response.Resp
func (h *Handler) Upload(ctx context.Context, c *app.RequestContext) {
	file, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, "上传失败: "+err.Error())
		return
	}
	if !allowedUploadExtension(file.Filename) {
		response.Fail(c, "不支持的文件格式，仅允许 jpg/png/gif/webp/mp4/mov/avi/wmv/flv/mkv")
		return
	}
	if file.Size <= 0 || file.Size > maxUploadSize {
		response.Fail(c, "上传文件大小必须在 20MB 以内")
		return
	}

	src, err := file.Open()
	if err != nil {
		response.Fail(c, "上传失败: "+err.Error())
		return
	}
	contentHeader := make([]byte, 512)
	readN, readErr := src.Read(contentHeader)
	_ = src.Close()
	if readErr != nil && readN == 0 {
		response.Fail(c, "无法读取上传文件")
		return
	}
	if !validateUploadContent(file.Filename, contentHeader[:readN]) {
		response.Fail(c, "文件内容与扩展名不匹配")
		return
	}

	now := time.Now()
	ext := strings.ToLower(filepath.Ext(file.Filename))
	filename := fmt.Sprintf("%d_%s", now.UnixNano(), filepath.Base(file.Filename))
	stored, err := storage.SaveMultipartFile(ctx, file, storage.SaveOptions{
		Prefix:   "uploads",
		Filename: filename,
		Now:      now,
	})
	if err != nil {
		response.Fail(c, "上传失败: "+err.Error())
		return
	}

	thumbURL := ""
	if stored.IsLocal && videoUploadExtensions[ext] {
		thumbName := filename[:len(filename)-len(ext)] + "_thumb.jpg"
		thumbPath := filepath.Join(filepath.Dir(stored.LocalPath), thumbName)
		if err := exec.CommandContext(ctx, "ffmpeg", "-y", "-i", stored.LocalPath, "-vframes", "1", "-q:v", "2", thumbPath).Run(); err == nil {
			thumbURL = strings.TrimSuffix(stored.RelativeURL, filename) + thumbName
		}
	}
	response.JSON(c, utils.H{
		"url": stored.URL, "thumb": thumbURL, "path": stored.Path,
		"filename": filename, "domain": media.StaticDomain(),
	})
}
