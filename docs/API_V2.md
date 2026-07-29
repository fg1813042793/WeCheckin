# API v2 接口说明

最后更新：2026-07-21

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
| 通知公告 | `/api/v2/admin/news` |
| 打卡项目 | `/api/v2/admin/enrollments` |
| 赛事活动 | `/api/v2/admin/events` |
| 字典 | `/api/v2/admin/dict` |
| 部门 | `/api/v2/admin/departments` |
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

完整方法、参数和响应以 Swagger 为准。

## 前端调用入口

- PC 管理后台：`admin/src/api/index.ts`
- uni-app 客户端：`frontend/api/index.js`
- uni-app 移动端管理页：`frontend/api/admin.js`

不要在页面中手写旧路径，例如：

- `/exam/list`
- `/survey/submit`
- `/passport/login_pwd`
- `/admin/user_list`

如确需新增接口，先在后端 `backend/cmd/routes_v2.go` 注册 `/api/v2` 路由，再同步前端 API 封装和 Swagger 文档。

## Swagger 更新

修改 v2 路由或 Swagger 注释后，在 `backend` 目录执行：

```bash
swag init -g cmd/main.go --output docs/swagger
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
GOCACHE=$PWD/.cache/go-build go test ./backend/cmd ./backend/internal/middleware
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
