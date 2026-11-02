package admin

import (
	"strings"

	"wecheckin/backend/internal/support/adminrouteperm"
)

var adminAPIPermissionNames = func() map[string]string {
	names := make(map[string]string)
	for _, declaration := range adminrouteperm.Declarations() {
		names[declaration.Perms] = declaration.Name
	}
	return names
}()

func missingAdminPermissionMessage(required string) string {
	entries := make([]string, 0)
	for _, code := range permissionCodes(required) {
		entry := adminrouteperm.KeyForPerms(code)
		if name := adminAPIPermissionNames[code]; name != "" {
			entry += "（" + name + "）"
		}
		entries = append(entries, entry)
	}
	message := "缺少权限-" + strings.Join(entries, " 或 ")
	if len(entries) > 1 {
		message += "（满足任一项即可）"
	}
	return message
}
