import assert from 'node:assert/strict'
import { existsSync, readFileSync } from 'node:fs'
import * as vue from 'vue'
import { executeTypeScriptModule } from './lib/typescript-runtime.mjs'

const path = new URL('../src/router/routeTabTitles.ts', import.meta.url)
assert.ok(existsSync(path), '页签需要按独立流程地址维护标题')
const route = vue.reactive({ path: '/workflow/definitions/7/designer', fullPath: '/workflow/definitions/7/designer', name: 'WorkflowDesigner', meta: { title: '流程设计器' } })
const titles = executeTypeScriptModule(readFileSync(path, 'utf8'), 'routeTabTitles.ts', { vue, 'vue-router': { useRoute: () => route } })
const plain = value => JSON.parse(JSON.stringify(value))
const a = titles.routeToVisitedTab(route)
assert.equal(a.title, '流程设计器 #7')
const bRoute = { ...plain(route), path: '/workflow/definitions/8/designer', fullPath: '/workflow/definitions/8/designer' }
const b = titles.routeToVisitedTab(bRoute)
assert.equal(b.title, '流程设计器 #8')
const tabs = vue.reactive([a, b])
assert.equal(titles.updateVisitedTabTitle(tabs, a.fullPath, titles.workflowDesignerTabTitle(7, 'IC业务-直属上级')), true)
assert.equal(tabs[0].title, 'IC业务-直属上级 · #7')
assert.equal(tabs[1].title, '流程设计器 #8')
assert.equal(titles.routeToVisitedTab(route, tabs[0]).title, tabs[0].title, '切换页签不能重置为固定路由标题')
const restored = titles.normalizeVisitedTabs(JSON.parse(JSON.stringify(tabs)))
assert.equal(restored[0].title, tabs[0].title, '刷新必须保留已解析的管理名称')
assert.equal(titles.normalizeVisitedTabs([{ ...a, title: '流程设计器', customTitle: undefined }])[0].title, '流程设计器 #7', '旧保存页签在未加载前也应以编号区分')
assert.equal(titles.normalizeVisitedTabs([a, a, b]).length, 2)
assert.equal(titles.updateVisitedTabTitle(tabs, '/closed', '迟到响应'), false, '不能重新创建已关闭的页签')
assert.equal(titles.updateVisitedTabTitle(tabs, a.fullPath, '  '), false)
assert.equal(titles.routeToVisitedTab({ ...route, path: '/login' }), null)
assert.equal(titles.routeToVisitedTab({ path: '/user', fullPath: '/user', name: 'User', meta: { title: '用户管理' } }).title, '用户管理')
for (const [kind, label] of [['survey', '问卷'], ['exam', '考试']]) {
  const target = { path: `/${kind}/designer`, fullPath: `/${kind}/designer?id=5`, name: `${kind}Designer`, meta: { title: `${label}设计器` } }
  assert.equal(titles.routeToVisitedTab(target).title, `${label}设计器 #5`)
  assert.equal(titles.routeToVisitedTab({ ...target, fullPath: target.path }).title, `新建${label}`)
  assert.equal(titles.normalizeVisitedTabs([{ ...target, title: `${label}设计器` }])[0].title, `${label}设计器 #5`)
  const created = titles.formDesignerTabTitle(kind, 5, '相同管理名称')
  assert.equal(created, `相同管理名称 · ${label}#5`)
  const draftTab = titles.routeToVisitedTab({ ...target, fullPath: target.path })
  const opened = [draftTab, a, titles.routeToVisitedTab(target)]
  assert.equal(titles.replaceVisitedTab(opened, draftTab.fullPath, target, created), true)
  assert.equal(opened.length, 2, '新建页签替换为已保存页签，不能残留重复入口')
  assert.equal(opened[0].fullPath, target.fullPath)
  assert.equal(opened[0].title, created)
  assert.equal(opened[1].fullPath, a.fullPath)
  assert.equal(titles.replaceVisitedTab(opened, '/closed', target, created), false)
}
assert.notEqual(titles.formDesignerTabTitle('survey', 5, '同名'), titles.formDesignerTabTitle('exam', 5, '同名'))

const node = () => ({ children: [], parent: null })
const renderer = vue.createRenderer({
  createElement: node, createText: node, createComment: node, setText() {}, setElementText() {}, patchProp() {},
  parentNode: value => value.parent, nextSibling: () => null,
  insert(child, parent) { child.parent = parent; parent.children.push(child) },
  remove(child) { if (child.parent) child.parent.children = child.parent.children.filter(value => value !== child) },
})
const pages = []
const Child = vue.defineComponent({ name: 'TitlePage', setup() { pages.push(titles.useAdminRouteTabTitle()); return () => null } })
const app = renderer.createApp({
  setup() {
    vue.provide(titles.ADMIN_ROUTE_TAB_TITLE_KEY, (fullPath, title) => titles.updateVisitedTabTitle(tabs, fullPath, title))
    return () => vue.h(vue.KeepAlive, null, { default: () => vue.h(Child, { key: route.fullPath }) })
  },
})
app.mount(node())
Object.assign(route, bRoute)
await vue.nextTick()
assert.equal(pages.length, 2)
pages[1]('IC业务-HRBP · #8')
pages[0]('迟到的上级流程名称 · #7')
assert.equal(tabs[0].title, '迟到的上级流程名称 · #7', '后台缓存页面的响应只更新自身页签')
assert.equal(tabs[1].title, 'IC业务-HRBP · #8')
assert.equal(route.meta.title, '流程设计器', '不得修改共享路由 meta')
app.unmount()
pages[0]('关闭后的迟到响应')
assert.equal(tabs[0].title, '迟到的上级流程名称 · #7', '已销毁页面不得再修改重开的页签')
console.log('route tab titles passed: workflow identity, persistence, rename, route switching, delayed responses and disposal')
