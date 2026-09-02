package workflowhandler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"

	workflowservice "wecheckin/backend/internal/service/admin/workflow"
	"wecheckin/backend/internal/support/storage"
)

const workflowLogoMaxSize int64 = 2 * 1024 * 1024

var workflowLogoContentTypes = map[string]string{
	".jpeg": "image/jpeg",
	".jpg":  "image/jpeg",
	".png":  "image/png",
	".webp": "image/webp",
}

func isWorkflowMultipart(c *app.RequestContext) bool {
	contentType := strings.ToLower(string(c.Request.Header.ContentType()))
	return strings.HasPrefix(contentType, "multipart/form-data")
}

func workflowCreateRequestFromMultipart(c *app.RequestContext) workflowservice.CreateRequest {
	return workflowservice.CreateRequest{
		Key:         string(c.FormValue("key")),
		Name:        string(c.FormValue("name")),
		Description: string(c.FormValue("description")),
		Category:    string(c.FormValue("category")),
		Draft:       workflowDraftFromMultipart(c),
	}
}

func workflowUpdateRequestFromMultipart(c *app.RequestContext) workflowservice.UpdateRequest {
	return workflowservice.UpdateRequest{
		Name:        string(c.FormValue("name")),
		Description: string(c.FormValue("description")),
		Category:    string(c.FormValue("category")),
		Draft:       workflowDraftFromMultipart(c),
	}
}

func workflowDraftFromMultipart(c *app.RequestContext) json.RawMessage {
	value := strings.TrimSpace(string(c.FormValue("draft")))
	if value == "" {
		return nil
	}
	return json.RawMessage(value)
}

func optionalWorkflowLogo(c *app.RequestContext) (*multipart.FileHeader, error) {
	form, err := c.MultipartForm()
	if err != nil {
		return nil, errors.New("读取流程 Logo 失败")
	}
	files := form.File["logo"]
	if len(files) == 0 {
		return nil, nil
	}
	if len(files) > 1 {
		return nil, errors.New("每个流程只能上传一个 Logo")
	}
	return files[0], nil
}

func validateWorkflowLogo(file *multipart.FileHeader) error {
	if file == nil {
		return errors.New("请选择流程 Logo")
	}
	if file.Size > workflowLogoMaxSize {
		return errors.New("流程 Logo 不能超过 2MB")
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	wantContentType, allowed := workflowLogoContentTypes[ext]
	if !allowed {
		return errors.New("流程 Logo 仅支持 PNG、JPG、JPEG 或 WebP 格式")
	}
	source, err := file.Open()
	if err != nil {
		return errors.New("读取流程 Logo 失败")
	}
	defer source.Close()
	content, err := io.ReadAll(io.LimitReader(source, workflowLogoMaxSize+1))
	if err != nil {
		return errors.New("读取流程 Logo 失败")
	}
	if int64(len(content)) > workflowLogoMaxSize {
		return errors.New("流程 Logo 不能超过 2MB")
	}
	if len(content) == 0 || http.DetectContentType(content) != wantContentType {
		return errors.New("流程 Logo 文件内容与格式不匹配")
	}
	return nil
}

func saveWorkflowLogo(ctx context.Context, file *multipart.FileHeader) (*storage.StoredFile, error) {
	if err := validateWorkflowLogo(file); err != nil {
		return nil, err
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	stored, err := storage.SaveMultipartFile(ctx, file, storage.SaveOptions{
		Prefix:   "uploads/workflow-logos",
		Filename: fmt.Sprintf("%d%s", time.Now().UnixNano(), ext),
	})
	if err != nil {
		return nil, fmt.Errorf("保存流程 Logo 失败: %w", err)
	}
	return stored, nil
}
