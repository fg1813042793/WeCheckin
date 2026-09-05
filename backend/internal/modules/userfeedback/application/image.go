package application

import (
	"bytes"
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
	if !utf8.ValidString(input.OriginalName) || !utf8.ValidString(input.ContentType) {
		return ValidatedImage{}, ErrInvalidArgument
	}
	originalName := strings.TrimSpace(input.OriginalName)
	if originalName == "" || utf8.RuneCountInString(originalName) > 255 {
		return ValidatedImage{}, ErrInvalidArgument
	}

	extension := strings.ToLower(filepath.Ext(originalName))
	expectedMIME, allowed := imageMIMEByExtension[extension]
	declaredMIME := strings.ToLower(strings.TrimSpace(input.ContentType))
	if !allowed || declaredMIME != expectedMIME || http.DetectContentType(input.Content) != expectedMIME {
		return ValidatedImage{}, ErrAttachmentTypeNotAllowed
	}

	return ValidatedImage{
		OriginalName: originalName,
		ContentType:  expectedMIME,
		SizeBytes:    uint64(len(input.Content)),
		Content:      bytes.Clone(input.Content),
	}, nil
}

func ValidateInitialMessage(content string, attachments []AttachmentInput) (ValidatedMessage, error) {
	return validateImageMessage(content, attachments, 0, true)
}

func ValidateSupplementMessage(content string, attachments []AttachmentInput, existingImageCount int64) (ValidatedMessage, error) {
	return validateImageMessage(content, attachments, existingImageCount, false)
}

func validateImageMessage(content string, attachments []AttachmentInput, existingImageCount int64, contentRequired bool) (ValidatedMessage, error) {
	if !utf8.ValidString(content) {
		return ValidatedMessage{}, ErrInvalidArgument
	}
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
	if len(attachments) > MaxImagesPerMessage ||
		existingImageCount > MaxImagesPerFeedback ||
		int64(len(attachments)) > MaxImagesPerFeedback-existingImageCount {
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
