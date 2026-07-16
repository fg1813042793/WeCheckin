package online

import (
	"context"
	"strconv"

	"github.com/redis/go-redis/v9"

	"wecheckin-backend/backend/internal/app/support/media"
	"wecheckin-backend/backend/internal/model"
	"wecheckin-backend/backend/pkg/database"
	rd "wecheckin-backend/backend/pkg/redis"
	"wecheckin-backend/backend/pkg/tokenutil"
)

func GetOnlineAdmins() ([]map[string]interface{}, error) {
	_, prefix := tokenutil.GetTokenConfig("admin")
	if rd.RDB == nil {
		return []map[string]interface{}{}, nil
	}
	setPrefix := prefix + "s:"
	authPrefix := prefix + "a:"

	entries, err := scanSets(setPrefix)
	if err != nil {
		return nil, err
	}
	loadBase := preloadAdminBase(entries)
	return buildRows(entries, authPrefix, loadBase), nil
}

func preloadAdminBase(entries []entry) func(uint64) (map[string]interface{}, bool) {
	uids := make([]uint, 0, len(entries))
	for _, e := range entries {
		uids = append(uids, uint(e.uid))
	}
	var admins []model.Admin
	if len(uids) > 0 {
		database.DB.Where("id IN ?", uids).Find(&admins)
	}
	adminByID := make(map[uint]*model.Admin, len(admins))
	for i := range admins {
		adminByID[admins[i].ID] = &admins[i]
	}

	roleIDs := make([]uint, 0, len(admins))
	for _, a := range admins {
		if a.RoleID > 0 {
			roleIDs = append(roleIDs, a.RoleID)
		}
	}
	var roles []model.Role
	if len(roleIDs) > 0 {
		database.DB.Where("id IN ?", roleIDs).Find(&roles)
	}
	roleByID := make(map[uint]string, len(roles))
	for _, r := range roles {
		roleByID[r.ID] = r.Name
	}

	return func(uid uint64) (map[string]interface{}, bool) {
		a, ok := adminByID[uint(uid)]
		if !ok {
			return nil, false
		}
		roleName := ""
		if a.RoleID > 0 {
			roleName = roleByID[a.RoleID]
		}
		return map[string]interface{}{
			"id":       a.ID,
			"name":     a.Name,
			"desc":     a.Desc,
			"pic":      media.FullURLWithStaticDomain(a.Pic),
			"type":     a.Type,
			"roleName": roleName,
			"loginCnt": a.LoginCnt,
		}, true
	}
}

func ForceOfflineAdmin(idStr, token string) error {
	_, prefix := tokenutil.GetTokenConfig("admin")
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

func BatchForceOfflineAdmin(items []struct {
	IDStr string `json:"idStr"`
	Token string `json:"token"`
}) (int, error) {
	_, prefix := tokenutil.GetTokenConfig("admin")
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

func AdminLogout(adminID uint, currentToken string) error {
	_, prefix := tokenutil.GetTokenConfig("admin")
	if rd.RDB == nil {
		return nil
	}
	redisCtx, cancel := rd.OperationContext(context.Background())
	defer cancel()
	idStr := strconv.Itoa(int(adminID))
	if currentToken == "" {
		tokens, _ := rd.RDB.SMembers(redisCtx, prefix+"s:"+idStr).Result()
		for _, t := range tokens {
			rd.RDB.Del(redisCtx, prefix+"a:"+t)
		}
		rd.RDB.Del(redisCtx, prefix+"s:"+idStr)
		return nil
	}
	rd.RDB.Del(redisCtx, prefix+"a:"+currentToken)
	rd.RDB.SRem(redisCtx, prefix+"s:"+idStr, currentToken)
	if count, _ := rd.RDB.SCard(redisCtx, prefix+"s:"+idStr).Result(); count == 0 {
		rd.RDB.Del(redisCtx, prefix+"s:"+idStr)
	}
	return nil
}
