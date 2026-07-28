## Context

后端 API 采用 Go + Hertz 框架，Swagger 文档通过 `swag` 工具从代码注释自动生成。当前 143 个 API 已有完整注释，但另有约 56 个接口分布在 7 个文件中完全没有 Swagger 注释，导致这些接口在文档中缺失。此外，`/passport/my_detail` 接口的返回结构已变动（新增 `domain` 字段），文档需要同步更新。

## Goals / Non-Goals

**Goals:**

- 为 7 个缺失注释的文件补全 Swagger 注解
- 重新生成 swagger.yaml/swagger.json/docs.go
- 更新 `/passport/my_detail` 的返回定义，补充 `domain` 字段

**Non-Goals:**

- 不修改 API 的业务逻辑
- 不改变 API 的路由或参数结构
- 不引入新的 swag 配置或自定义模板

## Decisions

1. **使用 `swag init` 自动生成文档** — 项目中已用此方式，维持一致。不手动编辑 swagger.yaml。
2. **注释风格保持一致** — 所有新增注释遵循现有格式：`@Tags` 用逗号分隔、`@Success 200 {object} response.Resp`、`@Router /path [method]`。
3. **不拆分 handlers 文件** — 缺失注释的 7 个文件不动结构，仅在现有 handler 函数上方添加注释。
4. **`/passport/my_detail` 使用自定义返回结构** — 不再返回 `*model.User`，改为 `map[string]interface{}`，因此在 swagger 注释中使用 `@Success 200 {object} response.Resp` 保持通用，不定义具体 schema（因为 swag 对 map 返回不够友好，且前端关心的是具体的字段名，已在代码中通过字段注释解决）。
5. **`/api/v2` 通过集中路由和辅助 Swagger 注释生成文档** — v2 路由集中在 `backend/cmd/routes_v2.go`，通用 Swagger 注释集中在 `backend/cmd/routes_v2_swagger.go`，生成产物仍使用 `swag init -g cmd/main.go --output docs/swagger`。

## Risks / Trade-offs

- **swag 版本兼容性** — 若本地 `swag` CLI 版本与项目要求不同，生成的文档格式可能有差异。应在项目 README 或 Makefile 中声明所需版本。
- **手动注释工作量大** — 56 个接口需要逐个编写注释，约 2-3 小时工作量。但这是必要的一次性投入。
- **map 返回类型的 Swagger 文档** — `/passport/my_detail` 返回 `map[string]interface{}` 无法在 swagger 中精确描述响应结构。trade-off：保持通用 `response.Resp`，在前端文档中说明。
- **v2 注释集中化** — v2 路由数量较多，集中注释能保证 Swagger operation 数量和实际路由保持同步；缺点是具体参数细节仍需结合 handler 和 `docs/API_V2.md` 阅读。
