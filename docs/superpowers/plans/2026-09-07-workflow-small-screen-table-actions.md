# Workflow Small-Screen Table And Actions Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让小屏 PC 完整访问流程表格列，并收紧详情底部操作按钮宽度。

**Architecture:** 在共享流程记录表中提供桌面横向滚动和固定操作列，发起审批历史表沿用相同的固定操作列策略；详情按钮仅调整稳定网格宽度。移动端卡片布局保持不变。

**Tech Stack:** uni-app、Vue 3、TypeScript、SCSS、Node.js 契约脚本

---

### Task 1: 增加失败的样式契约

**Files:**
- Modify: `h5app/scripts/check-workflow-module.mjs`

- [ ] **Step 1: 写入滚动、固定操作列和 96px 按钮断言**
- [ ] **Step 2: 运行 `node scripts/check-workflow-module.mjs`，确认因目标样式缺失而失败**

### Task 2: 实现小屏 PC 表格和按钮样式

**Files:**
- Modify: `h5app/src/pages/workflow/components/WorkflowRecordTable.vue`
- Modify: `h5app/src/pages/workflow/components/WorkflowStartPage.vue`
- Modify: `h5app/src/pages/workflow/components/WorkflowDetailPanel.vue`

- [ ] **Step 1: 让共享记录表桌面端支持横向滚动并固定操作列**
- [ ] **Step 2: 固定发起审批历史记录的操作列**
- [ ] **Step 3: 将详情底部操作按钮宽度调整为 96px**
- [ ] **Step 4: 复跑契约、类型检查和针对性 ESLint**

### Task 3: 构建和浏览器验证

**Files:**
- Test: `h5app/scripts/check-workflow-module.mjs`

- [ ] **Step 1: 运行 `npm run build:h5`，确认生产构建通过**
- [ ] **Step 2: 在 1024、1200、1366 宽度下检查横向滚动、固定操作列和按钮尺寸**
- [ ] **Step 3: 只提交本次文档、契约和三个流程组件文件**
