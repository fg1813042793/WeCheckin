package common

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"wecheckin/backend/internal/support/media"
	"wecheckin/backend/internal/support/storage"
	"wecheckin/backend/pkg/logger"
	"wecheckin/backend/pkg/response"
)

func RegisterUploadAndStatic(h *server.Hertz) {
	registerUploadRoutes(h)
	registerAdminSPARoutes(h)
}

func registerUploadRoutes(h *server.Hertz) {
	uploadDir := storage.LocalUploadRoot()
	os.MkdirAll(uploadDir, 0755)
	h.POST("/upload", func(ctx context.Context, c *app.RequestContext) {
		file, err := c.FormFile("file")
		if err != nil {
			response.FailInternal(ctx, c, "common.upload.form_file", "上传失败，请稍后重试", err)
			return
		}
		ext := strings.ToLower(filepath.Ext(file.Filename))
		if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" || ext == ".webp" {
			// image - continue
		} else if ext == ".mp4" || ext == ".mov" || ext == ".avi" || ext == ".wmv" || ext == ".flv" || ext == ".mkv" {
			// video - continue
		} else {
			response.Fail(c, "不支持的文件格式，仅允许 jpg/png/gif/webp/mp4/mov/avi/wmv/flv/mkv")
			return
		}
		if file.Size > 20*1024*1024 {
			response.Fail(c, "上传文件过大，最大20MB")
			return
		}
		now := time.Now()
		filename := fmt.Sprintf("%d_%s", now.UnixNano(), filepath.Base(file.Filename))
		stored, err := storage.SaveMultipartFile(ctx, file, storage.SaveOptions{
			Prefix:   "uploads",
			Filename: filename,
			Now:      now,
		})
		if err != nil {
			response.FailInternal(ctx, c, "common.upload.store", "上传失败，请稍后重试", err)
			return
		}
		thumbRelFile := ""
		if stored.IsLocal && ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" && ext != ".webp" {
			thumbName := filename[:len(filename)-len(ext)] + "_thumb.jpg"
			thumbDst := filepath.Join(filepath.Dir(stored.LocalPath), thumbName)
			if err := exec.Command("ffmpeg", "-y", "-i", stored.LocalPath, "-vframes", "1", "-q:v", "2", thumbDst).Run(); err == nil {
				thumbRelFile = strings.TrimSuffix(stored.RelativeURL, filename) + thumbName
			}
		}
		response.JSON(c, utils.H{"url": stored.URL, "thumb": thumbRelFile, "path": stored.Path, "filename": filename, "domain": media.StaticDomain()})
	})
	absUpload, _ := filepath.Abs(uploadDir)
	h.GET("/uploads/*filepath", func(ctx context.Context, c *app.RequestContext) {
		c.File(filepath.Join(absUpload, c.Param("filepath")))
	})
}

func registerAdminSPARoutes(h *server.Hertz) {
	adminDist := "../admin/dist"
	if _, err := os.Stat(adminDist); err == nil {
		h.GET("/*any", func(ctx context.Context, c *app.RequestContext) {
			path := string(c.Path())
			if strings.HasPrefix(path, "/admin/") || strings.HasPrefix(path, "/exam/") || strings.HasPrefix(path, "/survey/") || strings.HasPrefix(path, "/passport/") || strings.HasPrefix(path, "/upload/") || strings.HasPrefix(path, "/home/") {
				c.String(404, "Not Found")
				return
			}
			filePath := filepath.Join(adminDist, path)
			if fi, err := os.Stat(filePath); err == nil && !fi.IsDir() {
				c.File(filePath)
			} else {
				c.File(filepath.Join(adminDist, "index.html"))
			}
		})
		logger.Logger.Println("Admin SPA serving from", adminDist)
	}
}
