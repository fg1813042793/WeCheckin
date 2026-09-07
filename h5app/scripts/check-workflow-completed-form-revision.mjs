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
    "taskType: 'workflow' | 'form_revision'",
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

console.log('workflow completed form revision checks passed')
