package application

import (
	"context"
	"net/http"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

const (
	MaxImageSizeBytes       = 10 * 1024 * 1024
	MaxImagesPerMessage     = 6
	MaxImagesPerFeedback    = 30
	MaxFeedbackContentRunes = 5000

	StorageProviderLocal  = "local"
	StorageProviderAliyun = "aliyun"
)

type ValidatedImage struct {
	OriginalName string
	ContentType  string
	SizeBytes    uint64
	Content      []byte
}

type ValidatedMessage struct {
	Content string
	Images  []ValidatedImage
}

type StoredImage struct {
	StorageProvider string
	ObjectKey       string
	OriginalName    string
	ContentType     string
	SizeBytes       uint64
	URL             string
}

type ImageStorage interface {
	Save(ctx context.Context, image ValidatedImage) (StoredImage, error)
	Delete(ctx context.Context, image StoredImage) error
	PublicURL(ctx context.Context, objectKey string) string
}

var imageMIMEByExtension = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".webp": "image/webp",
}

func ValidateImage(input AttachmentInput) (ValidatedImage, error) {
	if uint64(len(input.Content)) > MaxImageSizeBytes {
		return ValidatedImage{}, ErrAttachmentTooLarge
	}

	extension := strings.ToLower(filepath.Ext(strings.TrimSpace(input.OriginalName)))
	expectedMIME, allowed := imageMIMEByExtension[extension]
	declaredMIME := strings.ToLower(strings.TrimSpace(input.ContentType))
	if !allowed || declaredMIME != expectedMIME || http.DetectContentType(input.Content) != expectedMIME {
		return ValidatedImage{}, ErrAttachmentTypeNotAllowed
	}

	return ValidatedImage{
		OriginalName: input.OriginalName,
		ContentType:  expectedMIME,
		SizeBytes:    uint64(len(input.Content)),
		Content:      input.Content,
	}, nil
}

func ValidateInitialMessage(content string, attachments []AttachmentInput) (ValidatedMessage, error) {
	return validateImageMessage(content, attachments, 0, true)
}

func ValidateSupplementMessage(content string, attachments []AttachmentInput, existingImageCount int64) (ValidatedMessage, error) {
	return validateImageMessage(content, attachments, existingImageCount, false)
}

func validateImageMessage(content string, attachments []AttachmentInput, existingImageCount int64, contentRequired bool) (ValidatedMessage, error) {
	trimmedContent := strings.TrimSpace(content)
	if utf8.RuneCountInString(trimmedContent) > MaxFeedbackContentRunes {
		return ValidatedMessage{}, ErrInvalidArgument
	}
	if contentRequired && trimmedContent == "" {
		return ValidatedMessage{}, ErrInvalidArgument
	}
	if !contentRequired && trimmedContent == "" && len(attachments) == 0 {
		return ValidatedMessage{}, ErrInvalidArgument
	}
	if existingImageCount < 0 {
		return ValidatedMessage{}, ErrInvalidArgument
	}
	if len(attachments) > MaxImagesPerMessage || existingImageCount+int64(len(attachments)) > MaxImagesPerFeedback {
		return ValidatedMessage{}, ErrAttachmentLimitExceeded
	}

	images := make([]ValidatedImage, 0, len(attachments))
	for _, attachment := range attachments {
		image, err := ValidateImage(attachment)
		if err != nil {
			return ValidatedMessage{}, err
		}
		images = append(images, image)
	}

	return ValidatedMessage{Content: trimmedContent, Images: images}, nil
}
