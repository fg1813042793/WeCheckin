# User Feedback Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在不修改现有绩效、问卷和流程业务行为的前提下，交付独立的用户反馈模块，使 H5App 用户可以提交文字和图片、查看与补充自己的反馈，Admin 可以总览、筛选、查看并更新反馈状态，状态变化可通过通用通知出站和 `taskd` 异步发送站内信并跳回反馈详情。

**Architecture:** Backend 新增 `internal/modules/userfeedback`，按 domain、application、infrastructure、transport 分层；多表写入通过 application 定义的事务 Store 接口完成，状态记录与通知 Outbox 在同一 GORM 事务内落库。附件复用通用 storage/media 能力，并补齐本地和阿里云对象删除补偿。Admin 使用现有 Admin UI 基础组件实现筛选列表、详情抽屉和状态对话框；H5App 使用现有权限菜单和动态页签体系实现反馈主页、新建与详情，不在 `pages.json` 注册业务动态页。

**Tech Stack:** Go、Hertz、GORM、MySQL、Redis/taskd、Vue 3、TypeScript、Element Plus、uni-app、uView Pro、Vite、pnpm/npm、Node 契约脚本。

---

## 实施约束

- 开始每个任务前执行 `git status --short`；当前工作区已有大量未提交的 Workflow/Admin/H5App 修改，只暂存本任务明确列出的文件。
- 不手工修改 `backend/docs/swagger/` 生成文件；完成 Swagger 注释后只运行仓库规定的生成命令。
- 新客户端接口使用 `/api/v2`，Admin 使用 `/api/v2/admin`，H5App 使用 `/api/v2/dingtalk/h5`。
- 数据库只通过 `backend/migrations/` 版本化 SQL 变更，不增加启动时 AutoMigrate。
- H5App 是当前钉钉 H5 的权威源；不修改已迁移或删除的旧钉钉 H5 目录。
- UI 文案写入现有 locale 资源；图标优先使用项目已有图标库。
- 每项提交前检查 `git diff --cached --name-status`，确认没有带入用户原有修改。

## Task 1: 建立数据库结构与权限迁移

**Files:**
- Create: `backend/migrations/20260905100000_create_user_feedback.sql`
- Create: `backend/migrations/20260905101000_add_user_feedback_permissions.sql`
- Create: `backend/test/internal/bootstrap/migrations/user_feedback_migration_test.go`

- [ ] **Step 1: 编写失败的迁移契约测试**

测试读取两份 SQL，断言四张表、唯一索引、查询索引、Admin 权限、H5App 权限和现有角色回填均存在：

~~~go
func TestUserFeedbackMigrationContainsSchemaAndPermissions(t *testing.T) {
    read := func(name string) string {
        source, err := os.ReadFile(filepath.Join("..", "..", "migrations", name))
        if err != nil {
            t.Fatal(err)
        }
        return string(source)
    }
    schema := read("20260905100000_create_user_feedback.sql")
    permissions := read("20260905101000_add_user_feedback_permissions.sql")

    for _, fragment := range []string{
        "CREATE TABLE IF NOT EXISTS `user_feedbacks`",
        "CREATE TABLE IF NOT EXISTS `user_feedback_messages`",
        "CREATE TABLE IF NOT EXISTS `user_feedback_attachments`",
        "CREATE TABLE IF NOT EXISTS `user_feedback_daily_sequences`",
        "uk_user_feedback_submitter_request",
        "uk_user_feedback_message_request",
    } {
        if !strings.Contains(schema, fragment) {
            t.Fatalf("schema migration missing %q", fragment)
        }
    }
    for _, key := range []string{
        "admin:menu:user-feedback", "user-feedback:list", "user-feedback:handle",
        "dingtalk_h5:menu:feedback", "dingtalk_h5:api:feedback:list",
        "dingtalk_h5:api:feedback:create", "dingtalk_h5:api:feedback:detail",
        "dingtalk_h5:api:feedback:supplement",
    } {
        if !strings.Contains(permissions, key) {
            t.Fatalf("permission migration missing %q", key)
        }
    }
}
~~~

- [ ] **Step 2: 运行测试并确认失败**

Run:

~~~bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./test/internal/bootstrap/migrations -run UserFeedback -count=1
~~~

Expected: FAIL，提示迁移文件不存在。

- [ ] **Step 3: 创建数据表迁移**

实现以下数据库约束：

- `user_feedbacks`：业务编号唯一、`submitter_id + create_request_id` 唯一；状态、提交人、处理人、最近活动时间建索引。
- `user_feedback_messages`：消息类型、作者、状态变化与请求 ID；`feedback_id + author_type + author_id + request_id` 唯一。
- `user_feedback_attachments`：只存 `storage_provider`、`object_key` 和安全元数据；反馈和消息建联合索引。
- `user_feedback_daily_sequences`：日期主键和当日序号。
- 时间字段使用 UTC 毫秒；主表包含 `version`、`resolved_at`、`closed_at`，不增加物理删除入口。

- [ ] **Step 4: 创建权限迁移**

写入：

- Admin 顶层菜单 `admin:menu:user-feedback`，路径 `/user-feedbacks`；按钮权限 `user-feedback:list`、`user-feedback:handle`；对应 Admin API category 和 method/path 权限记录。
- H5App 菜单 `dingtalk_h5:menu:feedback`；反馈 API category 和 list/create/detail/supplement 权限。
- 使用现有角色、菜单、API 关联表结构，为现有普通 H5App 角色回填全部反馈权限，为现有 Admin 超级管理员/默认管理角色回填菜单与 API 权限。
- 使用幂等插入方式，重复执行迁移不产生重复授权。

- [ ] **Step 5: 运行测试并提交**

~~~bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./test/internal/bootstrap/migrations -run UserFeedback -count=1
git add migrations/20260905100000_create_user_feedback.sql migrations/20260905101000_add_user_feedback_permissions.sql test/internal/bootstrap/migrations/user_feedback_migration_test.go
git commit -m "feat: 添加用户反馈数据与权限迁移"
~~~

## Task 2: 建立领域状态机与持久化模型

**Files:**
- Create: `backend/internal/model/userfeedback/models.go`
- Create: `backend/internal/modules/userfeedback/domain/status.go`
- Create: `backend/internal/modules/userfeedback/domain/status_test.go`
- Create: `backend/internal/modules/userfeedback/application/types.go`
- Create: `backend/internal/modules/userfeedback/application/errors.go`

- [ ] **Step 1: 编写状态机失败测试**

覆盖全部合法/非法流转、说明必填规则和用户可补充状态：

~~~go
func TestValidateTransition(t *testing.T) {
    tests := []struct {
        from, to Status
        note string
        wantErr error
    }{
        {StatusPending, StatusProcessing, "", nil},
        {StatusPending, StatusClosed, "", ErrTransitionNoteRequired},
        {StatusProcessing, StatusResolved, "已修复", nil},
        {StatusResolved, StatusProcessing, "", ErrTransitionNoteRequired},
        {StatusClosed, StatusProcessing, "重新处理", nil},
        {StatusPending, StatusResolved, "跳级", ErrTransitionNotAllowed},
        {StatusResolved, StatusResolved, "重复", ErrTransitionNotAllowed},
    }
    for _, tt := range tests {
        err := ValidateTransition(tt.from, tt.to, tt.note)
        require.ErrorIs(t, err, tt.wantErr)
    }
}
~~~

- [ ] **Step 2: 运行测试并确认失败**

~~~bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./internal/modules/userfeedback/domain -count=1
~~~

Expected: FAIL，领域包尚不存在。

- [ ] **Step 3: 实现状态、消息类型和错误**

~~~go
type Status string

const (
    StatusPending    Status = "pending"
    StatusProcessing Status = "processing"
    StatusResolved   Status = "resolved"
    StatusClosed     Status = "closed"
)

type MessageType string

const (
    MessageInitial    MessageType = "initial"
    MessageSupplement MessageType = "supplement"
    MessageStatus     MessageType = "status"
)

func (status Status) AllowsSupplement() bool {
    return status == StatusPending || status == StatusProcessing
}
~~~

`ValidateTransition` 必须集中实现流转矩阵和说明必填规则，HTTP 层不得复制状态机。

- [ ] **Step 4: 定义持久化模型与应用 DTO**

模型字段与迁移列逐一对应；application DTO 固定跨端字段：

~~~go
type Overview struct {
    Pending int64 `json:"pending"`
    Processing int64 `json:"processing"`
    Resolved int64 `json:"resolved"`
    Closed int64 `json:"closed"`
}

type UpdateStatusCommand struct {
    FeedbackID uint64
    AdminID uint
    Status domain.Status
    Note string
    NotifyUser bool
    Version uint64
    RequestID string
}
~~~

定义稳定错误：参数无效、未找到/无权访问、状态不可补充、流转非法、说明必填、版本冲突、请求重复、每日限额、附件限制和存储失败。

- [ ] **Step 5: 运行测试并提交**

~~~bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./internal/modules/userfeedback/domain -count=1
git add internal/model/userfeedback internal/modules/userfeedback/domain internal/modules/userfeedback/application/types.go internal/modules/userfeedback/application/errors.go
git commit -m "feat: 建立用户反馈领域模型"
~~~

## Task 3: 补齐通用对象删除和图片验证

**Files:**
- Modify: `backend/internal/support/storage/storage.go`
- Modify: `backend/internal/support/storage/aliyun.go`
- Modify: `backend/internal/support/storage/storage_test.go`
- Create: `backend/internal/modules/userfeedback/application/image.go`
- Create: `backend/internal/modules/userfeedback/application/image_test.go`
- Create: `backend/internal/modules/userfeedback/infrastructure/object_storage.go`
- Create: `backend/internal/modules/userfeedback/infrastructure/object_storage_test.go`

- [ ] **Step 1: 编写存储删除失败测试**

覆盖本地删除、阿里云 DELETE 方法与签名、404 幂等成功、其他非 2xx 返回脱敏错误：

~~~go
func TestDeleteStoredFileRemovesLocalObject(t *testing.T) {
    stored := writeStoredFixture(t)
    require.NoError(t, DeleteStoredFile(context.Background(), stored))
    _, err := os.Stat(stored.LocalPath)
    require.ErrorIs(t, err, os.ErrNotExist)
}

func TestDeleteAliyunUsesDeleteMethod(t *testing.T) {
    // httptest server asserts request.Method == http.MethodDelete
}
~~~

- [ ] **Step 2: 运行存储测试并确认失败**

~~~bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./internal/support/storage -run Delete -count=1
~~~

Expected: FAIL，统一删除入口尚不存在。

- [ ] **Step 3: 实现统一删除能力**

新增 `DeleteStoredFile(ctx, stored) error`：

- 本地对象删除不存在文件时视为成功。
- 阿里云按 `ObjectKey` 发起签名 DELETE；204、200、404 视为成功。
- 不把 access key、签名、响应正文或用户内容写入错误。
- 保留 `RemoveLocal` 兼容旧调用方，但新反馈模块只依赖统一删除接口。

- [ ] **Step 4: 编写图片校验失败测试**

使用真实 JPG/PNG/WebP 文件头，覆盖：

- 扩展名、声明 MIME、`http.DetectContentType` 三者不一致。
- 单张超过 10 MB。
- 首次/单次补充超过 6 张。
- 累计超过 30 张。
- 首次提交空文字和补充文字图片同时为空。

- [ ] **Step 5: 实现图片校验与存储适配**

application 只定义 `ImageStorage` 接口和经过验证的图片描述；infrastructure 负责调用 `storage.SaveMultipartFile`，对象前缀固定为 `uploads/feedback`，并通过 `media.FullURLWithStaticDomainContext` 按对象键动态生成响应 URL。

- [ ] **Step 6: 运行测试并提交**

~~~bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./internal/support/storage ./internal/modules/userfeedback/application ./internal/modules/userfeedback/infrastructure -run "Delete|Image|Storage" -count=1
git add internal/support/storage internal/modules/userfeedback/application/image.go internal/modules/userfeedback/application/image_test.go internal/modules/userfeedback/infrastructure/object_storage.go internal/modules/userfeedback/infrastructure/object_storage_test.go
git commit -m "feat: 增加反馈图片校验与删除补偿"
~~~

## Task 4: 实现 application 用例与事务契约

**Files:**
- Create: `backend/internal/modules/userfeedback/application/service.go`
- Create: `backend/internal/modules/userfeedback/application/create.go`
- Create: `backend/internal/modules/userfeedback/application/supplement.go`
- Create: `backend/internal/modules/userfeedback/application/query.go`
- Create: `backend/internal/modules/userfeedback/application/status.go`
- Create: `backend/internal/modules/userfeedback/application/service_test.go`

- [ ] **Step 1: 用 fake Store 编写失败测试**

覆盖：

- 新建反馈生成 `FB-YYYYMMDD-NNNN`，文字必填，每用户自然日最多 20 条。
- 自然日边界由注入的 application clock 和 `Asia/Shanghai` location 计算，不依赖数据库会话时区，测试覆盖 00:00 边界。
- 同一 `submitterId + requestId` 重放返回首次结果。
- H5 列表、总览、详情只能访问当前用户。
- 仅 pending/processing 可补充，补充事务内复核状态、版本和累计附件数。
- 状态更新使用乐观锁并追加状态消息。
- 需要通知时状态、消息和 Outbox 由同一 TransactionStore 写入。
- DB 提交失败后只清理本次上传对象。

事务接口按现有 Workflow 模式定义：

~~~go
type Store interface {
    InTransaction(context.Context, func(TransactionStore) error) error
    GetUserOverview(context.Context, uint) (Overview, error)
    ListUserFeedbacks(context.Context, uint, UserListQuery) (FeedbackList, error)
    GetUserFeedback(context.Context, uint64, uint) (*FeedbackDetail, error)
    GetAdminOverview(context.Context, AdminListQuery) (Overview, error)
    ListAdminFeedbacks(context.Context, AdminListQuery) (FeedbackList, error)
    GetAdminFeedback(context.Context, uint64) (*FeedbackDetail, error)
}

type TransactionStore interface {
    NextFeedbackNumber(context.Context, string) (string, error)
    CountCreatedByUserOnDate(context.Context, uint, int64, int64) (int64, error)
    FindCreateReplay(context.Context, uint, string) (*FeedbackDetail, bool, error)
    CreateFeedback(context.Context, CreateRecord) error
    LockFeedback(context.Context, uint64) (*FeedbackSnapshot, error)
    FindMessageReplay(context.Context, MessageRequestKey) (*FeedbackDetail, bool, error)
    AppendMessage(context.Context, MessageRecord) (uint64, error)
    AppendAttachments(context.Context, []AttachmentRecord) error
    UpdateSnapshot(context.Context, SnapshotUpdate) error
    EnqueueNotification(context.Context, NotificationOutboxRecord) error
}
~~~

- [ ] **Step 2: 运行测试并确认失败**

~~~bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./internal/modules/userfeedback/application -count=1
~~~

Expected: FAIL，service 尚未实现。

- [ ] **Step 3: 实现新建与补充**

实现顺序严格为：

1. 规范化并验证 content/requestId/version/图片。
2. 幂等查询。
3. 保存图片对象并记录本次对象列表。
4. 开事务，锁定/复核配额、状态、版本、累计附件数，写主表、消息和附件。
5. 事务失败后调用 `ImageStorage.Delete` 逐个补偿；只记录请求 ID、反馈 ID/编号、对象键和错误链。
6. 成功后返回统一 detail/list DTO。

- [ ] **Step 4: 实现查询和状态更新**

- 用户详情查询同时约束 feedback ID 与 submitter ID；不存在和越权统一返回 `ErrNotFound`。
- Admin 关键词最大 100 字，分页最大 100，时间范围合法。
- 状态事务先按 requestId 查询重放，再锁主表，校验 version 和领域流转，追加 status 消息，更新 handler/version/timestamps。
- `resolved_at` 在首次到达 resolved 时写入；重新打开不删除历史；当前 closed snapshot 在关闭时写 `closed_at`。
- notifyUser 为 true 时生成幂等键 `user-feedback-status:<feedbackId>:<statusMessageId>` 并在同一事务写 Outbox。
- Outbox 固定使用 channel=internal、notificationType=feedback_status、sourceType=user_feedback、sourceId=反馈 ID、recipientUserId=submitter_id；标题为“反馈 <feedbackNo> <状态名称>”，正文为处理说明。

- [ ] **Step 5: 运行测试并提交**

~~~bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./internal/modules/userfeedback/application -count=1
git add internal/modules/userfeedback/application
git commit -m "feat: 实现用户反馈应用服务"
~~~

## Task 5: 实现 GORM Store、聚合查询和并发保护

**Files:**
- Create: `backend/internal/modules/userfeedback/infrastructure/gorm_store.go`
- Create: `backend/internal/modules/userfeedback/infrastructure/gorm_store_test.go`
- Create: `backend/internal/modules/userfeedback/infrastructure/query.go`
- Create: `backend/internal/modules/userfeedback/infrastructure/query_test.go`
- Create: `backend/internal/modules/userfeedback/infrastructure/mapper.go`

- [ ] **Step 1: 编写 GORM 失败测试**

使用现有测试数据库约定覆盖：

- `InTransaction` 把同一个 tx Store 传给回调，任一步骤失败整体回滚。
- 每日序列使用行锁，编号可超过四位且不循环。
- 锁定反馈使用 `FOR UPDATE`；version 更新条件包含旧 version。
- 幂等冲突读取首次结果，不重复附件和 Outbox。
- 用户总览和 Admin 总览各只执行一次条件聚合。
- 列表批量查询用户、管理员和附件数量，不逐行查询。

聚合 SQL 目标：

~~~sql
SELECT
  SUM(feedback_status = 'pending') AS pending,
  SUM(feedback_status = 'processing') AS processing,
  SUM(feedback_status = 'resolved') AS resolved,
  SUM(feedback_status = 'closed') AS closed
FROM user_feedbacks
WHERE submitter_id = ?
~~~

- [ ] **Step 2: 运行测试并确认失败**

~~~bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./internal/modules/userfeedback/infrastructure -count=1
~~~

Expected: FAIL，GORM Store 尚不存在。

- [ ] **Step 3: 实现事务 Store**

- `InTransaction` 使用 `database.QueryContext(ctx)` 和 `db.WithContext(queryCtx).Transaction`。
- 事务内直接写 notification outbox 模型，payload/recipient 使用 `notificationoutbox/application` 的稳定结构编码，避免另起数据库连接。
- OnConflict 只用于幂等键；其他数据库错误原样向 application 返回并由 HTTP 层统一脱敏。
- 详情一次加载主表、消息、附件，再批量加载涉及的用户/管理员显示名。
- 附件响应 URL 在 mapper 中通过 `media.FullURLWithStaticDomainContext(ctx, "/"+objectKey)` 生成。

- [ ] **Step 4: 运行测试并提交**

~~~bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./internal/modules/userfeedback/infrastructure -count=1
git add internal/modules/userfeedback/infrastructure
git commit -m "feat: 实现用户反馈持久化"
~~~

## Task 6: 扩展通用通知类型并保持旧消息兼容

**Files:**
- Modify: `backend/internal/modules/notificationoutbox/application/types.go`
- Modify: `backend/internal/modules/notificationoutbox/application/service_test.go`
- Modify: `backend/internal/modules/notificationoutbox/infrastructure/channels.go`
- Modify: `backend/internal/modules/notificationoutbox/infrastructure/channels_test.go`
- Modify: `backend/internal/support/notificationstyle/style.go`
- Modify: `backend/internal/support/notificationstyle/style_test.go`

- [ ] **Step 1: 编写失败测试**

~~~go
func TestInternalChannelForwardsOptionalNotificationType(t *testing.T) {
    row := notificationRow(t, InternalRecipient{UserIDs: []uint{7}}, MessagePayload{
        Title: "反馈 FB-20260905-0001 已解决",
        Content: "问题已修复",
        NotificationType: "feedback_status",
        SourceType: "user_feedback",
        SourceID: "91",
    })
    require.NoError(t, channel.Deliver(context.Background(), row))
    require.Equal(t, "feedback_status", delivery.input.NotificationType)
}
~~~

同时断言旧 payload 没有 notificationType 时仍沿用现有默认类型。

- [ ] **Step 2: 运行测试并确认失败**

~~~bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./internal/modules/notificationoutbox/... ./internal/support/notificationstyle -run "NotificationType|FeedbackStatus" -count=1
~~~

Expected: FAIL，`MessagePayload` 尚无该字段或渠道未转发。

- [ ] **Step 3: 实现兼容扩展**

~~~go
type MessagePayload struct {
    Title string `json:"title"`
    Content string `json:"content"`
    NotificationType string `json:"notificationType,omitempty"`
    SourceType string `json:"sourceType,omitempty"`
    SourceID string `json:"sourceId,omitempty"`
}
~~~

`InternalChannel.Deliver` 把非空 NotificationType 传给 in-app service；空值保持旧逻辑。新增 `notificationstyle.TypeFeedbackStatus`，默认标签“用户反馈”、chat 图标、primary/info 色调。

- [ ] **Step 4: 运行测试并提交**

~~~bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./internal/modules/notificationoutbox/... ./internal/support/notificationstyle -count=1
git add internal/modules/notificationoutbox internal/support/notificationstyle
git commit -m "feat: 支持用户反馈站内信类型"
~~~

## Task 7: 暴露 H5App/Admin HTTP 接口并同步权限目录

**Files:**
- Create: `backend/internal/modules/userfeedback/transport/httpdingtalkh5/handler.go`
- Create: `backend/internal/modules/userfeedback/transport/httpdingtalkh5/handler_test.go`
- Create: `backend/internal/modules/userfeedback/transport/httpadmin/handler.go`
- Create: `backend/internal/modules/userfeedback/transport/httpadmin/handler_test.go`
- Modify: `backend/internal/routes/v2/dingtalkh5/routes.go`
- Modify: `backend/internal/routes/v2/admin/routes.go`
- Create: `backend/internal/routes/v2/dingtalkh5/user_feedback_routes_test.go`
- Create: `backend/internal/routes/v2/admin/user_feedback_routes_test.go`
- Modify: `backend/internal/support/appmenuperm/catalog.go`
- Modify: `backend/internal/support/appapiperm/catalog.go`
- Modify: `backend/internal/support/appapiperm/catalog_test.go`
- Modify: `backend/internal/support/adminmenuperm/declarations.go`
- Modify: `backend/internal/support/adminrouteperm/catalog.go`
- Modify: `backend/internal/middleware/admin/route_permissions.go`

- [ ] **Step 1: 编写 handler 与路由失败测试**

测试必须覆盖：

- H5 handler 只从 `dingtalkh5session.CurrentUser(c)` 读取用户 ID，不接受 submitterId。
- Admin handler 只从 context 的 admin 读取处理人 ID。
- multipart 总请求体在解析前限流，读取 `content`、`requestId`、`version` 和重复 `images`。
- H5 不属于当前用户的详情返回与不存在相同的响应。
- Admin status body 缺 version/requestId、非法状态和并发冲突映射为稳定业务错误。
- 五条 H5 路由/四条 Admin 路由注册在带权限中间件的 group 下。

- [ ] **Step 2: 运行测试并确认失败**

~~~bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./internal/modules/userfeedback/transport/... ./internal/routes/v2/admin ./internal/routes/v2/dingtalkh5 -run UserFeedback -count=1
~~~

Expected: FAIL，handler 和路由尚不存在。

- [ ] **Step 3: 实现 H5 HTTP 契约**

注册：

~~~text
GET  /api/v2/dingtalk/h5/user-feedbacks/overview
GET  /api/v2/dingtalk/h5/user-feedbacks
POST /api/v2/dingtalk/h5/user-feedbacks
GET  /api/v2/dingtalk/h5/user-feedbacks/:id
POST /api/v2/dingtalk/h5/user-feedbacks/:id/supplements
~~~

create/supplement 设置总请求体上限，先校验文件 header 数量和 Size，再交给 application 做真实 MIME 校验与保存。响应沿用 `response.JSON`，内部错误使用 `response.FailInternal`，不暴露 SQL/OSS 错误。

- [ ] **Step 4: 实现 Admin HTTP 契约**

注册：

~~~text
GET   /api/v2/admin/user-feedbacks/overview
GET   /api/v2/admin/user-feedbacks
GET   /api/v2/admin/user-feedbacks/:id
PATCH /api/v2/admin/user-feedbacks/:id/status
~~~

status 请求 DTO 精确为 status、note、notifyUser、version、requestId；notifyUser 未传时默认 true，可通过指针布尔区分 false 与缺省值。

- [ ] **Step 5: 同步代码权限目录**

- Admin menu declaration 使用顶层 TypeMenu，路径 `/user-feedbacks`；list/handle 为其子按钮。
- Admin 权限目录新增 user-feedback category；route permission 将三个 GET 映射 `user-feedback:list`，PATCH 映射 `user-feedback:handle`。
- H5 menu 排序位于流程之后；API category 下声明 overview/list/create/detail/supplement 的 method + path。
- 校验代码目录和 SQL 迁移中的 key、method、path 完全一致。

- [ ] **Step 6: 运行测试并提交**

~~~bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./internal/modules/userfeedback/... ./internal/routes/v2/admin ./internal/routes/v2/dingtalkh5 ./internal/support/appapiperm ./internal/support/adminrouteperm ./internal/support/adminmenuperm -count=1
git add internal/modules/userfeedback/transport internal/routes/v2/admin/routes.go internal/routes/v2/admin/user_feedback_routes_test.go internal/routes/v2/dingtalkh5/routes.go internal/routes/v2/dingtalkh5/user_feedback_routes_test.go internal/support/appmenuperm internal/support/appapiperm internal/support/adminmenuperm internal/support/adminrouteperm internal/middleware/admin/route_permissions.go
git commit -m "feat: 开放用户反馈管理接口"
~~~

## Task 8: 同步 Swagger 与后端开发文档

**Files:**
- Modify: `backend/internal/routes/v2/swagger/h5app.go`
- Modify: `backend/internal/routes/v2/swagger/request_models.go`
- Modify: `backend/internal/routes/v2/swagger/swagger.go`
- Modify: `backend/docs/development-guidelines.md`
- Regenerate: `backend/docs/swagger/docs.go`
- Regenerate: `backend/docs/swagger/swagger.json`
- Regenerate: `backend/docs/swagger/swagger.yaml`

- [ ] **Step 1: 增加 Swagger 契约检查**

在现有 Swagger/route 测试中断言 Admin 和 H5 反馈路径、multipart 字段、status 请求 DTO 均出现在生成文档中。

- [ ] **Step 2: 运行检查并确认失败**

~~~bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./internal/routes/v2/swagger ./test/internal/bootstrap/migrations -run "Swagger|UserFeedback" -count=1
~~~

Expected: FAIL，反馈接口尚未进入 Swagger。

- [ ] **Step 3: 添加注释并按规范生成 Swagger**

描述归属校验、图片限制、状态流转、幂等 requestId 和 version 冲突。执行以下规范命令，禁止手改生成文件：

~~~bash
cd backend
swag init -g main.go --dir ./cmd,./internal/routes/v2/swagger --parseDependency --output docs/swagger
~~~

- [ ] **Step 4: 补充 Backend 模块说明并提交**

文档只描述已经实现的目录、事务/Outbox 边界、存储补偿和测试入口，不把规划项写成已完成能力。

~~~bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./internal/routes/v2/swagger ./test/internal/bootstrap/migrations -count=1
git add internal/routes/v2/swagger docs/development-guidelines.md docs/swagger
git commit -m "docs: 补充用户反馈接口文档"
~~~

## Task 9: 建立 H5App API、multipart helper 和反馈路由键

**Files:**
- Modify: `h5app/src/api/dingtalk-h5/base.ts`
- Create: `h5app/src/api/user-feedback.ts`
- Create: `h5app/src/pages/feedback/feedback-route-keys.ts`
- Create: `h5app/src/pages/feedback/feedback.routes.ts`
- Create: `h5app/src/pages/feedback/feedback.menu.ts`
- Create: `h5app/src/pages/feedback/feedback-status.ts`
- Modify: `h5app/src/config/app-navigation.ts`
- Modify: `h5app/src/config/app-content-routes.ts`
- Modify: `h5app/src/config/app-icons.ts`
- Create: `h5app/scripts/check-user-feedback.mjs`
- Modify: `h5app/package.json`

- [ ] **Step 1: 编写失败的静态契约检查**

脚本断言：

- 顶层 menu key 为 `feedback`，权限 key 为 `dingtalk_h5:menu:feedback`。
- 固定主页键 `feedback`、新建键 `feedback:create`、详情键 `feedback:detail:<id>`。
- 五个 API path 和 permission key 与 Backend 一致。
- 状态颜色：pending/processing warning，resolved success，closed info。
- 动态详情解析拒绝空 ID 和非法 URI 编码。

- [ ] **Step 2: 运行检查并确认失败**

~~~bash
cd h5app
pnpm check:user-feedback
~~~

Expected: FAIL，脚本或反馈模块尚不存在。

- [ ] **Step 3: 实现 H5 专用多文件 multipart helper**

在 `base.ts` 新增 H5 条件编译函数 `uploadMultipartFiles`：

- 使用 H5 `FormData` 追加 content/requestId/version 和重复 images。
- 将 uni 选择图片返回的临时 URL 转成 Blob，保留安全文件名和 MIME。
- 使用现有 BASE_URL、Authorization、X-Client-Platform 和 ApiEnvelope 解析。
- 30 秒超时并支持 AbortController。
- 只在 `#ifdef H5` 分支使用 fetch/FormData；`#ifndef H5` 明确 reject “当前上传仅支持 H5”，避免非 H5 构建引用浏览器类型。
- 上传失败不清空页面本地图片和文字。

- [ ] **Step 4: 实现反馈类型/API/路由键**

API 类型固定与后端 DTO 一致；route key helpers 示例：

~~~ts
export const FEEDBACK_CONTENT_KEY = 'feedback'
export const FEEDBACK_CREATE_CONTENT_KEY = 'feedback:create'
export const FEEDBACK_DETAIL_PREFIX = 'feedback:detail:'

export function feedbackDetailContentKey(id: string | number) {
  const value = String(id).trim()
  return value ? FEEDBACK_DETAIL_PREFIX + encodeURIComponent(value) : ''
}
~~~

- [ ] **Step 5: 注册菜单、内容路由和脚本**

向 `AppContentView` 增加 feedback，导入 feedback menu/routes；详情 resolver 放在 workflow resolver 旁，未知动态 key 仍回退 dashboard。`package.json` 增加 `check:user-feedback`，不破坏现有脚本链。

- [ ] **Step 6: 运行检查并提交**

~~~bash
cd h5app
pnpm check:user-feedback
pnpm type-check
git add src/api/dingtalk-h5/base.ts src/api/user-feedback.ts src/pages/feedback src/config/app-navigation.ts src/config/app-content-routes.ts src/config/app-icons.ts scripts/check-user-feedback.mjs package.json
git commit -m "feat: 建立 H5 用户反馈契约"
~~~

## Task 10: 实现 H5App 我的反馈主页

**Files:**
- Create: `h5app/src/pages/feedback/components/FeedbackCenter.vue`
- Create: `h5app/src/pages/feedback/components/FeedbackStatusOverview.vue`
- Create: `h5app/src/pages/feedback/components/FeedbackList.vue`
- Modify: `h5app/src/pages/feedback/feedback.routes.ts`
- Modify: `h5app/src/locale/lang/zh-CN.json`
- Modify: `h5app/src/locale/lang/en-US.json`
- Modify: `h5app/scripts/check-user-feedback.mjs`

- [ ] **Step 1: 扩展失败契约检查**

断言主页具备：

- 四个状态统计入口、keyword/status 筛选、分页加载、空态、失败重试。
- 新建按钮打开 `feedback:create` 动态页。
- 进入页面、浏览器恢复前台、成功事件和点击统计项时刷新 overview。
- 不设置高频 interval/polling。
- 用户可见文案来自 locale。

- [ ] **Step 2: 运行检查并确认失败**

~~~bash
cd h5app
pnpm check:user-feedback
~~~

Expected: FAIL，主页组件尚不存在。

- [ ] **Step 3: 实现主页交互**

- 使用紧凑工作台布局，统计项不做营销式大卡片。
- 列表按 lastActivityAt 倒序显示 feedbackNo、summary、状态、imageCount、时间。
- 点击列表打开/聚焦同一个 `feedback:detail:<id>` 动态页签。
- watch 当前内容 key 和 refreshTick；H5 分支注册 visibilitychange，卸载时移除监听。
- 成功事件通过 appContent refresh 机制触发，而不是全局轮询。

- [ ] **Step 4: 运行检查并提交**

~~~bash
cd h5app
pnpm check:user-feedback
pnpm lint
pnpm type-check
git add src/pages/feedback src/locale/lang/zh-CN.json src/locale/lang/en-US.json scripts/check-user-feedback.mjs
git commit -m "feat: 添加 H5 我的反馈主页"
~~~

## Task 11: 实现 H5App 新建、详情、补充和未保存保护

**Files:**
- Create: `h5app/src/pages/feedback/components/FeedbackImagePicker.vue`
- Create: `h5app/src/pages/feedback/components/FeedbackCreatePage.vue`
- Create: `h5app/src/pages/feedback/components/FeedbackDetailPage.vue`
- Create: `h5app/src/pages/feedback/components/FeedbackTimeline.vue`
- Modify: `h5app/src/pages/feedback/feedback.routes.ts`
- Modify: `h5app/src/pages/index/index.vue`
- Modify: `h5app/scripts/check-user-feedback.mjs`
- Modify: `h5app/src/locale/lang/zh-CN.json`
- Modify: `h5app/src/locale/lang/en-US.json`

- [ ] **Step 1: 扩展失败契约检查**

覆盖：

- 新建首次进入时生成 requestId，网络重试复用同一值；成功后才生成下一值。
- 文字 5000 字和最多 6 图前端提示，后端仍为最终边界。
- 图片可选择、预览、移除并显示单图错误。
- pending/processing 显示补充区；resolved/closed 只读。
- 补充使用 detail 返回的 version，并为每次编辑生成稳定 requestId。
- 新建或补充存在未保存内容时注册 appContent close guard；成功/离开时注销。
- 成功后关闭 create 页、刷新 center 并打开 detail。
- `index.vue` 能从 URL 的 view 参数恢复 feedback 动态页。

- [ ] **Step 2: 运行检查并确认失败**

~~~bash
cd h5app
pnpm check:user-feedback
~~~

Expected: FAIL，动态页面行为尚不完整。

- [ ] **Step 3: 实现图片选择与新建页**

- 复用 WorkflowImagePicker 的交互原则，不直接依赖 Workflow 业务组件。
- 新建页使用多行文字、图片网格和明确提交按钮；提交期间禁用重复点击。
- 失败保留 content、requestId 和仍有效的本地图片。
- 成功后触发 refresh，关闭当前动态页并打开 detail。

- [ ] **Step 4: 实现详情与补充**

- 时间线按 createdAt 正序展示 initial、supplement、status。
- status 记录展示 fromStatus/toStatus、管理员显示名和处理说明。
- 图片点击调用 uni.previewImage。
- 补充允许只有文字或只有图片；提交成功清空编辑区并刷新 detail/version。
- 详情 API 404 显示不可访问空态，不从通知 payload 信任反馈内容。

- [ ] **Step 5: 合并 index.vue 现有修改并提交**

`h5app/src/pages/index/index.vue` 当前可能含用户未提交改动；先读取最新文件，只追加 feedback key normalization/open 逻辑，不覆盖 Workflow 变化。

~~~bash
cd h5app
pnpm check:user-feedback
pnpm lint
pnpm type-check
git add src/pages/feedback src/pages/index/index.vue src/locale/lang/zh-CN.json src/locale/lang/en-US.json scripts/check-user-feedback.mjs
git commit -m "feat: 完成 H5 反馈提交与详情"
~~~

## Task 12: 接通 H5App 站内信详情深链

**Files:**
- Modify: `h5app/src/api/notifications.ts`
- Modify: `h5app/src/components/app-notification-panel/app-notification-panel.vue`
- Modify: `h5app/scripts/check-user-feedback.mjs`

- [ ] **Step 1: 编写失败契约检查**

断言 `feedback_status` 显示“用户反馈”样式；`sourceType === 'user_feedback'` 且 sourceId 有效时：

1. 先调用标记已读。
2. 打开或聚焦 `feedback:detail:<sourceId>`。
3. 关闭通知面板。
4. sourceId 为空时只展示消息，不跳转。

- [ ] **Step 2: 运行检查并确认失败**

~~~bash
cd h5app
pnpm check:user-feedback
~~~

Expected: FAIL，通知面板尚不识别反馈来源。

- [ ] **Step 3: 实现通知映射和跳转**

导入 `feedbackDetailContentKey`，为反馈通知使用 chat 图标和统一色调；保留未知通知类型现有 fallback，不改变 Workflow 通知跳转。

- [ ] **Step 4: 运行检查并提交**

~~~bash
cd h5app
pnpm check:user-feedback
pnpm check:dingtalk-module
pnpm lint
pnpm type-check
git add src/api/notifications.ts src/components/app-notification-panel/app-notification-panel.vue scripts/check-user-feedback.mjs
git commit -m "feat: 接通反馈站内信详情跳转"
~~~

## Task 13: 建立 Admin API、路由、状态映射和权限保护

**Files:**
- Create: `admin/src/types/userFeedback.ts`
- Modify: `admin/src/api/index.ts`
- Modify: `admin/src/router/adminRoutes.ts`
- Create: `admin/src/views/user-feedback/userFeedbackStatus.ts`
- Create: `admin/scripts/check-user-feedback.mjs`
- Modify: `admin/package.json`
- Modify: `admin/scripts/check-navigation.mjs`
- Modify: `admin/scripts/check-icon-runtime.mjs`

- [ ] **Step 1: 编写失败的 Admin 契约检查**

断言：

- 路由 `/user-feedbacks` 使用 lazy component、`menuPath: '/user-feedbacks'` 和 `adminUi: { version: 1, pattern: 'filter-list' }`。
- API 包含 overview/list/detail/updateStatus，路径与 Backend 一致。
- 状态颜色映射与 H5 一致。
- 查看使用 `admin:menu:user-feedback:list`，处理使用 `admin:menu:user-feedback:handle`；按钮可见性和事件 handler 都检查 handle 权限。
- icon runtime 和 navigation 能解析新菜单图标/path。

- [ ] **Step 2: 运行检查并确认失败**

~~~bash
cd admin
npm run check:user-feedback
~~~

Expected: FAIL，脚本或 Admin 反馈模块尚不存在。

- [ ] **Step 3: 实现类型与 API**

~~~ts
export type UserFeedbackStatus = 'pending' | 'processing' | 'resolved' | 'closed'

export interface UpdateUserFeedbackStatusInput {
  status: UserFeedbackStatus
  note: string
  notifyUser: boolean
  version: number
  requestId: string
}
~~~

API 方法沿用 `admin/src/api/index.ts` 的 request 封装，不引入第二套 HTTP 客户端。

- [ ] **Step 4: 注册规范路由和权限辅助**

新路由必须通过 `check-admin-ui-contract`。读取并合并当前已修改的 navigation/icon 脚本，只增加 feedback 断言，不覆盖现有 Workflow/Admin 修复。

- [ ] **Step 5: 运行检查并提交**

~~~bash
cd admin
npm run check:user-feedback
npm run check:navigation
npm run check:icon-runtime
git add src/types/userFeedback.ts src/api/index.ts src/router/adminRoutes.ts src/views/user-feedback/userFeedbackStatus.ts scripts/check-user-feedback.mjs scripts/check-navigation.mjs scripts/check-icon-runtime.mjs package.json
git commit -m "feat: 建立后台用户反馈契约"
~~~

将 `check:user-feedback` 插入 `check:all`，使根级和 CI 的 Admin 检查默认覆盖反馈页面契约。

## Task 14: 实现 Admin 总览、列表、详情抽屉和状态对话框

**Files:**
- Create: `admin/src/views/user-feedback/index.vue`
- Create: `admin/src/views/user-feedback/components/UserFeedbackOverview.vue`
- Create: `admin/src/views/user-feedback/components/UserFeedbackDetailDrawer.vue`
- Create: `admin/src/views/user-feedback/components/UserFeedbackTimeline.vue`
- Create: `admin/src/views/user-feedback/components/UserFeedbackStatusDialog.vue`
- Modify: `admin/scripts/check-user-feedback.mjs`

- [ ] **Step 1: 扩展失败契约检查**

断言页面：

- 使用 AdminPageShell、AdminPageHeader、AdminSearchBar、AdminTablePanel。
- overview 四项可点击并刷新统计与列表。
- 筛选 feedbackNo/用户、status、handlerId、submittedFrom/submittedTo、keyword。
- 表格包含编号、摘要/首图、提交人、状态、处理人、图片数量、最近活动、提交时间和查看。
- 详情用右侧 AdminDrawer，时间线完整，底部更新状态按钮固定。
- 状态编辑使用独立 AdminDialog/ElDialog，`append-to-body`，不嵌套在 drawer DOM。
- 无 handle 权限时按钮不渲染且 submit handler 立即拒绝。
- 冲突/失败保留 status、note、notifyUser；成功才关闭并刷新 overview/list/detail。

- [ ] **Step 2: 运行检查并确认失败**

~~~bash
cd admin
npm run check:user-feedback
npm run check:admin-ui-contract
~~~

Expected: FAIL，页面组件尚不存在。

- [ ] **Step 3: 实现总览和列表**

- 使用紧凑统计条，避免卡片嵌套。
- overview 请求与列表请求分离；页面进入、回到浏览器前台、状态操作成功、点击统计项时刷新。
- 列表加载不拉取详情；首图使用后端已生成的绝对 URL。
- 统一 loading、empty、error 和分页状态。

- [ ] **Step 4: 实现详情与状态对话框**

- 抽屉读取 detail，展示初始反馈、补充和状态记录；图片可预览。
- 可选目标状态只显示领域状态机允许的下一状态。
- 必填说明随目标状态实时校验；notifyUser 默认 true。
- 打开对话框时生成 requestId，重试复用；成功后才清理。
- version 冲突提示刷新详情，不自动覆盖管理员输入。
- 对话框 append-to-body 并使用项目统一 z-index，不被固定导航/抽屉遮挡。

- [ ] **Step 5: 运行检查并提交**

~~~bash
cd admin
npm run check:user-feedback
npm run check:admin-ui-contract
npm run check:component-complexity
npm run build
git add src/views/user-feedback scripts/check-user-feedback.mjs
git commit -m "feat: 添加后台用户反馈总览"
~~~

## Task 15: 跨模块回归、全量构建和浏览器验证

**Files:**
- Modify only when a check exposes a feedback-related defect; do not reformat unrelated files.
- Modify: `docs/superpowers/specs/2026-09-05-user-feedback-design.md` only if implementation deliberately changes an approved contract.
- Modify: relevant module development guideline only when the delivered behavior requires a durable rule.

- [ ] **Step 1: 运行 Backend 全量测试**

~~~bash
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./... -count=1
~~~

Expected: PASS。若 loopback/httptest 仅因沙箱被拒绝，单独在获准环境重跑失败包并记录两次结果。

- [ ] **Step 2: 运行 Admin 全量检查**

~~~bash
cd admin
npm run check:all
~~~

Expected: PASS，包括 build 和 bundle budget。

- [ ] **Step 3: 运行 H5App 全量相关检查**

~~~bash
cd h5app
pnpm check:user-feedback
pnpm check:dingtalk-module
pnpm check:workflow-module
pnpm check:ui-style
pnpm lint
pnpm type-check
pnpm build:h5
~~~

Expected: PASS；Workflow 检查用于确认新增 feedback resolver 没破坏现有动态流程页。

- [ ] **Step 4: 运行根级复核**

~~~bash
cd ..
bash scripts/verify-local.sh
~~~

Expected: PASS。该脚本不覆盖 H5App，因此不能替代上一步。

- [ ] **Step 5: 启动本地服务并进行浏览器验证**

按现有启动脚本启动 backend、admin、h5app 和 taskd，使用 Browser 插件分别验证：

- PC Admin：1440x900 与窄桌面，菜单、统计筛选、详情抽屉、状态对话框层级、权限隐藏。
- H5App：390x844 与 PC 宽度，主页、新建、图片预览、详情、补充、只读状态、未保存关闭保护。
- 在 Admin 更新状态并保持 notifyUser 开启，等待 taskd 投递，点击 H5 站内信直接进入对应反馈详情。
- 检查浏览器控制台无错误、请求没有重复轮询、文字和按钮不溢出。
- 真钉钉免登、真机图片选择和部署环境通知点击由用户做最终手工验收，并在交付说明中列为未自动验证项。

- [ ] **Step 6: 检查迁移与工作区边界**

~~~bash
git status --short
git diff --check
git log --oneline --max-count=15
~~~

确认：

- 没有手工编辑生成物之外的构建产物进入 Git。
- 没有覆盖任务开始前的 Workflow/Admin/H5App 用户修改。
- 新表/路由/权限/DTO/前端 key/文档一致。
- 现有绩效、问卷、流程和定时任务行为未被改写。

- [ ] **Step 7: 提交最后的反馈相关修复**

只在前述检查确实产生反馈相关修复时执行；先用 `git diff --name-only` 列出文件，再逐个执行 `git add 文件路径`，并用 `git diff --cached --name-status` 确认暂存区只含用户反馈相关修复后提交：

~~~bash
git diff --cached --name-status
git commit -m "test: 完成用户反馈跨端回归"
~~~

## 完成交付清单

- [ ] H5App 用户只能查看和补充自己的反馈。
- [ ] 首次文字必填且不超过 5000 字；每次最多 6 图、累计最多 30 图、单图最大 10 MB，只允许 JPG/PNG/WebP。
- [ ] 新建、补充和状态更新幂等；状态与补充并发使用 version 和行锁。
- [ ] 四状态流转、说明必填和颜色映射在 Backend/Admin/H5App 一致。
- [ ] 状态消息与 notification_outbox 同一事务；taskd 异步投递，不占用 HTTP 投递时间。
- [ ] 本地与阿里云对象都支持数据库失败后的尽力删除补偿。
- [ ] H5App 顶层菜单、动态页签、未保存保护和站内信详情跳转完整。
- [ ] Admin 总览、筛选、详情抽屉、状态对话框和权限双重保护完整。
- [ ] Swagger、迁移、权限目录、路由、DTO 和开发文档一致。
- [ ] Backend、Admin、H5App 自动化检查及 PC/移动浏览器验证结果已记录。
