package workflowhandler

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
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

func TestWorkflowUpdateMultipartDisplayNamePresence(t *testing.T) {
	t.Run("omitted preserves existing value", func(t *testing.T) {
		request := workflowUpdateRequestFromMultipart(multipartWorkflowContext(t, nil))
		if request.DisplayName != nil {
			t.Fatalf("display name = %q, want omitted", *request.DisplayName)
		}
	})

	t.Run("explicit empty value clears display name", func(t *testing.T) {
		empty := ""
		request := workflowUpdateRequestFromMultipart(multipartWorkflowContext(t, &empty))
		if request.DisplayName == nil || *request.DisplayName != "" {
			t.Fatalf("display name = %#v, want explicit empty value", request.DisplayName)
		}
	})
}

func multipartWorkflowContext(t *testing.T, displayName *string) *app.RequestContext {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("name", "流程管理名称"); err != nil {
		t.Fatal(err)
	}
	if displayName != nil {
		if err := writer.WriteField("displayName", *displayName); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	ctx := app.NewContext(0)
	ctx.Request.Header.Set("Content-Type", writer.FormDataContentType())
	ctx.Request.SetBody(body.Bytes())
	return ctx
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
