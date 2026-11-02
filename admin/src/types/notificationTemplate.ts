import type { PageQuery } from '../api/types'

export type SendTemplateMessageType = 'oa' | 'link' | 'action_card'
export type SendTemplateTargetKind = 'app_home' | 'workflow_start' | 'url'

export interface SendTemplateTarget {
  kind: SendTemplateTargetKind
  definitionId?: number
  /** Response-only display name. Never included in mutation payloads. */
  definitionName?: string
  url?: string
}

export interface DingTalkSendTemplateInput {
  name: string
  enabled: boolean
  title: string
  content: string
  messageType: SendTemplateMessageType
  target: SendTemplateTarget
  buttonTitle: string
  headColor: string
}

export interface DingTalkSendTemplate extends DingTalkSendTemplateInput {
  id: number
  version: number
  updatedAt: number
}

export interface DingTalkSendTemplateUpdate extends DingTalkSendTemplateInput {
  version: number
}

export interface SendTemplateList {
  list: DingTalkSendTemplate[]
  total: number
  page: number
  pageSize: number
}

export interface TemplateWorkflowOption { id: number; name: string }
export interface TemplateWorkflowQuery extends PageQuery { id?: number }
export interface SendTemplateOptionQuery extends PageQuery { id?: number; targetKind?: SendTemplateTargetKind }
export interface TemplateWorkflowOptions {
  list: TemplateWorkflowOption[]
  total: number
  page: number
  pageSize: number
}
