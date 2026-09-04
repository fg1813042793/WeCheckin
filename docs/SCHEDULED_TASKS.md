# 通用定时任务运维说明

## 1. 服务边界

定时任务平台由现有 HTTP 后台和独立 `taskd` 进程组成：

- HTTP 后台只管理任务定义、运行记录和执行节点，不执行任务。
- `taskd --role=scheduler` 只扫描 MySQL 并向 Redis Streams 投递运行 ID。
- `taskd --role=worker` 只消费 Redis Streams 并执行处理器。
- `taskd --role=all` 在一个服务实例中同时运行 scheduler 和 worker，适合最低单服务部署。

`taskd` 不启动 Hertz，也不监听 HTTP 端口，因此不会占用现有 HTTP 服务的连接、goroutine 或端口资源。

## 2. 上线步骤

首次启用或版本升级时，必须先执行数据库迁移和权限菜单初始化：

```bash
cd backend
bash init.sh
```

确认 `scheduled_tasks`、`scheduled_task_runs`、`scheduled_task_run_logs` 和后台“定时任务”菜单已经生成后，再启动服务：

```bash
go run ./cmd
./start-taskd.sh
```

`start-taskd.sh` 默认以 `all` 角色同时运行 scheduler 和 worker。可以通过参数或环境变量覆盖：

```bash
./start-taskd.sh --role=worker --worker-id=worker-a
TASKD_ROLE=scheduler ./start-taskd.sh
```

Docker Compose 默认额外启动一个无端口的 `taskd --role=all` 服务：

```bash
cd backend
docker compose up -d --build mysql redis backend taskd
```

需要拆分容量时，可以分别运行 scheduler 和 worker。多个实例共享 MySQL 与 Redis，任务游标锁、运行唯一键和运行状态抢占负责去重：

```bash
./taskd --role=scheduler
./taskd --role=worker --worker-id=worker-a
./taskd --role=worker --worker-id=worker-b
```

同一集群中的 `worker-id` 必须唯一。未显式配置时使用“主机名-进程号”。

## 3. 内置任务

Go 任务不能填写函数名，只能选择服务端注册键。当前内置：

- `scheduled-task.cleanup`：按服务端运行记录和日志保留天数分批清理历史。
- `workflow.notification.dispatch_due`：派发已经到期的通用流程通知。
- `notification.in_app.send`：按任务执行时的组织数据向启用用户发送站内信。

建议为这两个任务分别建立定义，例如：

```json
{
  "handlerKey": "workflow.notification.dispatch_due",
  "params": { "limit": 100 }
}
```

站内信任务支持全部用户、指定部门和指定用户。指定部门包含当前所有子部门，例如：

```json
{
  "handlerKey": "notification.in_app.send",
  "params": {
    "title": "系统维护通知",
    "content": "今日 20:00 开始系统维护。",
    "scope": "departments",
    "departmentIds": [10, 11]
  }
}
```

任务保存接收范围规则，执行时重新解析当前启用用户；同一运行 ID 重试不会重复投递。创建或编辑该任务的管理员除定时任务管理权限外，还必须具有 `notification:send` 权限。运行日志只记录计划、发送、跳过和重放数量，不记录通知正文或完整收件人列表。

## 4. 安全配置

HTTP、Shell 和 SQL 都必须先在服务端配置，再由有独立高风险权限的管理员创建任务：

- HTTP 仅允许 `scheduled_task.http.allowed_hosts`/`allowed_cidrs` 中的目标；默认拒绝私网、回环、链路本地和云元数据地址。
- 敏感 HTTP 请求头从 `scheduled_task.http.credentials` 的服务端引用解析，不写进任务定义。
- Shell 默认关闭，只能执行 `shell_commands` 注册的绝对路径程序，参数和环境变量受白名单约束。
- SQL 默认关闭，只能使用 `sql_data_sources` 注册的数据源；普通 SQL 经 AST 校验并限制返回行数或影响行数。

生产环境不要在可被普通管理员读取的配置文件中保存凭据。应通过受控配置挂载或密钥管理系统生成最终配置。

## 5. 执行语义

MySQL 是唯一事实源，Redis 只承担投递和节点心跳。平台采用至少一次投递：

- Redis 暂时不可用时，运行记录保留为未投递，scheduler 恢复后重新发布；Redis 重启导致 Consumer Group 丢失时，worker 会自动重建。
- 重复 Redis 消息通过 MySQL 状态抢占消除常规重复执行。
- 外部 HTTP、Shell 或跨库 SQL 可能在“副作用已发生、结果未落库”时再次执行，任务应使用 `X-Scheduled-Run-ID` 或业务唯一键保证幂等。
- `waiting`、`queued`、`retry_wait` 可直接取消；`running` 写入取消请求，由 worker 取消处理器 Context。
- Worker 心跳超时后，scheduler 按任务快照的重试上限恢复为等待重试或失败。

流程发起处理器使用当前项目中的已发布通用流程定义，并从组织用户树选择一个或多个发起人：

- 配置保存流程定义 ID、版本策略、`starterIds`、流程变量和初始表单数据；最多选择 100 名发起人。
- 每次定时运行会为每名发起人创建独立流程实例，业务键为 `<run_id>:user:<user_id>`。
- 同一运行重试时，工作流应用服务按业务键返回已经创建的实例，不重复发起成功用户的流程。
- 历史任务中的单值 `starterId` 仍可执行，并继续使用原 `<run_id>` 业务键保证升级前后幂等。
- 创建或编辑流程类定时任务的管理员还需具备 `workflow:instance:start` 权限，才能读取已发布流程和组织用户选项。

启动和运行期间都会将临时数据库或 Redis 错误视为可恢复故障，按 1 秒起步、最大 30 秒的指数退避继续重试；单次连接或查询超时不会终止 `taskd`。启动等待可由 `SIGINT`/`SIGTERM` 正常中断。`role=all` 下 scheduler 和 worker 独立监督，其中一个组件异常重启时不会取消另一个组件。

扫描频率分为两组：

- `scheduler_poll_seconds`：到期任务、重试任务和串行等待任务扫描，默认 2 秒。
- `scheduler_recovery_seconds`：未投递消息和失联 Worker 恢复扫描，默认 30 秒。

可通过 `WECHECKIN_SCHEDULED_TASK_SCHEDULER_POLL_SECONDS` 和 `WECHECKIN_SCHEDULED_TASK_SCHEDULER_RECOVERY_SECONDS` 覆盖。查询仍保留独立超时，网络恢复后由连接池重新建立连接并继续扫描，不建议仅通过调大查询超时掩盖连接问题。

`taskd` 的 GORM 日志级别固定为 `Warn`，正常轮询和空结果查询不会逐条打印；慢 SQL、数据库错误以及运行时退避/恢复事件仍会记录。HTTP 服务的数据库日志级别不受此设置影响。

## 6. 排查顺序

1. 后台“执行节点”确认 worker 心跳、并发槽位和版本。
2. 后台“运行记录”查看状态、错误分类和分段日志。
3. 检查 MySQL 运行记录是否为 `queued` 且 `redis_message_id` 为空。
4. 检查 Redis Consumer Group 和 Pending 消息。
5. 检查 `taskd` 日志中的数据库、Redis、处理器配置和权限错误。

修改 Cron 或时区前使用后台预览。秒级 Cron 的最小间隔由 `minimum_second_interval` 限制，默认 5 秒。
