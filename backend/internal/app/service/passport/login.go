package passport

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"wecheckin/backend/internal/app/support/dept"
	"wecheckin/backend/internal/app/support/media"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/database"
	"wecheckin/backend/pkg/passwordutil"
	"wecheckin/backend/pkg/randutil"
	rd "wecheckin/backend/pkg/redis"
	"wecheckin/backend/pkg/tokenutil"
)

type LoginResponse struct {
	Token    interface{} `json:"token"`
	UserInfo *UserInfo   `json:"userInfo,omitempty"`
}

type UserInfo struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	Avatar     string `json:"avatar"`
	Desc       string `json:"desc"`
	MiniOpenID string `json:"miniOpenID"`
	Role       string `json:"role"`
	DeptID     uint   `json:"deptId"`
	DeptName   string `json:"deptName"`
}

func LoginUser(userID, addIP, device string) (*LoginResponse, error) {
	return LoginUserContext(context.Background(), userID, addIP, device)
}

func LoginUserContext(ctx context.Context, userID, addIP, device string) (*LoginResponse, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()

	var user model.User
	err := db.Where("`user_mini_openid` = ?", userID).First(&user).Error
	if err != nil {
		return &LoginResponse{Token: nil}, nil
	}
	db.Model(&user).Update("user_login_time", database.Now())
	db.Model(&user).UpdateColumn("user_login_cnt", user.LoginCnt+1)
	setUserRole(&user)
	fillUserRoleIDsContext(ctx, db, &user)
	token := randutil.HexString(32)
	storeUserTokenContext(ctx, &user, token, addIP, device)
	deptID := dept.UserDeptID(user.ID)
	return &LoginResponse{
		Token: token,
		UserInfo: &UserInfo{
			ID:         user.ID,
			Name:       user.Name,
			Avatar:     media.FullURLWithStaticDomain(user.Pic),
			Desc:       "点击完善个人信息",
			MiniOpenID: user.MiniOpenID,
			Role:       user.Role,
			DeptID:     deptID,
			DeptName:   "",
		},
	}, nil
}

func LoginByPwd(name, password, addIP, device string) (*LoginResponse, error) {
	return LoginByPwdContext(context.Background(), name, password, addIP, device)
}

func LoginByPwdContext(ctx context.Context, name, password, addIP, device string) (*LoginResponse, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()

	var user model.User
	err := db.Where("`user_name` = ? OR `user_mobile` = ?", name, name).First(&user).Error
	if err != nil {
		return nil, fmt.Errorf("账号或密码错误")
	}
	if !passwordutil.Verify(user.Password, password) {
		return nil, fmt.Errorf("账号或密码错误")
	}
	if passwordutil.NeedsRehash(user.Password) {
		newHash, err := passwordutil.Hash(password)
		if err != nil {
			return nil, err
		}
		db.Model(&user).Update("user_password", newHash)
		user.Password = newHash
	}
	db.Model(&user).Update("user_login_time", database.Now())
	db.Model(&user).UpdateColumn("user_login_cnt", user.LoginCnt+1)
	setUserRole(&user)
	fillUserRoleIDsContext(ctx, db, &user)
	token := randutil.HexString(32)
	storeUserTokenContext(ctx, &user, token, addIP, device)
	deptID := dept.UserDeptID(user.ID)
	var deptName string
	if deptID > 0 {
		var d model.Department
		db.First(&d, deptID)
		deptName = d.Name
	}
	return &LoginResponse{
		Token: token,
		UserInfo: &UserInfo{
			ID:         user.ID,
			Name:       user.Name,
			Avatar:     media.FullURLWithStaticDomain(user.Pic),
			Desc:       "点击完善个人信息",
			MiniOpenID: user.MiniOpenID,
			Role:       user.Role,
			DeptID:     deptID,
			DeptName:   deptName,
		},
	}, nil
}

func storeUserToken(user *model.User, token, addIP, device string) {
	storeUserTokenContext(context.Background(), user, token, addIP, device)
}

func storeUserTokenContext(ctx context.Context, user *model.User, token, addIP, device string) {
	expire, prefix := tokenutil.GetTokenConfig("user")
	if prefix == "" {
		prefix = "user_token:"
	}
	if rd.RDB == nil {
		return
	}
	keyAuth := prefix + "a:" + token
	idStr := strconv.Itoa(int(user.ID))
	keySet := prefix + "s:" + idStr
	redisCtx, cancel := rd.OperationContext(ctx)
	defer cancel()

	if tokenutil.IsUserSingleLogin() {
		if oldTokens, _ := rd.RDB.SMembers(redisCtx, keySet).Result(); len(oldTokens) > 0 {
			for _, t := range oldTokens {
				if t != token {
					rd.RDB.Del(redisCtx, prefix+"a:"+t)
				}
			}
			rd.RDB.Del(redisCtx, keySet)
		}
	}

	now := database.Now()
	info := map[string]interface{}{
		"id":         user.ID,
		"name":       user.Name,
		"mobile":     user.Mobile,
		"miniOpenID": user.MiniOpenID,
		"role":       user.Role,
		"roleId":     user.RoleID,
		"roleIds":    user.RoleIDs,
		"pic":        user.Pic,
		"loginIp":    addIP,
		"loginTime":  now,
		"device":     device,
	}
	jsonBytes, _ := json.Marshal(info)
	rd.RDB.Set(redisCtx, keyAuth, string(jsonBytes), expire)
	rd.RDB.SAdd(redisCtx, keySet, token)
	rd.RDB.Expire(redisCtx, keySet, expire*2)
}
