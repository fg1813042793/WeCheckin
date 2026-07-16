package adminauth

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"wecheckin-backend/backend/internal/app/support/media"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
	"wecheckin-backend/backend/pkg/passwordutil"
	"wecheckin-backend/backend/pkg/randutil"
	rd "wecheckin-backend/backend/pkg/redis"
	"wecheckin-backend/backend/pkg/tokenutil"
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
	db.Create(&model.Log{
		Type:      logType,
		Content:   content,
		AdminID:   adminID,
		AdminName: adminName,
		AdminDesc: adminDesc,
		AddTime:   database.Now(),
		AddIP:     addIP,
	})
}

type LoginResponse struct {
	Token     string `json:"token"`
	Name      string `json:"name"`
	Pic       string `json:"pic"`
	ID        uint   `json:"id"`
	Type      int    `json:"type"`
	RoleID    uint   `json:"roleId"`
	DataScope int    `json:"dataScope"`
	LoginCnt  int    `json:"loginCnt"`
}

func Login(name, password, addIP, device string) (*LoginResponse, error) {
	return LoginContext(context.Background(), name, password, addIP, device)
}

func LoginContext(ctx context.Context, name, password, addIP, device string) (*LoginResponse, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()

	var admin model.Admin
	err := db.Where("`admin_name` = ?", name).First(&admin).Error
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
		db.Model(&admin).Update("admin_password", newHash)
		admin.Password = newHash
	}
	if admin.Status != 1 {
		return nil, fmt.Errorf("账号已禁用")
	}
	token := genRandomString(32)
	admin.LoginCnt++
	admin.LoginTime = database.Now()
	db.Model(&admin).Updates(map[string]interface{}{
		"admin_login_cnt":  admin.LoginCnt,
		"admin_login_time": admin.LoginTime,
	})
	roleName := ""
	if admin.RoleID > 0 {
		var role model.Role
		if err := db.First(&role, admin.RoleID).Error; err == nil {
			roleName = role.Name
		}
	}
	storeAdminTokenContext(ctx, &admin, token, addIP, device, roleName)
	InsertLogContext(ctx, 1, "管理员登录", strconv.Itoa(int(admin.ID)), admin.Name, admin.Desc, addIP)
	var dataScope int
	if admin.RoleID > 0 {
		var role model.Role
		if err := db.First(&role, admin.RoleID).Error; err == nil {
			dataScope = role.DataScope
		}
	}
	return &LoginResponse{
		Token:     token,
		Name:      admin.Name,
		Pic:       media.FullURLWithStaticDomain(admin.Pic),
		ID:        admin.ID,
		Type:      admin.Type,
		RoleID:    admin.RoleID,
		DataScope: dataScope,
		LoginCnt:  admin.LoginCnt,
	}, nil
}

func storeAdminToken(admin *model.Admin, token, addIP, device, roleName string) {
	storeAdminTokenContext(context.Background(), admin, token, addIP, device, roleName)
}

func storeAdminTokenContext(ctx context.Context, admin *model.Admin, token, addIP, device, roleName string) {
	expire, prefix := tokenutil.GetTokenConfig("admin")
	now := database.Now()
	db, cancel := database.WithContext(ctx)
	defer cancel()
	db.Model(&model.Admin{}).Where("`id` = ?", admin.ID).Updates(map[string]interface{}{
		"admin_token":      token,
		"admin_token_time": now,
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
