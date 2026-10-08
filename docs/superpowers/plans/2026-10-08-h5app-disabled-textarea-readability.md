# H5App 禁用长文本可读性优化 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 `WorkflowTextarea` 的禁用态已填写内容从浅灰色改为正常正文色，同时保留明确的不可编辑视觉状态。

**Architecture:** 直接在 `WorkflowTextarea.vue` 组件层修改 disabled 样式，因此所有使用该组件的禁用长文本统一生效。H5 原生 textarea 与非 H5 uView textarea 分别增加正文色和 `-webkit-text-fill-color` 覆盖；placeholder、字数统计、背景和禁用光标保持原有层级。

**Tech Stack:** uni-app、Vue 3、TypeScript、uView Pro、SCSS、Node.js 结构契约脚本。

---

## 文件结构

- Modify: `h5app/scripts/check-workflow-textarea-ime.mjs` - 在现有 H5 IME 契约中增加禁用文字可读性和 placeholder 颜色断言。
- Modify: `h5app/src/pages/workflow/components/WorkflowTextarea.vue` - 调整 H5 与非 H5 disabled 文字颜色，不改输入事件和自动高度逻辑。
- Verify only: `h5app/package.json` - 使用当前工作区已有的 `check:workflow-textarea-ime` 脚本，不修改依赖。

### Task 1: 定义禁用长文本颜色契约

**Files:**
- Modify: `h5app/scripts/check-workflow-textarea-ime.mjs`
- Test: `h5app/scripts/check-workflow-textarea-ime.mjs`

- [ ] **Step 1: 增加 H5 原生禁用态断言**

在现有分支断言后加入：

```js
assert.match(
  source,
  /\.workflow-textarea__native--disabled \.workflow-textarea__native-control\s*\{[^}]*color:\s*var\(--u-main-color\);[^}]*-webkit-text-fill-color:\s*var\(--u-main-color\);[^}]*cursor:\s*not-allowed;/,
  'disabled native workflow textarea must keep readable filled text',
)

assert.match(
  source,
  /\.workflow-textarea__native--disabled \.workflow-textarea__native-control::placeholder\s*\{[^}]*color:\s*var\(--u-light-color\);[^}]*-webkit-text-fill-color:\s*var\(--u-light-color\);/,
  'disabled native workflow textarea placeholder must remain muted',
)
```

- [ ] **Step 2: 增加非 H5 uView 禁用态断言**

继续加入：

```js
assert.match(
  source,
  /:deep\(\.workflow-textarea__control\.u-textarea--disabled \.u-textarea__field\)[\s\S]*?color:\s*var\(--u-main-color\) !important;[\s\S]*?-webkit-text-fill-color:\s*var\(--u-main-color\) !important;/,
  'disabled uView workflow textarea must keep readable filled text',
)

assert.match(
  source,
  /:deep\(\.workflow-textarea__control\.u-textarea--disabled \.u-textarea__field::placeholder\)[\s\S]*?color:\s*var\(--u-light-color\) !important;[\s\S]*?-webkit-text-fill-color:\s*var\(--u-light-color\) !important;/,
  'disabled uView workflow textarea placeholder must remain muted',
)
```

- [ ] **Step 3: 运行检查并确认红灯**

Run:

```bash
cd h5app && pnpm check:workflow-textarea-ime
```

Expected: FAIL，首先指出 H5 原生 disabled textarea 尚未使用 `var(--u-main-color)`。

- [ ] **Step 4: 检查测试文件内容**

由于脚本当前为用户未跟踪文件，使用直接检查而不是普通 `git diff`：

```bash
git diff --check -- h5app/scripts/check-workflow-textarea-ime.mjs
git diff --no-index -- /dev/null h5app/scripts/check-workflow-textarea-ime.mjs || true
```

Expected: 无空白错误，原有 IME/原生 textarea 契约保持不变，仅新增颜色断言。

### Task 2: 提高所有禁用 WorkflowTextarea 的正文对比度

**Files:**
- Modify: `h5app/src/pages/workflow/components/WorkflowTextarea.vue:241-275`
- Test: `h5app/scripts/check-workflow-textarea-ime.mjs`

- [ ] **Step 1: 调整 H5 原生 disabled 正文颜色**

将现有禁用规则改为：

```scss
.workflow-textarea__native--disabled .workflow-textarea__native-control {
  color: var(--u-main-color);
  -webkit-text-fill-color: var(--u-main-color);
  cursor: not-allowed;
}

.workflow-textarea__native--disabled .workflow-textarea__native-control::placeholder {
  color: var(--u-light-color);
  -webkit-text-fill-color: var(--u-light-color);
}
```

- [ ] **Step 2: 调整非 H5 uView disabled 正文颜色**

在 uView textarea 深层样式后加入：

```scss
:deep(.workflow-textarea__control.u-textarea--disabled .u-textarea__field) {
  color: var(--u-main-color) !important;
  -webkit-text-fill-color: var(--u-main-color) !important;
}

:deep(.workflow-textarea__control.u-textarea--disabled .u-textarea__field::placeholder) {
  color: var(--u-light-color) !important;
  -webkit-text-fill-color: var(--u-light-color) !important;
}
```

不修改 `.workflow-textarea__native--disabled` 的背景，也不修改 `.workflow-textarea__native-count` 的辅助灰色。

- [ ] **Step 3: 运行契约检查并确认绿灯**

Run:

```bash
cd h5app && pnpm check:workflow-textarea-ime
```

Expected: 输出 `Workflow textarea IME checks passed`。

- [ ] **Step 4: 运行聚焦验证**

Run:

```bash
cd h5app
pnpm exec eslint src/pages/workflow/components/WorkflowTextarea.vue scripts/check-workflow-textarea-ime.mjs
pnpm type-check
pnpm check:ui-style
pnpm build:h5
```

Expected: 四条命令退出码均为 `0`。另运行 `pnpm lint` 和 `pnpm check:workflow-module` 记录既有无关基线失败。

- [ ] **Step 5: 检查最终差异**

Run:

```bash
git diff --check -- h5app/src/pages/workflow/components/WorkflowTextarea.vue h5app/scripts/check-workflow-textarea-ime.mjs
git diff -- h5app/src/pages/workflow/components/WorkflowTextarea.vue
```

Expected: `WorkflowTextarea.vue` 的行为差异仅包含 disabled 文字与 placeholder 颜色；已有 IME 和自动高度改动保持不变。两个目标文件都包含用户现有未提交工作，因此不自动暂存或提交实现。

### Task 3: 浏览器验证禁用长文本可读性

**Files:**
- Verify only: `h5app/src/pages/workflow/components/WorkflowTextarea.vue`

- [ ] **Step 1: 复用当前 H5App 开发服务器**

使用当前运行中的 `http://127.0.0.1:8087/`；若服务已停止，再运行：

```bash
cd h5app && pnpm dev:h5
```

- [ ] **Step 2: 验证标准 PC**

在 `1440x900` 的独立表单详情页打开一个有值的长文本分组，确认 textarea 计算颜色由约 `rgb(192, 196, 204)` 变为正文色，背景仍为浅灰且光标仍为 `not-allowed`。

- [ ] **Step 3: 验证 PC 小屏和手机**

在 `1024x768` 和 `390x844` 检查长文本内容清晰、完整换行且无横向溢出；禁用背景、边框和辅助字数统计保持原样。

- [ ] **Step 4: 检查运行时健康状态**

确认页面身份正确、无 Vite 错误覆盖层、控制台无本变更相关错误，并保存标准 PC 与手机截图到仓库外临时目录。
