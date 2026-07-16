package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"wecheckin-backend/backend/internal/app/support/media"
	"wecheckin-backend/backend/pkg/logger"
	"wecheckin-backend/backend/pkg/response"
)

func registerUploadAndStaticRoutes(h *server.Hertz) {
	registerUploadRoutes(h)
	registerAdminSPARoutes(h)
}

func registerUploadRoutes(h *server.Hertz) {
	uploadDir := "./uploads"
	os.MkdirAll(uploadDir, 0755)
	h.POST("/upload", func(ctx context.Context, c *app.RequestContext) {
		file, err := c.FormFile("file")
		if err != nil {
			response.Fail(c, "上传失败: "+err.Error())
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
		thumbRelFile := ""
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" && ext != ".webp" {
			thumbName := filename[:len(filename)-len(ext)] + "_thumb.jpg"
			thumbDst := filepath.Join(saveDir, thumbName)
			if err := exec.Command("ffmpeg", "-y", "-i", dst, "-vframes", "1", "-q:v", "2", thumbDst).Run(); err == nil {
				thumbRelFile = "/uploads/" + dateDir + "/" + thumbName
			}
		}
		response.JSON(c, utils.H{"url": relFile, "thumb": thumbRelFile, "path": filepath.Join(absUpload, relPath), "filename": filename, "domain": media.StaticDomain()})
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
