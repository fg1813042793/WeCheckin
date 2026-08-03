# 项目维护说明

## 构建产物与依赖目录

以下目录属于本地依赖、构建产物或运行时产物，不应提交到 Git：

- `node_modules/`
- `dist/`
- `unpackage/`
- `.hbuilderx/`
- `.cache/`
- `backend/bin/`
- `backend/uploads/`
- `backend/logs/`

当前检查结果：

- `frontend/dist` 未被 Git 跟踪。
- `frontend/unpackage` 未被 Git 跟踪。
- `frontend/node_modules` 未被 Git 跟踪。
- `admin/node_modules` 未被 Git 跟踪。
- `.gitignore` 已覆盖这些路径。

如果后续需要保留构建包用于发布，请放到独立发布制品系统或压缩包归档，不要直接提交到源码仓库。

## 本地质量门禁

后续做功能或 UI 调整时，建议按影响范围执行下面的检查：

- 后端：在 `backend` 目录执行 `GOCACHE=$PWD/../.cache/go-build go test ./...`
- 管理后台：在 `admin` 目录执行 `npm run check:all`
- 客户端：在 `frontend` 目录执行 `npm run check:all`
- 全量本地复核：在项目根目录执行 `bash scripts/verify-local.sh`
- 快速项目检查：在项目根目录执行 `bash scripts/check.sh`

当前门禁覆盖：

- 根目录质量门禁自检：确认后台、客户端和全量复核脚本都接入了关键检查。
- 部署配置检查：覆盖 Docker Compose、Nginx `/api/v2` 代理、日志轮转、备份恢复脚本和部署文档关键片段。
- 管理后台请求封装、API v2 迁移、构建配置、导航、基础 UI 壳、用户列表、P2 UI、图标运行时、题库组件拆分、共享 FormKit 类型、生产构建与 bundle 预算。
- 客户端配置、API v2 迁移、请求封装、生产调试日志、认证工具、表单逻辑、固定搜索区和搜索输入样式检查。
- 后端全包单元测试与结构保护测试。

注意：这些命令只能覆盖代码结构、构建和静态检查；问卷设计器、考试设计器、题库可视化编辑等复杂交互仍需要人工在浏览器中确认。

## API 文档维护

- 当前新增接口必须走 `/api/v2` 或 `/api/v2/admin`，详见 [API v2 接口说明](API_V2.md)。
- 修改 `backend/cmd/routes_v2.go` 或 `backend/cmd/routes_v2_swagger.go` 后，需要在 `backend` 目录执行 `swag init -g main.go --dir ./cmd --parseDependency --output docs/swagger`。
- 更新 Swagger 后至少在 `backend` 目录运行 `GOCACHE=$PWD/../.cache/go-build go test ./cmd ./internal/middleware`，确认 v2 route operation 数量和权限声明没有漂移。
- 前端新增接口调用后运行 `npm --prefix admin run check:request` 或 `npm --prefix frontend run check:request`，确认没有重新引入旧路径。
