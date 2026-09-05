package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"wecheckin/backend/internal/config"
)

const defaultUploadRoot = "./uploads"

type SaveOptions struct {
	Prefix   string
	Filename string
	Now      time.Time
}

type StoredFile struct {
	URL         string
	RelativeURL string
	Path        string
	ObjectKey   string
	Filename    string
	Domain      string
	LocalPath   string
	IsLocal     bool
}

func DeleteStoredFile(ctx context.Context, stored *StoredFile) error {
	if stored == nil {
		return nil
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if stored.IsLocal {
		if strings.TrimSpace(stored.LocalPath) == "" {
			return errors.New("删除本地存储对象失败: 本地路径为空")
		}
		if err := os.Remove(stored.LocalPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			var pathErr *os.PathError
			if errors.As(err, &pathErr) {
				err = pathErr.Err
			}
			return fmt.Errorf("删除本地存储对象失败: %w", err)
		}
		return nil
	}
	if strings.TrimSpace(stored.ObjectKey) == "" {
		return errors.New("删除存储对象失败: 对象键为空")
	}
	return deleteAliyun(ctx, stored.ObjectKey)
}

func SaveMultipartFile(ctx context.Context, file *multipart.FileHeader, options SaveOptions) (*StoredFile, error) {
	if file == nil {
		return nil, fmt.Errorf("上传文件为空")
	}
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("打开上传文件: %w", err)
	}
	defer src.Close()
	return SaveReader(ctx, src, file.Filename, options)
}

func SaveReader(ctx context.Context, src io.Reader, originalFilename string, options SaveOptions) (*StoredFile, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if src == nil {
		return nil, fmt.Errorf("上传文件为空")
	}

	now := options.Now
	if now.IsZero() {
		now = time.Now()
	}
	filename := strings.TrimSpace(options.Filename)
	if filename == "" {
		filename = fmt.Sprintf("%d_%s", now.UnixNano(), filepath.Base(originalFilename))
	}
	filename = filepath.Base(filename)
	prefix := strings.Trim(strings.TrimSpace(options.Prefix), "/")
	if prefix == "" {
		prefix = "uploads"
	}
	dateDir := now.Format("2006/01/02")
	objectKey := path.Join(prefix, dateDir, filename)
	contentType := uploadContentType(filename)

	ossType := strings.ToLower(strings.TrimSpace(currentOSSConfig().Type))
	switch ossType {
	case "", "local":
		return saveLocal(&contextReader{ctx: ctx, reader: src}, objectKey, filename)
	case "aliyun":
		return saveAliyun(ctx, src, objectKey, filename, contentType)
	case "tencent":
		return nil, fmt.Errorf("暂不支持腾讯云 COS，请将 oss.type 设置为 local 或 aliyun")
	default:
		return nil, fmt.Errorf("不支持的 oss.type: %s", ossType)
	}
}

type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (reader *contextReader) Read(buffer []byte) (int, error) {
	if err := reader.ctx.Err(); err != nil {
		return 0, err
	}
	n, err := reader.reader.Read(buffer)
	if ctxErr := reader.ctx.Err(); ctxErr != nil {
		return n, ctxErr
	}
	return n, err
}

func LocalUploadRoot() string {
	root := strings.TrimSpace(currentOSSConfig().Local.Path)
	if root == "" {
		return defaultUploadRoot
	}
	return root
}

func RemoveLocal(file *StoredFile) {
	if file == nil || !file.IsLocal || strings.TrimSpace(file.LocalPath) == "" {
		return
	}
	_ = os.Remove(file.LocalPath)
}

func currentOSSConfig() config.OSSConfig {
	if config.Cfg == nil {
		return config.OSSConfig{
			Type:  "local",
			Local: config.LocalOSSConfig{Path: defaultUploadRoot},
		}
	}
	return config.Cfg.OSS
}

func saveLocal(src io.Reader, objectKey, filename string) (*StoredFile, error) {
	root := LocalUploadRoot()
	localObjectPath := objectKey
	if !strings.HasPrefix(objectKey, "uploads/") {
		localObjectPath = objectKey
	} else {
		localObjectPath = strings.TrimPrefix(objectKey, "uploads/")
	}
	localPath := filepath.Join(root, filepath.FromSlash(localObjectPath))
	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		return nil, fmt.Errorf("创建目录失败: %w", err)
	}
	out, err := os.Create(localPath)
	if err != nil {
		return nil, fmt.Errorf("保存文件失败: %w", err)
	}
	defer out.Close()
	if _, err := io.Copy(out, src); err != nil {
		_ = os.Remove(localPath)
		return nil, fmt.Errorf("保存文件失败: %w", err)
	}
	absPath, _ := filepath.Abs(localPath)
	relativeURL := "/" + objectKey
	return &StoredFile{
		URL:         relativeURL,
		RelativeURL: relativeURL,
		Path:        absPath,
		ObjectKey:   objectKey,
		Filename:    filename,
		LocalPath:   absPath,
		IsLocal:     true,
	}, nil
}

func uploadContentType(filename string) string {
	if value := mime.TypeByExtension(strings.ToLower(filepath.Ext(filename))); value != "" {
		return value
	}
	return "application/octet-stream"
}
