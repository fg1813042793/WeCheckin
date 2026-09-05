package infrastructure

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path"
	"path/filepath"
	"strings"
	"time"

	"wecheckin/backend/internal/modules/userfeedback/application"
	"wecheckin/backend/internal/support/media"
	"wecheckin/backend/internal/support/storage"
)

const feedbackImagePrefix = "uploads/feedback"

type ObjectStorage struct {
	saveReader       func(context.Context, io.Reader, string, storage.SaveOptions) (*storage.StoredFile, error)
	deleteStoredFile func(context.Context, *storage.StoredFile) error
	fullURL          func(context.Context, string) string
	now              func() time.Time
}

func NewObjectStorage() *ObjectStorage {
	return &ObjectStorage{
		saveReader:       storage.SaveReader,
		deleteStoredFile: storage.DeleteStoredFile,
		fullURL:          media.FullURLWithStaticDomainContext,
		now:              time.Now,
	}
}

func (objectStorage *ObjectStorage) Save(ctx context.Context, image application.ValidatedImage) (application.StoredImage, error) {
	now := objectStorage.currentTime()
	filename := fmt.Sprintf("%d%s", now.UnixNano(), strings.ToLower(filepath.Ext(image.OriginalName)))
	stored, err := objectStorage.save(ctx, bytes.NewReader(image.Content), image.OriginalName, storage.SaveOptions{
		Prefix:   feedbackImagePrefix,
		Filename: filename,
		Now:      now,
	})
	if err != nil || stored == nil || strings.TrimSpace(stored.ObjectKey) == "" {
		return application.StoredImage{}, application.ErrStorageFailed
	}

	provider := application.StorageProviderAliyun
	if stored.IsLocal {
		provider = application.StorageProviderLocal
	}
	return application.StoredImage{
		StorageProvider: provider,
		ObjectKey:       stored.ObjectKey,
		OriginalName:    image.OriginalName,
		ContentType:     image.ContentType,
		SizeBytes:       image.SizeBytes,
		URL:             objectStorage.PublicURL(ctx, stored.ObjectKey),
	}, nil
}

func (objectStorage *ObjectStorage) Delete(ctx context.Context, image application.StoredImage) error {
	objectKey, ok := feedbackObjectKey(image.ObjectKey)
	if !ok {
		return application.ErrStorageFailed
	}

	stored := &storage.StoredFile{ObjectKey: objectKey}
	switch strings.ToLower(strings.TrimSpace(image.StorageProvider)) {
	case application.StorageProviderLocal:
		stored.IsLocal = true
		localObjectPath := strings.TrimPrefix(objectKey, "uploads/")
		stored.LocalPath = filepath.Join(storage.LocalUploadRoot(), filepath.FromSlash(localObjectPath))
	case application.StorageProviderAliyun:
	default:
		return application.ErrStorageFailed
	}
	if err := objectStorage.delete(ctx, stored); err != nil {
		return application.ErrStorageFailed
	}
	return nil
}

func (objectStorage *ObjectStorage) PublicURL(ctx context.Context, objectKey string) string {
	objectKey = strings.TrimLeft(strings.TrimSpace(objectKey), "/")
	if objectKey == "" {
		return ""
	}
	return objectStorage.url(ctx, "/"+objectKey)
}

func (objectStorage *ObjectStorage) save(ctx context.Context, reader io.Reader, originalName string, options storage.SaveOptions) (*storage.StoredFile, error) {
	if objectStorage != nil && objectStorage.saveReader != nil {
		return objectStorage.saveReader(ctx, reader, originalName, options)
	}
	return storage.SaveReader(ctx, reader, originalName, options)
}

func (objectStorage *ObjectStorage) delete(ctx context.Context, stored *storage.StoredFile) error {
	if objectStorage != nil && objectStorage.deleteStoredFile != nil {
		return objectStorage.deleteStoredFile(ctx, stored)
	}
	return storage.DeleteStoredFile(ctx, stored)
}

func (objectStorage *ObjectStorage) url(ctx context.Context, objectPath string) string {
	if objectStorage != nil && objectStorage.fullURL != nil {
		return objectStorage.fullURL(ctx, objectPath)
	}
	return media.FullURLWithStaticDomainContext(ctx, objectPath)
}

func (objectStorage *ObjectStorage) currentTime() time.Time {
	if objectStorage != nil && objectStorage.now != nil {
		return objectStorage.now()
	}
	return time.Now()
}

func feedbackObjectKey(value string) (string, bool) {
	trimmed := strings.Trim(strings.TrimSpace(value), "/")
	cleaned := path.Clean(trimmed)
	if cleaned != trimmed || !strings.HasPrefix(cleaned, feedbackImagePrefix+"/") {
		return "", false
	}
	return cleaned, true
}

var _ application.ImageStorage = (*ObjectStorage)(nil)
