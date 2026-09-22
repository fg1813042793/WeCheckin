import { readFileSync } from 'node:fs'

const source = readFileSync(
  new URL('../src/pages/workflow/components/WorkflowDetailPanel.vue', import.meta.url),
  'utf8',
)

for (const required of [
  'dingtalk_h5:button:workflow:approve',
  'dingtalk_h5:api:workflow:approve',
  'dingtalk_h5:button:workflow:reject',
  'dingtalk_h5:api:workflow:reject',
  'dingtalk_h5:button:workflow:return',
  'dingtalk_h5:api:workflow:return',
  'dingtalk_h5:button:workflow:submit',
  'dingtalk_h5:api:workflow:submit',
  'const canApprove = computed',
  'const canReject = computed',
  'const canReturn = computed',
  'const canSubmit = computed',
  'const canEditTaskForm = computed',
  'const hasTaskAction = computed',
]) {
  if (!source.includes(required))
    throw new Error(`workflow action permission wiring missing: ${required}`)
}

if (source.includes('dingtalk_h5:api:workflow:handle'))
  throw new Error('workflow detail must not retain the aggregate handle permission')

console.log('workflow action permission checks passed')
