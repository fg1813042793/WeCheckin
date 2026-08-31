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
const tree = loadTypeScriptModule('src/views/workflow/designer/flowTree.ts', { './graph': graph })

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
assert(!insertionDraft.edges.some(edge => edge.id === 'start_to_end'), '原连线应被替换')
assert(insertionDraft.edges.some(edge => edge.source === 'start' && edge.target === inserted.id), '应连接到新节点')
assert(insertionDraft.edges.some(edge => edge.source === inserted.id && edge.target === 'end'), '新节点应连接到原目标')

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

const removableBranchDraft = structuredClone(branchDraft)
const thirdBranch = graph.addBranch(removableBranchDraft, 'split')
assert(thirdBranch, '应能新增第三条分支')
assert(graph.removeBranch(removableBranchDraft, 'split', 'branch_a'), '应能删除指定分支')
assert(!removableBranchDraft.nodes.some(node => ['approval_a1', 'approval_a2'].includes(node.id)), '删除分支应移除整条分支上的节点')
assert(removableBranchDraft.edges.filter(edge => edge.source === 'split').length === 2, '删除后应保留另外两条分支')
assert(removableBranchDraft.nodes.some(node => node.id === thirdBranch.id), '删除分支不应影响其他分支')

console.log('workflow tree checks passed')
