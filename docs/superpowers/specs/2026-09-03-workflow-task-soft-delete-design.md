# 流程任务管理软删除设计

## 目标

Admin 流程任务列表支持删除操作，同时保留流程运行数据和审批审计记录。

## 删除语义

- 仅允许删除终态任务：`completed`、`approved`、`rejected`、`cancelled`。
- `waiting`、`pending` 任务仍参与流程执行，删除请求必须被后端拒绝。
- 删除采用管理侧软删除。删除后任务不再出现在 Admin 流程任务列表中，但仍保留在流程实例详情、流程图节点进度和审批历史中。
- 重复删除或任务不存在时返回明确的“任务不存在或已删除”错误。

## 数据模型

在 `workflow_process_tasks` 增加：

- `admin_deleted_at bigint NOT NULL DEFAULT 0`：管理员删除时间。
- `admin_deleted_by varchar(64) NOT NULL DEFAULT ''`：操作管理员 ID。
- 基于 `admin_deleted_at` 和任务创建时间的列表索引。

字段通过版本化迁移 SQL 创建，不使用运行时自动补齐。

## 后端接口

- 新增 `DELETE /api/v2/admin/workflow-tasks/:id`。
- 路由权限为 `workflow:task:delete`。
- 应用服务在事务内锁定任务、校验状态并写入软删除审计字段。
- Admin 任务列表查询统一过滤 `admin_deleted_at <> 0` 的记录。
- 流程状态加载、实例详情和历史查询不应用该过滤条件。

## 权限

迁移 SQL 新增：

- 按钮权限 `admin:menu:workflow:task:delete`。
- API 权限 `admin:api:workflow:task:delete`。

破坏性权限不自动授予现有角色或用户，由管理员在权限树中显式授权。

## Admin 交互

- 具备删除权限且任务处于终态时，在操作列显示红色“删除”按钮。
- 待激活、待处理任务不显示删除按钮；现有“处理”按钮逻辑保持不变。
- 点击删除后显示二次确认，成功后刷新当前列表，并在当前页无数据时回退上一页。

## 错误处理

- 缺少任务 ID、管理员身份、任务不存在或已删除、任务仍在运行分别返回清晰业务错误。
- 前端复用统一请求层展示后端错误，页面只补充删除成功反馈。

## 验证

- 应用服务测试覆盖终态删除、活动任务拒绝、任务不存在和审计字段。
- GORM 存储测试覆盖 Admin 列表过滤，实例详情仍保留被软删除任务。
- HTTP 处理器测试覆盖管理员身份和路由任务 ID。
- 权限测试覆盖路由映射、权限目录和迁移 SQL。
- Admin 结构检查覆盖权限判断、删除接口、确认框和按钮显示条件。
- 运行 Admin TypeScript 构建，并在浏览器验证终态任务删除流程。
