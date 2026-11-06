export type PublicPortalKind = 'survey' | 'exam'
export type PublicActivityState = 'upcoming' | 'active' | 'ended'

export interface PublicPortalSettings {
  enabled: boolean
  title: string
  description: string
  logoUrl: string
  showEnded: boolean
  version: number
}

export interface PublicPortalItemInput {
  kind: PublicPortalKind
  sourceId: number
  title: string
  description: string
  enabled: boolean
  sort: number
}

export interface PublicPortalItemUpdate extends PublicPortalItemInput {
  version: number
}

export interface PublicPortalAdminItem extends PublicPortalItemUpdate {
  id: number
  managementTitle: string
  category?: string
  displayTitle: string
  available: boolean
  unavailableReason: string
  canManage: boolean
}

export interface PublicPortalSourceOption {
  kind: PublicPortalKind
  sourceId: number
  managementTitle: string
  category?: string
  displayTitle: string
  startTime: number
  endTime: number
  loginRequired: boolean
  passwordRequired: boolean
}

export interface PublicPortalOptionsQuery {
  kind: PublicPortalKind
  keyword: string
  page: number
  pageSize: number
}

export interface PublicPortalOptions {
  list: PublicPortalSourceOption[]
  total: number
  page: number
  pageSize: number
}

export interface PublicPortalConfig {
  settings: PublicPortalSettings
  items: PublicPortalAdminItem[]
}

export interface PublicPortalDeleteResponse {
  id: number
}

export interface PublicActivityItem {
  id: number
  kind: PublicPortalKind
  sourceId: number
  title: string
  managementTitle?: string
  category?: string
  description: string
  startTime: number
  endTime: number
  state: PublicActivityState
  loginRequired: boolean
  passwordRequired: boolean
  url: string
}

export interface PublicActivities {
  enabled: boolean
  title: string
  description: string
  logoUrl: string
  items: PublicActivityItem[]
}
