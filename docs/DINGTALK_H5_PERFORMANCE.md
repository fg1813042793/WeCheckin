# 钉钉H5绩效系统

该模块将 `.test/performance-system-2` 的绩效收集流程迁移到当前项目中，前端位于 `dingtalk-h5`，后端接口统一使用独立前缀：

```text
/api/v2/dingtalk/h5
```

模块不复用现有客户端、管理后台、问卷、考试接口，也不套用现有后台权限中间件。

## 数据表

人员复用现有 `users` 表；绩效专用表使用 `dingtalk_h5_perf_` 前缀：

- `users`：钉钉 H5 人员直接复用现有用户表；账号使用 `user_mini_openid`，姓名使用 `user_name`，密码使用 `user_password`，角色、岗位、部门层级、直属上级、HRBP、负责部门写入 `user_obj.dingtalkH5Performance`。
- `dingtalk_h5_perf_sessions`：钉钉 H5 绩效模块登录会话。
- `dingtalk_h5_perf_reviews`：绩效考评单主表。
- `dingtalk_h5_perf_histories`：考评单流转记录。
- `dingtalk_h5_perf_templates`：默认目标、下月目标、价值观、分档系数模板。

旧版本创建过的 `dingtalk_h5_perf_users` 已废弃。执行 `backend/migrations/20260728110000_drop_dingtalk_h5_perf_users.sql` 会先把旧表人员合并到 `users.user_obj.dingtalkH5Performance`，再删除旧表。

`backend/init.sh` 维护脚本会补齐这些模型，并按 `schema_migrations` 跳过历史已执行迁移。对应手动迁移 SQL 位于：

```text
backend/migrations/20260727180000_add_dingtalk_h5_performance_tables.sql
```

## 流程状态

- `draft`：员工填写
- `manager_review`：上级评价
- `hrbp_review`：HRBP评价
- `employee_confirm`：员工确认
- `hr_final`：HRBP归档
- `completed`：已完成

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

`bootstrap` 仅返回当前登录用户 `user`、可见菜单/二级 tab `menus`、权限版本 `permissionVersion`，用于启动态校验和前端权限刷新判断；人员列表、考评单和模板分别通过 `/users`、`/reviews`、`/template` 按页面需要加载。

绩效单：

```text
GET    /api/v2/dingtalk/h5/reviews
POST   /api/v2/dingtalk/h5/reviews
GET    /api/v2/dingtalk/h5/reviews/:id
DELETE /api/v2/dingtalk/h5/reviews/:id
GET    /api/v2/dingtalk/h5/reviews/export
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

组织架构：

```text
GET    /api/v2/dingtalk/h5/users
POST   /api/v2/dingtalk/h5/users
PUT    /api/v2/dingtalk/h5/users/:id
DELETE /api/v2/dingtalk/h5/users/:id
```

## 初始数据

当新表为空时，模块会初始化默认模板和演示账号。默认密码为 `123456`。初始化只在表为空时执行，不会每次启动重复插入。
