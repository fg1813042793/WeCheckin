import type { WorkflowPublishedDefinition } from '@/types/workflow'

type WorkflowDefinitionNameSource = Pick<WorkflowPublishedDefinition, 'displayName' | 'name' | 'key'>

export function workflowDefinitionDisplayName(definition?: WorkflowDefinitionNameSource | null) {
  if (!definition)
    return '流程审批'
  return String(definition.displayName || '').trim()
    || String(definition.name || '').trim()
    || String(definition.key || '').trim()
    || '流程审批'
}
