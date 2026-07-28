## ADDED Requirements

### Requirement: 缺失的 API 接口添加 Swagger 注释

所有缺少 Swagger 注释的 handler 函数 SHALL 添加标准的 `@Tags`、`@Summary`、`@Param`、`@Success`、`@Router` 注释，以便 swag 工具能生成完整的文档。

#### Scenario: event 模块客户端 API 添加注释

- **WHEN** 执行 `swag init` 重新生成文档
- **THEN** `client/event.go` 中的 `GetEventList`、`ViewEvent`、`EventParticipate`、`GetMyEventList`、`GetMyEventRoles`、`GetMyManagedList`、`PostEventDynamic`、`GetEventDynamics`、`GetEventParticipantList`、`SaveEventScore`、`GetEventScores` 接口出现在 swagger.yaml 中

#### Scenario: geo 模块客户端 API 添加注释

- **WHEN** 执行 `swag init` 重新生成文档
- **THEN** `client/geo.go` 中的 `ReverseGeocode` 接口出现在 swagger.yaml 中

#### Scenario: admin_event 管理端 API 添加注释

- **WHEN** 执行 `swag init` 重新生成文档
- **THEN** `admin/admin_event.go` 中的约 21 个 API 接口出现在 swagger.yaml 中

#### Scenario: admin_department 管理端 API 添加注释

- **WHEN** 执行 `swag init` 重新生成文档
- **THEN** `admin/admin_department.go` 中的 `GetDeptTree`、`AddDept`、`EditDept`、`DelDept` 接口出现在 swagger.yaml 中

#### Scenario: admin_menu 管理端 API 添加注释

- **WHEN** 执行 `swag init` 重新生成文档
- **THEN** `admin/admin_menu.go` 中的约 7 个 API 接口出现在 swagger.yaml 中

#### Scenario: admin_dict 管理端 API 添加注释

- **WHEN** 执行 `swag init` 重新生成文档
- **THEN** `admin/admin_dict.go` 中的约 7 个 API 接口出现在 swagger.yaml 中

#### Scenario: admin_role 管理端 API 添加注释

- **WHEN** 执行 `swag init` 重新生成文档
- **THEN** `admin/admin_role.go` 中的 `GetRoleList`、`AddRole`、`EditRole`、`DelRole`、`DelRoles` 接口出现在 swagger.yaml 中

### Requirement: 已有注释的接口规范化

已有 Swagger 注释的 143 个接口 SHALL 通过人工校验确保参数描述、返回类型、Tags 分组一致且准确。

#### Scenario: 参数描述校验

- **WHEN** 检查所有已有 API 的 `@Param` 注释
- **THEN** 每个参数的 description、type、required 属性与实际代码逻辑一致

### Requirement: 文档生成命令

Swagger 文档 SHALL 通过统一的命令重新生成，确保 yaml/json/docs.go 三份文件保持同步。

#### Scenario: 使用 swag init 重新生成

- **WHEN** 在 backend 目录执行 `swag init`
- **THEN** `docs/swagger.yaml`、`docs/swagger.json`、`docs/docs.go` 全部更新且与代码注释一致

#### Scenario: API v2 路由出现在 Swagger 文档

- **WHEN** 在 backend 目录执行 `swag init -g cmd/main.go --output docs/swagger`
- **THEN** Swagger 文档包含 `/api/v2/home`、`/api/v2/auth/login`、`/api/v2/surveys/{id}`、`/api/v2/exams/{id}` 等公开接口
- **AND** Swagger 文档包含 `/api/v2/admin/users`、`/api/v2/admin/surveys`、`/api/v2/admin/exams`、`/api/v2/admin/survey-question-bank`、`/api/v2/admin/exam-question-bank` 等后台接口
- **AND** `backend/cmd/routes_v2_structure_test.go` 的 v2 operation 数量检查通过

## MODIFIED Requirements

### Requirement: `/passport/my_detail` 接口补充 domain 字段

`/passport/my_detail` 的 GET 响应 SHALL 在 `data` 中返回 `domain` 字段（静态资源域名），与 `user` 对象同级。

#### Scenario: 响应中包含 domain 字段

- **WHEN** 调用 `GET /passport/my_detail?user_id=xxx`
- **THEN** 响应 `data` 中包含 `"domain": "http://..."` 字段
