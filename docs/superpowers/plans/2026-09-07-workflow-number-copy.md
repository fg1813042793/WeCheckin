# Workflow Number Copy Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在钉钉 H5 流程中心列表和流程详情抽屉中，为流程单号提供统一的图标复制能力。

**Architecture:** 新增流程域内共用组件 `WorkflowCopyableText.vue`，封装文本省略、复制按钮、H5 标准剪贴板优先且 uView Pro `clipboard` 回退的复制策略与用户反馈。`WorkflowRecordTable.vue` 通过列级 `copyable` 配置启用该能力，`WorkflowDetailPanel.vue` 在详情标题中复用同一组件，不改变后端契约。

**Tech Stack:** Vue 3 `<script setup>`、TypeScript、uni-app、uView Pro、SCSS、Node.js 契约检查。

---

### Task 1: 为列表流程单号增加可复制渲染

**Files:**
- Create: `h5app/src/pages/workflow/components/WorkflowCopyableText.vue`
- Modify: `h5app/src/pages/workflow/components/WorkflowRecordTable.vue`
- Modify: `h5app/src/pages/workflow/components/WorkflowCenter.vue`
- Test: `h5app/scripts/check-workflow-module.mjs`

- [x] **Step 1: 先补失败契约检查**

在 `h5app/scripts/check-workflow-module.mjs` 中增加对共用复制组件、`copyable` 列和列表启用点的断言：

```js
{
  file: 'src/pages/workflow/components/WorkflowCopyableText.vue',
  patterns: [
    "import { clipboard } from 'uview-pro'",
    "uni.showToast({ title: '流程单号已复制', icon: 'success' })",
    "uni.showToast({ title: '复制失败，请重试', icon: 'none' })",
    '@click.stop="copyValue"',
    'aria-label="复制流程单号"',
  ],
},
{
  file: 'src/pages/workflow/components/WorkflowRecordTable.vue',
  patterns: [
    'copyable?: boolean',
    'v-else-if="column.copyable"',
    ':value="row.cells[column.key] || \'-\'"',
  ],
},
{
  file: 'src/pages/workflow/components/WorkflowCenter.vue',
  patterns: [
    "{ key: 'businessKey', label: '流程单号', width: 'minmax(180px, 1.35fr)', mobileHidden: true, copyable: true }",
  ],
},
```

- [x] **Step 2: 运行检查并确认先失败**

Run:

```bash
cd h5app && node scripts/check-workflow-module.mjs
```

Expected: FAIL，报告 `WorkflowCopyableText.vue` 不存在或缺少 `copyable` 相关内容。

- [x] **Step 3: 实现共用复制组件**

创建 `WorkflowCopyableText.vue`，核心实现为：

```ts
import { clipboard } from 'uview-pro'

const props = withDefaults(defineProps<{
  value?: string
}>(), {
  value: '',
})

const displayValue = computed(() => String(props.value || '').trim() || '-')
const copyable = computed(() => displayValue.value !== '-')

function copyValue() {
  if (!copyable.value)
    return
  clipboard(displayValue.value, {
    showToast: false,
    success: () => uni.showToast({ title: '流程单号已复制', icon: 'success' }),
    fail: () => uni.showToast({ title: '复制失败，请重试', icon: 'none' }),
  })
}
```

模板使用单行省略文本和紧凑图标按钮：

```vue
<view class="workflow-copyable-text">
  <text class="workflow-copyable-text__value" :title="displayValue">{{ displayValue }}</text>
  <u-button
    v-if="copyable"
    custom-class="workflow-copyable-text__button app-icon-button app-icon-button--small"
    title="复制流程单号"
    aria-label="复制流程单号"
    @click.stop="copyValue"
  >
    <view class="workflow-copyable-text__icon" aria-hidden="true" />
  </u-button>
</view>
```

复制图标使用两个边框方块的 CSS 绘制，按钮保持 `28px` 固定尺寸，不引入新图标依赖。

- [x] **Step 4: 通过列配置启用复制**

在 `WorkflowRecordTable.vue` 的列类型中增加：

```ts
copyable?: boolean
```

引入 `WorkflowCopyableText.vue`，并在普通文本分支前增加：

```vue
<WorkflowCopyableText
  v-else-if="column.copyable"
  :value="row.cells[column.key] || '-'"
/>
```

在 `WorkflowCenter.vue` 中只为流程单号列开启：

```ts
{ key: 'businessKey', label: '流程单号', width: 'minmax(180px, 1.35fr)', mobileHidden: true, copyable: true }
```

- [x] **Step 5: 运行定向检查**

Run:

```bash
cd h5app
node scripts/check-workflow-module.mjs
./node_modules/.bin/eslint scripts/check-workflow-module.mjs src/pages/workflow/components/WorkflowCopyableText.vue src/pages/workflow/components/WorkflowRecordTable.vue src/pages/workflow/components/WorkflowCenter.vue
```

Expected: 两个命令均 PASS。

- [x] **Step 6: 提交列表复制能力**

```bash
git add h5app/scripts/check-workflow-module.mjs h5app/src/pages/workflow/components/WorkflowCopyableText.vue h5app/src/pages/workflow/components/WorkflowRecordTable.vue h5app/src/pages/workflow/components/WorkflowCenter.vue
git commit -m "feat: 支持复制流程单号"
```

### Task 2: 在流程详情中复用复制组件

**Files:**
- Modify: `h5app/src/pages/workflow/components/WorkflowDetailPanel.vue`
- Test: `h5app/scripts/check-workflow-module.mjs`

- [x] **Step 1: 先增加详情接入的失败检查**

为 `WorkflowDetailPanel.vue` 增加以下契约断言：

```js
{
  file: 'src/pages/workflow/components/WorkflowDetailPanel.vue',
  patterns: [
    "import WorkflowCopyableText from './WorkflowCopyableText.vue'",
    '<WorkflowCopyableText :value="detail.instance.businessKey || \'-\'" />',
  ],
},
```

- [x] **Step 2: 运行检查并确认先失败**

Run:

```bash
cd h5app && node scripts/check-workflow-module.mjs
```

Expected: FAIL，报告 `WorkflowDetailPanel.vue` 缺少复制组件接入。

- [x] **Step 3: 调整详情标题区域**

导入 `WorkflowCopyableText.vue`，将原有完整的副标题 `<text>` 改为可容纳按钮的 `<view>`，业务编号部分改为：

```vue
<text>· 业务编号：</text>
<WorkflowCopyableText :value="detail.instance.businessKey || '-'" />
```

同时为 `.workflow-detail-panel__subtitle--business` 增加 `display: flex`、`align-items: center`、`flex-wrap: wrap` 和紧凑间距，保证长文本换行时复制按钮不溢出。

非历史展示中的“业务编号”摘要值同样改为 `WorkflowCopyableText`，保持详情组件内行为一致。

- [x] **Step 4: 运行定向检查**

Run:

```bash
cd h5app
node scripts/check-workflow-module.mjs
./node_modules/.bin/eslint scripts/check-workflow-module.mjs src/pages/workflow/components/WorkflowCopyableText.vue src/pages/workflow/components/WorkflowRecordTable.vue src/pages/workflow/components/WorkflowCenter.vue src/pages/workflow/components/WorkflowDetailPanel.vue
```

Expected: 两个命令均 PASS。

- [x] **Step 5: 运行 H5App 回归检查**

Run:

```bash
cd h5app
npm run type-check
node scripts/check-ui-style-guidelines.mjs
npm run build:h5
```

Expected: TypeScript、UI 规范检查和 H5 构建全部 PASS。如全量 lint 仍有与本次文件无关的既有基线问题，单独记录，不修改无关文件。

- [x] **Step 6: 执行浏览器验收**

在 H5 开发服务中验证：

1. 流程中心四类列表的流程单号后显示复制图标。
2. 点击图标后显示“流程单号已复制”，剪贴板内容与完整流程单号一致，且不打开详情。
3. 详情抽屉的业务编号后显示同款按钮，复制结果一致。
4. `1024px`、`920px` 和 `768px` 宽度下没有标题或表格溢出；列表原有横向滚动、固定操作列及手机卡片布局保持不变。

- [x] **Step 7: 提交详情接入与回归检查**

```bash
git add h5app/scripts/check-workflow-module.mjs h5app/src/pages/workflow/components/WorkflowDetailPanel.vue
git commit -m "feat: 在流程详情中复制单号"
```
