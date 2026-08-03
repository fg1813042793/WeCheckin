package dingtalkh5

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/pkg/database"
)

func LoginByAuthCodeContext(ctx context.Context, corpID, authCode, addIP, device string) (*LoginResponse, error) {
	return loginByAuthCodeContext(ctx, defaultDingTalkIdentityClient{}, corpID, authCode, addIP, device)
}

func loginByAuthCodeContext(ctx context.Context, client DingTalkIdentityClient, corpID, authCode, addIP, device string) (*LoginResponse, error) {
	corpID = strings.TrimSpace(corpID)
	if corpID == "" {
		return nil, fmt.Errorf("钉钉企业 CorpId 不能为空")
	}
	authCode = strings.TrimSpace(authCode)
	if authCode == "" {
		return nil, fmt.Errorf("免登授权码不能为空")
	}
	config, err := loadDingTalkH5CorpConfigContext(ctx, corpID)
	if err != nil {
		return nil, err
	}
	identity, err := client.ExchangeAuthCodeContext(ctx, config, authCode)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(identity.UserID) == "" {
		return nil, fmt.Errorf("钉钉身份异常")
	}

	db, cancel := database.WithContext(ctx)
	defer cancel()
	if db == nil {
		return nil, fmt.Errorf("钉钉账号未开通绩效系统，请联系管理员")
	}

	binding, err := loadDingTalkH5UserBindingDB(db, corpID, identity.UserID)
	var loadedUser *model.DingTalkH5PerfUser
	if err != nil {
		if isMissingDingTalkH5UserBindingTableError(err) {
			loadedUser, err = loadLegacyDingTalkH5UserByIdentityDB(ctx, db, config, identity)
			if err != nil {
				return nil, fmt.Errorf("钉钉账号未开通绩效系统，请联系管理员")
			}
		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			existingBinding, existingErr := loadAnyDingTalkH5UserBindingDB(db, corpID, identity.UserID)
			if existingErr == nil && existingBinding.Enabled != 1 {
				return nil, fmt.Errorf("钉钉账号绑定已停用，请联系管理员")
			}
			if existingErr != nil && !errors.Is(existingErr, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("钉钉账号未开通绩效系统，请联系管理员")
			}
			return nil, newDingTalkH5BindRequiredErrorContext(ctx, corpID, identity)
		} else {
			return nil, fmt.Errorf("钉钉账号未开通绩效系统，请联系管理员")
		}
	} else {
		loadedUser, err = loadPerfUserByIDDB(db, binding.UserID)
		if err != nil {
			return nil, fmt.Errorf("钉钉账号未开通绩效系统，请联系管理员")
		}
	}
	user := *loadedUser
	if user.Status != 1 {
		return nil, fmt.Errorf("账号已停用")
	}
	if err := hydratePerfUserWithUserDeptDB(db, &user); err != nil {
		return nil, err
	}
	// Make the local binding explicit: DingTalk only proves identity; WeCheckin still owns permissions.
	return issueDingTalkH5LoginResponseContext(ctx, db, &user, addIP, device)
}
