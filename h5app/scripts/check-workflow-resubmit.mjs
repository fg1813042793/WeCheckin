import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const detailPanel = readFileSync(
  new URL('../src/pages/workflow/components/WorkflowDetailPanel.vue', import.meta.url),
  'utf8',
)
const startPage = readFileSync(
  new URL('../src/pages/workflow/components/WorkflowStartPage.vue', import.meta.url),
  'utf8',
)

for (const snippet of [
  'const showResubmitAction = computed',
  'historyDrawer.value',
  'resubmitApplication.value',
  'canModifyApplication.value',
  'class="workflow-detail-panel__section-heading-actions"',
  'v-if="showResubmitAction"',
  'custom-class="workflow-detail-panel__resubmit-action"',
  '@click="modifyApplication"',
  '<text>再次提交</text>',
  'if (!resubmitApplication.value && !await confirmAction(actionTitle, content))',
  'appContent.setWorkflowStartSeed({',
  'formData: current.formData || {},',
]) {
  assert.ok(detailPanel.includes(snippet), `WorkflowDetailPanel.vue missing ${snippet}`)
}

for (const snippet of [
  'const seed = appContent.takeWorkflowStartSeed(value.id)',
  'formData.value = initialWorkflowFormData(value.form || [], seed.formData)',
  '已复制原申请内容，请修改后提交',
]) {
  assert.ok(startPage.includes(snippet), `WorkflowStartPage.vue missing ${snippet}`)
}

console.log('Workflow resubmit checks passed')
