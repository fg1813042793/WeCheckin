import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { compileScript, compileTemplate, parse } from '@vue/compiler-sfc'
import ts from 'typescript'
import * as vue from 'vue'

const read = path => readFileSync(new URL(`../src/${path}`, import.meta.url), 'utf8')
function execute(source, dependencies = {}) {
  const js = ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText
  const module = { exports: {} }
  // eslint-disable-next-line no-new-func
  new Function('require', 'module', 'exports', js)((name) => {
    assert.ok(name in dependencies, `unexpected import ${name}`)
    return dependencies[name]
  }, module, module.exports)
  return module.exports
}
const filename = 'pages/workflow/components/WorkflowReadOnlyGraph.vue'
const { descriptor } = parse(read(filename), { filename })
const script = compileScript(descriptor, { id: filename })
const template = compileTemplate({ source: descriptor.template.content, filename, id: filename, compilerOptions: { bindingMetadata: script.bindings } })
assert.deepEqual(template.errors, [])
const component = execute(script.content, {
  vue,
  '../workflow-graph-layout': execute(read('pages/workflow/workflow-graph-layout.ts')),
  '../workflow-status': execute(read('pages/workflow/workflow-status.ts')),
}).default
component.render = execute(template.code, { vue }).render
const node = (tag = '', text = '') => ({ tag, text, props: {}, children: [], parent: null })
const renderer = vue.createRenderer({
  createElement: tag => node(tag),
  createText: text => node('#text', text),
  createComment: () => node('#comment'),
  setText(element, text) { element.text = text },
  setElementText(element, text) {
    element.text = text
    element.children = []
  },
  patchProp(element, key, _previous, value) { element.props[key] = value },
  parentNode: element => element.parent,
  nextSibling: () => null,
  insert(child, parent) {
    child.parent = parent
    parent.children.push(child)
  },
  remove(child) {
    if (child.parent)
      child.parent.children = child.parent.children.filter(item => item !== child)
  },
})
const descendants = root => [root, ...root.children.flatMap(descendants)]
const text = root => [root.text, ...root.children.map(text)].join('')
async function settle() {
  for (let i = 0; i < 6; i++)
    await vue.nextTick()
}
function mount(rootComponent) {
  const root = node()
  const app = renderer.createApp(rootComponent)
  for (const tag of ['view', 'text', 'scroll-view', 'u-icon']) {
    app.component(tag, { setup: (_, { attrs, slots }) => () => vue.h(tag, attrs, slots.default?.()) })
  }
  app.config.errorHandler = (error) => {
    throw error
  }
  app.mount(root)
  return { root, app }
}
const props = vue.reactive({
  nodes: [
    { id: 'manager', type: 'approval', name: '上级评分', assigneeDisplay: 'David', approvalMode: 'single' },
    { id: 'employee', type: 'approval', name: '员工确认', assigneeDisplay: 'Foster', approvalMode: 'single', skipIfNoAssignee: true },
  ],
  edges: [{ id: 'e1', source: 'manager', target: 'employee' }],
  nodeProgress: [],
})
const graph = mount({ render: () => vue.h(component, props) })
function employeeCard() {
  return descendants(graph.root).find(item => String(item.props.class || '').split(' ').includes('workflow-graph__node') && text(item).includes('员工确认'))
}
try {
  await settle()
  assert.ok(!text(employeeCard()).includes('已跳过'), '开启开关但没有执行进度时不能误标')
  const originalStyle = JSON.stringify(employeeCard().props.style)
  props.nodeProgress = [{ nodeId: 'employee', nodeName: '员工确认', nodeType: 'approval', status: 'skipped' }]
  await settle()
  assert.match(text(employeeCard()), /已跳过/, '流程图必须展示后端实际跳过状态')
  assert.doesNotMatch(text(employeeCard()), /Foster|审批人：|单人审批/, '跳过节点不得显示当前解析出的审批人，避免误导')
  assert.ok(String(employeeCard().props.class).includes('workflow-graph__node--skipped'))
  assert.equal(JSON.stringify(employeeCard().props.style), originalStyle, '状态变化不能移动节点或连线')
  assert.match(text(graph.root), /David/, '正常节点仍显示审批人')
  for (const status of ['not_started', 'processing', 'completed', 'returned', 'terminated']) {
    props.nodeProgress = [{ nodeId: 'employee', status }]
    await settle()
    assert.doesNotMatch(text(employeeCard()), /已跳过/)
    assert.match(text(employeeCard()), /Foster/)
  }
}
finally { graph.app.unmount() }

// Exercise the actual detail-template binding so a correct graph cannot be left disconnected.
const detailSource = read('pages/workflow/components/WorkflowDetailPanel.vue')
const graphTag = detailSource.match(/<WorkflowReadOnlyGraph\b[\s\S]*?\/>/)[0]
const boundTemplate = compileTemplate({ source: graphTag, filename: 'DetailGraph.vue', id: 'detail-graph' })
let received
const detail = { nodes: props.nodes, edges: props.edges, nodeProgress: [{ nodeId: 'employee', status: 'skipped' }] }
const binding = mount({
  components: {
    WorkflowReadOnlyGraph: {
      props: ['nodes', 'edges', 'nodeProgress'],
      setup(value) {
        received = value
        return () => null
      },
    },
  },
  setup: () => ({ detail }),
  render: execute(boundTemplate.code, { vue }).render,
})
assert.equal(received.nodeProgress, detail.nodeProgress, '实例详情必须传入节点进度')
binding.app.unmount()
console.log('workflow graph progress passed: skipped display, assignee suppression, preview default, status refresh, stable layout and detail binding')
