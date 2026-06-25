## Context

当前 backend 目录混合了两种组织方式：
- **核心模块**按技术层组织：`internal/api/admin/`、`internal/api/client/`、`internal/service/`、`internal/model/`
- **exam/survey** 按功能组织：各自有 `api/`、`service/` 子目录
- **formkit** 已经按模块组织

这种不一致导致新功能无法快速确定文件归属，import 路径混杂，代码review 和重构成本高。

目标结构将统一为功能模块组织，每个模块目录包含 handler、service 和 model，模块间通过接口依赖，避免循环引用。

## Goals / Non-Goals

**Goals:**
- 所有 Go handler 按功能模块归类（passport、news、event、enroll、exam、survey、user 等）
- 每个模块自包含 handler + service + model（model 也可集中管理由团队决定）
- 清理根目录的构建产物和备份目录
- 更新 import 路径使项目可编译

**Non-Goals:**
- 不改变 API 路由路径（保持 `/admin/xxx`、`/api/xxx` 不变）
- 不重构业务逻辑或数据模型（仅移动文件、更新 package 名和 import）
- 不改变第三方依赖

## Decisions

### 1. 模块目录结构：扁平文件 vs 子目录
**选择：** 模块内扁平文件（`handler.go`、`service.go`、`model.go`）

理由：
- Go 的 package 名本身就是一层命名空间，无需再嵌套一层
- 大部分模块文件数在 1-5 个之间，扁平更简洁
- 避免 `internal/exam/api/` → `internal/exam/api/` 的递归嵌套

### 2. 模型（model）归属：分散到各模块
**选择：** model 拆散嵌入各模块

理由：
- 遵循"模块自包含"原则，每个模块的数据结构随模块一起维护
- 共享模型（如 response.Resp）保留在 `pkg/response/`
- 模块间共享的 model 可放在 `internal/model/`（如 User 被多个模块引用）

修正项：`User` 等在多个模块间共享的 model 保留集中 `internal/model/`。

### 3. 合并 admin/client handler
**选择：** 同一模块的 admin 和 client handler 合并到一个文件或同一 package

理由：
- news 有 admin_news.go 和 client/news.go → 合并为 `news/handler.go`（内部按功能函数命名区分）
- 同一模块的业务逻辑不应跨两个 package
- handler 注册仍在 `cmd/main.go` 中

### 4. pkg/ 保留不变
**选择：** `pkg/logger/`、`pkg/redis/`、`pkg/response/`、`pkg/tokenutil/` 保持原位

理由：
- pkg 是供外部使用的公共库，移至 internal 会失去复用能力
- 当前结构合理，无需变更

## 目标目录结构

```
backend/
├── cmd/main.go
├── internal/
│   ├── passport/         ← 通行证/认证（client handler + service）
│   ├── user/             ← 用户管理（admin handler + service）
│   ├── news/             ← 通知公告（admin+client handler + service）
│   ├── event/            ← 赛事活动（admin+client handler + service）
│   ├── enroll/           ← 打卡（admin+client handler + service）
│   ├── exam/             ← 考试（admin+client handler + service）
│   ├── survey/           ← 问卷（admin+client handler + service）
│   ├── formkit/          ← 表单工具（保持现有结构）
│   ├── department/       ← 部门管理（admin handler + service）
│   ├── role/             ← 角色管理（admin handler + service）
│   ├── menu/             ← 菜单管理（admin handler + service）
│   ├── dict/             ← 字典管理（admin handler + service）
│   ├── admin/            ← 管理员管理（admin handler + service）
│   ├── setup/            ← 系统设置（admin handler + service）
│   ├── fav/              ← 收藏（client handler + service）
│   ├── geo/              ← 地理编码（client handler）
│   ├── home/             ← 首页（admin+client handler + service）
│   ├── config/           ← 应用配置
│   ├── database/         ← 数据库初始化
│   ├── middleware/       ← HTTP 中间件
│   └── model/            ← 共享模型（User 等跨模块复用）
├── pkg/                  ← 保持不变
├── config/               ← 保持不变
├── docs/                 ← 保持不变
└── uploads/              ← 保持不变
```

## Risks / Trade-offs

- **编译风险** → 移动数十个文件后需要确保所有 import 路径正确更新，建议逐模块迁移、每步验证编译通过
- **git 历史丢失** → 纯文件移动 merge 会丢失行级历史，建议使用 `git mv` 或分批 commit 便于 trace
- **PR 变更量大** → 建议按模块分批提交 PR，而非一次性全量改动
- **共享模型依赖** → `User`、`Enroll` 等模型被多模块引用，需要评估拆散还是保留集中
