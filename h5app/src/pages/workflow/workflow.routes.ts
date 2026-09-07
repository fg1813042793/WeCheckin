import type { Component } from 'vue'
import WorkflowCenter from './components/WorkflowCenter.vue'
import WorkflowCompletedFormRevisionPage from './components/WorkflowCompletedFormRevisionPage.vue'
import WorkflowFormDetailPage from './components/WorkflowFormDetailPage.vue'
import WorkflowFormRevisionDetailPage from './components/WorkflowFormRevisionDetailPage.vue'
import WorkflowFormRevisionPage from './components/WorkflowFormRevisionPage.vue'
import WorkflowInstancePage from './components/WorkflowInstancePage.vue'
import WorkflowStartPage from './components/WorkflowStartPage.vue'
import WorkflowTaskPage from './components/WorkflowTaskPage.vue'
import {
  workflowCompletedFormRevisionInstanceIdFromContentKey,
  workflowDefinitionIdFromContentKey,
  workflowFormDetailInstanceIdFromContentKey,
  workflowFormRevisionDetailIdFromContentKey,
  workflowFormRevisionInstanceIdFromContentKey,
  workflowInstanceIdFromContentKey,
  workflowTaskIdFromContentKey,
  workflowTaskInstanceIdFromContentKey,
} from './workflow-route-keys'

export {
  workflowCompletedFormRevisionContentKey,
  workflowCompletedFormRevisionInstanceIdFromContentKey,
  workflowDefinitionIdFromContentKey,
  workflowFormDetailContentKey,
  workflowFormDetailInstanceIdFromContentKey,
  workflowFormRevisionContentKey,
  workflowFormRevisionDetailContentKey,
  workflowFormRevisionDetailIdFromContentKey,
  workflowFormRevisionInstanceIdFromContentKey,
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
  if (workflowFormDetailInstanceIdFromContentKey(key))
    return WorkflowFormDetailPage
  if (workflowFormRevisionInstanceIdFromContentKey(key))
    return WorkflowFormRevisionPage
  if (workflowCompletedFormRevisionInstanceIdFromContentKey(key))
    return WorkflowCompletedFormRevisionPage
  if (workflowFormRevisionDetailIdFromContentKey(key))
    return WorkflowFormRevisionDetailPage
  if (workflowTaskIdFromContentKey(key) && workflowTaskInstanceIdFromContentKey(key))
    return WorkflowTaskPage
  return workflowContentRoutes[key]
}
