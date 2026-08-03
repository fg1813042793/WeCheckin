# 钉钉H5绩效系统

该模块将 `.test/performance-system-2` 的绩效收集流程迁移到当前项目中，前端位于 `dingtalk-h5`，后端接口统一使用独立前缀：

```text
/api/v2/dingtalk/h5
```

模块不复用现有客户端、管理后台、问卷、考试接口，也不套用现有后台权限中间件。

## 数据表

人员复用现有 `users` 表；绩效专用表使用 `dingtalk_h5_perf_` 前缀：

- `users`：钉钉 H5 人员直接复用现有用户表；账号使用 `user_mini_openid`，姓名使用 `user_name`，密码使用 `user_password`，角色、岗位、部门层级、直属上级、HRBP、负责部门写入 `user_obj.dingtalkH5Performance`。
- `dingtalk_h5_corp_configs`：钉钉 H5 多企业配置表，按 `corp_id` 保存企业名称、AppKey、AppSecret、AgentId 和启用状态。
- `dingtalk_h5_user_bindings`：钉钉 H5 用户绑定表，使用 `corp_id + dingtalk_user_id` 唯一定位本地 `users.id`，避免多企业 userId 串号。
- `dingtalk_h5_perf_reviews`：绩效考评单主表。
- `dingtalk_h5_perf_histories`：考评单流转记录。
- `dingtalk_h5_perf_templates`：默认目标、下月目标、价值观、分档系数模板。

旧版本创建过的 `dingtalk_h5_perf_users` 已废弃。执行 `backend/migrations/20260728110000_drop_dingtalk_h5_perf_users.sql` 会先把旧表人员合并到 `users.user_obj.dingtalkH5Performance`，再删除旧表。

`backend/init.sh` 维护脚本会补齐这些模型，并按 `schema_migrations` 跳过历史已执行迁移。对应手动迁移 SQL 位于：

```text
backend/migrations/20260727180000_add_dingtalk_h5_performance_tables.sql
```

钉钉 H5 登录态已迁移到 Redis，不再使用 `dingtalk_h5_perf_sessions` 表。登录有效期、Redis Key 前缀、单点登录开关、首次自助绑定开关和绩效流程通知开关由管理后台“钉钉应用管理 / 配置选项”维护，对应配置键为 `TOKEN_DINGTALK_H5_EXPIRE`、`TOKEN_DINGTALK_H5_REDIS_PREFIX`、`DINGTALK_H5_SINGLE_LOGIN`、`DINGTALK_H5_SELF_BIND_ENABLED`、`DINGTALK_H5_NOTIFY_ENABLED`。

## 钉钉免登流程

钉钉内打开 H5 时，前端先通过 `/api/v2/dingtalk/h5/public-config` 读取后台配置的默认企业 `corpId` 和页面品牌信息，再确定当前 `corpId`：URL 查询参数 `?corpId=xxx` 优先，其次使用后台默认企业，最后使用构建环境变量 `VITE_DINGTALK_CORP_ID`。随后前端通过 `requestAuthCode(corpId)` 获取一次性免登授权码，并调用 `/api/v2/dingtalk/h5/sso-login`，请求体包含 `corpId` 和 `authCode`。

后端按 `corpId` 查询 `dingtalk_h5_corp_configs`，使用该企业对应的 AppKey/AppSecret 换取钉钉用户身份。拿到 DingTalk UserID 后，后端通过 `dingtalk_h5_user_bindings.corp_id + dingtalk_user_id` 找到本地 `users.id`。映射成功后签发系统自己的 `dingtalk_h5` Redis token，并返回用户、菜单、按钮权限、接口权限和权限版本。

如果钉钉身份没有绑定本地用户，且 `DINGTALK_H5_SELF_BIND_ENABLED` 开启，`/sso-login` 返回 `code=10020` 和短期一次性 `bindTicket`。前端显示“绑定系统账号”页，用户输入系统账号和密码后调用 `/api/v2/dingtalk/h5/bind-self`。后端校验票据、账号密码、用户状态、钉钉身份唯一性和系统账号唯一性后写入 `dingtalk_h5_user_bindings`，再按正常登录流程签发 token。

免登只替代密码校验，不替代权限校验。用户仍必须在管理后台存在、状态启用，并配置角色权限或用户额外授权。后续所有业务接口仍要携带 `DT_H5_TOKEN`，并继续经过钉钉 H5 接口权限和数据权限校验。

管理后台“钉钉应用管理 / 配置选项”维护企业应用列表和登录态配置。企业应用列表写入 `dingtalk_h5_corp_configs`；第一条配置会同步到以下旧 setup 键，供单企业旧部署和兼容逻辑读取：

- `DINGTALK_H5_CORP_ID`
- `DINGTALK_H5_APP_KEY`
- `DINGTALK_H5_APP_SECRET`
- `DINGTALK_H5_AGENT_ID`
- `TOKEN_DINGTALK_H5_EXPIRE`
- `TOKEN_DINGTALK_H5_REDIS_PREFIX`
- `DINGTALK_H5_SINGLE_LOGIN`
- `DINGTALK_H5_SELF_BIND_ENABLED`

迁移脚本 `backend/migrations/20260731123000_add_dingtalk_h5_multi_corp.sql` 会把旧 setup 单企业配置回填到 `dingtalk_h5_corp_configs`，并将旧 `users.user_mini_openid` 回填为该企业的 `dingtalk_h5_user_bindings.dingtalk_user_id`。若历史账号不是 DingTalk UserID，可以由管理员在绑定表中补齐关系，也可以开启首次自助绑定后由用户在钉钉内输入系统账号密码完成绑定。

多 CorpId 部署时，每个企业在钉钉工作台配置自己的 H5 地址，例如：

```text
https://oa.example.com/dingtalk-h5/?corpId=ding123
https://oa.example.com/dingtalk-h5/?corpId=ding456
```

## 钉钉通知

绩效流程通知由管理后台“钉钉应用配置 / 配置 / 绩效流程通知”控制，默认关闭。开启前需要在“企业应用”中配置对应企业的 `AgentId`，并在“钉钉用户绑定”中维护本地用户与钉钉 UserID 的绑定关系。

当前通知触发点：

- 员工在“本月绩效”提交自评后，若流程进入 `manager_review`，后端会异步发送钉钉工作通知给该考评单的直属上级。

通知只作为提醒，不参与主事务。钉钉接口失败、上级未绑定钉钉账号、企业应用未配置 AgentId 时，员工提交不会回滚，后端只记录 `[DingTalkH5Notify]` 日志，便于排查。

## 流程状态

- `draft`：员工填写
- `manager_review`：上级评价
- `hrbp_review`：HRBP评价
- `employee_confirm`：员工确认
- `hr_final`：HRBP归档
- `completed`：已完成

## 前端交互规范

- 所有删除按钮都必须增加二次确认弹窗。用户确认后才可以执行删除接口或修改本地删除状态；用户取消时不能发起删除请求，也不能改变页面数据。
- 所有钉钉 H5 删除接口都必须走软删除，禁止物理删除业务数据。考评单使用 `deleted_at/delete_by/delete_dept_id` 标记删除并在列表、详情、统计、导出中统一过滤；共享 `users` 表的人员删除只允许停用 `user_status=0`。

## 接口

登录：

```text
POST /api/v2/dingtalk/h5/login
POST /api/v2/dingtalk/h5/logout
PATCH /api/v2/dingtalk/h5/account/password
```

启动数据与模板：

```text
GET /api/v2/dingtalk/h5/bootstrap
GET /api/v2/dingtalk/h5/template
```

`bootstrap` 仅返回当前登录用户 `user`、可见菜单/二级菜单 `menus`、页面按钮权限 `buttonPermissionKeys`、权限版本 `permissionVersion`，用于启动态校验和前端权限刷新判断；人员列表、考评单和模板分别通过 `/users`、`/reviews`、`/template` 按页面需要加载。

## 权限配置

钉钉 H5 使用统一权限表，不再按绩效角色硬编码按钮显示：

- 菜单权限：`dingtalk_h5:menu:*`，控制左侧一级菜单和绩效管理子菜单。
- 按钮权限：`dingtalk_h5:button:*`，控制创建、保存、提交、退回、归档、导出、删除、编辑模板等页面按钮。
- 接口权限：`dingtalk_h5:api:*`，由后端接口权限中间件校验。
- 数据权限：`data:all`、`data:dept`、`data:self`、`data:custom`、`data:extra`，控制绩效单和人员列表的数据范围；其中 `data:extra` 为用户级额外可见部门/用户，不属于角色基础范围枚举。

菜单名称和图标来自统一权限表的 `permission_name`、`permission_icon`。管理后台“权限管理 / 钉钉 H5”可以编辑目录、菜单的图标；H5 端支持的图标键为 `dashboard`、`performance`、`mine`、`history`、`manager`、`hrbp`、`summary`、`org`、`template`、`account`。

管理后台在角色管理和用户编辑里配置“应用权限”。按钮要显示并可用，必须同时配置按钮权限和对应接口权限。例如“我的绩效”的创建按钮需要 `dingtalk_h5:button:review:create` 和 `dingtalk_h5:api:review:create`。

如果当前用户没有配置可替他人创建的数据权限，创建考评单时只能选择自己；后端也会按当前用户数据范围再次校验。替他人创建时，考评单的创建归属用户写为被创建的员工，更新人和流转操作人仍记录实际操作用户。

绩效单：

“我的绩效”入口展示当前登录员工名下流程未完成的绩效单，即 `draft`、`manager_review`、`hrbp_review`、`employee_confirm`、`hr_final` 状态，不再按当前月份限定。

```text
GET    /api/v2/dingtalk/h5/reviews
POST   /api/v2/dingtalk/h5/reviews
GET    /api/v2/dingtalk/h5/reviews/:id
DELETE /api/v2/dingtalk/h5/reviews/:id
GET    /api/v2/dingtalk/h5/reviews/export
```

`POST /reviews` 支持 `employeeIds` 多选批量创建：

```json
{
  "employeeIds": ["nick", "cube"],
  "period": "2026-07",
  "nextPeriod": "2026-08"
}
```

绩效流程动作：

```text
POST /api/v2/dingtalk/h5/reviews/:id/save-self
POST /api/v2/dingtalk/h5/reviews/:id/submit-self
POST /api/v2/dingtalk/h5/reviews/:id/submit-manager
POST /api/v2/dingtalk/h5/reviews/:id/submit-hrbp
POST /api/v2/dingtalk/h5/reviews/:id/confirm-result
POST /api/v2/dingtalk/h5/reviews/:id/dispute-result
POST /api/v2/dingtalk/h5/reviews/:id/withdraw
POST /api/v2/dingtalk/h5/reviews/:id/return-employee
POST /api/v2/dingtalk/h5/reviews/:id/return-manager
POST /api/v2/dingtalk/h5/reviews/:id/return-hrbp
POST /api/v2/dingtalk/h5/reviews/:id/finalize
```

撤回提交必须传入 `returnReason`，后端会把原因写入流转记录，例如 `撤销员工自评提交：原因内容`，前端“查看流程进度”直接展示该历史记录。

组织架构：

```text
GET    /api/v2/dingtalk/h5/users
POST   /api/v2/dingtalk/h5/users
PUT    /api/v2/dingtalk/h5/users/:id
DELETE /api/v2/dingtalk/h5/users/:id
```

## 初始数据

当新表为空时，模块会初始化默认模板和演示账号。默认密码为 `123456`。初始化只在表为空时执行，不会每次启动重复插入。
