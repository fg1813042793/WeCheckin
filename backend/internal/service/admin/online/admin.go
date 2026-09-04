package online

import (
	"context"
	"strconv"

	"github.com/redis/go-redis/v9"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/support/adminaccess"
	"wecheckin/backend/internal/support/media"
	"wecheckin/backend/pkg/database"
	rd "wecheckin/backend/pkg/redis"
	"wecheckin/backend/pkg/tokenutil"
)

func GetOnlineAdmins() ([]AdminSession, error) {
	return GetOnlineAdminsContext(context.Background())
}

func GetOnlineAdminsContext(ctx context.Context) ([]AdminSession, error) {
	_, prefix := tokenutil.GetTokenConfigContext(ctx, "admin")
	if rd.RDB == nil {
		return []AdminSession{}, nil
	}
	setPrefix := prefix + "s:"
	authPrefix := prefix + "a:"

	entries, err := scanSets(ctx, setPrefix)
	if err != nil {
		return nil, err
	}
	loadBase, err := preloadAdminBase(ctx, entries)
	if err != nil {
		return nil, err
	}
	return buildRows(ctx, entries, authPrefix, loadBase, func(row AdminSession, info SessionInfo) AdminSession {
		row.SessionInfo = info
		return row
	}), nil
}

func preloadAdminBase(ctx context.Context, entries []entry) (func(uint64) (AdminSession, bool), error) {
	uids := make([]uint, 0, len(entries))
	for _, e := range entries {
		uids = append(uids, uint(e.uid))
	}
	db, cancel := database.WithContext(ctx)
	defer cancel()
	var admins []model.Admin
	if len(uids) > 0 {
		if err := adminaccess.ApplyUserAdminAccessRoleFilter(db.Where("id IN ?", uids)).
			Find(&admins).Error; err != nil {
			return nil, err
		}
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
		if err := db.Where("id IN ?", roleIDs).Find(&roles).Error; err != nil {
			return nil, err
		}
	}
	roleByID := make(map[uint]string, len(roles))
	for _, r := range roles {
		roleByID[r.ID] = r.Name
	}

	return func(uid uint64) (AdminSession, bool) {
		a, ok := adminByID[uint(uid)]
		if !ok {
			return AdminSession{}, false
		}
		roleName := ""
		if a.RoleID > 0 {
			roleName = roleByID[a.RoleID]
		}
		return AdminSession{
			ID:       a.ID,
			Name:     a.Name,
			Desc:     a.Desc,
			Pic:      media.FullURLWithStaticDomain(a.Pic),
			Type:     a.Type,
			RoleName: roleName,
			LoginCnt: a.LoginCnt,
		}, true
	}, nil
}

func ForceOfflineAdmin(idStr, token string) error {
	return ForceOfflineAdminContext(context.Background(), idStr, token)
}

func ForceOfflineAdminContext(ctx context.Context, idStr, token string) error {
	_, prefix := tokenutil.GetTokenConfigContext(ctx, "admin")
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

func BatchForceOfflineAdmin(items []struct {
	IDStr string `json:"idStr"`
	Token string `json:"token"`
}) (int, error) {
	return BatchForceOfflineAdminContext(context.Background(), items)
}

func BatchForceOfflineAdminContext(ctx context.Context, items []struct {
	IDStr string `json:"idStr"`
	Token string `json:"token"`
}) (int, error) {
	_, prefix := tokenutil.GetTokenConfigContext(ctx, "admin")
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

func AdminLogout(adminID uint, currentToken string) error {
	return AdminLogoutContext(context.Background(), adminID, currentToken)
}

func AdminLogoutContext(ctx context.Context, adminID uint, currentToken string) error {
	_, prefix := tokenutil.GetTokenConfigContext(ctx, "admin")
	if rd.RDB == nil {
		return nil
	}
	redisCtx, cancel := rd.OperationContext(ctx)
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
