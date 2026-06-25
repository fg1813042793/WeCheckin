## ADDED Requirements

### Requirement: 按功能模块组织 backend 代码

系统 backend 目录结构 SHALL 从按技术层（handler/service/model）改为按功能模块组织。每个业务模块 SHALL 在一个独立目录中包含其 handler、service 和 model 代码。共享模型和工具类 MAY 保留在 `internal/model/` 和 `pkg/` 中。

#### Scenario: 模块目录创建

- **WHEN** 检查 `backend/internal/` 下的结构
- **THEN** SHALL 包含以下模块目录：`passport`、`user`、`news`、`event`、`enroll`、`exam`、`survey`、`formkit`、`department`、`role`、`menu`、`dict`、`admin`、`setup`、`fav`、`geo`、`home`

#### Scenario: 模块内容自包含

- **WHEN** 查看任一模块目录
- **THEN** SHALL 包含该模块的 handler（原 `internal/api/admin/` 或 `internal/api/client/` 中的对应文件）
- **AND** SHALL 包含该模块的 service（原 `internal/service/` 中的对应文件）
- **AND** MAY 包含该模块的 model（若数据模型仅该模块使用）

### Requirement: 合并分散的模块代码

`internal/exam/api/` 和 `internal/exam/service/` 中的代码 SHALL 合并到 `internal/exam/` 目录。`internal/survey/api/` 和 `internal/survey/service/` 中的代码 SHALL 合并到 `internal/survey/` 目录。

#### Scenario: exam 模块合并

- **WHEN** 检查 `internal/exam/` 目录
- **THEN** SHALL 包含原先 `api/` 和 `service/` 下的所有 handler 和 service 代码
- **AND** SHALL 不再存在 `internal/exam/api/` 和 `internal/exam/service/` 子目录

#### Scenario: survey 模块合并

- **WHEN** 检查 `internal/survey/` 目录
- **THEN** SHALL 包含原先 `api/` 和 `service/` 下的所有 handler 和 service 代码
- **AND** SHALL 不再存在 `internal/survey/api/` 和 `internal/survey/service/` 子目录

### Requirement: 项目编译通过

所有 import 路径 SHALL 更新为新的模块路径。`cmd/main.go` 中所有 handler 和 service 的引用 SHALL 更新为新的 package 路径。项目 SHALL 可通过 `go build` 编译成功。

#### Scenario: 编译验证

- **WHEN** 在项目根目录执行 `go build ./...`
- **THEN** SHALL 编译成功，无 import 错误

### Requirement: API 路由保持不变

文件移动后，所有 HTTP API 的路由路径 SHALL 保持不变。客户端访问的 URL 路径不因代码重组而变化。

#### Scenario: 路由路径稳定

- **WHEN** 检查 `cmd/main.go` 中的路由注册
- **THEN** 所有路由路径（如 `/admin/user/list`、`/api/passport/login`）SHALL 与重组前一致

### Requirement: 清理根目录产物

构建产物和备份目录 SHALL 从 backend 根目录移除。

#### Scenario: 清理构建产物

- **WHEN** 检查 `backend/` 根目录
- **THEN** `main.exe`、`wecheckin-server`、`stderr.txt`、`stdout.txt` SHALL 不存在
- **AND** `mcloud_bak/` 目录 SHALL 不存在
