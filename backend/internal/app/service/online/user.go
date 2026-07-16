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

func GetOnlineUsers() ([]map[string]interface{}, error) {
	_, prefix := tokenutil.GetTokenConfig("user")
	if rd.RDB == nil {
		return []map[string]interface{}{}, nil
	}
	setPrefix := prefix + "s:"
	authPrefix := prefix + "a:"

	entries, err := scanSets(setPrefix)
	if err != nil {
		return nil, err
	}
	loadBase := preloadUserBase(entries)
	return buildRows(entries, authPrefix, loadBase), nil
}

func preloadUserBase(entries []entry) func(uint64) (map[string]interface{}, bool) {
	uids := make([]uint, 0, len(entries))
	for _, e := range entries {
		uids = append(uids, uint(e.uid))
	}
	var users []model.User
	if len(uids) > 0 {
		database.DB.Where("id IN ?", uids).Find(&users)
	}
	userByID := make(map[uint]*model.User, len(users))
	for i := range users {
		userByID[users[i].ID] = &users[i]
	}

	return func(uid uint64) (map[string]interface{}, bool) {
		u, ok := userByID[uint(uid)]
		if !ok {
			return nil, false
		}
		return map[string]interface{}{
			"id":       u.ID,
			"name":     u.Name,
			"mobile":   u.Mobile,
			"pic":      media.FullURLWithStaticDomain(u.Pic),
			"loginCnt": u.LoginCnt,
		}, true
	}
}

func ForceOfflineUser(idStr, token string) error {
	_, prefix := tokenutil.GetTokenConfig("user")
	if rd.RDB == nil || token == "" {
		return nil
	}
	redisCtx, cancel := rd.OperationContext(context.Background())
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
	_, prefix := tokenutil.GetTokenConfig("user")
	if rd.RDB == nil || len(items) == 0 {
		return 0, nil
	}
	redisCtx, cancel := rd.OperationContext(context.Background())
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
