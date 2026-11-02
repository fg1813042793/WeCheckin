# 钉钉通知发送模板

## 配置与发送

在“通知记录管理 → 消息样式 → 发送模板”中新建模板。例如“绩效填写提醒”：填写默认标题和正文，消息类型选 OA，跳转目标选“发起流程”，再按流程管理名称搜索并选择已发布流程。候选项显示管理名称，搜索仍同时支持管理名称和用户显示名称。保存并启用模板。

回到“发送钉钉通知”，选择该模板。系统自动带入标题和正文，展示所选跳转目标；修改本次内容只影响此次发送。选择全部用户、部门或用户后，点击“发送”才触发钉钉投递。模板管理、切换、预览和刷新均不会发送消息。

“不使用模板”沿用原来的后台手动通知样式及企业 H5 入口。选用模板时，独立采用模板的 OA、链接或卡片类型与跳转目标，不会改动其他业务通知使用的全局样式。

候选加载失败不阻断明确未选模板的普通发送；已经选用模板的发送仍会被错误或过期版本阻止。弹窗中始终提供“刷新模板”，刷新不会覆盖已编辑的标题和正文；切换到新版模板需要确认替换。

模板管理支持新建、修改、复制、启用/停用、删除。修改或删除有版本检查；别人已修改的模板不能被旧编辑界面覆盖。发送时也校验所选版本，发现模板已修改、删除或停用时应刷新并重新选择，不会静默改用默认首页。

## 跳转目标

| 目标 | 行为 |
| --- | --- |
| 应用首页 | 使用收件人所属企业配置的 H5 应用入口 |
| 发起流程 | 使用流程定义 ID 打开 H5 的 `workflow:start:<id>` 页面，保存与发送时均检查流程仍已发布 |
| 自定义链接 | 使用完整 HTTPS 页面地址，不接受其他协议、用户名密码或控制字符；后端不请求该 URL |

系统依据收件人企业的 CorpID、App ID（只有 AgentId 时使用 `0_<AgentId>`）和 H5 地址生成钉钉 openapp 链接，无需操作者手写编码。流程目标会同时写入 HTTPS 顶层 `view` 查询参数和 uni-app hash 路由；即使钉钉客户端重写或丢弃 Fragment，H5App 仍可从顶层查询参数恢复目标。OA 的移动端 `message_url` 与桌面 `pc_message_url` 使用同一个生成地址。企业缺少跳转配置或目标不合法时应报告失败，不能悄悄回到首页。

OA 展示使用企业配置的“旧版工作通知（AgentId + OA）”通道；新版机器人通道沿用现有 sampleLink/卡片映射，跳转目标保持一致。

ActionCard 的可见正文会自动在顶部补充通知标题；正文模板已包含同一标题时不会重复。按钮仅支持文字和跳转地址，按钮颜色由钉钉客户端固定，当前 ActionCard 协议不提供颜色字段。

接收通知不意味着获得流程权限。收件人打开后仍需完成现有登录、用户绑定和发起权限校验；发起次数等限制同样有效。实际能否在目标钉钉客户端打开取决于已发布的 H5、企业应用及相关域名配置。

## 权限和接口

沿用既有权限，不自动新增或扩大授权。

| 方法与路径（前缀 `/api/v2/admin`） | 权限 |
| --- | --- |
| GET `/dingtalk-notification-templates` | `notification:style:list` |
| GET `/dingtalk-notification-templates/options` | `notification:dingtalk:send` |
| GET `/dingtalk-notification-templates/workflow-options` | `notification:style:edit` |
| POST `/dingtalk-notification-templates` | `notification:style:edit` |
| PUT `/dingtalk-notification-templates/{id}` | `notification:style:edit` |
| DELETE `/dingtalk-notification-templates/{id}?version=N` | `notification:style:edit` |

API 权限键带 `admin:api:` 前缀，菜单动作保持现有 `admin:menu:notification:style:list`、`admin:menu:notification:style:edit`、`admin:menu:notification:dingtalk-send`。

发送接口 POST `/dingtalk-notifications` 增加可选 `templateId` 与 `templateVersion`；指定模板时必须同时传递选择时的版本。标题、正文和收件范围仍按原有契约提交；服务端重新读取并校验模板，客户端不能通过发送请求覆盖目标链接或消息样式。模板版本在一次批次投递中保持一致，不同企业分别生成应用跳转链接。

模板操作归入通知审计，模板请求正文不写入普通访问日志。不会将模板 CRUD 当作实际消息发送事件。

## 上线

先备份并通过项目迁移工具执行 `backend/migrations/20261029120000_add_dingtalk_notification_templates.sql`，再发布后端和 Admin。主服务不自动建表；迁移仅建立模板表，不预置任何真实流程 ID、收件人或企业配置。

已有消息样式、通知记录及自动流程通知继续使用原数据。模板删除不会撤回已经发送的通知；客户端如何打开已发出的消息，以消息中当时生成的链接为准。
