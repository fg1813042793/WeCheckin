# Admin 站内信发送与定时投递设计

## 1. 目标

在管理后台提供可操作的站内信发送能力，并为通用定时任务平台注册“发送站内信” Go 任务。手动发送和定时发送都支持：

- 全部用户。
- 指定部门。
- 指定用户。

定时任务保存的是接收范围规则，执行时根据当前组织数据解析启用用户。因此部门成员变化会在下一次执行自动生效，已停用用户不再收到新消息。

本次复用现有 `notify` 表、H5App 站内信列表和未读数接口；不将 `news` 公告表改造成个人收件箱，不改动现有流程通知的数据契约。

## 2. 范围与界限

### 2.1 包含

- Admin 顶层“站内信管理”菜单和发送按钮。
- Admin 站内信列表、未读状态和手动发送对话框。
- 通用站内信 application service、收件人解析和 GORM 投递存储。
- `/api/v2/admin` 手动发送接口、Swagger 和权限契约。
- 内置 Go 任务 `notification.in_app.send`。
- 定时任务编辑器中针对该任务的结构化表单。
- 通过版本化 SQL 新增菜单、按钮和 API 权限，并回填已有管理角色。

### 2.2 不包含

- 不把 Admin “内容管理”中的 `news` 自动发送到站内信。
- 不新增邮件、短信、钉钉 OA 或群机器人渠道。
- 不新增消息模板、变量表达式或附件。
- 不修改 H5App 收件箱的已读和跳转交互。

## 3. 总体架构

新增 `internal/modules/inappnotification` 独立模块，避免 Admin handler、定时任务和 H5App 各自复制投递逻辑：

- `application`：请求校验、接收范围解析、去重、事务投递和幂等编排。
- `infrastructure`：基于 GORM 查询启用用户及用户部门，批量写入 `notify`。
- `transport/httpadmin`：只解析管理端 DTO、读取管理员上下文、调用 application service 并输出统一响应。

定时任务 `GoJob` 依赖 application service 的稳定接口，不在 `builtin_jobs.go` 中直接写 GORM 查询。Admin HTTP 服务与 `taskd` 分别注册相同的任务实现，保证后台保存时的配置校验与 worker 执行一致。

## 4. 接收范围契约

手动发送和定时任务共用以下结构：

```json
{
  "title": "系统维护通知",
  "content": "今日 20:00 开始系统维护。",
  "scope": "departments",
  "userIds": [],
  "departmentIds": [10, 11]
}
```

- `scope=all`：忽略 ID 列表，选择全部 `user_status=1` 的用户。
- `scope=departments`：`departmentIds` 至少一项，根据 `user_depts` 选择当前启用用户。
- `scope=users`：`userIds` 至少一项，仅选择当前仍存在且已启用的用户。

收件人统一按本地用户数字 ID 去重。指定部门包含选中部门及其当前所有子部门，子部门在每次执行时根据当前组织树解析，使新增或迁移的部门与用户能按规则自动生效。

`title` 去除首尾空白后必填，最长 255 字符；`content` 必填，最长 5000 字符。解析后没有可投递用户时返回业务错误，不生成空批次。

## 5. 手动发送 API

新增：

```text
POST /api/v2/admin/in-app-notifications
Permission: notification:send
```

请求体在通用发送结构基础上增加 `requestId`。`requestId` 由 Admin 在打开发送对话框时生成，同一次提交及网络重试复用相同值。

响应：

```json
{
  "requestId": "uuid",
  "sentCount": 25,
  "skippedCount": 2,
  "replayed": false
}
```

投递记录使用：

- `notify_type=admin_manual`
- `notify_source_type=admin_manual`
- `notify_source_id=requestId`
- `notify_user_id=本地用户 ID 字符串`
- `notify_is_read=0`

服务先在事务内检查同一 `source_type + source_id` 是否已有记录。已存在时直接返回已投递数量并标记 `replayed=true`；否则使用 `CreateInBatches` 在同一事务中写入所有收件人。Admin 提交期间禁用按钮，作为前端并发防护；后端 `requestId` 作为重试幂等边界。

## 6. Admin 交互与权限

- 新增顶层“站内信管理”菜单，路由为 `/notifications`，菜单权限为 `notification:list`。
- 旧 `/survey/notify` 路由保留兼容，共用新页面，不再将站内信功能继续放入问卷业务实现。
- 页面右上角增加“发送站内信”按钮，按钮权限为 `notification:send`。
- 发送对话框使用分段选择接收范围；指定部门或用户时复用 `WorkflowUserTreePicker`。
- 用户和部门选项复用现有工作流用户树 API，后端发送时仍重新校验用户状态和部门归属。
- 成功后关闭对话框、显示实际发送人数并刷新列表。校验错误保留用户已填内容。

权限编码统一为：

- `notification:list`：查看站内信列表和未读数。
- `notification:read`：更新当前管理员自己的已读状态。
- `notification:send`：手动向普通用户发送站内信。

菜单、按钮和 API 权限声明全部通过 `backend/migrations/` 中的 SQL 创建和回填，不依赖主服务启动时自动补齐。代码目录仍同步声明契约，供路由鉴权和一致性测试使用。

## 7. 定时任务

注册内置 Go 任务：

```text
Key:  notification.in_app.send
Name: 发送站内信
```

`params` 使用第 4 节的通用结构。任务编辑器检测到该 key 后，显示标题、正文、接收范围和用户树，其他 Go 任务仍保留原始 JSON 参数编辑能力。

管理员创建或修改 `notification.in_app.send` 任务时，除定时任务管理权限外还必须具有 `notification:send`。Admin 前端不向无权用户展示该 Go 任务，后端创建/修改接口再做一次不可绕过的 handler 配置授权校验。任务成功创建后由系统身份按计划执行，不依赖创建人当时在线。

执行时：

1. 解析并校验参数。
2. 按执行当时的组织数据解析启用收件人。
3. 使用 `notify_type=scheduled_task`、`notify_source_type=scheduled_task_run`、`notify_source_id=runID` 投递。
4. 如同一 `runID` 已存在投递记录，返回重放结果，不重复发送。
5. 运行日志记录 `runID`、规划收件人数、实际发送数、跳过数和重放状态，不记录正文或完整收件人列表。

数据库或连接类故障返回 `Temporary=true` 的 `HandlerError`，由定时任务平台按任务策略重试；参数错误和无有效收件人属于永久错误，不做无意义重试。

## 8. 事务与幂等

- 一次发送的幂等键是 `source_type + source_id`，每个收件人的唯一投递键由 `source_type + source_id + user_id` 计算。
- `notify` 表新增可空 `notify_delivery_key` 字段和唯一索引。历史通知及旧写入路径保持 `NULL`，只有新增的手动/定时站内信写入该键，因此不会限制现有流程通知的合法重复。
- 该键在事务内先检查后批量写入，保证一次投递不会部分成功；并发请求遇到唯一键冲突时回滚后重读已投递结果。
- 定时平台的同一 `runID` 只有一个 worker 可成功抢占运行状态，事务提交后 worker 失联的重试由已存在记录识别。
- Admin 手动发送使用客户端 `requestId` 吸收超时重试，且前端在请求进行中禁止再次提交。

## 9. 错误处理

- 缺少标题、正文、范围或必需 ID 列表：返回 400 和稳定业务错误。
- 无 `notification:send` 权限：返回 403，后端不依赖按钮隐藏作为安全边界。
- 所选用户不存在或已停用：计入 `skippedCount`；全部失效时整体失败。
- 数据库写入失败：回滚整批消息，不返回虚假成功数。
- 成功日志只记录批次键和数量；失败日志保留内部错误链，不记录消息正文。

## 10. 迁移与兼容

新增一个版本化 SQL 迁移，用于创建：

- `notify.notify_delivery_key` 可空字段和唯一索引。
- 顶层菜单 `admin:menu:notification`。
- 按钮权限 `admin:menu:notification:list`、`admin:menu:notification:read`、`admin:menu:notification:send`。
- API 分组 `admin:api-category:notification`。
- API 权限 `admin:api:notification:list`、`admin:api:notification:read`、`admin:api:notification:send`。
- 为现有拥有原 `survey:list` 站内通知查看能力的管理角色回填 `notification:list` 和 `notification:read`。
- `notification:send` 仅默认授予超级管理员或已有通知发布权限的管理角色，其他角色需在权限树显式授权。

旧 Admin 列表接口 `/survey-notifications` 在过渡期保留；新页面改用 `/in-app-notifications`。原有问卷通知、流程通知和 H5App 查询继续使用现有数据，不需要转换历史 `notify` 记录。

## 11. 测试与验证

### 11.1 Backend

- application 单元测试：标题/正文/范围校验，全部用户、部门和用户解析，停用用户过滤，去重，空收件人，投递成功，事务回滚，幂等重放。
- Admin handler 测试：正常发送、非法 JSON、业务校验失败和响应 DTO。
- 路由与权限测试：`method + path + permission` 完整对应，迁移 SQL 包含菜单、按钮、API 和回填语句。
- 定时任务测试：参数 schema 与校验、Admin/taskd 双端注册、执行数量、无效参数、临时错误和同 `runID` 重试。
- 更新 Swagger 声明并通过规定命令重新生成文档。

### 11.2 Admin

- API 契约和 TypeScript 类型检查。
- 页面权限检查：无发送权限不显示按钮，事件处理使用同一权限条件。
- 结构化定时任务表单测试：任务切换、参数回填、范围切换、用户树与校验。
- 运行 `npm run check:all` 和 `npm run build`。
- 在桌面宽度下验证站内信列表、发送对话框、发送中状态、成功反馈及定时任务编辑。

## 12. 验收标准

1. 有权限的管理员可从“站内信管理”向全部、指定部门或指定用户发送消息。
2. H5App 对应用户能收到消息、看到未读数并标记已读。
3. 定时任务编辑器可视化配置站内信内容和范围，`taskd` 能按 Cron 执行。
4. 同一手动 `requestId` 或定时 `runID` 重试不会重复投递。
5. 权限由 SQL 迁移创建且前后端一致，不依赖运行时自动补齐。
6. 原有问卷、流程、H5App 站内信和其他定时任务行为保持不变。
