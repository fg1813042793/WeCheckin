# 问卷设计器第一批体验优化实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 改善问卷设计器图标按钮可访问性、窄屏状态可见性、属性面板空间利用和移动端提示。

**Architecture:** 保持现有问卷状态、保存、历史和数据结构不变，在页面壳、属性面板标题及响应式样式中增加小范围交互。使用现有 `check-survey-designer-ux.mjs` 锁定行为契约，先失败再实现。

**Tech Stack:** Vue 3、TypeScript、Element Plus、CSS Media Queries、Node.js `assert` 静态契约检查、Vite。

---

### Task 1: 写入失败的 UX 契约

**Files:**
- Modify: `admin/scripts/check-survey-designer-ux.mjs`
- Test: `admin/scripts/check-survey-designer-ux.mjs`

- [ ] **Step 1: 增加页面壳契约**

要求 `SurveyDesigner.vue` 包含：

```js
for (const pattern of [
  'aria-label="编辑问卷"',
  'aria-label="撤销"',
  'aria-label="问卷设置面板"',
  'class="toolbar-compact-status"',
  'class="survey-mobile-notice"',
  "(selected || showSurveySettings)",
]) {
  assert.ok(designer.includes(pattern), `survey designer first-pass UX missing contract: ${pattern}`)
}
```

- [ ] **Step 2: 增加属性面板和响应式契约**

```js
for (const pattern of ['aria-label="关闭属性面板"', "emit('close')"]) {
  assert.ok(panelTitle.includes(pattern), `survey property panel close missing contract: ${pattern}`)
}
assert.ok(questionProperties.includes('function closePropertiesPanel('), 'survey property panel must support closing')
for (const pattern of ['.toolbar-compact-status', 'flex-wrap:wrap']) {
  assert.ok(designerStyle.includes(pattern), `survey narrow toolbar missing contract: ${pattern}`)
}
for (const pattern of ['.survey-mobile-notice', '.survey-main-content > :not(.survey-mobile-notice)']) {
  assert.ok(responsive.includes(pattern), `survey mobile notice missing contract: ${pattern}`)
}
```

- [ ] **Step 3: 运行检查并确认失败**

```bash
cd admin && npm run check:survey-designer-ux
```

Expected: FAIL，首个错误包含 `survey designer first-pass UX missing contract`。

### Task 2: 实现页面壳和属性面板优化

**Files:**
- Modify: `admin/src/views/survey/SurveyDesigner.vue`
- Modify: `admin/src/views/survey/components/SurveyDesignerPanelTitle.vue`
- Modify: `admin/src/views/survey/components/SurveyDesignerQuestionProperties.vue`
- Modify: `admin/src/views/survey/survey-designer.css`
- Modify: `admin/src/views/survey/survey-designer-responsive.css`
- Test: `admin/scripts/check-survey-designer-ux.mjs`

- [ ] **Step 1: 补齐按钮语义**

导航和工具栏图标按钮增加 `type="button"`、`aria-label`，模式按钮增加 `:aria-pressed`。问卷设置入口使用 Element Plus `Setting` 图标：

```vue
<el-tooltip content="问卷设置" placement="bottom">
  <el-button
    text
    size="small"
    class="toolbar-btn"
    :class="{ active: showSurveySettings }"
    aria-label="问卷设置面板"
    :aria-pressed="showSurveySettings"
    @click="openSurveySettings"
  >
    <el-icon><Setting /></el-icon>
  </el-button>
</el-tooltip>
```

- [ ] **Step 2: 增加紧凑状态和移动端提示**

```vue
<div class="toolbar-compact-status">
  <span class="toolbar-compact-title">{{ form.title || '未命名问卷' }}</span>
  <span class="toolbar-save-status" :class="saveStatusClass">{{ saveStatusLabel }}</span>
</div>

<section class="survey-mobile-notice" aria-label="移动端编辑提示">
  <strong>建议在桌面端编辑问卷</strong>
  <span>问卷设计器需要更宽的操作空间，请在宽度不小于 768px 的设备上使用。</span>
  <el-button type="primary" @click="goBack">返回问卷列表</el-button>
</section>
```

- [ ] **Step 3: 属性面板支持关闭**

`SurveyDesignerPanelTitle` 增加 `closable` prop 和 `close` emit，显示带 `Close` 图标的按钮。`SurveyDesignerQuestionProperties` 中所有标题传入 `closable` 和 `@close="closePropertiesPanel"`，关闭函数为：

```ts
function closePropertiesPanel() {
  selected.value = null
  selectedOptIdx.value = -1
  showSurveySettings.value = false
}
```

父页面仅在 `selected || showSurveySettings` 时显示属性面板。

- [ ] **Step 4: 增加响应式样式**

桌面隐藏 `.toolbar-compact-status`；`<=1080px` 时工具栏换行、完整状态区隐藏、紧凑状态占满第二行。`<768px` 时隐藏工作区直接子节点，仅显示 `.survey-mobile-notice`。

- [ ] **Step 5: 运行聚焦检查**

```bash
cd admin && npm run check:survey-designer-ux
```

Expected: PASS，输出 `survey designer UX checks passed`。

- [ ] **Step 6: 提交实现**

```bash
git add admin/scripts/check-survey-designer-ux.mjs admin/src/views/survey/SurveyDesigner.vue admin/src/views/survey/survey-designer.css admin/src/views/survey/survey-designer-responsive.css admin/src/views/survey/components/SurveyDesignerPanelTitle.vue admin/src/views/survey/components/SurveyDesignerQuestionProperties.vue
git commit -m "feat: 优化问卷设计器基础交互"
```

### Task 3: 构建和浏览器验证

**Files:**
- Verify: `admin/src/views/survey/SurveyDesigner.vue`
- Verify: `admin/src/views/survey/components/SurveyDesignerQuestionProperties.vue`
- Verify: `admin/src/views/survey/survey-designer.css`
- Verify: `admin/src/views/survey/survey-designer-responsive.css`

- [ ] **Step 1: 静态验证**

```bash
cd admin && npm run build
cd admin && npm run check:bundle
```

Expected: TypeScript、Vite 构建和 bundle 预算通过。

- [ ] **Step 2: 浏览器验证**

使用本地 Admin 登录后检查：

1. 标准桌面：未选择题目时属性面板隐藏；通过工具栏打开问卷设置；关闭按钮重新释放空间。
2. 1024px 和 768px：操作按钮保持首行，问卷标题和保存状态显示在第二行，无重叠和水平溢出。
3. 767px：只显示桌面端编辑提示和返回列表按钮。
4. DOM/AX：主要图标按钮具有可访问名称，模式按钮暴露 pressed 状态。
5. 控制台无本次改动引入的错误或警告。

- [ ] **Step 3: 检查差异和工作区**

```bash
git diff --check HEAD~1..HEAD -- admin/scripts/check-survey-designer-ux.mjs admin/src/views/survey
git status --short
```

Expected: 提交只包含本计划列出的文件；用户原有工作区修改保持不变。

