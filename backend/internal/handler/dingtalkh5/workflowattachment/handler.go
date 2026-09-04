package workflowattachment

import (
	"bytes"
	"context"
	"fmt"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/cloudwego/hertz/pkg/app"

	"wecheckin/backend/internal/support/media"
	"wecheckin/backend/internal/support/storage"
	"wecheckin/backend/pkg/response"
)

const maxAttachmentSize = 20 * 1024 * 1024

var attachmentExtensions = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
	".pdf": true, ".txt": true, ".csv": true,
	".doc": true, ".docx": true, ".xls": true, ".xlsx": true, ".ppt": true, ".pptx": true,
	".zip": true, ".rar": true, ".7z": true,
}

var imageAttachmentExtensions = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
}

var openXMLAttachmentExtensions = map[string]bool{
	".docx": true, ".xlsx": true, ".pptx": true, ".zip": true,
}

var legacyOfficeAttachmentExtensions = map[string]bool{
	".doc": true, ".xls": true, ".ppt": true,
}

type Handler struct{}

type UploadResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	MimeType string `json:"mimeType"`
	Size     int64  `json:"size"`
}

func NewHandler() *Handler { return &Handler{} }

func (h *Handler) Upload(ctx context.Context, c *app.RequestContext) {
	file, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, "上传失败: "+err.Error())
		return
	}
	originalName := attachmentBaseName(file.Filename)
	if originalName == "" || len([]rune(originalName)) > 255 {
		response.Fail(c, "附件名称不能为空且不能超过255个字符")
		return
	}
	if !allowedAttachmentExtension(originalName) {
		response.Fail(c, "不支持的附件格式")
		return
	}
	if file.Size <= 0 || file.Size > maxAttachmentSize {
		response.Fail(c, "附件大小必须在20MB以内")
		return
	}

	src, err := file.Open()
	if err != nil {
		response.Fail(c, "上传失败: "+err.Error())
		return
	}
	header := make([]byte, 512)
	readN, readErr := src.Read(header)
	_ = src.Close()
	if readErr != nil && readN == 0 {
		response.Fail(c, "无法读取附件内容")
		return
	}
	header = header[:readN]
	if !validateAttachmentContent(originalName, header) {
		response.Fail(c, "附件内容与扩展名不匹配")
		return
	}

	now := time.Now()
	ext := strings.ToLower(filepath.Ext(originalName))
	stored, err := storage.SaveMultipartFile(ctx, file, storage.SaveOptions{
		Prefix:   "uploads/workflow",
		Filename: fmt.Sprintf("%d%s", now.UnixNano(), ext),
		Now:      now,
	})
	if err != nil {
		response.Fail(c, "上传失败: "+err.Error())
		return
	}

	response.JSON(c, UploadResponse{
		ID:       stored.ObjectKey,
		Name:     originalName,
		URL:      attachmentPublicURL(ctx, c, stored.RelativeURL),
		MimeType: attachmentContentType(originalName, header),
		Size:     file.Size,
	})
}

func attachmentBaseName(filename string) string {
	filename = strings.ReplaceAll(strings.TrimSpace(filename), "\\", "/")
	return filepath.Base(filename)
}

func allowedAttachmentExtension(filename string) bool {
	return attachmentExtensions[strings.ToLower(filepath.Ext(filename))]
}

func validateAttachmentContent(filename string, content []byte) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	detected := http.DetectContentType(content)
	switch {
	case imageAttachmentExtensions[ext]:
		return strings.HasPrefix(detected, "image/")
	case ext == ".pdf":
		return bytes.HasPrefix(content, []byte("%PDF-"))
	case ext == ".txt" || ext == ".csv":
		return strings.HasPrefix(detected, "text/plain") || (utf8.Valid(content) && !bytes.ContainsRune(content, '\x00'))
	case openXMLAttachmentExtensions[ext]:
		return bytes.HasPrefix(content, []byte{'P', 'K'})
	case legacyOfficeAttachmentExtensions[ext]:
		return bytes.HasPrefix(content, []byte{0xd0, 0xcf, 0x11, 0xe0})
	case ext == ".rar":
		return bytes.HasPrefix(content, []byte("Rar!"))
	case ext == ".7z":
		return bytes.HasPrefix(content, []byte{'7', 'z', 0xbc, 0xaf, 0x27, 0x1c})
	default:
		return false
	}
}

func attachmentContentType(filename string, content []byte) string {
	if value := mime.TypeByExtension(strings.ToLower(filepath.Ext(filename))); value != "" {
		return strings.Split(value, ";")[0]
	}
	return strings.Split(http.DetectContentType(content), ";")[0]
}

func attachmentPublicURL(ctx context.Context, c *app.RequestContext, path string) string {
	domain := strings.TrimRight(strings.TrimSpace(media.StaticDomainContext(ctx)), "/")
	if domain != "" && !strings.Contains(domain, "localhost") && !strings.Contains(domain, "127.0.0.1") {
		return media.FullURL(path, domain)
	}
	host := strings.TrimSpace(string(c.Request.Host()))
	if host == "" {
		return media.FullURL(path, domain)
	}
	proto := strings.TrimSpace(strings.Split(string(c.Request.Header.Peek("X-Forwarded-Proto")), ",")[0])
	if proto == "" {
		proto = "http"
	}
	return strings.TrimRight(proto+"://"+host, "/") + path
}
