# API v2 接口说明

最后更新：2026-09-08

## 当前状态

项目已新增 `/api/v2` RESTful 接口层，并完成以下调用方迁移：

- PC 管理后台：`admin/src/api/index.ts`
- uni-app 客户端：`frontend/api/index.js`
- uni-app 内置移动端管理页：`frontend/api/admin.js`
- 钉钉 H5：`h5app/src/api/`，`h5app/` 是当前权威源码目录。

后台管理只使用 `/api/v2/admin/*` 路由和统一权限体系；旧版 `/admin/*` 后台路由已不再作为兼容入口。`/passport/*`、`/home/*`、`/survey/*`、`/exam/*` 等历史客户端路由如仍存在，仅用于兼容旧页面和小程序旧代码。新增页面和新增接口调用必须使用 `/api/v2`。已有单点 MySQL 部署升级时，同时参考 [单点 MySQL 部署兼容升级说明](SINGLE_NODE_MYSQL_UPGRADE.md)。

## 基础约定

公共问卷与考试入口使用 `GET /api/v2/public/activities`；后台使用 `/api/v2/admin/public-portal` 配置展示内容。权限、公开过滤、二维码、预览和迁移说明见 [公共活动入口](PUBLIC_ACTIVITY_PORTAL.md)。

用户反馈新增独立通知提交人接口：`POST /api/v2/admin/user-feedbacks/{id}/notify`，支持可编辑正文和 `internal` / `dingtalk` 渠道，复用反馈处理权限，详情见 [用户反馈处理与通知](USER_FEEDBACK_ACTIONS.md)。

- 后端默认地址：`http://localhost:8083`
- Swagger 页面：`http://localhost:8083/swagger/index.html`
- Swagger JSON：`http://localhost:8083/swagger/doc.json`
- 客户端 API 前缀：`/api/v2`
- 后台 API 前缀：`/api/v2/admin`
- 钉钉 H5 API 前缀：`/api/v2/dingtalk/h5`

请求认证：

- 客户端登录态使用 `Authorization` 请求头。
- 管理后台登录态也使用 `Authorization` 请求头，但请求路径必须以 `/api/v2/admin/` 开头，前端请求层会据此读取管理员 token。
- 钉钉 H5 登录态只允许通过 `Authorization` 请求头传递，兼容原始 Token 和 `Bearer <token>`，不接受 URL `?token=` 作为登录鉴权。公开答卷下载接口的 `?token=` 是绑定单份答卷的独立限时凭证，不属于登录 Token。
- `TOKEN_ADMIN_SLIDING_EXPIRE_ENABLED`、`TOKEN_USER_SLIDING_EXPIRE_ENABLED` 和 `TOKEN_DINGTALK_H5_SLIDING_EXPIRE_ENABLED` 分别控制 Admin、普通用户和钉钉 H5 会话是否在鉴权成功后刷新 Redis TTL。配置缺失或值为 `1` 时保持滑动过期；值为 `0` 时关闭续期，Token 按登录签发时设置的对应有效期固定到期。关闭开关不会重写现有 Token 的 TTL，现有 Token 从切换时剩余的 TTL 继续倒计时，新 Token 从登录时间开始固定倒计时。

响应格式保持项目统一结构：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {}
}
```

稳定接口的 `data` 使用命名 DTO 并在 Swagger 中声明。服务端错误在 HTTP 边界转换为稳定的公开文案；SQL、堆栈、本地路径和下游错误链只记入内部日志，不直接返回客户端。

用户新增和编辑（`POST /api/v2/admin/users`、`PUT /api/v2/admin/users/{id}`）会明确返回用户名重复、未填写登录密码、直属上级无效、角色/部门/用户/数据权限越界等已知业务原因。角色或直接权限超出管理员可授权范围时，错误保留对应权限 Key，例如“目标用户现有角色或所选角色超出当前管理员权限范围-dingtalk_h5:button:review:self_save”。接口权限通过不代表具有管理目标账号现有角色及数据范围的权限；即使保持原角色不变，也会校验目标账号的现有授权。未知内部错误仍返回通用提示，服务端分别以 `admin.user.add`、`admin.user.edit` 和请求 ID 记录错误；不向客户端透传数据库异常。

## 客户端公开接口

| 功能 | 方法 | 路径 |
| --- | --- | --- |
| 首页列表 | GET | `/api/v2/home` |
| 系统配置读取 | GET | `/api/v2/home/setup` |
| 用户扩展字段 | GET | `/api/v2/user-form-fields` |
| 获取钉钉登录配置 | GET | `/api/v2/auth/dingtalk-config` |
| 创建钉钉授权参数 | POST | `/api/v2/auth/dingtalk-authorization` |
| 使用钉钉授权码登录 | POST | `/api/v2/auth/dingtalk-login` |
| 密码登录 | POST | `/api/v2/auth/password-login` |
| 注册 | POST | `/api/v2/auth/register` |
| 逆地理编码 | GET | `/api/v2/geo/reverse` |
| 字典类型 | GET | `/api/v2/dict/types` |
| 字典项 | GET | `/api/v2/dict/items` |
| 活动列表 | GET | `/api/v2/events` |
| 活动详情 | GET | `/api/v2/events/{id}` |
| 问卷列表 | GET | `/api/v2/surveys` |
| 问卷详情 | GET | `/api/v2/surveys/{id}` |
| 提交问卷 | POST | `/api/v2/surveys/{id}/responses` |
| 问卷逻辑应用 | POST | `/api/v2/survey/apply` |
| 问卷答案校验 | POST | `/api/v2/survey/validate` |
| 考试列表 | GET | `/api/v2/exams` |
| 考试详情 | GET | `/api/v2/exams/{id}` |
| 提交考试 | POST | `/api/v2/exams/{id}/submissions` |
| 考试答案校验 | POST | `/api/v2/exams/{id}/validation` |
| 按 session 查看考试结果 | GET | `/api/v2/exam-results` |

问卷设置支持可选字符串 `settings.displayTitle`（显示标题）。管理端创建、编辑问卷时沿用 `settings` JSON 字符串保存；客户端问卷详情在 `settings` 对象中返回。填写页与设计器预览优先显示去除首尾空白后的显示标题，未配置、为空或不是字符串时回退到 `title`。`title` 仍是后台列表和统计使用的原标题，历史问卷无需迁移。

问卷答卷的 `browser`、`deviceType`、`platformType` 根据提交的 `device`（原始 User-Agent）解析。移动系统标记优先于 `Mac OS X`、`Linux` 等兼容标记；没有具体浏览器名称的嵌入环境显示为 `iOS WebView`、`Android WebView` 或 `WebKit`，信息不足时保留 `-`。后台读取答卷列表、详情和导出数据时，会根据原始 UA 补全缺失信息，并纠正历史 iOS/macOS、Android/Linux 误判，不覆盖原始 UA，也不要求用户重新提交。

客户端钉钉登录的企业配置、用户绑定、回调地址和多企业行为见 [客户端钉钉验证登录](CLIENT_DINGTALK_LOGIN.md)。公开的用户 ID 直接登录入口已移除；密码登录保留为备用方式。

### 考试管理名称与显示名称

考试的 `title` 是后台管理名称，`settings.displayTitle` 是考生显示名称（可选、最多100字）。管理端列表、记录和管理导出保留管理名称；编辑预览、PC 两种答题模版及 uni-app 考生列表/答题/结果页优先使用去除首尾空白的显示名称，未配置时回退管理名称。考生列表关键词可匹配两种名称。保存接口校验该设置并保留其他设置字段，不新增数据库列。

### 考试答卷导出

后台单份答卷导出的 Word、PDF 使用逐题阅读版式：顶部紧凑展示考生、成绩、状态与时间，按试卷分组排列题目；题目、得分与判题结果、用户答案、参考答案分层展示，长答案可以自然跨页。Excel 保留五列表格供筛选和对照。未配置参考答案显示“未设置参考答案”，考生没有回答仍显示“未作答”；待人工评分不再显示为零分，零满分项显示“不计分”。这些调整只改变导出呈现，不重算或回写答卷评分，也不改变流程表单导出版式。

考试设计器“投放与结果”可配置通知答卷下载格式（PDF、Word、Excel）、链接有效期（1 至 30 天，默认 7 天）和答卷下载服务地址。服务地址仅允许包含协议、域名和端口，例如 `http://101.37.205.158:8086`；它只用于答卷下载，与考试链接无关。通知模版使用 `{answerDownloadUrl}` 时该地址必填。交卷时为当前答卷生成 Redis 限时凭证并渲染免登录下载地址；未使用该变量的规则不创建凭证。公开下载接口为 `GET /api/v2/public/exam-record-exports/download?token=...`，考试 ID、答卷 ID 和格式均从凭证读取，不能由调用方覆盖。通知重试复用原链接且不刷新有效期；Redis 数据被清理、凭证过期或答卷删除后链接失效。

考试提交通知的 `{submitter}` 从当前已保存答卷的用户信息题中读取，按姓名、手机、邮箱、学号、工号依次取首个非空字符串（去除首尾空白），与问卷通知一致；不按普通题目的标题猜测身份。没有对应答案时回退到登录用户姓名，姓名不可用时显示“用户<ID>”，未登录且无身份答案时显示“匿名用户”。钉钉 Webhook 与站内通知共用此规则；仅影响新生成的通知，不修改已发送消息。

### 考试题目不计分

考试设计器的计分方式支持显式 `examScoreMode: "none"`。选择“不计分”或清空选择时保存该值，题目继续参与表单填写、必填校验和答题导航，答案正常保存，但不计入总分、得分、正确题数和人工待批改数量。该值优先于题型、旧题目分值和选项/填空内的残留分值；空计分方式的历史数据仍按原有规则处理。

Schema 评分结果保留原始题目 ordinal，不因跳过不计分题或分隔线而移动其他题目的结果。题库试卷路径也尊重题目 schema 中的 none。查看/导出无评分结果的不计分题时显示“不计分”，不显示待人工评分或参考答案；已有历史评分结果优先，不因当前定义后来改成 none 而回写或重算成绩。此项不新增数据库列，需同步更新后端、Admin 及使用考试功能的客户端。

后台批改上下文的 `questions` 同时包含计分题和不计分题；`unscored: true` 表示只展示题目、材料与考生答案，不显示参考答案或分数输入，也不进入人工评分提交。新交卷的 v2 题目快照保存全部可作答题目（排除分隔线等布局项），历史 v1 快照仍可读取。历史 schema 答卷缺失的不计分题，仅在当前配置可按答案 key 匹配、确有保存答案且不与原快照或评分记录冲突时补充，并显示“按当前配置补充展示”提示；不回写快照、答案或成绩。已删除或无法可靠匹配的历史题目不猜测恢复。

### 考试记录客户端信息

考试记录复用问卷的 User-Agent 识别规则，返回 `addIp`（来源 IP）、`device`（原始 UA）、`browser`、`deviceType`、`platformType`。开始试卷考试与交卷时均采集 IP/UA；请求中的 `device` 优先，缺失或空白时读取 HTTP `User-Agent`，最多保存 512 个字符，IP 由服务端 `ClientIP()` 解析。

浏览器、设备类型和平台类型在读取记录时由已有 `exam_r_device` 派生，不增加数据库列。后台列表、详情及客户端记录接口使用一致的标签；历史记录有 UA 时直接解析，无 UA 或无法识别的部分显示 `-`，不根据当前查看者的设备补造历史信息。

后台考试记录列表和详情展示这四项客户端信息，详情另可查看原始 UA。需要同时更新后端与管理端，无需执行数据库迁移；本次不改变评分、权限、分页或导出格式。

### 考试答卷人工批改

`GET/POST /api/v2/admin/exams/{id}/records/{recordId}/grading` 提供人工批改上下文和提交；需要独立 `exam-record:grade` 权限及考试数据范围。POST 使用 JSON `{expectedVersion, scores:[{questionId,score}], comment}`，服务端校验全部人工题、整数分值、题目满分与版本，并在事务中更新成绩和批改状态。已提交答案及已完成批改不能被重复交卷覆盖。

该功能需先执行人工批改迁移，权限、快照兼容和部署顺序见 [考试答卷人工批改](EXAM_MANUAL_GRADING.md)。

## 客户端登录后接口

管理端登录安全和操作审计的事件字段、只读查询、独立权限、导出上限、可信代理 IP 及升级步骤见 [审计日志说明](AUDIT_LOGS.md)。操作日志使用 `/api/v2/admin/logs`，登录安全日志使用 `/api/v2/admin/login-logs`，各自支持 `/{id}` 和 `/export`；旧日志删除入口只返回拒绝结果。用户、管理员和角色列表分别新增 `/users/export`、`/managers/export`、`/roles/export`，保持原列表的数据范围与查看权限，支持受审计的当前分页导出。

日志中的 `snapshotStatus`、`snapshotReason` 独立表示快照采集状态和原因；历史记录默认为 `legacy_unknown`。普通配置只有通过明确白名单与标量校验才返回真实值；敏感值继续脱敏，旧隐藏值不还原。部署前需执行 `20261029150000_add_audit_snapshot_status.sql`。

| 功能 | 方法 | 路径 |
| --- | --- | --- |
| 当前用户详情 | GET | `/api/v2/me` |
| 更新当前用户 | PUT | `/api/v2/me` |
| 修改当前用户密码 | PATCH | `/api/v2/me/password` |
| 获取手机号 | POST | `/api/v2/me/phone` |
| 退出登录 | POST | `/api/v2/me/logout` |
| 我的收藏 | GET | `/api/v2/me/favorites` |
| 新增/更新收藏 | POST | `/api/v2/me/favorites` |
| 删除收藏 | DELETE | `/api/v2/me/favorites/{oid}` |
| 检查收藏 | GET | `/api/v2/me/favorites/check` |
| 我的打卡项目 | GET | `/api/v2/me/enrollments` |
| 我的打卡用户 | GET | `/api/v2/me/enrollment-users` |
| 我的打卡记录 | GET | `/api/v2/me/enrollment-records` |
| 我的打卡日历 | GET | `/api/v2/me/enrollment-calendar` |
| 我的某日打卡记录 | GET | `/api/v2/me/enrollment-day-records` |
| 我的活动/赛事 | GET | `/api/v2/me/events` |
| 我的活动角色 | GET | `/api/v2/me/event-roles` |
| 我管理的活动/赛事 | GET | `/api/v2/me/managed-events` |
| 我的问卷答卷 | GET | `/api/v2/me/survey-responses` |
| 答卷详情 | GET | `/api/v2/me/survey-responses/{id}` |
| 我的考试记录 | GET | `/api/v2/me/exam-records` |
| 通知列表 | GET | `/api/v2/news` |
| 通知分类 | GET | `/api/v2/news/categories` |
| 通知详情 | GET | `/api/v2/news/{id}` |
| 打卡列表 | GET | `/api/v2/enrollments` |
| 打卡详情 | GET | `/api/v2/enrollments/{id}` |
| 某日打卡数据 | GET | `/api/v2/enrollments/{id}/join-days` |
| 执行打卡 | POST | `/api/v2/enrollments/{id}/joins` |
| 提交报名表单 | POST | `/api/v2/enrollments/{id}/submissions` |
| 参与活动 | POST | `/api/v2/events/{id}/participants` |
| 参与用户 | GET | `/api/v2/events/{id}/participants` |
| 活动动态 | GET | `/api/v2/events/{id}/dynamics` |
| 发布动态 | POST | `/api/v2/events/{id}/dynamics` |
| 活动成绩 | GET | `/api/v2/events/{id}/scores` |
| 保存成绩 | POST | `/api/v2/events/{id}/scores` |
| 开始考试 | POST | `/api/v2/exams/{id}/start` |
| 考试记录详情 | GET | `/api/v2/exam-records/{id}` |
| 保存考试答案 | PUT | `/api/v2/exam-records/{id}/answers` |

## 管理后台接口

管理后台统一使用 `/api/v2/admin` 前缀。常用资源如下：

密码策略通过 `GET/PUT /api/v2/admin/password-policy` 读取和保存。策略默认关闭；启用后可配置最小长度、字符要求、有效期、到期提醒、最近密码数量及首次登录强制改密。管理员、普通客户端和钉钉 H5 登录响应统一返回 `passwordChangeRequired`、`passwordExpired`、`passwordExpiresAt` 与 `passwordExpiryWarning`。必须改密的会话只能访问当前账号改密和退出接口，其他请求返回 `errorCode=PASSWORD_CHANGE_REQUIRED`。普通客户端使用 `PATCH /api/v2/me/password`，Admin 使用 `PATCH /api/v2/admin/me/password`，钉钉 H5 使用 `PATCH /api/v2/dingtalk/h5/account/password`。

密码修改或管理员重置密码成功后，服务端立即撤销该用户在 Admin、普通客户端和钉钉 H5 的全部登录 Token，包括发起改密的当前会话；调用端清除本地登录态并返回登录页。账号状态变为非正常状态或账号被删除时执行相同撤销，重新启用账号不会恢复旧 Token，用户必须重新登录。

用户列表通过 `PATCH /api/v2/admin/users/{id}/password` 重置密码，表单参数 `password` 为管理员输入的新密码（至少 6 位且不超过 72 字节）。显式空值不合法；旧调用方省略该参数时仍兼容使用 `123456`，同样受密码策略约束。新密码不得与当前密码相同，启用策略后继续校验历史密码和适用的临时密码强度规则。Admin 弹窗会展示具体错误并保留输入供修改重试；未知内部错误返回“密码重置失败，请稍后重试”，服务端以 `admin.user.reset_password` 和请求 ID 记录错误。路由及按钮继续使用既有用户编辑权限和目标用户数据范围。

| 资源 | 路径前缀 |
| --- | --- |
| 登录/退出 | `/api/v2/admin/auth` |
| 控制台 | `/api/v2/admin/home` |
| 用户 | `/api/v2/admin/users` |
| 在线用户 | `/api/v2/admin/user-sessions` |
| 管理员 | `/api/v2/admin/managers` |
| 在线管理员 | `/api/v2/admin/admin-sessions` |
| 日志 | `/api/v2/admin/logs` |
| 系统设置 | `/api/v2/admin/settings` |
| 密码策略 | `/api/v2/admin/password-policy` |
| 后台文件上传 | `/api/v2/admin/uploads` |
| 通知公告 | `/api/v2/admin/news` |
| 打卡项目 | `/api/v2/admin/enrollments` |
| 赛事活动 | `/api/v2/admin/events` |
| 字典 | `/api/v2/admin/dict` |
| 部门 | `/api/v2/admin/departments` |
| 岗位 | `/api/v2/admin/positions` |
| 角色 | `/api/v2/admin/roles` |
| 权限管理 | `/api/v2/admin/permissions` |
| 当前管理员菜单/权限 | `/api/v2/admin/me` |
| 钉钉通知发送模板 | `/api/v2/admin/dingtalk-notification-templates` |
| 问卷 | `/api/v2/admin/surveys` |
| 问卷答卷 | `/api/v2/admin/survey-responses` |
| 问卷题库 | `/api/v2/admin/survey-question-bank` |
| 问卷资源 | `/api/v2/admin/survey-resources` |
| 考试 | `/api/v2/admin/exams` |
| 考试题库 | `/api/v2/admin/exam-question-bank` |
| 考试资源 | `/api/v2/admin/exam-resources` |
| 流程定义 | `/api/v2/admin/workflow-definitions` |
| 定时任务 | `/api/v2/admin/scheduled-tasks` |
| 定时任务运行记录 | `/api/v2/admin/scheduled-task-runs` |
| 定时任务执行节点 | `/api/v2/admin/scheduled-task-workers` |

钉钉手动通知可选择独立发送模板：`GET /dingtalk-notification-templates/options` 提供启用模板，`POST /dingtalk-notifications` 通过可选 `templateId`、`templateVersion` 选择模板，并提交本次标题、正文及收件范围。服务端校验模板版本、启用状态和流程发布状态，按收件企业生成跳转地址。模板管理复用通知样式查看/编辑权限，发送候选复用钉钉发送权限；完整接口、使用与迁移说明见 [钉钉通知发送模板](DINGTALK_NOTIFICATION_TEMPLATES.md)。

流程草稿与已发布流程响应支持可选 `historyImport`：配置历史来源、状态、字段/明细列映射及覆盖策略。保存复用既有流程更新接口和即时发起配置，不修改历史发布版本；没有配置或未启用时保留 H5 手工导入向导。契约与使用方式见 [历史表单导入规则](WORKFLOW_HISTORY_IMPORT_RULES.md)。

字典接口约定：

- `/api/v2/dict/types` 和 `/api/v2/dict/items` 只返回启用的字典类型与字典项，并隐藏历史版本用于表示类型的空值占位记录。
- 管理端通过 `POST /api/v2/admin/dict/types` 创建类型，通过 `PUT /api/v2/admin/dict/types/{typeCode}` 修改名称、状态和备注；`typeCode` 创建后不可修改。
- `DELETE /api/v2/admin/dict/types/{typeCode}/items` 只清空数据并保留类型；`DELETE /api/v2/admin/dict/types/{typeCode}` 删除类型及其数据。

后台配置与上传约定：

- 管理后台读取内容配置使用 `GET /api/v2/admin/settings/content?key=...`，需要 `setup:list`，不再借用公开的 `/api/v2/home/setup`。
- 图片和视频上传统一使用 `POST /api/v2/admin/uploads`，需要 `upload:create`，文件字段名为 `file`，最大 20MB，并校验扩展名与文件内容类型。
- `/upload` 仅保留给尚未迁移的历史客户端；Admin 页面不得再调用该匿名兼容入口。

完整方法、参数和响应以 Swagger 为准。定时任务的运行语义、处理器安全边界和部署方式见 [通用定时任务](SCHEDULED_TASKS.md)。

### 流程定义 Logo

`POST /api/v2/admin/workflow-definitions`、`POST /api/v2/admin/workflow-definitions/:id/copy` 和 `PUT /api/v2/admin/workflow-definitions/:id` 保留原有 JSON 请求格式，同时支持 `multipart/form-data`：

- 文本字段使用 `key`、`name`、`category`、`description` 和可选的 `draft`。
- 图片字段使用 `logo`，仅支持 PNG、JPG/JPEG、WebP，最大 2MB。
- 修改时传 `removeLogo=true` 可移除当前 Logo；同时提交图片时以新图片为准。
- 列表和详情响应通过 `logoUrl` 返回可访问的完整图片地址；Logo 不进入设计草稿和 BPMN，但发布时会随名称、分类和说明写入版本元数据快照。

复制流程定义时只复用源流程当前的设计草稿。新流程的 `key`、`name`、`category`、`description` 和 `logo` 均由本次请求重新提供，不继承源流程信息；新流程固定以草稿状态和版本 `0` 创建，不复制发布版本及版本历史。该接口复用 `workflow:add` 权限。

### 流程定义版本

- `GET /api/v2/admin/workflow-definitions/{id}/versions`：查询版本名称、发布说明、自动变更摘要、引用数量以及删除可用状态。
- `GET /api/v2/admin/workflow-definitions/{id}/versions/{version}/changes`：查询相对发布基准版本的结构化变更；可使用 `compareTo` 指定其他对比版本。
- `DELETE /api/v2/admin/workflow-definitions/{id}/versions/{version}`：物理删除未被流程实例或发起草稿引用的非当前版本；版本号不会重排或复用。
- `POST /api/v2/admin/workflow-definitions/{id}/versions/{version}/rollback`：复制目标版本内容并发布为新版本，历史版本和运行中实例保持不变。
- 旧版本没有元数据快照时，名称从其设计器 JSON 恢复，分类、说明和 Logo 沿用当前流程信息，不伪造历史值。

### 钉钉 H5 任务提交错误

`POST /api/v2/dingtalk/h5/workflows/tasks/{id}/complete` 沿用业务响应信封。提交失败时，H5 审批页显示接口公开的错误原因，保留未提交表单并恢复操作按钮，不自动重试。页面处理“通过/提交办理”的反馈时关闭该请求的公共 toast，避免重复提示覆盖具体原因。

流转过程中未解析到处理人时，返回 `code: 1` 和“流程节点未解析到处理人，请联系管理员检查审批人配置”；事务回滚，本次任务不会被标记为已完成。未知内部错误继续使用通用提示，SQL、堆栈和内部错误链不向用户公开。

### 无审批人时自动跳过审批节点

流程定义的审批节点可设置 `skipIfNoAssignee: true`，缺省或 `false` 保持原行为。仅 `approval` 节点可启用，仍须配置合法的审批人来源；解析成功但候选为空（包含只含空白候选或逐级审批全部层级为空）时，沿原有流程连线继续，兼容单人、依次、并行和会签。解析报错、数据库异常及逐级审批“缺失时报错”策略仍然阻断，不因本开关跳过。

自动跳过会写入 `node_skipped` 历史记录及原因，节点进度为 `skipped`，不创建虚假任务、通过记录或该节点的待办通知。后续节点照常审批；若后续直接结束，实例正常完成。开关需保存并发布新版本，新发起实例使用新版本；已发起实例继续使用绑定的发布版本。不新增接口、权限或数据库列。

### 管理后台流程实例作废

- `POST /api/v2/admin/workflow-instances/{id}/void`：仅允许将已完成实例变更为 `voided`（已作废），请求体必须包含不超过 500 个字符的 `reason`。
- 作废保留实例、表单、任务和流转历史，客户端仍可查看并再次提交；流程完成次数只统计 `completed`，因此已作废实例不再占用次数。
- 存在进行中的完成后表单修订时拒绝作废，需先完成或取消修订。
- 作废会发布 `instance_voided` 生命周期事件，但不会自动回滚流程完成时已经执行的外部业务动作。

### 钉钉 H5 流程汇总

流程汇总使用独立的管理访问接口，不放宽“我的申请/待办/已处理/抄送”接口：

- `GET /api/v2/dingtalk/h5/workflows/summary/definitions`：查询全部已发布流程定义。
- `GET /api/v2/dingtalk/h5/workflows/summary/definitions/{id}`：查询指定已发布流程定义。
- `GET /api/v2/dingtalk/h5/workflows/summary/instances`：按流程定义、发起人、版本、状态和时间分页查询实例，`pageSize` 仅支持 20 或 50。
- `GET /api/v2/dingtalk/h5/workflows/summary/instances/{id}`：在数据范围内查询只读实例详情。
- `GET /api/v2/dingtalk/h5/workflows/summary/export`：导出当前页选中的最多 50 个实例，格式为 `pdf`、`xlsx` 或 `docx`。

汇总列表、详情和导出均按实例发起人实时所属部门应用统一 `data:*` 权限。批量 PDF/Word 返回 ZIP，批量 Excel 返回一个工作簿且每个实例一个工作表；任一实例不属于指定流程或超出数据范围时，整批拒绝。

### 钉钉 H5 流程催办

- `POST /api/v2/dingtalk/h5/workflows/instances/{id}/reminders`：由流程发起人提醒指定当前节点的待处理人，请求体为 `{ "nodeId": "approval_1" }`。
- 仅运行中流程的发起人可操作；串行、并行或会签均只通知当前处于 `pending` 状态的人工任务处理人，自己不会收到催办。
- 同一流程实例和节点每 30 分钟最多提醒一次，每日最多 3 次；成功催办会写入流转记录和通知 Outbox。

### 钉钉 H5 办理后修改表单

- `PATCH /api/v2/dingtalk/h5/workflows/instances/{id}/form-data`：允许实际办理过流程节点的用户在流程结束前修改授权字段。
- 流程设计节点必须开启“办理完成后允许修改表单”；可修改字段取该用户已办理且已开启能力节点的 `write` 字段并集，参与路由条件判断的字段始终不可修改。
- 请求体包含 `expectedRevision`、非空 `formData` 补丁和必填 `reason`。服务端使用 `expectedRevision` 做乐观锁校验，普通任务提交导致表单变化时也会递增版本。
- `notification` 可省略；启用时收件人只能选择该流程的发起人、处理人或抄送人，渠道支持 `in_app` 和 `dingtalk_oa`。
- 修改、审计记录与通知 Outbox 在同一事务中提交；修改成功后会新增“表单已修改”流转记录。历史流程版本未配置该能力时保持不可修改。
- H5 入口同时要求 `dingtalk_h5:button:workflow:form-revise` 与 `dingtalk_h5:api:workflow:form-revise` 权限，后端接口权限是最终边界。

### 钉钉 H5 已完成流程表单修订

已完成流程不恢复为运行中状态。实际办理过且在流程设计中启用了“流程完成后允许修订”的人工节点处理人，可以从该节点发起独立表单修订：

- `GET /api/v2/dingtalk/h5/workflows/instances/{id}/form-revisions`：查询源实例的修订记录。
- `POST /api/v2/dingtalk/h5/workflows/instances/{id}/form-revisions/preview`：校验并预览修订模式与确认路径，不写数据库。
- `POST /api/v2/dingtalk/h5/workflows/instances/{id}/form-revisions`：创建修订请求；创建时重新执行与预览相同的完整校验。
- `GET /api/v2/dingtalk/h5/workflows/form-revisions/{id}`：查询修订前数据、补丁、候选表单和全部确认任务。
- `POST /api/v2/dingtalk/h5/workflows/form-revisions/{id}/cancel`：修订发起人在尚无确认任务被处理时取消请求。
- `POST /api/v2/dingtalk/h5/workflows/form-revision-tasks/{id}/complete`：实际任务处理人通过或驳回修订。

预览和创建请求示例：

```json
{
  "sourceNodeId": "manager_review",
  "expectedRevision": 3,
  "formData": {
    "finalResult": "修订后的结果"
  },
  "reason": "修正最终验收结果"
}
```

预览响应中的 `mode` 为 `direct` 或 `downstream_review`；`changedFields` 是变更字段标签，`confirmationStages` 是依据原实例实际经过的后续人工节点和当前组织关系解析出的确认阶段。配置为低风险直接修订的字段使用 `direct`，在创建事务内立即更新源实例表单并递增 `formRevision`；其他字段使用 `downstream_review`，全部确认通过且基准版本未变化时才生效。

修订状态为 `pending`、`applied`、`rejected`、`cancelled` 或 `conflict`；确认任务状态为 `waiting`、`pending`、`approved`、`rejected` 或 `cancelled`。任务动作只允许：

```json
{
  "action": "approve",
  "comment": "确认修订内容无误",
  "images": []
}
```

同一实例同一时间只允许一个 `pending` 修订。`expectedRevision` 与源实例当前版本不一致，或最终通过时源实例版本已变化，返回冲突；后者会将修订记为 `conflict`，不会覆盖较新的表单。相同处理人以相同动作重复处理已完成任务时幂等返回原结果，换动作或处理其他终态任务返回 409。

新增接口使用 HTTP 状态表达错误：参数、非法字段、无确认路径和不可取消为 400；非实际节点处理人或确认任务处理人为 403；实例、修订请求或任务不存在为 404；活动修订、表单版本变化和重复终态冲突为 409。创建和处理入口分别要求 `workflow:form-revision-create` 与 `workflow:form-revision-handle` 权限。

### 钉钉 H5 流程附件

- `POST /api/v2/dingtalk/h5/workflows/attachments`：上传流程表单附件，文件字段名为 `file`，单文件最大 20MB。
- 支持 JPG/JPEG、PNG、GIF、WebP、PDF、TXT、CSV、Word、Excel、PowerPoint、ZIP、RAR 和 7Z；后端同时校验扩展名与文件内容特征。
- 新表单数据使用 `{ "id", "name", "url", "mimeType", "size" }[]` 保存附件元数据；后端和运行时继续接受历史 `string[]` 附件值，不修改历史实例数据。
- 接口权限为 `dingtalk_h5:api:workflow:attachment`，升级迁移会向已有流程发起或处理权限的角色和用户回填该权限。

## 前端调用入口

- PC 管理后台：`admin/src/api/index.ts`
- uni-app 客户端：`frontend/api/index.js`
- uni-app 移动端管理页：`frontend/api/admin.js`
- 钉钉 H5：`h5app/src/api/`

不要在页面中手写旧路径，例如：

- `/exam/list`
- `/survey/submit`
- `/passport/login_pwd`
- `/admin/user_list`

如确需新增接口，先在后端 `backend/internal/routes/v2` 下按 `admin`、`client`、`dingtalkh5` 分类注册 `/api/v2` 路由，再同步前端 API 封装和 Swagger 文档。

## Swagger 更新

修改 v2 路由后，请同步 `backend/internal/routes/v2/swagger/swagger.go`，然后在 `backend` 目录执行：

```bash
swag init -g main.go --dir ./cmd,./internal/routes/v2/swagger --parseDependency --output docs/swagger
```

生成文件包括：

- `backend/docs/swagger/docs.go`
- `backend/docs/swagger/swagger.json`
- `backend/docs/swagger/swagger.yaml`

## 验证命令

接口迁移相关检查：

```bash
npm --prefix admin run check:request
npm --prefix frontend run check:request
cd backend
GOCACHE=$PWD/../.cache/go-build go test ./cmd ./internal/middleware
```

全量本地复核：

```bash
bash scripts/verify-local.sh
```

构建验证：

```bash
npm --prefix admin run build
npm --prefix frontend run build:h5
```
