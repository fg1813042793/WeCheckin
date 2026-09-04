import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

function read(path) {
  return readFileSync(new URL(path, import.meta.url), 'utf8')
}

const agents = read('../AGENTS.md')
const developmentGuidelines = read('../docs/development-guidelines.md')
const uiGuidelines = read('../docs/ui-layout-style-guidelines.md')
const commonStyle = read('../src/common/style.scss')
const layoutStyle = read('../src/common/ui-layout.scss')
const appShell = read('../src/components/app-shell/app-shell.vue')
const workflowRuntimeForm = read('../src/pages/workflow/components/WorkflowRuntimeForm.vue')
const workflowSummary = read('../src/pages/workflow/components/WorkflowSummarySection.vue')

assert.match(agents, /ui-layout-style-guidelines\.md/)
assert.match(developmentGuidelines, /ui-layout-style-guidelines\.md/)

for (const rule of [
  '### PC 小屏幕统一样式（强制）',
  '`769px - 1023px`',
  '`app-pc-control-scope`',
  '`app-workflow-form`',
  'Teleport',
  '左右滚动',
  '向上展开',
]) {
  assert.match(developmentGuidelines, new RegExp(rule.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')))
}

for (const heading of [
  '响应式断点与验证尺寸',
  '页面布局',
  '表格页面',
  '搜索与筛选栏',
  'PC 小窗口控件',
  '弹窗与抽屉',
  '新页面验收清单',
]) {
  assert.match(uiGuidelines, new RegExp(`## ${heading}`))
}

for (const token of [
  '--app-page-content-max-width',
  '--app-control-height',
  '--app-input-font-size',
  '--app-icon-size',
  '--app-button-icon-size',
  '--app-icon-button-size',
  '--app-loading-size',
  '--app-counter-font-size',
  '--app-table-header-height',
  '--app-dialog-width-medium',
]) {
  assert.match(layoutStyle, new RegExp(token))
}

for (const className of [
  'app-layout-page',
  'app-page-header',
  'app-filter-bar',
  'app-data-table',
  'app-dialog-shell',
  'app-pc-control-scope',
]) {
  assert.match(layoutStyle, new RegExp(`\\.${className}`))
}

assert.match(commonStyle, /@import ['"]\.\/ui-layout\.scss['"];/)
assert.match(layoutStyle, /@media screen and \(max-width: 768px\)/)
assert.match(layoutStyle, /@media screen and \(min-width: 769px\) and \(max-width: 1023px\)/)
assert.match(layoutStyle, /@media screen and \(min-width: 1440px\)/)
assert.match(layoutStyle, /\.app-pc-control-scope\s*\{[\s\S]*?font-size:\s*14px;/)
for (const selector of [
  '.u-btn.u-size-small',
  '.u-input__input:not(.u-input__textarea)',
  '.uni-input-placeholder',
  '.uni-textarea-placeholder',
  '.u-textarea__count',
  '.u-input__count',
  '.u-tag.u-size-mini',
  '.u-icon__icon',
  '.u-icon__img',
  '.u-btn .u-icon__icon',
  '.u-loading-circle',
  '.app-icon-button',
  '.u-checkbox__icon-wrap',
  '.u-radio__icon-wrap',
  '.u-pagination .custom-class',
]) {
  assert.match(layoutStyle, new RegExp(selector.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')))
}
assert.match(layoutStyle, /\.app-pc-control-scope \.u-input:not\(:has\(\.u-input__textarea\)\)\)\s*\{[^}]*height:\s*var\(--app-control-height\) !important;/)
assert.match(layoutStyle, /\.app-pc-control-scope \.uni-input-placeholder\)[^{]*\{[^}]*font-size:\s*var\(--app-input-font-size\) !important;/)
assert.match(layoutStyle, /\.app-pc-control-scope \.uni-textarea-placeholder\)[^{]*\{[^}]*font-size:\s*var\(--app-input-font-size\) !important;/)
assert.match(layoutStyle, /\.app-pc-control-scope \.u-icon__icon\)[^{]*\{[^}]*font-size:\s*var\(--app-icon-size\) !important;/)
assert.match(layoutStyle, /\.app-pc-control-scope \.u-btn \.u-icon__icon\)[^{]*\{[^}]*font-size:\s*var\(--app-button-icon-size\) !important;/)
assert.match(layoutStyle, /\.app-pc-control-scope \.u-loading-circle\)[^{]*\{[^}]*width:\s*var\(--app-loading-size\) !important;/)
assert.match(layoutStyle, /\.app-icon-button\)\s*\{[^}]*width:\s*var\(--app-icon-button-size, 32px\) !important;/)
assert.match(layoutStyle, /\.app-pc-control-scope \.u-textarea__count\)[^{]*\{[^}]*font-size:\s*var\(--app-counter-font-size\) !important;/)
assert.match(appShell, /app-shell__content app-pc-control-scope/)
assert.match(appShell, /:class="\{ 'app-shell__tabs--overflow': tabsOverflow \}"/)
assert.match(appShell, /app-shell__tab-scroll-control--left/)
assert.match(appShell, /app-shell__tab-scroll-control--right/)
assert.match(appShell, /\.app-shell__tabs--overflow\s*\{[^}]*padding:\s*0 46px;/)
assert.match(appShell, /\.app-shell__tab-scroll-control\s*\{[^}]*position:\s*absolute;[^}]*top:\s*50%;[^}]*transform:\s*translateY\(-50%\);/)
assert.match(workflowRuntimeForm, /class="workflow-form app-workflow-form app-pc-control-scope"/)
for (const selector of [
  '.app-workflow-form .workflow-form__field-label',
  '.app-workflow-form .workflow-detail__empty',
  '.app-workflow-form .workflow-control__empty',
  '.app-workflow-form .workflow-picker',
  '.app-workflow-form .workflow-attachment__item',
  '.app-pc-control-scope .workflow-start-page__form-card',
]) {
  assert.match(layoutStyle, new RegExp(selector.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')))
}
assert.match(layoutStyle, /\.app-workflow-form \.workflow-form__field-label\)[^{]*\{[^}]*font-size:\s*13px !important;/)
assert.match(layoutStyle, /\.app-workflow-form \.workflow-detail__empty\)[^{]*\{[^}]*min-height:\s*96px !important;/)
assert.match(layoutStyle, /\.app-workflow-form \.workflow-picker\)[^{]*\{[^}]*min-height:\s*36px !important;/)
assert.match(workflowSummary, /\.workflow-summary__filter-actions[\s\S]*grid-column: 5 \/ 7;/)
assert.match(workflowSummary, /\.workflow-summary__filter-actions[^{]*\{[^}]*width:\s*fit-content;/)

console.log('H5App UI style guideline checks passed')
