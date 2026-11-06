package adminuser

import (
	"errors"
	"regexp"

	"gorm.io/gorm"
	"wecheckin/backend/internal/support/permissiondelegation"
)

var publicPermissionKeyPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+(?::[a-zA-Z0-9_-]+)+$`)

// UserWriteErrorMessage exposes only known business errors, never an arbitrary
// wrapped error or database message. An empty result requires internal logging.
func UserWriteErrorMessage(err error) string {
	switch {
	case errors.Is(err, ErrUserNameRequired):
		return "请输入用户名"
	case errors.Is(err, ErrUserNameTaken):
		return "用户名已存在，请更换后重试"
	case errors.Is(err, ErrLoginPasswordRequired):
		return "请先设置登录密码，再分配角色"
	case errors.Is(err, ErrSelfManager):
		return "用户不能选择自己作为直属上级"
	case errors.Is(err, ErrManagerUnavailable):
		return "直属上级用户不存在或已停用"
	case errors.Is(err, ErrInvalidAdmin):
		return "未登录或管理员身份失效，请重新登录"
	case errors.Is(err, permissiondelegation.ErrReservedRole):
		return "只有超级管理员可以管理超级管理员角色"
	case errors.Is(err, permissiondelegation.ErrRoleExceeded):
		if errors.Is(err, permissiondelegation.ErrPermissionExceeded) {
			return "目标用户现有角色或所选角色超出当前管理员权限范围" + userWritePermissionSuffix(err)
		}
		if errors.Is(err, permissiondelegation.ErrDataScopeExceeded) {
			return "目标用户现有角色或所选角色的数据范围超出当前管理员可授权范围"
		}
		return "所选角色或目标用户现有角色不存在、已停用或超出当前管理员可授权范围"
	case errors.Is(err, permissiondelegation.ErrPermissionExceeded):
		return "不能授予或管理超出当前管理员范围的权限" + userWritePermissionSuffix(err)
	case errors.Is(err, permissiondelegation.ErrDataScopeExceeded):
		return "目标用户的数据权限超出当前管理员可授权范围"
	case errors.Is(err, permissiondelegation.ErrDepartmentExceeded):
		return "所属部门或附加数据部门不在当前管理员可管理范围内，请检查部门选择"
	case errors.Is(err, permissiondelegation.ErrUserExceeded):
		return "直属上级或附加数据用户不在当前管理员可管理范围内"
	case errors.Is(err, gorm.ErrRecordNotFound):
		return "用户不存在或不在当前管理员可管理范围内"
	default:
		return ""
	}
}

func userWritePermissionSuffix(err error) string {
	var detail *permissiondelegation.PermissionExceededError
	if errors.As(err, &detail) && len(detail.PermissionKey) <= 191 && publicPermissionKeyPattern.MatchString(detail.PermissionKey) {
		return "-" + detail.PermissionKey
	}
	return ""
}
