# H5App 流程表单详情分组查看 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让独立流程表单详情页在 PC 端支持与发起、审批页面一致的分组/连续显示切换，同时让手机端保持连续显示。

**Architecture:** 复用当前工作区已经实现的 `WorkflowRuntimeForm` `continuous/tabs` 展示能力和 `workflow_start_form_presentation_v1` 存储键。独立详情页自行维护桌面宽度状态，将计算后的展示模式传给只读表单；现有历史上下分栏预览和后端数据契约不变。

**Tech Stack:** uni-app、Vue 3 Composition API、TypeScript、uView Pro、SCSS、Node.js 结构契约脚本。

---

## 文件结构

- Modify: `h5app/scripts/check-workflow-form-tabs.mjs` - 将独立表单详情页加入现有页签展示契约。
- Modify: `h5app/src/pages/workflow/components/WorkflowFormDetailPage.vue` - 增加展示偏好、桌面判断、切换按钮和表单 `presentation` 绑定。
- Verify only: `h5app/package.json` - 使用当前工作区已增加的 `check:workflow-form-tabs` 脚本，不修改其余脚本或依赖。

### Task 1: 用现有契约脚本定义详情页行为

**Files:**
- Modify: `h5app/scripts/check-workflow-form-tabs.mjs`
- Test: `h5app/scripts/check-workflow-form-tabs.mjs`

- [ ] **Step 1: 读取独立详情页源码**

在现有 `detailSource` 后加入：

```js
const formDetailSource = readFileSync(
  new URL('../src/pages/workflow/components/WorkflowFormDetailPage.vue', import.meta.url),
  'utf8',
)
```

- [ ] **Step 2: 增加独立详情页的失败契约**

在审批页断言之后加入：

```js
for (const snippet of [
  "type WorkflowFormPresentation = 'continuous' | 'tabs'",
  "const formPresentationStorageKey = 'workflow_start_form_presentation_v1'",
  'const formPresentation = ref<WorkflowFormPresentation>(loadFormPresentation())',
  'const desktopFormPresentation = ref(resolveDesktopFormPresentation())',
  'const effectiveFormPresentation = computed<WorkflowFormPresentation>',
  "return desktopFormPresentation.value ? formPresentation.value : 'continuous'",
  'function toggleFormPresentation()',
  'v-if="detail && desktopFormPresentation"',
  'custom-class="workflow-form-detail-page__presentation-action"',
  '分组显示',
  '连续显示',
  ':presentation="effectiveFormPresentation"',
  'class="workflow-form-detail-page__header-actions"',
]) {
  assert.ok(formDetailSource.includes(snippet), `WorkflowFormDetailPage.vue missing form presentation behavior: ${snippet}`)
}

assert.equal(
  formDetailSource.match(/:presentation="effectiveFormPresentation"/g)?.length || 0,
  1,
  'standalone form detail should bind the shared presentation mode once',
)

const formDetailActions = formDetailSource.match(/<view class="workflow-form-detail-page__header-actions">[\s\S]*?<u-tag[\s\S]*?<\/view>/)?.[0] || ''
assert.ok(formDetailActions, 'standalone form detail header actions missing')
assert.ok(
  formDetailActions.indexOf('workflow-form-detail-page__presentation-action') < formDetailActions.indexOf('<u-tag'),
  'standalone form detail presentation toggle must stay before the status tag',
)
```

- [ ] **Step 3: 运行契约检查，确认红灯**

Run:

```bash
cd h5app && pnpm check:workflow-form-tabs
```

Expected: FAIL，首先指出 `WorkflowFormDetailPage.vue` 缺少 `WorkflowFormPresentation` 或存储键，证明独立详情页尚未接入。

- [ ] **Step 4: 检查测试差异**

Run:

```bash
git diff --check -- h5app/scripts/check-workflow-form-tabs.mjs
git diff -- h5app/scripts/check-workflow-form-tabs.mjs
```

Expected: 无空白错误；保留原有发起页、审批页断言，仅新增独立详情页断言。

### Task 2: 接入独立详情页的共享展示偏好

**Files:**
- Modify: `h5app/src/pages/workflow/components/WorkflowFormDetailPage.vue:1-120`
- Modify: `h5app/src/pages/workflow/components/WorkflowFormDetailPage.vue:136-250`
- Test: `h5app/scripts/check-workflow-form-tabs.mjs`

- [ ] **Step 1: 增加展示类型与生命周期导入**

将 Vue 导入改为：

```ts
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
```

并在组件声明前增加：

```ts
type WorkflowFormPresentation = 'continuous' | 'tabs'
```

- [ ] **Step 2: 增加共享偏好和桌面状态**

在现有加载状态后加入：

```ts
const formPresentationStorageKey = 'workflow_start_form_presentation_v1'
const formPresentation = ref<WorkflowFormPresentation>(loadFormPresentation())
const desktopFormPresentation = ref(resolveDesktopFormPresentation())
let formPresentationResizeHandler: (() => void) | null = null
```

在 `fieldAccess` 后加入：

```ts
const effectiveFormPresentation = computed<WorkflowFormPresentation>(() => {
  return desktopFormPresentation.value ? formPresentation.value : 'continuous'
})
```

- [ ] **Step 3: 增加桌面判断、偏好切换和生命周期同步**

在 `watch` 前加入：

```ts
function loadFormPresentation(): WorkflowFormPresentation {
  return uni.getStorageSync(formPresentationStorageKey) === 'tabs' ? 'tabs' : 'continuous'
}

function resolveDesktopFormPresentation() {
  try {
    const info = uni.getSystemInfoSync()
    const width = Number(info.windowWidth || info.screenWidth || 0)
    if (width > 0)
      return width > 768
  }
  catch {
    // 读取失败时继续使用 H5 媒体查询。
  }

  // #ifdef H5
  if (typeof window !== 'undefined' && typeof window.matchMedia === 'function')
    return !window.matchMedia('(max-width: 768px)').matches
  // #endif

  return false
}

function syncDesktopFormPresentation() {
  desktopFormPresentation.value = resolveDesktopFormPresentation()
}

function toggleFormPresentation() {
  if (!desktopFormPresentation.value)
    return
  formPresentation.value = formPresentation.value === 'tabs' ? 'continuous' : 'tabs'
  uni.setStorageSync(formPresentationStorageKey, formPresentation.value)
}

onMounted(() => {
  syncDesktopFormPresentation()
  // #ifdef H5
  formPresentationResizeHandler = syncDesktopFormPresentation
  window.addEventListener('resize', formPresentationResizeHandler)
  // #endif
})

onBeforeUnmount(() => {
  // #ifdef H5
  if (formPresentationResizeHandler)
    window.removeEventListener('resize', formPresentationResizeHandler)
  // #endif
  formPresentationResizeHandler = null
})
```

- [ ] **Step 4: 将切换按钮放到状态标签左侧**

用以下操作区包裹现有状态标签：

```vue
<view class="workflow-form-detail-page__header-actions">
  <u-button
    v-if="detail && desktopFormPresentation"
    custom-class="workflow-form-detail-page__presentation-action"
    size="small"
    plain
    @click="toggleFormPresentation"
  >
    <u-icon :name="formPresentation === 'tabs' ? 'list' : 'grid'" size="15px" color="#0f766e" />
    <text>{{ formPresentation === 'tabs' ? '连续显示' : '分组显示' }}</text>
  </u-button>
  <u-tag
    v-if="detail"
    :text="workflowInstanceStatusMeta(detail.instance.status).label"
    :type="workflowInstanceStatusMeta(detail.instance.status).type"
    size="mini"
  />
</view>
```

- [ ] **Step 5: 将有效展示模式传给只读表单**

在 `WorkflowRuntimeForm` 上增加：

```vue
:presentation="effectiveFormPresentation"
```

- [ ] **Step 6: 增加操作区样式**

在标题样式后加入：

```scss
.workflow-form-detail-page__header-actions {
  flex: 0 0 auto;
  display: flex;
  align-items: center;
  gap: 10px;
}

:deep(.workflow-form-detail-page__presentation-action) {
  width: 104px;
  margin: 0;
  color: #0f766e;
}

:deep(.workflow-form-detail-page__presentation-action .u-button__text) {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}
```

移动端由 `v-if="desktopFormPresentation"` 隐藏切换按钮，不新增移动端专用覆盖。

- [ ] **Step 7: 运行契约检查，确认绿灯**

Run:

```bash
cd h5app && pnpm check:workflow-form-tabs
```

Expected: 输出 `Workflow desktop form tabs checks passed`。

- [ ] **Step 8: 运行针对性和模块验证**

Run:

```bash
cd h5app
pnpm exec eslint src/pages/workflow/components/WorkflowFormDetailPage.vue scripts/check-workflow-form-tabs.mjs
pnpm type-check
pnpm check:ui-style
pnpm build:h5
```

Expected: 四条命令退出码均为 `0`。另运行 `pnpm lint` 记录全量基线；若仍是已知的无关遗留错误，保留输出并以针对性 ESLint 作为本变更证据。

- [ ] **Step 9: 检查最终差异并提交**

Run:

```bash
git diff --check -- h5app/scripts/check-workflow-form-tabs.mjs h5app/src/pages/workflow/components/WorkflowFormDetailPage.vue
git diff -- h5app/scripts/check-workflow-form-tabs.mjs h5app/src/pages/workflow/components/WorkflowFormDetailPage.vue
git add h5app/scripts/check-workflow-form-tabs.mjs h5app/src/pages/workflow/components/WorkflowFormDetailPage.vue
git commit -m "feat: 支持流程表单详情分组查看"
```

Expected: 提交只包含契约脚本和独立详情页，不包含其他工作区流程改动。

### Task 3: 浏览器验证详情页展示模式

**Files:**
- Verify only: `h5app/src/pages/workflow/components/WorkflowFormDetailPage.vue`

- [ ] **Step 1: 启动 H5App 开发服务器**

Run:

```bash
cd h5app && pnpm dev:h5 -- --host 127.0.0.1 --port 5175
```

Expected: H5App 在 `http://127.0.0.1:5175/` 启动并连接当前本地后端。

- [ ] **Step 2: 验证标准 PC**

在约 `1440x900` 视口打开一条流程记录并进入独立“表单详情”页。确认切换按钮位于状态标签左侧；选择“分组显示”后顶部出现字段/分组页签，点击页签只展示对应内容，再切回连续显示可显示完整表单。

- [ ] **Step 3: 验证共享偏好**

在详情页选择分组显示，返回发起页或审批任务页，确认仍使用分组模式；再切换为连续显示并返回详情页，确认详情页读取同一偏好。

- [ ] **Step 4: 验证 PC 小屏和手机**

在 `1024x768` 检查标题、按钮、状态标签和页签无重叠或横向页面溢出。在 `390x844` 检查切换按钮不显示、表单连续展示、字段与明细列表正常滚动。

- [ ] **Step 5: 检查运行时健康状态**

确认页面身份和 URL 正确、无 Vite 错误覆盖层、浏览器控制台无本变更相关错误，并保存 PC 分组、PC 连续和手机连续状态截图到仓库外临时目录。
