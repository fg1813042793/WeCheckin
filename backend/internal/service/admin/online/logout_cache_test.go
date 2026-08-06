package online

import (
	"os"
	"strings"
	"testing"
)

func TestForceOfflineUserContextReturnsRedisCleanupErrors(t *testing.T) {
	src, err := os.ReadFile("user.go")
	if err != nil {
		t.Fatalf("read user.go: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		"func RemoveUserTokenContext(ctx context.Context, token string) error",
		"return rd.RDB.Del(redisCtx, prefix+\"a:\"+token).Err()",
		"if err := rd.RDB.SRem(redisCtx, setKey, token).Err(); err != nil",
		"count, err := rd.RDB.SCard(redisCtx, setKey).Result()",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("online session cleanup should handle Redis errors with %q", snippet)
		}
	}
	if strings.Contains(text, "rd.RDB.Del(redisCtx, prefix+\"a:\"+token)\n") {
		t.Fatalf("online session cleanup must not ignore auth token Redis delete errors")
	}
}
