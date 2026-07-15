package service

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"

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

	entries, err := scanOnlineSets(setPrefix)
	if err != nil {
		return nil, err
	}
	loadBase := preloadUserBase(entries)
	return buildOnlineRows(entries, authPrefix, loadBase), nil
}

// preloadUserBase fetches all user records in 1 batched query and returns a lookup closure.
func preloadUserBase(entries []onlineEntry) func(uint64) (map[string]interface{}, bool) {
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
			"pic":      GetFullURL(u.Pic),
			"loginCnt": u.LoginCnt,
		}, true
	}
}

// onlineEntry is a (setKey, uid, tokens) tuple built from Redis SCAN+SMEMBERS.
type onlineEntry struct {
	setKey string
	uid    uint64
	tokens []string
}

// scanOnlineSets SCANs all `s:` Set keys and returns per-user entries with
// their current token members. No DB or per-token I/O.
func scanOnlineSets(setPrefix string) ([]onlineEntry, error) {
	var cursor uint64
	var setKeys []string
	for {
		ks, c, err := rd.RDB.Scan(rd.Ctx, cursor, setPrefix+"*", 100).Result()
		if err != nil {
			return nil, err
		}
		setKeys = append(setKeys, ks...)
		cursor = c
		if cursor == 0 {
			break
		}
	}

	entries := make([]onlineEntry, 0, len(setKeys))
	for _, setKey := range setKeys {
		idStr := strings.TrimPrefix(setKey, setPrefix)
		uid, _ := strconv.ParseUint(idStr, 10, 64)
		if uid == 0 {
			continue
		}
		tokens, _ := rd.RDB.SMembers(rd.Ctx, setKey).Result()
		if len(tokens) == 0 {
			continue
		}
		entries = append(entries, onlineEntry{setKey, uid, tokens})
	}
	return entries, nil
}

// buildOnlineRows takes the entries from scanOnlineSets, fetches per-token info
// in a single pipelined round trip, joins with the per-user base info from
// `loadBase`, and prunes dead token references from Sets.
func buildOnlineRows(entries []onlineEntry, authPrefix string, loadBase func(uid uint64) (map[string]interface{}, bool)) []map[string]interface{} {
	// Pipeline: for each token, GET a:{token} + TTL a:{token}
	pipe := rd.RDB.Pipeline()
	type cmd struct {
		token string
		get   *redis.StringCmd
		ttl   *redis.DurationCmd
	}
	allCmds := make([]cmd, 0)
	for _, e := range entries {
		for _, t := range e.tokens {
			allCmds = append(allCmds, cmd{
				token: t,
				get:   pipe.Get(rd.Ctx, authPrefix+t),
				ttl:   pipe.TTL(rd.Ctx, authPrefix+t),
			})
		}
	}
	if len(allCmds) > 0 {
		_, _ = pipe.Exec(rd.Ctx)
	}

	result := make([]map[string]interface{}, 0)
	idx := 0
	for _, e := range entries {
		base, ok := loadBase(e.uid)
		if !ok {
			idx += len(e.tokens)
			continue
		}
		var deadTokens []string
		for _, t := range e.tokens {
			jsonStr, err := allCmds[idx].get.Result()
			ttl, _ := allCmds[idx].ttl.Result()
			idx++
			if err != nil {
				deadTokens = append(deadTokens, t)
				continue
			}
			var info struct {
				LoginIP   string `json:"loginIp"`
				LoginTime int64  `json:"loginTime"`
				Device    string `json:"device"`
			}
			row := map[string]interface{}{}
			for k, v := range base {
				row[k] = v
			}
			row["token"] = t
			row["ttl"] = int(ttl.Seconds())
			if json.Unmarshal([]byte(jsonStr), &info) == nil {
				row["loginIp"] = info.LoginIP
				row["loginTime"] = info.LoginTime
				row["device"] = info.Device
			} else {
				row["loginIp"] = ""
				row["loginTime"] = int64(0)
				row["device"] = ""
			}
			result = append(result, row)
		}
		if len(deadTokens) > 0 {
			rd.RDB.SRem(rd.Ctx, e.setKey, anyToIface(deadTokens)...)
		}
	}
	return result
}

func anyToIface(ss []string) []interface{} {
	out := make([]interface{}, len(ss))
	for i, s := range ss {
		out[i] = s
	}
	return out
}

func ForceOfflineUser(idStr, token string) error {
	_, prefix := tokenutil.GetTokenConfig("user")
	if rd.RDB == nil || token == "" {
		return nil
	}
	rd.RDB.Del(rd.Ctx, prefix+"a:"+token)
	rd.RDB.SRem(rd.Ctx, prefix+"s:"+idStr, token)
	if count, _ := rd.RDB.SCard(rd.Ctx, prefix+"s:"+idStr).Result(); count == 0 {
		rd.RDB.Del(rd.Ctx, prefix+"s:"+idStr)
	}
	return nil
}

// BatchForceOfflineUser 批量踢人。items = [{idStr, token}, ...]。
// 用 Redis pipeline 一次完成所有 SREM + DEL，对每个用户最后 SCard==0 时再 DEL Set。
func BatchForceOfflineUser(items []struct {
	IDStr string `json:"idStr"`
	Token string `json:"token"`
}) (int, error) {
	_, prefix := tokenutil.GetTokenConfig("user")
	if rd.RDB == nil || len(items) == 0 {
		return 0, nil
	}
	// Group tokens by user id (one user may have multiple devices selected)
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
		pipe.Del(rd.Ctx, authKeys...)
		pipe.SRem(rd.Ctx, prefix+"s:"+idStr, anyToIface(tokens)...)
	}
	if _, err := pipe.Exec(rd.Ctx); err != nil && err != redis.Nil {
		return 0, err
	}
	// 如果 Set 变空就 DEL（清理空 Set）
	for idStr := range byID {
		setKey := prefix + "s:" + idStr
		if n, _ := rd.RDB.SCard(rd.Ctx, setKey).Result(); n == 0 {
			rd.RDB.Del(rd.Ctx, setKey)
		}
	}
	return len(items), nil
}

func GetOnlineAdmins() ([]map[string]interface{}, error) {
	_, prefix := tokenutil.GetTokenConfig("admin")
	if rd.RDB == nil {
		return []map[string]interface{}{}, nil
	}
	setPrefix := prefix + "s:"
	authPrefix := prefix + "a:"

	entries, err := scanOnlineSets(setPrefix)
	if err != nil {
		return nil, err
	}
	loadBase := preloadAdminBase(entries)
	return buildOnlineRows(entries, authPrefix, loadBase), nil
}

// preloadAdminBase fetches all admin records and their roles in 2 batched queries
// (instead of 2N queries), and returns a lookup closure.
func preloadAdminBase(entries []onlineEntry) func(uint64) (map[string]interface{}, bool) {
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
			"pic":      GetFullURL(a.Pic),
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
	rd.RDB.Del(rd.Ctx, prefix+"a:"+token)
	rd.RDB.SRem(rd.Ctx, prefix+"s:"+idStr, token)
	if count, _ := rd.RDB.SCard(rd.Ctx, prefix+"s:"+idStr).Result(); count == 0 {
		rd.RDB.Del(rd.Ctx, prefix+"s:"+idStr)
	}
	return nil
}

// BatchForceOfflineAdmin 批量踢管理员（pipeline 一次完成）。
func BatchForceOfflineAdmin(items []struct {
	IDStr string `json:"idStr"`
	Token string `json:"token"`
}) (int, error) {
	_, prefix := tokenutil.GetTokenConfig("admin")
	if rd.RDB == nil || len(items) == 0 {
		return 0, nil
	}
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
		pipe.Del(rd.Ctx, authKeys...)
		pipe.SRem(rd.Ctx, prefix+"s:"+idStr, anyToIface(tokens)...)
	}
	if _, err := pipe.Exec(rd.Ctx); err != nil && err != redis.Nil {
		return 0, err
	}
	for idStr := range byID {
		setKey := prefix + "s:" + idStr
		if n, _ := rd.RDB.SCard(rd.Ctx, setKey).Result(); n == 0 {
			rd.RDB.Del(rd.Ctx, setKey)
		}
	}
	return len(items), nil
}

func AdminLogout(adminID uint, currentToken string) error {
	_, prefix := tokenutil.GetTokenConfig("admin")
	if rd.RDB == nil {
		return nil
	}
	idStr := strconv.Itoa(int(adminID))
	if currentToken == "" {
		// 兜底：无 token 时清空该 adminID 的所有 session
		tokens, _ := rd.RDB.SMembers(rd.Ctx, prefix+"s:"+idStr).Result()
		for _, t := range tokens {
			rd.RDB.Del(rd.Ctx, prefix+"a:"+t)
		}
		rd.RDB.Del(rd.Ctx, prefix+"s:"+idStr)
		return nil
	}
	rd.RDB.Del(rd.Ctx, prefix+"a:"+currentToken)
	rd.RDB.SRem(rd.Ctx, prefix+"s:"+idStr, currentToken)
	if count, _ := rd.RDB.SCard(rd.Ctx, prefix+"s:"+idStr).Result(); count == 0 {
		rd.RDB.Del(rd.Ctx, prefix+"s:"+idStr)
	}
	return nil
}
