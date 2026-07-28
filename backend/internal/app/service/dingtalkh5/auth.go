package dingtalkh5

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"

	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
	"wecheckin-backend/backend/pkg/passwordutil"
	"wecheckin-backend/backend/pkg/randutil"
)

var normalizeUserIDRegexp = regexp.MustCompile(`[^a-z0-9_.-]+`)

func NormalizeUserID(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	return strings.Trim(normalizeUserIDRegexp.ReplaceAllString(value, ""), ".-_")
}

func LoginContext(ctx context.Context, name, password, addIP, device string) (*LoginResponse, error) {
	if err := EnsureSeedContext(ctx); err != nil {
		return nil, err
	}
	name = strings.TrimSpace(name)
	password = strings.TrimSpace(password)
	if name == "" || password == "" {
		return nil, fmt.Errorf("请填写账号和密码")
	}

	db, cancel := database.WithContext(ctx)
	defer cancel()
	var user model.DingTalkH5PerfUser
	if err := db.Where("`user_mini_openid` = ? OR `user_name` = ?", NormalizeUserID(name), name).First(&user).Error; err != nil {
		return nil, fmt.Errorf("账号或密码错误")
	}
	hydratePerfUser(&user)
	if user.Status != 1 {
		return nil, fmt.Errorf("账号已停用")
	}
	if !passwordutil.Verify(user.Password, password) {
		return nil, fmt.Errorf("账号或密码错误")
	}
	if passwordutil.NeedsRehash(user.Password) {
		if hash, err := passwordutil.Hash(password); err == nil {
			db.Model(&model.User{}).Where("`user_mini_openid` = ?", user.Account).Update("user_password", hash)
			user.Password = hash
		}
	}

	token := randutil.HexString(32)
	now := database.Now()
	session := model.DingTalkH5PerfSession{
		Token:       token,
		UserAccount: user.Account,
		ExpiresAt:   now + int64((7*24*time.Hour)/time.Millisecond),
		AddIP:       addIP,
		Device:      device,
		AddTime:     now,
	}
	if err := db.Create(&session).Error; err != nil {
		return nil, err
	}
	return &LoginResponse{Token: token, UserInfo: userDTO(user)}, nil
}

func AuthenticateContext(ctx context.Context, token string) (*model.DingTalkH5PerfUser, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, fmt.Errorf("未登录")
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var session model.DingTalkH5PerfSession
	now := database.Now()
	if err := db.Where("token = ? AND expires_at > ?", token, now).First(&session).Error; err != nil {
		return nil, fmt.Errorf("登录已过期")
	}
	var user model.DingTalkH5PerfUser
	if err := db.Where("`user_mini_openid` = ? AND `user_status` = 1", session.UserAccount).First(&user).Error; err != nil {
		return nil, fmt.Errorf("登录账号异常")
	}
	hydratePerfUser(&user)
	return &user, nil
}

func LogoutContext(ctx context.Context, token string) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	if db == nil || strings.TrimSpace(token) == "" {
		return nil
	}
	return db.Where("token = ?", strings.TrimSpace(token)).Delete(&model.DingTalkH5PerfSession{}).Error
}

func ChangePasswordContext(ctx context.Context, current *model.DingTalkH5PerfUser, currentPassword, newPassword, confirmPassword string) error {
	currentPassword = strings.TrimSpace(currentPassword)
	newPassword = strings.TrimSpace(newPassword)
	confirmPassword = strings.TrimSpace(confirmPassword)
	if current == nil {
		return fmt.Errorf("未登录")
	}
	if !passwordutil.Verify(current.Password, currentPassword) {
		return fmt.Errorf("当前密码不正确")
	}
	if len(newPassword) < 6 {
		return fmt.Errorf("新密码至少 6 位")
	}
	if newPassword != confirmPassword {
		return fmt.Errorf("两次输入的新密码不一致")
	}
	hash, err := passwordutil.Hash(newPassword)
	if err != nil {
		return err
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return db.Model(&model.User{}).Where("`user_mini_openid` = ?", current.Account).Update("user_password", hash).Error
}

func currentUserByAccount(ctx context.Context, account string) (*model.DingTalkH5PerfUser, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return loadPerfUserByAccountDB(db, account)
}

func isNotFound(err error) bool {
	return err == gorm.ErrRecordNotFound
}

func jsonUnmarshal(raw string, target interface{}) error {
	if strings.TrimSpace(raw) == "" {
		raw = "{}"
	}
	return json.Unmarshal([]byte(raw), target)
}
