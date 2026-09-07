# H5App 流程筛选布局 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让流程中心四个记录页签和汇总页的查询按钮在 PC 端有空间时紧跟筛选条件、空间不足时自然换行，并收窄记录页签的流程分类选择框。

**Architecture:** 保留两个现有筛选组件及其查询逻辑，只把桌面端固定网格改为按 DOM 顺序排列的可换行 Flex。移动端继续使用单列 Grid 和两列操作按钮，避免影响触控布局。

**Tech Stack:** uni-app、Vue 3、TypeScript、SCSS、Node.js 结构契约脚本、uView Pro。

---

### Task 1: 为响应式筛选布局增加失败契约

**Files:**
- Modify: `h5app/scripts/check-ui-style-guidelines.mjs`
- Test: `h5app/scripts/check-ui-style-guidelines.mjs`

- [ ] **Step 1: 读取流程中心组件并声明布局契约**

在脚本中读取 `WorkflowCenter.vue`，增加断言，要求列表和汇总筛选容器使用 `flex-wrap: wrap`、操作区使用 `flex: 0 0 auto`，并要求列表分类字段存在紧凑 class：

```js
const workflowCenter = read('../src/pages/workflow/components/WorkflowCenter.vue')

assert.match(workflowCenter, /workflow-center__filter-field--category/)
assert.match(workflowCenter, /\.workflow-center__record-filters\s*\{[^}]*display:\s*flex;[^}]*flex-wrap:\s*wrap;/)
assert.match(workflowCenter, /\.workflow-center__filter-field--category\s*\{[^}]*flex:\s*0 1 160px;[^}]*max-width:\s*160px;/)
assert.match(workflowCenter, /\.workflow-center__filter-actions\s*\{[^}]*flex:\s*0 0 auto;/)
assert.match(workflowSummary, /\.workflow-summary__filters\s*\{[^}]*display:\s*flex;[^}]*flex-wrap:\s*wrap;/)
assert.match(workflowSummary, /\.workflow-summary__filter-actions\s*\{[^}]*flex:\s*0 0 auto;/)
```

- [ ] **Step 2: 运行契约并确认按预期失败**

Run: `cd h5app && pnpm check:ui-style`

Expected: FAIL，首个失败说明 `WorkflowCenter.vue` 尚不存在分类紧凑 class 或筛选容器仍为 Grid。

### Task 2: 调整四个记录页签的共享筛选布局

**Files:**
- Modify: `h5app/src/pages/workflow/components/WorkflowCenter.vue`
- Test: `h5app/scripts/check-ui-style-guidelines.mjs`

- [ ] **Step 1: 标记流程分类字段**

给流程分类外层增加专用 class：

```vue
<view class="workflow-center__filter-field workflow-center__filter-field--category">
```

- [ ] **Step 2: 将桌面筛选网格改为可换行 Flex**

使用以下宽度职责：普通字段约 `170px`，分类约 `160px`，日期范围约 `200px`，字段不主动吃满行尾空间，按钮不压缩并跟在最后一个条件后面：

```scss
.workflow-center__record-filters {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  gap: 12px;
}

.workflow-center__filter-field {
  min-width: 150px;
  max-width: 220px;
  flex: 0 1 170px;
}

.workflow-center__filter-field--category {
  min-width: 150px;
  flex: 0 1 160px;
  max-width: 160px;
}

.workflow-center__filter-field--date {
  min-width: 200px;
  max-width: 240px;
  flex: 0 1 200px;
}

.workflow-center__filter-actions {
  flex: 0 0 auto;
}
```

- [ ] **Step 3: 保留紧凑 PC 和手机布局**

`max-width: 900px` 保留两列 Grid，让最后一个条件和操作区可以共享一行；`max-width: 768px` 把字段的最小/最大宽度复位并保持单列：

```scss
@media screen and (max-width: 768px) {
  .workflow-center__filter-field,
  .workflow-center__filter-field--category,
  .workflow-center__filter-field--date {
    width: 100%;
    min-width: 0;
    max-width: none;
  }
}
```

### Task 3: 调整汇总页筛选布局

**Files:**
- Modify: `h5app/src/pages/workflow/components/WorkflowSummarySection.vue`
- Test: `h5app/scripts/check-ui-style-guidelines.mjs`

- [ ] **Step 1: 移除桌面固定行列定位**

把 `.workflow-summary__filters` 改为可换行 Flex，并删除各筛选项的 `grid-column`、`grid-row` 和 `@media (max-width: 1200px)` 固定定位。

- [ ] **Step 2: 设置字段宽度职责**

```scss
.workflow-summary__filter {
  min-width: 160px;
  max-width: 220px;
  flex: 1 1 180px;
}

.workflow-summary__filter--status,
.workflow-summary__filter--version {
  min-width: 120px;
  max-width: 140px;
  flex: 0 1 140px;
}

.workflow-summary__filter--range {
  min-width: 280px;
  max-width: 340px;
  flex: 1 1 280px;
}

.workflow-summary__filter-actions {
  width: fit-content;
  flex: 0 0 auto;
}
```

- [ ] **Step 3: 保持手机单列和按钮两列等宽**

在 `max-width: 768px` 中继续使用单列 Grid，并为所有字段复位 `min-width`、`max-width` 与 `width`，操作按钮保持两列。

- [ ] **Step 4: 运行契约并确认通过**

Run: `cd h5app && pnpm check:ui-style`

Expected: `H5App UI style guideline checks passed`。

### Task 4: 静态、构建与浏览器验证

**Files:**
- Verify: `h5app/src/pages/workflow/components/WorkflowCenter.vue`
- Verify: `h5app/src/pages/workflow/components/WorkflowSummarySection.vue`

- [ ] **Step 1: 运行模块和静态检查**

Run: `cd h5app && pnpm check:workflow-module && pnpm lint && pnpm type-check`

Expected: 所有命令退出码为 `0`。

- [ ] **Step 2: 运行 H5 构建**

Run: `cd h5app && pnpm build:h5`

Expected: 构建完成且退出码为 `0`。

- [ ] **Step 3: 用应用内 Browser 验证视口**

在 `1023x768`、`1366x768`、`390x844` 验证：

```text
流程中心 -> 我的申请/我的待办/已处理/抄送我的 -> 筛选条件 -> 按钮同行或自然换行
流程中心 -> 汇总 -> 筛选条件 -> 按钮紧跟最后一个条件
```

检查页面非空、无框架错误层、控制台无相关错误、页面 `scrollWidth <= clientWidth`，并点击查询或重置确认交互可用。
