package redis

import (
	"context"
	"fmt"

	goredis "github.com/redis/go-redis/v9"

	"wecheckin-backend/backend/internal/config"
)

var RDB *goredis.Client
var Ctx = context.Background()

func Init(cfg config.RedisConfig) error {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	RDB = goredis.NewClient(&goredis.Options{
		Addr:     addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	return RDB.Ping(Ctx).Err()
}

func Close() {
	if RDB != nil {
		RDB.Close()
	}
}
