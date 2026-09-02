package workflowhandler

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValidateWorkflowLogoRejectsInvalidFiles(t *testing.T) {
	validPNG := multipartFileHeader(t, "logo.png", append([]byte("\x89PNG\r\n\x1a\n"), make([]byte, 32)...))
	if err := validateWorkflowLogo(validPNG); err != nil {
		t.Fatalf("valid PNG rejected: %v", err)
	}

	invalidContent := multipartFileHeader(t, "logo.png", []byte("not an image"))
	if err := validateWorkflowLogo(invalidContent); err == nil {
		t.Fatal("fake PNG must be rejected")
	}

	invalidExtension := multipartFileHeader(t, "logo.svg", []byte("<svg></svg>"))
	if err := validateWorkflowLogo(invalidExtension); err == nil {
		t.Fatal("SVG logo must be rejected")
	}

	oversized := multipartFileHeader(t, "logo.png", []byte("\x89PNG\r\n\x1a\n"))
	oversized.Size = workflowLogoMaxSize + 1
	if err := validateWorkflowLogo(oversized); err == nil {
		t.Fatal("oversized logo must be rejected")
	}
}

func multipartFileHeader(t *testing.T, filename string, content []byte) *multipart.FileHeader {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("logo", filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	if err := request.ParseMultipartForm(3 * 1024 * 1024); err != nil {
		t.Fatal(err)
	}
	return request.MultipartForm.File["logo"][0]
}
