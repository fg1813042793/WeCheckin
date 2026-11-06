package infrastructure

import (
	"context"
	"errors"
	"testing"
	"time"

	redis "github.com/redis/go-redis/v9"
	exportlinkapp "wecheckin/backend/internal/modules/examrecordexportlink/application"
)

type redisClientStub struct {
	key        string
	value      interface{}
	expiration time.Duration
	getValue   string
	getErr     error
}

func (client *redisClientStub) Set(_ context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	client.key, client.value, client.expiration = key, value, expiration
	return redis.NewStatusResult("OK", nil)
}

func (client *redisClientStub) Get(_ context.Context, key string) *redis.StringCmd {
	client.key = key
	return redis.NewStringResult(client.getValue, client.getErr)
}

func TestRedisStorePersistsAndLoadsGrant(t *testing.T) {
	client := &redisClientStub{}
	store := NewRedisStoreWithClient(client)
	grant := exportlinkapp.Grant{ExamID: 5, RecordID: 19, Format: "pdf", ExpiresAt: time.Date(2026, 11, 10, 0, 0, 0, 0, time.UTC)}
	if err := store.Put(context.Background(), "digest", grant, 7*24*time.Hour); err != nil {
		t.Fatal(err)
	}
	if client.key != "exam_record_export_link:digest" || client.expiration != 7*24*time.Hour {
		t.Fatalf("redis write key=%q ttl=%v", client.key, client.expiration)
	}
	client.getValue = string(client.value.([]byte))
	loaded, err := store.Get(context.Background(), "digest")
	if err != nil || loaded != grant {
		t.Fatalf("loaded=%#v err=%v", loaded, err)
	}
}

func TestRedisStoreMapsMissingGrant(t *testing.T) {
	store := NewRedisStoreWithClient(&redisClientStub{getErr: redis.Nil})
	if _, err := store.Get(context.Background(), "missing"); !errors.Is(err, exportlinkapp.ErrGrantNotFound) {
		t.Fatalf("error = %v", err)
	}
}
