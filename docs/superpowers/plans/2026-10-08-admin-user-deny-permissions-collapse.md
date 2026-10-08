# Admin 用户禁止权限默认折叠 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 Admin 用户管理新增、编辑表单中的“禁止权限”四个面板改为默认收起、可整体展开且不影响权限勾选数据。

**Architecture:** 在现有用户页面中增加一个独立的 Vue `ref` 作为纯展示状态，通过 `v-show` 同时控制两列禁止权限面板。状态在 `resetForm()` 中恢复为 `false`，从而覆盖新增和编辑两个入口；现有权限树实例、表单字段和保存逻辑保持不变。

**Tech Stack:** Vue 3 Composition API、TypeScript、Element Plus、Node.js 结构契约脚本、Vite。

---

## 文件结构

- Modify: `admin/scripts/check-user-list-ui.mjs` - 增加禁止权限折叠交互的静态契约检查。
- Modify: `admin/src/views/user/index.vue` - 增加折叠标题、展示状态、表单重置和两列面板显隐绑定。
- Verify only: `admin/package.json` - 使用现有 `check:user-list-ui`、`check:all` 和 `dev` 脚本，不修改依赖或脚本定义。

### Task 1: 用结构契约定义默认折叠行为

**Files:**
- Modify: `admin/scripts/check-user-list-ui.mjs`
- Test: `admin/scripts/check-user-list-ui.mjs`

- [ ] **Step 1: 写入失败的结构检查**

在 `requiredSnippets` 中加入以下契约：

```js
  'const denyApplicationPermissionExpanded = ref(false)',
  ':icon="denyApplicationPermissionExpanded ? \'ArrowUp\' : \'ArrowDown\'"',
  ':aria-expanded="denyApplicationPermissionExpanded"',
  '@click.stop="denyApplicationPermissionExpanded = !denyApplicationPermissionExpanded"',
  "{{ denyApplicationPermissionExpanded ? '收起' : '展开' }}",
  'denyApplicationPermissionExpanded.value = false',
```

并在通用片段检查后增加两列显隐绑定数量检查：

```js
const denyPermissionColumnCount = source.match(/<section v-show="denyApplicationPermissionExpanded"/g)?.length || 0
if (denyPermissionColumnCount !== 2) {
  throw new Error(`admin user deny permission collapse column count invalid: ${denyPermissionColumnCount}`)
}
```

- [ ] **Step 2: 运行局部检查，确认红灯原因正确**

Run:

```bash
cd admin && npm run check:user-list-ui
```

Expected: FAIL，错误首先指出缺少 `const denyApplicationPermissionExpanded = ref(false)`，证明新契约尚未实现。

- [ ] **Step 3: 检查测试改动范围**

Run:

```bash
git diff --check -- admin/scripts/check-user-list-ui.mjs
git diff -- admin/scripts/check-user-list-ui.mjs
```

Expected: 无空白错误，差异只包含上述折叠契约。

### Task 2: 实现禁止权限整体折叠

**Files:**
- Modify: `admin/src/views/user/index.vue:421`
- Modify: `admin/src/views/user/index.vue:562`
- Modify: `admin/src/views/user/index.vue:1184`
- Test: `admin/scripts/check-user-list-ui.mjs`

- [ ] **Step 1: 将禁止权限标题改为可操作分隔栏**

用以下模板替换现有仅含标题的 `permission-section-divider`：

```vue
<div class="permission-section-divider">
  <span>禁止权限</span>
  <el-button
    link
    type="primary"
    size="small"
    :icon="denyApplicationPermissionExpanded ? 'ArrowUp' : 'ArrowDown'"
    :aria-expanded="denyApplicationPermissionExpanded"
    @click.stop="denyApplicationPermissionExpanded = !denyApplicationPermissionExpanded"
  >
    {{ denyApplicationPermissionExpanded ? '收起' : '展开' }}
  </el-button>
</div>
```

- [ ] **Step 2: 让两列禁止权限面板共享显隐状态**

为客户端菜单列和接口列分别增加同一个 `v-show`：

```vue
<section v-show="denyApplicationPermissionExpanded" class="permission-column permission-column--menu">
```

```vue
<section v-show="denyApplicationPermissionExpanded" class="permission-column permission-column--api">
```

继续使用 `v-show`，使树组件保持挂载，收起和再次展开不会重新创建树或丢失当前勾选。

- [ ] **Step 3: 增加默认状态并在表单重置时收起**

在现有 `userApplicationPermissionExpanded` 旁增加：

```ts
const denyApplicationPermissionExpanded = ref(false)
```

在 `resetForm()` 末尾、调用 `resetExtraDataPickerFilters()` 前增加：

```ts
denyApplicationPermissionExpanded.value = false
```

新增和编辑入口都先调用 `resetForm()`，因此无需在 `showAdd()` 和 `showEdit()` 重复赋值。

- [ ] **Step 4: 运行局部检查，确认绿灯**

Run:

```bash
cd admin && npm run check:user-list-ui
```

Expected: PASS，无异常输出。

- [ ] **Step 5: 运行完整 Admin 门禁**

Run:

```bash
cd admin && npm run check:all
```

Expected: 所有结构检查、TypeScript 构建、Vite 构建和 bundle 检查通过。若工作区中其他未提交流程改动导致无关失败，保留完整错误并单独运行 `npm run check:user-list-ui` 与 `npm run build` 证明本变更。

- [ ] **Step 6: 检查最终差异**

Run:

```bash
git diff --check -- admin/scripts/check-user-list-ui.mjs admin/src/views/user/index.vue
git diff -- admin/scripts/check-user-list-ui.mjs admin/src/views/user/index.vue
```

Expected: 无空白错误；页面差异仅包含折叠标题、两个 `v-show`、状态声明和重置赋值。

- [ ] **Step 7: 提交实现**

```bash
git add admin/scripts/check-user-list-ui.mjs admin/src/views/user/index.vue
git commit -m "feat: 默认折叠用户禁止权限"
```

### Task 3: 浏览器验证用户管理表单

**Files:**
- Verify only: `admin/src/views/user/index.vue`

- [ ] **Step 1: 启动 Admin 开发服务器**

Run:

```bash
cd admin && npm run dev -- --host 127.0.0.1 --port 4174
```

Expected: Vite 在 `http://127.0.0.1:4174/` 启动；保持进程运行直到验证结束。

- [ ] **Step 2: 验证新增用户表单默认收起**

进入用户管理页，点击“增加用户”。确认“用户扩展应用权限”展开后，“禁止权限”只显示标题、箭头和“展开”按钮，四个禁止权限面板不可见；页面无框架错误覆盖层，控制台无相关错误。

- [ ] **Step 3: 验证展开、勾选和再次展开**

点击“展开”，确认四个禁止权限面板按两列完整显示。勾选任一禁止权限，点击“收起”再点击“展开”，确认刚才的勾选仍然存在。

- [ ] **Step 4: 验证重新打开表单恢复默认状态**

关闭新增用户表单并重新打开，确认“禁止权限”重新处于收起状态。对任一用户打开编辑表单，确认同样默认收起；展开后已有禁止权限勾选正确回显。

- [ ] **Step 5: 验证桌面和窄窗口布局**

分别在约 `1440x900` 和 `1024x768` 视口检查标题、按钮、两列面板无重叠、裁切或横向溢出，并保存验证截图到仓库外临时目录。
