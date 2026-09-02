package scheduledtaskmodel

import "time"

const (
	HandlerTypeGo       = "go"
	HandlerTypeWorkflow = "workflow"
	HandlerTypeHTTP     = "http"
	HandlerTypeShell    = "shell"
	HandlerTypeSQL      = "sql"

	CronPrecisionMinute = "minute"
	CronPrecisionSecond = "second"

	MisfirePolicySkip      = "skip"
	MisfirePolicyFireOnce  = "fire_once"
	MisfirePolicyCatchUp   = "catch_up"
	ConcurrencyPolicySkip  = "skip"
	ConcurrencyPolicyQueue = "queue_once"
	ConcurrencyPolicyAllow = "allow"

	TriggerTypeScheduled   = "scheduled"
	TriggerTypeManual      = "manual"
	TriggerTypeManualRetry = "manual_retry"
	TriggerTypeMisfire     = "misfire"

	RunStatusWaiting   = "waiting"
	RunStatusQueued    = "queued"
	RunStatusRunning   = "running"
	RunStatusRetryWait = "retry_wait"
	RunStatusSuccess   = "success"
	RunStatusFailed    = "failed"
	RunStatusCanceled  = "canceled"
	RunStatusSkipped   = "skipped"
)

type Task struct {
	ID                uint64    `json:"id" gorm:"primaryKey;column:id;comment:Task ID"`
	Code              string    `json:"code" gorm:"size:100;column:code;uniqueIndex:uk_scheduled_task_code;comment:Stable task code"`
	Name              string    `json:"name" gorm:"size:200;column:name;comment:Task name"`
	Description       string    `json:"description" gorm:"size:500;column:description;comment:Task description"`
	HandlerType       string    `json:"handlerType" gorm:"size:24;column:handler_type;index:idx_scheduled_tasks_handler,priority:1;comment:Handler type"`
	HandlerConfigJSON string    `json:"handlerConfigJson" gorm:"type:mediumtext;column:handler_config_json;comment:Handler configuration JSON"`
	CronExpression    string    `json:"cronExpression" gorm:"size:120;column:cron_expression;comment:Cron expression"`
	CronPrecision     string    `json:"cronPrecision" gorm:"size:16;column:cron_precision;comment:Cron precision"`
	Timezone          string    `json:"timezone" gorm:"size:64;column:timezone;comment:IANA timezone"`
	Enabled           int       `json:"enabled" gorm:"column:enabled;index:idx_scheduled_tasks_due,priority:1;index:idx_scheduled_tasks_handler,priority:2;comment:Enabled status"`
	MisfirePolicy     string    `json:"misfirePolicy" gorm:"size:24;column:misfire_policy;comment:Misfire policy"`
	MaxCatchUp        int       `json:"maxCatchUp" gorm:"column:max_catch_up;comment:Maximum catch-up runs"`
	ConcurrencyPolicy string    `json:"concurrencyPolicy" gorm:"size:24;column:concurrency_policy;comment:Concurrency policy"`
	TimeoutSeconds    int       `json:"timeoutSeconds" gorm:"column:timeout_seconds;comment:Execution timeout seconds"`
	MaxRetries        int       `json:"maxRetries" gorm:"column:max_retries;comment:Maximum automatic retries"`
	RetryBackoffJSON  string    `json:"retryBackoffJson" gorm:"type:mediumtext;column:retry_backoff_json;comment:Retry backoff configuration JSON"`
	LastScheduledAt   int64     `json:"lastScheduledAt" gorm:"column:last_scheduled_at;comment:Last processed schedule"`
	NextRunAt         int64     `json:"nextRunAt" gorm:"column:next_run_at;index:idx_scheduled_tasks_due,priority:2;comment:Next schedule"`
	Version           int64     `json:"version" gorm:"column:version;comment:Optimistic lock version"`
	CreatedBy         uint64    `json:"createdBy" gorm:"column:created_by;comment:Creator admin user ID"`
	UpdatedBy         uint64    `json:"updatedBy" gorm:"column:updated_by;comment:Updater admin user ID"`
	DeletedAt         int64     `json:"deletedAt" gorm:"column:deleted_at;index:idx_scheduled_tasks_deleted;comment:Soft deletion time"`
	AddTime           int64     `json:"addTime" gorm:"column:add_time;comment:Creation time"`
	EditTime          int64     `json:"editTime" gorm:"column:edit_time;comment:Update time"`
	CreatedAt         time.Time `json:"-"`
	UpdatedAt         time.Time `json:"-"`
}

func (Task) TableName() string { return "scheduled_tasks" }

type Run struct {
	ID                string    `json:"id" gorm:"size:64;primaryKey;column:id;comment:Run ID"`
	RunKey            string    `json:"runKey" gorm:"size:191;column:run_key;uniqueIndex:uk_scheduled_task_run_key;comment:Idempotency key"`
	TaskID            uint64    `json:"taskId" gorm:"column:task_id;index:idx_scheduled_task_runs_task_time,priority:1;comment:Task ID"`
	ParentRunID       string    `json:"parentRunId" gorm:"size:64;column:parent_run_id;index:idx_scheduled_task_runs_parent;comment:Manual retry source"`
	TriggerType       string    `json:"triggerType" gorm:"size:24;column:trigger_type;comment:Trigger type"`
	Status            string    `json:"status" gorm:"size:24;column:run_status;index:idx_scheduled_task_runs_due,priority:1;index:idx_scheduled_task_runs_recovery,priority:1;comment:Run status"`
	TaskSnapshotJSON  string    `json:"taskSnapshotJson" gorm:"type:mediumtext;column:task_snapshot_json;comment:Task snapshot JSON"`
	ScheduledAt       int64     `json:"scheduledAt" gorm:"column:scheduled_at;index:idx_scheduled_task_runs_task_time,priority:2;comment:Scheduled UTC time"`
	CoalescedCount    int       `json:"coalescedCount" gorm:"column:coalesced_count;comment:Coalesced trigger count"`
	Attempt           int       `json:"attempt" gorm:"column:attempt;comment:Current attempt"`
	WorkerID          string    `json:"workerId" gorm:"size:160;column:worker_id;comment:Worker ID"`
	RedisMessageID    string    `json:"redisMessageId" gorm:"size:80;column:redis_message_id;comment:Redis message ID"`
	QueuedAt          int64     `json:"queuedAt" gorm:"column:queued_at;comment:Queue time"`
	StartedAt         int64     `json:"startedAt" gorm:"column:started_at;comment:Start time"`
	FinishedAt        int64     `json:"finishedAt" gorm:"column:finished_at;comment:Finish time"`
	HeartbeatAt       int64     `json:"heartbeatAt" gorm:"column:heartbeat_at;index:idx_scheduled_task_runs_recovery,priority:2;comment:Worker heartbeat"`
	NextRetryAt       int64     `json:"nextRetryAt" gorm:"column:next_retry_at;index:idx_scheduled_task_runs_due,priority:2;comment:Next retry time"`
	CancelRequestedAt int64     `json:"cancelRequestedAt" gorm:"column:cancel_requested_at;comment:Cancellation request time"`
	ResultSummary     string    `json:"resultSummary" gorm:"size:2000;column:result_summary;comment:Redacted result summary"`
	ErrorCode         string    `json:"errorCode" gorm:"size:64;column:error_code;comment:Classified error code"`
	ErrorSummary      string    `json:"errorSummary" gorm:"size:2000;column:error_summary;comment:Redacted error summary"`
	AddTime           int64     `json:"addTime" gorm:"column:add_time;comment:Creation time"`
	EditTime          int64     `json:"editTime" gorm:"column:edit_time;comment:Update time"`
	CreatedAt         time.Time `json:"-"`
	UpdatedAt         time.Time `json:"-"`
}

func (Run) TableName() string { return "scheduled_task_runs" }

type RunLog struct {
	ID        uint64    `json:"id" gorm:"primaryKey;column:id;comment:Log segment ID"`
	RunID     string    `json:"runId" gorm:"size:64;column:run_id;index:idx_scheduled_task_run_logs_run_seq,priority:1;comment:Run ID"`
	Sequence  int       `json:"sequence" gorm:"column:log_sequence;index:idx_scheduled_task_run_logs_run_seq,priority:2;comment:Log sequence"`
	Level     string    `json:"level" gorm:"size:16;column:log_level;comment:Log level"`
	Stage     string    `json:"stage" gorm:"size:40;column:log_stage;comment:Execution stage"`
	Content   string    `json:"content" gorm:"type:mediumtext;column:log_content;comment:Redacted log content"`
	LogTime   int64     `json:"logTime" gorm:"column:log_time;comment:Log time"`
	AddTime   int64     `json:"addTime" gorm:"column:add_time;comment:Creation time"`
	CreatedAt time.Time `json:"-"`
}

func (RunLog) TableName() string { return "scheduled_task_run_logs" }
