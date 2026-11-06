package adminauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"gorm.io/gorm"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/service/accountsecurity"
	"wecheckin/backend/internal/support/adminaccess"
	"wecheckin/backend/internal/support/auditlog"
	"wecheckin/backend/internal/support/media"
	permissionsupport "wecheckin/backend/internal/support/permission"
	"wecheckin/backend/pkg/database"
	"wecheckin/backend/pkg/passwordutil"
	"wecheckin/backend/pkg/randutil"
	rd "wecheckin/backend/pkg/redis"
	"wecheckin/backend/pkg/tokenutil"
)

func genRandomString(length int) string {
	return randutil.HexString(length)
}

func InsertLog(logType int, content, adminID, adminName, adminDesc, addIP string) {
	InsertLogContext(context.Background(), logType, content, adminID, adminName, adminDesc, addIP)
}

func InsertLogContext(ctx context.Context, logType int, content, adminID, adminName, adminDesc, addIP string) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	adminIDUint, _ := strconv.ParseUint(adminID, 10, 64)
	now := database.Now()
	db.Create(&model.Log{
		Type:      logType,
		Content:   content,
		AdminID:   uint(adminIDUint),
		UpdateBy:  uint(adminIDUint),
		AdminName: adminName,
		AdminDesc: adminDesc,
		AddTime:   now,
		EditTime:  now,
		AddIP:     addIP,
	})
}

type LoginResponse struct {
	Token                  string   `json:"token"`
	Name                   string   `json:"name"`
	Pic                    string   `json:"pic"`
	ID                     uint     `json:"id"`
	Type                   int      `json:"type"`
	RoleID                 uint     `json:"roleId"`
	RoleName               string   `json:"roleName"`
	RoleIDs                []uint   `json:"roleIds"`
	RoleNames              []string `json:"roleNames"`
	DataScope              int      `json:"dataScope"`
	LoginCnt               int      `json:"loginCnt"`
	PasswordChangeRequired bool     `json:"passwordChangeRequired"`
	PasswordExpired        bool     `json:"passwordExpired"`
	PasswordExpiresAt      int64    `json:"passwordExpiresAt"`
	PasswordExpiryWarning  bool     `json:"passwordExpiryWarning"`
}

func Login(name, password, addIP, device string) (*LoginResponse, error) {
	return LoginContext(context.Background(), name, password, addIP, device)
}

func LoginContext(ctx context.Context, name, password, addIP, device string) (*LoginResponse, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if db == nil {
		return nil, loginStorageError(ctx, "system_error")
	}

	var admin model.Admin
	err := db.Where("`user_mobile` = ? OR `user_name` = ? OR `user_account` = ?", name, name, name).First(&admin).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			auditlog.SetReason(ctx, "invalid_credentials", "账号或密码错误")
		} else {
			auditlog.SetReason(ctx, "system_error", "登录服务异常")
		}
		return nil, fmt.Errorf("账号或密码错误")
	}
	if !passwordutil.Verify(admin.Password, password) {
		auditlog.SetReason(ctx, "invalid_credentials", "账号或密码错误")
		return nil, fmt.Errorf("账号或密码错误")
	}
	if passwordutil.NeedsRehash(admin.Password) {
		newHash, err := passwordutil.Hash(password)
		if err != nil {
			auditlog.SetReason(ctx, "system_error", "登录服务异常")
			return nil, err
		}
		changedAt := database.Now()
		db.Model(&admin).Updates(map[string]interface{}{"user_password": newHash, "user_password_changed_at": changedAt})
		admin.Password = newHash
		admin.PasswordChangedAt = changedAt
	}
	if admin.Status != 1 {
		auditlog.SetReason(ctx, "account_disabled", "账号已禁用")
		return nil, fmt.Errorf("账号已禁用")
	}
	role, roleIDs, roleNames, err := adminLoginAccess(ctx, db, &admin)
	if err != nil {
		return nil, err
	}
	admin.RoleIDs = roleIDs
	admin.RoleNames = roleNames
	passwordState, err := accountsecurity.EvaluatePasswordStateContext(ctx, admin.PasswordChangedAt, admin.PasswordMustChange == 1)
	if err != nil {
		return nil, err
	}
	token := genRandomString(32)
	admin.LoginCnt++
	admin.LoginTime = database.Now()
	db.Model(&admin).Updates(map[string]interface{}{
		"user_login_cnt":  admin.LoginCnt,
		"user_login_time": admin.LoginTime,
	})
	if err := storeAdminTokenContext(ctx, &admin, token, addIP, device, role.Name, roleIDs, roleNames); err != nil {
		return nil, err
	}
	auditlog.SetActor(ctx, admin.ID, admin.Name, "admin")
	return &LoginResponse{
		Token:                  token,
		Name:                   admin.Name,
		Pic:                    media.FullURLWithStaticDomain(admin.Pic),
		ID:                     admin.ID,
		Type:                   admin.Type,
		RoleID:                 admin.RoleID,
		RoleName:               role.Name,
		RoleIDs:                roleIDs,
		RoleNames:              roleNames,
		DataScope:              role.DataScope,
		LoginCnt:               admin.LoginCnt,
		PasswordChangeRequired: passwordState.PasswordChangeRequired,
		PasswordExpired:        passwordState.PasswordExpired,
		PasswordExpiresAt:      passwordState.PasswordExpiresAt,
		PasswordExpiryWarning:  passwordState.PasswordExpiryWarning,
	}, nil
}

func adminLoginRole(ctx context.Context, db *gorm.DB, admin *model.Admin) (model.Role, error) {
	return adminaccess.UserAllowsAdminAccessContext(ctx, db, admin.ID, admin.RoleID)
}

func adminLoginAccess(ctx context.Context, db *gorm.DB, admin *model.Admin) (model.Role, []uint, []string, error) {
	roleIDs, err := permissionsupport.ActiveRoleIDsForUserContext(ctx, db, admin.ID, admin.RoleID)
	if err != nil {
		auditlog.SetReason(ctx, "system_error", "后台权限查询失败")
		return model.Role{}, nil, nil, err
	}
	role, err := adminaccess.UserAllowsAdminAccessWithRoleIDsContext(ctx, db, admin.ID, roleIDs)
	if err != nil {
		// The access package exposes these fixed denials as plain errors.
		switch err.Error() {
		case "当前用户未分配后台角色", "当前角色已停用或不存在", "当前用户未获得后台入口权限":
			auditlog.SetReason(ctx, "admin_access_denied", "未获得后台访问权限")
		default:
			auditlog.SetReason(ctx, "system_error", "后台权限查询失败")
		}
		return model.Role{}, nil, nil, err
	}
	roleNames, err := roleNamesByIDsContext(ctx, db, roleIDs)
	if err != nil {
		auditlog.SetReason(ctx, "system_error", "后台权限查询失败")
		return model.Role{}, nil, nil, err
	}
	return role, roleIDs, roleNames, nil
}

func roleNamesByIDsContext(ctx context.Context, db *gorm.DB, roleIDs []uint) ([]string, error) {
	if len(roleIDs) == 0 {
		return nil, nil
	}
	var rows []model.Role
	if err := db.WithContext(ctx).Select("id", "role_name").Where("`id` IN ?", roleIDs).Find(&rows).Error; err != nil {
		return nil, err
	}
	namesByID := make(map[uint]string, len(rows))
	for _, row := range rows {
		namesByID[row.ID] = row.Name
	}
	names := make([]string, 0, len(roleIDs))
	for _, roleID := range roleIDs {
		if name := namesByID[roleID]; name != "" {
			names = append(names, name)
		}
	}
	return names, nil
}

func storeAdminToken(admin *model.Admin, token, addIP, device, roleName string) error {
	if admin == nil {
		return loginStorageError(context.Background(), "system_error")
	}
	return storeAdminTokenContext(context.Background(), admin, token, addIP, device, roleName, []uint{admin.RoleID}, admin.RoleNames)
}

func storeAdminTokenContext(ctx context.Context, admin *model.Admin, token, addIP, device, roleName string, roleIDs []uint, roleNames []string) (resultErr error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if db == nil || admin == nil || admin.ID == 0 || token == "" {
		return loginStorageError(ctx, "system_error")
	}
	if rd.RDB == nil {
		return loginStorageError(ctx, "session_store_failed")
	}
	expire, prefix := tokenutil.GetTokenConfigContext(ctx, "admin")
	now := database.Now()
	keyAuth := prefix + "a:" + token
	idStr := strconv.Itoa(int(admin.ID))
	keySet := prefix + "s:" + idStr
	redisCtx, redisCancel := rd.OperationContext(ctx)
	defer redisCancel()

	if tokenutil.IsAdminSingleLoginContext(ctx) {
		oldTokens, err := rd.RDB.SMembers(redisCtx, keySet).Result()
		if err != nil {
			return loginStorageError(ctx, "session_store_failed")
		}
		if len(oldTokens) > 0 {
			for _, t := range oldTokens {
				if t != token {
					if err := rd.RDB.Del(redisCtx, prefix+"a:"+t).Err(); err != nil {
						return loginStorageError(ctx, "session_store_failed")
					}
				}
			}
			if err := rd.RDB.Del(redisCtx, keySet).Err(); err != nil {
				return loginStorageError(ctx, "session_store_failed")
			}
		}
	}

	info := map[string]interface{}{
		"id":                 admin.ID,
		"name":               admin.Name,
		"type":               admin.Type,
		"roleId":             admin.RoleID,
		"roleName":           roleName,
		"roleIds":            roleIDs,
		"roleNames":          roleNames,
		"desc":               admin.Desc,
		"passwordChangedAt":  admin.PasswordChangedAt,
		"passwordMustChange": admin.PasswordMustChange == 1,
		"loginIp":            addIP,
		"loginTime":          now,
		"device":             device,
	}
	jsonBytes, err := json.Marshal(info)
	if err != nil {
		return loginStorageError(ctx, "session_store_failed")
	}
	// A partially created session must not survive a failed login. Cleanup is
	// best effort and only removes this attempt's token from a shared index.
	defer func() {
		if resultErr != nil {
			rd.RDB.Del(redisCtx, keyAuth)
			rd.RDB.SRem(redisCtx, keySet, token)
		}
	}()
	if err := rd.RDB.Set(redisCtx, keyAuth, string(jsonBytes), expire).Err(); err != nil {
		return loginStorageError(ctx, "session_store_failed")
	}
	if err := rd.RDB.SAdd(redisCtx, keySet, token).Err(); err != nil {
		return loginStorageError(ctx, "session_store_failed")
	}
	if ok, err := rd.RDB.Expire(redisCtx, keySet, expire*2).Result(); err != nil || !ok {
		return loginStorageError(ctx, "session_store_failed")
	}
	if err := db.Model(&model.Admin{}).Where("`id` = ?", admin.ID).Updates(map[string]interface{}{
		"user_admin_token":      token,
		"user_admin_token_time": now,
	}).Error; err != nil {
		return loginStorageError(ctx, "session_store_failed")
	}
	return nil
}

func loginStorageError(ctx context.Context, reasonCode string) error {
	auditlog.SetReason(ctx, reasonCode, "登录会话保存失败")
	return fmt.Errorf("登录失败，请稍后重试")
}
