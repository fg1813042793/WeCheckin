# API v2 接口说明

最后更新：2026-09-02

## 当前状态

项目已新增 `/api/v2` RESTful 接口层，并完成以下调用方迁移：

- PC 管理后台：`admin/src/api/index.ts`
- uni-app 客户端：`frontend/api/index.js`
- uni-app 内置移动端管理页：`frontend/api/admin.js`

后台管理只使用 `/api/v2/admin/*` 路由和统一权限体系；旧版 `/admin/*` 后台路由已不再作为兼容入口。`/passport/*`、`/home/*`、`/survey/*`、`/exam/*` 等历史客户端路由如仍存在，仅用于兼容旧页面和小程序旧代码。新增页面和新增接口调用必须使用 `/api/v2`。已有单点 MySQL 部署升级时，同时参考 [单点 MySQL 部署兼容升级说明](SINGLE_NODE_MYSQL_UPGRADE.md)。

## 基础约定

- 后端默认地址：`http://localhost:8083`
- Swagger 页面：`http://localhost:8083/swagger/index.html`
- Swagger JSON：`http://localhost:8083/swagger/doc.json`
- 客户端 API 前缀：`/api/v2`
- 后台 API 前缀：`/api/v2/admin`

请求认证：

- 客户端登录态使用 `Authorization` 请求头。
- 管理后台登录态也使用 `Authorization` 请求头，但请求路径必须以 `/api/v2/admin/` 开头，前端请求层会据此读取管理员 token。

响应格式保持项目统一结构：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {}
}
```

## 客户端公开接口

| 功能 | 方法 | 路径 |
| --- | --- | --- |
| 首页列表 | GET | `/api/v2/home` |
| 系统配置读取 | GET | `/api/v2/home/setup` |
| 用户扩展字段 | GET | `/api/v2/user-form-fields` |
| 微信/用户标识登录 | POST | `/api/v2/auth/login` |
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

## 客户端登录后接口

| 功能 | 方法 | 路径 |
| --- | --- | --- |
| 当前用户详情 | GET | `/api/v2/me` |
| 更新当前用户 | PUT | `/api/v2/me` |
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

### 钉钉 H5 流程附件

- `POST /api/v2/dingtalk/h5/workflows/attachments`：上传流程表单附件，文件字段名为 `file`，单文件最大 20MB。
- 支持 JPG/JPEG、PNG、GIF、WebP、PDF、TXT、CSV、Word、Excel、PowerPoint、ZIP、RAR 和 7Z；后端同时校验扩展名与文件内容特征。
- 新表单数据使用 `{ "id", "name", "url", "mimeType", "size" }[]` 保存附件元数据；后端和运行时继续接受历史 `string[]` 附件值，不修改历史实例数据。
- 接口权限为 `dingtalk_h5:api:workflow:attachment`，升级迁移会向已有流程发起或处理权限的角色和用户回填该权限。

## 前端调用入口

- PC 管理后台：`admin/src/api/index.ts`
- uni-app 客户端：`frontend/api/index.js`
- uni-app 移动端管理页：`frontend/api/admin.js`

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
