# 用户反馈处理与通知

管理后台“用户反馈”列表和详情均提供“处理”和“通知提交人”。拥有 `admin:menu:user-feedback:handle` 的管理员可以使用这些入口；后端两个操作均校验已有 `user-feedback:handle` API 权限。仅有查看权限的账号不能处理或发送通知。

## 处理反馈

打开处理弹窗前加载最新详情。沿用现有状态流转、处理说明、版本校验和可选站内信开关，通过 `PATCH /api/v2/admin/user-feedbacks/{id}/status` 提交。

## 独立通知

通知弹窗显示反馈编号、提交人和当前状态，管理员自行编辑正文，并选择“站内通知”或“钉钉通知”，默认站内通知。接收人固定为该条反馈的提交人，不提供任意用户选择。

接口：`POST /api/v2/admin/user-feedbacks/{id}/notify`。

```json
{
  "channel": "internal",
  "content": "问题已修复，请确认是否恢复正常。",
  "version": 3,
  "requestId": "8f9f4f33-f9b8-4f2f-9b26-f83c33d4f17f"
}
```

- `channel`：`internal` 或 `dingtalk`；省略时使用站内通知。
- `content`：去除首尾空白后为 1～5000 字。
- `version`：最新反馈版本，冲突时刷新后重试。
- `requestId`：1～64 字的幂等标识，同一次提交的网络重试必须沿用，新的通知使用新的标识。
- 响应：最新 `FeedbackDetail`，接口成功表示已提交发送，不代表外部渠道已经送达。

独立通知不改变反馈状态或处理人。后端在同一事务中追加 `messageType=notification` 的管理员消息，更新活动时间和版本，并写入 `notification_outbox`；管理员和提交人均可在反馈时间线查看正文。请求不能指定收件人或管理员身份。

钉钉渠道在受理前检查提交人的有效绑定和启用通知的企业配置。实际发送复用后台钉钉通知的企业选择及消息样式规则；没有可用绑定或配置时，保留填写内容并提示改用站内通知。

## 运行与兼容

- 复用已有反馈表、消息表、Outbox 和处理权限，不新增表或权限 key。
- 部署需要同步更新 Backend、Admin、H5App 及 `taskd`。`taskd` 的通知队列任务负责投递，失败按现有策略重试；钉钉投递结果在通知记录中查询。
- 队列幂等避免同一请求重复入队。外部钉钉投递仍遵循现有至少一次语义：外部发送成功而本地结果持久化失败等极端情况可能导致重试，不能承诺外部消息严格只送达一次。
- 原状态更新接口及其通知开关保持兼容。

## 本次验证（2026-11-02）

- Backend：`GOCACHE=$PWD/../.cache/go-build go test ./internal/routes/v2/swagger ./internal/modules/userfeedback/... ./internal/modules/notificationoutbox/... ./internal/modules/inappnotification/... ./cmd/taskd -count=1` 通过；反馈路由、权限映射和审计分类的定向测试通过。
- Admin：`npm run check:user-feedback`、`npm run build` 通过。
- H5App：`pnpm --config.verify-deps-before-run=false run type-check`、`build:h5` 及受影响文件的 ESLint 检查通过。此参数仅跳过本机运行脚本前自动重装依赖，不改变源码或依赖版本。
- 浏览器：使用隔离端口和模拟 API，在 1600×1000、900×800 下验证处理、站内通知、钉钉通知、未绑定提示、失败保留、幂等重试、详情记录和只读权限；没有调用真实通知接口。
- 全量 Backend 仍有范围外的 Swagger、钉钉配置、流程升级权限声明、审计分类、配置安全及 HTTP 错误边界失败；Admin `check:all` 停在现有流程实例版本升级弹窗检查；H5App 全量 lint 有 213 个既有错误，反馈全量脚本停在既有通知跳转调用顺序断言。本次相关专项检查通过，未修改这些范围外问题。
- 未部署生产环境，真实企业的钉钉消息送达未验证。
