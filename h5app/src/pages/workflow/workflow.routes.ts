import type { Component } from 'vue'
import WorkflowCenter from './components/WorkflowCenter.vue'
import WorkflowInstancePage from './components/WorkflowInstancePage.vue'
import WorkflowStartPage from './components/WorkflowStartPage.vue'
import WorkflowTaskPage from './components/WorkflowTaskPage.vue'
import {
  workflowDefinitionIdFromContentKey,
  workflowInstanceIdFromContentKey,
  workflowTaskIdFromContentKey,
  workflowTaskInstanceIdFromContentKey,
} from './workflow-route-keys'

export {
  workflowDefinitionIdFromContentKey,
  workflowInstanceContentKey,
  workflowInstanceIdFromContentKey,
  workflowStartContentKey,
  workflowTaskContentKey,
  workflowTaskIdFromContentKey,
  workflowTaskInstanceIdFromContentKey,
} from './workflow-route-keys'

export const workflowContentRoutes: Record<string, Component> = {
  workflow: WorkflowCenter,
}

export function resolveWorkflowContentComponent(key: string): Component | undefined {
  if (workflowDefinitionIdFromContentKey(key) > 0)
    return WorkflowStartPage
  if (workflowInstanceIdFromContentKey(key))
    return WorkflowInstancePage
  if (workflowTaskIdFromContentKey(key) && workflowTaskInstanceIdFromContentKey(key))
    return WorkflowTaskPage
  return workflowContentRoutes[key]
}
