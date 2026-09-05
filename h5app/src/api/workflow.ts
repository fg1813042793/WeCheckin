import type {
  WorkflowAttachment,
  WorkflowCommentRequest,
  WorkflowCompleteTaskRequest,
  WorkflowInstanceDetail,
  WorkflowInstanceList,
  WorkflowInstanceQuery,
  WorkflowMutationResult,
  WorkflowOptionSource,
  WorkflowOverview,
  WorkflowPublishedDefinition,
  WorkflowRemindInstanceRequest,
  WorkflowRemindInstanceResult,
  WorkflowReviseFormRequest,
  WorkflowSaveStartDraftRequest,
  WorkflowStartDraft,
  WorkflowStartRequest,
  WorkflowSummaryExportFormat,
  WorkflowSummaryQuery,
  WorkflowTaskList,
  WorkflowTaskQuery,
} from '@/types/workflow'
import { authToken, buildApiUrl, del, get, patch, post, put, uploadFile } from '@/api/dingtalk-h5/base'

const WORKFLOW_API = '/api/v2/dingtalk/h5/workflows'

export function listWorkflowCategories() {
  return get<string[]>(`${WORKFLOW_API}/categories`)
}

export function getWorkflowOverview() {
  return get<WorkflowOverview>(`${WORKFLOW_API}/overview`)
}

export function listWorkflowDefinitions() {
  return get<WorkflowPublishedDefinition[]>(`${WORKFLOW_API}/definitions`)
}

export function getWorkflowDefinition(id: number) {
  return get<WorkflowPublishedDefinition>(`${WORKFLOW_API}/definitions/${id}`)
}

export function uploadWorkflowAttachment(filePath: string) {
  return uploadFile<WorkflowAttachment>(`${WORKFLOW_API}/attachments`, filePath)
}

export function startWorkflowInstance(data: WorkflowStartRequest) {
  return post<WorkflowMutationResult>(`${WORKFLOW_API}/instances`, data)
}

export function getWorkflowStartDraft(definitionId: number) {
  return get<WorkflowStartDraft | null>(`${WORKFLOW_API}/drafts/${definitionId}`)
}

export function saveWorkflowStartDraft(definitionId: number, data: WorkflowSaveStartDraftRequest) {
  return put<WorkflowStartDraft>(`${WORKFLOW_API}/drafts/${definitionId}`, data)
}

export function deleteWorkflowStartDraft(definitionId: number) {
  return del<void>(`${WORKFLOW_API}/drafts/${definitionId}`)
}

export function listWorkflowInstances(query: WorkflowInstanceQuery = {}) {
  return get<WorkflowInstanceList>(`${WORKFLOW_API}/instances`, query)
}

export function getWorkflowInstance(id: string) {
  return get<WorkflowInstanceDetail>(`${WORKFLOW_API}/instances/${encodeURIComponent(id)}`)
}

export function listWorkflowSummaryDefinitions() {
  return get<WorkflowPublishedDefinition[]>(`${WORKFLOW_API}/summary/definitions`)
}

export function getWorkflowSummaryDefinition(id: number) {
  return get<WorkflowPublishedDefinition>(`${WORKFLOW_API}/summary/definitions/${id}`)
}

export function listWorkflowSummaryInstances(query: WorkflowSummaryQuery) {
  return get<WorkflowInstanceList>(`${WORKFLOW_API}/summary/instances`, query)
}

export function getWorkflowSummaryInstance(id: string) {
  return get<WorkflowInstanceDetail>(`${WORKFLOW_API}/summary/instances/${encodeURIComponent(id)}`)
}

export function workflowSummaryExportUrl(
  definitionId: number,
  instanceIds: string[],
  format: WorkflowSummaryExportFormat,
) {
  return buildApiUrl(`${WORKFLOW_API}/summary/export`, {
    definitionId,
    instanceIds: instanceIds.join(','),
    format,
    token: authToken(),
  })
}

export function deleteWorkflowInstance(id: string) {
  return del<void>(`${WORKFLOW_API}/instances/${encodeURIComponent(id)}`)
}

export function commentWorkflowInstance(id: string, data: WorkflowCommentRequest) {
  return post<void>(
    `${WORKFLOW_API}/instances/${encodeURIComponent(id)}/comments`,
    data,
  )
}

export function remindWorkflowInstance(id: string, data: WorkflowRemindInstanceRequest) {
  return post<WorkflowRemindInstanceResult>(
    `${WORKFLOW_API}/instances/${encodeURIComponent(id)}/reminders`,
    data,
  )
}

export function reviseWorkflowInstanceForm(id: string, data: WorkflowReviseFormRequest) {
  return patch<WorkflowMutationResult>(
    `${WORKFLOW_API}/instances/${encodeURIComponent(id)}/form-data`,
    data,
  )
}

export function withdrawWorkflowInstance(id: string, reason = '') {
  return post<WorkflowMutationResult>(
    `${WORKFLOW_API}/instances/${encodeURIComponent(id)}/withdraw`,
    { reason },
  )
}

export function listWorkflowTasks(query: WorkflowTaskQuery = {}) {
  return get<WorkflowTaskList>(`${WORKFLOW_API}/tasks`, query)
}

export function completeWorkflowTask(id: string, data: WorkflowCompleteTaskRequest) {
  return post<WorkflowMutationResult>(
    `${WORKFLOW_API}/tasks/${encodeURIComponent(id)}/complete`,
    data,
  )
}

export function requestWorkflowOptionSource(source: WorkflowOptionSource) {
  const url = String(source.url || '').trim()
  if (String(source.method || '').toUpperCase() === 'POST') {
    return post<unknown>(url, {})
  }
  return get<unknown>(url)
}
