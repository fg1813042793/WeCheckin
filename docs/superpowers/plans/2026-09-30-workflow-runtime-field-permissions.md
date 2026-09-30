# Workflow Runtime Field Permissions Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans. Steps use checkbox syntax.

**Goal:** Admin 保存字段权限后立即影响所有匹配的运行中和未来流程实例，不增加发布版本号。

**Architecture:** 使用 `workflow_definitions` 独立 JSON + revision 保存实时字段权限。加载已发布定义时按 node/field 稳定 ID 安全覆盖；版本比较前对旧快照应用当前运行时权限。

**Tech Stack:** Go、GORM、MySQL migration、Vue 3、TypeScript。

---

### Task 1: Runtime field-permission core
- [ ] 先写 `runtime_field_permissions_test.go`，覆盖编解码、过滤旧节点字段和克隆。
- [ ] 实现 `runtime_field_permissions.go` 并运行 `go test ./internal/workflowcore`。

### Task 2: Persistence and migration
- [ ] 为 `WorkflowDefinition` 增加 JSON/revision 字段。
- [ ] 新增版本化 SQL 和迁移契约测试。
- [ ] 定义加载时应用运行时字段权限。

### Task 3: Admin save and version semantics
- [ ] Create/Update/GetDetail/Publish/Rollback 接入运行时字段权限。
- [ ] 仅字段权限变化不产生发布差异，保存时 revision 自增。
- [ ] 运行 Admin workflow service tests。

### Task 4: Admin contract and verification
- [ ] 增加 `fieldPermissionsRevision` 类型。
- [ ] 字段权限页面增加立即生效提示。
- [ ] 运行 Backend 聚焦测试、Admin 工作流检查与构建。
