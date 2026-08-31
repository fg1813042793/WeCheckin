# 通用 OA 流程能力 V1 Implementation Plan

**Goal:** 在不接入绩效流程的前提下，把现有管理员工作流引擎补齐为可由普通用户发起、处理和审计的通用 OA 后端能力。

**Architecture:** 发布定义 JSON 同时承载流程图和表单 Schema；领域引擎负责状态转换与撤回/取消；应用服务负责表单权限、用户数据范围、事务和生命周期事件；GORM 仓储保存表单快照；管理员和普通用户使用独立 HTTP transport。

**Tech Stack:** Go 1.24、GORM、MySQL、Hertz。

> 当前工作区已有进行中的 workflow 修改，本轮只增量完善，不自动提交，也不修改绩效代码。

## Task 1: 表单 Schema 与节点字段权限

**Files:**
- Modify: `backend/internal/workflow/types.go`
- Modify: `backend/internal/workflow/validation.go`
- Modify: `backend/internal/workflow/definition_test.go`

- [x] 先增加表单字段和权限校验测试并确认 RED。
- [x] 增加通用字段、约束、选项和节点字段权限模型。
- [x] 增加启动数据和节点补丁校验函数。

## Task 2: 表单数据与撤回/取消领域能力

**Files:**
- Modify: `backend/internal/modules/workflow/domain/runtime.go`
- Modify: `backend/internal/modules/workflow/domain/engine.go`
- Modify: `backend/internal/modules/workflow/domain/engine_test.go`

- [x] 先增加表单数据保留、发起人撤回和管理员取消测试并确认 RED。
- [x] 增加表单数据状态与撤回/取消状态转换。
- [x] 为新动作追加历史事件并取消活动任务、令牌。

## Task 3: 应用服务与用户数据范围

**Files:**
- Modify: `backend/internal/modules/workflow/application/service.go`
- Modify: `backend/internal/modules/workflow/application/types.go`
- Modify: `backend/internal/modules/workflow/application/service_test.go`

- [x] 先增加表单校验、字段写权限、我的申请/待办/详情和撤回测试并确认 RED。
- [x] 实现用户侧命令与查询，所有 actor 从调用方传入且由 transport 绑定认证上下文。
- [x] 增加生命周期事件发布接口，默认空实现保持独立部署。

## Task 4: 持久化与迁移

**Files:**
- Modify: `backend/internal/model/workflow/runtime.go`
- Modify: `backend/internal/modules/workflow/infrastructure/gorm_store.go`
- Modify: `backend/internal/modules/workflow/infrastructure/mapping.go`
- Create: `backend/migrations/20260831100000_add_workflow_form_data.sql`
- Create: `backend/test/internal/bootstrap/migrations/workflow_form_data_migration_test.go`

- [x] 先增加迁移和映射测试并确认 RED。
- [x] 保存、加载和返回表单快照。
- [x] 增加实例锁定和用户可见范围查询。

## Task 5: 普通用户 API 与管理员取消接口

**Files:**
- Create: `backend/internal/modules/workflow/transport/httpclient/handler.go`
- Create: `backend/internal/modules/workflow/transport/httpclient/handler_test.go`
- Modify: `backend/internal/modules/workflow/transport/httpadmin/handler.go`
- Modify: `backend/internal/routes/v2/client/routes.go`
- Modify: `backend/internal/routes/v2/admin/routes.go`

- [x] 先增加认证身份不可伪造、我的列表强制过滤和撤回测试并确认 RED。
- [x] 注册普通用户定义、发起、申请、待办、详情、审批和撤回接口。
- [x] 注册管理员实例取消接口。

## Task 6: 验证边界

- [x] 运行 workflow 核心、应用、仓储和 transport 测试。
- [x] 运行迁移结构测试和相关路由测试。
- [x] 搜索确认本轮没有修改绩效模块或建立绩效适配器。
