package redis

import (
	"context"
	"fmt"

	goredis "github.com/redis/go-redis/v9"

	"wecheckin/backend/internal/config"
)

var RDB *goredis.Client
var Ctx = context.Background()

func Init(cfg config.RedisConfig) error {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	client := goredis.NewClient(&goredis.Options{
		Addr:     addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	if err := client.Ping(Ctx).Err(); err != nil {
		_ = client.Close()
		return err
	}
	if RDB != nil {
		_ = RDB.Close()
	}
	RDB = client
	return nil
}

func Close() {
	if RDB != nil {
		RDB.Close()
	}
}
