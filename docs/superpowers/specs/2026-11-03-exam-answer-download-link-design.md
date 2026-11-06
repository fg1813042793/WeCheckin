# 考试答卷通知下载链接设计

## 目标

考试交卷后的 `POST_STAT` 通知模版支持 `{answerDownloadUrl}`，收件人可通过限时免登录链接下载当前用户的单份考试答卷。下载格式和有效期在考试设计器“投放与结果”中配置。

## 用户配置

- `answerDownloadFormat`：`pdf`、`docx`、`xlsx`，默认 `pdf`。
- `answerDownloadExpireDays`：整数 `1-30`，默认 `7`。
- `answerDownloadBaseUrl`：下载服务地址，仅允许协议、域名和端口；通知模版使用 `{answerDownloadUrl}` 时必须配置，与考试链接无关。
- `{answerDownloadUrl}`：仅考试通知模版可用，代表当前提交记录的下载链接。
- 只有实际使用该变量的通知规则才创建 Redis 凭证；旧考试和旧模版行为保持不变。

## 链路与安全

1. 交卷记录持久化后，`poststat` 检查通知模版是否包含 `{answerDownloadUrl}`。
2. 使用 256 位随机 Token，Redis 只以 Token 的 SHA-256 摘要作为 key，value 记录考试 ID、答卷 ID、格式和过期时间，TTL 等于配置天数。
3. 链接只使用 `answerDownloadBaseUrl`，路径固定为 `/api/v2/public/exam-record-exports/download?token=...`，不得读取或回退到考试链接。
4. 公开接口不接受考试 ID、答卷 ID或格式参数；解析 Token 后调用现有单份答卷导出渲染。
5. Token 查询参数由现有访问日志敏感字段规则脱敏。Token 无效、过期、答卷已删除时不得返回答卷内容。
6. Outbox 保存已渲染正文，通知重试复用原链接，不刷新 TTL。

## 失败语义

- Redis 不可用、下载服务地址无效或凭证创建失败时，交卷本身仍成功；该条交卷后通知不入队，并记录内部错误。
- 无效或过期链接返回 HTTP 410 和用户可理解的中文提示。
- 凭证有效但答卷已删除时返回 HTTP 410。
- 导出渲染的内部故障返回 HTTP 500，不暴露内部错误。

## 管理端

- “投放与结果”增加下载格式、有效期和独立下载服务地址控件。
- 保存考试时校验格式、天数和 `answerDownloadBaseUrl`；模版使用下载变量时服务地址必填。
- 考试通知模版管理页和自定义逻辑编辑器展示 `{answerDownloadUrl}` 变量说明。

## 验证

- 设置归一化覆盖合法值、默认值和越界拒绝。
- Redis Store 覆盖保存 TTL、摘要 key、解析、过期和缺失。
- 通知渲染覆盖变量存在/不存在、人工评分和 Outbox 重试正文稳定性。
- 公开下载接口覆盖 PDF、Word、Excel、无效 Token、已删除答卷和内容响应头。
- Admin 契约检查覆盖“投放与结果”字段、保存/加载和模版变量提示。

