package passwordsession

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	rd "wecheckin/backend/pkg/redis"
	"wecheckin/backend/pkg/tokenutil"
)

func audienceRole(audience Audience) string {
	switch audience {
	case AudienceAdmin:
		return "admin"
	case AudienceDingTalkH5:
		return "dingtalk_h5"
	default:
		return "user"
	}
}

func RefreshCurrentContext(ctx context.Context, audience Audience, userID uint, token string, changedAt int64) error {
	token = strings.TrimSpace(strings.TrimPrefix(token, "Bearer "))
	if userID == 0 || token == "" || rd.RDB == nil {
		return fmt.Errorf("登录会话无效")
	}
	role := audienceRole(audience)
	expire, prefix := tokenutil.GetTokenConfigContext(ctx, role)
	authKey := prefix + "a:" + token
	setKey := prefix + "s:" + strconv.FormatUint(uint64(userID), 10)
	redisCtx, cancel := rd.OperationContext(ctx)
	defer cancel()
	raw, err := rd.RDB.Get(redisCtx, authKey).Result()
	if err != nil {
		return fmt.Errorf("登录已过期或已被强制下线")
	}
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return fmt.Errorf("登录信息异常")
	}
	payload["passwordChangedAt"] = changedAt
	payload["passwordMustChange"] = false
	encoded, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	tokens, err := rd.RDB.SMembers(redisCtx, setKey).Result()
	if err != nil {
		return err
	}
	for _, oldToken := range tokens {
		if oldToken != token {
			if err := rd.RDB.Del(redisCtx, prefix+"a:"+oldToken).Err(); err != nil {
				return err
			}
		}
	}
	pipe := rd.RDB.TxPipeline()
	pipe.Set(redisCtx, authKey, encoded, expire)
	pipe.Del(redisCtx, setKey)
	pipe.SAdd(redisCtx, setKey, token)
	pipe.Expire(redisCtx, setKey, expire*2)
	_, err = pipe.Exec(redisCtx)
	return err
}

func InvalidateAllContext(ctx context.Context, userID uint) error {
	if userID == 0 || rd.RDB == nil {
		return nil
	}
	redisCtx, cancel := rd.OperationContext(ctx)
	defer cancel()
	for _, audience := range []Audience{AudienceAdmin, AudienceClient, AudienceDingTalkH5} {
		_, prefix := tokenutil.GetTokenConfigContext(ctx, audienceRole(audience))
		setKey := prefix + "s:" + strconv.FormatUint(uint64(userID), 10)
		tokens, err := rd.RDB.SMembers(redisCtx, setKey).Result()
		if err != nil {
			return err
		}
		pipe := rd.RDB.TxPipeline()
		for _, token := range tokens {
			pipe.Del(redisCtx, prefix+"a:"+token)
		}
		pipe.Del(redisCtx, setKey)
		if _, err := pipe.Exec(redisCtx); err != nil {
			return err
		}
	}
	return nil
}
