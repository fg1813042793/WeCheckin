import request from '../utils/request'
import type { PageQuery } from './types'
import type {
  DingTalkSendTemplate, DingTalkSendTemplateInput, DingTalkSendTemplateUpdate,
  SendTemplateList, SendTemplateOptionQuery, TemplateWorkflowOptions, TemplateWorkflowQuery,
} from '../types/notificationTemplate'

const base = '/api/v2/admin/dingtalk-notification-templates'
const jsonConfig = {
  headers: { 'Content-Type': 'application/json' },
  transformRequest: [(data: unknown) => JSON.stringify(data)],
}

export const notificationTemplatesApi = {
  list(params: PageQuery) { return request.get<SendTemplateList>(base, { params }) },
  options(params: SendTemplateOptionQuery) { return request.get<SendTemplateList>(`${base}/options`, { params }) },
  workflowOptions(params: TemplateWorkflowQuery) {
    return request.get<TemplateWorkflowOptions>(`${base}/workflow-options`, { params })
  },
  create(data: DingTalkSendTemplateInput) {
    return request.post<DingTalkSendTemplate, DingTalkSendTemplateInput>(base, data, jsonConfig)
  },
  update(id: number, data: DingTalkSendTemplateUpdate) {
    return request.put<DingTalkSendTemplate, DingTalkSendTemplateUpdate>(`${base}/${id}`, data, jsonConfig)
  },
  remove(id: number, version: number) {
    return request.delete(`${base}/${id}`, { params: { version } })
  },
}
