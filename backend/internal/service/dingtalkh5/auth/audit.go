package auth

import (
	"context"
	"errors"
	"net"

	"gorm.io/gorm"
	"wecheckin/backend/internal/support/auditlog"
)

func auditAuthLookupFailure(ctx context.Context, err error, missingCode, missingReason string) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		auditlog.SetReason(ctx, missingCode, missingReason)
	} else {
		auditlog.SetReason(ctx, "system_error", "认证服务异常")
	}
}

func auditAuthCodeExchangeFailure(ctx context.Context, err error) {
	var networkErr net.Error
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) || errors.As(err, &networkErr) {
		auditlog.SetReason(ctx, "system_error", "钉钉认证服务异常")
		return
	}
	// The upstream client returns untyped errors; do not infer invalid credentials
	// from an arbitrary provider message or copy its potentially sensitive content.
	auditlog.SetReason(ctx, "auth_code_exchange_failed", "钉钉授权码换取身份失败")
}

func auditBindingFailure(ctx context.Context, err error) {
	// Binding business denials currently use fixed plain errors.
	switch err.Error() {
	case "钉钉账号绑定已停用，请联系管理员", "该系统账号绑定已停用，请联系管理员":
		auditlog.SetReason(ctx, "binding_disabled", "账号绑定已停用")
	case "该钉钉账号已绑定其他系统用户，请联系管理员", "该系统账号已绑定其他钉钉用户，请联系管理员":
		auditlog.SetReason(ctx, "binding_conflict", "账号已绑定其他用户")
	default:
		auditlog.SetReason(ctx, "system_error", "账号绑定服务异常")
	}
}
