package service

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"

	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
	"wecheckin-backend/backend/pkg/passwordutil"
	rd "wecheckin-backend/backend/pkg/redis"
	"wecheckin-backend/backend/pkg/tokenutil"
)

// genRandomString 使用 crypto/rand 生成 length 个十六进制字符（length 必须为偶数）。
// 熵为 length*4 bit；长度 32 即 128 bit，远高于 UUID v4。
// 熵源失败时 panic（系统熵不可用 = 整个服务都不可信）。
func genRandomString(length int) string {
	if length <= 0 || length%2 != 0 {
		panic(fmt.Sprintf("genRandomString: length must be a positive even number, got %d", length))
	}
	b := make([]byte, length/2)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("genRandomString: crypto/rand failed: %v", err))
	}
	return hex.EncodeToString(b)
}

func InsertLog(logType int, content, adminID, adminName, adminDesc, addIP string) {
	database.DB.Create(&model.Log{
		Type:      logType,
		Content:   content,
		AdminID:   adminID,
		AdminName: adminName,
		AdminDesc: adminDesc,
		AddTime:   database.Now(),
		AddIP:     addIP,
	})
}

func AdminLogin(name, password, addIP, device string) (map[string]interface{}, error) {
	var admin model.Admin
	err := database.DB.Where("`admin_name` = ?", name).First(&admin).Error
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
		database.DB.Model(&admin).Update("admin_password", newHash)
		admin.Password = newHash
	}
	if admin.Status != 1 {
		return nil, fmt.Errorf("账号已禁用")
	}
	token := genRandomString(32)
	admin.LoginCnt++
	admin.LoginTime = database.Now()
	database.DB.Model(&admin).Updates(map[string]interface{}{
		"admin_login_cnt":  admin.LoginCnt,
		"admin_login_time": admin.LoginTime,
	})
	roleName := ""
	if admin.RoleID > 0 {
		var role model.Role
		if err := database.DB.First(&role, admin.RoleID).Error; err == nil {
			roleName = role.Name
		}
	}
	storeAdminToken(&admin, token, addIP, device, roleName)
	InsertLog(1, "管理员登录", strconv.Itoa(int(admin.ID)), admin.Name, admin.Desc, addIP)
	var dataScope int
	if admin.RoleID > 0 {
		var role model.Role
		if err := database.DB.First(&role, admin.RoleID).Error; err == nil {
			dataScope = role.DataScope
		}
	}
	result := map[string]interface{}{
		"token":     token,
		"name":      admin.Name,
		"pic":       GetFullURL(admin.Pic),
		"id":        admin.ID,
		"type":      admin.Type,
		"roleId":    admin.RoleID,
		"dataScope": dataScope,
		"loginCnt":  admin.LoginCnt,
	}
	return result, nil
}

func storeAdminToken(admin *model.Admin, token, addIP, device, roleName string) {
	expire, prefix := tokenutil.GetTokenConfig("admin")
	now := database.Now()
	database.DB.Model(&model.Admin{}).Where("`id` = ?", admin.ID).Updates(map[string]interface{}{
		"admin_token":      token,
		"admin_token_time": now,
	})
	if rd.RDB != nil {
		keyAuth := prefix + "a:" + token
		idStr := strconv.Itoa(int(admin.ID))
		keySet := prefix + "s:" + idStr

		// 单端登录模式：踢掉同 adminID 的所有旧 token
		if tokenutil.IsAdminSingleLogin() {
			if oldTokens, _ := rd.RDB.SMembers(rd.Ctx, keySet).Result(); len(oldTokens) > 0 {
				for _, t := range oldTokens {
					if t != token {
						rd.RDB.Del(rd.Ctx, prefix+"a:"+t)
					}
				}
				rd.RDB.Del(rd.Ctx, keySet)
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
		// Set TTL on s: 2x expire to keep it alive through long sessions
		// (a: slides on every request; s: only updates on token add/remove)
		rd.RDB.Set(rd.Ctx, keyAuth, string(jsonBytes), expire)
		rd.RDB.SAdd(rd.Ctx, keySet, token)
		rd.RDB.Expire(rd.Ctx, keySet, expire*2)
	}
}
