# 纯 Go 工作流引擎 V1 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在现有 Go 后端内完成可启动、可审批、可查询、可审计的工作流 V1，并保留现有设计器和 BPMN 导出能力。

**Architecture:** 发布版本 JSON 是运行依据；纯 Go 领域引擎计算状态转换；应用服务使用 GORM 事务持久化实例、令牌、任务、变量和历史；管理后台通过独立实例/待办 API 操作。外部 Flowable HTTP 客户端从运行链路移除。

**Tech Stack:** Go 1.24、GORM、MySQL、Hertz、Vue 3、Element Plus、Vitest/Node 构建检查。

---

> 当前工作区已有用户进行中的修改，本轮不执行自动提交，只做小范围增量修改和验证。

## Task 1: 领域引擎测试与状态模型

**Files:**
- Create: `backend/internal/modules/workflow/domain/engine_test.go`
- Replace: `backend/internal/modules/workflow/domain/runtime.go`
- Create: `backend/internal/modules/workflow/domain/engine.go`
- Delete: `backend/internal/modules/workflow/infrastructure/flowable/client.go`
- Delete: `backend/internal/modules/workflow/infrastructure/flowable/client_test.go`

- [ ] 先写线性审批、排他分支、并行汇聚、顺序审批和会签失败测试。
- [ ] 运行 `GOCACHE=$PWD/.cache/go-build go test ./internal/modules/workflow/domain -count=1`，确认 RED。
- [ ] 实现最小领域模型和引擎，确认测试转为 GREEN。

## Task 2: 运行时模型与迁移

**Files:**
- Create: `backend/internal/model/workflow/runtime.go`
- Modify: `backend/internal/model/aliases.go`
- Modify: `backend/internal/bootstrap/migrate.go`
- Create: `backend/migrations/20260829143000_create_go_workflow_runtime_tables.sql`
- Create: `backend/test/internal/bootstrap/migrations/workflow_runtime_migration_test.go`

- [ ] 先写迁移结构测试并确认 RED。
- [ ] 新增实例、令牌、任务、变量、历史表及索引。
- [ ] 注册模型与迁移，运行迁移测试确认 GREEN。

## Task 3: 应用服务与事务

**Files:**
- Create: `backend/internal/service/admin/workflow/runtime_service.go`
- Create: `backend/internal/service/admin/workflow/runtime_service_test.go`

- [ ] 先写发布版本加载、重复业务键、任务重复完成测试。
- [ ] 实现审批人解析、聚合加载和事务保存。
- [ ] 发布时记录 `go:<definitionID>:<version>` 引擎版本标识。

## Task 4: 管理 API 与权限

**Files:**
- Modify: `backend/internal/handler/admin/workflow/handler.go`
- Modify: `backend/internal/routes/v2/admin/routes.go`
- Modify: `backend/internal/middleware/admin/route_permissions.go`
- Modify: `backend/internal/support/adminrouteperm/catalog.go`
- Modify: `backend/internal/support/adminmenuperm/declarations.go`

- [ ] 新增实例启动、列表、详情、待办列表和任务完成接口。
- [ ] 为运行接口增加独立菜单按钮和 API 权限声明。
- [ ] 运行路由权限完整性测试。

## Task 5: 管理后台运行视图

**Files:**
- Modify: `admin/src/api/index.ts`
- Modify: `admin/src/router/adminRoutes.ts`
- Create: `admin/src/views/workflow/instances/index.vue`

- [ ] 增加实例列表、状态筛选和详情入口。
- [ ] 增加待办操作区，完成审批时要求二次确认和审批意见。
- [ ] 保持设计器与运行页面分层，不在设计器中写运行逻辑。

## Task 6: 集成验证

- [ ] 运行 `GOCACHE=$PWD/.cache/go-build go test ./internal/modules/workflow/... ./internal/workflow/... -count=1`。
- [ ] 运行 `GOCACHE=$PWD/.cache/go-build go test ./internal/service/admin/workflow ./internal/middleware/admin -count=1`。
- [ ] 运行迁移结构测试。
- [ ] 运行管理后台 TypeScript 检查和构建。
- [ ] 记录未接入绩效模块等 V1 明确边界。

