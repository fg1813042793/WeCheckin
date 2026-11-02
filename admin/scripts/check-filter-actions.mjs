import assert from 'node:assert/strict'
import { readFileSync, readdirSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { dirname, join, relative } from 'node:path'
import * as vue from 'vue'
import { parse, compileScript } from '@vue/compiler-sfc'
import { renderToString } from '@vue/server-renderer'
import { executeTypeScriptModule } from './lib/typescript-runtime.mjs'

const root = join(dirname(fileURLToPath(import.meta.url)), '..')
function component(path, imports = {}) {
  const filename = join(root, path)
  const { descriptor } = parse(readFileSync(filename, 'utf8'), { filename })
  const script = compileScript(descriptor, { id: path, inlineTemplate: true })
  return executeTypeScriptModule(script.content, filename + '.ts', { vue, ...imports }).default
}
const actions = component('src/components/admin-ui/AdminFilterActions.vue')
async function render(target, props = {}, slots = {}) {
  const buttons = []
  const app = vue.createSSRApp({ render: () => vue.h(target, props, slots) })
  app.component('ElButton', { inheritAttrs: false, setup: (_, { attrs, slots }) => { buttons.push({ ...attrs, label: slots.default?.().map(node => node.children).join('') }); return () => vue.h('button', attrs, slots.default?.()) } })
  app.component('ElCollapseTransition', { setup: (_, { slots }) => () => slots.default?.() })
  return { html: await renderToString(app), buttons }
}
let searches = 0, resets = 0
const regular = await render(actions, { onSearch: () => searches++, onReset: () => resets++ })
assert.equal(regular.buttons.length, 2)
assert.deepEqual(regular.buttons.map(button => button.label), ['查询', '重置'])
assert.equal(regular.buttons[0].type, 'primary')
assert.ok(!regular.buttons[1].type)
assert.equal(regular.buttons[0]['native-type'], 'button')
assert.ok(regular.buttons.every(button => !button.icon))
regular.buttons[0].onClick(); regular.buttons[1].onClick()
assert.equal(searches, 1); assert.equal(resets, 1)
const native = await render(actions, { nativeType: 'submit', onSearch: () => searches++ })
assert.equal(native.buttons[0]['native-type'], 'submit')
native.buttons[0].onClick()
assert.equal(searches, 1, 'native form submit must not also emit click search')
const locked = await render(actions, { disabled: true, onSearch: () => searches++, onReset: () => resets++ })
assert.ok(locked.buttons.every(button => button.disabled))
locked.buttons[0].onClick(); locked.buttons[1].onClick()
assert.equal(searches, 1); assert.equal(resets, 1)
const busy = await render(actions, { loading: true, onSearch: () => searches++ })
assert.equal(busy.buttons[0].loading, true)
busy.buttons[0].onClick(); assert.equal(searches, 1)

const searchBar = component('src/components/admin-ui/AdminSearchBar.vue', {
  './AdminFilterActions.vue': { __esModule: true, default: actions },
  '@element-plus/icons-vue': { ArrowDown: () => vue.h('i'), ArrowUp: () => vue.h('i') },
})
const bar = await render(searchBar, { title: '', onSearch: () => searches++, onReset: () => resets++ }, { default: () => vue.h('input', { placeholder: '最后一个筛选项' }) })
assert.ok(bar.html.indexOf('最后一个筛选项') < bar.html.indexOf('admin-filter-actions'))
assert.equal(bar.buttons[0]['native-type'], 'submit')
bar.buttons[0].onClick(); assert.equal(searches, 1)
bar.buttons[1].onClick(); assert.equal(resets, 2)

const searchBarSource = readFileSync(join(root, 'src/components/admin-ui/AdminSearchBar.vue'), 'utf8')
assert.match(
  searchBarSource,
  /\.admin-ui-search-bar__content\s*\{[^}]*padding:\s*12px\s+14px;/s,
  'AdminSearchBar content must use the shared 12px 14px filter padding',
)

const userFeedbackSource = readFileSync(join(root, 'src/views/user-feedback/index.vue'), 'utf8')
const userFeedbackSearchBar = userFeedbackSource.match(/<AdminSearchBar\b[\s\S]*?<\/AdminSearchBar>/)?.[0] || ''
assert.ok(userFeedbackSearchBar, 'user feedback page must render AdminSearchBar')
assert.match(userFeedbackSearchBar, /^<AdminSearchBar\b[^>]*\btitle=""/, 'user feedback filter must not render a visible search-bar heading')
assert.ok(!/^<AdminSearchBar\b[^>]*\bcollapsible\b/.test(userFeedbackSearchBar), 'user feedback filter must use the compact always-visible layout')
assert.ok(!/<el-form-item\b[^>]*\blabel=/.test(userFeedbackSearchBar), 'list filters must not render visible form-item labels')
for (const accessibleControl of ['el-input', 'el-select', 'el-input-number', 'el-date-picker']) {
  const matches = userFeedbackSearchBar.match(new RegExp(`<${accessibleControl}\\b[^>]*>`, 'g')) || []
  assert.ok(matches.every(tag => /\baria-label=/.test(tag)), `${accessibleControl} filters must keep an accessible name with aria-label`)
}

const userListFiltersSource = readFileSync(join(root, 'src/views/user/components/UserListFilters.vue'), 'utf8')
assert.match(
  userListFiltersSource,
  /\.user-list-filters\s*\{[^}]*padding:\s*12px\s+14px;/s,
  'user list filters must use the shared 12px 14px filter padding',
)

const auditFiltersSource = readFileSync(join(root, 'src/views/log/AuditLogFilters.vue'), 'utf8')
assert.match(
  auditFiltersSource,
  /\.audit-search-card\s+:deep\(\.el-card__body\)\s*\{[^}]*padding:\s*12px\s+14px;/s,
  'audit log filters must use the shared 12px 14px filter padding',
)

const operationsListSource = readFileSync(join(root, 'src/styles/operations-list.css'), 'utf8')
assert.match(
  operationsListSource,
  /body \.feedback-page \.admin-ui-search-bar__content\s*\{[^}]*padding:\s*12px\s+14px;/s,
  'feedback page must not override the shared 12px 14px filter padding',
)

// Detect future hand-written reset/query pairs that escape the shared presentation.
function files(dir) { return readdirSync(dir, { withFileTypes: true }).flatMap(item => item.isDirectory() ? files(join(dir, item.name)) : item.name.endsWith('.vue') ? [join(dir, item.name)] : []) }
for (const filename of files(join(root, 'src/views'))) {
  const { descriptor } = parse(readFileSync(filename, 'utf8'))
  const template = descriptor.template?.content || ''
  assert.ok(!/<el-button\b[^>]*>\s*重置\s*<\/el-button>/.test(template), `${relative(root, filename)} must use AdminFilterActions for query/reset`)
  for (const filterBar of template.match(/<AdminSearchBar\b[\s\S]*?<\/AdminSearchBar>/g) || []) {
    assert.ok(!/<el-form-item\b[^>]*\blabel=/.test(filterBar), `${relative(root, filename)} AdminSearchBar filters must not render visible form-item labels`)
  }
  for (const rightBlock of template.match(/<div\b[^>]*class="[^"]*admin-toolbar__right[^"]*"[^>]*>[\s\S]*?<\/div>/g) || []) {
    assert.ok(!/<AdminFilterActions\b/.test(rightBlock), `${relative(root, filename)} must keep filter actions beside the last filter`)
    assert.ok(!/>\s*(查询|搜索|重置)\s*</.test(rightBlock), `${relative(root, filename)} must not right-align query/reset actions`)
  }
}
console.log('Filter actions checks passed: order, appearance props, click/reset, native submit, disabled/loading guards and page coverage')
