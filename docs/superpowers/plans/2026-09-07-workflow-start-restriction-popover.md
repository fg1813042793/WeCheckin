# Workflow Start Restriction Popover Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在 H5App 发起审批卡片上增加仅 PC 生效的流程限制悬浮提示，并阻止当前不可发起流程进入表单。

**Architecture:** 新增无 UI 依赖的 `workflow-start-restrictions.ts`，将后端返回的开放规则、可用状态和个人次数状态格式化为统一展示模型。`WorkflowCenter.vue` 只消费展示模型并渲染 CSS 悬浮层，手机断点下不显示；后端继续负责最终资格校验。

**Tech Stack:** Vue 3、TypeScript、uni-app H5、uView Pro、SCSS、现有 Node 契约检查。

---

### Task 1: 流程限制展示模型

**Files:**
- Create: `h5app/src/pages/workflow/workflow-start-restrictions.ts`
- Modify: `h5app/scripts/check-workflow-module.mjs`

- [ ] **Step 1: 写入失败的格式化测试**

在 `check-workflow-module.mjs` 中登记新文件，并通过 `loadTypeScriptModule` 校验：长期有效、固定时间、每周开放、每月开放、跨月窗口、周期次数、次数耗尽和重置时间。

- [ ] **Step 2: 确认测试因实现缺失而失败**

Run: `cd h5app && pnpm check:workflow-module`

Expected: FAIL，提示缺少 `workflow-start-restrictions.ts` 或其导出函数。

- [ ] **Step 3: 实现最小格式化模块**

导出：

```ts
export interface WorkflowStartRestrictionSummary {
  allowed: boolean
  statusLabel: string
  statusTone: 'success' | 'warning' | 'danger'
  reason: string
  items: Array<{ key: string, label: string, value: string }>
}

export function workflowStartRestrictionSummary(
  definition: WorkflowPublishedDefinition,
): WorkflowStartRestrictionSummary
```

状态优先级与后端一致：先判断 `availabilityStatus`，再判断 `startLimitStatus.allowed`。剩余次数只读取后端状态，不在前端重新统计。

- [ ] **Step 4: 确认格式化测试通过**

Run: `cd h5app && pnpm check:workflow-module`

Expected: PASS，输出 `workflow module structure checks passed`。

### Task 2: PC 悬浮层与不可发起拦截

**Files:**
- Modify: `h5app/src/pages/workflow/components/WorkflowCenter.vue`
- Modify: `h5app/scripts/check-workflow-module.mjs`

- [ ] **Step 1: 写入失败的页面契约检查**

要求页面包含 `workflowStartRestrictionSummary`、`workflow-definition__restriction`、键盘聚焦入口，以及仅允许悬停指针和 `min-width: 769px` 显示浮层的媒体查询。

- [ ] **Step 2: 确认页面契约检查失败**

Run: `cd h5app && pnpm check:workflow-module`

Expected: FAIL，提示 `WorkflowCenter.vue` 缺少限制浮层结构。

- [ ] **Step 3: 实现卡片交互**

- 卡片增加 `role="button"`、`tabindex="0"` 和 Enter/Space 键处理。
- 点击前读取统一展示模型；不可发起时显示 `reason`，可发起时保持现有新开流程 Tab 行为。
- PC 悬停或聚焦卡片时显示状态、开放时间、次数、发起范围和可选重置时间。
- 浮层使用绝对定位和稳定宽度，不改变网格尺寸；各列末端卡片调整对齐，避免超出内容区。
- 默认隐藏浮层，仅在 `(hover: hover) and (pointer: fine) and (min-width: 769px)` 下开放显示。

- [ ] **Step 4: 确认页面契约检查通过**

Run: `cd h5app && pnpm check:workflow-module`

Expected: PASS。

### Task 3: 全量验证

**Files:**
- Verify only

- [ ] **Step 1: 检查代码风格和类型**

Run: `cd h5app && pnpm lint && pnpm type-check`

Expected: PASS。

- [ ] **Step 2: 检查 UI 契约与 H5 构建**

Run: `cd h5app && pnpm check:ui-style && pnpm build:h5`

Expected: PASS。

- [ ] **Step 3: 静态核对移动端隔离**

确认默认样式隐藏浮层，只有精确指针且宽度至少 `769px` 时显示；手机端没有新增点击分支、弹窗或信息图标。

- [ ] **Step 4: 检查改动范围**

Run: `git diff --check -- h5app/src/pages/workflow h5app/scripts/check-workflow-module.mjs docs/superpowers`

Expected: 无输出；不覆盖工作区已有流程标题、搜索和业务期间相关修改。
