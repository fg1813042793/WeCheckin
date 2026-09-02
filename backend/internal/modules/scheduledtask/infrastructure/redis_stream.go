package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	redis "github.com/redis/go-redis/v9"

	"wecheckin/backend/internal/modules/scheduledtask/application"
)

type RedisCommands interface {
	XGroupCreateMkStream(context.Context, string, string, string) *redis.StatusCmd
	XAdd(context.Context, *redis.XAddArgs) *redis.StringCmd
	XReadGroup(context.Context, *redis.XReadGroupArgs) *redis.XStreamSliceCmd
	XAck(context.Context, string, string, ...string) *redis.IntCmd
	XAutoClaim(context.Context, *redis.XAutoClaimArgs) *redis.XAutoClaimCmd
	Set(context.Context, string, interface{}, time.Duration) *redis.StatusCmd
	Scan(context.Context, uint64, string, int64) *redis.ScanCmd
	Get(context.Context, string) *redis.StringCmd
}

type QueueMessage = application.QueueMessage
type WorkerHeartbeat = application.WorkerHeartbeat

type RedisStreamQueue struct {
	client          RedisCommands
	streamKey       string
	groupName       string
	workerKeyPrefix string
}

func NewRedisStreamQueue(client RedisCommands, keyPrefix string) *RedisStreamQueue {
	keyPrefix = strings.Trim(strings.TrimSpace(keyPrefix), ":")
	if keyPrefix == "" {
		keyPrefix = "wecheckin"
	}
	return &RedisStreamQueue{
		client:          client,
		streamKey:       keyPrefix + ":scheduled-task:runs",
		groupName:       keyPrefix + ":scheduled-task:workers",
		workerKeyPrefix: keyPrefix + ":scheduled-task:worker:",
	}
}

func (queue *RedisStreamQueue) EnsureGroup(ctx context.Context) error {
	if queue == nil || queue.client == nil {
		return errors.New("scheduled task Redis client is not initialized")
	}
	err := queue.client.XGroupCreateMkStream(ctx, queue.streamKey, queue.groupName, "0").Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		return fmt.Errorf("create scheduled task consumer group: %w", err)
	}
	return nil
}

func (queue *RedisStreamQueue) PublishRun(ctx context.Context, runID string) (string, error) {
	if queue == nil || queue.client == nil {
		return "", errors.New("scheduled task Redis client is not initialized")
	}
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return "", errors.New("scheduled task run ID is required")
	}
	messageID, err := queue.client.XAdd(ctx, &redis.XAddArgs{
		Stream: queue.streamKey,
		Values: map[string]interface{}{"run_id": runID},
	}).Result()
	if err != nil {
		return "", fmt.Errorf("publish scheduled task run: %w", err)
	}
	return messageID, nil
}

func (queue *RedisStreamQueue) Read(ctx context.Context, consumer string, count int64, block time.Duration) ([]QueueMessage, error) {
	streams, err := queue.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group: queue.groupName, Consumer: consumer,
		Streams: []string{queue.streamKey, ">"}, Count: count, Block: block,
	}).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read scheduled task stream: %w", err)
	}
	return decodeQueueMessages(streams), nil
}

func (queue *RedisStreamQueue) Ack(ctx context.Context, messageIDs ...string) error {
	if len(messageIDs) == 0 {
		return nil
	}
	if err := queue.client.XAck(ctx, queue.streamKey, queue.groupName, messageIDs...).Err(); err != nil {
		return fmt.Errorf("ack scheduled task stream: %w", err)
	}
	return nil
}

func (queue *RedisStreamQueue) AutoClaim(ctx context.Context, consumer string, minIdle time.Duration, start string, count int64) ([]QueueMessage, string, error) {
	messages, next, err := queue.client.XAutoClaim(ctx, &redis.XAutoClaimArgs{
		Stream: queue.streamKey, Group: queue.groupName, Consumer: consumer,
		MinIdle: minIdle, Start: start, Count: count,
	}).Result()
	if errors.Is(err, redis.Nil) {
		return nil, start, nil
	}
	if err != nil {
		return nil, start, fmt.Errorf("claim scheduled task stream: %w", err)
	}
	return decodeMessages(messages), next, nil
}

func (queue *RedisStreamQueue) HeartbeatWorker(ctx context.Context, heartbeat WorkerHeartbeat, ttl time.Duration) error {
	heartbeat.WorkerID = strings.TrimSpace(heartbeat.WorkerID)
	if heartbeat.WorkerID == "" || strings.ContainsAny(heartbeat.WorkerID, "*?[] \t\r\n") {
		return errors.New("invalid scheduled task worker ID")
	}
	payload, err := json.Marshal(heartbeat)
	if err != nil {
		return err
	}
	if err := queue.client.Set(ctx, queue.workerKeyPrefix+heartbeat.WorkerID, string(payload), ttl).Err(); err != nil {
		return fmt.Errorf("write scheduled task worker heartbeat: %w", err)
	}
	return nil
}

func (queue *RedisStreamQueue) ListWorkers(ctx context.Context) ([]WorkerHeartbeat, error) {
	workers := make([]WorkerHeartbeat, 0)
	cursor := uint64(0)
	for {
		keys, next, err := queue.client.Scan(ctx, cursor, queue.workerKeyPrefix+"*", 100).Result()
		if err != nil {
			return nil, fmt.Errorf("scan scheduled task workers: %w", err)
		}
		for _, key := range keys {
			payload, err := queue.client.Get(ctx, key).Result()
			if errors.Is(err, redis.Nil) {
				continue
			}
			if err != nil {
				return nil, fmt.Errorf("read scheduled task worker %s: %w", key, err)
			}
			var heartbeat WorkerHeartbeat
			if err := json.Unmarshal([]byte(payload), &heartbeat); err != nil {
				return nil, fmt.Errorf("decode scheduled task worker %s: %w", key, err)
			}
			workers = append(workers, heartbeat)
		}
		if next == 0 {
			break
		}
		cursor = next
	}
	sort.Slice(workers, func(i, j int) bool { return workers[i].WorkerID < workers[j].WorkerID })
	return workers, nil
}

func decodeQueueMessages(streams []redis.XStream) []QueueMessage {
	result := make([]QueueMessage, 0)
	for _, stream := range streams {
		result = append(result, decodeMessages(stream.Messages)...)
	}
	return result
}

func decodeMessages(messages []redis.XMessage) []QueueMessage {
	result := make([]QueueMessage, 0, len(messages))
	for _, message := range messages {
		raw, ok := message.Values["run_id"]
		if !ok {
			continue
		}
		runID := strings.TrimSpace(fmt.Sprint(raw))
		if runID == "" {
			continue
		}
		result = append(result, QueueMessage{MessageID: message.ID, RunID: runID})
	}
	return result
}

var _ application.QueuePublisher = (*RedisStreamQueue)(nil)
