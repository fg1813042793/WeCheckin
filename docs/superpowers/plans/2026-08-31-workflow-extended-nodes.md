# 通用工作流扩展节点 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为通用 Go 工作流和管理端设计器增加办理、抄送、自动动作、定时等待节点，以及按指定用户分别发起流程的能力。

**Architecture:** 发布版本 JSON 继续作为 Go 引擎唯一运行依据。人工办理复用任务聚合，抄送和自动动作同步执行，定时节点把到期时间保存在实例变量中并由幂等恢复命令推进；按用户发起由管理端编排多个独立启动请求。

**Tech Stack:** Go 1.24、Hertz、GORM、Vue 3、TypeScript、Element Plus、Vite。

---

### Task 1: 扩展流程定义与校验

**Files:**
- Modify: `backend/internal/workflow/types.go`
- Modify: `backend/internal/workflow/validation.go`
- Test: `backend/internal/workflow/definition_test.go`

- [ ] 先写定义校验测试，构造包含 `handle`、`cc`、`automation`、`timer` 的线性流程并断言校验通过。
- [ ] 添加缺少处理人、空自动变量、非法等待秒数的失败用例并确认测试因未知节点类型失败。
- [ ] 增加节点常量、`AutomationConfig`、`TimerConfig` 和对应校验逻辑。
- [ ] 运行 `GOCACHE=$PWD/.cache/go-build go test ./internal/workflow -run 'TestValidateDefinition' -count=1` 并确认通过。

### Task 2: 扩展 BPMN 导出

**Files:**
- Modify: `backend/internal/workflow/compiler.go`
- Test: `backend/internal/workflow/definition_test.go`

- [ ] 先写 BPMN 失败测试，断言输出包含办理 `userTask`、抄送/自动动作 `serviceTask` 和定时 `intermediateCatchEvent`。
- [ ] 为四类节点补充编码函数；定时持续时间输出 ISO-8601 `PT<n>S`。
- [ ] 运行 `GOCACHE=$PWD/.cache/go-build go test ./internal/workflow -run 'TestCompileBPMN' -count=1` 并确认通过。

### Task 3: 扩展领域运行时

**Files:**
- Modify: `backend/internal/modules/workflow/domain/runtime.go`
- Modify: `backend/internal/modules/workflow/domain/engine.go`
- Test: `backend/internal/modules/workflow/domain/engine_test.go`

- [ ] 先写办理节点测试，断言创建单人任务、拒绝 `approve/reject`、接受 `submit` 并完成实例。
- [ ] 写抄送和自动动作测试，断言不创建任务、写历史、合并变量并继续。
- [ ] 写固定时钟的定时节点测试，断言启动后等待、到期前不推进、到期后推进、重复恢复为 0。
- [ ] 增加 `submit` 动作、`completed` 任务状态、新历史类型、时钟注入和 `ResumeTimers`。
- [ ] 运行 `GOCACHE=$PWD/.cache/go-build go test ./internal/modules/workflow/domain -count=1` 并确认通过。

### Task 4: 接入定时恢复应用服务与管理接口

**Files:**
- Modify: `backend/internal/modules/workflow/application/service.go`
- Modify: `backend/internal/modules/workflow/application/service_test.go`
- Modify: `backend/internal/modules/workflow/transport/httpadmin/handler.go`
- Modify: `backend/internal/modules/workflow/transport/httpadmin/handler_test.go`
- Modify: `backend/internal/routes/v2/admin/routes.go`
- Modify: `backend/internal/middleware/admin/route_permissions.go`

- [ ] 先写应用服务失败测试，断言按实例加锁、恢复到期令牌并保存状态。
- [ ] 先写 HTTP 失败测试，断言实例 ID 和认证管理员身份传入服务。
- [ ] 增加 `POST /api/v2/admin/workflow-instances/:id/resume`，复用 `workflow:instance:start` 权限。
- [ ] 运行工作流应用、传输层和权限映射定向测试并确认通过。

### Task 5: 扩展管理端设计器

**Files:**
- Modify: `admin/src/views/workflow/types.ts`
- Modify: `admin/src/views/workflow/designer/graph.ts`
- Modify: `admin/src/views/workflow/designer/components/FlowInsertButton.vue`
- Modify: `admin/src/views/workflow/designer/components/WorkflowCanvas.vue`
- Modify: `admin/src/views/workflow/designer/components/WorkflowSequence.vue`
- Modify: `admin/src/views/workflow/designer/components/WorkflowNodeCard.vue`
- Modify: `admin/src/views/workflow/designer/components/NodeInspector.vue`
- Modify: `admin/src/views/workflow/designer/components/WorkflowFieldPermissions.vue`
- Modify: `admin/src/views/workflow/designer/index.vue`
- Modify: `admin/scripts/check-workflow-tree.mjs`
- Modify: `admin/scripts/check-workflow-form-designer.mjs`

- [ ] 先扩充静态检查，要求四类节点、配置字段和办理字段权限存在，并确认检查失败。
- [ ] 扩展节点类型、插入逻辑、菜单、卡片摘要和检查器。
- [ ] 将办理节点加入字段权限矩阵，并允许配置明细表新增/删除行。
- [ ] 运行 `npm run check:workflow-tree` 与 `npm run check:workflow-form-designer` 并确认通过。

### Task 6: 增加按用户独立发起和办理任务交互

**Files:**
- Modify: `admin/src/api/index.ts`
- Modify: `admin/src/views/workflow/types.ts`
- Modify: `admin/src/views/workflow/instances/index.vue`
- Modify: `admin/src/views/workflow/tasks/index.vue`
- Modify: `admin/scripts/check-workflow-runtime-pages.mjs`

- [ ] 先扩充运行页静态检查，要求目标用户选择、`targetUserId`、独立业务标识、`submit` 和定时恢复 API，并确认检查失败。
- [ ] 发起对话框增加普通/按用户模式；按用户模式逐个创建实例并汇总结果。
- [ ] 任务处理页根据实例节点类型为办理节点显示“提交”，审批节点仍显示“通过/驳回”。
- [ ] 实例详情增加“推进到期节点”动作和反馈。
- [ ] 运行 `npm run check:workflow-runtime-pages` 与管理端类型构建并确认通过。

### Task 7: 全量验证与差异复核

**Files:**
- Verify only: all files above

- [ ] 运行 `GOCACHE=$PWD/.cache/go-build go test ./internal/workflow ./internal/modules/workflow/... ./internal/middleware/admin -count=1`。
- [ ] 运行 `npm run check:workflow-tree && npm run check:workflow-form-designer && npm run check:workflow-runtime-pages`。
- [ ] 运行 `npm run build`。
- [ ] 使用 `git diff --check` 检查空白错误，使用 `git diff -- <相关文件>` 确认没有覆盖无关改动。
