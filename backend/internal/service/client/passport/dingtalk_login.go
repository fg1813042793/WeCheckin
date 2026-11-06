package passport

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"

	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/service/accountsecurity"
	onlineservice "wecheckin/backend/internal/service/admin/online"
	dingtalkconfig "wecheckin/backend/internal/service/dingtalkh5/config"
	"wecheckin/backend/internal/support/auditlog"
	"wecheckin/backend/internal/support/dept"
	"wecheckin/backend/internal/support/media"
	"wecheckin/backend/pkg/database"
	"wecheckin/backend/pkg/randutil"
)

const dingTalkOAuthAuthorizeEndpoint = "https://login.dingtalk.com/oauth2/auth"

var (
	ErrDingTalkLoginNotConfigured = errors.New("dingtalk client login not configured")
	ErrDingTalkInvalidRedirect    = errors.New("invalid dingtalk oauth redirect uri")
	ErrDingTalkBindingRequired    = errors.New("dingtalk account is not bound")
	ErrDingTalkBindingDisabled    = errors.New("dingtalk account binding is disabled")
	ErrDingTalkBindingAmbiguous   = errors.New("dingtalk account binding is ambiguous")
	ErrDingTalkUserDisabled       = errors.New("bound local user is disabled")
)

type DingTalkLoginCorpOption struct {
	CorpID   string `json:"corpId"`
	CorpName string `json:"corpName"`
}

type DingTalkLoginConfigResponse struct {
	Enabled     bool                      `json:"enabled"`
	CorpOptions []DingTalkLoginCorpOption `json:"corpOptions"`
}

type DingTalkAuthorizationResponse struct {
	AuthorizationURL string `json:"authorizationUrl"`
	State            string `json:"state"`
	ClientID         string `json:"clientId"`
	RedirectURI      string `json:"redirectUri"`
}

func DingTalkLoginConfigContext(ctx context.Context) (*DingTalkLoginConfigResponse, error) {
	configs, err := dingtalkconfig.ListDingTalkH5CorpConfigsContext(ctx)
	if err != nil {
		return nil, err
	}
	options := make([]DingTalkLoginCorpOption, 0, len(configs))
	for _, config := range configs {
		if config.Enabled != 1 || strings.TrimSpace(config.AppKey) == "" || strings.TrimSpace(config.AppSecret) == "" {
			continue
		}
		name := strings.TrimSpace(config.CorpName)
		if name == "" {
			name = strings.TrimSpace(config.CorpID)
		}
		options = append(options, DingTalkLoginCorpOption{CorpID: strings.TrimSpace(config.CorpID), CorpName: name})
	}
	return &DingTalkLoginConfigResponse{Enabled: len(options) > 0, CorpOptions: options}, nil
}

func CreateDingTalkAuthorizationContext(ctx context.Context, corpID, redirectURI string) (*DingTalkAuthorizationResponse, error) {
	corpID = strings.TrimSpace(corpID)
	if corpID == "" {
		auditlog.SetReason(ctx, "invalid_request", "缺少钉钉企业标识")
		return nil, ErrDingTalkLoginNotConfigured
	}
	if err := validateDingTalkOAuthRedirectURI(redirectURI); err != nil {
		auditlog.SetReason(ctx, "invalid_redirect", "钉钉登录回调地址无效")
		return nil, err
	}
	config, err := dingtalkconfig.LoadDingTalkH5CorpConfigContext(ctx, corpID)
	if err != nil {
		auditlog.SetReason(ctx, "system_error", "钉钉认证配置不可用")
		return nil, fmt.Errorf("%w: %v", ErrDingTalkLoginNotConfigured, err)
	}
	state := randutil.HexString(24)
	authorizationURL, err := buildDingTalkAuthorizationURL(config.AppKey, redirectURI, state)
	if err != nil {
		auditlog.SetReason(ctx, "system_error", "钉钉认证配置不可用")
		return nil, err
	}
	return &DingTalkAuthorizationResponse{
		AuthorizationURL: authorizationURL,
		State:            state,
		ClientID:         strings.TrimSpace(config.AppKey),
		RedirectURI:      strings.TrimSpace(redirectURI),
	}, nil
}

func LoginByDingTalkOAuthCodeContext(ctx context.Context, corpID, authCode, addIP, device string) (*LoginResponse, error) {
	return loginByDingTalkOAuthCodeContext(ctx, dingtalkconfig.DefaultDingTalkOAuthIdentityClient(), corpID, authCode, addIP, device)
}

func loginByDingTalkOAuthCodeContext(ctx context.Context, client dingtalkconfig.DingTalkOAuthIdentityClient, corpID, authCode, addIP, device string) (*LoginResponse, error) {
	corpID = strings.TrimSpace(corpID)
	authCode = strings.TrimSpace(authCode)
	if corpID == "" || authCode == "" {
		if authCode == "" {
			auditlog.SetReason(ctx, "auth_code_invalid", "钉钉授权码无效")
		} else {
			auditlog.SetReason(ctx, "invalid_request", "缺少钉钉企业标识")
		}
		return nil, fmt.Errorf("钉钉登录参数不完整")
	}
	config, err := dingtalkconfig.LoadDingTalkH5CorpConfigContext(ctx, corpID)
	if err != nil {
		auditlog.SetReason(ctx, "system_error", "钉钉认证配置不可用")
		return nil, fmt.Errorf("%w: %v", ErrDingTalkLoginNotConfigured, err)
	}
	identity, err := client.ExchangeOAuthCodeContext(ctx, config, authCode)
	if err != nil {
		var networkErr net.Error
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) || errors.As(err, &networkErr) {
			auditlog.SetReason(ctx, "system_error", "钉钉认证服务异常")
		} else {
			auditlog.SetReason(ctx, "auth_code_exchange_failed", "钉钉授权码换取身份失败")
		}
		return nil, err
	}
	if strings.TrimSpace(identity.UnionID) == "" {
		auditlog.SetReason(ctx, "auth_identity_invalid", "钉钉身份异常")
		return nil, fmt.Errorf("钉钉授权用户身份异常")
	}

	db, cancel := database.WithContext(ctx)
	defer cancel()
	if db == nil {
		auditlog.SetReason(ctx, "system_error", "认证服务异常")
		return nil, fmt.Errorf("数据库未初始化")
	}
	binding, err := loadDingTalkOAuthBindingDB(db, corpID, identity.UnionID)
	if err != nil {
		switch {
		case errors.Is(err, ErrDingTalkBindingRequired):
			auditlog.SetReason(ctx, "binding_required", "钉钉账号未绑定系统用户")
		case errors.Is(err, ErrDingTalkBindingDisabled):
			auditlog.SetReason(ctx, "binding_disabled", "账号绑定已停用")
		case errors.Is(err, ErrDingTalkBindingAmbiguous):
			auditlog.SetReason(ctx, "binding_conflict", "钉钉账号存在重复绑定")
		default:
			auditlog.SetReason(ctx, "system_error", "账号绑定查询失败")
		}
		return nil, err
	}
	var user model.User
	if err := db.Where("`id` = ?", binding.UserID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			auditlog.SetReason(ctx, "binding_required", "钉钉账号未绑定有效系统用户")
			return nil, ErrDingTalkBindingRequired
		}
		auditlog.SetReason(ctx, "system_error", "登录账号查询失败")
		return nil, err
	}
	if user.Status != 1 {
		auditlog.SetReason(ctx, "account_disabled", "账号已停用")
		return nil, ErrDingTalkUserDisabled
	}
	setUserRole(&user)
	fillUserRoleIDsContext(ctx, db, &user)
	if err := db.Model(&user).Updates(map[string]interface{}{
		"user_login_time": database.Now(),
		"user_login_cnt":  gorm.Expr("user_login_cnt + 1"),
	}).Error; err != nil {
		auditlog.SetReason(ctx, "system_error", "登录状态保存失败")
		return nil, err
	}
	return issueDingTalkClientLoginResponseContext(ctx, db, &user, addIP, device)
}

func loadDingTalkOAuthBindingDB(db *gorm.DB, corpID, unionID string) (*model.DingTalkH5UserBinding, error) {
	var enabled []model.DingTalkH5UserBinding
	if err := db.Where("`corp_id` = ? AND `union_id` = ? AND `enabled` = 1", corpID, strings.TrimSpace(unionID)).Limit(2).Find(&enabled).Error; err != nil {
		return nil, err
	}
	if len(enabled) > 1 {
		return nil, ErrDingTalkBindingAmbiguous
	}
	if len(enabled) == 1 {
		return &enabled[0], nil
	}
	var count int64
	if err := db.Model(&model.DingTalkH5UserBinding{}).Where("`corp_id` = ? AND `union_id` = ?", corpID, strings.TrimSpace(unionID)).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrDingTalkBindingDisabled
	}
	return nil, ErrDingTalkBindingRequired
}

func issueDingTalkClientLoginResponseContext(ctx context.Context, db *gorm.DB, user *model.User, addIP, device string) (*LoginResponse, error) {
	if user == nil || user.ID == 0 {
		auditlog.SetReason(ctx, "system_error", "登录信息异常")
		return nil, fmt.Errorf("登录信息异常")
	}
	passwordState, err := accountsecurity.EvaluatePasswordStateContext(ctx, user.PasswordChangedAt, user.PasswordMustChange == 1)
	if err != nil {
		return nil, err
	}
	token := randutil.HexString(32)
	if err := onlineservice.StoreUserSessionContext(ctx, user, token, addIP, device); err != nil {
		auditlog.SetReason(ctx, "system_error", "登录会话创建失败")
		return nil, err
	}
	deptID := dept.UserDeptIDContext(ctx, user.ID)
	var deptName string
	if deptID > 0 {
		var department model.Department
		if err := db.Select("dept_name").First(&department, deptID).Error; err == nil {
			deptName = department.Name
		}
	}
	auditlog.SetActor(ctx, user.ID, user.Name, "user")
	response := &LoginResponse{
		Token: token,
		UserInfo: &UserInfo{
			ID: user.ID, Name: user.Name, Avatar: media.FullURLWithStaticDomain(user.Pic), Desc: "点击完善个人信息",
			MiniOpenID: user.MiniOpenID, Role: user.Role, DeptID: deptID, DeptName: deptName,
		},
	}
	applyPasswordState(response, passwordState)
	return response, nil
}

func buildDingTalkAuthorizationURL(appKey, redirectURI, state string) (string, error) {
	appKey = strings.TrimSpace(appKey)
	state = strings.TrimSpace(state)
	if appKey == "" || state == "" {
		return "", ErrDingTalkLoginNotConfigured
	}
	if err := validateDingTalkOAuthRedirectURI(redirectURI); err != nil {
		return "", err
	}
	endpoint, _ := url.Parse(dingTalkOAuthAuthorizeEndpoint)
	query := endpoint.Query()
	query.Set("redirect_uri", strings.TrimSpace(redirectURI))
	query.Set("response_type", "code")
	query.Set("client_id", appKey)
	query.Set("scope", "openid")
	query.Set("state", state)
	query.Set("prompt", "consent")
	endpoint.RawQuery = query.Encode()
	return endpoint.String(), nil
}

func validateDingTalkOAuthRedirectURI(value string) error {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > 1000 {
		return ErrDingTalkInvalidRedirect
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.User != nil || parsed.Fragment != "" {
		return ErrDingTalkInvalidRedirect
	}
	switch strings.ToLower(parsed.Scheme) {
	case "https":
		if parsed.Hostname() == "" {
			return ErrDingTalkInvalidRedirect
		}
		return nil
	case "http":
		if parsed.Hostname() != "" {
			return nil
		}
	}
	return ErrDingTalkInvalidRedirect
}
