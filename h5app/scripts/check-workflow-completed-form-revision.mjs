import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')

function requireSnippets(file, snippets) {
  const source = fs.readFileSync(path.join(root, file), 'utf8')
  for (const snippet of snippets) {
    if (!source.includes(snippet))
      throw new Error(`${file} missing completed form revision contract: ${snippet}`)
  }
}

const expectations = [
  ['src/types/workflow.ts', [
    'WorkflowCompletedRevisionNode',
    'WorkflowFormRevisionDetail',
    'taskType: \'workflow\' | \'form_revision\'',
  ]],
  ['src/api/workflow.ts', [
    'listWorkflowFormRevisions',
    'previewWorkflowFormRevision',
    'createWorkflowFormRevision',
    'getWorkflowFormRevision',
    'cancelWorkflowFormRevision',
    'completeWorkflowFormRevisionTask',
  ]],
  ['src/pages/workflow/workflow-route-keys.ts', [
    'workflowCompletedFormRevisionContentKey',
    'workflowFormRevisionDetailContentKey',
  ]],
  ['src/pages/workflow/workflow.routes.ts', [
    'WorkflowCompletedFormRevisionPage',
    'WorkflowFormRevisionDetailPage',
  ]],
]

for (const [file, snippets] of expectations)
  requireSnippets(file, snippets)

requireSnippets('src/pages/workflow/components/WorkflowCompletedFormRevisionPage.vue', [
  '选择来源节点',
  '修改字段',
  '修改原因',
  '提交后立即生效',
  '需要重新确认',
])
requireSnippets('src/pages/workflow/components/WorkflowFormRevisionDetailPage.vue', [
  '修改前',
  '修改后',
  '确认记录',
  '取消修订',
  '通过',
  '驳回',
])

console.log('workflow completed form revision checks passed')
