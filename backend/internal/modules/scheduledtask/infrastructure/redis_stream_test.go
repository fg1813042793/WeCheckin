package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	redis "github.com/redis/go-redis/v9"
)

func TestRedisStreamQueueUsesPrefixedKeysAndRunIDOnlyMessages(t *testing.T) {
	client := &fakeRedisCommands{}
	queue := NewRedisStreamQueue(client, "test-env")

	if err := queue.EnsureGroup(context.Background()); err != nil {
		t.Fatal(err)
	}
	messageID, err := queue.PublishRun(context.Background(), "run-42")
	if err != nil {
		t.Fatal(err)
	}
	if messageID != "100-0" {
		t.Fatalf("message id = %q", messageID)
	}
	if client.groupStream != "test-env:scheduled-task:runs" || client.group != "test-env:scheduled-task:workers" {
		t.Fatalf("group creation = %q/%q", client.groupStream, client.group)
	}
	if client.addArgs.Stream != client.groupStream || len(client.addArgs.Values.(map[string]interface{})) != 1 || client.addArgs.Values.(map[string]interface{})["run_id"] != "run-42" {
		t.Fatalf("xadd args = %#v", client.addArgs)
	}
}

func TestRedisStreamQueueReadsAcknowledgesAndAutoClaims(t *testing.T) {
	client := &fakeRedisCommands{
		readStreams: []redis.XStream{{Stream: "test-env:scheduled-task:runs", Messages: []redis.XMessage{{ID: "101-0", Values: map[string]interface{}{"run_id": "run-1"}}}}},
		claimed:     []redis.XMessage{{ID: "99-0", Values: map[string]interface{}{"run_id": "run-old"}}},
	}
	queue := NewRedisStreamQueue(client, "test-env")

	messages, err := queue.Read(context.Background(), "worker-a", 10, 2*time.Second)
	if err != nil || len(messages) != 1 || messages[0].RunID != "run-1" {
		t.Fatalf("read messages = %#v, err = %v", messages, err)
	}
	if client.readArgs.Group != "test-env:scheduled-task:workers" || client.readArgs.Consumer != "worker-a" {
		t.Fatalf("read args = %#v", client.readArgs)
	}
	if err := queue.Ack(context.Background(), "101-0"); err != nil {
		t.Fatal(err)
	}
	claimed, next, err := queue.AutoClaim(context.Background(), "worker-a", time.Minute, "0-0", 10)
	if err != nil || len(claimed) != 1 || claimed[0].RunID != "run-old" || next != "200-0" {
		t.Fatalf("claimed = %#v, next = %q, err = %v", claimed, next, err)
	}
	if client.ackIDs[0] != "101-0" || client.claimArgs.MinIdle != time.Minute {
		t.Fatalf("ack/claim = %#v / %#v", client.ackIDs, client.claimArgs)
	}
}

func TestRedisStreamQueueTreatsExistingConsumerGroupAsSuccess(t *testing.T) {
	client := &fakeRedisCommands{groupErr: errors.New("BUSYGROUP Consumer Group name already exists")}
	if err := NewRedisStreamQueue(client, "test-env").EnsureGroup(context.Background()); err != nil {
		t.Fatalf("ensure existing group: %v", err)
	}
}

func TestWorkerHeartbeatUsesTTLAndCanBeListed(t *testing.T) {
	client := &fakeRedisCommands{}
	queue := NewRedisStreamQueue(client, "test-env")
	heartbeat := WorkerHeartbeat{WorkerID: "node-a", Role: "all", Version: "v1", StartedAt: 10, LastHeartbeat: 20, CurrentRuns: 2, WorkerCount: 4}
	if err := queue.HeartbeatWorker(context.Background(), heartbeat, 30*time.Second); err != nil {
		t.Fatal(err)
	}
	if client.setKey != "test-env:scheduled-task:worker:node-a" || client.setExpiration != 30*time.Second {
		t.Fatalf("heartbeat set = %q / %s", client.setKey, client.setExpiration)
	}
	client.scanKeys = []string{client.setKey}
	client.getValues = map[string]string{client.setKey: client.setValue.(string)}
	workers, err := queue.ListWorkers(context.Background())
	if err != nil || len(workers) != 1 || workers[0].WorkerID != "node-a" {
		t.Fatalf("workers = %#v, err = %v", workers, err)
	}
	var encoded WorkerHeartbeat
	if err := json.Unmarshal([]byte(client.setValue.(string)), &encoded); err != nil || encoded.CurrentRuns != 2 {
		t.Fatalf("heartbeat JSON = %#v, err = %v", encoded, err)
	}
}

type fakeRedisCommands struct {
	groupStream   string
	group         string
	groupErr      error
	addArgs       *redis.XAddArgs
	readArgs      *redis.XReadGroupArgs
	readStreams   []redis.XStream
	ackIDs        []string
	claimArgs     *redis.XAutoClaimArgs
	claimed       []redis.XMessage
	setKey        string
	setValue      interface{}
	setExpiration time.Duration
	scanKeys      []string
	getValues     map[string]string
}

func (client *fakeRedisCommands) XGroupCreateMkStream(_ context.Context, stream, group, _ string) *redis.StatusCmd {
	client.groupStream, client.group = stream, group
	return redis.NewStatusResult("OK", client.groupErr)
}

func (client *fakeRedisCommands) XAdd(_ context.Context, args *redis.XAddArgs) *redis.StringCmd {
	client.addArgs = args
	return redis.NewStringResult("100-0", nil)
}

func (client *fakeRedisCommands) XReadGroup(_ context.Context, args *redis.XReadGroupArgs) *redis.XStreamSliceCmd {
	client.readArgs = args
	return redis.NewXStreamSliceCmdResult(client.readStreams, nil)
}

func (client *fakeRedisCommands) XAck(_ context.Context, _, _ string, ids ...string) *redis.IntCmd {
	client.ackIDs = append([]string(nil), ids...)
	return redis.NewIntResult(int64(len(ids)), nil)
}

func (client *fakeRedisCommands) XAutoClaim(ctx context.Context, args *redis.XAutoClaimArgs) *redis.XAutoClaimCmd {
	client.claimArgs = args
	cmd := redis.NewXAutoClaimCmd(ctx)
	cmd.SetVal(client.claimed, "200-0")
	return cmd
}

func (client *fakeRedisCommands) Set(_ context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	client.setKey, client.setValue, client.setExpiration = key, value, expiration
	return redis.NewStatusResult("OK", nil)
}

func (client *fakeRedisCommands) Scan(_ context.Context, _ uint64, _ string, _ int64) *redis.ScanCmd {
	return redis.NewScanCmdResult(client.scanKeys, 0, nil)
}

func (client *fakeRedisCommands) Get(_ context.Context, key string) *redis.StringCmd {
	value, ok := client.getValues[key]
	if !ok {
		return redis.NewStringResult("", redis.Nil)
	}
	return redis.NewStringResult(value, nil)
}
