package online

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/support/media"
	"wecheckin/backend/pkg/database"
	rd "wecheckin/backend/pkg/redis"
	"wecheckin/backend/pkg/tokenutil"
)

func GetOnlineUsers() ([]UserSession, error) {
	return GetOnlineUsersContext(context.Background())
}

type userSessionRole struct {
	role   string
	source string
}

var userSessionRoles = []userSessionRole{
	{role: "user", source: "客户端"},
	{role: "dingtalk_h5", source: "钉钉 H5"},
}

func GetOnlineUsersContext(ctx context.Context) ([]UserSession, error) {
	if rd.RDB == nil {
		return []UserSession{}, nil
	}
	type roleEntries struct {
		source     string
		authPrefix string
		entries    []entry
	}
	groups := make([]roleEntries, 0, len(userSessionRoles))
	allEntries := make([]entry, 0)
	for _, item := range userSessionRoles {
		_, prefix := tokenutil.GetTokenConfigContext(ctx, item.role)
		entries, err := scanSets(ctx, prefix+"s:")
		if err != nil {
			return nil, err
		}
		groups = append(groups, roleEntries{
			source:     item.source,
			authPrefix: prefix + "a:",
			entries:    entries,
		})
		allEntries = append(allEntries, entries...)
	}
	loadBase, err := preloadUserBase(ctx, allEntries)
	if err != nil {
		return nil, err
	}
	rows := make([]UserSession, 0, len(allEntries))
	for _, group := range groups {
		groupRows := buildRows(ctx, group.entries, group.authPrefix, loadBase, func(row UserSession, info SessionInfo) UserSession {
			info.Source = group.source
			row.SessionInfo = info
			return row
		})
		rows = append(rows, groupRows...)
	}
	return rows, nil
}

func StoreUserSessionContext(ctx context.Context, user *model.User, token, addIP, device string) error {
	return storeUserSessionForRoleContext(ctx, "user", tokenutil.IsUserSingleLoginContext(ctx), user, token, addIP, device)
}

func StoreDingTalkH5SessionContext(ctx context.Context, user *model.User, token, addIP, device string) error {
	return storeUserSessionForRoleContext(ctx, "dingtalk_h5", tokenutil.IsDingTalkH5SingleLoginContext(ctx), user, token, addIP, device)
}

func UpdateDingTalkH5SessionUserContext(ctx context.Context, user *model.User, token string) error {
	return updateUserSessionForRoleContext(ctx, "dingtalk_h5", user, token)
}

func storeUserSessionForRoleContext(ctx context.Context, role string, singleLogin bool, user *model.User, token, addIP, device string) error {
	if user == nil || user.ID == 0 || token == "" {
		return fmt.Errorf("登录信息异常")
	}
	expire, prefix := tokenutil.GetTokenConfigContext(ctx, role)
	if rd.RDB == nil {
		return fmt.Errorf("服务异常")
	}
	keyAuth := prefix + "a:" + token
	idStr := strconv.Itoa(int(user.ID))
	keySet := prefix + "s:" + idStr
	redisCtx, cancel := rd.OperationContext(ctx)
	defer cancel()

	if singleLogin {
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
	info := storedUserSessionPayload{
		ID:         user.ID,
		Name:       user.Name,
		Mobile:     user.Mobile,
		MiniOpenID: user.MiniOpenID,
		Role:       user.Role,
		RoleID:     user.RoleID,
		Pic:        user.Pic,
		LoginIP:    addIP,
		LoginTime:  now,
		Device:     device,
	}
	jsonBytes, _ := json.Marshal(info)
	if err := rd.RDB.Set(redisCtx, keyAuth, string(jsonBytes), expire).Err(); err != nil {
		return err
	}
	if err := rd.RDB.SAdd(redisCtx, keySet, token).Err(); err != nil {
		rd.RDB.Del(redisCtx, keyAuth)
		return err
	}
	rd.RDB.Expire(redisCtx, keySet, expire*2)
	return nil
}

func updateUserSessionForRoleContext(ctx context.Context, role string, user *model.User, token string) error {
	if user == nil || user.ID == 0 || strings.TrimSpace(token) == "" {
		return fmt.Errorf("登录信息异常")
	}
	token = strings.TrimSpace(token)
	expire, prefix := tokenutil.GetTokenConfigContext(ctx, role)
	if rd.RDB == nil {
		return fmt.Errorf("服务异常")
	}
	payload, err := loadUserSessionPayloadContext(ctx, role, token)
	if err != nil {
		return err
	}
	payload.ID = user.ID
	payload.Name = user.Name
	payload.Mobile = user.Mobile
	payload.MiniOpenID = user.MiniOpenID
	payload.Role = user.Role
	payload.RoleID = user.RoleID
	payload.Pic = user.Pic
	jsonBytes, _ := json.Marshal(payload)
	redisCtx, cancel := rd.OperationContext(ctx)
	defer cancel()
	return rd.RDB.Set(redisCtx, prefix+"a:"+token, string(jsonBytes), expire).Err()
}

func RequireUserSessionContext(ctx context.Context, token string) error {
	_, err := loadUserSessionPayloadContext(ctx, "user", token)
	return err
}

func DingTalkH5SessionAccountContext(ctx context.Context, token string) (string, error) {
	payload, err := loadUserSessionPayloadContext(ctx, "dingtalk_h5", token)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(payload.MiniOpenID) == "" {
		return "", fmt.Errorf("登录信息异常")
	}
	return payload.MiniOpenID, nil
}

func loadUserSessionPayloadContext(ctx context.Context, role, token string) (storedUserSessionPayload, error) {
	var payload storedUserSessionPayload
	if token == "" {
		return payload, fmt.Errorf("未登录")
	}
	expire, prefix := tokenutil.GetTokenConfigContext(ctx, role)
	if rd.RDB == nil {
		return payload, fmt.Errorf("服务异常")
	}
	redisCtx, cancel := rd.OperationContext(ctx)
	defer cancel()
	raw, err := rd.RDB.Get(redisCtx, prefix+"a:"+token).Result()
	if err != nil {
		return payload, fmt.Errorf("登录已过期或已被强制下线")
	}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return payload, fmt.Errorf("登录信息异常")
	}
	rd.RDB.Expire(redisCtx, prefix+"a:"+token, expire)
	return payload, nil
}

func RemoveUserSessionContext(ctx context.Context, userID uint, token string) error {
	if userID == 0 {
		return RemoveUserTokenContext(ctx, token)
	}
	return ForceOfflineUserContext(ctx, strconv.Itoa(int(userID)), token)
}

func RemoveDingTalkH5SessionContext(ctx context.Context, userID uint, token string) error {
	return removeUserSessionForRoleContext(ctx, "dingtalk_h5", userID, token)
}

func RemoveUserTokenContext(ctx context.Context, token string) error {
	return removeUserTokenForRoleContext(ctx, "user", token)
}

func removeUserSessionForRoleContext(ctx context.Context, role string, userID uint, token string) error {
	if userID == 0 {
		return removeUserTokenForRoleContext(ctx, role, token)
	}
	return forceOfflineUserForRoleContext(ctx, role, strconv.Itoa(int(userID)), token)
}

func removeUserTokenForRoleContext(ctx context.Context, role, token string) error {
	token = strings.TrimSpace(token)
	_, prefix := tokenutil.GetTokenConfigContext(ctx, role)
	if rd.RDB == nil || token == "" {
		return nil
	}
	redisCtx, cancel := rd.OperationContext(ctx)
	defer cancel()
	return rd.RDB.Del(redisCtx, prefix+"a:"+token).Err()
}

func preloadUserBase(ctx context.Context, entries []entry) (func(uint64) (UserSession, bool), error) {
	uids := make([]uint, 0, len(entries))
	for _, e := range entries {
		uids = append(uids, uint(e.uid))
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var users []model.User
	if len(uids) > 0 {
		if err := db.Where("id IN ?", uids).Find(&users).Error; err != nil {
			return nil, err
		}
	}
	userByID := make(map[uint]*model.User, len(users))
	for i := range users {
		userByID[users[i].ID] = &users[i]
	}

	return func(uid uint64) (UserSession, bool) {
		u, ok := userByID[uint(uid)]
		if !ok {
			return UserSession{}, false
		}
		return UserSession{
			ID:       u.ID,
			Name:     u.Name,
			Mobile:   u.Mobile,
			Pic:      media.FullURLWithStaticDomain(u.Pic),
			LoginCnt: u.LoginCnt,
		}, true
	}, nil
}

func ForceOfflineUser(idStr, token string) error {
	return ForceOfflineUserContext(context.Background(), idStr, token)
}

func ForceOfflineUserContext(ctx context.Context, idStr, token string) error {
	idStr = strings.TrimSpace(idStr)
	token = strings.TrimSpace(token)
	if rd.RDB == nil || token == "" {
		return nil
	}
	for _, item := range userSessionRoles {
		if err := forceOfflineUserForRoleContext(ctx, item.role, idStr, token); err != nil {
			return err
		}
	}
	return nil
}

func forceOfflineUserForRoleContext(ctx context.Context, role, idStr, token string) error {
	_, prefix := tokenutil.GetTokenConfigContext(ctx, role)
	redisCtx, cancel := rd.OperationContext(ctx)
	defer cancel()
	if err := rd.RDB.Del(redisCtx, prefix+"a:"+token).Err(); err != nil {
		return err
	}
	if idStr == "" {
		return nil
	}
	setKey := prefix + "s:" + idStr
	if err := rd.RDB.SRem(redisCtx, setKey, token).Err(); err != nil {
		return err
	}
	count, err := rd.RDB.SCard(redisCtx, setKey).Result()
	if err != nil {
		return err
	}
	if count == 0 {
		if err := rd.RDB.Del(redisCtx, setKey).Err(); err != nil {
			return err
		}
	}
	return nil
}

func BatchForceOfflineUser(items []struct {
	IDStr string `json:"idStr"`
	Token string `json:"token"`
}) (int, error) {
	return BatchForceOfflineUserContext(context.Background(), items)
}

func BatchForceOfflineUserContext(ctx context.Context, items []struct {
	IDStr string `json:"idStr"`
	Token string `json:"token"`
}) (int, error) {
	if rd.RDB == nil || len(items) == 0 {
		return 0, nil
	}
	for _, item := range userSessionRoles {
		if err := batchForceOfflineUserForRoleContext(ctx, item.role, items); err != nil {
			return 0, err
		}
	}
	return len(items), nil
}

func batchForceOfflineUserForRoleContext(ctx context.Context, role string, items []struct {
	IDStr string `json:"idStr"`
	Token string `json:"token"`
}) error {
	_, prefix := tokenutil.GetTokenConfigContext(ctx, role)
	redisCtx, cancel := rd.OperationContext(ctx)
	defer cancel()
	byID := make(map[string][]string, len(items))
	for _, it := range items {
		if it.Token == "" {
			continue
		}
		byID[it.IDStr] = append(byID[it.IDStr], it.Token)
	}

	pipe := rd.RDB.Pipeline()
	for idStr, tokens := range byID {
		authKeys := make([]string, len(tokens))
		for i, t := range tokens {
			authKeys[i] = prefix + "a:" + t
		}
		pipe.Del(redisCtx, authKeys...)
		pipe.SRem(redisCtx, prefix+"s:"+idStr, stringSliceToInterface(tokens)...)
	}
	if _, err := pipe.Exec(redisCtx); err != nil && err != redis.Nil {
		return err
	}
	for idStr := range byID {
		setKey := prefix + "s:" + idStr
		if n, _ := rd.RDB.SCard(redisCtx, setKey).Result(); n == 0 {
			rd.RDB.Del(redisCtx, setKey)
		}
	}
	return nil
}
