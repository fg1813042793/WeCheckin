# 通用定时任务平台 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在 WeCheckin 中实现基于 MySQL 和 Redis Streams 的独立定时任务服务，完整支持 Go、流程发起、HTTP、Shell、SQL 处理器及后台管理页面。

**Architecture:** HTTP 服务仅提供配置和管理 API；独立 `cmd/taskd` 以 `scheduler`、`worker` 或 `all` 角色运行。MySQL 保存任务和执行事实，Redis Streams 负责投递，数据库唯一运行键和状态条件更新保证多实例安全。

**Tech Stack:** Go 1.24.5、Hertz、GORM、MySQL 8、go-redis v9/Redis Streams、robfig/cron v3、xwb1989/sqlparser、Vue 3、TypeScript、Element Plus、Vite。

**Workspace rule:** 当前工作区包含用户未提交的通用工作流改动。实现必须在当前工作区进行，不回滚任何现有修改；功能代码不自动提交，避免提交共享文件中的用户改动。计划文档可单独提交。

---

## 文件结构

新增调度模块采用与现有 workflow 模块一致的分层：

```text
backend/internal/model/scheduledtask/models.go
backend/internal/modules/scheduledtask/domain/types.go
backend/internal/modules/scheduledtask/domain/schedule.go
backend/internal/modules/scheduledtask/application/service.go
backend/internal/modules/scheduledtask/application/handlers.go
backend/internal/modules/scheduledtask/infrastructure/gorm_store.go
backend/internal/modules/scheduledtask/infrastructure/redis_stream.go
backend/internal/modules/scheduledtask/infrastructure/handler_go.go
backend/internal/modules/scheduledtask/infrastructure/handler_workflow.go
backend/internal/modules/scheduledtask/infrastructure/handler_http.go
backend/internal/modules/scheduledtask/infrastructure/handler_shell.go
backend/internal/modules/scheduledtask/infrastructure/handler_sql.go
backend/internal/modules/scheduledtask/runtime/scheduler.go
backend/internal/modules/scheduledtask/runtime/worker.go
backend/internal/modules/scheduledtask/transport/httpadmin/handler.go
backend/cmd/taskd/main.go
admin/src/views/scheduled-task/tasks/index.vue
admin/src/views/scheduled-task/runs/index.vue
admin/src/views/scheduled-task/workers/index.vue
```

领域层负责 Cron、补偿、并发和状态规则；应用层负责管理用例和处理器注册；基础设施层负责 GORM、Redis 和具体处理器；runtime 只编排 Scheduler/Worker 生命周期；HTTP 和 Vue 不直接依赖内部执行细节。

### Task 1: 数据库迁移与持久化模型

**Files:**
- Create: `backend/migrations/20260901130000_create_scheduled_task_tables.sql`
- Create: `backend/test/internal/bootstrap/migrations/scheduled_task_migration_test.go`
- Create: `backend/internal/model/scheduledtask/models.go`
- Modify: `backend/internal/model/aliases.go`

- [ ] **Step 1: 写失败的迁移结构测试**

测试读取迁移文件并断言三张表、运行键唯一索引、到期扫描索引、任务外键字段和日志顺序索引存在：

```go
func TestScheduledTaskMigrationCreatesRuntimeTables(t *testing.T) {
    source := readMigration(t, "20260901130000_create_scheduled_task_tables.sql")
    for _, want := range []string{
        "CREATE TABLE IF NOT EXISTS `scheduled_tasks`",
        "CREATE TABLE IF NOT EXISTS `scheduled_task_runs`",
        "CREATE TABLE IF NOT EXISTS `scheduled_task_run_logs`",
        "UNIQUE KEY `uk_scheduled_task_run_key` (`run_key`)",
        "KEY `idx_scheduled_tasks_due` (`enabled`,`next_run_at`)",
        "KEY `idx_scheduled_task_runs_recovery` (`run_status`,`heartbeat_at`)",
    } {
        if !strings.Contains(source, want) { t.Fatalf("migration missing %q", want) }
    }
}
```

- [ ] **Step 2: 运行测试并确认失败**

Run: `cd backend && GOCACHE=$PWD/.cache/go-build go test ./test/internal/bootstrap/migrations -run TestScheduledTaskMigrationCreatesRuntimeTables -count=1`

Expected: FAIL，提示迁移文件不存在。

- [ ] **Step 3: 创建迁移和 GORM 模型**

模型定义 `Task`、`Run`、`RunLog`，字段名严格对应设计文档；时间字段使用毫秒 `int64`，配置和快照使用 `MEDIUMTEXT`。状态常量至少包含：

```go
const (
    RunStatusWaiting   = "waiting"
    RunStatusQueued    = "queued"
    RunStatusRunning   = "running"
    RunStatusRetryWait = "retry_wait"
    RunStatusSuccess   = "success"
    RunStatusFailed    = "failed"
    RunStatusCanceled  = "canceled"
    RunStatusSkipped   = "skipped"
)
```

- [ ] **Step 4: 运行模型和迁移测试**

Run: `cd backend && GOCACHE=$PWD/.cache/go-build go test ./internal/model/... ./test/internal/bootstrap/migrations -run 'ScheduledTask|scheduled_task' -count=1`

Expected: PASS。

### Task 2: 配置与 Cron 领域规则

**Files:**
- Modify: `backend/internal/config/config.go`
- Modify: `backend/internal/config/config_test.go`
- Modify: `backend/config/config.yaml`
- Modify: `backend/config/config.prod.yaml`
- Modify: `backend/go.mod`
- Modify: `backend/go.sum`
- Create: `backend/internal/modules/scheduledtask/domain/types.go`
- Create: `backend/internal/modules/scheduledtask/domain/schedule.go`
- Create: `backend/internal/modules/scheduledtask/domain/schedule_test.go`

- [ ] **Step 1: 写配置默认值和 Cron 失败测试**

覆盖分钟/秒表达式、IANA 时区、5 秒最短间隔、未来 5 次预览、非法字段数量、三种补偿和三种并发策略：

```go
func TestParseScheduleSupportsMinuteAndSecondPrecision(t *testing.T) {
    minute, err := ParseSchedule("minute", "*/5 * * * *", "Asia/Shanghai", 5*time.Second)
    if err != nil || minute == nil { t.Fatalf("minute schedule: %v", err) }
    second, err := ParseSchedule("second", "*/10 * * * * *", "Asia/Shanghai", 5*time.Second)
    if err != nil || second == nil { t.Fatalf("second schedule: %v", err) }
}

func TestSecondScheduleRejectsIntervalBelowConfiguredMinimum(t *testing.T) {
    _, err := ParseSchedule("second", "*/2 * * * * *", "UTC", 5*time.Second)
    if !errors.Is(err, ErrScheduleTooFrequent) { t.Fatalf("error=%v", err) }
}
```

- [ ] **Step 2: 运行测试并确认失败**

Run: `cd backend && GOCACHE=$PWD/.cache/go-build go test ./internal/config ./internal/modules/scheduledtask/domain -count=1`

Expected: FAIL，提示配置字段或 `ParseSchedule` 未定义。

- [ ] **Step 3: 添加 robfig/cron 并实现领域规则**

Run: `cd backend && go get github.com/robfig/cron/v3@v3.0.1`

使用显式 5 段/6 段 Parser，禁用描述符；`Preview` 返回 UTC 毫秒和任务时区显示值。`ComputeDue` 返回要创建的运行和错过汇总，不访问数据库。

- [ ] **Step 4: 完成配置结构和环境变量绑定**

新增 `ScheduledTaskConfig`、`ShellCommandConfig`、`HTTPPolicyConfig`、`SQLDataSourceConfig`，默认 `enable_shell=false`、`enable_sql=false`，并绑定 Worker 数、轮询周期和开关环境变量。

- [ ] **Step 5: 运行测试**

Run: `cd backend && GOCACHE=$PWD/.cache/go-build go test ./internal/config ./internal/modules/scheduledtask/domain -count=1`

Expected: PASS。

### Task 3: 应用服务、任务 CRUD 与运行生成

**Files:**
- Create: `backend/internal/modules/scheduledtask/application/service.go`
- Create: `backend/internal/modules/scheduledtask/application/service_test.go`
- Create: `backend/internal/modules/scheduledtask/infrastructure/gorm_store.go`
- Create: `backend/internal/modules/scheduledtask/infrastructure/gorm_store_test.go`

- [ ] **Step 1: 写失败的应用服务测试**

使用 FakeStore 覆盖任务新增/编辑校验、启用时计算 `next_run_at`、计划运行键去重、`queue_once` 合并、立即执行和手工重试创建新运行：

```go
func TestCreateTaskValidatesHandlerAndComputesNextRun(t *testing.T) {
    service := NewService(store, registry, schedulerConfig)
    task, err := service.CreateTask(ctx, 7, CreateTaskRequest{
        Code: "workflow-monthly", Name: "月度流程", HandlerType: "workflow",
        CronPrecision: "minute", CronExpression: "0 9 1 * *", Timezone: "Asia/Shanghai",
    })
    if err != nil { t.Fatal(err) }
    if task.NextRunAt == 0 { t.Fatal("next run is required") }
}
```

- [ ] **Step 2: 运行测试并确认失败**

Run: `cd backend && GOCACHE=$PWD/.cache/go-build go test ./internal/modules/scheduledtask/application ./internal/modules/scheduledtask/infrastructure -count=1`

Expected: FAIL，提示服务和 Store 未定义。

- [ ] **Step 3: 实现 Store 接口和 GORM Store**

接口包含 CRUD、分页查询、到期任务锁定、运行插入、抢占、心跳、终态、恢复投递、日志追加和保留期清理。到期生成必须在事务中更新任务游标并依赖 `run_key` 唯一约束幂等。

- [ ] **Step 4: 实现应用服务**

保存前调用 Cron 和处理器 `ValidateConfig`；修改任务使用 `version` 乐观锁；软删除前禁用任务；立即执行只创建 `queued` 运行并调用 QueuePublisher。

- [ ] **Step 5: 运行测试**

Run: `cd backend && GOCACHE=$PWD/.cache/go-build go test ./internal/modules/scheduledtask/application ./internal/modules/scheduledtask/infrastructure -count=1`

Expected: PASS。

### Task 4: Redis Streams 队列

**Files:**
- Create: `backend/internal/modules/scheduledtask/infrastructure/redis_stream.go`
- Create: `backend/internal/modules/scheduledtask/infrastructure/redis_stream_test.go`

- [ ] **Step 1: 写队列契约失败测试**

通过 FakeRedisCommands 覆盖创建 Consumer Group、`XADD` 投递 `run_id`、`XREADGROUP` 消费、`XACK`、`XAUTOCLAIM` 和 Worker TTL 心跳；键必须带配置前缀。

- [ ] **Step 2: 运行测试并确认失败**

Run: `cd backend && GOCACHE=$PWD/.cache/go-build go test ./internal/modules/scheduledtask/infrastructure -run 'RedisStream|WorkerHeartbeat' -count=1`

Expected: FAIL，提示 Redis Stream 适配器未定义。

- [ ] **Step 3: 实现 Redis Stream 适配器**

定义窄接口 `Queue`，消息只包含 `run_id`；Consumer Group 固定为 `<prefix>:scheduled-task:workers`，Worker 名称使用稳定节点 ID。Redis 错误不改变 MySQL 运行终态。

- [ ] **Step 4: 运行测试**

Run: `cd backend && GOCACHE=$PWD/.cache/go-build go test ./internal/modules/scheduledtask/infrastructure -run 'RedisStream|WorkerHeartbeat' -count=1`

Expected: PASS。

### Task 5: Scheduler、Worker 与独立 taskd

**Files:**
- Create: `backend/internal/modules/scheduledtask/runtime/scheduler.go`
- Create: `backend/internal/modules/scheduledtask/runtime/scheduler_test.go`
- Create: `backend/internal/modules/scheduledtask/runtime/worker.go`
- Create: `backend/internal/modules/scheduledtask/runtime/worker_test.go`
- Create: `backend/cmd/taskd/main.go`
- Create: `backend/cmd/taskd/main_test.go`

- [ ] **Step 1: 写运行时失败测试**

覆盖 Scheduler 生成并投递到期运行、恢复未投递运行、唤醒 `retry_wait`/`waiting`、Worker 抢占失败时 ACK、执行成功后落库再 ACK、失败退避、取消、超时、Pending 认领和优雅关闭。

- [ ] **Step 2: 运行测试并确认失败**

Run: `cd backend && GOCACHE=$PWD/.cache/go-build go test ./internal/modules/scheduledtask/runtime ./cmd/taskd -count=1`

Expected: FAIL，提示 Scheduler/Worker 未定义。

- [ ] **Step 3: 实现 Scheduler 和 Worker**

Scheduler 每个动作使用独立超时 Context；Worker 使用有界 goroutine 池和每运行 Context。数据库状态更新失败时不 ACK；终态落库成功后 ACK。永久错误直接失败，暂时错误按任务快照进入 `retry_wait`。

- [ ] **Step 4: 实现 taskd 启动与信号关闭**

`main.go` 复用现有配置、数据库、Redis 和日志初始化，校验 `--role=scheduler|worker|all`，注册处理器后启动对应角色。禁止启动 Hertz Server。

- [ ] **Step 5: 运行测试**

Run: `cd backend && GOCACHE=$PWD/.cache/go-build go test ./internal/modules/scheduledtask/runtime ./cmd/taskd -count=1`

Expected: PASS。

### Task 6: 处理器注册、Go 与流程发起

**Files:**
- Create: `backend/internal/modules/scheduledtask/application/handlers.go`
- Create: `backend/internal/modules/scheduledtask/application/handlers_test.go`
- Create: `backend/internal/modules/scheduledtask/infrastructure/handler_go.go`
- Create: `backend/internal/modules/scheduledtask/infrastructure/handler_workflow.go`
- Create: `backend/internal/modules/scheduledtask/infrastructure/handler_workflow_test.go`
- Modify: `backend/internal/modules/workflow/application/service.go`
- Modify: `backend/internal/modules/workflow/application/service_test.go`

- [ ] **Step 1: 写注册表和流程幂等失败测试**

注册表拒绝重复键和未知处理器；流程任务按发起人使用 `business_type=scheduled_task`、`business_key=<run_id>:user:<user_id>`，相同运行重试返回已有实例而不创建第二条。历史单值发起人配置继续兼容 `<run_id>` 业务键。

- [ ] **Step 2: 运行测试并确认失败**

Run: `cd backend && GOCACHE=$PWD/.cache/go-build go test ./internal/modules/scheduledtask/... ./internal/modules/workflow/application -run 'HandlerRegistry|WorkflowHandler|Idempotent' -count=1`

Expected: FAIL。

- [ ] **Step 3: 实现处理器注册和内置 Go 任务**

处理器元数据返回键、名称、配置 Schema 和风险级别。首版注册 `scheduled-task.cleanup` 和 `workflow.notification.dispatch_due`；Go 任务只能从注册表选择。

- [ ] **Step 4: 实现流程处理器和幂等启动**

流程配置包含已发布定义 ID、版本策略、固定版本、发起人数组和表单 JSON。管理端使用流程下拉和组织用户树多选；扩展工作流应用服务，使业务唯一键冲突时查询并返回已有实例；不改变普通后台发起行为。

- [ ] **Step 5: 运行测试**

Run: `cd backend && GOCACHE=$PWD/.cache/go-build go test ./internal/modules/scheduledtask/... ./internal/modules/workflow/application -run 'HandlerRegistry|WorkflowHandler|Idempotent' -count=1`

Expected: PASS。

### Task 7: HTTP/Webhook 处理器

**Files:**
- Create: `backend/internal/modules/scheduledtask/infrastructure/handler_http.go`
- Create: `backend/internal/modules/scheduledtask/infrastructure/handler_http_test.go`

- [ ] **Step 1: 写 SSRF 与执行失败测试**

覆盖允许域名成功、回环/链路本地/云元数据拒绝、DNS 解析后地址复检、重定向复检、请求响应大小、超时、凭据引用和 `X-Scheduled-Run-ID`。

- [ ] **Step 2: 运行测试并确认失败**

Run: `cd backend && GOCACHE=$PWD/.cache/go-build go test ./internal/modules/scheduledtask/infrastructure -run HTTPHandler -count=1`

Expected: FAIL。

- [ ] **Step 3: 实现受控 HTTP Client**

使用自定义 `DialContext` 校验最终 IP，禁用代理继承，按策略控制重定向；只从服务端凭据注册表解析敏感请求头。错误映射为暂时或永久错误。

- [ ] **Step 4: 运行测试**

Run: `cd backend && GOCACHE=$PWD/.cache/go-build go test ./internal/modules/scheduledtask/infrastructure -run HTTPHandler -count=1`

Expected: PASS。

### Task 8: Shell 与 SQL 处理器

**Files:**
- Create: `backend/internal/modules/scheduledtask/infrastructure/handler_shell.go`
- Create: `backend/internal/modules/scheduledtask/infrastructure/handler_shell_test.go`
- Create: `backend/internal/modules/scheduledtask/infrastructure/handler_sql.go`
- Create: `backend/internal/modules/scheduledtask/infrastructure/handler_sql_test.go`
- Modify: `backend/go.mod`
- Modify: `backend/go.sum`

- [ ] **Step 1: 写 Shell 安全失败测试**

覆盖总开关、命令键白名单、参数规则、工作目录、环境变量白名单、禁止 Shell 解析、超时取消和输出截断。

- [ ] **Step 2: 写 SQL AST 与事务失败测试**

覆盖只读/写入权限、单语句、DDL/多语句拒绝、参数化执行、注册存储过程键、超时、最大行数、最大影响行数和超限回滚。测试使用 FakeSQLExecutor，不连接生产数据库。

- [ ] **Step 3: 运行测试并确认失败**

Run: `cd backend && GOCACHE=$PWD/.cache/go-build go test ./internal/modules/scheduledtask/infrastructure -run 'ShellHandler|SQLHandler' -count=1`

Expected: FAIL。

- [ ] **Step 4: 添加 MySQL AST 解析依赖**

Run: `cd backend && go get github.com/xwb1989/sqlparser@v0.0.0-20180606152119-120387863bf2`

Expected: `go.mod` 和 `go.sum` 只增加轻量 MySQL 解析器及其必要依赖。

- [ ] **Step 5: 实现 Shell 和 SQL 处理器**

Shell 使用 `exec.CommandContext(path, args...)`。普通 SQL 使用 `sqlparser.Parse` 后按 `*sqlparser.Select`、`*sqlparser.Insert`、`*sqlparser.Update`、`*sqlparser.Delete` 做白名单判断；存储过程使用服务端注册的过程键和参数数组，不解析后台输入的 `CALL` 字符串。执行器只解析服务端数据源键，查询流式计数，写入在事务内检查 `RowsAffected` 后提交。

- [ ] **Step 6: 运行测试**

Run: `cd backend && GOCACHE=$PWD/.cache/go-build go test ./internal/modules/scheduledtask/infrastructure -run 'ShellHandler|SQLHandler' -count=1`

Expected: PASS。

### Task 9: 管理 API、菜单与权限

**Files:**
- Create: `backend/internal/modules/scheduledtask/transport/httpadmin/handler.go`
- Create: `backend/internal/modules/scheduledtask/transport/httpadmin/handler_test.go`
- Modify: `backend/internal/routes/v2/admin/routes.go`
- Modify: `backend/internal/middleware/admin/route_permissions.go`
- Modify: `backend/internal/middleware/admin/permission_test.go`
- Modify: `backend/internal/support/adminmenuperm/declarations.go`
- Create: `backend/internal/support/adminmenuperm/scheduled_task_structure_test.go`
- Modify: `backend/internal/support/adminrouteperm/catalog.go`
- Create: `backend/internal/support/adminrouteperm/scheduled_task_structure_test.go`
- Modify: `backend/internal/routes/v2/swagger/swagger.go`
- Modify: `backend/docs/swagger/docs.go`
- Modify: `backend/docs/swagger/swagger.json`
- Modify: `backend/docs/swagger/swagger.yaml`

- [ ] **Step 1: 写路由、菜单和权限失败测试**

断言设计文档中的全部 API、一级“任务调度”目录、三个子菜单和高风险按钮权限均存在，并确保通配路由顺序不会吞掉 `/cron-preview`、`/handlers` 或 `/workers`。

- [ ] **Step 2: 运行测试并确认失败**

Run: `cd backend && GOCACHE=$PWD/.cache/go-build go test ./internal/modules/scheduledtask/transport/httpadmin ./internal/middleware/admin ./internal/support/adminmenuperm ./internal/support/adminrouteperm -count=1`

Expected: FAIL。

- [ ] **Step 3: 实现 HTTP Handler 和路由**

Handler 只解析参数和认证管理员，调用应用服务并使用现有 `response.JSON/Fail`。立即执行在 Redis 失败时返回已创建运行及 `dispatchPending=true`。

- [ ] **Step 4: 注册菜单和权限**

新增 `scheduled-task:*` 权限目录、按钮声明、API 分类和路由权限映射；不自动扩大现有角色授权。

- [ ] **Step 5: 补充并生成 Swagger**

在 `internal/routes/v2/swagger/swagger.go` 为全部定时任务管理路由增加 `AdminToken` 注解，然后运行：

Run: `cd backend && go generate ./cmd`

Expected: `docs/swagger` 三个生成文件包含 `/api/v2/admin/scheduled-tasks`、运行记录和 Worker 接口。

- [ ] **Step 6: 运行测试**

Run: `cd backend && GOCACHE=$PWD/.cache/go-build go test ./internal/modules/scheduledtask/transport/httpadmin ./internal/middleware/admin ./internal/support/adminmenuperm ./internal/support/adminrouteperm -count=1`

Expected: PASS。

### Task 10: 管理端 API 与三个页面

**Files:**
- Modify: `admin/src/api/index.ts`
- Modify: `admin/src/router/adminRoutes.ts`
- Create: `admin/src/views/scheduled-task/types.ts`
- Create: `admin/src/views/scheduled-task/tasks/index.vue`
- Create: `admin/src/views/scheduled-task/components/TaskEditorDrawer.vue`
- Create: `admin/src/views/scheduled-task/runs/index.vue`
- Create: `admin/src/views/scheduled-task/workers/index.vue`
- Create: `admin/scripts/check-scheduled-task.mjs`
- Modify: `admin/package.json`

- [ ] **Step 1: 写失败的静态回归检查**

脚本断言三条路由、API 方法、Cron 预览、分钟/秒和时区控件、五类处理器、权限判断、运行详情、重试/取消、Worker 心跳与离线状态均存在。

- [ ] **Step 2: 运行检查并确认失败**

Run: `cd admin && npm run check:scheduled-task`

Expected: FAIL，提示脚本或页面缺失。

- [ ] **Step 3: 实现类型和 API**

定义 Task、Run、RunLog、Worker、HandlerMetadata、分页响应和各处理器配置类型；API 使用现有 `ADMIN_V2`、`encodePath` 和 `jsonConfig`。

- [ ] **Step 4: 实现任务列表和编辑抽屉**

使用现有 `admin-page/admin-card/admin-toolbar` 视觉体系。操作使用图标按钮和 Tooltip；编辑抽屉 `append-to-body`。Cron 预览由后端接口计算，不在前端复制解析规则。

- [ ] **Step 5: 实现执行记录和节点页面**

运行详情抽屉展示快照、摘要和日志；取消/重试按权限显示。节点页根据心跳时间显示在线/离线和 Worker Pool 使用率。

- [ ] **Step 6: 运行前端检查和构建**

Run: `cd admin && npm run check:scheduled-task && npm run build`

Expected: PASS；Vite 仅允许已有依赖注释警告。

### Task 11: Docker、启动配置与运维文档

**Files:**
- Modify: `backend/Dockerfile`
- Modify: `backend/docker-compose.yml`
- Create: `backend/docs/scheduled-task-operations.md`
- Create: `backend/cmd/taskd/deployment_structure_test.go`

- [ ] **Step 1: 写部署结构失败测试**

断言 Dockerfile 生成 `main` 和 `taskd` 两个二进制，Compose 新增 `taskd` 服务并复用数据库/Redis环境变量，HTTP 容器命令仍只运行 `main`。

- [ ] **Step 2: 运行测试并确认失败**

Run: `cd backend && GOCACHE=$PWD/.cache/go-build go test ./cmd/taskd -run Deployment -count=1`

Expected: FAIL。

- [ ] **Step 3: 修改镜像和 Compose**

同一镜像复制两个二进制；新增 `taskd` 服务，命令为 `./taskd --role=all --env=prod`，依赖 MySQL/Redis 健康，不暴露 HTTP 端口。

- [ ] **Step 4: 编写运维文档**

说明迁移、最小部署、角色拆分、Redis 恢复、Worker 离线、Shell/SQL 开关、凭据/数据源注册和日志清理。

- [ ] **Step 5: 运行测试**

Run: `cd backend && GOCACHE=$PWD/.cache/go-build go test ./cmd/taskd -run Deployment -count=1`

Expected: PASS。

### Task 12: 全量验证与页面验收准备

**Files:**
- Modify only if verification finds an in-scope defect.

- [ ] **Step 1: 格式化并检查变更边界**

Run: `cd backend && gofmt -w internal/model/scheduledtask internal/modules/scheduledtask cmd/taskd`

Run: `git diff --check`

Expected: PASS，且不回滚任何既有工作区修改。

- [ ] **Step 2: 后端全量测试**

Run: `cd backend && GOCACHE=$PWD/.cache/go-build go test ./...`

Expected: PASS；若存在确认过的历史基线失败，单独列出并保证 scheduledtask 范围全部通过。

- [ ] **Step 3: 管理端全量检查**

Run: `cd admin && npm run check:all`

Expected: PASS。

- [ ] **Step 4: 本地运行 taskd 验证启动角色**

Run: `cd backend && GOCACHE=$PWD/.cache/go-build go run ./cmd/taskd --role=all --env=dev`

Expected: 成功连接 MySQL/Redis、注册 Consumer Group、写入节点心跳；验证后优雅停止，不留下运行会话。

- [ ] **Step 5: 浏览器验证**

登录后台后验证任务列表、五类编辑表单、Cron 预览、立即执行、运行详情和节点状态；记录控制台错误和桌面/移动视口截图。若无登录凭据，明确交由用户完成最终手工页面验证。

---

## 实施顺序与检查点

按 Task 1-5 完成可运行的调度内核；Task 6-8 补齐五类处理器；Task 9-10 完成后台闭环；Task 11-12 完成部署和全量验证。每个 Task 必须先看到目标测试失败，再做最小实现并看到同一测试通过。

由于当前工作区存在用户未提交改动，执行期间不自动创建功能提交；每个检查点使用 `git status --short` 和限定路径的 `git diff --check` 核对边界。
