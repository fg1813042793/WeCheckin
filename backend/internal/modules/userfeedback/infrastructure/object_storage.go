package infrastructure

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"path"
	"path/filepath"
	"strings"
	"time"

	"wecheckin/backend/internal/modules/userfeedback/application"
	"wecheckin/backend/internal/support/media"
	"wecheckin/backend/internal/support/storage"
)

const (
	feedbackImagePrefix   = "uploads/feedback"
	storageProviderLocal  = "local"
	storageProviderAliyun = "aliyun"
)

type ObjectStorage struct {
	saveReader       func(context.Context, io.Reader, string, storage.SaveOptions) (*storage.StoredFile, error)
	deleteStoredFile func(context.Context, *storage.StoredFile) error
	randomID         func() (string, error)
	fullURL          func(context.Context, string) string
	now              func() time.Time
}

func NewObjectStorage() *ObjectStorage {
	return &ObjectStorage{
		saveReader:       storage.SaveReader,
		deleteStoredFile: storage.DeleteStoredFile,
		randomID:         secureRandomID,
		fullURL:          media.FullURLWithStaticDomainContext,
		now:              time.Now,
	}
}

func (objectStorage *ObjectStorage) Save(ctx context.Context, image application.ValidatedImage) (application.StoredImage, error) {
	if err := ctx.Err(); err != nil {
		return application.StoredImage{}, err
	}
	randomID, err := objectStorage.nextRandomID()
	if err != nil {
		return application.StoredImage{}, newObjectStorageError(ctx, err)
	}
	now := objectStorage.currentTime()
	filename := randomID + strings.ToLower(filepath.Ext(image.OriginalName))
	stored, err := objectStorage.save(ctx, bytes.NewReader(image.Content), image.OriginalName, storage.SaveOptions{
		Prefix:   feedbackImagePrefix,
		Filename: filename,
		Now:      now,
	})
	if err != nil {
		return application.StoredImage{}, newObjectStorageError(ctx, err)
	}
	if stored == nil {
		return application.StoredImage{}, newObjectStorageError(ctx, errEmptyStorageResult)
	}
	objectKey, ok := feedbackObjectKey(stored.ObjectKey)
	if !ok {
		return application.StoredImage{}, newObjectStorageError(ctx, errInvalidFeedbackObjectKey)
	}

	provider := storageProviderAliyun
	if stored.IsLocal {
		provider = storageProviderLocal
	}
	return application.StoredImage{
		StorageProvider: provider,
		ObjectKey:       objectKey,
		OriginalName:    image.OriginalName,
		ContentType:     image.ContentType,
		SizeBytes:       image.SizeBytes,
		URL:             objectStorage.PublicURL(ctx, objectKey),
	}, nil
}

func (objectStorage *ObjectStorage) Delete(ctx context.Context, image application.StoredImage) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	objectKey, ok := feedbackObjectKey(image.ObjectKey)
	if !ok {
		return newObjectStorageError(ctx, errInvalidFeedbackObjectKey)
	}

	stored := &storage.StoredFile{ObjectKey: objectKey}
	switch strings.ToLower(strings.TrimSpace(image.StorageProvider)) {
	case storageProviderLocal:
		stored.IsLocal = true
		localObjectPath := strings.TrimPrefix(objectKey, "uploads/")
		stored.LocalPath = filepath.Join(storage.LocalUploadRoot(), filepath.FromSlash(localObjectPath))
	case storageProviderAliyun:
	default:
		return newObjectStorageError(ctx, errUnsupportedStorageProvider)
	}
	if err := objectStorage.delete(ctx, stored); err != nil {
		return newObjectStorageError(ctx, err)
	}
	return nil
}

func (objectStorage *ObjectStorage) PublicURL(ctx context.Context, objectKey string) string {
	objectKey, ok := feedbackObjectKey(objectKey)
	if !ok {
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

func (objectStorage *ObjectStorage) nextRandomID() (string, error) {
	if objectStorage != nil && objectStorage.randomID != nil {
		return objectStorage.randomID()
	}
	return secureRandomID()
}

func secureRandomID() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(value[:]), nil
}

func feedbackObjectKey(value string) (string, bool) {
	trimmed := strings.Trim(strings.TrimSpace(value), "/")
	cleaned := path.Clean(trimmed)
	if cleaned != trimmed || !strings.HasPrefix(cleaned, feedbackImagePrefix+"/") {
		return "", false
	}
	return cleaned, true
}

var (
	errEmptyStorageResult         = errors.New("empty feedback storage result")
	errInvalidFeedbackObjectKey   = errors.New("invalid feedback object key")
	errUnsupportedStorageProvider = errors.New("unsupported feedback storage provider")
)

type objectStorageError struct {
	cause error
}

func (err *objectStorageError) Error() string {
	return application.ErrStorageFailed.Error()
}

func (err *objectStorageError) SafeCleanupLogMessage() string {
	return application.ErrStorageFailed.Error()
}

func (err *objectStorageError) Unwrap() []error {
	return []error{application.ErrStorageFailed, err.cause}
}

func newObjectStorageError(ctx context.Context, cause error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return &objectStorageError{cause: cause}
}

var _ application.ImageStorage = (*ObjectStorage)(nil)
