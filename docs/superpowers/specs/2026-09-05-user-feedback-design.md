# 用户反馈模块设计

## 1. 目标

新增独立、通用的用户反馈能力：

- H5App 增加顶层“我的反馈”菜单，普通用户可以查看状态统计和反馈列表、新建反馈、查看详情，并在处理中继续补充文字或图片。
- Admin 增加“用户反馈”总览，管理员可以筛选和查看反馈、更新处理状态，并选择是否向提交人发送站内信。
- 站内信点击后直接打开对应反馈详情，形成“反馈处理状态变化 -> 用户收到通知 -> 返回详情查看”的闭环。

本次新增 `userfeedback` 独立业务模块，不放入问卷、流程或绩效模块，不修改现有绩效业务流程。图片存储和站内信投递复用项目已有通用基础设施。

## 2. 范围与界限

### 2.1 包含

- 用户反馈领域模型、状态机、应用服务、存储实现和 Admin/H5App HTTP 接口。
- 用户反馈、反馈消息、附件和每日编号序列表的版本化 SQL 迁移。
- H5App 顶层菜单、状态统计、列表、新建、详情和补充反馈交互。
- Admin 状态统计、筛选列表、详情抽屉和状态更新对话框。
- 图片上传限制、真实 MIME 校验、对象清理和媒体 URL 输出。
- Admin 菜单/API 权限、H5App 菜单/API 权限及现有角色回填。
- 基于现有通用通知出站模块和 `taskd` 的异步站内信投递。
- Swagger、前后端类型、契约检查、自动化测试和浏览器验证。

### 2.2 不包含

- 不新增用户之间的即时聊天或管理员自由回复会话。
- 不新增短信、邮件、钉钉 OA 或群机器人通知。
- 不支持普通用户修改或删除已经提交的历史内容。
- 不支持物理删除反馈；“已关闭”承担归档语义。
- 不修改现有问卷反馈、流程通知和绩效业务的数据结构或行为。

## 3. 总体架构

Backend 新增 `internal/modules/userfeedback`：

- `domain`：状态枚举、状态流转规则、消息类型和稳定领域错误。
- `application`：权限上下文、校验、事务、幂等、编号生成、图片协作和通知出站编排。
- `infrastructure`：GORM 查询与持久化、统计聚合和对象存储适配。
- `transport/httpadmin`：Admin DTO、参数解析、权限边界和统一响应。
- `transport/httpdingtalkh5`：H5App DTO、当前用户归属约束和统一响应。

HTTP handler 不直接执行多表写入。领域状态流转不依赖 Hertz、GORM、Redis 或对象存储。Admin 与 H5App 共用同一 application service，但使用不同的用例入口和授权上下文。

图片继续使用 `internal/support/storage`。为保证数据库失败后可以清理已经上传的对象，需要为本地存储和阿里云对象存储补齐统一的删除能力；业务模块只依赖抽象接口，不判断具体存储供应商。

状态通知写入现有 `notification_outbox`，由独立 `taskd` worker 投递到 `notify` 表。HTTP 服务只提交业务事务和出站记录，不同步执行站内信投递，不占用额外 HTTP 请求时间。

## 4. 状态与流转

反馈状态固定为：

| 状态 | 显示名称 | 语义 | H5App 是否允许补充 |
| --- | --- | --- | --- |
| `pending` | 待处理 | 已提交，尚未开始处理 | 是 |
| `processing` | 处理中 | 管理员已接手或重新打开 | 是 |
| `resolved` | 已解决 | 已给出处理结论，等待归档 | 否 |
| `closed` | 已关闭 | 已归档，不再处理 | 否 |

允许的状态流转：

```text
pending -> processing
pending -> closed
processing -> resolved
resolved -> closed
resolved -> processing
closed -> processing
```

- `pending -> closed` 表示无需继续处理，处理说明必填。
- `resolved -> processing` 和 `closed -> processing` 表示重新打开，处理说明必填。
- `processing -> resolved`、`resolved -> closed` 的处理说明必填。
- 不允许跳过处理中直接从 `pending` 变为 `resolved`。
- 不允许提交与当前状态相同的更新。
- 所有状态变更都追加处理记录，不覆盖历史记录。
- 主表保留当前状态、当前处理人、最后更新时间等查询快照，详情以消息时间线为完整事实记录。

状态颜色在 Admin 与 H5App 保持一致：待处理和处理中使用黄色系，已解决使用绿色系，已关闭使用灰色系。

## 5. 数据模型

### 5.1 `user_feedbacks`

反馈主表保存列表和统计需要的当前快照：

- `id`：主键。
- `feedback_no`：用户可见编号，唯一，例如 `FB-20260905-0001`。
- `submitter_id`：提交人的本地用户 ID，建立列表索引。
- `create_request_id`：新建请求幂等键，与提交人组成唯一索引。
- `status`：当前状态。
- `handler_id`：最近一次处理该反馈的管理员 ID，可空。
- `version`：乐观锁版本号，每次状态变化或用户补充后递增。
- `last_activity_at`：最近一次原始提交、补充或状态变化时间。
- `resolved_at`、`closed_at`：对应状态首次到达时间，可空；重新打开时不删除历史消息，当前快照字段按状态规则更新。
- `created_at`、`updated_at`。

不在主表复制完整用户名。列表和详情按本地用户 ID 查询当前用户显示名；反馈编号和历史消息保证用户改名后记录仍可追踪。

### 5.2 `user_feedback_messages`

每一条内容或处理动作保存为独立消息：

- `id`、`feedback_id`。
- `message_type`：`initial`、`supplement` 或 `status`。
- `author_type`：`user` 或 `admin`。
- `author_id`：本地用户或管理员 ID。
- `content`：原始文字、补充文字或处理说明；无文字的图片补充允许为空。
- `from_status`、`to_status`：仅状态消息使用。
- `request_id`：客户端请求幂等键。
- `created_at`。

唯一索引按业务入口设置：

- 新建反馈：主表 `submitter_id + create_request_id` 唯一。
- 用户补充：`feedback_id + author_type + author_id + request_id` 唯一。
- 状态更新：`feedback_id + author_type + author_id + request_id` 唯一。

同一请求重试返回首次提交结果，不重复创建消息、附件或通知。

### 5.3 `user_feedback_attachments`

- `id`、`feedback_id`、`message_id`。
- `storage_provider`、`object_key`。
- `original_name`、`content_type`、`size_bytes`、`sort_order`。
- `created_at`。

数据库只保存对象键和必要元数据，不保存带域名的完整 URL。响应 DTO 通过现有媒体地址构造能力动态生成可访问 URL。

### 5.4 `user_feedback_daily_sequences`

- `sequence_date`：日期主键。
- `current_value`：当日最新序号。

新建反馈事务内锁定当日序列并递增，生成 `FB-YYYYMMDD-NNNN`。超过四位时允许自然扩展，不截断或循环，避免编号碰撞。

所有表通过 `backend/migrations/` 下的版本化 SQL 创建，不依赖服务启动时 AutoMigrate。

## 6. 图片规则与一致性

### 6.1 输入限制

- 首次提交文字必填，去除首尾空白后最长 5000 字符。
- 首次提交最多 6 张图片。
- 用户补充可以只有文字、只有图片，或两者同时存在，但不能同时为空。
- 每次补充最多 6 张图片；单条反馈累计最多 30 张图片。
- 单张图片最大 10 MB。
- 只允许 JPG、PNG、WebP。
- 后端同时检查扩展名、声明的 MIME 和文件头识别出的真实 MIME，不信任客户端 `Content-Type`。
- multipart 请求设置总请求体上限，计算字段和全部图片后的最大允许大小，超限在进入存储前拒绝。

### 6.2 保存流程

新建和补充采用以下顺序：

1. 解析并校验文字、请求 ID、文件数量、文件大小和真实类型。
2. 查询已有附件数量；补充操作在事务内再次锁定反馈并复核状态和总数。
3. 将合法图片写入 `uploads/feedback/<yyyy>/<mm>/...` 对应对象键。
4. 开启数据库事务，写入反馈、消息、附件和幂等信息。
5. 数据库提交失败时，通过统一存储删除接口尽力清理本次已上传对象。
6. 清理失败只记录请求 ID、反馈业务键和对象键，不记录图片内容或用户文字；后续维护任务可以按无数据库引用的对象键清理孤儿文件。

对象存储与数据库无法构成同一 ACID 事务，因此“数据库事务 + 失败补偿删除 + 可审计孤儿清理”是本次的一致性边界。接口只有在数据库提交成功后才返回成功。

## 7. H5App 接口

接口前缀：`/api/v2/dingtalk/h5/user-feedbacks`。

### 7.1 状态统计

```text
GET /overview
Permission: dingtalk_h5:api:feedback:list
```

只对当前登录用户的数据执行一次条件聚合，返回：

```json
{
  "pending": 1,
  "processing": 2,
  "resolved": 3,
  "closed": 4
}
```

不加载列表记录，不接受客户端传入用户 ID。

### 7.2 列表

```text
GET /?page=&pageSize=&status=&keyword=
Permission: dingtalk_h5:api:feedback:list
```

只返回当前用户自己的反馈。列表项包含 `id`、`feedbackNo`、`summary`、`status`、`imageCount`、`lastActivityAt` 和 `createdAt`。

### 7.3 新建

```text
POST /
Content-Type: multipart/form-data
Permission: dingtalk_h5:api:feedback:create
```

字段为 `content`、`requestId` 和重复的 `images`。`requestId` 在进入新建页面时生成，同一次提交和网络重试复用。每个用户自然日最多新建 20 条反馈，超过后返回稳定业务错误；限制不影响已有反馈的补充。

### 7.4 详情

```text
GET /:id
Permission: dingtalk_h5:api:feedback:detail
```

详情包含主表快照、提交人显示名、全部消息时间线和每条消息的附件。查询必须同时约束 `id` 和当前用户 ID；不存在和不属于当前用户统一返回未找到。

### 7.5 补充

```text
POST /:id/supplements
Content-Type: multipart/form-data
Permission: dingtalk_h5:api:feedback:supplement
```

字段为 `content`、`requestId`、`version` 和重复的 `images`。仅 `pending`、`processing` 可补充。事务内锁定反馈后再次检查状态、版本和累计附件数量，避免管理员状态更新与用户补充并发时越过限制。

## 8. Admin 接口

接口前缀：`/api/v2/admin/user-feedbacks`。

### 8.1 状态统计

```text
GET /overview
Permission: user-feedback:list
```

根据当前筛选权限范围执行一次条件聚合，返回四个状态数量，不读取列表记录。

### 8.2 列表

```text
GET /?page=&pageSize=&status=&submitterId=&handlerId=&keyword=&submittedFrom=&submittedTo=
Permission: user-feedback:list
```

关键词匹配反馈编号和消息文字。返回分页结果，不为每行执行独立的用户、附件或消息查询。

### 8.3 详情

```text
GET /:id
Permission: user-feedback:list
```

返回反馈快照、提交人、处理人、完整消息时间线和附件，供详情抽屉使用。

### 8.4 更新状态

```text
PATCH /:id/status
Permission: user-feedback:handle
```

请求体：

```json
{
  "status": "resolved",
  "note": "问题已修复，请刷新后重试。",
  "notifyUser": true,
  "version": 3,
  "requestId": "uuid"
}
```

- `notifyUser` 默认 `true`，管理员可以关闭。
- 处理说明按第 4 节状态规则校验。
- `version` 用于检测页面数据过期；冲突时返回稳定的并发更新错误并要求刷新。
- 事务内锁定反馈、校验流转、追加状态消息、更新主表快照；需要通知时同时写入通知出站记录。
- 重复 `requestId` 返回首次操作结果，不重复变更状态或发送站内信。

## 9. 通知投递与详情跳转

状态更新对话框包含目标状态、处理说明和“发送站内信”开关，开关默认开启。开启后写入现有通用通知出站：

- `channel=internal`
- `type=feedback_status`
- `sourceType=user_feedback`
- `sourceId=反馈 ID`
- `recipientUserId=提交人本地用户 ID`
- `title=反馈 <feedbackNo> <状态名称>`
- `content=处理说明`
- 幂等键：`user-feedback-status:<feedbackId>:<statusMessageId>`

现有通知出站 `MessagePayload` 需要增加可选 `NotificationType` 字段。未提供时保持原默认类型，避免影响已有流程、定时任务和管理员手动站内信；用户反馈明确传入 `feedback_status`。

状态变更、状态消息和通知出站记录在同一数据库事务中提交。事务成功后即表示处理状态可靠保存，`taskd` 独立拉取出站消息并按现有重试策略投递。短暂的 Redis、数据库或 worker 故障只会延迟通知，不会丢失已提交的出站记录，也不会让 HTTP 服务同步等待通知投递。

H5App 通知面板识别 `sourceType=user_feedback`，点击后：

1. 标记通知已读。
2. 关闭通知面板。
3. 打开或聚焦动态页签 `feedback:detail:<id>`。
4. 详情接口仍执行当前用户归属校验，不能依赖通知内容越权查看。

前后端为 `feedback_status` 增加统一的图标与颜色映射；未知通知类型继续使用现有默认样式。

## 10. 权限与迁移

### 10.1 Admin

- 顶层菜单：`admin:menu:user-feedback`，路由 `/user-feedbacks`。
- 查看权限：`user-feedback:list`，覆盖总览、列表和详情。
- 处理权限：`user-feedback:handle`，覆盖状态更新及该次更新中的可选站内信。

前端按钮显示和事件处理使用同一权限判断；后端路由权限是最终安全边界。菜单、API 分组、method + path 声明和已有角色授权通过 SQL 迁移同步。

### 10.2 H5App

- 菜单：`dingtalk_h5:menu:feedback`。
- 列表与统计：`dingtalk_h5:api:feedback:list`。
- 新建：`dingtalk_h5:api:feedback:create`。
- 详情：`dingtalk_h5:api:feedback:detail`。
- 补充：`dingtalk_h5:api:feedback:supplement`。

迁移为现有普通 H5App 角色回填上述菜单和 API 权限，使功能上线后默认可见，同时保留后台按角色关闭入口或具体操作的能力。所有 H5App 接口仍要求有效钉钉 H5 登录态，并从登录态读取唯一的本地用户 ID。

## 11. H5App 交互

### 11.1 “我的反馈”主页

- 作为顶层应用菜单进入，不放入头像下拉菜单。
- 顶部显示待处理、处理中、已解决、已关闭四个紧凑统计项；点击统计项筛选列表并刷新统计。
- 列表按最近活动时间倒序，显示反馈编号、文字摘要、状态、图片数量和最近更新时间。
- 页面提供“新建反馈”按钮，打开动态内部页 `feedback:create`。
- 进入页面、切回浏览器前台、新建或补充成功、点击统计项时刷新统计。
- 提供加载、空数据、失败、分页加载和重试状态。

### 11.2 新建反馈

- 使用完整动态内部页，不使用空间不足的弹窗。
- 提供多行文字、图片选择/预览/移除和提交按钮。
- 显示已选图片数量与单张上传错误，不显示实现说明文字。
- 请求进行中禁用重复提交；失败保留已填写内容和仍可用的本地图片。
- 页面有未保存内容时关闭或切换页签需要确认。
- 成功后关闭新建页，刷新主页并打开新反馈详情。

### 11.3 反馈详情

- 动态页签键为 `feedback:detail:<id>`；重复打开同一反馈时聚焦已有页签并刷新。
- 顶部显示反馈编号、当前状态、提交时间和最近更新时间。
- 按时间顺序显示原始反馈、用户补充和管理员处理记录；图片支持预览。
- `pending`、`processing` 显示补充入口；`resolved`、`closed` 只读。
- 补充编辑区同样具有未保存关闭保护。

## 12. Admin 交互

- 新增 `/user-feedbacks` 页面，使用 `AdminPageShell`、`AdminPageHeader`、`AdminSearchBar` 和 `AdminTablePanel` 等现有规范组件。
- 顶部显示四个可点击的紧凑状态统计项，点击后设置状态筛选并刷新统计与列表。
- 筛选项包括反馈编号/用户、状态、处理人、提交时间和关键词。
- 表格列包括反馈编号、内容摘要与首张缩略图、提交人、状态、处理人、图片数量、最近更新时间、提交时间和查看操作。
- 查看操作打开右侧详情抽屉，展示原始反馈、图片、用户补充和状态时间线。
- 有 `user-feedback:handle` 权限时，抽屉底部固定显示“更新状态”按钮。
- 状态更新使用独立对话框，不嵌套抽屉；对话框挂载到 `body`，避免被后台固定导航或详情抽屉遮挡。
- 提交失败保留目标状态、说明和通知开关；版本冲突提示刷新详情后再处理。

## 13. 错误、安全与可观测性

- H5App 用户只能查询和补充自己的反馈；Admin 访问由明确权限控制。
- 不接受客户端提交 `submitterId`、`handlerId` 或通知接收人，均从认证上下文确定。
- 非法状态、越权访问、版本冲突、重复请求、上传限制和存储故障使用稳定业务错误，在 HTTP 边界统一映射。
- 图片文件名不直接作为对象键，下载/预览响应使用安全的内容类型和既有媒体访问策略。
- 搜索参数、分页大小和时间范围有后端上限；关键词查询使用参数化 SQL。
- 日志记录 request ID、反馈 ID/编号、状态变化、对象键和错误链，不记录反馈正文、图片内容、token 或完整用户数据。
- 新建反馈按用户自然日限制 20 条；限流判断由后端执行，前端提示只改善体验。
- 状态统计使用条件聚合；列表一次性预加载所需用户和附件计数，避免 N+1 查询。
- 不增加高频后台轮询。页面进入、浏览器恢复前台、操作成功和用户主动点击时刷新；站内信继续复用现有未读刷新机制。

## 14. 兼容性

- 新表、新路由、新权限和新通知类型均为增量能力，没有历史用户反馈数据需要转换。
- 现有通知出站 payload 新字段为可选，旧生产者和旧记录保持原行为。
- H5App 通知面板对未知 `sourceType` 保持现有展示方式，只有 `user_feedback` 增加详情跳转。
- 用户反馈附件使用独立对象前缀，不改变问卷、流程表单和其他上传目录。
- H5App 菜单由权限目录返回；未迁移权限时入口不可见，不影响其他应用菜单。

## 15. 测试与验证

### 15.1 Backend

- 领域测试：全部合法和非法状态流转、必填说明、补充允许状态。
- application 测试：新建、补充、详情归属、每日限额、幂等重放、乐观锁冲突和并发状态复核。
- 图片测试：扩展名/MIME/文件头不一致、大小、单次数量、累计数量、本地与阿里云删除补偿。
- 事务测试：反馈/消息/附件原子写入，状态/历史/出站原子写入，数据库失败后的对象清理。
- 通知测试：`feedback_status` payload、出站幂等键、`taskd` 内部渠道投递和用户收件人。
- HTTP 测试：H5App 用户隔离、Admin 权限、非法 multipart、分页筛选、状态统计只执行聚合。
- 契约测试：路由、权限目录、菜单目录、迁移 SQL、Swagger 和 DTO 一致。
- 运行：

  ```bash
  cd backend
  GOCACHE=$PWD/../.cache/go-build go test ./... -count=1
  ```

### 15.2 Admin

- API 类型、状态映射、筛选、统计卡片联动、详情抽屉和状态对话框测试。
- 权限检查同时覆盖按钮可见性和事件保护。
- 检查对话框层级、提交中状态、版本冲突和失败后表单保留。
- 运行 `npm run check:all` 和 `npm run build`。
- 在桌面和窄桌面宽度下进行浏览器验证。

### 15.3 H5App

- 菜单权限、API 类型、动态页签键、新建/详情/补充、未保存保护和通知跳转契约检查。
- 检查上传失败重试、图片预览、只读状态、状态筛选和前台恢复刷新。
- 运行 `pnpm lint`、`pnpm type-check`、现有模块契约检查及 `pnpm build:h5`。
- 在 PC 和手机宽度下进行浏览器验证；真实钉钉免登、图片选择和通知点击仍由用户在目标环境做最终验收。

## 16. 实施顺序

1. 增加数据库迁移、GORM 模型、状态领域规则和权限目录。
2. 完成 Backend application/infrastructure、图片删除能力和 H5App/Admin API。
3. 扩展通知出站类型并完成 `taskd` 投递与 H5App 详情深链。
4. 完成 H5App“我的反馈”主页、新建页和详情页。
5. 完成 Admin 总览、列表、详情抽屉和状态更新对话框。
6. 同步 Swagger、类型、权限回填、契约检查和开发文档。
7. 运行分模块自动化检查、全量 Backend 测试和 PC/移动浏览器验证。

## 17. 验收标准

1. 有 H5App 权限的用户能从顶层“我的反馈”进入，查看自己的四类状态统计和反馈列表。
2. 用户能提交必填文字和最多 6 张合法图片；重复提交不会产生重复反馈。
3. 用户能在待处理、处理中继续追加文字或图片，已解决、已关闭时详情只读。
4. 有处理权限的管理员能查看完整历史，并且只能按规定状态机更新状态。
5. 管理员可在每次状态更新时选择是否发送站内信，默认发送；通知失败只延迟投递，不回滚已成功的状态事务。
6. 用户点击反馈状态站内信后直接打开自己的反馈详情，不能借助反馈 ID 越权访问他人记录。
7. 图片类型、大小、数量、累计上限和失败清理均由后端强制执行。
8. Admin/H5App 菜单、API、按钮权限与 SQL 迁移一致，现有角色按设计获得默认授权。
9. 原有问卷、流程、绩效、站内信和定时任务行为保持不变。
