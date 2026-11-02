# 钉钉手动通知模板

用户要求在“发送钉钉通知”中直接选择模板，并在“消息样式”管理模板，配置后能打开指定流程的发起页面。

## 交互

- 消息样式弹窗增加“业务消息样式 / 发送模板”页签。原业务消息样式继续服务于自动流程通知。
- 发送模板提供列表、新建、编辑、复制、启用/停用、删除。模板字段为名称、默认标题、默认正文、消息类型（OA / 链接 / 卡片）、跳转目标、OA 颜色和卡片按钮文字。
- ActionCard 可见正文必须展示通知标题；正文已包含标题时不重复。按钮颜色由钉钉客户端固定，不提供无效的颜色配置项。
- 跳转目标支持应用首页、已发布流程的发起页、完整 HTTPS 地址。流程通过可搜索选择器选取，管理员不拼 CorpID、AppID 或 URL 编码。
- 发送弹窗增加模板选择。默认“不使用模板（沿用系统通知样式）”；选中后带入标题、正文，展示消息类型和目标摘要。允许修改本次标题/正文，不修改模板。切换模板若有未发送编辑，先确认替换。取消后保持原选择和内容。
- 模板的目标与样式固定于所选版本；模板已停用、删除、修改或目标流程不再发布时，发送明确报错，不能静默回退首页。所有收件范围保留。
- 选择、管理、预览均不发送消息，只有发送按钮触发外部投递。

## 数据与权限

使用独立 `dingtalk_notification_templates` 表和版本化迁移，主服务不建表。每条模板有数值 ID、整数 version 和创建/更新人及时间。更新、删除校验 version，避免覆盖别人编辑；删除只影响未来选择，不撤回历史消息。

复用现有权限：模板管理读取对应 notification:style:list，增改删和流程候选对应 notification:style:edit；发送弹窗候选对应 notification:dingtalk:send。前后端一致，发送者无需获得模板编辑或工作流管理权限。

模板候选只返回启用项。流程候选只提供已发布定义的 ID/管理名称，分页检索；搜索同时匹配管理名称和用户显示名称，配置流程目标与发送时均校验发布状态。收件人打开页面后仍执行现有登录和流程发起权限校验。

## 契约

`DingTalkSendTemplate`：`id, name, enabled, title, content, messageType, target, buttonTitle, headColor, version, updatedAt`。

`target`：`{ kind: 'app_home' | 'workflow_start' | 'url', definitionId?: number, definitionName?: string, url?: string }`；definitionName 仅响应展示，输入不作为可信依据。

- GET `/api/v2/admin/dingtalk-notification-templates`：管理列表，`keyword,page,pageSize`，返回 `{list,total,page,pageSize}`，每页默认20、最多100条。
- GET `/api/v2/admin/dingtalk-notification-templates/options`：启用模板，同样支持关键词与分页，返回 `{list,total,page,pageSize}`；发送选择器须可检索/继续加载，不能静默截断。
- GET `/api/v2/admin/dingtalk-notification-templates/workflow-options`：`keyword,page,pageSize,id`，返回 `{list:[{id,name}],total,page,pageSize}`，id 用于解析编辑时已选流程。
- POST `/api/v2/admin/dingtalk-notification-templates`：创建；body 为模板可编辑字段，无 id/updatedAt。
- PUT `/api/v2/admin/dingtalk-notification-templates/:id`：更新；body 带 version。
- DELETE `/api/v2/admin/dingtalk-notification-templates/:id?version=N`：删除。
- POST `/api/v2/admin/dingtalk-notifications` 扩展可选 `templateId`、`templateVersion`；同时提交最终标题/正文与既有收件范围。不接受客户端伪造模板目标或样式。

名称最多80字，标题255字，正文5000字，URL2048字符，按钮20字，颜色8位十六进制。模板配置完整且可使用后才能保存；重复名称报错。自定义链接仅允许 HTTPS、无用户名密码、无控制字符；不在后端请求链接。

投递按收件企业配置生成链接。workflow_start 同时在 HTTPS 顶层 query 和 `#/pages/index/index?view=workflow:start:<id>` hash 路由中写入 `view`，并保留基础地址已有查询参数及 corpId；再通过既有 openapp 包装器生成企业应用链接。H5App 合并读取浏览器顶层 query、hash query 与 uni-app onLoad 参数，onLoad 参数优先。OA 的 message_url 与 pc_message_url 使用同一目标。自定义 URL 原样作为目标页面并经过相同包装。模板发送不套用全局 admin_manual 样式，未选择模板时保持原有行为。

## 验证

后端验证 CRUD、版本冲突、禁用/删除/失效流程拒绝、URL 校验、多企业链接及原路径兼容；发送测试使用假钉钉客户端。Admin 验证模板 CRUD、候选失败重试、选中/切换确认、临时内容修改和实际提交字段。浏览器使用模拟 API 在桌面与窄屏验证，不发送真实通知。
