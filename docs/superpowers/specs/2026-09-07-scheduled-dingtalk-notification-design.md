# 定时钉钉通知注册任务设计

## 目标

在管理后台的“Go 注册任务”中增加独立的“发送钉钉通知”任务，使管理员可以按 Cron 周期向全部用户、指定部门或指定用户发送企业内部钉钉工作通知。

## 现状

- `system.workflow-business-event-dispatch` 只派发工作流业务事件 Outbox，不直接发送通知。
- `system.notification-outbox-dispatch` 只派发通用通知 Outbox，目前装配站内信和 Webhook 渠道；其中 Webhook 不是企业内部钉钉工作通知。
- `notification.in_app.send` 已作为可配置 Go 注册任务提供“发送站内信”，但没有对应的钉钉注册任务。
- 管理后台已经具有手工发送钉钉通知的收件人解析、企业配置选择、消息模板和投递实现。

## 方案

新增注册键 `notification.dingtalk.send`，中文名称为“发送钉钉通知”。任务参数沿用站内信任务的结构：

- `title`：通知标题，必填，最多 255 个字符。
- `content`：通知正文，必填，最多 5000 个字符。
- `scope`：`all`、`departments` 或 `users`。
- `departmentIds`：按部门发送时必填。
- `userIds`：按用户发送时必填。

任务执行时调用现有 `inappnotification` 应用服务的 `SendDingTalk` 方法，使用 `scheduled_task` 通知类型加载钉钉消息样式，并复用既有钉钉用户绑定、企业应用配置和分企业投递逻辑。每次调度运行使用运行 ID 作为来源 ID，运行日志只记录计划、成功、跳过和失败数量，不记录通知正文。

## 权限与界面

- 创建或修改该任务必须具有 `notification:dingtalk:send` API 权限。
- 管理后台只有具备 `admin:menu:notification:dingtalk-send` 时才显示该注册任务。
- 任务编辑器复用现有通知表单，但根据任务类型分别校验站内信或钉钉发送权限。
- 收件人选项继续使用现有统一用户和部门数据源。

## 装配边界

- Admin API 注册表加入新任务，以便处理器元数据接口返回该选项并校验任务配置。
- `taskd` 注册表加入同一任务，并把通知服务改为带钉钉投递器的实例，保证后台可配置的任务在执行进程中也确实存在。
- 两个现有系统派发任务不改名、不改职责、不迁移历史数据。

## 错误处理

- 参数错误返回 `invalid_config`。
- 没有有效收件人返回 `no_recipients`。
- 钉钉投递基础设施不可用或解析失败返回可重试的 `notification_delivery_failed`。
- 某些企业或用户投递失败时沿用现有批量投递语义，在任务结果中记录 `failedCount`，不因单个企业失败中断其他企业发送。

## 验证

- 新任务元数据、参数校验、调用映射、结果和日志脱敏单元测试。
- 风险权限映射测试。
- `taskd` 与 Admin 注册表装配契约测试。
- Admin 静态契约检查及 TypeScript 检查。
- Backend 受影响包测试；条件允许时运行 Backend 全量测试。
