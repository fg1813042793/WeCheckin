package storage

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"wecheckin/backend/internal/config"
)

func TestSaveMultipartFileStoresLocalUploadUnderConfiguredRoot(t *testing.T) {
	oldCfg := config.Cfg
	uploadRoot := t.TempDir()
	config.Cfg = &config.Config{OSS: config.OSSConfig{
		Type:  "local",
		Local: config.LocalOSSConfig{Path: uploadRoot},
	}}
	t.Cleanup(func() { config.Cfg = oldCfg })

	file := multipartFileHeader(t, "avatar.png", "hello")
	stored, err := SaveMultipartFile(context.Background(), file, SaveOptions{
		Prefix:   "uploads",
		Filename: "avatar.png",
		Now:      time.Date(2026, 8, 6, 10, 0, 0, 0, time.Local),
	})
	if err != nil {
		t.Fatalf("save local upload: %v", err)
	}

	if !stored.IsLocal {
		t.Fatalf("stored file should be local")
	}
	if stored.URL != "/uploads/2026/08/06/avatar.png" {
		t.Fatalf("url = %q", stored.URL)
	}
	if !strings.HasPrefix(stored.Path, uploadRoot) {
		t.Fatalf("path = %q, want under %q", stored.Path, uploadRoot)
	}
	content, err := os.ReadFile(filepath.Join(uploadRoot, "2026", "08", "06", "avatar.png"))
	if err != nil {
		t.Fatalf("read stored file: %v", err)
	}
	if string(content) != "hello" {
		t.Fatalf("content = %q", content)
	}
}

func TestAliyunObjectURLUsesBucketEndpointAndEscapesKey(t *testing.T) {
	got := aliyunObjectURL("oss-cn-hangzhou.aliyuncs.com", "demo-bucket", "uploads/2026/08/06/头像.png")
	want := "https://demo-bucket.oss-cn-hangzhou.aliyuncs.com/uploads/2026/08/06/%E5%A4%B4%E5%83%8F.png"
	if got != want {
		t.Fatalf("aliyun object url = %q, want %q", got, want)
	}
}

func TestSaveMultipartFileStoresAliyunUploadWithRelativeResourcePath(t *testing.T) {
	oldCfg := config.Cfg
	config.Cfg = &config.Config{OSS: config.OSSConfig{
		Type: "aliyun",
		Aliyun: config.AliyunOSSConfig{
			AccessKeyID:     "ak",
			AccessKeySecret: "sk",
			Endpoint:        "oss-cn-hangzhou.aliyuncs.com",
			Bucket:          "demo-bucket",
		},
	}}
	t.Cleanup(func() { config.Cfg = oldCfg })

	oldClient := http.DefaultClient
	var uploadURL string
	http.DefaultClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		uploadURL = req.URL.String()
		if req.Method != http.MethodPut {
			t.Fatalf("method = %s, want PUT", req.Method)
		}
		if !strings.HasPrefix(req.Header.Get("Authorization"), "OSS ak:") {
			t.Fatalf("authorization header = %q", req.Header.Get("Authorization"))
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("")),
			Request:    req,
		}, nil
	})}
	t.Cleanup(func() { http.DefaultClient = oldClient })

	file := multipartFileHeader(t, "avatar.png", "hello")
	stored, err := SaveMultipartFile(context.Background(), file, SaveOptions{
		Prefix:   "uploads",
		Filename: "avatar.png",
		Now:      time.Date(2026, 8, 6, 10, 0, 0, 0, time.Local),
	})
	if err != nil {
		t.Fatalf("save aliyun upload: %v", err)
	}

	wantURL := "/uploads/2026/08/06/avatar.png"
	if stored.URL != wantURL || stored.RelativeURL != wantURL {
		t.Fatalf("url = %q relative = %q, want %q", stored.URL, stored.RelativeURL, wantURL)
	}
	if stored.Path != "uploads/2026/08/06/avatar.png" {
		t.Fatalf("path = %q", stored.Path)
	}
	if stored.Domain != "" {
		t.Fatalf("domain = %q, want empty so static domain config owns public access", stored.Domain)
	}
	if uploadURL != "https://demo-bucket.oss-cn-hangzhou.aliyuncs.com/uploads/2026/08/06/avatar.png" {
		t.Fatalf("upload URL = %q", uploadURL)
	}
}

func multipartFileHeader(t *testing.T, filename, content string) *multipart.FileHeader {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write([]byte(content)); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}
	reader := multipart.NewReader(&body, writer.Boundary())
	form, err := reader.ReadForm(1024)
	if err != nil {
		t.Fatalf("read multipart form: %v", err)
	}
	files := form.File["file"]
	if len(files) != 1 {
		t.Fatalf("multipart file count = %d", len(files))
	}
	return files[0]
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}
