package application

import (
	"errors"
	"net/url"
	"path"
	"strings"

	"wecheckin/backend/internal/workflowcore"
)

const (
	maxWorkflowImageCount = 9
	maxWorkflowImageSize  = 20 * 1024 * 1024
)

var (
	ErrWorkflowImageInvalid = errors.New("流程图片数据无效")
	ErrWorkflowImageTooMany = errors.New("流程图片最多上传9张")
)

var workflowImageExtensions = map[string]struct{}{
	".jpg": {}, ".jpeg": {}, ".png": {}, ".gif": {}, ".webp": {},
}

func normalizeWorkflowImages(images []workflowcore.FormAttachment) ([]workflowcore.FormAttachment, error) {
	if len(images) == 0 {
		return nil, nil
	}
	if len(images) > maxWorkflowImageCount {
		return nil, ErrWorkflowImageTooMany
	}
	result := make([]workflowcore.FormAttachment, 0, len(images))
	seen := make(map[string]struct{}, len(images))
	for _, image := range images {
		image.ID = strings.TrimSpace(image.ID)
		image.Name = strings.TrimSpace(image.Name)
		image.URL = strings.TrimSpace(image.URL)
		image.MimeType = strings.ToLower(strings.TrimSpace(strings.Split(image.MimeType, ";")[0]))
		if !validWorkflowImage(image) {
			return nil, ErrWorkflowImageInvalid
		}
		if _, exists := seen[image.ID]; exists {
			return nil, ErrWorkflowImageInvalid
		}
		seen[image.ID] = struct{}{}
		result = append(result, image)
	}
	return result, nil
}

func validWorkflowImage(image workflowcore.FormAttachment) bool {
	if image.ID == "" || image.Name == "" || image.URL == "" || image.Size <= 0 || image.Size > maxWorkflowImageSize {
		return false
	}
	if !strings.HasPrefix(image.ID, "uploads/workflow/") || path.Clean(image.ID) != image.ID {
		return false
	}
	if _, ok := workflowImageExtensions[strings.ToLower(path.Ext(image.Name))]; !ok {
		return false
	}
	if !strings.HasPrefix(image.MimeType, "image/") {
		return false
	}
	parsed, err := url.Parse(image.URL)
	if err != nil || parsed.User != nil || (parsed.Scheme != "" && parsed.Scheme != "http" && parsed.Scheme != "https") {
		return false
	}
	return parsed.Path == "/"+image.ID
}
