## ADDED Requirements

### Requirement: Swagger Tags 分类重组

所有接口的 Swagger `@Tags` 注释 SHALL 遵循统一的命名规范：`@Tags 客户端, 模块名` 或 `@Tags PC端, 模块名`，不再使用 `客户端 API` / `管理端 API` 作为二级分组。

#### Scenario: main.go 中 tag.name 定义更新

- **WHEN** 查看 `cmd/main.go` 中的 `@tag.name` 定义
- **THEN** `客户端 API` 改为 `客户端`，`管理端 API` 改为 `PC端`

#### Scenario: 客户端接口 Tags 统一为"客户端, 模块名"

- **WHEN** 查看任意客户端 handler 文件的 `@Tags` 注释
- **THEN** 格式统一为 `// @Tags 客户端, 模块名`

#### Scenario: 管理端接口 Tags 统一为"PC端, 模块名"

- **WHEN** 查看任意管理端 handler 文件的 `@Tags` 注释
- **THEN** 格式统一为 `// @Tags PC端, 模块名`

#### Scenario: 模块名去掉"-客户端"后缀

- **WHEN** 查看原 `赛事活动-客户端`、`考试-客户端`、`问卷-客户端` 的 Tags
- **THEN** 改为 `赛事活动`、`考试`、`问卷`

#### Scenario: 特殊 Tags 归类

- **WHEN** 查看 `// @Tags 表单工具`（无分类）
- **THEN** 客户端侧改为 `// @Tags 客户端, 表单工具`，管理端侧改为 `// @Tags PC端, 表单工具`
- **WHEN** 查看 `// @Tags 报名`（无分类）
- **THEN** 改为 `// @Tags 客户端, 报名`

### Requirement: 文档重新生成

修改完成后 SHALL 重新执行 `swag init` 生成最新文档。

#### Scenario: swag init 执行

- **WHEN** 在 backend 目录执行 `swag init`
- **THEN** `docs/swagger/swagger.yaml` 中的 tags 数组更新为 `客户端` 和 `PC端`，所有接口按新分类归组
