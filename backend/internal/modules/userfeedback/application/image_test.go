package application

import (
	"errors"
	"math"
	"os"
	"strings"
	"testing"
)

func TestImageApplicationDoesNotOwnConcreteStorageProviders(t *testing.T) {
	source, err := os.ReadFile("image.go")
	if err != nil {
		t.Fatalf("read image.go: %v", err)
	}
	for _, name := range []string{"StorageProviderLocal", "StorageProviderAliyun"} {
		if strings.Contains(string(source), name) {
			t.Errorf("application image boundary declares concrete provider %s", name)
		}
	}
}

func TestValidateImageAcceptsMatchingExtensionDeclaredMIMEAndContent(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		contentType string
	}{
		{name: "jpg", filename: "photo.jpg", contentType: "image/jpeg"},
		{name: "jpeg", filename: "photo.jpeg", contentType: "image/jpeg"},
		{name: "png", filename: "photo.png", contentType: "image/png"},
		{name: "webp", filename: "photo.webp", contentType: "image/webp"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := validImageInput(test.filename, test.contentType)
			image, err := ValidateImage(input)
			if err != nil {
				t.Fatalf("ValidateImage() error = %v", err)
			}
			if image.OriginalName != input.OriginalName {
				t.Errorf("OriginalName = %q, want %q", image.OriginalName, input.OriginalName)
			}
			if image.ContentType != test.contentType {
				t.Errorf("ContentType = %q, want %q", image.ContentType, test.contentType)
			}
			if image.SizeBytes != uint64(len(input.Content)) {
				t.Errorf("SizeBytes = %d, want %d", image.SizeBytes, len(input.Content))
			}
			if string(image.Content) != string(input.Content) {
				t.Fatal("validated content differs from input")
			}
		})
	}
}

func TestValidateImageRejectsUnsupportedOrMismatchedTypes(t *testing.T) {
	tests := []struct {
		name  string
		input AttachmentInput
	}{
		{
			name:  "unsupported extension",
			input: AttachmentInput{OriginalName: "photo.gif", ContentType: "image/gif", Content: []byte("GIF89a")},
		},
		{
			name:  "unsupported declared mime",
			input: AttachmentInput{OriginalName: "photo.jpg", ContentType: "application/octet-stream", Content: minimalJPEG()},
		},
		{
			name:  "extension and declared mime differ",
			input: AttachmentInput{OriginalName: "photo.png", ContentType: "image/jpeg", Content: minimalPNG()},
		},
		{
			name:  "declared mime and content differ",
			input: AttachmentInput{OriginalName: "photo.jpg", ContentType: "image/jpeg", Content: minimalPNG()},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ValidateImage(test.input)
			if !errors.Is(err, ErrAttachmentTypeNotAllowed) {
				t.Fatalf("ValidateImage() error = %v, want ErrAttachmentTypeNotAllowed", err)
			}
			if err != ErrAttachmentTypeNotAllowed {
				t.Fatalf("ValidateImage() exposed unstable error = %v", err)
			}
		})
	}
}

func TestValidateImageAllowsExactlyTenMiB(t *testing.T) {
	content := make([]byte, MaxImageSizeBytes)
	copy(content, minimalPNG())
	image, err := ValidateImage(AttachmentInput{
		OriginalName: "limit.png",
		ContentType:  "image/png",
		SizeBytes:    uint64(len(content)),
		Content:      content,
	})
	if err != nil {
		t.Fatalf("ValidateImage(exact limit) error = %v", err)
	}
	if image.SizeBytes != MaxImageSizeBytes {
		t.Fatalf("SizeBytes = %d, want %d", image.SizeBytes, MaxImageSizeBytes)
	}
}

func TestValidateImageRejectsMoreThanTenMiB(t *testing.T) {
	content := make([]byte, MaxImageSizeBytes+1)
	copy(content, minimalPNG())
	_, err := ValidateImage(AttachmentInput{
		OriginalName: "too-large.png",
		ContentType:  "image/png",
		SizeBytes:    uint64(len(content)),
		Content:      content,
	})
	if !errors.Is(err, ErrAttachmentTooLarge) {
		t.Fatalf("ValidateImage(over limit) error = %v, want ErrAttachmentTooLarge", err)
	}
}

func TestValidateInitialMessageTrimsAndRequiresContent(t *testing.T) {
	message, err := ValidateInitialMessage("  中文反馈  \n", nil)
	if err != nil {
		t.Fatalf("ValidateInitialMessage() error = %v", err)
	}
	if message.Content != "中文反馈" {
		t.Fatalf("Content = %q, want trimmed Chinese content", message.Content)
	}

	if _, err := ValidateInitialMessage(" \t\n ", []AttachmentInput{validImageInput("photo.png", "image/png")}); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("ValidateInitialMessage(empty content) error = %v, want ErrInvalidArgument", err)
	}
}

func TestValidateInitialMessageCountsContentByRunes(t *testing.T) {
	if _, err := ValidateInitialMessage(strings.Repeat("中", MaxFeedbackContentRunes), nil); err != nil {
		t.Fatalf("ValidateInitialMessage(exact rune limit) error = %v", err)
	}
	if _, err := ValidateInitialMessage(strings.Repeat("中", MaxFeedbackContentRunes+1), nil); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("ValidateInitialMessage(over rune limit) error = %v, want ErrInvalidArgument", err)
	}
}

func TestValidateInitialMessageAllowsAtMostSixImages(t *testing.T) {
	if _, err := ValidateInitialMessage("content", validImageInputs(MaxImagesPerMessage)); err != nil {
		t.Fatalf("ValidateInitialMessage(exact image limit) error = %v", err)
	}
	if _, err := ValidateInitialMessage("content", validImageInputs(MaxImagesPerMessage+1)); !errors.Is(err, ErrAttachmentLimitExceeded) {
		t.Fatalf("ValidateInitialMessage(over image limit) error = %v, want ErrAttachmentLimitExceeded", err)
	}
}

func TestValidateSupplementMessageAllowsTextOnlyOrImagesOnly(t *testing.T) {
	textOnly, err := ValidateSupplementMessage("  补充文字  ", nil, 0)
	if err != nil {
		t.Fatalf("ValidateSupplementMessage(text only) error = %v", err)
	}
	if textOnly.Content != "补充文字" || len(textOnly.Images) != 0 {
		t.Fatalf("text-only message = %#v", textOnly)
	}

	imagesOnly, err := ValidateSupplementMessage(" \n ", validImageInputs(1), 0)
	if err != nil {
		t.Fatalf("ValidateSupplementMessage(images only) error = %v", err)
	}
	if imagesOnly.Content != "" || len(imagesOnly.Images) != 1 {
		t.Fatalf("images-only message = %#v", imagesOnly)
	}
}

func TestValidateSupplementMessageRejectsEmptyMessageAndLongContent(t *testing.T) {
	if _, err := ValidateSupplementMessage(" \n ", nil, 0); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("ValidateSupplementMessage(empty) error = %v, want ErrInvalidArgument", err)
	}
	if _, err := ValidateSupplementMessage(strings.Repeat("中", MaxFeedbackContentRunes+1), nil, 0); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("ValidateSupplementMessage(long content) error = %v, want ErrInvalidArgument", err)
	}
}

func TestValidateSupplementMessageEnforcesPerMessageAndCumulativeLimits(t *testing.T) {
	if _, err := ValidateSupplementMessage("", validImageInputs(MaxImagesPerMessage), MaxImagesPerFeedback-MaxImagesPerMessage); err != nil {
		t.Fatalf("ValidateSupplementMessage(exact cumulative limit) error = %v", err)
	}
	if _, err := ValidateSupplementMessage("", validImageInputs(MaxImagesPerMessage+1), 0); !errors.Is(err, ErrAttachmentLimitExceeded) {
		t.Fatalf("ValidateSupplementMessage(over per-message limit) error = %v, want ErrAttachmentLimitExceeded", err)
	}
	if _, err := ValidateSupplementMessage("", validImageInputs(MaxImagesPerMessage), MaxImagesPerFeedback-MaxImagesPerMessage+1); !errors.Is(err, ErrAttachmentLimitExceeded) {
		t.Fatalf("ValidateSupplementMessage(over cumulative limit) error = %v, want ErrAttachmentLimitExceeded", err)
	}
}

func TestValidateSupplementMessageRejectsOverflowingExistingImageCount(t *testing.T) {
	_, err := ValidateSupplementMessage("", validImageInputs(1), math.MaxInt64)
	if !errors.Is(err, ErrAttachmentLimitExceeded) {
		t.Fatalf("ValidateSupplementMessage(MaxInt64 count) error = %v, want ErrAttachmentLimitExceeded", err)
	}
}

func validImageInputs(count int) []AttachmentInput {
	inputs := make([]AttachmentInput, count)
	for index := range inputs {
		inputs[index] = validImageInput("photo.png", "image/png")
	}
	return inputs
}

func validImageInput(filename, contentType string) AttachmentInput {
	var content []byte
	switch contentType {
	case "image/jpeg":
		content = minimalJPEG()
	case "image/png":
		content = minimalPNG()
	case "image/webp":
		content = minimalWebP()
	default:
		panic("unsupported test content type: " + contentType)
	}
	return AttachmentInput{
		OriginalName: filename,
		ContentType:  contentType,
		SizeBytes:    uint64(len(content)),
		Content:      content,
	}
}

func minimalJPEG() []byte {
	return []byte{0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00, 0x01, 0x01, 0x00}
}

func minimalPNG() []byte {
	return []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a}
}

func minimalWebP() []byte {
	return []byte{'R', 'I', 'F', 'F', 0x04, 0x00, 0x00, 0x00, 'W', 'E', 'B', 'P', 'V', 'P', '8', ' '}
}
