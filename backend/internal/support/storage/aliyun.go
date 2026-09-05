package storage

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

var aliyunHTTPClient = &http.Client{Timeout: 30 * time.Second}

func saveAliyun(ctx context.Context, src io.Reader, objectKey, filename, contentType string) (*StoredFile, error) {
	return saveAliyunWithClient(ctx, aliyunHTTPClient, src, objectKey, filename, contentType)
}

func deleteAliyun(ctx context.Context, objectKey string) error {
	return deleteAliyunWithClient(ctx, aliyunHTTPClient, objectKey)
}

func deleteAliyunWithClient(ctx context.Context, client *http.Client, objectKey string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if client == nil {
		return fmt.Errorf("阿里云 OSS HTTP 客户端未初始化")
	}
	cfg := currentOSSConfig().Aliyun
	accessKeyID := strings.TrimSpace(cfg.AccessKeyID)
	accessKeySecret := strings.TrimSpace(cfg.AccessKeySecret)
	endpoint := strings.TrimSpace(cfg.Endpoint)
	bucket := strings.TrimSpace(cfg.Bucket)
	if accessKeyID == "" || accessKeySecret == "" || endpoint == "" || bucket == "" || strings.TrimSpace(objectKey) == "" {
		return fmt.Errorf("阿里云 OSS 删除配置不完整")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, aliyunObjectURL(endpoint, bucket, objectKey), nil)
	if err != nil {
		return fmt.Errorf("创建阿里云 OSS 删除请求失败")
	}
	date := time.Now().UTC().Format(http.TimeFormat)
	req.Header.Set("Date", date)
	req.Header.Set("Authorization", aliyunAuthorizationForMethod(http.MethodDelete, accessKeyID, accessKeySecret, bucket, objectKey, "", date))

	resp, err := client.Do(req)
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		return fmt.Errorf("删除阿里云 OSS 对象失败")
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 512))
	switch resp.StatusCode {
	case http.StatusOK, http.StatusNoContent, http.StatusNotFound:
		return nil
	default:
		return fmt.Errorf("删除阿里云 OSS 对象失败: HTTP %d", resp.StatusCode)
	}
}

func saveAliyunWithClient(ctx context.Context, client *http.Client, src io.Reader, objectKey, filename, contentType string) (*StoredFile, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if client == nil {
		return nil, fmt.Errorf("阿里云 OSS HTTP 客户端未初始化")
	}
	cfg := currentOSSConfig().Aliyun
	accessKeyID := strings.TrimSpace(cfg.AccessKeyID)
	accessKeySecret := strings.TrimSpace(cfg.AccessKeySecret)
	endpoint := strings.TrimSpace(cfg.Endpoint)
	bucket := strings.TrimSpace(cfg.Bucket)
	if accessKeyID == "" || accessKeySecret == "" || endpoint == "" || bucket == "" {
		return nil, fmt.Errorf("阿里云 OSS 配置不完整")
	}

	body, err := io.ReadAll(src)
	if err != nil {
		return nil, fmt.Errorf("读取上传文件失败: %w", err)
	}

	publicURL := aliyunObjectURL(endpoint, bucket, objectKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, publicURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	date := time.Now().UTC().Format(http.TimeFormat)
	req.Header.Set("Date", date)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", aliyunAuthorization(accessKeyID, accessKeySecret, bucket, objectKey, contentType, date))

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("上传阿里云 OSS 失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("上传阿里云 OSS 失败: %s", resp.Status)
	}

	relativeURL := "/" + objectKey
	return &StoredFile{
		URL:         relativeURL,
		RelativeURL: relativeURL,
		Path:        objectKey,
		ObjectKey:   objectKey,
		Filename:    filename,
		IsLocal:     false,
	}, nil
}

func aliyunAuthorization(accessKeyID, accessKeySecret, bucket, objectKey, contentType, date string) string {
	return aliyunAuthorizationForMethod(http.MethodPut, accessKeyID, accessKeySecret, bucket, objectKey, contentType, date)
}

func aliyunAuthorizationForMethod(method, accessKeyID, accessKeySecret, bucket, objectKey, contentType, date string) string {
	canonicalResource := "/" + bucket + "/" + objectKey
	stringToSign := strings.Join([]string{
		method,
		"",
		contentType,
		date,
		canonicalResource,
	}, "\n")
	mac := hmac.New(sha1.New, []byte(accessKeySecret))
	_, _ = mac.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return "OSS " + accessKeyID + ":" + signature
}

func aliyunObjectURL(endpoint, bucket, objectKey string) string {
	return strings.TrimRight(aliyunBucketEndpoint(endpoint, bucket), "/") + "/" + escapeObjectKey(objectKey)
}

func aliyunBucketEndpoint(endpoint, bucket string) string {
	endpoint = strings.TrimSpace(endpoint)
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		endpoint = "https://" + endpoint
	}
	parsed, err := url.Parse(endpoint)
	if err != nil || parsed.Host == "" {
		return strings.TrimRight(endpoint, "/")
	}
	host := parsed.Host
	if !strings.HasPrefix(host, bucket+".") {
		host = bucket + "." + host
	}
	parsed.Host = host
	parsed.Path = ""
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String()
}

func escapeObjectKey(objectKey string) string {
	parts := strings.Split(path.Clean("/"+objectKey), "/")
	escaped := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}
		escaped = append(escaped, url.PathEscape(part))
	}
	return strings.Join(escaped, "/")
}
