# 流程用户树目录选择实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让流程用户树在多人模式下支持通过部门目录批量选择用户。

**Architecture:** 行为统一实现在共享 `WorkflowUserTreePicker` 中，由现有 `multiple` 属性决定部门节点是否可选。部门勾选在前端展开为去重后的用户 ID，不修改 API 和后端执行契约。

**Tech Stack:** Vue 3, TypeScript strict, Element Plus `el-tree`, Node.js 结构回归检查。

---

### Task 1: 回归检查

**Files:**
- Modify: `admin/scripts/check-scheduled-task.mjs`

- [ ] 在结构检查中断言共享用户树包含多人模式目录启用和目录子用户收集逻辑。
- [ ] 运行 `npm run check:scheduled-task`，确认现有实现上失败。

### Task 2: 共享选择器

**Files:**
- Modify: `admin/src/views/workflow/components/WorkflowUserTreePicker.vue`

- [ ] 构建部门树时，仅在非多人模式禁用目录节点。
- [ ] 收集所选目录子树中的用户 ID，勾选时批量加入，取消时批量移除。
- [ ] 保留单个用户的现有选择和去重行为。

### Task 3: 验证

**Files:**
- Verify: `admin/`

- [ ] 运行 `npm run check:scheduled-task`。
- [ ] 运行 `npm run check:all`。
- [ ] 在浏览器检查多人模式的部门勾选、取消和单人模式的目录禁用。
