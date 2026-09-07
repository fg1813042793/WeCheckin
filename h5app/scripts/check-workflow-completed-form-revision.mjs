import assert from 'node:assert/strict'
import { Buffer } from 'node:buffer'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import ts from 'typescript'

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')

function requireSnippets(file, snippets) {
  const source = fs.readFileSync(path.join(root, file), 'utf8')
  for (const snippet of snippets) {
    if (!source.includes(snippet))
      throw new Error(`${file} missing completed form revision contract: ${snippet}`)
  }
}

async function loadTypeScriptModule(file) {
  const source = fs.readFileSync(path.join(root, file), 'utf8')
  const output = ts.transpileModule(source, {
    compilerOptions: {
      module: ts.ModuleKind.ESNext,
      target: ts.ScriptTarget.ES2022,
    },
  })
  return import(`data:text/javascript;base64,${Buffer.from(output.outputText).toString('base64')}`)
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
requireSnippets('src/pages/workflow/components/WorkflowCenter.vue', [
  'task.taskType === \'form_revision\'',
  'workflowFormRevisionDetailContentKey',
])
requireSnippets('src/components/app-notification-panel/app-notification-panel.vue', [
  'sourceType === \'workflow_form_revision\'',
  'openWorkflowNotification',
])
requireSnippets('src/pages/notifications/components/NotificationHistoryPage.vue', [
  'workflowFormRevisionDetailContentKey',
  'openWorkflowNotification',
])

const workflowNotification = await loadTypeScriptModule('src/pages/notifications/workflow-notification-open.ts')
const routeDependencies = {
  workflowInstanceContentKey: sourceID => `workflow:instance:${encodeURIComponent(sourceID)}`,
  workflowFormRevisionDetailContentKey: sourceID => `workflow:form-revision-detail:${encodeURIComponent(sourceID)}`,
}
assert.deepEqual(workflowNotification.workflowNotificationDynamicTab({
  sourceType: 'workflow_form_revision',
  sourceId: 'revision/1',
  type: 'instance_form_revision_requested',
  title: '修订待确认',
}, routeDependencies), {
  key: 'workflow:form-revision-detail:revision%2F1',
  label: '修订待确认',
  icon: 'edit-pen',
  path: '/pages/index/index?view=workflow%3Aform-revision-detail%3Arevision%252F1',
})
assert.equal(workflowNotification.workflowNotificationLabel({
  type: 'instance_form_revision_conflict',
}), '修订冲突')

let opened = false
assert.equal(await workflowNotification.openWorkflowNotification({
  sourceType: 'workflow_form_revision',
  sourceId: 'revision-1',
  title: '修订待确认',
}, {
  ...routeDependencies,
  markRead: async () => false,
  openDynamicTab: () => {
    opened = true
    return true
  },
  closePanel: () => {},
}), true)
assert.equal(opened, false, 'mark-read failure must keep the notification detail open')

console.log('workflow completed form revision checks passed')
