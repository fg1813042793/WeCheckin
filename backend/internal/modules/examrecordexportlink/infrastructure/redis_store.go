package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	redis "github.com/redis/go-redis/v9"
	exportlinkapp "wecheckin/backend/internal/modules/examrecordexportlink/application"
)

const redisGrantPrefix = "exam_record_export_link:"

type RedisClient interface {
	Set(context.Context, string, interface{}, time.Duration) *redis.StatusCmd
	Get(context.Context, string) *redis.StringCmd
}

type RedisStore struct {
	client RedisClient
}

func NewRedisStore(client *redis.Client) *RedisStore {
	if client == nil {
		return &RedisStore{}
	}
	return &RedisStore{client: client}
}

func NewRedisStoreWithClient(client RedisClient) *RedisStore { return &RedisStore{client: client} }

func (store *RedisStore) Put(ctx context.Context, digest string, grant exportlinkapp.Grant, ttl time.Duration) error {
	if store == nil || store.client == nil {
		return exportlinkapp.ErrStoreMissing
	}
	payload, err := json.Marshal(grant)
	if err != nil {
		return err
	}
	return store.client.Set(ctx, redisGrantPrefix+digest, payload, ttl).Err()
}

func (store *RedisStore) Get(ctx context.Context, digest string) (exportlinkapp.Grant, error) {
	if store == nil || store.client == nil {
		return exportlinkapp.Grant{}, exportlinkapp.ErrStoreMissing
	}
	raw, err := store.client.Get(ctx, redisGrantPrefix+digest).Result()
	if errors.Is(err, redis.Nil) {
		return exportlinkapp.Grant{}, exportlinkapp.ErrGrantNotFound
	}
	if err != nil {
		return exportlinkapp.Grant{}, err
	}
	var grant exportlinkapp.Grant
	if json.Unmarshal([]byte(raw), &grant) != nil {
		return exportlinkapp.Grant{}, exportlinkapp.ErrGrantNotFound
	}
	return grant, nil
}
