package adminuser

import (
	"errors"
	"fmt"
	"testing"

	"gorm.io/gorm"
	"wecheckin/backend/internal/support/permissiondelegation"
)

func TestUserWriteErrorMessageOnlyExposesKnownBusinessErrors(t *testing.T) {
	for _, test := range []struct {
		err  error
		want string
	}{
		{ErrUserNameRequired, "请输入用户名"},
		{fmt.Errorf("save: %w", ErrUserNameTaken), "用户名已存在，请更换后重试"},
		{ErrLoginPasswordRequired, "请先设置登录密码，再分配角色"},
		{ErrSelfManager, "用户不能选择自己作为直属上级"},
		{ErrManagerUnavailable, "直属上级用户不存在或已停用"},
		{ErrInvalidAdmin, "未登录或管理员身份失效，请重新登录"},
		{gorm.ErrRecordNotFound, "用户不存在或不在当前管理员可管理范围内"},
		{permissiondelegation.ErrDepartmentExceeded, "所属部门或附加数据部门不在当前管理员可管理范围内，请检查部门选择"},
		{permissiondelegation.ErrUserExceeded, "直属上级或附加数据用户不在当前管理员可管理范围内"},
		{permissiondelegation.ErrRoleExceeded, "所选角色或目标用户现有角色不存在、已停用或超出当前管理员可授权范围"},
		{permissiondelegation.ErrDataScopeExceeded, "目标用户的数据权限超出当前管理员可授权范围"},
		{permissiondelegation.ErrReservedRole, "只有超级管理员可以管理超级管理员角色"},
		{nil, ""},
		{errors.New("SQL password=private_token connection failed"), ""},
		{errors.New("用户名已存在，请更换后重试"), ""},
		{fmt.Errorf("SQL password=secret: %w", permissiondelegation.ErrRoleExceeded), "所选角色或目标用户现有角色不存在、已停用或超出当前管理员可授权范围"},
	} {
		if got := UserWriteErrorMessage(test.err); got != test.want {
			t.Errorf("error %T: message=%q want=%q", test.err, got, test.want)
		}
	}
}

func TestUserWritePermissionErrorIdentifiesSafePermissionKey(t *testing.T) {
	ceiling := permissiondelegation.Ceiling{}
	err := ceiling.ValidatePermissionKeys([]string{"admin:api:menu:list"})
	if !errors.Is(err, permissiondelegation.ErrPermissionExceeded) {
		t.Fatal(err)
	}
	if got := UserWriteErrorMessage(err); got != "不能授予或管理超出当前管理员范围的权限-admin:api:menu:list" {
		t.Fatalf("message=%q", got)
	}
	roleError := fmt.Errorf("%w: %w", permissiondelegation.ErrRoleExceeded, err)
	if got := UserWriteErrorMessage(roleError); got != "目标用户现有角色或所选角色超出当前管理员权限范围-admin:api:menu:list" {
		t.Fatalf("role permission reason was lost: %q", got)
	}
	err = ceiling.ValidatePermissionKeys([]string{"admin:api:menu:list\nprivate-data"})
	if got := UserWriteErrorMessage(err); got != "不能授予或管理超出当前管理员范围的权限" {
		t.Fatalf("untrusted detail escaped: %q", got)
	}
}
