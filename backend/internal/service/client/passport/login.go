package passport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/support/auditlog"
	"wecheckin/backend/internal/support/dept"
	"wecheckin/backend/internal/support/media"
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
	if db == nil {
		return nil, loginStorageError(ctx, "system_error")
	}

	var user model.User
	err := db.Where("`user_mini_openid` = ?", userID).First(&user).Error
	if err != nil {
		auditLoginLookupFailure(ctx, err)
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, loginStorageError(ctx, "system_error")
		}
		return &LoginResponse{Token: nil}, nil
	}
	db.Model(&user).Update("user_login_time", database.Now())
	db.Model(&user).UpdateColumn("user_login_cnt", user.LoginCnt+1)
	setUserRole(&user)
	fillUserRoleIDsContext(ctx, db, &user)
	token := randutil.HexString(32)
	if err := storeUserTokenContext(ctx, &user, token, addIP, device); err != nil {
		return nil, err
	}
	deptID := dept.UserDeptIDContext(ctx, user.ID)
	auditlog.SetActor(ctx, user.ID, user.Name, "user")
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
	if db == nil {
		return nil, loginStorageError(ctx, "system_error")
	}

	var user model.User
	err := db.Where("`user_name` = ? OR `user_mobile` = ?", name, name).First(&user).Error
	if err != nil {
		auditLoginLookupFailure(ctx, err)
		return nil, fmt.Errorf("账号或密码错误")
	}
	if !passwordutil.Verify(user.Password, password) {
		auditlog.SetReason(ctx, "invalid_credentials", "账号或密码错误")
		return nil, fmt.Errorf("账号或密码错误")
	}
	if passwordutil.NeedsRehash(user.Password) {
		newHash, err := passwordutil.Hash(password)
		if err != nil {
			auditlog.SetReason(ctx, "system_error", "登录服务异常")
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
	if err := storeUserTokenContext(ctx, &user, token, addIP, device); err != nil {
		return nil, err
	}
	deptID := dept.UserDeptIDContext(ctx, user.ID)
	var deptName string
	if deptID > 0 {
		var d model.Department
		db.First(&d, deptID)
		deptName = d.Name
	}
	auditlog.SetActor(ctx, user.ID, user.Name, "user")
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

func auditLoginLookupFailure(ctx context.Context, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		auditlog.SetReason(ctx, "invalid_credentials", "账号或密码错误")
	} else {
		auditlog.SetReason(ctx, "system_error", "登录服务异常")
	}
}

func storeUserToken(user *model.User, token, addIP, device string) error {
	return storeUserTokenContext(context.Background(), user, token, addIP, device)
}

func storeUserTokenContext(ctx context.Context, user *model.User, token, addIP, device string) (resultErr error) {
	if user == nil || user.ID == 0 || token == "" {
		return loginStorageError(ctx, "system_error")
	}
	if rd.RDB == nil {
		return loginStorageError(ctx, "session_store_failed")
	}
	expire, prefix := tokenutil.GetTokenConfigContext(ctx, "user")
	if prefix == "" {
		prefix = "user_token:"
	}
	keyAuth := prefix + "a:" + token
	idStr := strconv.Itoa(int(user.ID))
	keySet := prefix + "s:" + idStr
	redisCtx, cancel := rd.OperationContext(ctx)
	defer cancel()

	if tokenutil.IsUserSingleLoginContext(ctx) {
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
	jsonBytes, err := json.Marshal(info)
	if err != nil {
		return loginStorageError(ctx, "session_store_failed")
	}
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
	return nil
}

func loginStorageError(ctx context.Context, reasonCode string) error {
	auditlog.SetReason(ctx, reasonCode, "登录会话保存失败")
	return fmt.Errorf("登录失败，请稍后重试")
}
