export type ID = string | number

export type QueryPrimitive = string | number | boolean | null | undefined
export type QueryParams = Record<string, QueryPrimitive>
export type FormPayload = Record<string, unknown>

export interface AdminListRecord {
  id?: ID
  name?: string
  title?: string
  status?: number
  [key: string]: unknown
}

export interface DepartmentNode {
  id: number
  name: string
  children?: DepartmentNode[]
  [key: string]: unknown
}

export interface PageQuery extends QueryParams {
  page?: number
  pageSize?: number
  keyword?: string
  sort?: string
}

export interface PageResult<T> {
  list: T[]
  total: number
  page?: number
  size?: number
}

export type InAppNotificationScope = 'all' | 'departments' | 'users'
export type NotificationTone = 'primary' | 'success' | 'warning' | 'danger' | 'info'
export type DingTalkMessageType = 'auto' | 'text' | 'image' | 'voice' | 'file' | 'link' | 'oa' | 'markdown' | 'action_card'

export interface DingTalkNotificationTemplate {
  messageType: DingTalkMessageType
  title: string
  content: string
  url: string
  picUrl: string
  sourceName: string
  mediaId: string
  duration: number
  buttonTitle: string
  headColor: string
}

export interface NotificationStyle {
  type: string
  label: string
  icon: string
  tone: NotificationTone
  dingTalk: DingTalkNotificationTemplate
}

export interface NotificationStyleConfig {
  version: number
  styles: NotificationStyle[]
}

export interface InAppNotificationItem {
  id: number
  title: string
  content: string
  type: string
  sourceType: string
  sourceId: string
  recipientUserId: string
  recipientName: string
  isRead: number
  addTime: number
}

export interface InAppNotificationQuery extends PageQuery {
  title?: string
  recipientName?: string
  sourceType?: string
  type?: string
  isRead?: number
  addTimeFrom?: number
  addTimeTo?: number
}

export interface InAppNotificationList {
  list: InAppNotificationItem[]
  total: number
  page: number
  pageSize: number
}

export interface InAppNotificationRecipientOptions {
  users: AdminUser[]
  departments: DepartmentNode[]
}

export interface InAppNotificationSendPayload {
  requestId: string
  title: string
  content: string
  scope: InAppNotificationScope
  userIds?: number[]
  departmentIds?: number[]
}

export interface InAppNotificationSendResult {
  sourceId: string
  plannedCount: number
  sentCount: number
  skippedCount: number
  failedCount: number
  replayed: boolean
}

export type DingTalkNotificationSendPayload = Omit<InAppNotificationSendPayload, 'requestId'>
export type DingTalkNotificationSendResult = InAppNotificationSendResult

export interface NotificationStyleTestPayload {
  requestId: string
  notificationType: string
  title: string
  content: string
  userIds: number[]
}

export interface AdminMenuItem {
  id: string | number
  name: string
  type: 0 | 1 | 2
  status: number
  path?: string
  icon?: string
  children?: AdminMenuItem[]
}

export interface AdminLoginData {
  token: string
  id?: number
  name?: string
  pic?: string
  roleId?: number
  roleIds?: number[]
}

export interface DingTalkCorpConfig {
  corpId?: string
  corpName?: string
  appKey?: string
  agentId?: string
  unifiedAppId?: string
  appUrl?: string
  notifyEnabled?: number
  notifyMode?: string
  robotCode?: string
  appSecretSet?: boolean
}

export interface DingTalkSettings extends DingTalkCorpConfig {
  corpConfigs?: DingTalkCorpConfig[]
  tokenExpire?: string
  redisPrefix?: string
  singleLogin?: number
  selfBind?: number
  appName?: string
  logoText?: string
  logoUrl?: string
}

export interface EnrollItem extends AdminListRecord {
  id?: ID
  title?: string
  img?: string
  desc?: string
  cateName?: string
  order?: number
  allowRepeat?: number
  dailyLimit?: number
  publishDeptIds?: string
  timeStart?: number
  timeEnd?: number
  forms?: string
  joinForms?: string
}

export interface EventUserReference {
  userId?: ID
  userID?: ID
  name?: string
}

export interface EventItem extends AdminListRecord {
  id?: ID
  title?: string
  type?: number
  status?: number
  img?: string
  desc?: string
  rules?: string
  order?: number
  publishDeptIds?: string
  regStart?: number
  regEnd?: number
  eventStart?: number
  eventEnd?: number
  forms?: string
  scoreFields?: string
  organizers?: EventUserReference[]
  assistants?: EventUserReference[]
  referees?: EventUserReference[]
  cateName?: string
}

export interface NewsItem extends AdminListRecord {
  id?: ID
  title?: string
  cateName?: string
  order?: number
  desc?: string
  content?: string
  img?: string
  publishDeptIds?: string
  vouch?: number
}

export interface ApplicationPermissionTree {
  client: AdminListRecord[]
  dingtalkH5: AdminListRecord[]
  clientApi: AdminListRecord[]
  dingtalkH5Api: AdminListRecord[]
}

export interface FormkitTypeMeta {
  type: string
  displayName: string
  category: string
  defaultProps: Record<string, unknown>
}

export interface FormkitQuestion {
  id: string
  type: string
  title: string
  description?: string
  required: boolean
  placeholder?: string
  props?: unknown
  validate?: unknown[]
  logic?: unknown[]
  calcValue?: unknown
}

export interface FormkitReport {
  title: string
  count: number
  stats: AdminListRecord[]
  table: {
    headers: string[]
    rows: Array<{ userId: string; addTime: number; values: unknown[] }>
  }
}

export interface AdminUser {
  id: number
  name: string
  mobile?: string
  avatar?: string
  pic?: string
  status: number
  forms?: string
  loginCnt?: number
  addTime?: number
  loginTime?: number
  deptIds?: number[]
  account?: string
  adminEnabled?: number
  adminType?: number
  roleId?: number
  roleName?: string
  roleIds?: number[]
  roleNames?: string[]
  allowPermissionKeys?: string[]
  denyPermissionKeys?: string[]
  extraDataDeptIds?: number[]
  extraDataUserIds?: number[]
  adminDeptIds?: number[]
  positionId?: number
  positionName?: string
  managerUserId?: number
  managerUserName?: string
}

export interface SurveyItem {
  id: number
  title: string
  description?: string
  category?: string
  status: number
  mode?: string
  addTime?: number
  responseCount?: number
}

export interface ExamItem {
  id: number
  title: string
  description?: string
  category?: string
  status: number
  duration?: number
  addTime?: number
}

export interface QuestionBankItem {
  id: number
  title: string
  type: string
  schema: string
  category?: string
  tags?: string
  status?: number
  deptId?: number
  createBy?: number
  addTime?: number
}

export interface ResourceItem {
  id: number
  surveyId?: number
  examId?: number
  type: string
  url: string
  filename: string
  path: string
  domain?: string
  addTime?: number
}

export interface DictTypeSummary {
  typeCode: string
  typeName: string
  status: number
  remark: string
  itemCnt: number
  addTime: number
  editTime: number
}

export interface DictItem {
  id: number
  typeCode: string
  typeName: string
  label: string
  value: string
  sort: number
  status: number
  remark: string
  addTime: number
  editTime: number
}

export interface DictTypePayload {
  typeCode: string
  typeName: string
  status: number
  remark: string
}

export interface DictItemPayload {
  typeCode: string
  label: string
  value: string
  sort: number
  status: number
  remark: string
}
