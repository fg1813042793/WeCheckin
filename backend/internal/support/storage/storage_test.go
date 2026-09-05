package storage

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"wecheckin/backend/internal/config"
)

func TestDeleteStoredFileTreatsNilAsSuccess(t *testing.T) {
	if err := DeleteStoredFile(context.Background(), nil); err != nil {
		t.Fatalf("DeleteStoredFile(nil) error = %v, want nil", err)
	}
}

func TestDeleteStoredFileRejectsEmptyLocalPath(t *testing.T) {
	err := DeleteStoredFile(context.Background(), &StoredFile{IsLocal: true})
	if err == nil {
		t.Fatal("DeleteStoredFile(empty local path) error = nil")
	}
	if err.Error() != "删除本地存储对象失败: 本地路径为空" {
		t.Fatalf("DeleteStoredFile(empty local path) error = %q", err)
	}
	if errors.Is(err, os.ErrNotExist) {
		t.Fatalf("DeleteStoredFile(empty local path) incorrectly treated as missing file: %v", err)
	}
}

func TestDeleteStoredFileRejectsEmptyObjectKey(t *testing.T) {
	err := DeleteStoredFile(context.Background(), &StoredFile{})
	if err == nil {
		t.Fatal("DeleteStoredFile(empty object key) error = nil")
	}
	if err.Error() != "删除存储对象失败: 对象键为空" {
		t.Fatalf("DeleteStoredFile(empty object key) error = %q", err)
	}
}

func TestDeleteStoredFileRemovesLocalPath(t *testing.T) {
	localPath := filepath.Join(t.TempDir(), "feedback.png")
	if err := os.WriteFile(localPath, []byte("image"), 0600); err != nil {
		t.Fatalf("write local object: %v", err)
	}

	err := DeleteStoredFile(context.Background(), &StoredFile{
		LocalPath: localPath,
		IsLocal:   true,
	})
	if err != nil {
		t.Fatalf("DeleteStoredFile(local) error = %v", err)
	}
	if _, err := os.Stat(localPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("local object still exists or stat failed: %v", err)
	}
}

func TestDeleteStoredFileTreatsMissingLocalPathAsSuccess(t *testing.T) {
	missingPath := filepath.Join(t.TempDir(), "missing.png")
	if err := DeleteStoredFile(context.Background(), &StoredFile{LocalPath: missingPath, IsLocal: true}); err != nil {
		t.Fatalf("DeleteStoredFile(missing local) error = %v, want nil", err)
	}
}

func TestDeleteStoredFileWrapsLocalDeleteError(t *testing.T) {
	nonEmptyDir := filepath.Join(t.TempDir(), "private-feedback-name.png")
	if err := os.Mkdir(nonEmptyDir, 0755); err != nil {
		t.Fatalf("mkdir local object: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nonEmptyDir, "child"), []byte("keep"), 0600); err != nil {
		t.Fatalf("write child: %v", err)
	}

	err := DeleteStoredFile(context.Background(), &StoredFile{LocalPath: nonEmptyDir, IsLocal: true})
	if err == nil {
		t.Fatal("DeleteStoredFile(non-empty directory) error = nil")
	}
	if !strings.Contains(err.Error(), "删除本地存储对象失败") {
		t.Fatalf("DeleteStoredFile(non-empty directory) error = %v, want contextual error", err)
	}
	if strings.Contains(err.Error(), nonEmptyDir) {
		t.Fatalf("DeleteStoredFile(non-empty directory) leaked full local path: %v", err)
	}
	if strings.Contains(err.Error(), filepath.Base(nonEmptyDir)) {
		t.Fatalf("DeleteStoredFile(non-empty directory) leaked user filename: %v", err)
	}
}

func TestDeleteStoredFileHonorsCanceledContextForLocalObject(t *testing.T) {
	localPath := filepath.Join(t.TempDir(), "feedback.png")
	if err := os.WriteFile(localPath, []byte("image"), 0600); err != nil {
		t.Fatalf("write local object: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := DeleteStoredFile(ctx, &StoredFile{LocalPath: localPath, IsLocal: true})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("DeleteStoredFile(canceled local) error = %v, want context.Canceled", err)
	}
	if _, err := os.Stat(localPath); err != nil {
		t.Fatalf("canceled delete changed local object: %v", err)
	}
}

func TestDeleteStoredFileSignsEscapedAliyunDeleteRequest(t *testing.T) {
	withAliyunTestConfig(t)
	oldClient := aliyunHTTPClient
	objectKey := "uploads/feedback/2026/09/05/反馈 图.png"
	aliyunHTTPClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodDelete {
			t.Fatalf("method = %s, want DELETE", req.Method)
		}
		wantPath := "/uploads/feedback/2026/09/05/%E5%8F%8D%E9%A6%88%20%E5%9B%BE.png"
		if req.URL.EscapedPath() != wantPath {
			t.Fatalf("escaped path = %q, want %q", req.URL.EscapedPath(), wantPath)
		}
		date := req.Header.Get("Date")
		if date == "" {
			t.Fatal("Date header is empty")
		}
		wantAuthorization := expectedAliyunAuthorization(http.MethodDelete, "ak", "sk", "demo-bucket", objectKey, "", date)
		if got := req.Header.Get("Authorization"); got == "" || got != wantAuthorization {
			t.Fatalf("Authorization = %q, want DELETE signature %q", got, wantAuthorization)
		}
		return storageHTTPResponse(req, http.StatusNoContent, ""), nil
	})}
	t.Cleanup(func() { aliyunHTTPClient = oldClient })

	err := DeleteStoredFile(context.Background(), &StoredFile{ObjectKey: objectKey})
	if err != nil {
		t.Fatalf("DeleteStoredFile(aliyun) error = %v", err)
	}
}

func TestDeleteStoredFileDoesNotFollowAliyunRedirect(t *testing.T) {
	withSensitiveAliyunTestConfig(t)
	objectKey := "uploads/feedback/private-user-content.png"
	methods := make([]string, 0, 2)
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		methods = append(methods, req.Method)
		if len(methods) == 1 {
			resp := storageHTTPResponse(req, http.StatusFound, "private-redirect-response")
			resp.Header.Set("Location", "https://redirect.example.test/private-target")
			return resp, nil
		}
		return storageHTTPResponse(req, http.StatusOK, "redirect followed"), nil
	})}

	err := deleteAliyunWithClient(context.Background(), client, objectKey)
	if err == nil {
		t.Fatal("deleteAliyunWithClient(redirect) error = nil")
	}
	if err.Error() != "删除阿里云 OSS 对象失败: HTTP 302" {
		t.Fatalf("deleteAliyunWithClient(redirect) error = %q", err)
	}
	if len(methods) != 1 || methods[0] != http.MethodDelete {
		t.Fatalf("request methods = %v, want one DELETE and no redirected GET", methods)
	}
	assertErrorOmits(t, err, objectKey, "private-user-content.png", "redirect.example.test", "private-target", "private-redirect-response")
}

func TestDeleteStoredFileTreatsAliyunSuccessAndMissingStatusesAsSuccess(t *testing.T) {
	for _, statusCode := range []int{http.StatusOK, http.StatusNoContent, http.StatusNotFound} {
		statusCode := statusCode
		t.Run(fmt.Sprintf("status_%d", statusCode), func(t *testing.T) {
			withAliyunTestConfig(t)
			oldClient := aliyunHTTPClient
			aliyunHTTPClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return storageHTTPResponse(req, statusCode, "already gone"), nil
			})}
			t.Cleanup(func() { aliyunHTTPClient = oldClient })

			if err := DeleteStoredFile(context.Background(), &StoredFile{ObjectKey: "uploads/feedback/image.png"}); err != nil {
				t.Fatalf("DeleteStoredFile status %d error = %v, want nil", statusCode, err)
			}
		})
	}
}

func TestDeleteStoredFileHonorsCanceledContextForAliyunObject(t *testing.T) {
	withAliyunTestConfig(t)
	oldClient := aliyunHTTPClient
	calls := 0
	aliyunHTTPClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		calls++
		return nil, errors.New("request should not be sent")
	})}
	t.Cleanup(func() { aliyunHTTPClient = oldClient })
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := DeleteStoredFile(ctx, &StoredFile{ObjectKey: "uploads/feedback/image.png"})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("DeleteStoredFile(canceled aliyun) error = %v, want context.Canceled", err)
	}
	if calls != 0 {
		t.Fatalf("round trip calls = %d, want 0", calls)
	}
}

func TestDeleteStoredFileSanitizesAliyunRequestConstructionErrorAndPreservesCause(t *testing.T) {
	oldCfg := config.Cfg
	endpoint := "http://[private-endpoint-secret"
	objectKey := "uploads/feedback/private-user-content.png"
	config.Cfg = &config.Config{OSS: config.OSSConfig{Aliyun: config.AliyunOSSConfig{
		Endpoint:        endpoint,
		Bucket:          "demo-bucket",
		AccessKeyID:     "sensitive-access-key",
		AccessKeySecret: "sensitive-secret",
	}}}
	t.Cleanup(func() { config.Cfg = oldCfg })

	err := deleteAliyunWithClient(context.Background(), &http.Client{}, objectKey)
	if err == nil {
		t.Fatal("deleteAliyunWithClient(invalid endpoint) error = nil")
	}
	if err.Error() != "删除阿里云 OSS 对象失败" {
		t.Fatalf("deleteAliyunWithClient(invalid endpoint) error text = %q", err)
	}
	assertErrorOmits(t, err, endpoint, objectKey, "private-user-content.png", "missing ']'", "sensitive-access-key", "sensitive-secret")
	var urlErr *url.Error
	if !errors.As(err, &urlErr) {
		t.Fatalf("deleteAliyunWithClient(invalid endpoint) error = %v, want wrapped *url.Error", err)
	}
	if !strings.Contains(urlErr.Error(), "private-endpoint-secret") {
		t.Fatalf("wrapped URL error = %q, want original construction detail", urlErr)
	}
}

func TestDeleteStoredFileSanitizesAliyunStatusError(t *testing.T) {
	withSensitiveAliyunTestConfig(t)
	oldClient := aliyunHTTPClient
	objectKey := "uploads/feedback/private-user-content.png"
	aliyunHTTPClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		resp := storageHTTPResponse(req, http.StatusForbidden, "private-response-body")
		resp.Status = "403 private-user-content.png sensitive-access-key sensitive-secret " + req.Header.Get("Authorization")
		return resp, nil
	})}
	t.Cleanup(func() { aliyunHTTPClient = oldClient })

	err := DeleteStoredFile(context.Background(), &StoredFile{ObjectKey: objectKey})
	if err == nil {
		t.Fatal("DeleteStoredFile(forbidden aliyun) error = nil")
	}
	assertErrorOmits(t, err, "private-user-content.png", "sensitive-access-key", "sensitive-secret", "Authorization", "private-response-body", "OSS sensitive-access-key:")
}

func TestDeleteStoredFileSanitizesAliyunTransportError(t *testing.T) {
	withSensitiveAliyunTestConfig(t)
	oldClient := aliyunHTTPClient
	transportErr := errors.New("private-user-content.png sensitive-access-key sensitive-secret")
	aliyunHTTPClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return nil, fmt.Errorf("%w %s", transportErr, req.Header.Get("Authorization"))
	})}
	t.Cleanup(func() { aliyunHTTPClient = oldClient })

	err := DeleteStoredFile(context.Background(), &StoredFile{ObjectKey: "uploads/feedback/private-user-content.png"})
	if err == nil {
		t.Fatal("DeleteStoredFile(transport failure) error = nil")
	}
	if !errors.Is(err, transportErr) {
		t.Fatalf("DeleteStoredFile(transport failure) error = %v, want transport cause", err)
	}
	if err.Error() != "删除阿里云 OSS 对象失败" {
		t.Fatalf("DeleteStoredFile(transport failure) error text = %q", err)
	}
	assertErrorOmits(t, err, "private-user-content.png", "sensitive-access-key", "sensitive-secret", "OSS sensitive-access-key:")
}

func TestDeleteStoredFilePreservesAliyunTransportContextCause(t *testing.T) {
	withAliyunTestConfig(t)
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return nil, context.DeadlineExceeded
	})}

	err := deleteAliyunWithClient(context.Background(), client, "uploads/feedback/image.png")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("deleteAliyunWithClient(deadline transport failure) error = %v, want context deadline", err)
	}
	if err.Error() != "删除阿里云 OSS 对象失败" {
		t.Fatalf("deleteAliyunWithClient(deadline transport failure) error text = %q", err)
	}
}

func TestSaveReaderStoresLocalUploadWithoutMultipartReconstruction(t *testing.T) {
	oldCfg := config.Cfg
	uploadRoot := t.TempDir()
	config.Cfg = &config.Config{OSS: config.OSSConfig{
		Type:  "local",
		Local: config.LocalOSSConfig{Path: uploadRoot},
	}}
	t.Cleanup(func() { config.Cfg = oldCfg })

	stored, err := SaveReader(context.Background(), strings.NewReader("image-bytes"), "original.png", SaveOptions{
		Prefix:   "uploads/feedback",
		Filename: "stored.png",
		Now:      time.Date(2026, 9, 5, 10, 0, 0, 0, time.Local),
	})
	if err != nil {
		t.Fatalf("SaveReader() error = %v", err)
	}
	if stored.ObjectKey != "uploads/feedback/2026/09/05/stored.png" {
		t.Fatalf("ObjectKey = %q", stored.ObjectKey)
	}
	content, err := os.ReadFile(filepath.Join(uploadRoot, "feedback", "2026", "09", "05", "stored.png"))
	if err != nil {
		t.Fatalf("read stored reader content: %v", err)
	}
	if string(content) != "image-bytes" {
		t.Fatalf("stored content = %q", content)
	}
}

func TestSaveReaderRemovesPartialLocalFileOnCopyFailure(t *testing.T) {
	oldCfg := config.Cfg
	uploadRoot := t.TempDir()
	config.Cfg = &config.Config{OSS: config.OSSConfig{
		Type:  "local",
		Local: config.LocalOSSConfig{Path: uploadRoot},
	}}
	t.Cleanup(func() { config.Cfg = oldCfg })

	_, err := SaveReader(context.Background(), &partialErrorReader{}, "original.png", SaveOptions{
		Prefix:   "uploads/feedback",
		Filename: "partial.png",
		Now:      time.Date(2026, 9, 5, 10, 0, 0, 0, time.Local),
	})
	if err == nil {
		t.Fatal("SaveReader(partial failure) error = nil")
	}
	localPath := filepath.Join(uploadRoot, "feedback", "2026", "09", "05", "partial.png")
	if _, statErr := os.Stat(localPath); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("partial local file remains after save failure: %v", statErr)
	}
}

func TestSaveReaderHonorsPreCanceledContextBeforeReading(t *testing.T) {
	oldCfg := config.Cfg
	uploadRoot := t.TempDir()
	config.Cfg = &config.Config{OSS: config.OSSConfig{
		Type:  "local",
		Local: config.LocalOSSConfig{Path: uploadRoot},
	}}
	t.Cleanup(func() { config.Cfg = oldCfg })
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	reader := &trackingReader{content: []byte("must-not-read")}

	_, err := SaveReader(ctx, reader, "original.png", SaveOptions{
		Prefix:   "uploads/feedback",
		Filename: "pre-canceled.png",
		Now:      time.Date(2026, 9, 5, 10, 0, 0, 0, time.Local),
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("SaveReader(pre-canceled) error = %v, want context.Canceled", err)
	}
	if reader.reads != 0 {
		t.Fatalf("SaveReader(pre-canceled) reads = %d, want 0", reader.reads)
	}
	localPath := filepath.Join(uploadRoot, "feedback", "2026", "09", "05", "pre-canceled.png")
	if _, statErr := os.Stat(localPath); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("pre-canceled save created local file: %v", statErr)
	}
}

func TestSaveReaderRemovesPartialLocalFileWhenContextCanceledDuringCopy(t *testing.T) {
	oldCfg := config.Cfg
	uploadRoot := t.TempDir()
	config.Cfg = &config.Config{OSS: config.OSSConfig{
		Type:  "local",
		Local: config.LocalOSSConfig{Path: uploadRoot},
	}}
	t.Cleanup(func() { config.Cfg = oldCfg })
	ctx, cancel := context.WithCancel(context.Background())
	reader := &cancelAfterReadReader{cancel: cancel}

	_, err := SaveReader(ctx, reader, "original.png", SaveOptions{
		Prefix:   "uploads/feedback",
		Filename: "mid-copy.png",
		Now:      time.Date(2026, 9, 5, 10, 0, 0, 0, time.Local),
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("SaveReader(mid-copy cancel) error = %v, want context.Canceled", err)
	}
	localPath := filepath.Join(uploadRoot, "feedback", "2026", "09", "05", "mid-copy.png")
	if _, statErr := os.Stat(localPath); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("mid-copy cancellation left partial local file: %v", statErr)
	}
}

func TestSaveReaderAliyunReadHonorsCancellationBeforeSendingRequest(t *testing.T) {
	withAliyunTestConfig(t)
	oldClient := aliyunHTTPClient
	requests := 0
	aliyunHTTPClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		requests++
		return storageHTTPResponse(req, http.StatusOK, ""), nil
	})}
	t.Cleanup(func() { aliyunHTTPClient = oldClient })
	ctx, cancel := context.WithCancel(context.Background())
	reader := &cancelThenErrorReader{cancel: cancel}

	_, err := SaveReader(ctx, reader, "original.png", SaveOptions{
		Prefix:   "uploads/feedback",
		Filename: "canceled.png",
		Now:      time.Date(2026, 9, 5, 10, 0, 0, 0, time.Local),
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("SaveReader(aliyun read canceled) error = %v, want context.Canceled", err)
	}
	if reader.reads != 1 {
		t.Fatalf("SaveReader(aliyun read canceled) reads = %d, want 1", reader.reads)
	}
	if requests != 0 {
		t.Fatalf("SaveReader(aliyun read canceled) requests = %d, want 0", requests)
	}
}

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

	oldClient := aliyunHTTPClient
	var uploadURL string
	aliyunHTTPClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
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
	t.Cleanup(func() { aliyunHTTPClient = oldClient })

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

func TestAliyunHTTPClientHasFiniteTimeout(t *testing.T) {
	if aliyunHTTPClient == nil || aliyunHTTPClient.Timeout != 30*time.Second {
		t.Fatalf("aliyun HTTP client timeout = %v, want 30s", aliyunHTTPClient)
	}
}

func TestSaveAliyunHonorsCanceledContextWithoutSendingRequest(t *testing.T) {
	withAliyunTestConfig(t)
	calls := 0
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		calls++
		return nil, errors.New("request should not be sent")
	})}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := saveAliyunWithClient(ctx, client, strings.NewReader("body"), "uploads/test.txt", "test.txt", "text/plain")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("saveAliyunWithClient() error = %v, want context cancellation", err)
	}
	if calls != 0 {
		t.Fatalf("round trip calls = %d, want 0", calls)
	}
}

func TestSaveAliyunDoesNotExposeUpstreamResponseBody(t *testing.T) {
	withAliyunTestConfig(t)
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusBadGateway,
			Status:     "502 Bad Gateway",
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("upstream-secret-detail")),
			Request:    req,
		}, nil
	})}

	_, err := saveAliyunWithClient(context.Background(), client, strings.NewReader("body"), "uploads/test.txt", "test.txt", "text/plain")
	if err == nil {
		t.Fatal("saveAliyunWithClient() error = nil")
	}
	if strings.Contains(err.Error(), "upstream-secret-detail") {
		t.Fatalf("OSS error leaked upstream response body: %v", err)
	}
}

func withAliyunTestConfig(t *testing.T) {
	t.Helper()
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
}

func withSensitiveAliyunTestConfig(t *testing.T) {
	t.Helper()
	oldCfg := config.Cfg
	config.Cfg = &config.Config{OSS: config.OSSConfig{
		Type: "aliyun",
		Aliyun: config.AliyunOSSConfig{
			AccessKeyID:     "sensitive-access-key",
			AccessKeySecret: "sensitive-secret",
			Endpoint:        "oss-cn-hangzhou.aliyuncs.com",
			Bucket:          "demo-bucket",
		},
	}}
	t.Cleanup(func() { config.Cfg = oldCfg })
}

func expectedAliyunAuthorization(method, accessKeyID, accessKeySecret, bucket, objectKey, contentType, date string) string {
	canonicalResource := "/" + bucket + "/" + objectKey
	stringToSign := strings.Join([]string{method, "", contentType, date, canonicalResource}, "\n")
	mac := hmac.New(sha1.New, []byte(accessKeySecret))
	_, _ = mac.Write([]byte(stringToSign))
	return "OSS " + accessKeyID + ":" + base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func storageHTTPResponse(req *http.Request, statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Status:     fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode)),
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    req,
	}
}

func assertErrorOmits(t *testing.T, err error, forbidden ...string) {
	t.Helper()
	for _, value := range forbidden {
		if strings.Contains(err.Error(), value) {
			t.Errorf("error %q leaked %q", err, value)
		}
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

type partialErrorReader struct {
	wrote bool
}

type trackingReader struct {
	content []byte
	reads   int
}

func (reader *trackingReader) Read(buffer []byte) (int, error) {
	reader.reads++
	if len(reader.content) == 0 {
		return 0, io.EOF
	}
	n := copy(buffer, reader.content)
	reader.content = reader.content[n:]
	return n, nil
}

type cancelAfterReadReader struct {
	cancel context.CancelFunc
	read   bool
}

func (reader *cancelAfterReadReader) Read(buffer []byte) (int, error) {
	if reader.read {
		return 0, io.EOF
	}
	reader.read = true
	n := copy(buffer, "partial")
	reader.cancel()
	return n, nil
}

type cancelThenErrorReader struct {
	cancel context.CancelFunc
	reads  int
}

func (reader *cancelThenErrorReader) Read(buffer []byte) (int, error) {
	reader.reads++
	if reader.reads > 1 {
		return 0, errors.New("reader continued after cancellation")
	}
	n := copy(buffer, "partial")
	reader.cancel()
	return n, nil
}

func (reader *partialErrorReader) Read(buffer []byte) (int, error) {
	if reader.wrote {
		return 0, errors.New("forced reader failure")
	}
	reader.wrote = true
	return copy(buffer, "partial"), nil
}
