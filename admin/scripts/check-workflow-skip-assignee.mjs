import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import * as vue from 'vue'
import { parse, compileScript } from '@vue/compiler-sfc'
import { executeTypeScriptModule } from './lib/typescript-runtime.mjs'

const read = file => readFileSync(new URL(`../src/${file}`, import.meta.url), 'utf8')
const path = 'views/workflow/designer/components/NodeInspector.vue'
const script = compileScript(parse(read(path), { filename: path }).descriptor, { id: path }).content
const graph = executeTypeScriptModule(read('views/workflow/designer/graph.ts'), 'graph.ts')
const catalog = executeTypeScriptModule(read('views/workflow/designer/workflowFieldCatalog.ts'), 'catalog.ts')
const layout = executeTypeScriptModule(read('views/workflow/formLayout.ts'), 'formLayout.ts')
const { cloneDraft } = executeTypeScriptModule(read('types/workflow.ts'), 'workflowTypes.ts')
const component = executeTypeScriptModule(script, path, {
  vue,
  'element-plus': { ElMessage: { warning() {}, success() {}, error() {} } },
  '@element-plus/icons-vue': {},
  '../graph': graph,
  '../workflowFieldCatalog': catalog,
  '../../formLayout': layout,
  './WorkflowResultNotificationEditor.vue': {},
}).default
const draft = { key: 'skip-test', name: '无审批人跳过', form: [], nodes: [{ id: 'approval', type: 'approval', name: '上级评价', approvalMode: 'single', assignee: { type: 'manager', value: 'direct_manager' } }], edges: [] }
const props = vue.reactive({ draft, selectedNodeId: 'approval', readonly: false })
const changes = []
let state
const node = () => ({ children: [] })
const renderer = vue.createRenderer({ createElement: node, createText: node, createComment: node, setText() {}, setElementText() {}, patchProp() {}, parentNode: () => null, nextSibling: () => null, insert(child, parent) { parent.children.push(child) }, remove() {} })
const app = renderer.createApp({ setup() { state = vue.proxyRefs(component.setup(props, { expose() {}, emit: event => changes.push(event) })); return () => null } })
app.mount(node())
try {
  assert.notEqual(state.selectedNode.skipIfNoAssignee, true, '旧节点必须默认关闭')
  assert.equal(typeof state.updateSkipIfNoAssignee, 'function', '审批人配置必须允许启用无审批人跳过')
  state.updateSkipIfNoAssignee(true)
  assert.equal(state.selectedNode.skipIfNoAssignee, true)
  assert.equal(changes.at(-1), 'change')
  props.draft = cloneDraft(JSON.parse(JSON.stringify(props.draft)))
  await vue.nextTick()
  assert.equal(state.selectedNode.skipIfNoAssignee, true, '保存重载必须保留开关')
  props.readonly = true
  const before = changes.length
  state.updateSkipIfNoAssignee(false)
  assert.equal(state.selectedNode.skipIfNoAssignee, true)
  assert.equal(changes.length, before, '只读状态禁止修改及触发保存')
  props.readonly = false
  state.updateSkipIfNoAssignee(false)
  assert.ok(!('skipIfNoAssignee' in state.selectedNode), '关闭后恢复缺省行为')
  props.draft.nodes.push({ id: 'handle', type: 'handle', name: '办理' })
  props.selectedNodeId = 'handle'
  await vue.nextTick()
  state.updateSkipIfNoAssignee(true)
  assert.ok(!('skipIfNoAssignee' in state.selectedNode), '不能将跳过审批策略写入其他节点')
} finally { app.unmount() }

for (const [source, start, end, exported] of [
  [read('views/workflow/instances/index.vue'), 'function historyEventLabel(', 'function notificationStatusMeta(', 'historyEventLabel'],
  [readFileSync(new URL('../../h5app/src/pages/workflow/components/WorkflowDetailPanel.vue', import.meta.url), 'utf8'), 'function historyLabel(', 'function historyActorName(', 'historyLabel'],
]) {
  const body = source.slice(source.indexOf(start), source.indexOf(end))
  const labels = executeTypeScriptModule(`${body}\nexport { ${exported} }`, 'historyLabel.ts')
  assert.equal(labels[exported]('node_skipped'), '审批节点已自动跳过', 'Admin/H5 应展示可理解的跳过记录')
}
console.log('workflow skip-assignee configuration passed: default off, toggle, JSON reload, readonly and approval-only')
