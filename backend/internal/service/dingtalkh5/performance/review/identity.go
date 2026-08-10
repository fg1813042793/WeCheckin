package review

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	usersvc "wecheckin/backend/internal/service/dingtalkh5/performance/user"
	permissionsupport "wecheckin/backend/internal/support/permission"
)

var normalizeUserIDRegexp = regexp.MustCompile(`[^a-z0-9_.-]+`)

func NormalizeUserID(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	return strings.Trim(normalizeUserIDRegexp.ReplaceAllString(value, ""), ".-_")
}

func sanitizeAvatarURL(value string) (string, error) {
	value = strings.TrimSpace(value)
	if len([]rune(value)) > 500 {
		return "", fmt.Errorf("头像地址不能超过 500 个字符")
	}
	if value == "" {
		return "", nil
	}
	lower := strings.ToLower(value)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") || strings.HasPrefix(value, "/") {
		return value, nil
	}
	return "", fmt.Errorf("头像地址需使用 http(s) 或站内相对路径")
}

func activeRoleIDsForPerfUserContext(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser) ([]uint, error) {
	if user == nil {
		return nil, nil
	}
	roleIDs := usersvc.UniqueUintIDs(append([]uint{user.RoleID}, user.RoleIDs...))
	if len(roleIDs) > 0 || db == nil || user.ID == 0 {
		user.RoleIDs = roleIDs
		return roleIDs, nil
	}
	roleIDs, err := permissionsupport.ActiveRoleIDsForUserContext(ctx, db, user.ID, user.RoleID)
	if err != nil {
		return nil, err
	}
	user.RoleIDs = roleIDs
	return roleIDs, nil
}
