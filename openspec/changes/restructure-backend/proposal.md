## Why

当前 backend 目录按技术层组织（api/、service/、model/），同时 exam、survey 等模块又内嵌自己的 api/service 子目录，导致逻辑分散、查找不便。这种混合结构在模块增多后维护成本上升，新功能难以快速定位归属。改为按功能模块组织，使每个模块自包含，提升可读性和可维护性。

## What Changes

- **移除** `internal/api/` 目录，handler 分散到各模块目录
- **移除** `internal/service/` 目录，业务逻辑分散到各模块目录
- **移除** `internal/model/` 目录，数据模型分散到各模块目录
- **移除** `internal/exam/api/`、`internal/exam/service/`，合入 `internal/exam/`
- **移除** `internal/survey/api/`、`internal/survey/service/`，合入 `internal/survey/`
- **创建** `internal/passport/`、`internal/user/`、`internal/news/`、`internal/event/`、`internal/enroll/`、`internal/exam/`、`internal/survey/`、`internal/formkit/`、`internal/department/`、`internal/role/`、`internal/menu/`、`internal/dict/`、`internal/admin/`、`internal/setup/`、`internal/fav/`、`internal/geo/`、`internal/home/`
- **更新** `cmd/main.go` 中所有 import 路径和路由注册
- **更新** `pkg/` 下代码的 import 路径
- **清理**根目录构建产物（`main.exe`、`wecheckin-server`、`stderr.txt`、`stdout.txt`）
- **移除** `mcloud_bak/` 备份目录

## Capabilities

### New Capabilities

- `backend-modular`: Backend 目录按功能模块化重组，包括 handler/service/model 的迁移

### Modified Capabilities

（无现有 spec 变更）

## Impact

- `backend/cmd/main.go`：所有 import 路径变更，路由函数引用路径变更
- `backend/internal/api/admin/*.go`（11 个文件）：迁移到各模块目录，更新 package 名
- `backend/internal/api/client/*.go`（8 个文件）：迁移到各模块目录，更新 package 名
- `backend/internal/service/*.go`（7 个文件）：迁移到各模块目录，更新 package 名
- `backend/internal/exam/api/*.go` + `backend/internal/exam/service/*.go`：合并到 `internal/exam/`
- `backend/internal/survey/api/*.go` + `backend/internal/survey/service/*.go`：合并到 `internal/survey/`
- `backend/internal/model/*.go`：模型拆散嵌入各模块或保留统一 model 包
- `backend/pkg/response/response.go`、`pkg/redis/redis.go` 等：可能涉及 import 路径调整
