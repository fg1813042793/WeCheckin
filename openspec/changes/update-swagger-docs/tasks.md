## 1. 客户端 event 模块 Swagger 注释

- [x] 1.1 为 `client/event.go` 中 `GetEventList` 添加 Swagger 注释（GET, Tags: 赛事活动-客户端, 客户端 API, Param: page/pageSize/user_id/keyword/type）
- [x] 1.2 为 `client/event.go` 中 `ViewEvent` 添加 Swagger 注释（GET, Param: id/user_id）
- [x] 1.3 为 `client/event.go` 中 `EventParticipate` 添加 Swagger 注释（POST, Param: event_id/user_id/forms）
- [x] 1.4 为 `client/event.go` 中 `GetMyEventList` 添加 Swagger 注释（GET, Param: user_id/type/status/page/pageSize）
- [x] 1.5 为 `client/event.go` 中 `GetMyEventRoles` 添加 Swagger 注释（GET, Param: user_id）
- [x] 1.6 为 `client/event.go` 中 `GetMyManagedList` 添加 Swagger 注释（GET, Param: user_id/type/status/keyword/page/pageSize）
- [x] 1.7 为 `client/event.go` 中 `PostEventDynamic` 添加 Swagger 注释（POST, Param: event_id/user_id/title/content/images/videos）
- [x] 1.8 为 `client/event.go` 中 `GetEventDynamics` 添加 Swagger 注释（GET, Param: event_id/page/pageSize）
- [x] 1.9 为 `client/event.go` 中 `GetEventParticipantList` 添加 Swagger 注释（GET, Param: event_id）
- [x] 1.10 为 `client/event.go` 中 `SaveEventScore` 添加 Swagger 注释（POST, Param: event_id/participant_id/score/judge_id）
- [x] 1.11 为 `client/event.go` 中 `GetEventScores` 添加 Swagger 注释（GET, Param: event_id/page/pageSize）

## 2. 客户端 geo 模块 Swagger 注释

- [x] 2.1 为 `client/geo.go` 中 `ReverseGeocode` 添加 Swagger 注释（GET, Tags: 地理编码, 客户端 API, Param: lat/lng）

## 3. 管理端 admin_event 模块 Swagger 注释

- [x] 3.1 为 `admin/admin_event.go` 中 `GetAdminEventList` 添加 Swagger 注释
- [x] 3.2 为 `admin/admin_event.go` 中 `GetAdminEventDetail` 添加 Swagger 注释
- [x] 3.3 为 `admin/admin_event.go` 中 `InsertEvent` 添加 Swagger 注释
- [x] 3.4 为 `admin/admin_event.go` 中 `EditEvent` 添加 Swagger 注释
- [x] 3.5 为 `admin/admin_event.go` 中 `DelEvent` 添加 Swagger 注释
- [x] 3.6 为 `admin/admin_event.go` 中 `DelEvents` 添加 Swagger 注释
- [x] 3.7 为 `admin/admin_event.go` 中 `StatusEvent` 添加 Swagger 注释
- [x] 3.8 为 `admin/admin_event.go` 中 `VouchEvent` 添加 Swagger 注释
- [x] 3.9 为 `admin/admin_event.go` 中 `TopEvent` 添加 Swagger 注释
- [x] 3.10 为 `admin/admin_event.go` 中 `GetEventParticipantList` 添加 Swagger 注释
- [x] 3.11 为 `admin/admin_event.go` 中 `DelEventParticipant` 添加 Swagger 注释
- [x] 3.12 为 `admin/admin_event.go` 中 `DelEventParticipants` 添加 Swagger 注释
- [x] 3.13 为 `admin/admin_event.go` 中 `EditEventParticipant` 添加 Swagger 注释
- [x] 3.14 为 `admin/admin_event.go` 中 `PostEventDynamic` 添加 Swagger 注释
- [x] 3.15 为 `admin/admin_event.go` 中 `GetEventDynamics` 添加 Swagger 注释
- [x] 3.16 为 `admin/admin_event.go` 中 `EditEventDynamic` 添加 Swagger 注释
- [x] 3.17 为 `admin/admin_event.go` 中 `DelEventDynamic` 添加 Swagger 注释
- [x] 3.18 为 `admin/admin_event.go` 中 `DelEventDynamics` 添加 Swagger 注释
- [x] 3.19 为 `admin/admin_event.go` 中 `GetEventScores` 添加 Swagger 注释
- [x] 3.20 为 `admin/admin_event.go` 中 `EditEventScore` 添加 Swagger 注释
- [x] 3.21 为 `admin/admin_event.go` 中 `GetDeptUsers` 添加 Swagger 注释

## 4. 管理端 admin_department 模块 Swagger 注释

- [x] 4.1 为 `admin/admin_department.go` 中 `GetDeptTree` 添加 Swagger 注释
- [x] 4.2 为 `admin/admin_department.go` 中 `AddDept` 添加 Swagger 注释
- [x] 4.3 为 `admin/admin_department.go` 中 `EditDept` 添加 Swagger 注释
- [x] 4.4 为 `admin/admin_department.go` 中 `DelDept` 添加 Swagger 注释

## 5. 管理端 admin_menu 模块 Swagger 注释

- [x] 5.1 为 `admin/admin_menu.go` 中 `GetMenuTree` 添加 Swagger 注释
- [x] 5.2 为 `admin/admin_menu.go` 中 `GetMenuList` 添加 Swagger 注释
- [x] 5.3 为 `admin/admin_menu.go` 中 `AddMenu` 添加 Swagger 注释
- [x] 5.4 为 `admin/admin_menu.go` 中 `EditMenu` 添加 Swagger 注释
- [x] 5.5 为 `admin/admin_menu.go` 中 `DelMenu` 添加 Swagger 注释
- [x] 5.6 为 `admin/admin_menu.go` 中 `GetAdminMenus` 添加 Swagger 注释
- [x] 5.7 为 `admin/admin_menu.go` 中 `GetAdminPerms` 添加 Swagger 注释

## 6. 管理端 admin_dict 模块 Swagger 注释

- [x] 6.1 为 `admin/admin_dict.go` 中 `GetDictTypes` 添加 Swagger 注释
- [x] 6.2 为 `admin/admin_dict.go` 中 `GetDictByType` 添加 Swagger 注释
- [x] 6.3 为 `admin/admin_dict.go` 中 `AddDictItem` 添加 Swagger 注释
- [x] 6.4 为 `admin/admin_dict.go` 中 `EditDictItem` 添加 Swagger 注释
- [x] 6.5 为 `admin/admin_dict.go` 中 `DelDictItem` 添加 Swagger 注释
- [x] 6.6 为 `admin/admin_dict.go` 中 `DelDictByType` 添加 Swagger 注释
- [x] 6.7 为 `admin/admin_dict.go` 中 `EditDictTypeName` 添加 Swagger 注释

## 7. 管理端 admin_role 模块 Swagger 注释

- [x] 7.1 为 `admin/admin_role.go` 中 `GetRoleList` 添加 Swagger 注释
- [x] 7.2 为 `admin/admin_role.go` 中 `AddRole` 添加 Swagger 注释
- [x] 7.3 为 `admin/admin_role.go` 中 `EditRole` 添加 Swagger 注释
- [x] 7.4 为 `admin/admin_role.go` 中 `DelRole` 添加 Swagger 注释
- [x] 7.5 为 `admin/admin_role.go` 中 `DelRoles` 添加 Swagger 注释

## 8. 更新 /passport/my_detail 文档

- [x] 8.1 校验 `passport.go` 中 `GetMyDetail` 的 Swagger 注释，确认响应包含 `domain` 字段说明

## 9. 重新生成 Swagger 文档

- [x] 9.1 在 backend 目录执行 `swag init` 重新生成 swagger.yaml、swagger.json、docs.go
- [x] 9.2 验证生成的文档中新增的 56 个接口已完整列出（共 199 个 paths）
- [x] 9.3 验证 `/passport/my_detail` 文档已更新

## 10. 同步 API v2 文档

- [x] 10.1 重新生成 Swagger，确认 `/api/v2` 和 `/api/v2/admin` 路由出现在 swagger.yaml、swagger.json、docs.go 中
- [x] 10.2 新增 `docs/API_V2.md`，说明 v2 接口前缀、调用入口、认证约定、Swagger 更新和验证命令
- [x] 10.3 更新 README、部署排障、HBuilder 调试、测试数据、维护规范和迁移说明中的接口口径
