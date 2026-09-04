import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const source = readFileSync(
  new URL('../src/pages/workflow/components/WorkflowDetailPanel.vue', import.meta.url),
  'utf8',
)

assert.match(
  source,
  /\.workflow-detail-panel--history-drawer \.workflow-detail-panel__application-actions--comment-only\s*\{[\s\S]*?min-height:\s*56px;[\s\S]*?grid-template-columns:\s*104px;/,
)
assert.match(
  source,
  /\.workflow-detail-panel--history-drawer[\s\S]*?application-actions--comment-only\s*:deep\(\.workflow-detail-panel__application-action\)[\s\S]*?height:\s*36px;[\s\S]*?border-color:\s*#0f766e;/,
)
assert.match(
  source,
  /@media[^{]*\(max-width:\s*768px\)[\s\S]*?\.workflow-detail-panel--history-drawer \.workflow-detail-panel__application-actions--comment-only\s*\{[\s\S]*?grid-template-columns:\s*1fr;/,
)

console.log('workflow history action layout checks passed')
