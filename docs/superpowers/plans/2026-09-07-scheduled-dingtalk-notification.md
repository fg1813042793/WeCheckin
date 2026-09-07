# Scheduled DingTalk Notification Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在定时任务的 Go 注册任务列表中提供独立的“发送钉钉通知”，并复用现有企业内部钉钉通知链路。

**Architecture:** 新任务实现 `scheduledtask/infrastructure.GoJob`，将与站内信相同的收件参数转换成 `inappnotification/application.SendInput` 后调用现有 `SendDingTalk`。Admin API 与 `taskd` 注册同一个任务；管理端复用通知收件表单，并按任务渠道分别执行权限过滤和校验。

**Tech Stack:** Go 1.24、GORM、Vue 3、TypeScript、Element Plus、Node 静态契约脚本

---

### Task 1: 新增钉钉通知 Go Job

**Files:**
- Create: `backend/internal/modules/scheduledtask/infrastructure/dingtalk_notification_job.go`
- Create: `backend/internal/modules/scheduledtask/infrastructure/dingtalk_notification_job_test.go`
- Reuse: `backend/internal/modules/scheduledtask/infrastructure/in_app_notification_job.go`

- [ ] **Step 1: Write the failing metadata and execution tests**

新增 fake sender，并断言任务键、中文名称、来源 ID、发送统计和日志脱敏：

```go
func TestDingTalkNotificationJobSendsWithRunIDAndCounts(t *testing.T) {
	sender := &dingTalkNotificationSenderStub{result: inappnotificationapp.SendResult{
		SourceID: "run-1", PlannedCount: 3, SentCount: 2, SkippedCount: 1, FailedCount: 1,
	}}
	job := NewDingTalkNotificationJob(sender)
	result, err := job.Execute(context.Background(), "run-1", json.RawMessage(
		`{"title":"定时钉钉通知","content":"敏感正文","scope":"users","userIds":[4,9]}`,
	), &captureRunLogger{})
	if err != nil { t.Fatal(err) }
	if job.Key() != "notification.dingtalk.send" || job.Name() != "发送钉钉通知" { t.Fatal("unexpected metadata") }
	if sender.input.SourceType != inappnotificationapp.SourceScheduledTaskRun || sender.input.SourceID != "run-1" { t.Fatal("unexpected source") }
	if result.Data["failedCount"] != 1 { t.Fatalf("result = %#v", result) }
}
```

- [ ] **Step 2: Run test and verify RED**

```bash
cd backend && GOCACHE=$PWD/../.cache/go-build go test ./internal/modules/scheduledtask/infrastructure -run DingTalkNotificationJob -count=1
```

Expected: FAIL because `NewDingTalkNotificationJob` does not exist.

- [ ] **Step 3: Implement the minimal job**

实现独立 sender 接口、任务元数据、与站内信一致的 JSON Schema、参数解析、错误映射和只包含计数的日志：

```go
const DingTalkNotificationJobKey = "notification.dingtalk.send"

type DingTalkNotificationSender interface {
	SendDingTalk(context.Context, inappnotificationapp.SendInput) (inappnotificationapp.SendResult, error)
}

func (job *DingTalkNotificationJob) Execute(ctx context.Context, runID string, raw json.RawMessage, logger scheduledtaskapp.RunLogger) (scheduledtaskapp.HandlerResult, error) {
	input, err := decodeInAppNotificationParams(raw, runID)
	if err != nil {
		return scheduledtaskapp.HandlerResult{}, &scheduledtaskapp.HandlerError{Code: "invalid_config", Summary: err.Error()}
	}
	result, err := job.sender.SendDingTalk(ctx, input)
	if err != nil {
		if errors.Is(err, inappnotificationapp.ErrNoRecipients) {
			return scheduledtaskapp.HandlerResult{}, &scheduledtaskapp.HandlerError{Code: "no_recipients", Summary: err.Error()}
		}
		return scheduledtaskapp.HandlerResult{}, &scheduledtaskapp.HandlerError{Code: "notification_delivery_failed", Summary: err.Error(), Temporary: true}
	}
	if logger == nil {
		logger = scheduledtaskapp.NopRunLogger{}
	}
	_ = logger.Log(ctx, "info", "delivery", fmt.Sprintf(
		"planned=%d sent=%d skipped=%d failed=%d",
		result.PlannedCount, result.SentCount, result.SkippedCount, result.FailedCount,
	))
	return scheduledtaskapp.HandlerResult{
		Summary: fmt.Sprintf("sent %d DingTalk notifications", result.SentCount),
		Data: map[string]interface{}{
			"sourceId": result.SourceID, "plannedCount": result.PlannedCount,
			"sentCount": result.SentCount, "skippedCount": result.SkippedCount,
			"failedCount": result.FailedCount,
		},
	}, nil
}
```

- [ ] **Step 4: Run the focused test and verify GREEN**

Run the command from Step 2. Expected: PASS.

### Task 2: 增加高风险权限映射

**Files:**
- Modify: `backend/internal/modules/scheduledtask/transport/httpadmin/notification_risk_test.go`
- Modify: `backend/internal/modules/scheduledtask/transport/httpadmin/handler.go`

- [ ] **Step 1: Write the failing permission test**

```go
func TestRiskPermissionRequiresDingTalkSendForDingTalkNotificationJob(t *testing.T) {
	permission := riskPermission(scheduledtaskmodel.HandlerTypeGo,
		json.RawMessage(`{"handlerKey":"notification.dingtalk.send","params":{}}`))
	if permission != "notification:dingtalk:send" { t.Fatalf("riskPermission() = %q", permission) }
}
```

- [ ] **Step 2: Run test and verify RED**

```bash
cd backend && GOCACHE=$PWD/../.cache/go-build go test ./internal/modules/scheduledtask/transport/httpadmin -run RiskPermission -count=1
```

Expected: FAIL because the new job has no permission mapping.

- [ ] **Step 3: Add key-to-permission mapping**

```go
switch strings.TrimSpace(value.HandlerKey) {
case scheduledtaskinfra.InAppNotificationJobKey:
	return "notification:send"
case scheduledtaskinfra.DingTalkNotificationJobKey:
	return "notification:dingtalk:send"
}
```

- [ ] **Step 4: Run the permission tests and verify GREEN**

Run the command from Step 2. Expected: PASS.

### Task 3: 在 Admin 与 taskd 注册同一任务

**Files:**
- Modify: `backend/internal/routes/v2/admin/routes.go`
- Modify: `backend/internal/routes/v2/admin/in_app_notification_routes_test.go`
- Modify: `backend/cmd/taskd/main.go`
- Modify: `backend/cmd/taskd/main_test.go`

- [ ] **Step 1: Add failing wiring contract assertions**

Admin 路由装配测试要求 `scheduledtaskinfra.NewDingTalkNotificationJob(notificationService)`。`taskd` 装配测试要求以下片段：

```go
inappnotificationapp.NewServiceWithDingTalk(
inappnotificationinfra.NewDingTalkDelivery(db, nil)
scheduledtaskinfra.NewDingTalkNotificationJob(notificationService)
```

- [ ] **Step 2: Run wiring tests and verify RED**

```bash
cd backend && GOCACHE=$PWD/../.cache/go-build go test ./internal/routes/v2/admin ./cmd/taskd -count=1
```

Expected: FAIL on missing wiring snippets.

- [ ] **Step 3: Register the task in both processes**

Admin 的 `NewHandlerRegistry` 参数增加：

```go
scheduledtaskinfra.NewDingTalkNotificationJob(notificationService),
```

`taskd` 的通知服务改为：

```go
notificationService := inappnotificationapp.NewServiceWithDingTalk(
	notificationStore,
	inappnotificationinfra.NewDingTalkDelivery(db, nil),
)
```

并把同一个钉钉任务加入 `taskd` 注册表。

- [ ] **Step 4: Run wiring tests and verify GREEN**

Run the command from Step 2. Expected: PASS.

### Task 4: 让任务编辑器支持钉钉通知任务

**Files:**
- Modify: `admin/src/views/scheduled-task/components/TaskEditorDialog.vue`
- Modify: `admin/scripts/check-scheduled-task.mjs`

- [ ] **Step 1: Extend the static contract first**

在检查脚本中要求以下标识存在：

```js
'notification.dingtalk.send'
'admin:menu:notification:dingtalk-send'
'dingTalkNotificationRecipientOptions'
'isNotificationJob'
```

- [ ] **Step 2: Run contract and verify RED**

```bash
cd admin && node scripts/check-scheduled-task.mjs
```

Expected: FAIL on the first missing DingTalk integration snippet.

- [ ] **Step 3: Generalize the notification task form**

增加两个渠道权限和统一判定：

```ts
const canDingTalkNotificationSend = computed(() => hasPerm('admin:menu:notification:dingtalk-send'))
const isInAppNotificationJob = computed(() => handlerConfig.handlerKey === 'notification.in_app.send')
const isDingTalkNotificationJob = computed(() => handlerConfig.handlerKey === 'notification.dingtalk.send')
const isNotificationJob = computed(() => isInAppNotificationJob.value || isDingTalkNotificationJob.value)
```

Go Job 下拉按各自权限过滤，通知表单、收件人加载、参数构造和字段校验改用 `isNotificationJob`。钉钉任务通过 `adminApi.dingTalkNotificationRecipientOptions()` 获取收件选项，站内信任务继续调用原接口。

- [ ] **Step 4: Run contract and TypeScript checks**

```bash
cd admin && node scripts/check-scheduled-task.mjs && npm run build
```

Expected: PASS.

### Task 5: 格式化和回归验证

**Files:**
- Verify all files changed in Tasks 1-4.

- [ ] **Step 1: Format changed Go files**

```bash
cd backend && gofmt -w internal/modules/scheduledtask/infrastructure/dingtalk_notification_job.go internal/modules/scheduledtask/infrastructure/dingtalk_notification_job_test.go internal/modules/scheduledtask/transport/httpadmin/handler.go internal/modules/scheduledtask/transport/httpadmin/notification_risk_test.go internal/routes/v2/admin/routes.go internal/routes/v2/admin/in_app_notification_routes_test.go cmd/taskd/main.go cmd/taskd/main_test.go
```

- [ ] **Step 2: Run backend focused regression tests**

```bash
cd backend && GOCACHE=$PWD/../.cache/go-build go test ./internal/modules/scheduledtask/infrastructure ./internal/modules/scheduledtask/transport/httpadmin ./internal/routes/v2/admin ./cmd/taskd -count=1
```

Expected: PASS.

- [ ] **Step 3: Run Admin verification**

```bash
cd admin && npm run check:all
```

Expected: PASS, or report unrelated baseline failures separately.

- [ ] **Step 4: Inspect the scoped diff**

```bash
git diff --check
git status --short
```

Confirm only task-related files plus pre-existing user changes are present, and no generated output or unrelated file was altered.
