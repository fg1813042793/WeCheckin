package infrastructure

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"wecheckin/backend/internal/config"
	"wecheckin/backend/internal/modules/userfeedback/application"
	"wecheckin/backend/internal/support/media"
	"wecheckin/backend/internal/support/storage"
)

func TestObjectStorageImplementsImageStorage(t *testing.T) {
	var _ application.ImageStorage = NewObjectStorage()
}

func TestObjectStorageSaveUsesFeedbackPrefixAndMapsMetadata(t *testing.T) {
	now := time.Date(2026, 9, 5, 14, 30, 0, 123, time.Local)
	ctx := context.WithValue(context.Background(), staticDomainContextKey{}, "https://cdn.example.test")
	var savedContent string
	var savedOriginalName string
	var savedOptions storage.SaveOptions
	var publicURLPath string
	objectStorage := &ObjectStorage{
		saveReader: func(_ context.Context, reader io.Reader, originalName string, options storage.SaveOptions) (*storage.StoredFile, error) {
			content, err := io.ReadAll(reader)
			if err != nil {
				t.Fatalf("read save input: %v", err)
			}
			savedContent = string(content)
			savedOriginalName = originalName
			savedOptions = options
			return &storage.StoredFile{
				ObjectKey: "uploads/feedback/2026/09/05/generated.png",
				Filename:  "generated.png",
			}, nil
		},
		deleteStoredFile: storage.DeleteStoredFile,
		fullURL: func(ctx context.Context, path string) string {
			publicURLPath = path
			return ctx.Value(staticDomainContextKey{}).(string) + path
		},
		now: func() time.Time { return now },
	}
	image := application.ValidatedImage{
		OriginalName: "用户截图.png",
		ContentType:  "image/png",
		SizeBytes:    12,
		Content:      []byte("image-content"),
	}

	stored, err := objectStorage.Save(ctx, image)
	if err != nil {
		t.Fatalf("ObjectStorage.Save() error = %v", err)
	}
	if savedContent != "image-content" || savedOriginalName != image.OriginalName {
		t.Fatalf("save input content=%q originalName=%q", savedContent, savedOriginalName)
	}
	if savedOptions.Prefix != feedbackImagePrefix {
		t.Fatalf("save prefix = %q, want %q", savedOptions.Prefix, feedbackImagePrefix)
	}
	if savedOptions.Now != now {
		t.Fatalf("save time = %v, want %v", savedOptions.Now, now)
	}
	if filepath.Ext(savedOptions.Filename) != ".png" || savedOptions.Filename == image.OriginalName {
		t.Fatalf("generated filename = %q", savedOptions.Filename)
	}
	if stored.StorageProvider != application.StorageProviderAliyun || stored.ObjectKey != "uploads/feedback/2026/09/05/generated.png" {
		t.Fatalf("stored provider/key = %q/%q", stored.StorageProvider, stored.ObjectKey)
	}
	if stored.OriginalName != image.OriginalName || stored.ContentType != image.ContentType || stored.SizeBytes != image.SizeBytes {
		t.Fatalf("stored metadata = %#v", stored)
	}
	if publicURLPath != "/uploads/feedback/2026/09/05/generated.png" || stored.URL != "https://cdn.example.test"+publicURLPath {
		t.Fatalf("stored URL = %q, path = %q", stored.URL, publicURLPath)
	}
}

func TestObjectStoragePublicURLIsGeneratedForEachContext(t *testing.T) {
	objectStorage := &ObjectStorage{
		fullURL: func(ctx context.Context, path string) string {
			return ctx.Value(staticDomainContextKey{}).(string) + path
		},
	}
	objectKey := "uploads/feedback/2026/09/05/image.png"
	first := objectStorage.PublicURL(context.WithValue(context.Background(), staticDomainContextKey{}, "https://one.example"), objectKey)
	second := objectStorage.PublicURL(context.WithValue(context.Background(), staticDomainContextKey{}, "https://two.example"), objectKey)
	if first != "https://one.example/"+objectKey || second != "https://two.example/"+objectKey {
		t.Fatalf("dynamic URLs = %q and %q", first, second)
	}
}

func TestObjectStoragePublicURLUsesMediaStaticDomain(t *testing.T) {
	ctx := context.Background()
	objectKey := "uploads/feedback/2026/09/05/image.png"
	want := media.FullURLWithStaticDomainContext(ctx, "/"+objectKey)
	if got := NewObjectStorage().PublicURL(ctx, objectKey); got != want {
		t.Fatalf("PublicURL() = %q, want %q", got, want)
	}
}

func TestObjectStorageSaveReturnsStableStorageError(t *testing.T) {
	objectStorage := &ObjectStorage{
		saveReader: func(context.Context, io.Reader, string, storage.SaveOptions) (*storage.StoredFile, error) {
			return nil, errors.New("private storage failure")
		},
		now: time.Now,
	}

	_, err := objectStorage.Save(context.Background(), application.ValidatedImage{OriginalName: "image.png", Content: []byte("image")})
	if err != application.ErrStorageFailed {
		t.Fatalf("ObjectStorage.Save() error = %v, want stable ErrStorageFailed", err)
	}
	if strings.Contains(err.Error(), "private storage failure") {
		t.Fatalf("ObjectStorage.Save() leaked storage error: %v", err)
	}
}

func TestObjectStorageDeleteDelegatesLocalAndAliyunObjects(t *testing.T) {
	oldCfg := config.Cfg
	uploadRoot := t.TempDir()
	config.Cfg = &config.Config{OSS: config.OSSConfig{Local: config.LocalOSSConfig{Path: uploadRoot}}}
	t.Cleanup(func() { config.Cfg = oldCfg })

	tests := []struct {
		name          string
		provider      string
		wantLocal     bool
		wantLocalPath string
	}{
		{
			name:          "local",
			provider:      application.StorageProviderLocal,
			wantLocal:     true,
			wantLocalPath: filepath.Join(uploadRoot, "feedback", "2026", "09", "05", "image.png"),
		},
		{name: "aliyun", provider: application.StorageProviderAliyun},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var delegated *storage.StoredFile
			objectStorage := &ObjectStorage{
				deleteStoredFile: func(_ context.Context, stored *storage.StoredFile) error {
					copy := *stored
					delegated = &copy
					return nil
				},
			}
			image := application.StoredImage{
				StorageProvider: test.provider,
				ObjectKey:       "uploads/feedback/2026/09/05/image.png",
			}

			if err := objectStorage.Delete(context.Background(), image); err != nil {
				t.Fatalf("ObjectStorage.Delete() error = %v", err)
			}
			if delegated == nil || delegated.ObjectKey != image.ObjectKey || delegated.IsLocal != test.wantLocal || delegated.LocalPath != test.wantLocalPath {
				t.Fatalf("delegated stored file = %#v", delegated)
			}
		})
	}
}

func TestObjectStorageDeleteCompensatesLocalSave(t *testing.T) {
	oldCfg := config.Cfg
	uploadRoot := t.TempDir()
	config.Cfg = &config.Config{OSS: config.OSSConfig{
		Type:  "local",
		Local: config.LocalOSSConfig{Path: uploadRoot},
	}}
	t.Cleanup(func() { config.Cfg = oldCfg })
	objectStorage := NewObjectStorage()

	stored, err := objectStorage.Save(context.Background(), application.ValidatedImage{
		OriginalName: "feedback.png",
		ContentType:  "image/png",
		SizeBytes:    8,
		Content:      []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a},
	})
	if err != nil {
		t.Fatalf("ObjectStorage.Save(local) error = %v", err)
	}
	localPath := filepath.Join(uploadRoot, filepath.FromSlash(strings.TrimPrefix(stored.ObjectKey, "uploads/")))
	if _, err := os.Stat(localPath); err != nil {
		t.Fatalf("saved local image is missing: %v", err)
	}
	if err := objectStorage.Delete(context.Background(), stored); err != nil {
		t.Fatalf("ObjectStorage.Delete(local) error = %v", err)
	}
	if _, err := os.Stat(localPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("local image remains after compensation: %v", err)
	}
}

func TestObjectStorageDeleteReturnsStableStorageError(t *testing.T) {
	objectStorage := &ObjectStorage{
		deleteStoredFile: func(context.Context, *storage.StoredFile) error {
			return errors.New("private delete failure")
		},
	}
	err := objectStorage.Delete(context.Background(), application.StoredImage{
		StorageProvider: application.StorageProviderAliyun,
		ObjectKey:       "uploads/feedback/image.png",
	})
	if err != application.ErrStorageFailed {
		t.Fatalf("ObjectStorage.Delete() error = %v, want stable ErrStorageFailed", err)
	}
	if strings.Contains(err.Error(), "private delete failure") {
		t.Fatalf("ObjectStorage.Delete() leaked storage error: %v", err)
	}
}

type staticDomainContextKey struct{}
