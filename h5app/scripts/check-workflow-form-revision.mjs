import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const read = file => fs.readFileSync(path.join(root, file), 'utf8')

const expectations = [
  ['src/types/workflow.ts', ['WorkflowFormRevisionCapability', 'WorkflowReviseFormRequest', 'formRevision: WorkflowFormRevisionCapability']],
  ['src/api/workflow.ts', ['reviseWorkflowInstanceForm', '/form-data', 'patch<WorkflowMutationResult>']],
  ['src/pages/workflow/workflow-route-keys.ts', ['workflowFormRevisionContentKey', 'workflowFormRevisionInstanceIdFromContentKey']],
  ['src/pages/workflow/workflow.routes.ts', ['WorkflowFormRevisionPage', 'workflowFormRevisionInstanceIdFromContentKey']],
  ['src/pages/workflow/components/WorkflowCenter.vue', [':form-revision-action="activeTab === \'handled\'"']],
  ['src/pages/workflow/components/WorkflowDetailPanel.vue', ['formRevisionAction', 'dingtalk_h5:button:workflow:form-revise', 'workflowFormRevisionContentKey', '修改表单', 'instance_form_revised: \'表单已修改\'']],
  ['src/pages/workflow/components/WorkflowFormRevisionPage.vue', ['WorkflowRuntimeForm', 'WorkflowParticipantSelect', 'reviseWorkflowInstanceForm', '展示你已办理节点可见的字段，仅可修改配置为可写的字段', '修改原因', '站内信', '钉钉 OA', 'registerTabCloseGuard', 'hasUnsavedChanges: () => hasUnsavedChanges.value']],
  ['src/stores/appContent.ts', ['saveDraft?: () => Promise<boolean>', 'canSaveTabDraft']],
  ['src/components/app-shell/app-shell.vue', ['pendingCloseCanSaveDraft', 'v-if="pendingCloseCanSaveDraft"']],
  ['src/components/app-notification-panel/app-notification-panel.vue', ['instance_form_revised: { label: \'表单修改\'']],
]

for (const [file, snippets] of expectations) {
  const source = read(file)
  for (const snippet of snippets) {
    if (!source.includes(snippet))
      throw new Error(`${file} missing workflow form revision contract: ${snippet}`)
  }
}

const revisionPage = read('src/pages/workflow/components/WorkflowFormRevisionPage.vue')
if (!/const fieldAccess = computed[\s\S]*?workflowFieldAccessMap\([\s\S]*?'hidden',[\s\S]*?\n\)/.test(revisionPage))
  throw new Error('workflow form revision must hide fields without explicit write access')
if (!revisionPage.includes('class="workflow-form-revision-page__content"'))
  throw new Error('workflow form revision content must use a centered container')
if (!/\.workflow-form-revision-page__body\s*\{[\s\S]*?width:\s*min\(var\(--app-pc-content-max-width, 1080px\), 100%\);[\s\S]*?margin:\s*0 auto;/.test(revisionPage))
  throw new Error('workflow form revision body must reuse the centered approval page width')
if (!/\.workflow-form-revision-page__content\s*\{[\s\S]*?padding:\s*20px;[\s\S]*?border:\s*1px solid #dfe5ee;[\s\S]*?background:\s*#fff;/.test(revisionPage))
  throw new Error('workflow form revision content must reuse the approval form card')

const detailPanel = read('src/pages/workflow/components/WorkflowDetailPanel.vue')
if (!/\.workflow-detail-panel__application-actions\s*\{[\s\S]*?grid-template-columns:\s*repeat\(auto-fit, 120px\);[\s\S]*?justify-content:\s*start;/.test(detailPanel))
  throw new Error('workflow detail application actions must use fixed-width left-aligned columns')
if (!/\.workflow-detail-panel__application-actions--comment-only\s*\{[\s\S]*?grid-template-columns:\s*120px;[\s\S]*?justify-content:\s*start;/.test(detailPanel))
  throw new Error('workflow detail comment-only action must remain fixed-width and left-aligned')
if (!/:deep\(\.workflow-detail-panel__application-action\)\s*\{[\s\S]*?width:\s*100%;[\s\S]*?height:\s*40px;/.test(detailPanel))
  throw new Error('workflow detail application actions must keep a fixed 120x40px size')
if (!detailPanel.includes(`const commentPopupMode = computed(() => mobileInteractionDialog.value ? 'bottom' : 'center')`))
  throw new Error('workflow comment form must use a bottom popup on mobile')
if (!detailPanel.includes(':mode="commentPopupMode"') || !detailPanel.includes(':width="commentPopupWidth"') || !detailPanel.includes(':safe-area-inset-bottom="true"'))
  throw new Error('workflow comment popup must expose responsive mode, width, and safe area')
if (!/\.workflow-interaction-dialog--comment\s*\{[\s\S]*?max-height:\s*92vh;[\s\S]*?overflow-y:\s*auto;/.test(detailPanel))
  throw new Error('workflow comment form must scroll within the mobile viewport')
if (!/\.workflow-interaction-dialog--comment[\s\S]*?\.workflow-interaction-dialog__actions\s*\{[\s\S]*?position:\s*sticky;[\s\S]*?bottom:\s*0;/.test(detailPanel))
  throw new Error('workflow comment actions must remain visible on mobile')
if (!detailPanel.includes(`const rejectPopupMode = computed(() => mobileInteractionDialog.value ? 'bottom' : 'center')`))
  throw new Error('workflow reject form must use a bottom popup on mobile')
if (!/v-model="rejectVisible"[\s\S]*?:mode="rejectPopupMode"[\s\S]*?:width="rejectPopupWidth"[\s\S]*?:safe-area-inset-bottom="true"[\s\S]*?workflow-interaction-dialog--reject/.test(detailPanel))
  throw new Error('workflow reject popup must expose responsive mode, width, safe area, and mobile styling')
if (!/\.workflow-interaction-dialog--reject\s*\{[\s\S]*?max-height:\s*92vh;[\s\S]*?overflow-y:\s*auto;/.test(detailPanel))
  throw new Error('workflow reject form must scroll within the mobile viewport')
if (!/\.workflow-interaction-dialog--reject[\s\S]*?\.workflow-interaction-dialog__actions\s*\{[\s\S]*?position:\s*sticky;[\s\S]*?bottom:\s*0;/.test(detailPanel))
  throw new Error('workflow reject actions must remain visible on mobile')
if (!detailPanel.includes(`const returnPopupMode = computed(() => mobileInteractionDialog.value ? 'bottom' : 'center')`))
  throw new Error('workflow return form must use a bottom popup on mobile')
if (!/v-model="returnVisible"[\s\S]*?:mode="returnPopupMode"[\s\S]*?:width="returnPopupWidth"[\s\S]*?:safe-area-inset-bottom="true"[\s\S]*?workflow-interaction-dialog--return/.test(detailPanel))
  throw new Error('workflow return popup must expose responsive mode, width, safe area, and mobile styling')
if (!/\.workflow-interaction-dialog--return\s*\{[\s\S]*?max-height:\s*92vh;[\s\S]*?overflow-y:\s*auto;/.test(detailPanel))
  throw new Error('workflow return form must scroll within the mobile viewport')
if (!/\.workflow-interaction-dialog--return[\s\S]*?\.workflow-interaction-dialog__actions\s*\{[\s\S]*?position:\s*sticky;[\s\S]*?bottom:\s*0;/.test(detailPanel))
  throw new Error('workflow return actions must remain visible on mobile')
if (!/@media screen and \(max-width: 768px\)[\s\S]*?\.workflow-detail-panel__actions\s+:deep\(\.workflow-detail-panel__action\)\s*\{[\s\S]*?height:\s*46px;[\s\S]*?padding:\s*0 8px;[\s\S]*?white-space:\s*nowrap;/.test(detailPanel))
  throw new Error('workflow mobile task actions must keep consistent button sizing without wrapping')
if (!/\.workflow-detail-panel__actions\s+:deep\(\.workflow-detail-panel__action text\)\s*\{[\s\S]*?flex-shrink:\s*0;[\s\S]*?white-space:\s*nowrap;/.test(detailPanel))
  throw new Error('workflow mobile comment label must remain on one line')
if (!detailPanel.includes('workflow-detail-panel__action--comment') || !/\.workflow-detail-panel__actions\s+:deep\(\.workflow-detail-panel__action--comment \.u-icon\)\s*\{[\s\S]*?display:\s*none;/.test(detailPanel))
  throw new Error('workflow mobile comment action must match the text-only task action format')
if (!/const showTaskActionBar = computed\(\(\) => \{[\s\S]*?pagePresentation\.value && mobileInteractionDialog\.value && activeSection\.value === 'graph'[\s\S]*?return false[\s\S]*?return pagePresentation\.value \|\| canHandle\.value \|\| canWithdraw\.value \|\| showCommentAction\.value/.test(detailPanel))
  throw new Error('workflow graph must hide task actions only in mobile page presentation')
if (!detailPanel.includes('v-if="showTaskActionBar"'))
  throw new Error('workflow task action bar must use the responsive visibility condition')

console.log('workflow form revision checks passed')
