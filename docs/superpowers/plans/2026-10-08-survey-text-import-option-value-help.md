# 问卷文本导入选项值说明实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在 Admin 问卷设计器文本导入弹窗中明确展示 `A. [选项值] 选项文本` 语法，并用回归契约防止说明再次丢失。

**Architecture:** 保持 `parseSurveyText` 和导入数据结构不变，只修改文本导入弹窗的帮助文案、占位示例和必要的换行样式。复用现有 `check-survey-designer-ux.mjs` 作为静态 UI 契约门禁，先让新增断言失败，再用最小 UI 修改使其通过。

**Tech Stack:** Vue 3、TypeScript、Element Plus、Node.js `assert` 静态契约检查、Vite。

---

## 文件结构

- 修改 `admin/scripts/check-survey-designer-ux.mjs`：锁定文本导入弹窗必须包含显式选项值语法和可复制示例。
- 修改 `admin/src/views/survey/components/SurveyDesignerTextImportDialog.vue`：补充顶部提示、占位示例和提示文本换行样式。
- 不修改 `admin/src/views/survey/survey-designer-text-transfer.ts`：当前解析器已支持方括号显式值语法。

### Task 1: 以契约驱动补充选项值帮助文案

**Files:**
- Modify: `admin/scripts/check-survey-designer-ux.mjs`
- Modify: `admin/src/views/survey/components/SurveyDesignerTextImportDialog.vue`
- Test: `admin/scripts/check-survey-designer-ux.mjs`

- [ ] **Step 1: 写入失败契约**

在读取 `textImportDialog` 后的现有组件契约附近加入以下断言：

```js
for (const pattern of [
  '指定值：A. [选项值] 选项文本',
  'A. [red] 红色',
]) {
  assert.ok(textImportDialog.includes(pattern), `survey text import help missing contract: ${pattern}`)
}
```

- [ ] **Step 2: 运行聚焦检查并确认正确失败**

Run:

```bash
cd admin && npm run check:survey-designer-ux
```

Expected: FAIL，错误包含 `survey text import help missing contract: 指定值：A. [选项值] 选项文本`，证明检查针对缺失说明生效。

- [ ] **Step 3: 更新顶部格式提示**

将现有格式提示替换为同时说明默认格式和显式值格式：

```vue
<el-tag type="info" class="format-hint">
  选项：A. 选项文本；指定值：A. [选项值] 选项文本
</el-tag>
```

- [ ] **Step 4: 更新占位示例**

将颜色题示例改为同时展示显式值和默认字母值：

```vue
placeholder="粘贴题目文本&#10;&#10;格式示例：&#10;1. 您最喜欢的颜色？&#10;A. [red] 红色&#10;B. [blue] 蓝色&#10;C. 绿色&#10;&#10;2. 您的姓名&#10;&#10;3. [多选题] 您的爱好&#10;A. 阅读&#10;B. 运动&#10;C. 音乐"
```

其中 `red`、`blue` 会成为选项值，`C. 绿色` 继续使用 `C` 作为默认选项值。

- [ ] **Step 5: 保证帮助文本在弹窗内完整显示**

将工具栏和格式提示样式调整为允许窄窗口换行：

```css
.import-toolbar { display:flex; flex-wrap:wrap; align-items:center; gap:8px; margin-bottom:8px; }
.format-hint { height:auto; font-size:11px; line-height:20px; white-space:normal; }
```

- [ ] **Step 6: 运行聚焦检查并确认通过**

Run:

```bash
cd admin && npm run check:survey-designer-ux
```

Expected: PASS，输出 `survey designer UX checks passed`。

- [ ] **Step 7: 提交契约和实现**

```bash
git add admin/scripts/check-survey-designer-ux.mjs admin/src/views/survey/components/SurveyDesignerTextImportDialog.vue
git commit -m "fix: 补充问卷导入选项值说明"
```

### Task 2: 全量验证和浏览器冒烟

**Files:**
- Verify: `admin/scripts/check-survey-designer-ux.mjs`
- Verify: `admin/src/views/survey/components/SurveyDesignerTextImportDialog.vue`

- [ ] **Step 1: 运行 Admin 全量门禁**

Run:

```bash
cd admin && npm run check:all
```

Expected: 所有 Admin 静态检查、TypeScript 构建、Vite 构建和 bundle 检查通过。若出现与本次两个文件无关的既有失败，单独记录失败命令与证据，不修改无关模块。

- [ ] **Step 2: 启动 Admin 开发服务器**

Run:

```bash
cd admin && npm run dev -- --host 127.0.0.1 --port 18083
```

Expected: Vite 输出本地地址 `http://127.0.0.1:18083/`。若端口占用，改用下一个空闲端口并记录实际地址。

- [ ] **Step 3: 浏览器验证文本导入弹窗**

在浏览器中进入 Admin 问卷设计器并打开“文本导入”，检查：

1. 顶部完整显示 `选项：A. 选项文本；指定值：A. [选项值] 选项文本`。
2. 输入框占位示例完整显示 `A. [red] 红色` 和 `B. [blue] 蓝色`。
3. 在标准桌面宽度和窄窗口下，提示可换行，不遮挡粘贴/上传切换控件及输入区域。
4. 粘贴示例后预览题数正常，导入按钮可用；导入后颜色选项值分别为 `red`、`blue`、`C`。

- [ ] **Step 4: 检查最终差异**

Run:

```bash
git diff 314cd46..HEAD -- admin/scripts/check-survey-designer-ux.mjs admin/src/views/survey/components/SurveyDesignerTextImportDialog.vue
git status --short
```

Expected: 功能差异只包含契约断言、帮助文案、示例和换行样式；工作区原有其他修改仍保留且未被纳入本次提交。
