import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import * as vue from 'vue'
import { createRouter, createMemoryHistory } from 'vue-router'
import { compileTemplate } from '@vue/compiler-sfc'
import { executeTypeScriptModule } from './lib/typescript-runtime.mjs'

const read = path => readFileSync(new URL(`../${path}`, import.meta.url), 'utf8')
const { createAdminRouteComponentCache, getRouteCacheName } = executeTypeScriptModule(read('src/router/routeComponentCache.ts'), 'routeComponentCache.ts', { vue })
const helpers = executeTypeScriptModule(read('src/views/log/auditLogHelpers.ts'), 'auditLogHelpers.ts')
const requests = [], pending = [], states = [], errors = []
let hold = false, mounts = 0
const list = category => query => {
  const result = { data: { list: [{ id: `${category}-${requests.length + 1}`, content: 'record' }], total: 1, page: query.page, pageSize: query.pageSize } }
  requests.push({ category, query })
  return hold ? new Promise(resolve => pending.push(() => resolve(result))) : Promise.resolve(result)
}
const { useAuditLogs } = executeTypeScriptModule(read('src/views/log/useAuditLogs.ts'), 'useAuditLogs.ts', {
  vue, 'element-plus': { ElMessageBox: { confirm: () => Promise.reject('unused') } },
  '@/utils/permission': { hasPerm: () => true },
  '@/api': { adminApi: { logList: list('operation'), loginLogList: list('login') } },
  './auditLogHelpers': helpers,
})

// A real Vue renderer/RouterView/KeepAlive exercises hook ordering, not mocked hooks.
const node = (type, text = '') => ({ type, text, parent: null, children: [], props: {} })
const remove = child => {
  if (!child.parent) return
  const list = child.parent.children, at = list.indexOf(child)
  if (at >= 0) list.splice(at, 1)
  child.parent = null
}
const renderer = vue.createRenderer({
  createElement: tag => node(tag), createText: text => node('text', text), createComment: text => node('comment', text),
  setText: (target, text) => { target.text = text }, setElementText: (target, text) => { target.text = text; target.children = [] },
  patchProp: (target, key, _old, value) => { target.props[key] = value },
  parentNode: target => target.parent,
  nextSibling: target => target.parent?.children[target.parent.children.indexOf(target) + 1] || null,
  insert: (child, parent, anchor = null) => {
    remove(child)
    const at = anchor ? parent.children.indexOf(anchor) : -1
    if (at < 0) parent.children.push(child); else parent.children.splice(at, 0, child)
    child.parent = parent
  }, remove,
})
const LogPage = vue.defineComponent({ name: 'LogPage', setup() {
  mounts++
  const state = useAuditLogs(); states.push(state)
  return () => vue.h('div', { class: 'logs' }, state.rows.value.map(row => row.id).join(','))
} })
const OtherPage = vue.defineComponent({ name: 'OtherPage', render: () => vue.h('div', 'other') })
const router = createRouter({ history: createMemoryHistory(), routes: [
  { path: '/log', name: 'Log', component: LogPage }, { path: '/other', name: 'Other', component: OtherPage },
] })
const cache = createAdminRouteComponentCache()
const included = vue.ref([getRouteCacheName('/log', 'Log'), getRouteCacheName('/other', 'Other')])
const frame = vue.ref(0)
const routeTemplate = read('src/views/layout/index.vue').match(/<router-view\b[\s\S]*?<\/router-view>/)?.[0]
assert.ok(routeTemplate, 'exercise the actual route cache template, including its prop binding')
const compiled = compileTemplate({ source: routeTemplate, filename: 'route-cache-harness.vue', id: 'route-cache-test' })
assert.deepEqual(compiled.errors, [])
const { render: renderRoute } = executeTypeScriptModule(compiled.code, 'route-cache-harness.ts', { vue })
const Root = vue.defineComponent({
  setup: () => ({ permsReady: true, cachedRouteNames: included, getCachedRouteComponent: cache.resolve }),
  render(...args) { frame.value; return renderRoute.apply(this, args) },
})
const settle = async () => { for (let i = 0; i < 8; i++) { await vue.nextTick(); await Promise.resolve() } }
const navigate = async path => { await router.push(path); await settle() }
const app = renderer.createApp(Root)
app.use(router); app.config.errorHandler = error => errors.push(error)
app.config.warnHandler = message => errors.push(new Error(message))
await router.push('/log'); await router.isReady(); app.mount(node('root')); await settle()
assert.equal(mounts, 1); assert.equal(requests.length, 1); assert.equal(states[0].rows.value.length, 1)
frame.value++; await settle()
assert.equal(mounts, 1, 'layout updates must not replace the cached route component')
assert.equal(requests.length, 1)
for (let i = 0; i < 3; i++) {
  await navigate('/other'); await navigate('/log')
  assert.equal(mounts, 1, 'returning through RouterView reuses one component instance')
  assert.equal(states[0].rows.value.length, 1, 'logs must load again after reactivation')
  assert.equal(requests.length, i + 2, 'reactivation issues exactly one list request')
}
states[0].category.value = 'login'; await settle()
await navigate('/other'); await navigate('/log')
assert.equal(states[0].category.value, 'login')
assert.equal(requests.at(-1).category, 'login')
assert.equal(states[0].rows.value.length, 1)

await navigate('/other'); hold = true; await navigate('/log')
assert.equal(pending.length, 1)
await navigate('/other'); pending.shift()(); await settle()
assert.equal(states[0].rows.value.length, 0, 'late responses cannot revive a deactivated page')
hold = false; await navigate('/log')
assert.equal(states[0].rows.value.length, 1)
assert.equal(states[0].loading.value, false)

// Closing a tab evicts its state; reopening mounts a fresh, loading page.
await navigate('/other')
included.value = [getRouteCacheName('/other', 'Other')]
cache.remove('/log'); await settle()
included.value.push(getRouteCacheName('/log', 'Log'))
await navigate('/log')
assert.equal(mounts, 2); assert.equal(states.at(-1).rows.value.length, 1)

// Menu-level caching disabled must still support repeated visits by remounting.
included.value = []; await settle()
await navigate('/other'); await navigate('/log')
assert.equal(mounts, 3); assert.equal(states.at(-1).rows.value.length, 1)
assert.deepEqual(errors, [])
app.unmount(); await settle()
assert.deepEqual(errors, [])
console.log('Route cache runtime passed: stable wrappers, real KeepAlive reentry, login tab, late responses, closed tabs and disabled caching')
