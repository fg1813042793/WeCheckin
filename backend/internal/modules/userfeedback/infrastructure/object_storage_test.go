package infrastructure

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
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
		randomID: func() (string, error) {
			return strings.Repeat("a", 32), nil
		},
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
	if savedOptions.Filename != strings.Repeat("a", 32)+".png" {
		t.Fatalf("generated filename = %q", savedOptions.Filename)
	}
	if stored.StorageProvider != "aliyun" || stored.ObjectKey != "uploads/feedback/2026/09/05/generated.png" {
		t.Fatalf("stored provider/key = %q/%q", stored.StorageProvider, stored.ObjectKey)
	}
	if stored.OriginalName != image.OriginalName || stored.ContentType != image.ContentType || stored.SizeBytes != image.SizeBytes {
		t.Fatalf("stored metadata = %#v", stored)
	}
	if publicURLPath != "/uploads/feedback/2026/09/05/generated.png" || stored.URL != "https://cdn.example.test"+publicURLPath {
		t.Fatalf("stored URL = %q, path = %q", stored.URL, publicURLPath)
	}
}

func TestSecureRandomIDReturnsLongHexValues(t *testing.T) {
	first, err := secureRandomID()
	if err != nil {
		t.Fatalf("secureRandomID() first error = %v", err)
	}
	second, err := secureRandomID()
	if err != nil {
		t.Fatalf("secureRandomID() second error = %v", err)
	}
	if len(first) != 32 || len(second) != 32 {
		t.Fatalf("secureRandomID() lengths = %d and %d, want 32", len(first), len(second))
	}
	if _, err := hex.DecodeString(first); err != nil {
		t.Fatalf("secureRandomID() first value is not hex: %q", first)
	}
	if _, err := hex.DecodeString(second); err != nil {
		t.Fatalf("secureRandomID() second value is not hex: %q", second)
	}
	if first == second {
		t.Fatalf("secureRandomID() returned duplicate values: %q", first)
	}
}

func TestObjectStorageFixedTimeSequentialAndConcurrentSavesUseUniqueKeys(t *testing.T) {
	fixedTime := time.Date(2026, 9, 5, 14, 30, 0, 123, time.Local)
	var sequence atomic.Uint64
	objectStorage := &ObjectStorage{
		randomID: func() (string, error) {
			return fmt.Sprintf("%032x", sequence.Add(1)), nil
		},
		saveReader: func(_ context.Context, _ io.Reader, _ string, options storage.SaveOptions) (*storage.StoredFile, error) {
			return &storage.StoredFile{
				ObjectKey: path.Join(options.Prefix, options.Now.Format("2006/01/02"), options.Filename),
			}, nil
		},
		fullURL: func(_ context.Context, objectPath string) string { return objectPath },
		now:     func() time.Time { return fixedTime },
	}
	image := application.ValidatedImage{OriginalName: "image.png", Content: []byte("image")}

	const saveCount = 34
	keys := make(chan string, saveCount)
	errs := make(chan error, saveCount)
	save := func() {
		stored, err := objectStorage.Save(context.Background(), image)
		if err != nil {
			errs <- err
			return
		}
		keys <- stored.ObjectKey
	}
	save()
	save()
	var group sync.WaitGroup
	for range saveCount - 2 {
		group.Add(1)
		go func() {
			defer group.Done()
			save()
		}()
	}
	group.Wait()
	close(errs)
	close(keys)

	for err := range errs {
		t.Fatalf("ObjectStorage.Save() error = %v", err)
	}
	unique := make(map[string]struct{}, saveCount)
	for key := range keys {
		if !strings.HasPrefix(key, "uploads/feedback/2026/09/05/") || filepath.Ext(key) != ".png" {
			t.Fatalf("saved object key = %q", key)
		}
		if _, exists := unique[key]; exists {
			t.Fatalf("duplicate object key = %q", key)
		}
		unique[key] = struct{}{}
	}
	if len(unique) != saveCount {
		t.Fatalf("unique object keys = %d, want %d", len(unique), saveCount)
	}
}

func TestObjectStorageRandomFailureDoesNotSave(t *testing.T) {
	randomCause := errors.New("private random source failure")
	var saveCalls atomic.Int64
	objectStorage := &ObjectStorage{
		randomID: func() (string, error) { return "", randomCause },
		saveReader: func(context.Context, io.Reader, string, storage.SaveOptions) (*storage.StoredFile, error) {
			saveCalls.Add(1)
			return nil, nil
		},
		now: time.Now,
	}

	_, err := objectStorage.Save(context.Background(), application.ValidatedImage{OriginalName: "image.png", Content: []byte("image")})
	assertStableObjectStorageError(t, err, randomCause)
	if got := saveCalls.Load(); got != 0 {
		t.Fatalf("save calls = %d, want 0", got)
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

func TestObjectStoragePublicURLRejectsObjectKeyOutsideFeedbackPrefix(t *testing.T) {
	objectStorage := &ObjectStorage{
		fullURL: func(context.Context, string) string {
			t.Fatal("full URL generator must not be called for an invalid object key")
			return ""
		},
	}
	if got := objectStorage.PublicURL(context.Background(), "uploads/avatar/2026/09/05/image.png"); got != "" {
		t.Fatalf("PublicURL() = %q, want empty", got)
	}
}

func TestObjectStorageSaveReturnsStableStorageError(t *testing.T) {
	storageCause := errors.New("private storage failure")
	objectStorage := &ObjectStorage{
		saveReader: func(context.Context, io.Reader, string, storage.SaveOptions) (*storage.StoredFile, error) {
			return nil, storageCause
		},
		randomID: func() (string, error) { return strings.Repeat("a", 32), nil },
		now:      time.Now,
	}

	_, err := objectStorage.Save(context.Background(), application.ValidatedImage{OriginalName: "image.png", Content: []byte("image")})
	assertStableObjectStorageError(t, err, storageCause)
}

func TestObjectStorageSavePreservesContextErrors(t *testing.T) {
	for _, contextErr := range []error{context.Canceled, context.DeadlineExceeded} {
		t.Run(contextErr.Error(), func(t *testing.T) {
			objectStorage := &ObjectStorage{
				saveReader: func(context.Context, io.Reader, string, storage.SaveOptions) (*storage.StoredFile, error) {
					return nil, fmt.Errorf("wrapped context failure: %w", contextErr)
				},
				randomID: func() (string, error) { return strings.Repeat("a", 32), nil },
				now:      time.Now,
			}

			_, err := objectStorage.Save(context.Background(), application.ValidatedImage{OriginalName: "image.png", Content: []byte("image")})
			if !errors.Is(err, contextErr) {
				t.Fatalf("ObjectStorage.Save() error = %v, want errors.Is(_, %v)", err, contextErr)
			}
			if err.Error() != contextErr.Error() {
				t.Fatalf("ObjectStorage.Save() error text = %q, want %q", err.Error(), contextErr.Error())
			}
		})
	}
}

func TestObjectStorageSaveRejectsObjectKeyOutsideFeedbackPrefix(t *testing.T) {
	objectStorage := &ObjectStorage{
		saveReader: func(context.Context, io.Reader, string, storage.SaveOptions) (*storage.StoredFile, error) {
			return &storage.StoredFile{ObjectKey: "uploads/avatar/2026/09/05/image.png"}, nil
		},
		randomID: func() (string, error) { return strings.Repeat("a", 32), nil },
		fullURL: func(context.Context, string) string {
			t.Fatal("full URL generator must not be called for an invalid object key")
			return ""
		},
		now: time.Now,
	}

	_, err := objectStorage.Save(context.Background(), application.ValidatedImage{OriginalName: "image.png", Content: []byte("image")})
	if !errors.Is(err, application.ErrStorageFailed) {
		t.Fatalf("ObjectStorage.Save() error = %v, want errors.Is(_, ErrStorageFailed)", err)
	}
	if err.Error() != application.ErrStorageFailed.Error() {
		t.Fatalf("ObjectStorage.Save() error text = %q", err.Error())
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
			provider:      "local",
			wantLocal:     true,
			wantLocalPath: filepath.Join(uploadRoot, "feedback", "2026", "09", "05", "image.png"),
		},
		{name: "aliyun", provider: "aliyun"},
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
	deleteCause := errors.New("private delete failure")
	objectStorage := &ObjectStorage{
		deleteStoredFile: func(context.Context, *storage.StoredFile) error {
			return deleteCause
		},
	}
	err := objectStorage.Delete(context.Background(), application.StoredImage{
		StorageProvider: "aliyun",
		ObjectKey:       "uploads/feedback/image.png",
	})
	assertStableObjectStorageError(t, err, deleteCause)
}

func TestObjectStorageDeletePreservesContextErrors(t *testing.T) {
	for _, contextErr := range []error{context.Canceled, context.DeadlineExceeded} {
		t.Run(contextErr.Error(), func(t *testing.T) {
			objectStorage := &ObjectStorage{
				deleteStoredFile: func(context.Context, *storage.StoredFile) error {
					return fmt.Errorf("wrapped context failure: %w", contextErr)
				},
			}
			err := objectStorage.Delete(context.Background(), application.StoredImage{
				StorageProvider: "aliyun",
				ObjectKey:       "uploads/feedback/image.png",
			})
			if !errors.Is(err, contextErr) {
				t.Fatalf("ObjectStorage.Delete() error = %v, want errors.Is(_, %v)", err, contextErr)
			}
			if err.Error() != contextErr.Error() {
				t.Fatalf("ObjectStorage.Delete() error text = %q, want %q", err.Error(), contextErr.Error())
			}
		})
	}
}

func assertStableObjectStorageError(t *testing.T, err, cause error) {
	t.Helper()
	if !errors.Is(err, application.ErrStorageFailed) {
		t.Fatalf("error = %v, want errors.Is(_, ErrStorageFailed)", err)
	}
	if !errors.Is(err, cause) {
		t.Fatalf("error = %v, want errors.Is(_, cause)", err)
	}
	if err.Error() != application.ErrStorageFailed.Error() {
		t.Fatalf("error text = %q, want %q", err.Error(), application.ErrStorageFailed.Error())
	}
	if strings.Contains(err.Error(), cause.Error()) {
		t.Fatalf("error leaked cause: %v", err)
	}
}

type staticDomainContextKey struct{}
