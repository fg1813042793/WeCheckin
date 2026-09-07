import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const read = file => fs.readFileSync(path.join(root, file), 'utf8')

const typeSource = read('src/types/workflow.ts')
const inspectorSource = read('src/views/workflow/designer/components/NodeInspector.vue')
const fieldCatalogSource = read('src/views/workflow/designer/workflowFieldCatalog.ts')
const notificationPageSource = read('src/views/notification/index.vue')
const notificationStyleSource = read('src/views/notification/components/NotificationStyleDialog.vue')

for (const snippet of [
  'interface WorkflowPostHandleEditConfig',
  'postHandleEdit?: WorkflowPostHandleEditConfig',
	'WorkflowCompletedRevisionConfig',
	'completedRevision?: WorkflowCompletedRevisionConfig',
  "| 'instance_form_revised'",
]) {
  if (!typeSource.includes(snippet))
    throw new Error(`workflow type contract missing: ${snippet}`)
}

for (const snippet of [
  '办理完成后允许修改表单',
  'updatePostHandleEditEnabled',
	'流程完成后允许修订',
	'允许直接修订的低风险字段',
	'updateCompletedRevisionEnabled',
	'updateCompletedRevisionDirectFields',
  "['approval', 'handle'].includes(selectedNode.type)",
]) {
  if (!inspectorSource.includes(snippet))
    throw new Error(`workflow node inspector missing: ${snippet}`)
}

if (!fieldCatalogSource.includes('workflowCompletedRevisionBlockedFieldTypes'))
  throw new Error('workflow field catalog missing completed revision eligibility metadata')

if (!notificationPageSource.includes("{ label: '表单修改', value: 'instance_form_revised' }"))
  throw new Error('notification page missing instance form revised type')
if (!notificationStyleSource.includes("instance_form_revised: '表单修改'") || !notificationStyleSource.includes("value: 'edit-pen'"))
  throw new Error('notification style dialog missing instance form revised type')

console.log('workflow form revision checks passed')
