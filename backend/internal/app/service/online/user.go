package online

import (
	"context"

	"github.com/redis/go-redis/v9"

	"wecheckin-backend/backend/internal/app/support/media"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
	rd "wecheckin-backend/backend/pkg/redis"
	"wecheckin-backend/backend/pkg/tokenutil"
)

func GetOnlineUsers() ([]UserSession, error) {
	return GetOnlineUsersContext(context.Background())
}

func GetOnlineUsersContext(ctx context.Context) ([]UserSession, error) {
	_, prefix := tokenutil.GetTokenConfig("user")
	if rd.RDB == nil {
		return []UserSession{}, nil
	}
	setPrefix := prefix + "s:"
	authPrefix := prefix + "a:"

	entries, err := scanSets(ctx, setPrefix)
	if err != nil {
		return nil, err
	}
	loadBase, err := preloadUserBase(ctx, entries)
	if err != nil {
		return nil, err
	}
	return buildRows(ctx, entries, authPrefix, loadBase, func(row UserSession, info SessionInfo) UserSession {
		row.SessionInfo = info
		return row
	}), nil
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
	_, prefix := tokenutil.GetTokenConfig("user")
	if rd.RDB == nil || token == "" {
		return nil
	}
	redisCtx, cancel := rd.OperationContext(ctx)
	defer cancel()
	rd.RDB.Del(redisCtx, prefix+"a:"+token)
	rd.RDB.SRem(redisCtx, prefix+"s:"+idStr, token)
	if count, _ := rd.RDB.SCard(redisCtx, prefix+"s:"+idStr).Result(); count == 0 {
		rd.RDB.Del(redisCtx, prefix+"s:"+idStr)
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
	_, prefix := tokenutil.GetTokenConfig("user")
	if rd.RDB == nil || len(items) == 0 {
		return 0, nil
	}
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
		return 0, err
	}
	for idStr := range byID {
		setKey := prefix + "s:" + idStr
		if n, _ := rd.RDB.SCard(redisCtx, setKey).Result(); n == 0 {
			rd.RDB.Del(redisCtx, setKey)
		}
	}
	return len(items), nil
}
