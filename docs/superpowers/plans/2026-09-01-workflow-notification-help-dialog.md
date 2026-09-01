# Workflow Notification Help Dialog Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在流程节点通知配置区增加消息配置说明弹窗。

**Architecture:** 仅修改节点检查器的展示层，不改变流程定义和后端通知契约。使用现有 Element Plus 图标、Tooltip、Dialog 和 Table 组件，静态脚本锁定关键内容。

**Tech Stack:** Vue 3、TypeScript、Element Plus、Node.js 静态回归脚本。

---

### Task 1: 通知说明结构回归

**Files:**
- Modify: `admin/scripts/check-workflow-tree.mjs`
- Test: `admin/scripts/check-workflow-tree.mjs`

- [ ] **Step 1: 写入失败检查**

要求 `NodeInspector.vue` 包含 `notificationHelpVisible`、`QuestionFilled`、`append-to-body`、五个占位符和发送长度限制。

- [ ] **Step 2: 验证检查失败**

Run: `cd admin && npm run check:workflow-tree`

Expected: FAIL，提示缺少通知配置说明弹窗。

### Task 2: 通知说明弹窗

**Files:**
- Modify: `admin/src/views/workflow/designer/components/NodeInspector.vue`
- Test: `admin/scripts/check-workflow-tree.mjs`

- [ ] **Step 1: 实现最小弹窗**

在通知区标题旁增加问号图标按钮；弹窗展示配置步骤、占位符表格、模板示例、渲染结果和长度限制，并使用 `append-to-body`。

- [ ] **Step 2: 验证静态检查和构建**

Run: `cd admin && npm run check:workflow-tree && npm run build`

Expected: 两个命令均成功。

- [ ] **Step 3: 浏览器验证**

在管理端流程设计器打开审批或办理节点，点击通知说明图标，确认弹窗显示且不被节点抽屉遮挡。
