package admin

import (
	"context"
	"log"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/support/adminaccess"
	permissionsupport "wecheckin/backend/internal/support/permission"
	"wecheckin/backend/pkg/database"
	"wecheckin/backend/pkg/logger"
)

func AdminPerm() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		adminVal, _ := c.Get("admin")
		admin := adminVal.(*model.Admin)

		db, cancel := database.WithContext(ctx)
		defer cancel()
		roleIDs := admin.RoleIDs
		if len(roleIDs) == 0 {
			var err error
			roleIDs, err = permissionsupport.ActiveRoleIDsForUserContext(ctx, db, admin.ID, admin.RoleID)
			if err != nil {
				auditAdminPermissionDenied(admin, string(c.Path()), "", "role_lookup_failed")
				c.JSON(consts.StatusOK, utils.H{
					"code": 1,
					"msg":  "权限校验失败",
				})
				c.Abort()
				return
			}
			admin.RoleIDs = roleIDs
		}

		// The reserved super admin role bypasses route permission checks.
		if adminaccess.HasReservedSuperAdminRoleWithRoleIDsContext(ctx, db, roleIDs) {
			c.Next(ctx)
			return
		}

		// No role assigned - no access
		if len(roleIDs) == 0 {
			auditAdminPermissionDenied(admin, string(c.Path()), "", "role_missing")
			c.JSON(consts.StatusOK, utils.H{
				"code": 1,
				"msg":  "未分配角色，无权限",
			})
			c.Abort()
			return
		}

		path := string(c.Path())
		// Remove query string if any
		if idx := strings.Index(path, "?"); idx != -1 {
			path = path[:idx]
		}

		required, ok := adminRoutePermission(string(c.Method()), path)
		if !ok {
			auditAdminPermissionDenied(admin, path, "", "route_not_declared")
			c.JSON(consts.StatusOK, utils.H{
				"code": 1,
				"msg":  "无权限访问",
			})
			c.Abort()
			return
		}
		if required == "" {
			// Explicitly allowed after login.
			c.Next(ctx)
			return
		}

		requiredCodes := permissionCodes(required)
		for _, code := range requiredCodes {
			apiKey := permissionsupport.AdminAPIPermissionKey(code)
			if effect, ok, err := permissionsupport.SubjectPermissionEffectContext(ctx, db, permissionsupport.SubjectUser, admin.ID, apiKey); err == nil && ok && effect == permissionsupport.EffectDeny {
				continue
			}
			if ok, err := permissionsupport.SubjectHasPermissionWithRoleIDsContext(ctx, db, admin.ID, roleIDs, apiKey); err == nil && ok {
				c.Next(ctx)
				return
			}
		}

		c.JSON(consts.StatusOK, utils.H{
			"code": 1,
			"msg":  "无权限访问",
		})
		auditAdminPermissionDenied(admin, path, required, "permission_missing")
		c.Abort()
	}
}

func adminRoutePermission(method, path string) (string, bool) {
	method = strings.ToUpper(method)
	if required, ok := routeMethodPerms[method+" "+path]; ok {
		return required, true
	}
	for _, route := range routeMethodPermPatterns {
		if route.method == method && routePatternMatches(route.path, path) {
			return route.perm, true
		}
	}
	return "", false
}

func routePatternMatches(pattern, path string) bool {
	patternParts := routePatternParts(pattern)
	pathParts := routePatternParts(path)
	if len(patternParts) != len(pathParts) {
		return false
	}
	for i := range patternParts {
		if strings.HasPrefix(patternParts[i], ":") {
			continue
		}
		if patternParts[i] != pathParts[i] {
			return false
		}
	}
	return true
}

func routePatternParts(path string) []string {
	path = strings.Trim(path, "/")
	if path == "" {
		return nil
	}
	return strings.Split(path, "/")
}

func auditAdminPermissionDenied(admin *model.Admin, path, required, reason string) {
	writer := logger.Logger
	if writer == nil {
		writer = log.Default()
	}
	if admin == nil {
		writer.Printf("[AdminPermDenied] admin=<nil> path=%s required=%s reason=%s", path, required, reason)
		return
	}
	writer.Printf("[AdminPermDenied] adminId=%d roleId=%d path=%s required=%s reason=%s", admin.ID, admin.RoleID, path, required, reason)
}

func permissionCodes(required string) []string {
	codes := make([]string, 0)
	seen := map[string]bool{}
	for _, item := range strings.Split(required, ",") {
		item = strings.TrimSpace(item)
		if item == "" || seen[item] {
			continue
		}
		seen[item] = true
		codes = append(codes, item)
	}
	return codes
}
