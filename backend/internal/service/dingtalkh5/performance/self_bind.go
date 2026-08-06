package performance

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/database"
	"wecheckin/backend/pkg/passwordutil"
	"wecheckin/backend/pkg/randutil"
	rd "wecheckin/backend/pkg/redis"
)

const (
	DingTalkH5BindRequiredCode = 10020

	dingTalkH5SelfBindEnabledKey   = "DINGTALK_H5_SELF_BIND_ENABLED"
	dingTalkH5SelfBindTicketTTLKey = "DINGTALK_H5_SELF_BIND_TICKET_TTL"
	dingTalkH5SelfBindTicketPrefix = "dingtalk_h5_bind_ticket:"
	dingTalkH5SelfBindDefaultTTL   = 10 * time.Minute
)

type DingTalkH5BindRequiredResponse struct {
	BindTicket           string                 `json:"bindTicket"`
	CorpID               string                 `json:"corpId"`
	DingTalkUserIDMasked string                 `json:"dingTalkUserIdMasked"`
	UnionIDMasked        string                 `json:"unionIdMasked,omitempty"`
	ExpiresIn            int64                  `json:"expiresIn"`
	AppConfig            DingTalkH5AppConfigDTO `json:"appConfig"`
	AppTitle             string                 `json:"appTitle"`
	AppName              string                 `json:"appName"`
	LogoText             string                 `json:"logoText"`
	LogoURL              string                 `json:"logoUrl"`
}

type DingTalkH5BindRequiredError struct {
	Response DingTalkH5BindRequiredResponse
}

func (e *DingTalkH5BindRequiredError) Error() string {
	return "钉钉账号未绑定系统用户"
}

func DingTalkH5BindRequiredData(err error) (DingTalkH5BindRequiredResponse, bool) {
	var bindErr *DingTalkH5BindRequiredError
	if errors.As(err, &bindErr) && bindErr != nil {
		return bindErr.Response, true
	}
	return DingTalkH5BindRequiredResponse{}, false
}

type dingTalkH5SelfBindTicket struct {
	CorpID         string `json:"corpId"`
	DingTalkUserID string `json:"dingTalkUserId"`
	UnionID        string `json:"unionId"`
	IssuedAt       int64  `json:"issuedAt"`
	ExpiresAt      int64  `json:"expiresAt"`
}

func SelfBindEnabledContext(ctx context.Context) bool {
	value := strings.ToLower(strings.TrimSpace(dingTalkH5SetupValueContext(ctx, dingTalkH5SelfBindEnabledKey)))
	if value == "" {
		return true
	}
	return value == "1" || value == "true" || value == "yes" || value == "on"
}

func newDingTalkH5BindRequiredErrorContext(ctx context.Context, corpID string, identity DingTalkUserIdentity) error {
	if !SelfBindEnabledContext(ctx) {
		return fmt.Errorf("钉钉账号未开通绩效系统，请联系管理员")
	}
	ticket, expiresIn, err := createDingTalkH5SelfBindTicketContext(ctx, corpID, identity)
	if err != nil {
		return fmt.Errorf("自助绑定暂不可用，请联系管理员")
	}
	appConfig := dingTalkH5AppConfigContext(ctx)
	return &DingTalkH5BindRequiredError{
		Response: DingTalkH5BindRequiredResponse{
			BindTicket:           ticket,
			CorpID:               strings.TrimSpace(corpID),
			DingTalkUserIDMasked: maskDingTalkIdentity(identity.UserID),
			UnionIDMasked:        maskDingTalkIdentity(identity.UnionID),
			ExpiresIn:            int64(expiresIn.Seconds()),
			AppConfig:            appConfig,
			AppTitle:             appConfig.AppTitle,
			AppName:              appConfig.AppName,
			LogoText:             appConfig.LogoText,
			LogoURL:              appConfig.LogoURL,
		},
	}
}

func BindSelfContext(ctx context.Context, bindTicket, account, password, addIP, device string) (*LoginResponse, error) {
	bindTicket = strings.TrimSpace(bindTicket)
	account = strings.TrimSpace(account)
	password = strings.TrimSpace(password)
	if !SelfBindEnabledContext(ctx) {
		return nil, fmt.Errorf("钉钉自助绑定已关闭")
	}
	if bindTicket == "" {
		return nil, fmt.Errorf("绑定会话已过期，请重新打开应用")
	}
	if account == "" || password == "" {
		return nil, fmt.Errorf("请填写系统账号和密码")
	}

	ticket, err := readDingTalkH5SelfBindTicketContext(ctx, bindTicket)
	if err != nil {
		return nil, err
	}

	db, cancel := database.WithContext(ctx)
	defer cancel()
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	baseUser, err := loadSelfBindUserByAccountDB(db, account)
	if err != nil {
		return nil, fmt.Errorf("账号或密码错误")
	}
	if baseUser.Status != 1 {
		return nil, fmt.Errorf("账号已停用")
	}
	if !passwordutil.Verify(baseUser.Password, password) {
		return nil, fmt.Errorf("账号或密码错误")
	}
	if passwordutil.NeedsRehash(baseUser.Password) {
		if hash, err := passwordutil.Hash(password); err == nil {
			_ = db.Model(&model.User{}).Where("`id` = ?", baseUser.ID).Update("user_password", hash).Error
		}
	}

	if err := bindDingTalkH5IdentityToUserDB(db, ticket, baseUser.ID); err != nil {
		return nil, err
	}
	_ = deleteDingTalkH5SelfBindTicketContext(ctx, bindTicket)

	user, err := loadPerfUserByIDDB(db, baseUser.ID)
	if err != nil {
		return nil, fmt.Errorf("账号异常")
	}
	if user.Status != 1 {
		return nil, fmt.Errorf("账号已停用")
	}
	return issueDingTalkH5LoginResponseContext(ctx, db, user, addIP, device)
}

func bindDingTalkH5IdentityToUserDB(db *gorm.DB, ticket dingTalkH5SelfBindTicket, userID uint) error {
	if db == nil || userID == 0 {
		return fmt.Errorf("绑定信息异常")
	}
	return db.Transaction(func(tx *gorm.DB) error {
		identityBinding, err := loadAnyDingTalkH5UserBindingDB(tx, ticket.CorpID, ticket.DingTalkUserID)
		if err == nil {
			if identityBinding.Enabled != 1 {
				return fmt.Errorf("钉钉账号绑定已停用，请联系管理员")
			}
			if identityBinding.UserID != userID {
				return fmt.Errorf("该钉钉账号已绑定其他系统用户，请联系管理员")
			}
			return tx.Model(identityBinding).Updates(map[string]interface{}{
				"union_id":  ticket.UnionID,
				"edit_time": database.Now(),
			}).Error
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		var disabledUserBindingCount int64
		if err := tx.Model(&model.DingTalkH5UserBinding{}).
			Where("`corp_id` = ? AND `user_id` = ? AND `enabled` <> 1", ticket.CorpID, userID).
			Count(&disabledUserBindingCount).Error; err != nil {
			return err
		}
		if disabledUserBindingCount > 0 {
			return fmt.Errorf("该系统账号绑定已停用，请联系管理员")
		}

		var otherUserBindingCount int64
		if err := tx.Model(&model.DingTalkH5UserBinding{}).
			Where("`corp_id` = ? AND `user_id` = ? AND `enabled` = 1 AND `dingtalk_user_id` <> ?", ticket.CorpID, userID, ticket.DingTalkUserID).
			Count(&otherUserBindingCount).Error; err != nil {
			return err
		}
		if otherUserBindingCount > 0 {
			return fmt.Errorf("该系统账号已绑定其他钉钉用户，请联系管理员")
		}

		if strings.TrimSpace(ticket.UnionID) != "" {
			var unionConflictCount int64
			if err := tx.Model(&model.DingTalkH5UserBinding{}).
				Where("`corp_id` = ? AND `union_id` = ? AND `enabled` = 1 AND `user_id` <> ?", ticket.CorpID, ticket.UnionID, userID).
				Count(&unionConflictCount).Error; err != nil {
				return err
			}
			if unionConflictCount > 0 {
				return fmt.Errorf("该钉钉账号已绑定其他系统用户，请联系管理员")
			}
		}

		now := database.Now()
		return tx.Create(&model.DingTalkH5UserBinding{
			CorpID:         ticket.CorpID,
			DingTalkUserID: ticket.DingTalkUserID,
			UnionID:        ticket.UnionID,
			UserID:         userID,
			Enabled:        1,
			AddTime:        now,
			EditTime:       now,
		}).Error
	})
}

func loadSelfBindUserByAccountDB(db *gorm.DB, account string) (*model.User, error) {
	account = strings.TrimSpace(account)
	normalizedAccount := NormalizeUserID(account)
	var user model.User
	if err := db.Where(
		"`user_mini_openid` = ? OR `user_name` = ? OR `user_account` = ? OR `user_mobile` = ?",
		normalizedAccount,
		account,
		account,
		account,
	).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func createDingTalkH5SelfBindTicketContext(ctx context.Context, corpID string, identity DingTalkUserIdentity) (string, time.Duration, error) {
	if rd.RDB == nil {
		return "", 0, fmt.Errorf("redis is not initialized")
	}
	ttl := dingTalkH5SelfBindTicketTTLContext(ctx)
	now := time.Now()
	ticket := randutil.HexString(48)
	payload := dingTalkH5SelfBindTicket{
		CorpID:         strings.TrimSpace(corpID),
		DingTalkUserID: strings.TrimSpace(identity.UserID),
		UnionID:        strings.TrimSpace(identity.UnionID),
		IssuedAt:       now.UnixMilli(),
		ExpiresAt:      now.Add(ttl).UnixMilli(),
	}
	if payload.CorpID == "" || payload.DingTalkUserID == "" {
		return "", 0, fmt.Errorf("invalid dingTalk bind ticket payload")
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", 0, err
	}
	redisCtx, cancel := rd.OperationContext(ctx)
	defer cancel()
	if err := rd.RDB.Set(redisCtx, dingTalkH5SelfBindTicketRedisKey(ticket), string(raw), ttl).Err(); err != nil {
		return "", 0, err
	}
	return ticket, ttl, nil
}

func readDingTalkH5SelfBindTicketContext(ctx context.Context, ticket string) (dingTalkH5SelfBindTicket, error) {
	if rd.RDB == nil {
		return dingTalkH5SelfBindTicket{}, fmt.Errorf("绑定会话已过期，请重新打开应用")
	}
	redisCtx, cancel := rd.OperationContext(ctx)
	defer cancel()
	raw, err := rd.RDB.Get(redisCtx, dingTalkH5SelfBindTicketRedisKey(ticket)).Result()
	if errors.Is(err, goredis.Nil) {
		return dingTalkH5SelfBindTicket{}, fmt.Errorf("绑定会话已过期，请重新打开应用")
	}
	if err != nil {
		return dingTalkH5SelfBindTicket{}, fmt.Errorf("绑定会话读取失败，请稍后重试")
	}
	var payload dingTalkH5SelfBindTicket
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return dingTalkH5SelfBindTicket{}, fmt.Errorf("绑定会话异常，请重新打开应用")
	}
	if strings.TrimSpace(payload.CorpID) == "" || strings.TrimSpace(payload.DingTalkUserID) == "" || time.Now().UnixMilli() > payload.ExpiresAt {
		return dingTalkH5SelfBindTicket{}, fmt.Errorf("绑定会话已过期，请重新打开应用")
	}
	return payload, nil
}

func deleteDingTalkH5SelfBindTicketContext(ctx context.Context, ticket string) error {
	if rd.RDB == nil {
		return nil
	}
	redisCtx, cancel := rd.OperationContext(ctx)
	defer cancel()
	return rd.RDB.Del(redisCtx, dingTalkH5SelfBindTicketRedisKey(ticket)).Err()
}

func dingTalkH5SelfBindTicketRedisKey(ticket string) string {
	return dingTalkH5SelfBindTicketPrefix + strings.TrimSpace(ticket)
}

func dingTalkH5SelfBindTicketTTLContext(ctx context.Context) time.Duration {
	value := strings.TrimSpace(dingTalkH5SetupValueContext(ctx, dingTalkH5SelfBindTicketTTLKey))
	if value == "" {
		return dingTalkH5SelfBindDefaultTTL
	}
	if seconds, err := strconv.Atoi(value); err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}
	if duration, err := time.ParseDuration(value); err == nil && duration > 0 {
		return duration
	}
	return dingTalkH5SelfBindDefaultTTL
}

func maskDingTalkIdentity(value string) string {
	value = strings.TrimSpace(value)
	runes := []rune(value)
	if len(runes) == 0 {
		return ""
	}
	if len(runes) <= 4 {
		return string(runes[:1]) + "***"
	}
	return string(runes[:2]) + "****" + string(runes[len(runes)-2:])
}
