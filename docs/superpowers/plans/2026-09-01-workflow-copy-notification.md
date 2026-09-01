# 通用工作流抄送与通知实施计划

> 设计基线：`docs/superpowers/specs/2026-09-01-workflow-copy-notification-design.md`

## 实施约束

- 只扩展独立通用 workflow 模块和 admin 流程设计/实例页面，不修改绩效流程。
- `h5app` 本期保持不变，工作流中心和深链属于第二期。
- 当前工作区的工作流文件已有大量未提交改动，必须基于现状增量编辑，禁止还原或覆盖。
- 计划文档单独提交；生产代码实施阶段不自动分步提交，避免把用户原有未提交改动一起纳入提交。
- 按 TDD 顺序执行：先增加会失败的定向测试，确认失败原因，再写最小实现并复跑。

## 目标架构

- 流程定义为 `cc`、`notify`、`approval`、`handle` 提供统一 `notification` 配置。
- 领域引擎只产生 `Participant` 和 `NotificationIntent`，不依赖站内通知或钉钉。
- GORM 存储在流程事务内保存参与人，并把通知意图展开成每接收人、每渠道一条 Outbox。
- 应用服务在事务提交后调用通用 dispatcher；渠道失败只改变 Outbox 状态，不回滚流程。
- 站内渠道写入现有 `notify` 表；钉钉渠道复用 `DingTalkWorkNotificationClient`，不依赖绩效通知包。
- 客户端实例查询支持 `started`、`handled`、`copied`，抄送人仅获得只读访问权。

## Task 1：流程定义、校验与 BPMN 契约

**文件：**

- 修改：`backend/internal/workflow/types.go`
- 修改：`backend/internal/workflow/validation.go`
- 修改：`backend/internal/workflow/definition_test.go`
- 修改：`backend/internal/workflow/compiler.go`

**步骤：**

1. 在 `definition_test.go` 增加失败测试，覆盖：
   - `notify` 节点必须配置有效 `assignee` 和启用的通知配置。
   - 只允许 `in_app`、`dingtalk_oa` 渠道，渠道不能为空或重复。
   - 标题、正文模板不能为空且不能超过原文长度上限。
   - 未知 `{{token}}` 校验失败。
   - `cc`、`approval`、`handle` 未配置 `notification` 时仍兼容通过。
2. 运行：

   ```bash
   cd backend
   GOCACHE=$PWD/.cache/go-build go test ./internal/workflow -run 'TestValidate.*Notification|TestCompile.*Notify' -count=1
   ```

   预期：因缺少 `NodeTypeNotify`、`NotificationConfig` 和校验而失败。
3. 在 `types.go` 增加：
   - `NodeTypeNotify`。
   - `NotificationChannelInApp`、`NotificationChannelDingTalkOA`。
   - `NotificationConfig{Enabled, Channels, Title, Content}`。
   - `Node.Notification`。
   - 独立通知校验错误码。
4. 在 `validation.go` 增加模板白名单解析和节点级通知校验；`notify` 强制启用，其他节点保持可选。
5. 在 `compiler.go` 把 `notify` 输出为 `serviceTask`，并让 `cc`、`notify`、人工任务通过 Flowable 扩展属性保留通知配置。
6. 复跑定向测试和完整 `./internal/workflow` 测试。

## Task 2：领域参与人和通知意图

**文件：**

- 修改：`backend/internal/modules/workflow/domain/runtime.go`
- 修改：`backend/internal/modules/workflow/domain/engine.go`
- 修改：`backend/internal/modules/workflow/domain/engine_test.go`

**步骤：**

1. 先写领域失败测试：
   - `cc` 生成去重的 `Participant{Role: cc}`、`node_cc` 历史和通知意图，并继续流转。
   - `notify` 只生成 `node_notify` 历史和通知意图，不生成参与人或任务。
   - 办理任务和并行审批的每个 `pending` 任务生成 `task_arrived` 意图。
   - 顺序审批初始只通知第一人，后续任务激活时再通知对应处理人。
   - 未启用通知时不产生意图。
2. 运行：

   ```bash
   cd backend
   GOCACHE=$PWD/.cache/go-build go test ./internal/modules/workflow/domain -run 'Test.*(CC|Notify|TaskArrival)' -count=1
   ```

3. 在领域模型增加：
   - `ParticipantRoleCC`、`Participant`。
   - `NotificationKindNodeCC`、`NotificationKindNodeNotify`、`NotificationKindTaskArrived`。
   - `NotificationIntent`，携带实例、节点、任务、接收人、流程名称和通知配置快照。
   - `State.Participants`、`State.NotificationIntents`。
   - `HistoryNodeNotify`。
4. 扩展 `advanceToken` 支持 `notify`，把 CC、通知节点和任务激活的副作用统一写入领域状态。
5. 复跑领域定向测试和整个 domain 包测试。

## Task 3：参与人、Outbox 与权限迁移

**文件：**

- 新增：`backend/migrations/20260901110000_add_workflow_notifications.sql`
- 新增：`backend/test/internal/bootstrap/migrations/workflow_notification_migration_test.go`
- 修改：`backend/internal/model/workflow/runtime.go`
- 修改：`backend/internal/model/aliases.go`
- 修改：`backend/internal/bootstrap/database_safety_test.go`

**步骤：**

1. 先写迁移结构测试，要求 SQL 包含：
   - `workflow_instance_participants` 及 `(instance_id,user_id,participant_role,node_id)` 唯一键。
   - `workflow_notification_outbox`、唯一 `dedupe_key`、状态/到期查询索引。
   - `workflow:notification:list`、`workflow:notification:retry` 权限。
   - 不写入 `permission_grants`，不自动授权角色。
2. 运行：

   ```bash
   cd backend
   GOCACHE=$PWD/.cache/go-build go test ./test/internal/bootstrap/migrations -run TestWorkflowNotificationMigration -count=1
   ```

3. 新增 SQL 和对应 GORM 模型 `InstanceParticipant`、`NotificationOutbox`，常量覆盖角色、类型、渠道和状态。
4. 保持运行时表由版本化 SQL 管理，不把两个新表加入 `autoMigrate`；增加安全测试锁定这一约束。
5. 复跑迁移与 bootstrap 定向测试。

## Task 4：事务内持久化与抄送查询权限

**文件：**

- 修改：`backend/internal/modules/workflow/application/types.go`
- 修改：`backend/internal/modules/workflow/application/service.go`
- 修改：`backend/internal/modules/workflow/application/service_test.go`
- 修改：`backend/internal/modules/workflow/infrastructure/gorm_store.go`
- 修改：`backend/internal/modules/workflow/infrastructure/mapping.go`
- 修改：`backend/internal/modules/workflow/infrastructure/mapping_test.go`

**步骤：**

1. 扩展应用层 fake store 的失败测试，覆盖：
   - 创建/保存状态后必须在同一事务调用 `PersistEffects`。
   - `PersistEffects` 失败会让启动或任务完成整体失败。
   - 成功时返回本次新增 Outbox ID 供提交后投递。
   - `scope=copied` 设置参与人过滤，`scope=handled` 设置任务参与过滤。
   - `GetMyInstance` 允许 CC 参与人查看，但任务处理权限不受影响。
2. 在 `TransactionStore` 增加 `PersistEffects`，在 `Store` 增加 `HasParticipant`；为 `InstanceQuery` 增加 `ScopeUserID`、`Scope`。
3. GORM 实现：
   - 参与人使用唯一键 `OnConflict DoNothing`。
   - 通知意图按渠道展开并生成稳定 `dedupe_key`，模板在入队时渲染为 payload 快照。
   - 解析发起人显示名称，缺失时回退到本地用户 ID。
   - `copied` 和 `handled` 使用 `EXISTS` 子查询，避免 JOIN 造成实例重复和分页总数错误。
4. `GetMyInstance` 先保持原发起人/任务检查，再通过 `HasParticipant(..., cc)` 放行只读详情。
5. 复跑 application 和 infrastructure 定向测试。

## Task 5：通用 Outbox dispatcher 与渠道适配器

**文件：**

- 新增：`backend/internal/modules/workflow/application/notification.go`
- 新增：`backend/internal/modules/workflow/application/notification_test.go`
- 新增：`backend/internal/modules/workflow/infrastructure/notification_store.go`
- 新增：`backend/internal/modules/workflow/infrastructure/notification_channels.go`
- 新增：`backend/internal/modules/workflow/infrastructure/notification_dispatcher_test.go`
- 复用：`backend/internal/service/dingtalkh5/config/dingtalk_oapi.go`

**步骤：**

1. 先写 dispatcher 失败测试，使用 fake repository 和 fake channel 覆盖：
   - `pending`/到期 `failed` 抢占为 `sending`。
   - 成功变为 `sent`，失败按 1 分钟、5 分钟、30 分钟、2 小时、12 小时退避。
   - 第五次失败变为 `dead`。
   - 超过 10 分钟的 `sending` 可重新抢占。
   - 单条人工重试重置状态并立即发送。
   - 两个渠道独立成功或失败。
2. 定义应用层 `NotificationDispatcher`、查询 DTO 和 repository/channel 端口，不暴露 GORM 模型到 transport。
3. GORM notification store 实现列表、抢占、标记成功/失败、到期扫描和人工重置。
4. 站内渠道在数据库事务中：
   - 锁定 Outbox。
   - 若未发送则插入 `model.Notify`。
   - 同事务更新 Outbox 为 `sent`，保证进程中断时不会重复站内消息。
5. 钉钉渠道：
   - 根据发起人和接收人的有效绑定确定并固定 `corp_id`。
   - 按企业和相同 payload 合并接收人。
   - 调用 `DingTalkWorkNotificationClient.SendWorkNotificationContext`。
   - 使用企业 `AppURL`，来源名为通用流程，不引用绩效通知包。
6. 运行：

   ```bash
   cd backend
   GOCACHE=$PWD/.cache/go-build go test ./internal/modules/workflow/application ./internal/modules/workflow/infrastructure -run 'Test.*Notification' -count=1
   ```

## Task 6：应用服务提交后投递与客户端范围查询

**文件：**

- 修改：`backend/internal/modules/workflow/application/service.go`
- 修改：`backend/internal/modules/workflow/application/service_test.go`
- 修改：`backend/internal/modules/workflow/transport/httpclient/handler.go`
- 修改：`backend/internal/modules/workflow/transport/httpclient/handler_test.go`
- 修改：`backend/internal/routes/v2/client/routes.go`

**步骤：**

1. 先写失败测试：
   - 启动、完成任务、推进定时器提交后只投递本次新 Outbox。
   - dispatcher 发送失败不改变流程接口成功结果。
   - 客户端 `scope` 只接受 `started`、`handled`、`copied`，默认 `started`。
   - 用户 ID 只取认证上下文。
2. 为 Service 增加带 dispatcher 的构造入口；保留现有构造器作为无通知兼容入口，避免影响不相关测试和调用方。
3. 在事务闭包外调用即时投递，严禁在数据库事务内访问钉钉。
4. 客户端实例列表透传并校验 `scope`，详情继续调用只读的 `GetMyInstance`。
5. 在客户端路由中装配真实 dispatcher。
6. 复跑 application、httpclient 和 client routes 定向测试。

## Task 7：管理接口、路由与权限

**文件：**

- 修改：`backend/internal/modules/workflow/transport/httpadmin/handler.go`
- 修改：`backend/internal/modules/workflow/transport/httpadmin/handler_test.go`
- 修改：`backend/internal/routes/v2/admin/routes.go`
- 修改：`backend/internal/middleware/admin/route_permissions.go`
- 修改：`backend/internal/middleware/admin/permission_test.go`
- 修改：`backend/internal/routes/v2/swagger/swagger.go`
- 生成更新：`backend/docs/swagger/docs.go`
- 生成更新：`backend/docs/swagger/swagger.json`
- 生成更新：`backend/docs/swagger/swagger.yaml`

**步骤：**

1. 先写 handler 和权限失败测试，覆盖查询参数、路由 ID 优先、认证管理员、单条重试和到期投递。
2. 新增：
   - `GET /api/v2/admin/workflow-notifications`
   - `POST /api/v2/admin/workflow-notifications/dispatch-due`
   - `POST /api/v2/admin/workflow-notifications/:id/retry`
3. `dispatch-due` 路由必须在 `:id/retry` 参数路由前注册，避免静态路径被参数吞掉。
4. 权限映射：查询使用 `workflow:notification:list`，两个投递动作使用 `workflow:notification:retry`。
5. 管理和客户端路由都使用同一种 dispatcher 装配方法，避免渠道配置漂移。
6. 运行 httpadmin、middleware、routes 和 Swagger 定向测试；按项目现有方式重新生成 Swagger。

## Task 8：管理端设计器通知节点和配置

**文件：**

- 修改：`admin/src/views/workflow/types.ts`
- 修改：`admin/src/views/workflow/designer/graph.ts`
- 修改：`admin/src/views/workflow/designer/components/FlowInsertButton.vue`
- 修改：`admin/src/views/workflow/designer/components/NodeInspector.vue`
- 修改：`admin/src/views/workflow/designer/components/WorkflowNodeCard.vue`
- 修改：`admin/scripts/check-workflow-tree.mjs`

**步骤：**

1. 先扩展 `check-workflow-tree.mjs`，要求：
   - 节点联合类型包含 `notify`。
   - 插入菜单分别显示“抄送”和“通知”。
   - 新建四类相关节点具有规格约定的默认通知配置。
   - `notify` 可插入、删除并保留流程连线。
   - NodeInspector 有通知开关、渠道复选、标题和正文配置。
2. 运行 `npm run check:workflow-tree`，确认因实现缺失失败。
3. 增加 `WorkflowNotificationConfig` 和渠道类型，`cloneDraft` 保持旧定义缺省配置不被自动开启。
4. 更新 `graph.ts` 默认节点工厂：
   - 新审批/办理默认任务到达通知。
   - 新 CC 默认通知。
   - 新 notify 强制通知。
5. 在检查器复用一块通知配置 UI；使用复选框表示渠道、开关表示可选通知，不为 `notify` 展示关闭开关。
6. 节点卡片显示真实节点名称和渠道摘要，新增独立 `notify` 色彩但保持现有紧凑尺寸。
7. 复跑 `check:workflow-tree`。

## Task 9：管理端投递记录与重试界面

**文件：**

- 修改：`admin/src/api/index.ts`
- 修改：`admin/src/views/workflow/types.ts`
- 修改：`admin/src/views/workflow/instances/index.vue`
- 修改：`admin/scripts/check-workflow-runtime-pages.mjs`

**步骤：**

1. 先扩展静态回归检查，要求 API、权限控制、状态列表、单条重试和到期投递入口存在，并锁定 `node_cc` 文案为“已记录抄送”。
2. 新增 API：通知列表、单条重试、到期投递。
3. 在实例详情增加不嵌套卡片的“通知投递”区，展示接收人、类型、渠道、状态、尝试次数、错误和时间。
4. 只在具备 `admin:menu:workflow:notification:retry` 时显示重试动作；列表查看使用对应 list 菜单权限。
5. 对 `failed`、`dead` 提供单条重试；工具栏提供一次“投递到期通知”，执行后刷新当前实例记录。
6. 修正历史事件：`node_cc` 为“已记录抄送”，新增 `node_notify` 为“通知节点已触发”。
7. 运行：

   ```bash
   cd admin
   npm run check:workflow-tree
   npm run check:workflow-runtime-pages
   ```

## Task 10：集成验证与差异审计

**文件：**

- 核对所有本计划涉及文件，不修改 `h5app` 和绩效目录。

**步骤：**

1. 格式化新增/修改 Go 文件：

   ```bash
   cd backend
   gofmt -w <本次修改的 Go 文件>
   ```

2. 后端定向验证：

   ```bash
   cd backend
   GOCACHE=$PWD/.cache/go-build go test ./internal/workflow ./internal/modules/workflow/... ./internal/middleware/admin ./internal/routes/v2/... ./test/internal/bootstrap/migrations -count=1
   ```

3. 后端完整测试：

   ```bash
   cd backend
   GOCACHE=$PWD/.cache/go-build go test ./... -count=1
   ```

4. 管理端验证：

   ```bash
   cd admin
   npm run check:workflow-tree
   npm run check:workflow-runtime-pages
   npm run check:workflow-form-designer
   npm run build
   ```

5. 差异审计：
   - `git diff --check` 无空白错误。
   - `git status --short` 中 `h5app` 状态与实施前一致。
   - `rg` 确认 workflow 新代码没有导入 `dingtalkh5/performance`。
   - 核对没有回滚用户原有未提交文件，没有修改绩效业务流程。
6. 向用户报告自动化结果和仍需手工验证的钉钉真实发送、管理端页面交互；不宣称已完成第二期 H5 页面。
