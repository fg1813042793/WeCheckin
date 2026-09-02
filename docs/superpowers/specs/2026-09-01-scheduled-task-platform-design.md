# 通用定时任务平台设计

## 1. 目标与范围

为 WeCheckin 后台补充独立的通用定时任务平台，统一承载系统任务、流程定时发起、HTTP/Webhook、Shell 和 SQL 任务。调度和执行进程必须与 HTTP 服务隔离，避免任务负载占用后台接口的 Worker、CPU 和内存配额。

首版包含：

- 分钟级和秒级 Cron，每个任务独立配置时区。
- 多实例安全的任务扫描、Redis 队列投递和执行抢占。
- 停机补偿、并发控制、超时、取消、自动重试和手工重试。
- Go、流程发起、HTTP/Webhook、Shell、SQL 五类处理器。
- 任务定义、执行记录、分段日志和执行节点管理页面。
- 独立的权限、操作审计和高风险处理器安全控制。

本期不包含：

- 修改或迁移现有绩效业务流程。
- 在 `h5app` 增加定时任务页面。
- 允许后台直接输入 Go 函数名、任意 Shell 字符串或多语句 SQL。
- 承诺外部副作用严格恰好执行一次。
- 在 HTTP 服务进程内以内嵌模式启动调度器或 Worker。

## 2. 总体架构

系统拆分为三个运行角色：

1. HTTP 服务：提供任务配置、查询、启停、立即执行、取消和重试接口，不扫描或消费任务。
2. Scheduler：扫描 MySQL 中到期任务，生成执行实例并投递 Redis Stream。
3. Worker：通过 Redis Consumer Group 消费执行实例，执行处理器并把状态和日志写回 MySQL。

新增独立程序 `backend/cmd/taskd`，支持以下启动角色：

```text
taskd --role=scheduler
taskd --role=worker
taskd --role=all
```

只部署一个任务服务时使用 `--role=all`，同一进程内启动 Scheduler 和有界 Worker Pool。需要扩容时可以分别部署多个 Scheduler 和 Worker。无论使用哪种角色，`taskd` 始终与 HTTP 服务分开运行。

MySQL 是任务定义、执行实例和执行日志的唯一事实源。Redis 只承担队列、Consumer Group 和节点心跳职责；Redis 消息丢失或短暂不可用时，可以根据 MySQL 中的执行状态恢复投递。

Redis 使用现有 `go-redis` 客户端实现 Streams 适配器，基础键约定为：

```text
wecheckin:scheduled-task:runs
wecheckin:scheduled-task:workers
```

键前缀允许通过配置覆盖，以隔离不同环境。

## 3. 调度数据流

Scheduler 按固定周期查询 `enabled = 1 AND next_run_at <= now` 的任务。对每个到期任务在数据库事务中完成：

1. 按 Cron、时区和停机补偿策略计算应生成的计划时间。
2. 写入执行实例，计划执行使用稳定运行键去重。
3. 更新任务的 `last_scheduled_at` 和 `next_run_at`。
4. 提交事务后把执行实例 ID 投递到 Redis Stream。

多个 Scheduler 可以同时扫描。计划运行键和数据库唯一约束保证同一任务、同一计划时间只生成一个执行实例。事务提交后 Redis 投递失败时，执行实例保持可恢复状态；Scheduler 周期扫描未投递或投递超时的实例并重新发布。

Worker 从 Consumer Group 读取执行实例 ID，然后使用条件更新把可执行状态原子变更为 `running`。抢占失败表示消息已被其他 Worker 处理，当前 Worker 直接确认该 Redis 消息。Worker 只有在数据库已经记录成功、失败、取消或等待重试后才确认消息。

## 4. 数据模型

### 4.1 任务定义 `scheduled_tasks`

核心字段：

| 字段 | 说明 |
| --- | --- |
| `id` | 主键 |
| `code` | 稳定且唯一的任务编码 |
| `name` | 任务名称 |
| `handler_type` | `go`、`workflow`、`http`、`shell`、`sql` |
| `handler_config_json` | 处理器配置，不保存明文凭据 |
| `cron_expression` | 5 段分钟级或 6 段秒级 Cron |
| `cron_precision` | `minute` 或 `second` |
| `timezone` | IANA 时区名称 |
| `enabled` | 是否启用 |
| `misfire_policy` | `skip`、`fire_once`、`catch_up` |
| `max_catch_up` | 逐次补执行的最大数量 |
| `concurrency_policy` | `skip`、`queue_once`、`allow` |
| `timeout_seconds` | 单次尝试超时 |
| `max_retries` | 自动重试次数，范围 0 至 5 |
| `retry_backoff_json` | 重试退避配置 |
| `last_scheduled_at` | 最近已处理的计划时间 |
| `next_run_at` | 下一次扫描时间，统一保存 UTC |
| `version` | 乐观锁版本 |
| `created_by`、`updated_by` | 管理员用户 ID |
| `deleted_at` | 软删除时间 |
| `add_time`、`edit_time` | 创建和更新时间 |

秒级任务的最短间隔默认限制为 5 秒，并允许通过服务端配置提高；修改任务时必须重新计算 `next_run_at`。

### 4.2 执行实例 `scheduled_task_runs`

核心字段：

| 字段 | 说明 |
| --- | --- |
| `id` | 主键 |
| `run_key` | 全局唯一运行键 |
| `task_id` | 任务 ID |
| `parent_run_id` | 手工重试来源，可空 |
| `trigger_type` | `scheduled`、`manual`、`manual_retry`、`misfire` |
| `status` | 执行状态 |
| `task_snapshot_json` | 本次执行使用的完整任务快照 |
| `scheduled_at` | 计划时间，统一保存 UTC |
| `coalesced_count` | 合并的触发或错过次数 |
| `attempt` | 当前自动尝试次数 |
| `worker_id` | 抢占任务的 Worker 标识 |
| `redis_message_id` | 最近一次 Redis 消息 ID |
| `queued_at`、`started_at`、`finished_at` | 生命周期时间 |
| `heartbeat_at` | 运行心跳 |
| `cancel_requested_at` | 取消请求时间 |
| `result_summary` | 脱敏且截断的结果摘要 |
| `error_code`、`error_summary` | 分类错误和脱敏错误摘要 |
| `add_time`、`edit_time` | 创建和更新时间 |

计划任务运行键格式为 `task:<taskId>:scheduled:<scheduledAtUnixMilli>`。手工执行使用 UUID 组成独立运行键。自动重试继续使用同一执行实例并递增 `attempt`；管理员对失败记录执行手工重试时创建新实例，并通过 `parent_run_id` 关联原实例。

### 4.3 分段日志 `scheduled_task_run_logs`

分段日志保存 `run_id`、序号、级别、阶段、内容和时间。单条日志、单次执行日志总量和返回列表均设置上限；超过上限时写入一条截断标记，不继续保存剩余输出。

执行摘要默认保留 90 天，详细分段日志默认保留 30 天。清理周期和保留时间可配置，清理由注册的 Go 处理器执行。

## 5. 执行状态机

执行状态包括：

```text
waiting -> queued -> running -> success
                    |       -> retry_wait -> queued
                    |       -> failed
                    |       -> canceled
waiting/queued/retry_wait -> canceled
skipped
```

- `waiting`：采用 `queue_once` 且同一任务已有运行实例时，保留一个等待实例，不投递 Redis。
- `queued`：已具备执行条件，可投递或已经投递 Redis。
- `running`：由 Worker 通过数据库条件更新抢占。
- `retry_wait`：自动重试尚未到期，到期后重新进入 `queued`。
- `success`、`failed`、`canceled`、`skipped`：终态。

Worker 运行期间定期更新心跳并检查 `cancel_requested_at`。HTTP、Shell、SQL 使用可取消的 Context；Go 和流程处理器接口必须遵守 Context。Worker 失联或执行超时后，Scheduler 根据尝试次数把实例恢复为 `retry_wait` 或标记为 `failed`。

`queue_once` 最多保留一个 `waiting` 实例；后续触发合并到该实例并递增 `coalesced_count`。前一个实例进入终态后，等待实例转为 `queued`。

## 6. Cron、补偿与并发规则

Cron 支持：

- 分钟级：标准 5 段表达式。
- 秒级：包含秒字段的 6 段表达式。
- 每任务独立 IANA 时区。
- 后台校验并预览未来 5 次执行时间。

计划时间统一换算为 UTC 存储，运行键使用 UTC 时间去重。Cron 解析仍按任务时区计算，夏令时重复或跳过的本地时间不会产生相同 UTC 运行键。

停机补偿策略：

- `skip`：推进到第一个未来时间，写一条带错过次数的 `skipped` 汇总记录。
- `fire_once`：生成一次补偿执行，`coalesced_count` 记录总错过次数。
- `catch_up`：按计划时间顺序生成最多 `max_catch_up` 个执行，其余错过次数写入 `skipped` 汇总记录。

并发策略：

- `skip`：已有 `running`、`queued`、`retry_wait` 或 `waiting` 实例时跳过新触发并记录原因。
- `queue_once`：最多排队一个等待实例，后续触发合并。
- `allow`：每个触发独立生成执行实例，但仍受 Worker Pool 全局并发上限约束。

## 7. 处理器设计

所有处理器实现统一接口：

```go
type Handler interface {
    Type() string
    ValidateConfig(ctx context.Context, raw json.RawMessage) error
    Execute(ctx context.Context, run RunContext) (Result, error)
}
```

`RunContext` 至少包含稳定 `run_id`、任务快照、尝试次数、触发方式和结构化日志写入器。处理器必须返回分类错误，供重试策略区分超时、暂时错误、配置错误和永久错误。

### 7.1 Go 处理器

Go 任务通过代码注册稳定处理器键。后台从处理器元数据接口选择处理器并按其 JSON Schema 配置参数，不能输入函数名。通知到期投递、日志清理等系统任务可注册为内置 Go 处理器。

### 7.2 流程发起处理器

配置包含已发布流程定义、一个或多个固定发起人、表单初始 JSON，以及 `latest_published` 或 `fixed_version` 版本策略，默认使用最新发布版本。每次执行重新校验：

- 每名发起人存在、有效且位于流程允许发起范围内。
- 流程定义存在并具有符合策略的发布版本。
- 表单初始值通过对应发布版本的表单校验。

处理器调用通用工作流应用服务发起流程，不能直接写流程运行表。每名发起人创建独立实例，使用 `<run_id>:user:<user_id>` 作为幂等业务键；实际创建的流程实例继续固定其启动时的发布版本。历史单值 `starterId` 配置继续使用原 `run_id` 业务键，避免升级后重试产生重复实例。

### 7.3 HTTP/Webhook 处理器

支持常用 HTTP 方法、Query、请求头、JSON Body、超时和预期状态码。任务只保存服务端凭据引用，不保存 Token、密码或签名密钥明文。

服务端配置允许访问的域名和 CIDR。请求发出前和连接时都校验解析地址，默认阻止回环、链路本地、云元数据地址和未授权私网，限制重定向、请求体和响应体大小。每次请求自动附带 `X-Scheduled-Run-ID`，供接收方幂等处理。

### 7.4 Shell 处理器

Shell 能力默认关闭。任务只能选择服务端配置的命令键；命令注册项定义可执行文件绝对路径、工作目录、允许的环境变量和参数规则。

执行使用 `exec.CommandContext` 和参数数组，不调用 Shell 解析器，禁止管道、重定向、命令替换和拼接命令。`taskd` 必须以非 root 用户运行，并限制超时、标准输出和标准错误总量。

### 7.5 SQL 处理器

SQL 能力默认关闭。首版只支持服务端注册的 MySQL 数据源；任务只引用数据源键，凭据使用独立的最小权限数据库账号。SQL 配置包含语句、参数和读写模式，后续增加其他数据库驱动时继续复用处理器接口，不改变任务与执行模型。

使用 SQL AST 解析器校验恰好一条语句：

- 只读模式只允许查询类语句。
- 写入模式允许参数化 `INSERT`、`UPDATE`、`DELETE`；存储过程调用必须选择服务端注册的过程键并传入参数数组，不能在任务中直接输入任意 `CALL`。
- 首版禁止 DDL、多语句和字符串拼接参数。

SQL 在事务中执行，限制超时、最大返回行数和最大影响行数；超过限制时回滚。

## 8. 安全与权限

新增权限码：

```text
scheduled-task:list
scheduled-task:add
scheduled-task:edit
scheduled-task:delete
scheduled-task:status
scheduled-task:run
scheduled-task:run:list
scheduled-task:run:retry
scheduled-task:run:cancel
scheduled-task:worker:list
scheduled-task:http
scheduled-task:shell
scheduled-task:sql:read
scheduled-task:sql:write
```

Shell、SQL 和 HTTP 任务除基础编辑或执行权限外，还必须具备对应类型权限。高风险能力由服务端配置总开关和管理员权限双重控制。任务新增、修改、启停、删除、立即执行、重试和取消均写入现有后台操作日志。

API、任务快照、执行日志和错误摘要均不得返回凭据明文。HTTP 凭据、Shell 环境和 SQL 数据源只通过服务端注册键引用。

## 9. 管理端设计

后台新增一级目录“任务调度”，包含：

### 9.1 定时任务

- 按名称、编码、类型和状态筛选。
- 展示 Cron、时区、下次执行时间、最近结果和启停状态。
- 支持新增、编辑、复制、启停、立即执行、查看记录和软删除。
- 编辑抽屉根据处理器类型渲染配置项。
- Cron 编辑器提供分钟/秒模式、时区、表达式校验和未来 5 次预览。
- 无高风险权限时不显示 Shell、SQL 或受控 HTTP 类型及其操作入口。

### 9.2 执行记录

- 按任务、触发方式、状态、Worker 和时间范围筛选。
- 详情抽屉展示任务快照、计划和实际时间、尝试次数、合并次数、结果摘要和分段日志。
- 支持手工重试失败实例，以及取消 `waiting`、`queued`、`retry_wait` 或可中断的 `running` 实例。

### 9.3 执行节点

- 展示节点 ID、角色、版本、启动时间、最近心跳、当前执行数和 Worker Pool 使用情况。
- `taskd` 周期写入带 TTL 的 Redis 心跳；超时节点显示离线。

## 10. HTTP 接口

```text
GET    /api/v2/admin/scheduled-tasks
POST   /api/v2/admin/scheduled-tasks
GET    /api/v2/admin/scheduled-tasks/:id
PUT    /api/v2/admin/scheduled-tasks/:id
DELETE /api/v2/admin/scheduled-tasks/:id
PATCH  /api/v2/admin/scheduled-tasks/:id/status
POST   /api/v2/admin/scheduled-tasks/:id/run

GET    /api/v2/admin/scheduled-task-runs
GET    /api/v2/admin/scheduled-task-runs/:id
POST   /api/v2/admin/scheduled-task-runs/:id/retry
POST   /api/v2/admin/scheduled-task-runs/:id/cancel

GET    /api/v2/admin/scheduled-task-handlers
POST   /api/v2/admin/scheduled-tasks/cron-preview
GET    /api/v2/admin/scheduled-task-workers
```

“立即执行”和“手工重试”接口只在事务内创建执行实例，事务提交后尝试投递 Redis，不在 HTTP 请求内执行任务。投递失败时接口返回“任务已创建、等待调度服务恢复投递”，而不是把执行实例回滚。

## 11. 取消、重试与关闭

取消未运行实例时直接条件更新为 `canceled`。取消运行实例时写入 `cancel_requested_at`；Worker 在心跳周期内取消 Context。处理器无法及时响应 Context 时，达到超时后由恢复流程处理，界面显示取消正在等待处理。

自动重试沿用原执行实例和 `run_id`，按任务退避配置从 `retry_wait` 回到 `queued`。配置错误、权限错误和白名单拒绝属于永久错误，不自动重试。管理员手工重试创建新实例并关联原实例。

`taskd` 收到停止信号后停止扫描和拉取新消息，等待当前任务在配置的宽限时间内结束。宽限时间结束后取消运行 Context，保留数据库状态和心跳信息供其他 Worker 恢复。

## 12. 一致性与错误处理

平台采用至少一次投递语义。数据库唯一运行键和状态抢占可以消除常规重复，但 HTTP、Shell 或跨库 SQL 在“外部副作用已经发生、结果尚未写回”时仍可能再次执行。管理端必须展示处理器是否具备幂等保障；高风险任务应使用 `run_id` 或业务唯一键自行去重。

- Redis 不可用：MySQL 中保留执行实例，恢复后重新投递。
- MySQL 不可用：Scheduler 和 Worker 退避并停止生成或执行新任务。
- 消息重复：数据库状态抢占失败后直接确认重复消息。
- Worker 失联：其他 Worker 认领 Redis Pending 消息，Scheduler 按心跳恢复数据库状态。
- 配置无效：禁止保存或禁止启用；运行时检测到配置漂移则标记永久失败。
- 日志或响应过大：截断并记录明确标识，不影响任务终态落库。

## 13. 配置与部署

现有 Redis 配置继续复用，新增任务服务配置至少包含：

```yaml
scheduled_task:
  redis_prefix: wecheckin
  scheduler_poll_interval: 5s
  recovery_interval: 30s
  worker_concurrency: 8
  heartbeat_interval: 10s
  worker_stale_after: 60s
  shutdown_grace_period: 30s
  second_cron_min_interval: 5s
  run_retention_days: 90
  log_retention_days: 30
  enable_http: true
  enable_shell: false
  enable_sql: false
```

Shell 命令、HTTP 凭据、允许访问地址和 SQL 数据源均在服务端配置中注册。生产环境最低部署为现有 HTTP 服务加一个：

```bash
go run ./cmd/taskd --role=all
```

## 14. 测试与验收

后端单元和集成测试至少覆盖：

- 分钟/秒 Cron、时区、未来时间预览和非法表达式。
- 多 Scheduler 扫描同一计划时间只生成一个实例。
- `skip`、`fire_once`、`catch_up` 停机补偿。
- `skip`、`queue_once`、`allow` 并发策略。
- Redis 消息重复、投递丢失、Pending 认领和 Worker 失联恢复。
- 自动重试、手工重试、取消、超时和优雅关闭。
- 五类处理器的成功、配置错误、暂时错误和永久错误。
- HTTP SSRF 防护、Shell 白名单、SQL AST 和读写权限校验。
- 工作流固定发起人、发布版本策略、表单校验和 `run_id` 幂等。
- 菜单声明、路由权限和操作审计。

管理端测试至少覆盖：

- 类型和权限控制下的任务表单显示。
- Cron 预览、启停、立即执行、重试和取消交互。
- 执行详情、分段日志和 Worker 离线状态。
- 静态回归检查、TypeScript 类型检查和生产构建。

验收时以一个 `taskd --role=all` 实例验证完整调度闭环，再增加第二个实例验证去重、Consumer Group 和失联恢复。用户负责最终手工页面测试；自动化验证必须覆盖本次改动范围。
