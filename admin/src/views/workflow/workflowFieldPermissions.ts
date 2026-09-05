import type { WorkflowDraft, WorkflowFormField } from './types'

function collectCalculationFieldKeys(fields: WorkflowFormField[], target: Set<string>) {
  for (const field of fields) {
    if (field.type === 'calculation' && field.key) target.add(field.key)
    if (field.type === 'group' && Array.isArray(field.fields)) {
      collectCalculationFieldKeys(field.fields, target)
    }
  }
}

export function normalizeWorkflowCalculationPermissions(draft: WorkflowDraft): boolean {
  const calculationFieldKeys = new Set<string>()
  collectCalculationFieldKeys(Array.isArray(draft.form) ? draft.form : [], calculationFieldKeys)
  if (calculationFieldKeys.size === 0) return false

  let changed = false
  for (const node of Array.isArray(draft.nodes) ? draft.nodes : []) {
    for (const permission of Array.isArray(node.formPermissions) ? node.formPermissions : []) {
      if (!calculationFieldKeys.has(permission.field)) continue
      if (permission.access === 'write') {
        permission.access = 'read'
        changed = true
      }
      if (permission.actions !== undefined) {
        delete permission.actions
        changed = true
      }
    }
  }
  return changed
}
