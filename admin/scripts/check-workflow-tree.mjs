import fs from 'node:fs'
import path from 'node:path'
import { createRequire } from 'node:module'
import ts from 'typescript'

const root = process.cwd()
const require = createRequire(import.meta.url)

function loadTypeScriptModule(relativePath, overrides = {}) {
  const filename = path.join(root, relativePath)
  const source = fs.readFileSync(filename, 'utf8')
  const output = ts.transpileModule(source, {
    compilerOptions: {
      module: ts.ModuleKind.CommonJS,
      target: ts.ScriptTarget.ES2020,
      esModuleInterop: true,
    },
    fileName: filename,
  }).outputText
  const module = { exports: {} }
  const runner = new Function('module', 'exports', 'require', output)
  runner(module, module.exports, request => request in overrides ? overrides[request] : require(request))
  return module.exports
}

function assert(condition, message) {
  if (!condition) throw new Error(message)
}

const graph = loadTypeScriptModule('src/views/workflow/designer/graph.ts')
const workflowTypes = loadTypeScriptModule('src/views/workflow/types.ts')
const tree = loadTypeScriptModule('src/views/workflow/designer/flowTree.ts', { './graph': graph })
const layout = loadTypeScriptModule('src/views/workflow/designer/layout.ts')
const typeSource = fs.readFileSync(path.join(root, 'src/views/workflow/types.ts'), 'utf8')
const insertSource = fs.readFileSync(path.join(root, 'src/views/workflow/designer/components/FlowInsertButton.vue'), 'utf8')
const inspectorSource = fs.readFileSync(path.join(root, 'src/views/workflow/designer/components/NodeInspector.vue'), 'utf8')
const cardSource = fs.readFileSync(path.join(root, 'src/views/workflow/designer/components/WorkflowNodeCard.vue'), 'utf8')
const sequenceSource = fs.readFileSync(path.join(root, 'src/views/workflow/designer/components/WorkflowSequence.vue'), 'utf8')
const canvasSource = fs.readFileSync(path.join(root, 'src/views/workflow/designer/components/WorkflowCanvas.vue'), 'utf8')
const graphNodeSource = fs.readFileSync(path.join(root, 'src/views/workflow/designer/components/WorkflowGraphNode.vue'), 'utf8')
const curveEdgeSource = fs.readFileSync(path.join(root, 'src/views/workflow/designer/components/WorkflowCurveEdge.vue'), 'utf8')
const layoutSource = fs.readFileSync(path.join(root, 'src/views/workflow/designer/layout.ts'), 'utf8')

const insertionDraft = {
  key: 'insert_test',
  name: '指定连线插入',
  nodes: [
    { id: 'start', type: 'start', name: '发起人' },
    { id: 'end', type: 'end', name: '流程结束' },
  ],
  edges: [{ id: 'start_to_end', source: 'start', target: 'end' }],
}
const inserted = graph.insertNodeAtEdge(insertionDraft, 'start_to_end', 'approval')
assert(inserted?.type === 'approval', '指定连线应插入审批节点')
assert(inserted?.notification?.enabled === true, '新审批节点应默认开启任务到达通知')
assert(inserted?.notification?.channels?.join(',') === 'in_app,dingtalk_oa', '新审批节点应默认启用站内和钉钉 OA 通知')
assert(inserted?.notification?.content === '你有一项待处理任务：{{nodeName}}', '新审批节点的默认通知模板无效')
assert(!insertionDraft.edges.some(edge => edge.id === 'start_to_end'), '原连线应被替换')
assert(insertionDraft.edges.some(edge => edge.source === 'start' && edge.target === inserted.id), '应连接到新节点')
assert(insertionDraft.edges.some(edge => edge.source === inserted.id && edge.target === 'end'), '新节点应连接到原目标')

const anchoredDraft = {
  key: 'anchor_test',
  name: '连接点保留',
  nodes: [
    { id: 'start', type: 'start', name: '发起人' },
    { id: 'end', type: 'end', name: '流程结束' },
  ],
  edges: [{ id: 'anchored', source: 'start', target: 'end', sourceHandle: 'right', targetHandle: 'left' }],
}
const anchoredNode = graph.insertNodeAtEdge(anchoredDraft, 'anchored', 'approval')
const anchoredBefore = anchoredDraft.edges.find(edge => edge.source === 'start' && edge.target === anchoredNode.id)
const anchoredAfter = anchoredDraft.edges.find(edge => edge.source === anchoredNode.id && edge.target === 'end')
assert(anchoredBefore?.sourceHandle === 'right' && !anchoredBefore?.targetHandle, '插入节点应保留原来源锚点并重置新目标锚点')
assert(!anchoredAfter?.sourceHandle && anchoredAfter?.targetHandle === 'left', '插入节点应保留原目标锚点并重置新来源锚点')
assert(graph.removeNode(anchoredDraft, anchoredNode.id), '应能删除刚插入的审批节点')
assert(anchoredDraft.edges[0]?.sourceHandle === 'right' && anchoredDraft.edges[0]?.targetHandle === 'left', '删除中间节点应恢复原连线两端锚点')

for (const type of ['handle', 'cc', 'notify', 'automation', 'timer']) {
  const draft = {
    key: `insert_${type}`,
    name: `插入 ${type}`,
    nodes: [
      { id: 'start', type: 'start', name: '发起人' },
      { id: 'end', type: 'end', name: '流程结束' },
    ],
    edges: [{ id: 'start_to_end', source: 'start', target: 'end' }],
  }
  const node = graph.insertNodeAtEdge(draft, 'start_to_end', type)
  assert(node?.type === type, `指定连线应插入 ${type} 节点`)
  assert(draft.edges.some(edge => edge.source === 'start' && edge.target === node.id), `${type} 节点应连接原起点`)
  assert(draft.edges.some(edge => edge.source === node.id && edge.target === 'end'), `${type} 节点应连接原终点`)
  if (type === 'handle') {
    assert(node.notification?.enabled === true && node.notification.content === '你有一项待处理任务：{{nodeName}}', '新办理节点应默认开启任务到达通知')
  }
  if (type === 'cc') {
    assert(node.name === '抄送' && node.notification?.enabled === true && node.notification.content === '{{starterName}} 发起的流程已抄送给你', '新抄送节点的默认通知无效')
  }
  if (type === 'notify') {
    assert(node.name === '通知' && node.notification?.enabled === true && node.notification.content === '流程已到达 {{nodeName}}', '新通知节点的默认配置无效')
    assert(graph.removeNode(draft, node.id), '应能删除通知节点')
    assert(draft.edges.some(edge => edge.source === 'start' && edge.target === 'end'), '删除通知节点应恢复原流程连线')
  }
}

const legacyDraft = workflowTypes.cloneDraft({
  schemaVersion: 1,
  key: 'legacy',
  name: '旧流程',
  form: [],
  nodes: [{ id: 'approval', type: 'approval', name: '审批' }],
  edges: [],
})
assert(legacyDraft.nodes[0]?.notification === undefined, '旧流程缺少 notification 时不得自动开启')

for (const [source, snippet, message] of [
  [typeSource, "'handle' | 'cc' | 'notify' | 'automation' | 'timer'", '流程节点类型缺少通知节点'],
  [typeSource, "export type WorkflowNotificationChannel = 'in_app' | 'dingtalk_oa'", '流程类型缺少通知渠道'],
  [typeSource, 'notification?: WorkflowNotificationConfig', '流程节点缺少通知配置'],
  [typeSource, "'initiator'", '处理人来源类型缺少发起人'],
  [typeSource, 'sourceHandle?: WorkflowEdgeHandle', '流程连线缺少来源连接点字段'],
  [typeSource, 'targetHandle?: WorkflowEdgeHandle', '流程连线缺少目标连接点字段'],
  [insertSource, "choose('handle')", '插入菜单缺少办理节点'],
  [insertSource, "choose('cc')", '插入菜单缺少抄送节点'],
  [insertSource, '<b>抄送</b>', '插入菜单应独立显示抄送'],
  [insertSource, "choose('notify')", '插入菜单缺少通知节点'],
  [insertSource, '<b>通知</b>', '插入菜单应独立显示通知'],
  [insertSource, "choose('automation')", '插入菜单缺少自动动作节点'],
  [insertSource, "choose('timer')", '插入菜单缺少定时节点'],
  [inspectorSource, "selectedNode.type === 'automation'", '节点配置缺少自动动作配置'],
  [inspectorSource, 'delaySeconds', '节点配置缺少等待时长配置'],
  [inspectorSource, 'label="发起人" value="initiator"', '节点配置缺少发起人处理人来源'],
  [inspectorSource, "selectedNode.assignee?.type !== 'initiator'", '发起人处理人来源不应显示处理人标识输入'],
  [inspectorSource, '任务到达通知', '审批和办理节点缺少通知开关'],
  [inspectorSource, 'notificationChannels', '节点配置缺少通知渠道复选'],
  [inspectorSource, '通知标题', '节点配置缺少通知标题'],
  [inspectorSource, '通知正文', '节点配置缺少通知正文'],
  [inspectorSource, "selectedNode.type !== 'notify'", '通知节点不应允许关闭通知'],
  [inspectorSource, 'notificationHelpVisible', '节点配置缺少通知消息说明弹窗状态'],
  [inspectorSource, 'QuestionFilled', '通知配置标题旁缺少说明图标'],
  [inspectorSource, 'title="通知消息配置说明"', '通知说明弹窗缺少清晰标题'],
  [inspectorSource, 'append-to-body', '通知说明弹窗必须挂载到 body，避免被节点抽屉遮挡'],
  [inspectorSource, '{{workflowName}}', '通知说明缺少流程名称占位符'],
  [inspectorSource, '{{nodeName}}', '通知说明缺少节点名称占位符'],
  [inspectorSource, '{{starterName}}', '通知说明缺少发起人占位符'],
  [inspectorSource, '{{instanceId}}', '通知说明缺少流程实例占位符'],
  [inspectorSource, '{{taskId}}', '通知说明缺少任务占位符'],
  [inspectorSource, '发送时标题最多 64 个字符，正文最多 1000 个字符', '通知说明缺少最终发送长度限制'],
  [inspectorSource, '不支持直接读取表单字段或流程变量', '通知说明缺少模板安全边界'],
  [cardSource, "node.type === 'handle'", '节点卡片缺少办理节点展示'],
  [cardSource, "node.type === 'notify'", '节点卡片缺少通知节点展示'],
  [cardSource, 'notificationChannelSummary', '节点卡片缺少通知渠道摘要'],
  [cardSource, "initiator: '发起人'", '节点卡片缺少发起人来源说明'],
  [cardSource, "node.type === 'start'", '开始节点缺少独立事件样式'],
  [cardSource, 'class="start-node__event"', '开始节点缺少单圆环事件主体'],
  [cardSource, '<VideoPlay />', '开始节点缺少启动图标'],
  [cardSource, '.start-node__event {', '开始节点缺少稳定的圆形尺寸'],
  [sequenceSource, 'class="branch-lane__tail"', '短分支末尾缺少连接汇聚线的伸缩尾线'],
  [sequenceSource, '.branch-lane__tail { width: 1px; min-height: 0; flex: 1;', '分支尾线不能随最长分支高度自动伸展'],
  [canvasSource, "from '@vue-flow/core'", '流程画布未接入 Vue Flow'],
  [canvasSource, '@node-drag-stop="persistNodePosition"', '流程节点拖拽后未持久化坐标'],
  [canvasSource, '@edge-update="persistEdgeAnchor"', '流程连线端点调整后未持久化连接点'],
  [canvasSource, 'ConnectionMode.Loose', '流程画布未启用四向连接点模式'],
  [canvasSource, 'connection.source !== edge.source || connection.target !== edge.target', '调整连接点时缺少流程拓扑保护'],
  [canvasSource, 'connectedHandles:', '流程节点缺少已连接端点状态'],
  [canvasSource, 'workflow: markRaw(WorkflowCurveEdge)', '流程画布未注册曲线连线'],
  [graphNodeSource, 'v-for="anchor in anchors"', '自定义流程节点缺少四向连接点'],
  [graphNodeSource, ':id="anchor.id"', '流程连接点缺少可持久化标识'],
  [graphNodeSource, "'workflow-anchor--connected': data.connectedHandles.includes(anchor.id)", '已连接端点缺少可见状态'],
  [graphNodeSource, '.workflow-anchor { opacity: 0;', '未连接端点默认应隐藏'],
  [graphNodeSource, "'workflow-graph-node--gateway': data.node.gatewayMode", '网关节点缺少独立的菱形连接点容器'],
  [graphNodeSource, 'class="gateway-node__content"', '网关名称和类型标识应显示在菱形内部'],
  [graphNodeSource, '.workflow-graph-node--gateway { width: 136px; height: 136px;', '网关外接框尺寸不足以承载放大的菱形'],
  [graphNodeSource, '.gateway-node__main { display: grid; width: 96px; height: 96px;', '网关菱形框未按设计放大'],
  [curveEdgeSource, 'getSimpleBezierPath', '流程连线未使用贝塞尔曲线'],
  [curveEdgeSource, '<FlowInsertButton', '曲线连线未保留插入节点入口'],
  [layoutSource, "from '@dagrejs/dagre'", '旧流程缺少自动布局能力'],
  [layoutSource, "if (node.type === 'start') return { width: 120, height: 76 }", '自动布局未同步开始事件尺寸'],
  [layoutSource, "if (node.gatewayMode) return { width: 136, height: 136 }", '自动布局未同步网关菱形尺寸'],
]) {
  assert(source.includes(snippet), message)
}

assert(!inspectorSource.includes('发起权限'), '发起权限应移动到发布流程弹窗')
assert(!inspectorSource.includes('updateInitiatorScope'), '节点配置不应继续修改发布期发起范围')

const branchDraft = {
  key: 'tree_test',
  name: '树形流程',
  nodes: [
    { id: 'start', type: 'start', name: '发起人' },
    { id: 'split', type: 'exclusive', name: '条件分支', gatewayMode: 'split' },
    { id: 'approval_a1', type: 'approval', name: '分支一审批' },
    { id: 'approval_a2', type: 'approval', name: '分支一复核' },
    { id: 'approval_b', type: 'approval', name: '分支二审批' },
    { id: 'join', type: 'exclusive', name: '条件汇聚', gatewayMode: 'join' },
    { id: 'after', type: 'approval', name: '汇聚后审批' },
    { id: 'end', type: 'end', name: '流程结束' },
  ],
  edges: [
    { id: 'e1', source: 'start', target: 'split' },
    { id: 'branch_a', source: 'split', target: 'approval_a1', name: '条件一' },
    { id: 'a1_a2', source: 'approval_a1', target: 'approval_a2' },
    { id: 'a2_join', source: 'approval_a2', target: 'join' },
    { id: 'branch_b', source: 'split', target: 'approval_b', name: '默认条件', default: true },
    { id: 'b_join', source: 'approval_b', target: 'join' },
    { id: 'join_after', source: 'join', target: 'after' },
    { id: 'after_end', source: 'after', target: 'end' },
  ],
}

assert(graph.findPairedJoin(branchDraft, 'split')?.id === 'join', '长分支应找到共同汇聚节点')
const workflowTree = tree.buildWorkflowTree(branchDraft)
assert(workflowTree.items[0]?.kind === 'node' && workflowTree.items[0].node.id === 'start', '树应从发起人开始')
const branchItem = workflowTree.items.find(item => item.kind === 'branch')
assert(branchItem?.kind === 'branch', '树中应包含分支组')
assert(branchItem.branches.length === 2, '分支组应保留两条路径')
assert(branchItem.branches[0].sequence.items.length === 2, '长分支应完整保留两个审批节点')
assert(workflowTree.items.some(item => item.kind === 'node' && item.node.id === 'after'), '汇聚后的主流程节点应继续渲染')
assert(workflowTree.items.at(-1)?.kind === 'node' && workflowTree.items.at(-1).node.id === 'end', '树应以结束节点收尾')

const positions = layout.layoutWorkflow(branchDraft)
assert(positions.size === branchDraft.nodes.length, '自动布局应为每个流程节点生成坐标')
assert(positions.get('start').y < positions.get('split').y, '自动布局应按流程方向从上到下排列')
assert(positions.get('approval_a1').x !== positions.get('approval_b').x, '分支节点不应重叠在同一横向位置')
assert(positions.get('join').y > positions.get('approval_a2').y, '汇聚节点应位于最长分支节点之后')

const shuffledPermissionDraft = {
  ...structuredClone(branchDraft),
  nodes: [
    branchDraft.nodes.find(node => node.id === 'after'),
    branchDraft.nodes.find(node => node.id === 'approval_b'),
    branchDraft.nodes.find(node => node.id === 'start'),
    branchDraft.nodes.find(node => node.id === 'approval_a2'),
    branchDraft.nodes.find(node => node.id === 'join'),
    branchDraft.nodes.find(node => node.id === 'approval_a1'),
    branchDraft.nodes.find(node => node.id === 'split'),
    branchDraft.nodes.find(node => node.id === 'end'),
  ].filter(Boolean),
}
assert(
  tree.workflowPermissionNodes(shuffledPermissionDraft).map(node => node.id).join(',') === 'start,approval_a1,approval_a2,approval_b,after',
  '字段权限节点应按主线和分支的流程顺序排列，而不是按 nodes 数组顺序排列',
)

const removableBranchDraft = structuredClone(branchDraft)
const thirdBranch = graph.addBranch(removableBranchDraft, 'split')
assert(thirdBranch, '应能新增第三条分支')
assert(graph.removeBranch(removableBranchDraft, 'split', 'branch_a'), '应能删除指定分支')
assert(!removableBranchDraft.nodes.some(node => ['approval_a1', 'approval_a2'].includes(node.id)), '删除分支应移除整条分支上的节点')
assert(removableBranchDraft.edges.filter(edge => edge.source === 'split').length === 2, '删除后应保留另外两条分支')
assert(removableBranchDraft.nodes.some(node => node.id === thirdBranch.id), '删除分支不应影响其他分支')

console.log('workflow tree checks passed')
