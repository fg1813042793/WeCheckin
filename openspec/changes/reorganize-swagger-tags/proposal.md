## Why

Swagger 文档中接口 Tags 分类方式不统一，客户端接口分散在不同 Tags 下，管理端使用"管理端 API"作为二级分组。为提升文档可读性和检索效率，将所有客户端接口统一归入"客户端"大类、管理端接口归入"PC端"大类，再按模块细分。

## What Changes

- 修改 `cmd/main.go` 中的 `@tag.name` 定义：`客户端 API` → `客户端`，`管理端 API` → `PC端`
- 将所有 `// @Tags XXX, 客户端 API` 改为 `// @Tags 客户端, XXX`
- 将所有 `// @Tags XXX, 管理端 API` 改为 `// @Tags PC端, XXX`
- 特殊处理：`// @Tags 表单工具`（无分类）需确认归属；`// @Tags 报名`（无分类）需归入`客户端`
- 重新生成 swagger.yaml/swagger.json/docs.go

## Capabilities

### New Capabilities

- `swagger-tags-reorg`: Swagger 文档 Tags 分类重组

### Modified Capabilities

（无现有 spec 变更）

## Impact

- `backend/cmd/main.go`：`@tag.name` 定义
- `backend/internal/api/client/passport.go`、`event.go`、`geo.go`、`enroll.go`、`fav.go`、`news.go`、`home.go`
- `backend/internal/api/admin/*.go`（共 11 个文件）
- `backend/internal/survey/api/admin_survey.go`、`client_survey.go`、`admin_formkit.go`
- `backend/internal/exam/api/client_exam.go`、`admin_exam.go`
- `backend/docs/swagger/swagger.yaml`、`swagger.json`、`docs.go`
