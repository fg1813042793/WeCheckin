package account

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"

	dingtalkh5service "wecheckin/backend/internal/service/dingtalkh5/account"
	"wecheckin/backend/internal/support/dingtalkh5session"
	"wecheckin/backend/internal/support/media"
	"wecheckin/backend/internal/support/storage"
	"wecheckin/backend/pkg/response"
)

type Handler struct{}

func NewHandler() *Handler { return &Handler{} }

const avatarUploadMaxSize = 5 * 1024 * 1024

func (h *Handler) ChangePassword(ctx context.Context, c *app.RequestContext) {
	user, ok := dingtalkh5session.CurrentUser(c)
	if !ok {
		response.Fail(c, "未登录")
		return
	}
	var req struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
		ConfirmPassword string `json:"confirmPassword"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	if err := dingtalkh5service.ChangePasswordContext(ctx, user, req.CurrentPassword, req.NewPassword, req.ConfirmPassword); err != nil {
		response.FailInternal(ctx, c, "dingtalkh5.account.handler", "操作失败，请稍后重试", err)
		return
	}
	response.JSON(c, nil)
}

func (h *Handler) UpdateProfile(ctx context.Context, c *app.RequestContext) {
	user, ok := dingtalkh5session.CurrentUser(c)
	if !ok {
		response.Fail(c, "未登录")
		return
	}
	var req dingtalkh5service.AccountProfilePayload
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	data, err := dingtalkh5service.UpdateAccountProfileContext(ctx, user, dingtalkh5session.CurrentToken(c), req)
	if err != nil {
		response.FailInternal(ctx, c, "dingtalkh5.account.handler", "操作失败，请稍后重试", err)
		return
	}
	response.JSON(c, data)
}

func (h *Handler) UploadAvatar(ctx context.Context, c *app.RequestContext) {
	user, ok := dingtalkh5session.CurrentUser(c)
	if !ok {
		response.Fail(c, "未登录")
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		response.FailInternal(ctx, c, "dingtalkh5.account.handler", "上传失败，请稍后重试", err)
		return
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !isAvatarExt(ext) {
		response.Fail(c, "不支持的文件格式，仅允许 jpg/png/gif/webp")
		return
	}
	if file.Size <= 0 {
		response.Fail(c, "上传文件为空")
		return
	}
	if file.Size > avatarUploadMaxSize {
		response.Fail(c, "上传文件过大，最大5MB")
		return
	}

	now := time.Now()
	account := dingtalkh5service.NormalizeUserID(user.Account)
	if account == "" {
		account = strconv.FormatUint(uint64(user.ID), 10)
	}
	filename := fmt.Sprintf("%d_avatar_%s%s", now.UnixNano(), account, ext)
	stored, err := storage.SaveMultipartFile(ctx, file, storage.SaveOptions{
		Prefix:   "uploads",
		Filename: filename,
		Now:      now,
	})
	if err != nil {
		response.FailInternal(ctx, c, "dingtalkh5.account.handler", "上传失败，请稍后重试", err)
		return
	}

	publicURL := avatarPublicURL(ctx, c, stored.RelativeURL)
	response.JSON(c, utils.H{
		"url":      publicURL,
		"path":     stored.RelativeURL,
		"filename": filename,
	})
}

func isAvatarExt(ext string) bool {
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
		return true
	default:
		return false
	}
}

func avatarPublicURL(ctx context.Context, c *app.RequestContext, path string) string {
	domain := strings.TrimRight(strings.TrimSpace(media.StaticDomainContext(ctx)), "/")
	if domain != "" && !strings.Contains(domain, "localhost") && !strings.Contains(domain, "127.0.0.1") {
		return media.FullURL(path, domain)
	}
	host := strings.TrimSpace(string(c.Request.Host()))
	if host == "" {
		return media.FullURL(path, domain)
	}
	proto := strings.TrimSpace(string(c.Request.Header.Peek("X-Forwarded-Proto")))
	if strings.Contains(proto, ",") {
		proto = strings.TrimSpace(strings.Split(proto, ",")[0])
	}
	if proto == "" {
		proto = "http"
	}
	return strings.TrimRight(proto+"://"+host, "/") + path
}
