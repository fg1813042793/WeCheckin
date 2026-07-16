package online

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"

	rd "wecheckin-backend/backend/pkg/redis"
)

type entry struct {
	setKey string
	uid    uint64
	tokens []string
}

func scanSets(setPrefix string) ([]entry, error) {
	redisCtx, cancel := rd.OperationContext(context.Background())
	defer cancel()
	var cursor uint64
	var setKeys []string
	for {
		ks, c, err := rd.RDB.Scan(redisCtx, cursor, setPrefix+"*", 100).Result()
		if err != nil {
			return nil, err
		}
		setKeys = append(setKeys, ks...)
		cursor = c
		if cursor == 0 {
			break
		}
	}

	entries := make([]entry, 0, len(setKeys))
	for _, setKey := range setKeys {
		idStr := strings.TrimPrefix(setKey, setPrefix)
		uid, _ := strconv.ParseUint(idStr, 10, 64)
		if uid == 0 {
			continue
		}
		tokens, _ := rd.RDB.SMembers(redisCtx, setKey).Result()
		if len(tokens) == 0 {
			continue
		}
		entries = append(entries, entry{setKey, uid, tokens})
	}
	return entries, nil
}

func buildRows(entries []entry, authPrefix string, loadBase func(uid uint64) (map[string]interface{}, bool)) []map[string]interface{} {
	redisCtx, cancel := rd.OperationContext(context.Background())
	defer cancel()
	pipe := rd.RDB.Pipeline()
	type tokenCmd struct {
		token string
		get   *redis.StringCmd
		ttl   *redis.DurationCmd
	}
	allCmds := make([]tokenCmd, 0)
	for _, e := range entries {
		for _, t := range e.tokens {
			allCmds = append(allCmds, tokenCmd{
				token: t,
				get:   pipe.Get(redisCtx, authPrefix+t),
				ttl:   pipe.TTL(redisCtx, authPrefix+t),
			})
		}
	}
	if len(allCmds) > 0 {
		_, _ = pipe.Exec(redisCtx)
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
			rd.RDB.SRem(redisCtx, e.setKey, stringSliceToInterface(deadTokens)...)
		}
	}
	return result
}

func stringSliceToInterface(ss []string) []interface{} {
	out := make([]interface{}, len(ss))
	for i, s := range ss {
		out[i] = s
	}
	return out
}
