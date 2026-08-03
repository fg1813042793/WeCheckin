package dingtalkh5

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"gorm.io/gorm"

	onlineservice "wecheckin/backend/internal/app/service/online"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/database"
	"wecheckin/backend/pkg/passwordutil"
	"wecheckin/backend/pkg/randutil"
)

var normalizeUserIDRegexp = regexp.MustCompile(`[^a-z0-9_.-]+`)

const maxSessionDeviceLength = 255

func NormalizeUserID(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	return strings.Trim(normalizeUserIDRegexp.ReplaceAllString(value, ""), ".-_")
}

func sessionDevice(device string) string {
	device = strings.TrimSpace(device)
	runes := []rune(device)
	if len(runes) <= maxSessionDeviceLength {
		return device
	}
	return string(runes[:maxSessionDeviceLength])
}

func LoginContext(ctx context.Context, name, password, addIP, device string) (*LoginResponse, error) {
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
	if err := hydratePerfUserWithUserDeptDB(db, &user); err != nil {
		return nil, err
	}
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

	return issueDingTalkH5LoginResponseContext(ctx, db, &user, addIP, device)
}

func issueDingTalkH5LoginResponseContext(ctx context.Context, db *gorm.DB, user *model.DingTalkH5PerfUser, addIP, device string) (*LoginResponse, error) {
	if user == nil {
		return nil, fmt.Errorf("登录信息异常")
	}
	token := randutil.HexString(32)
	safeDevice := sessionDevice(device)
	if err := onlineservice.StoreDingTalkH5SessionContext(ctx, onlineUserFromDingTalkH5User(user), token, addIP, safeDevice); err != nil {
		return nil, err
	}
	bootstrap := bootstrapForUserDB(ctx, db, user)
	return &LoginResponse{
		Token:                 token,
		UserInfo:              bootstrap.User,
		Menus:                 bootstrap.Menus,
		AppConfig:             bootstrap.AppConfig,
		AppTitle:              bootstrap.AppTitle,
		AppName:               bootstrap.AppName,
		LogoText:              bootstrap.LogoText,
		LogoURL:               bootstrap.LogoURL,
		ButtonPermissionKeys:  bootstrap.ButtonPermissionKeys,
		ButtonPermissionReady: bootstrap.ButtonPermissionReady,
		APIPermissionKeys:     bootstrap.APIPermissionKeys,
		APIPermissionReady:    bootstrap.APIPermissionReady,
		PermissionVersion:     bootstrap.PermissionVersion,
	}, nil
}

func AuthenticateContext(ctx context.Context, token string) (*model.DingTalkH5PerfUser, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, fmt.Errorf("未登录")
	}
	account, err := onlineservice.DingTalkH5SessionAccountContext(ctx, token)
	if err != nil {
		return nil, err
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var user model.DingTalkH5PerfUser
	if err := db.Where("`user_mini_openid` = ? AND `user_status` = 1", account).First(&user).Error; err != nil {
		return nil, fmt.Errorf("登录账号异常")
	}
	if err := hydratePerfUserWithUserDeptDB(db, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func LogoutContext(ctx context.Context, current *model.DingTalkH5PerfUser, token string) error {
	if strings.TrimSpace(token) == "" {
		return nil
	}
	token = strings.TrimSpace(token)

	userID := uint(0)
	if current != nil {
		userID = current.ID
	}
	if userID == 0 {
		account, err := onlineservice.DingTalkH5SessionAccountContext(ctx, token)
		if err == nil && strings.TrimSpace(account) != "" {
			db, cancel := database.WithContext(ctx)
			defer cancel()
			if db != nil {
				user, err := loadPerfUserByAccountDB(db, account)
				if err == nil && user != nil {
					userID = user.ID
				}
			}
		}
	}

	return onlineservice.RemoveDingTalkH5SessionContext(ctx, userID, token)
}

func onlineUserFromDingTalkH5User(user *model.DingTalkH5PerfUser) *model.User {
	if user == nil {
		return nil
	}
	return &model.User{
		ID:         user.ID,
		MiniOpenID: user.Account,
		Status:     user.Status,
		Name:       user.Name,
		Pic:        user.Pic,
		Password:   user.Password,
		RoleID:     user.RoleID,
		RoleIDs:    user.RoleIDs,
	}
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

func UpdateAccountProfileContext(ctx context.Context, current *model.DingTalkH5PerfUser, token string, payload AccountProfilePayload) (*AccountProfileResponse, error) {
	if current == nil {
		return nil, fmt.Errorf("未登录")
	}
	nextAccount := NormalizeUserID(firstNonEmpty(payload.Account, current.Account))
	if nextAccount == "" {
		return nil, fmt.Errorf("请填写账号")
	}
	avatar, err := sanitizeAvatarURL(payload.Avatar)
	if err != nil {
		return nil, err
	}
	accountChanged := nextAccount != NormalizeUserID(current.Account)
	if accountChanged {
		if !passwordutil.Verify(current.Password, strings.TrimSpace(payload.CurrentPassword)) {
			return nil, fmt.Errorf("修改账号需填写正确的当前密码")
		}
	}

	db, cancel := database.WithContext(ctx)
	defer cancel()
	var updated model.DingTalkH5PerfUser
	if err := db.Transaction(func(tx *gorm.DB) error {
		existing, err := loadPerfUserByIDDB(tx, current.ID)
		if err != nil {
			return fmt.Errorf("登录账号异常")
		}
		if nextAccount != NormalizeUserID(existing.Account) {
			if err := validateAccountAvailable(tx, nextAccount, existing.Account); err != nil {
				return err
			}
		}
		now := database.Now()
		updates := map[string]interface{}{
			"user_mini_openid": nextAccount,
			"user_pic":         avatar,
			"user_edit_time":   now,
		}
		if err := tx.Model(&model.User{}).Where("`id` = ?", existing.ID).Updates(updates).Error; err != nil {
			return err
		}
		if accountChanged {
			if err := syncPerfAccountReferences(tx, NormalizeUserID(existing.Account), nextAccount); err != nil {
				return err
			}
		}
		loaded, err := loadPerfUserByIDDB(tx, existing.ID)
		if err != nil {
			return err
		}
		updated = *loaded
		return nil
	}); err != nil {
		return nil, err
	}

	if strings.TrimSpace(token) != "" {
		if err := onlineservice.UpdateDingTalkH5SessionUserContext(ctx, onlineUserFromDingTalkH5User(&updated), strings.TrimSpace(token)); err != nil {
			return nil, err
		}
	}
	dto := userDTO(updated)
	return &AccountProfileResponse{User: dto}, nil
}

func validateAccountAvailable(db *gorm.DB, nextAccount, existingAccount string) error {
	nextAccount = NormalizeUserID(nextAccount)
	existingAccount = NormalizeUserID(existingAccount)
	if nextAccount == "" {
		return fmt.Errorf("请填写账号")
	}
	var duplicate model.User
	err := db.Where("`user_mini_openid` = ? AND `user_mini_openid` <> ?", nextAccount, existingAccount).First(&duplicate).Error
	if err == nil {
		return fmt.Errorf("账号已存在")
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	return nil
}

func sanitizeAvatarURL(value string) (string, error) {
	value = strings.TrimSpace(value)
	if len([]rune(value)) > 500 {
		return "", fmt.Errorf("头像地址不能超过 500 个字符")
	}
	if value == "" {
		return "", nil
	}
	lower := strings.ToLower(value)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") || strings.HasPrefix(value, "/") {
		return value, nil
	}
	return "", fmt.Errorf("头像地址需使用 http(s) 或站内相对路径")
}

func syncPerfAccountReferences(tx *gorm.DB, oldAccount, nextAccount string) error {
	oldAccount = NormalizeUserID(oldAccount)
	nextAccount = NormalizeUserID(nextAccount)
	if oldAccount == "" || nextAccount == "" || oldAccount == nextAccount {
		return nil
	}
	now := database.Now()
	reviewUpdates := []struct {
		column string
	}{
		{"employee_account"},
		{"manager_account"},
		{"hrbp_account"},
		{"hrbp_reviewer_account"},
	}
	for _, item := range reviewUpdates {
		if err := notDeletedReviewQuery(tx.Model(&model.DingTalkH5PerfReview{})).
			Where(item.column+" = ?", oldAccount).
			Updates(map[string]interface{}{item.column: nextAccount, "edit_time": now}).Error; err != nil {
			return err
		}
	}
	if err := tx.Model(&model.DingTalkH5PerfHistory{}).
		Where("`by_account` = ?", oldAccount).
		Update("by_account", nextAccount).Error; err != nil {
		return err
	}
	return syncPerfUserMetadataAccountReferences(tx, oldAccount, nextAccount)
}

func syncPerfUserMetadataAccountReferences(tx *gorm.DB, oldAccount, nextAccount string) error {
	var users []model.DingTalkH5PerfUser
	if err := tx.Where("`user_obj` LIKE ?", "%"+oldAccount+"%").Find(&users).Error; err != nil {
		return err
	}
	for _, user := range users {
		hydratePerfUser(&user)
		changed := false
		if NormalizeUserID(user.ManagerAccount) == oldAccount {
			user.ManagerAccount = nextAccount
			changed = true
		}
		if NormalizeUserID(user.HRBPAccount) == oldAccount {
			user.HRBPAccount = nextAccount
			changed = true
		}
		if !changed {
			continue
		}
		if err := tx.Model(&model.User{}).
			Where("`id` = ?", user.ID).
			Updates(map[string]interface{}{
				"user_obj":       encodePerfUserObj(user.Obj, user),
				"user_edit_time": database.Now(),
			}).Error; err != nil {
			return err
		}
	}
	return nil
}

func currentUserByAccount(ctx context.Context, account string) (*model.DingTalkH5PerfUser, error) {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return loadPerfUserByAccountDB(db, account)
}

func jsonUnmarshal(raw string, target interface{}) error {
	if strings.TrimSpace(raw) == "" {
		raw = "{}"
	}
	return json.Unmarshal([]byte(raw), target)
}
