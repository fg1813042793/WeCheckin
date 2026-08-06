package adminauth

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"gorm.io/gorm"
	"wecheckin/backend/internal/support/adminaccess"
	"wecheckin/backend/internal/support/media"
	permissionsupport "wecheckin/backend/internal/support/permission"
	"wecheckin/backend/internal/model"
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
	Token     string   `json:"token"`
	Name      string   `json:"name"`
	Pic       string   `json:"pic"`
	ID        uint     `json:"id"`
	Type      int      `json:"type"`
	RoleID    uint     `json:"roleId"`
	RoleName  string   `json:"roleName"`
	RoleIDs   []uint   `json:"roleIds"`
	RoleNames []string `json:"roleNames"`
	DataScope int      `json:"dataScope"`
	LoginCnt  int      `json:"loginCnt"`
}

func Login(name, password, addIP, device string) (*LoginResponse, error) {
	return LoginContext(context.Background(), name, password, addIP, device)
}

func LoginContext(ctx context.Context, name, password, addIP, device string) (*LoginResponse, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()

	var admin model.Admin
	err := db.Where("`user_mobile` = ? OR `user_name` = ? OR `user_account` = ?", name, name, name).First(&admin).Error
	if err != nil {
		return nil, fmt.Errorf("账号或密码错误")
	}
	if !passwordutil.Verify(admin.Password, password) {
		return nil, fmt.Errorf("账号或密码错误")
	}
	if passwordutil.NeedsRehash(admin.Password) {
		newHash, err := passwordutil.Hash(password)
		if err != nil {
			return nil, err
		}
		db.Model(&admin).Update("user_password", newHash)
		admin.Password = newHash
	}
	if admin.Status != 1 {
		return nil, fmt.Errorf("账号已禁用")
	}
	role, roleIDs, roleNames, err := adminLoginAccess(ctx, db, &admin)
	if err != nil {
		return nil, err
	}
	admin.RoleIDs = roleIDs
	admin.RoleNames = roleNames
	token := genRandomString(32)
	admin.LoginCnt++
	admin.LoginTime = database.Now()
	db.Model(&admin).Updates(map[string]interface{}{
		"user_login_cnt":  admin.LoginCnt,
		"user_login_time": admin.LoginTime,
	})
	storeAdminTokenContext(ctx, &admin, token, addIP, device, role.Name, roleIDs, roleNames)
	InsertLogContext(ctx, 1, "管理员登录", strconv.Itoa(int(admin.ID)), admin.Name, admin.Desc, addIP)
	return &LoginResponse{
		Token:     token,
		Name:      admin.Name,
		Pic:       media.FullURLWithStaticDomain(admin.Pic),
		ID:        admin.ID,
		Type:      admin.Type,
		RoleID:    admin.RoleID,
		RoleName:  role.Name,
		RoleIDs:   roleIDs,
		RoleNames: roleNames,
		DataScope: role.DataScope,
		LoginCnt:  admin.LoginCnt,
	}, nil
}

func adminLoginRole(ctx context.Context, db *gorm.DB, admin *model.Admin) (model.Role, error) {
	return adminaccess.UserAllowsAdminAccessContext(ctx, db, admin.ID, admin.RoleID)
}

func adminLoginAccess(ctx context.Context, db *gorm.DB, admin *model.Admin) (model.Role, []uint, []string, error) {
	roleIDs, err := permissionsupport.ActiveRoleIDsForUserContext(ctx, db, admin.ID, admin.RoleID)
	if err != nil {
		return model.Role{}, nil, nil, err
	}
	role, err := adminaccess.UserAllowsAdminAccessWithRoleIDsContext(ctx, db, admin.ID, roleIDs)
	if err != nil {
		return model.Role{}, nil, nil, err
	}
	roleNames, err := roleNamesByIDsContext(ctx, db, roleIDs)
	if err != nil {
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

func storeAdminToken(admin *model.Admin, token, addIP, device, roleName string) {
	storeAdminTokenContext(context.Background(), admin, token, addIP, device, roleName, []uint{admin.RoleID}, admin.RoleNames)
}

func storeAdminTokenContext(ctx context.Context, admin *model.Admin, token, addIP, device, roleName string, roleIDs []uint, roleNames []string) {
	expire, prefix := tokenutil.GetTokenConfig("admin")
	now := database.Now()
	db, cancel := database.WithContext(ctx)
	defer cancel()
	db.Model(&model.Admin{}).Where("`id` = ?", admin.ID).Updates(map[string]interface{}{
		"user_admin_token":      token,
		"user_admin_token_time": now,
	})
	if rd.RDB == nil {
		return
	}
	keyAuth := prefix + "a:" + token
	idStr := strconv.Itoa(int(admin.ID))
	keySet := prefix + "s:" + idStr
	redisCtx, redisCancel := rd.OperationContext(ctx)
	defer redisCancel()

	if tokenutil.IsAdminSingleLogin() {
		if oldTokens, _ := rd.RDB.SMembers(redisCtx, keySet).Result(); len(oldTokens) > 0 {
			for _, t := range oldTokens {
				if t != token {
					rd.RDB.Del(redisCtx, prefix+"a:"+t)
				}
			}
			rd.RDB.Del(redisCtx, keySet)
		}
	}

	info := map[string]interface{}{
		"id":        admin.ID,
		"name":      admin.Name,
		"type":      admin.Type,
		"roleId":    admin.RoleID,
		"roleName":  roleName,
		"roleIds":   roleIDs,
		"roleNames": roleNames,
		"desc":      admin.Desc,
		"loginIp":   addIP,
		"loginTime": now,
		"device":    device,
	}
	jsonBytes, _ := json.Marshal(info)
	rd.RDB.Set(redisCtx, keyAuth, string(jsonBytes), expire)
	rd.RDB.SAdd(redisCtx, keySet, token)
	rd.RDB.Expire(redisCtx, keySet, expire*2)
}
