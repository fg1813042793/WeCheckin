## Why

后端约 56 个 API 接口缺少 Swagger 注释，导致 swagger.yaml 文档不完整，影响前端开发和第三方对接时的接口查阅效率。现有 143 个已注释的接口也需要统一校验和规范化。

## What Changes

- 为 7 个缺失 Swagger 注释的 handler 文件添加 `@Tags`、`@Summary`、`@Param`、`@Router` 等注解（约 56 个接口）
- 校验已有注释的接口，修正不规范的描述与参数定义
- 重新生成 swagger.yaml/swagger.json/docs.go 文档
- 更新 `/passport/my_detail` 接口的返回响应定义，补充 `domain` 字段

## Capabilities

### New Capabilities

- `event-api-docs`: 添加赛事活动模块客户端和管理端 API 的 Swagger 文档
- `geo-api-docs`: 添加地理编码客户端 API 的 Swagger 文档
- `department-api-docs`: 添加部门管理 API 的 Swagger 文档
- `menu-api-docs`: 添加菜单权限 API 的 Swagger 文档
- `dict-api-docs`: 添加字典管理 API 的 Swagger 文档
- `role-api-docs`: 添加角色管理 API 的 Swagger 文档

### Modified Capabilities

- `passport-api-docs`: `/passport/my_detail` 响应增加 `domain` 字段定义

## Impact

- `backend/internal/api/client/event.go`、`geo.go`
- `backend/internal/api/admin/admin_event.go`、`admin_department.go`、`admin_menu.go`、`admin_dict.go`、`admin_role.go`
- `backend/docs/swagger/swagger.yaml`、`swagger.json`、`docs.go`

## Current Status

2026-07-18 更新：项目已新增 `/api/v2` RESTful 路由层，Swagger 产物已重新生成并包含 v2 公开接口、客户端登录后接口和 `/api/v2/admin` 后台接口。当前接口总览以 `docs/API_V2.md` 与 `backend/docs/swagger/` 为准。
