package application

import "errors"

var (
	ErrInvalidArgument   = errors.New("请求参数不正确")
	ErrInvalidLogo       = errors.New("Logo 必须是站内绝对路径或不含凭证的 HTTP(S) 地址")
	ErrVersionConflict   = errors.New("配置已更新，请刷新后重试")
	ErrForbidden         = errors.New("未登录或无权管理该来源")
	ErrSourceUnavailable = errors.New("来源不存在、未发布或不是公开活动")
	ErrItemNotFound      = errors.New("展示项不存在，请刷新后重试")
	ErrDuplicateSource   = errors.New("该来源已添加，请勿重复添加")
	ErrItemLimit         = errors.New("最多配置200个展示项")
	ErrSchemaMissing     = errors.New("公共入口数据结构未就绪，请联系管理员在维护窗口执行 backend/init.sh 后重试")
)

// PublicErrorMessage only exposes known errors, never database/driver details.
func PublicErrorMessage(err error) string {
	for _, known := range []error{ErrInvalidArgument, ErrInvalidLogo, ErrVersionConflict, ErrForbidden, ErrSourceUnavailable, ErrItemNotFound, ErrDuplicateSource, ErrItemLimit, ErrSchemaMissing} {
		if errors.Is(err, known) {
			return known.Error()
		}
	}
	return ""
}
