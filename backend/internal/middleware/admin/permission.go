package admin

import (
	"context"
	"errors"
	"log"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"gorm.io/gorm"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/support/adminaccess"
	permissionsupport "wecheckin/backend/internal/support/permission"
	"wecheckin/backend/pkg/database"
	"wecheckin/backend/pkg/logger"
	"wecheckin/backend/pkg/response"
)

func AdminPerm() app.HandlerFunc {
	return adminPermWithLookup(adminHasAPIPermission)
}

type adminAPIPermissionLookup func(context.Context, *gorm.DB, *model.Admin, []uint, string) (bool, error)

func adminPermWithLookup(lookup adminAPIPermissionLookup) app.HandlerFunc {
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
			message := "未分配角色，无权限"
			if required, ok := adminRoutePermission(string(c.Method()), string(c.Path())); ok && required != "" {
				message = missingAdminPermissionMessage(required) + "；当前账号未分配角色"
			}
			c.JSON(consts.StatusOK, utils.H{
				"code": 1,
				"msg":  message,
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
				"msg":  "接口未配置权限规则，请联系管理员",
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
		var lookupError error
		for _, code := range requiredCodes {
			apiKey := permissionsupport.AdminAPIPermissionKey(code)
			allowed, err := lookup(ctx, db, admin, roleIDs, apiKey)
			if err != nil {
				lookupError = err
				continue
			}
			if allowed {
				c.Next(ctx)
				return
			}
		}
		if lookupError != nil {
			auditAdminPermissionDenied(admin, path, required, "permission_lookup_failed")
			response.FailInternal(ctx, c, "admin.permission", "权限校验失败，请稍后重试", lookupError)
			c.Abort()
			return
		}

		c.JSON(consts.StatusOK, utils.H{
			"code": 1,
			"msg":  missingAdminPermissionMessage(required),
		})
		auditAdminPermissionDenied(admin, path, required, "permission_missing")
		c.Abort()
	}
}

func adminHasAPIPermission(ctx context.Context, db *gorm.DB, admin *model.Admin, roleIDs []uint, apiKey string) (bool, error) {
	if db == nil {
		return false, errors.New("admin permission database is unavailable")
	}
	effect, found, err := permissionsupport.SubjectPermissionEffectContext(ctx, db, permissionsupport.SubjectUser, admin.ID, apiKey)
	if err != nil {
		return false, err
	}
	if found && effect == permissionsupport.EffectDeny {
		return false, nil
	}
	return permissionsupport.SubjectHasPermissionWithRoleIDsContext(ctx, db, admin.ID, roleIDs, apiKey)
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
