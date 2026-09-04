import request from '../utils/request'
import type {
  AdminUser,
  AdminLoginData,
  AdminMenuItem,
  AdminListRecord,
  ApplicationPermissionTree,
  DepartmentNode,
  DingTalkNotificationSendPayload,
  DingTalkNotificationSendResult,
  EnrollItem,
  EventItem,
  FormkitQuestion,
  FormkitReport,
  FormkitTypeMeta,
  NewsItem,
  DictItem,
  DictItemPayload,
  DictTypePayload,
  DictTypeSummary,
  ExamItem,
  FormPayload,
  ID,
  InAppNotificationList,
  InAppNotificationRecipientOptions,
  InAppNotificationSendPayload,
  InAppNotificationSendResult,
  PageQuery,
  PageResult,
  QueryParams,
  QuestionBankItem,
  ResourceItem,
  SurveyItem
} from './types'
import type {
  WorkflowDefinitionDetail,
  WorkflowDefinitionSummary,
  WorkflowInstanceDetail,
  WorkflowInstanceSummary,
  WorkflowNotificationList,
  WorkflowPublishResult,
  WorkflowPublishedDefinition,
  WorkflowTaskSummary,
  WorkflowValidationResult,
  WorkflowVersion,
  WorkflowVersionChangeSummary,
} from '../types/workflow'
import type {
  CronOccurrence,
  HandlerMetadata,
  PageData as ScheduledTaskPageData,
  ScheduledTask,
  ScheduledTaskRun,
  ScheduledTaskRunDetail,
  ScheduledTaskWorker,
} from '../types/scheduledTask'

export type TemplatePreset = { label: string; value: string }
type WorkflowInitiatorConfig = { scope: 'all' | 'specified'; userIds?: number[]; departmentIds?: number[]; excludedUserIds?: number[] }

const API_V2 = '/api/v2'
const ADMIN_V2 = `${API_V2}/admin`

const jsonConfig = {
  transformRequest: [(data: unknown) => (typeof data === 'string' ? data : JSON.stringify(data))],
  headers: { 'Content-Type': 'application/json' }
}

function encodePath(value: ID | null | undefined, fallback: ID = '') {
  return encodeURIComponent(String(value ?? fallback))
}

function idFrom(data: { id?: ID } | ID | null | undefined, fallback: ID = '') {
  if (data && typeof data === 'object') return data.id ?? fallback
  return data ?? fallback
}

function deleteBody(data: unknown) {
  return { data }
}

function vouchValue(data: { vouch?: number | string; status?: number | string; isVouch?: number | string }) {
  return data.vouch ?? data.isVouch ?? data.status ?? 0
}

export const adminApi = {
  login(data: { name: string; password: string }) {
    return request.post<AdminLoginData, { name: string; password: string }>(`${ADMIN_V2}/auth/login`, data)
  },
  home() {
    return request.get(`${ADMIN_V2}/home`)
  },
  clearVouch() {
    return request.delete(`${ADMIN_V2}/home/recommendations`)
  },
  // 用户管理
  userList(params?: PageQuery) {
    return request.get<PageResult<AdminUser>>(`${ADMIN_V2}/users`, { params })
  },
  userDetailById(id: ID) {
    return request.get<AdminUser>(`${ADMIN_V2}/users/${encodePath(id)}`)
  },
  userAdd(data: FormPayload) {
    return request.post(`${ADMIN_V2}/users`, data)
  },
  userEdit(data: FormPayload & { id?: ID }) {
    return request.put(`${ADMIN_V2}/users/${encodePath(idFrom(data))}`, data)
  },
  userStatus(data: { id: ID; status: number | string; reason?: string }) {
    return request.patch(`${ADMIN_V2}/users/${encodePath(data.id)}/status`, data)
  },
  userDel(data: { id: ID }) {
    return request.delete(`${ADMIN_V2}/users/${encodePath(data.id)}`, deleteBody(data))
  },
  userDels(data: { ids: string }) {
    return request.delete(`${ADMIN_V2}/users`, deleteBody(data))
  },
  userResetPwd(data: { id: ID }) {
    return request.patch(`${ADMIN_V2}/users/${encodePath(data.id)}/password`, data)
  },
  userFormFields() {
    return request.get<AdminListRecord[]>(`${API_V2}/user-form-fields`)
  },
  userFormFieldSave(data: FormPayload) {
    return request.put(`${ADMIN_V2}/users/form-fields`, data)
  },
  // 打卡管理
  enrollList(params?: PageQuery) {
    return request.get<PageResult<EnrollItem>>(`${ADMIN_V2}/enrollments`, { params })
  },
  enrollDetail(id: ID) {
    return request.get<EnrollItem>(`${ADMIN_V2}/enrollments/${encodePath(id)}`)
  },
  enrollInsert(data: FormPayload) {
    return request.post(`${ADMIN_V2}/enrollments`, data)
  },
  enrollEdit(data: FormPayload & { id?: ID }) {
    return request.put(`${ADMIN_V2}/enrollments/${encodePath(idFrom(data))}`, data)
  },
  enrollDel(data: { id: ID }) {
    return request.delete(`${ADMIN_V2}/enrollments/${encodePath(data.id)}`, deleteBody(data))
  },
  enrollDels(data: { ids: string }) {
    return request.delete(`${ADMIN_V2}/enrollments`, deleteBody(data))
  },
  enrollStatus(data: { id: ID; status: number | string }) {
    return request.patch(`${ADMIN_V2}/enrollments/${encodePath(data.id)}/status`, data)
  },
  enrollSort(data: { id: ID; sort: number | string }) {
    return request.patch(`${ADMIN_V2}/enrollments/${encodePath(data.id)}/sort`, data)
  },
  enrollVouch(data: { id: ID; vouch?: number | string; status?: number | string; isVouch?: number | string }) {
    return request.patch(`${ADMIN_V2}/enrollments/${encodePath(data.id)}/recommendation`, { ...data, vouch: vouchValue(data) })
  },
  enrollClear(data: { id?: ID } = {}) {
    return request.post(`${ADMIN_V2}/enrollments/${encodePath(data.id, 0)}/clear`, data)
  },
  enrollJoinList(params?: PageQuery & { enrollId?: ID }) {
    return request.get<PageResult<AdminListRecord>>(`${ADMIN_V2}/enrollments/${encodePath(params?.enrollId, 0)}/joins`, { params })
  },
  enrollUserList(params?: PageQuery & { enrollId?: ID }) {
    return request.get<AdminListRecord[]>(`${ADMIN_V2}/enrollments/${encodePath(params?.enrollId, 0)}/users`, { params })
  },
  enrollStats(params?: QueryParams & { enrollId?: ID; id?: ID }) {
    return request.get<AdminListRecord[]>(`${ADMIN_V2}/enrollments/${encodePath(params?.enrollId ?? params?.id, 0)}/stats`, { params })
  },
  enrollRemoveUser(data: { id?: ID; userId?: ID; enrollId?: ID }) {
    return request.delete(`${ADMIN_V2}/enrollments/${encodePath(data.enrollId ?? data.id, 0)}/users/${encodePath(data.userId)}`, deleteBody(data))
  },
  enrollRemoveUsers(data: { ids?: string; userIds?: string; enrollId?: ID }) {
    return request.delete(`${ADMIN_V2}/enrollments/${encodePath(data.enrollId, 0)}/users`, deleteBody(data))
  },
  enrollJoinDel(data: { id?: ID; enrollJoinId?: ID; enrollId?: ID }) {
    return request.delete(`${ADMIN_V2}/enrollments/${encodePath(data.enrollId, 0)}/joins/${encodePath(data.enrollJoinId ?? data.id)}`, deleteBody(data))
  },
  enrollJoinDels(data: { ids: string; enrollId?: ID }) {
    return request.delete(`${ADMIN_V2}/enrollments/${encodePath(data.enrollId, 0)}/joins`, deleteBody(data))
  },
  enrollJoinDataExport(params?: QueryParams & { enrollId?: ID }) {
    return request.post<string, null>(`${ADMIN_V2}/enrollments/${encodePath(params?.enrollId, 0)}/export`, null, { params })
  },
  // 内容管理
  newsList(params?: PageQuery) {
    return request.get<PageResult<NewsItem>>(`${ADMIN_V2}/news`, { params })
  },
  newsDetail(id: ID) {
    return request.get<NewsItem>(`${ADMIN_V2}/news/${encodePath(id)}`)
  },
  newsInsert(data: FormPayload) {
    return request.post(`${ADMIN_V2}/news`, data)
  },
  newsEdit(data: FormPayload & { id?: ID }) {
    return request.put(`${ADMIN_V2}/news/${encodePath(idFrom(data))}`, data)
  },
  newsDel(data: { id: ID }) {
    return request.delete(`${ADMIN_V2}/news/${encodePath(data.id)}`, deleteBody(data))
  },
  newsDels(data: { ids: string }) {
    return request.delete(`${ADMIN_V2}/news`, deleteBody(data))
  },
  newsStatus(data: { id: ID; status: number | string }) {
    return request.patch(`${ADMIN_V2}/news/${encodePath(data.id)}/status`, data)
  },
  newsVouch(data: { id: ID; vouch?: number | string; status?: number | string; isVouch?: number | string }) {
    return request.patch(`${ADMIN_V2}/news/${encodePath(data.id)}/recommendation`, { ...data, vouch: vouchValue(data) })
  },
  newsSort(data: { id: ID; sort: number | string }) {
    return request.patch(`${ADMIN_V2}/news/${encodePath(data.id)}/sort`, data)
  },
  // 管理员管理
  mgrList(params?: PageQuery) {
    return request.get<PageResult<AdminListRecord>>(`${ADMIN_V2}/managers`, { params })
  },
  mgrDetail(id: ID) {
    return request.get<AdminListRecord>(`${ADMIN_V2}/managers/${encodePath(id)}`)
  },
  mgrInsert(data: FormPayload) {
    return request.post(`${ADMIN_V2}/managers`, data)
  },
  mgrEdit(data: FormPayload & { id?: ID }) {
    return request.put(`${ADMIN_V2}/managers/${encodePath(idFrom(data))}`, data)
  },
  mgrDel(data: { id: ID }) {
    return request.delete(`${ADMIN_V2}/managers/${encodePath(data.id)}`, deleteBody(data))
  },
  mgrDels(data: { ids: string }) {
    return request.delete(`${ADMIN_V2}/managers`, deleteBody(data))
  },
  mgrStatus(data: { id: ID; status: number }) {
    return request.patch(`${ADMIN_V2}/managers/${encodePath(data.id)}/status`, data)
  },
  mgrPwd(data: { id?: ID; password?: string; oldPassword?: string; newPassword?: string }) {
    if (data.id) {
      return request.patch(`${ADMIN_V2}/managers/${encodePath(data.id)}/password`, data)
    }
    return request.patch(`${ADMIN_V2}/me/password`, data)
  },
  // 操作日志
  logList(params?: PageQuery) {
    return request.get<PageResult<AdminListRecord>>(`${ADMIN_V2}/logs`, { params })
  },
  logDels(data: { ids: string }) {
    return request.delete(`${ADMIN_V2}/logs`, deleteBody(data))
  },
  logClear() {
    return request.delete(`${ADMIN_V2}/logs`)
  },
  // 设置
  setupSetContent(data: FormPayload) {
    return request.put(`${ADMIN_V2}/settings/content`, data)
  },
  setupGetContent(key: string) {
    return request.get<string | null>(`${ADMIN_V2}/settings/content`, { params: { key } })
  },
  // 字典管理
  dictTypes() {
    return request.get<DictTypeSummary[]>(`${ADMIN_V2}/dict/types`)
  },
  dictTypeAdd(data: DictTypePayload) {
    return request.post(`${ADMIN_V2}/dict/types`, data)
  },
  dictTypeEdit(typeCode: string, data: Omit<DictTypePayload, 'typeCode'>) {
    return request.put(`${ADMIN_V2}/dict/types/${encodePath(typeCode)}`, data)
  },
  dictTypeDelete(typeCode: string) {
    return request.delete(`${ADMIN_V2}/dict/types/${encodePath(typeCode)}`)
  },
  dictItems(typeCode: string) {
    return request.get<DictItem[]>(`${ADMIN_V2}/dict/items`, { params: { typeCode } })
  },
  dictActiveItems(typeCode: string) {
    return request.get<DictItem[]>(`${API_V2}/dict/items`, { params: { typeCode } })
  },
  dictAdd(data: DictItemPayload & { typeName?: string }) {
    return request.post(`${ADMIN_V2}/dict/items`, data)
  },
  dictEdit(data: Partial<DictItemPayload> & { id: ID }) {
    return request.put(`${ADMIN_V2}/dict/items/${encodePath(idFrom(data))}`, data)
  },
  dictDel(data: { id: ID }) {
    return request.delete(`${ADMIN_V2}/dict/items/${encodePath(data.id)}`, deleteBody(data))
  },
  dictTypeClearItems(typeCode: string) {
    return request.delete(`${ADMIN_V2}/dict/types/${encodePath(typeCode)}/items`, deleteBody({ typeCode }))
  },
  dictClear(typeCode: string) {
    return request.delete(`${ADMIN_V2}/dict/types/${encodePath(typeCode)}/items`, deleteBody({ typeCode }))
  },
  dictEditTypeName(data: { typeCode: string; typeName: string; oldTypeCode?: string }) {
    return request.patch(`${ADMIN_V2}/dict/types/${encodePath(data.oldTypeCode ?? data.typeCode)}`, data)
  },
  // 部门管理
  deptTree() {
    return request.get<DepartmentNode[]>(`${ADMIN_V2}/departments/tree`)
  },
  deptAdd(data: FormPayload) {
    return request.post(`${ADMIN_V2}/departments`, data)
  },
  deptEdit(data: FormPayload & { id?: ID }) {
    return request.put(`${ADMIN_V2}/departments/${encodePath(idFrom(data))}`, data)
  },
  deptDel(data: { id: ID }) {
    return request.delete(`${ADMIN_V2}/departments/${encodePath(data.id)}`, deleteBody(data))
  },
  // 岗位管理
  positionList(params?: PageQuery) {
    return request.get<PageResult<AdminListRecord>>(`${ADMIN_V2}/positions`, { params })
  },
  positionAdd(data: FormPayload) {
    return request.post(`${ADMIN_V2}/positions`, data)
  },
  positionEdit(data: FormPayload & { id?: ID }) {
    return request.put(`${ADMIN_V2}/positions/${encodePath(idFrom(data))}`, data)
  },
  positionDel(data: { id: ID }) {
    return request.delete(`${ADMIN_V2}/positions/${encodePath(data.id)}`, deleteBody(data))
  },
  // 角色管理
  roleList(params?: PageQuery) {
    return request.get<PageResult<AdminListRecord>>(`${ADMIN_V2}/roles`, { params })
  },
  roleAdd(data: FormPayload) {
    return request.post(`${ADMIN_V2}/roles`, data)
  },
  roleEdit(data: FormPayload & { id?: ID }) {
    return request.put(`${ADMIN_V2}/roles/${encodePath(idFrom(data))}`, data)
  },
  roleDel(data: { id: ID }) {
    return request.delete(`${ADMIN_V2}/roles/${encodePath(data.id)}`, deleteBody(data))
  },
  roleDels(data: { ids: string }) {
    return request.delete(`${ADMIN_V2}/roles`, deleteBody(data))
  },
  appPermissionTree() {
    return request.get<ApplicationPermissionTree>(`${ADMIN_V2}/roles/application-permissions`)
  },
  // 权限管理
  permissionTree(params?: { platform?: string; types?: string }) {
    return request.get<AdminListRecord[]>(`${ADMIN_V2}/permissions/tree`, { params })
  },
  permissionList(params?: { platform?: string; types?: string }) {
    return request.get(`${ADMIN_V2}/permissions`, { params })
  },
  permissionAdd(data: FormPayload) {
    return request.post(`${ADMIN_V2}/permissions`, data)
  },
  permissionEdit(data: FormPayload & { key?: ID; permissionKey?: ID; originalKey?: ID }) {
    const key = data.originalKey ?? data.key ?? data.permissionKey
    return request.put(`${ADMIN_V2}/permissions/${encodePath(key)}`, data)
  },
  permissionDel(data: { key?: ID; permissionKey?: ID }) {
    const key = data.key ?? data.permissionKey
    return request.delete(`${ADMIN_V2}/permissions/${encodePath(key)}`, deleteBody(data))
  },
  // 赛事活动管理
  eventList(params?: PageQuery) {
    return request.get<PageResult<EventItem>>(`${ADMIN_V2}/events`, { params })
  },
  eventDetail(id: ID) {
    return request.get<EventItem>(`${ADMIN_V2}/events/${encodePath(id)}`)
  },
  eventInsert(data: FormPayload) {
    return request.post(`${ADMIN_V2}/events`, data)
  },
  eventEdit(data: FormPayload & { id?: ID }) {
    return request.put(`${ADMIN_V2}/events/${encodePath(idFrom(data))}`, data)
  },
  eventDel(data: { id: ID }) {
    return request.delete(`${ADMIN_V2}/events/${encodePath(data.id)}`, deleteBody(data))
  },
  eventDels(data: { ids: string }) {
    return request.delete(`${ADMIN_V2}/events`, deleteBody(data))
  },
  eventStatus(data: { id: ID; status: number | string }) {
    return request.patch(`${ADMIN_V2}/events/${encodePath(data.id)}/status`, data)
  },
  eventParticipantList(params?: PageQuery & { eventId?: ID }) {
    return request.get<PageResult<AdminListRecord>>(`${ADMIN_V2}/events/${encodePath(params?.eventId, 0)}/participants`, { params })
  },
  eventParticipantDel(data: { id: ID; eventId?: ID }) {
    return request.delete(`${ADMIN_V2}/events/${encodePath(data.eventId, 0)}/participants/${encodePath(data.id)}`, deleteBody(data))
  },
  eventParticipantDels(data: { ids: string; eventId?: ID }) {
    return request.delete(`${ADMIN_V2}/events/${encodePath(data.eventId, 0)}/participants`, deleteBody(data))
  },
  eventDynamics(params?: PageQuery & { eventId?: ID }) {
    return request.get<PageResult<AdminListRecord>>(`${ADMIN_V2}/events/${encodePath(params?.eventId, 0)}/dynamics`, { params })
  },
  eventDynamicAdd(data: FormPayload & { eventId?: ID }) {
    return request.post(`${ADMIN_V2}/events/${encodePath(data.eventId, 0)}/dynamics`, data)
  },
  eventDynamicEdit(data: FormPayload & { id?: ID; eventId?: ID }) {
    return request.put(`${ADMIN_V2}/events/${encodePath(data.eventId, 0)}/dynamics/${encodePath(idFrom(data))}`, data)
  },
  eventDynamicDel(data: { id: ID; eventId?: ID }) {
    return request.delete(`${ADMIN_V2}/events/${encodePath(data.eventId, 0)}/dynamics/${encodePath(data.id)}`, deleteBody(data))
  },
  eventDynamicDels(data: { ids: string; eventId?: ID }) {
    return request.delete(`${ADMIN_V2}/events/${encodePath(data.eventId, 0)}/dynamics`, deleteBody(data))
  },
  eventScores(params?: PageQuery & { eventId?: ID }) {
    return request.get<PageResult<AdminListRecord>>(`${ADMIN_V2}/events/${encodePath(params?.eventId, 0)}/scores`, { params })
  },
  eventScoreEdit(data: FormPayload & { id?: ID; eventId?: ID }) {
    if (data.id) {
      return request.put(`${ADMIN_V2}/events/${encodePath(data.eventId, 0)}/scores/${encodePath(data.id)}`, data)
    }
    return request.post(`${ADMIN_V2}/events/${encodePath(data.eventId, 0)}/scores`, data)
  },
  eventVouch(data: { id: ID; vouch?: number | string; status?: number | string; isVouch?: number | string }) {
    return request.patch(`${ADMIN_V2}/events/${encodePath(data.id)}/recommendation`, { ...data, vouch: vouchValue(data) })
  },
  eventTop(data: { id: ID; top?: number | string; status?: number | string; isTop?: number | string }) {
    return request.patch(`${ADMIN_V2}/events/${encodePath(data.id)}/top`, { ...data, top: data.top ?? data.isTop ?? data.status ?? 0 })
  },
  deptUsers(params?: QueryParams) {
    return request.get<PageResult<AdminListRecord>>(`${ADMIN_V2}/event-dept-users`, { params })
  },
  // 当前管理员的菜单和权限
  adminMenus() {
    return request.get<AdminMenuItem[]>(`${ADMIN_V2}/me/menus`)
  },
  adminPerms() {
    return request.get<string[]>(`${ADMIN_V2}/me/perms`)
  },
  // 在线用户
  onlineUsers() {
    return request.get<AdminListRecord[]>(`${ADMIN_V2}/user-sessions`)
  },
  onlineAdmins() {
    return request.get<AdminListRecord[]>(`${ADMIN_V2}/admin-sessions`)
  },
  forceOfflineAdmin(data: { id: string | number, token: string }) {
    return request.post(`${ADMIN_V2}/admin-sessions/${encodePath(data.id)}/force-offline`, data)
  },
  forceOfflineUser(data: { id: string | number, token: string }) {
    return request.post(`${ADMIN_V2}/user-sessions/${encodePath(data.id)}/force-offline`, data)
  },
  batchForceOfflineAdmin(items: { idStr: string | number, token: string }[]) {
    return request.post<{ count: number }, typeof items>(`${ADMIN_V2}/admin-sessions/batch-force-offline`, items, jsonConfig)
  },
  batchForceOfflineUser(items: { idStr: string | number, token: string }[]) {
    return request.post<{ count: number }, typeof items>(`${ADMIN_V2}/user-sessions/batch-force-offline`, items, jsonConfig)
  },
  adminLogout() {
    return request.post(`${ADMIN_V2}/auth/logout`)
  },
  // Formkit (题型元信息 / schema 校验 / 表达式试算) - 已合并到 survey
  formkitTypes() {
    return request.get<FormkitTypeMeta[]>(`${ADMIN_V2}/survey-types`)
  },
  formkitParseSchema(schema: string) {
    return request.post<{ questions: FormkitQuestion[] }, { schema: string }>(`${ADMIN_V2}/survey-schema/parse`, { schema }, jsonConfig)
  },
  formkitEval(data: { expr: string; env: Record<string, unknown>; asBool?: boolean }) {
    return request.post<{ value: unknown }, typeof data>(`${ADMIN_V2}/survey-expressions/evaluate`, data, jsonConfig)
  },
  formkitReportEnroll(enrollId: ID) {
    return request.get<FormkitReport>(`${ADMIN_V2}/survey-report/enroll`, { params: { enrollId } })
  },
  formkitReportEvent(eventId: ID) {
    return request.get<FormkitReport>(`${ADMIN_V2}/survey-report/event`, { params: { eventId } })
  },
  formkitSaveToBank(data: FormPayload) {
    return request.post(`${ADMIN_V2}/survey-question-bank`, { ...data, fromFormkit: true }, jsonConfig)
  },
  // 题库 + 考试 (P7 -> 已合并到 survey)
  examQuestionList(params?: PageQuery & { category?: string; type?: string }) {
    return request.get<PageResult<QuestionBankItem>>(`${ADMIN_V2}/survey-question-bank`, { params })
  },
  examQuestionInsert(data: FormPayload) {
    return request.post(`${ADMIN_V2}/survey-question-bank`, data, jsonConfig)
  },
  examQuestionEdit(data: FormPayload & { id?: ID }) {
    return request.put(`${ADMIN_V2}/survey-question-bank/${encodePath(idFrom(data))}`, data, jsonConfig)
  },
  examQuestionDel(data: { id: number }) {
    return request.delete(`${ADMIN_V2}/survey-question-bank/${encodePath(data.id)}`, deleteBody(data))
  },
  examPaperList(params?: PageQuery) {
    return request.get(`${ADMIN_V2}/surveys`, { params })
  },
  examPaperDetail(id: number) {
    return request.get(`${ADMIN_V2}/surveys/${encodePath(id)}`)
  },
  // Survey 独立子系统
  surveyList(params?: PageQuery) {
    return request.get<PageResult<SurveyItem>>(`${ADMIN_V2}/surveys`, { params })
  },
  surveyDetail(id: number) {
    return request.get<{ survey: SurveyItem; responseCount: number; schema: string }>(`${ADMIN_V2}/surveys/${encodePath(id)}`)
  },
  surveyInsert(data: FormPayload) {
    return request.post(`${ADMIN_V2}/surveys`, data)
  },
  surveyEdit(data: FormPayload & { id?: ID }) {
    return request.put(`${ADMIN_V2}/surveys/${encodePath(idFrom(data))}`, data, jsonConfig)
  },
  surveyDel(data: { id: number }) {
    return request.delete(`${ADMIN_V2}/surveys/${encodePath(data.id)}`, deleteBody(data))
  },
  surveyStatus(data: { id: number; status: number }) {
    return request.patch(`${ADMIN_V2}/surveys/${encodePath(data.id)}/status`, data)
  },
  surveyCopy(data: { id: number }) {
    return request.post(`${ADMIN_V2}/surveys/${encodePath(data.id)}/copy`, data)
  },
  surveyResponseList(params: { surveyId: number; page?: number; pageSize?: number; keyword?: string }) {
    return request.get(`${ADMIN_V2}/surveys/${encodePath(params.surveyId)}/responses`, { params })
  },
  surveyResponseDetail(id: number) {
    return request.get(`${ADMIN_V2}/survey-responses/${encodePath(id)}`)
  },
  surveyResponseDel(data: { id: number }) {
    return request.delete(`${ADMIN_V2}/survey-responses/${encodePath(data.id)}`, deleteBody(data))
  },
  surveyResponseBatchDel(data: { ids: string }) {
    return request.delete(`${ADMIN_V2}/survey-responses`, deleteBody(data))
  },
  surveyStatistic(surveyId: number) {
    return request.get(`${ADMIN_V2}/surveys/${encodePath(surveyId)}/statistics`, { params: { surveyId } })
  },
  surveyChannelList(surveyId: number) {
    return request.get(`${ADMIN_V2}/surveys/${encodePath(surveyId)}/channels`, { params: { surveyId } })
  },
  surveyChannelInsert(data: FormPayload & { surveyId?: ID }) {
    return request.post(`${ADMIN_V2}/surveys/${encodePath(data.surveyId, 0)}/channels`, data, jsonConfig)
  },
  surveyChannelDel(data: { id: number }) {
    return request.delete(`${ADMIN_V2}/survey-channels/${encodePath(data.id)}`, deleteBody(data))
  },
  surveyResponseExport(surveyId: number) {
    return request.get(`${ADMIN_V2}/surveys/${encodePath(surveyId)}/responses/export`, { params: { surveyId }, responseType: 'blob' })
  },
  surveyResourceList(params: { surveyId: number; resType?: string }) {
    return request.get<ResourceItem[]>(`${ADMIN_V2}/surveys/${encodePath(params.surveyId)}/resources`, { params })
  },
  surveyResourceDelete(data: { id: number }) {
    return request.delete(`${ADMIN_V2}/survey-resources/${encodePath(data.id)}`, deleteBody(data))
  },
  // ==================== Exam 独立子系统 ====================
  examList(params?: PageQuery) {
    return request.get<PageResult<ExamItem>>(`${ADMIN_V2}/exams`, { params })
  },
  examDetail(id: number) {
    return request.get<{ survey: ExamItem; responseCount: number; schema: string }>(`${ADMIN_V2}/exams/${encodePath(id)}`)
  },
  examSave(data: FormPayload & { id?: ID }) {
    if (data.id) {
      return request.put(`${ADMIN_V2}/exams/${encodePath(data.id)}`, data, jsonConfig)
    }
    return request.post(`${ADMIN_V2}/exams`, data, jsonConfig)
  },
  examDelete(data: { id: number }) {
    return request.delete(`${ADMIN_V2}/exams/${encodePath(data.id)}`, deleteBody(data))
  },
  examStatus(data: { id: number; status: number }) {
    return request.patch(`${ADMIN_V2}/exams/${encodePath(data.id)}/status`, data)
  },
  examRecordList(params?: PageQuery & { examId?: ID }) {
    return request.get(`${ADMIN_V2}/exams/${encodePath(params?.examId, 0)}/records`, { params })
  },
  examRecordDetail(id: number) {
    return request.get(`${ADMIN_V2}/exams/0/records/${encodePath(id)}`, { params: { id } })
  },
  examRecordDel(data: { id: number; examId?: ID }) {
    return request.delete(`${ADMIN_V2}/exams/${encodePath(data.examId, 0)}/records/${encodePath(data.id)}`, deleteBody(data))
  },
  examRecordBatchDel(data: { ids: string; examId?: ID }) {
    return request.delete(`${ADMIN_V2}/exams/${encodePath(data.examId, 0)}/records`, deleteBody(data))
  },
  examStatistics(examId: number) {
    return request.get(`${ADMIN_V2}/exams/${encodePath(examId)}/statistics`, { params: { examId } })
  },
  // ==================== Survey Question Bank ====================
  surveyQuestionBankList(params?: PageQuery & { category?: string; type?: string }) {
    return request.get<PageResult<QuestionBankItem>>(`${ADMIN_V2}/survey-question-bank`, { params })
  },
  surveyQuestionBankInsert(data: FormPayload) {
    return request.post(`${ADMIN_V2}/survey-question-bank`, data, jsonConfig)
  },
  surveyQuestionBankEdit(data: FormPayload & { id?: ID }) {
    return request.put(`${ADMIN_V2}/survey-question-bank/${encodePath(idFrom(data))}`, data, jsonConfig)
  },
  surveyQuestionBankDel(data: { id: number }) {
    return request.delete(`${ADMIN_V2}/survey-question-bank/${encodePath(data.id)}`, deleteBody(data))
  },
  surveyQuestionBankCategories() {
    return request.get<string[]>(`${ADMIN_V2}/survey-question-bank/categories`)
  },
  examResourceList(params: { examId: number; resType?: string }) {
    return request.get<ResourceItem[]>(`${ADMIN_V2}/exams/${encodePath(params.examId)}/resources`, { params })
  },
  examResourceDelete(data: { id: number }) {
    return request.delete(`${ADMIN_V2}/exam-resources/${encodePath(data.id)}`, deleteBody(data))
  },
  // ==================== Exam Question Bank ====================
  examQuestionBankList(params?: PageQuery & { category?: string; type?: string }) {
    return request.get<PageResult<QuestionBankItem>>(`${ADMIN_V2}/exam-question-bank`, { params })
  },
  examQuestionBankInsert(data: FormPayload) {
    return request.post(`${ADMIN_V2}/exam-question-bank`, data, jsonConfig)
  },
  examQuestionBankEdit(data: FormPayload & { id?: ID }) {
    return request.put(`${ADMIN_V2}/exam-question-bank/${encodePath(idFrom(data))}`, data, jsonConfig)
  },
  examQuestionBankDel(data: { id: number }) {
    return request.delete(`${ADMIN_V2}/exam-question-bank/${encodePath(data.id)}`, deleteBody(data))
  },
  examQuestionBankCategories() {
    return request.get<string[]>(`${ADMIN_V2}/exam-question-bank/categories`)
  },
  surveyNotifyList(params?: QueryParams) {
    return request.get(`${ADMIN_V2}/survey-notifications`, { params })
  },
  surveyNotifyRead(data: { id?: number; all?: boolean; userId?: string }) {
    return request.patch(`${ADMIN_V2}/survey-notifications/${encodePath(data.id, 0)}/read`, data, jsonConfig)
  },
  surveyNotifyUnreadCount(params?: { userId?: string }) {
    return request.get(`${ADMIN_V2}/survey-notifications/unread-count`, { params })
  },
  inAppNotificationList(params?: PageQuery) {
    return request.get<InAppNotificationList>(`${ADMIN_V2}/in-app-notifications`, { params })
  },
  inAppNotificationUnreadCount() {
    return request.get<{ count: number }>(`${ADMIN_V2}/in-app-notifications/unread-count`)
  },
  inAppNotificationRecipientOptions() {
    return request.get<InAppNotificationRecipientOptions>(`${ADMIN_V2}/in-app-notifications/recipient-options`)
  },
  inAppNotificationMarkRead(id: ID) {
    return request.patch(`${ADMIN_V2}/in-app-notifications/${encodePath(id)}/read`)
  },
  inAppNotificationMarkAllRead() {
    return request.patch(`${ADMIN_V2}/in-app-notifications/read-all`)
  },
  inAppNotificationSend(data: InAppNotificationSendPayload) {
    return request.post<InAppNotificationSendResult, InAppNotificationSendPayload>(`${ADMIN_V2}/in-app-notifications`, data, jsonConfig)
  },
  dingTalkNotificationRecipientOptions() {
    return request.get<InAppNotificationRecipientOptions>(`${ADMIN_V2}/dingtalk-notifications/recipient-options`)
  },
  dingTalkNotificationSend(data: DingTalkNotificationSendPayload) {
    return request.post<DingTalkNotificationSendResult, DingTalkNotificationSendPayload>(`${ADMIN_V2}/dingtalk-notifications`, data, jsonConfig)
  },
  surveyTemplatePresetsGet() {
    return request.get<TemplatePreset[]>(`${ADMIN_V2}/survey-template-presets`)
  },
  surveyTemplatePresetsSave(data: { presets: TemplatePreset[] }) {
    return request.put(`${ADMIN_V2}/survey-template-presets`, data, jsonConfig)
  },
  // 流程定义与版本管理
  workflowDefinitionList(params?: PageQuery & { category?: string; status?: number | string }) {
    return request.get<PageResult<WorkflowDefinitionSummary>>(`${ADMIN_V2}/workflow-definitions`, { params })
  },
  workflowDefinitionDetail(id: ID) {
    return request.get<WorkflowDefinitionDetail>(`${ADMIN_V2}/workflow-definitions/${encodePath(id)}`)
  },
  workflowDefinitionCreate(data: FormPayload | FormData) {
    return request.post(`${ADMIN_V2}/workflow-definitions`, data, data instanceof FormData ? undefined : jsonConfig)
  },
  workflowDefinitionCopy(id: ID, data: FormPayload | FormData) {
    return request.post<WorkflowDefinitionDetail>(`${ADMIN_V2}/workflow-definitions/${encodePath(id)}/copy`, data, data instanceof FormData ? undefined : jsonConfig)
  },
  workflowDefinitionUpdate(id: ID, data: FormPayload | FormData) {
    return request.put(`${ADMIN_V2}/workflow-definitions/${encodePath(id)}`, data, data instanceof FormData ? undefined : jsonConfig)
  },
  workflowDefinitionDelete(id: ID) {
    return request.delete(`${ADMIN_V2}/workflow-definitions/${encodePath(id)}`)
  },
  workflowDefinitionValidate(id: ID) {
    return request.post<WorkflowValidationResult>(`${ADMIN_V2}/workflow-definitions/${encodePath(id)}/validate`)
  },
  workflowDefinitionPublish(id: ID, data: { initiator?: WorkflowInitiatorConfig; note?: string } = {}) {
    return request.post<WorkflowPublishResult, typeof data>(`${ADMIN_V2}/workflow-definitions/${encodePath(id)}/publish`, data, jsonConfig)
  },
  workflowDefinitionVersions(id: ID) {
    return request.get<WorkflowVersion[]>(`${ADMIN_V2}/workflow-definitions/${encodePath(id)}/versions`)
  },
  workflowDefinitionVersionChanges(id: ID, version: number, compareTo?: number) {
    return request.get<WorkflowVersionChangeSummary>(`${ADMIN_V2}/workflow-definitions/${encodePath(id)}/versions/${encodePath(version)}/changes`, {
      params: compareTo ? { compareTo } : undefined,
    })
  },
  workflowDefinitionVersionDelete(id: ID, version: number) {
    return request.delete(`${ADMIN_V2}/workflow-definitions/${encodePath(id)}/versions/${encodePath(version)}`)
  },
  workflowDefinitionVersionRollback(id: ID, version: number, data: { note?: string } = {}) {
    return request.post<WorkflowPublishResult, typeof data>(`${ADMIN_V2}/workflow-definitions/${encodePath(id)}/versions/${encodePath(version)}/rollback`, data, jsonConfig)
  },
  workflowOrgApproverIdentities() {
    return request.get(`${ADMIN_V2}/workflow-org-approver-identities`)
  },
  workflowOrgApproverAssignments(params?: { subjectType?: 'department' | 'user'; subjectId?: ID; departmentId?: ID; identityCode?: string }) {
    return request.get(`${ADMIN_V2}/workflow-org-approver-assignments`, { params })
  },
  workflowOrgApproverAssignmentsSave(data: FormPayload & { subjectType?: 'department' | 'user'; subjectId?: ID; departmentId?: ID; identityCode?: string; userIds?: ID[] }) {
    return request.put(`${ADMIN_V2}/workflow-org-approver-assignments`, data, jsonConfig)
  },
  // 通用工作流运行时
  workflowPublishedDefinitionList(params?: QueryParams) {
    return request.get<WorkflowPublishedDefinition[]>(`${ADMIN_V2}/workflow-published-definitions`, { params })
  },
  workflowPublishedDefinitionDetail(id: ID) {
    return request.get<WorkflowPublishedDefinition>(`${ADMIN_V2}/workflow-published-definitions/${encodePath(id)}`)
  },
  workflowUserOptions(params?: PageQuery) {
    return request.get<PageResult<AdminUser>>(`${ADMIN_V2}/workflow-user-options`, { params })
  },
  workflowDepartmentOptions() {
    return request.get(`${ADMIN_V2}/workflow-department-options`)
  },
  workflowInstanceList(params?: PageQuery & {
    definitionId?: ID
    status?: string
    businessType?: string
    businessKey?: string
    starterId?: string
  }) {
    return request.get<PageResult<WorkflowInstanceSummary>>(`${ADMIN_V2}/workflow-instances`, { params })
  },
  workflowInstanceStart(data: FormPayload) {
    return request.post<{ instanceId: string }, FormPayload>(`${ADMIN_V2}/workflow-instances`, data, jsonConfig)
  },
  workflowInstanceDetail(id: ID) {
    return request.get<WorkflowInstanceDetail>(`${ADMIN_V2}/workflow-instances/${encodePath(id)}`)
  },
  workflowInstanceDelete(id: ID) {
    return request.delete<{ deleted: number }>(`${ADMIN_V2}/workflow-instances/${encodePath(id)}`)
  },
  workflowInstanceBatchDelete(ids: ID[]) {
    return request.delete<{ deleted: number }>(`${ADMIN_V2}/workflow-instances`, deleteBody({ ids }))
  },
  workflowInstanceResume(id: ID) {
    return request.post<{ advanced: number }, Record<string, never>>(`${ADMIN_V2}/workflow-instances/${encodePath(id)}/resume`, {}, jsonConfig)
  },
  workflowTaskList(params?: PageQuery & { instanceId?: string; assigneeId?: string; status?: string }) {
    return request.get<PageResult<WorkflowTaskSummary>>(`${ADMIN_V2}/workflow-tasks`, { params })
  },
  workflowTaskComplete(id: ID, data: FormPayload) {
    return request.post(`${ADMIN_V2}/workflow-tasks/${encodePath(id)}/complete`, data, jsonConfig)
  },
  workflowTaskDelete(id: ID) {
    return request.delete<{ id: string }>(`${ADMIN_V2}/workflow-tasks/${encodePath(id)}`)
  },
  workflowNotificationList(params?: PageQuery & {
    instanceId?: string
    recipientUserId?: string
    kind?: string
    channel?: string
    status?: string
  }) {
    return request.get<WorkflowNotificationList>(`${ADMIN_V2}/workflow-notifications`, { params })
  },
  workflowNotificationRetry(id: ID) {
    return request.post(`${ADMIN_V2}/workflow-notifications/${encodePath(id)}/retry`, {}, jsonConfig)
  },
  workflowNotificationDispatchDue(data: { limit?: number } = {}) {
    return request.post<{ dispatched: number }, typeof data>(`${ADMIN_V2}/workflow-notifications/dispatch-due`, data, jsonConfig)
  },
  // 通用定时任务
  scheduledTaskList(params?: QueryParams) {
    return request.get<ScheduledTaskPageData<ScheduledTask>>(`${ADMIN_V2}/scheduled-tasks`, { params })
  },
  scheduledTaskDetail(id: ID) {
    return request.get<ScheduledTask>(`${ADMIN_V2}/scheduled-tasks/${encodePath(id)}`)
  },
  scheduledTaskCreate(data: object) {
    return request.post(`${ADMIN_V2}/scheduled-tasks`, data, jsonConfig)
  },
  scheduledTaskUpdate(id: ID, data: object) {
    return request.put(`${ADMIN_V2}/scheduled-tasks/${encodePath(id)}`, data, jsonConfig)
  },
  scheduledTaskDelete(id: ID) {
    return request.delete(`${ADMIN_V2}/scheduled-tasks/${encodePath(id)}`)
  },
  scheduledTaskStatus(id: ID, data: { enabled: boolean; version: number }) {
    return request.patch<ScheduledTask, typeof data>(`${ADMIN_V2}/scheduled-tasks/${encodePath(id)}/status`, data, jsonConfig)
  },
  scheduledTaskRun(id: ID) {
    return request.post<{ dispatchPending: boolean }, Record<string, never>>(`${ADMIN_V2}/scheduled-tasks/${encodePath(id)}/run`, {}, jsonConfig)
  },
  scheduledTaskCronPreview(data: FormPayload) {
    return request.post<{ occurrences: CronOccurrence[] }, FormPayload>(`${ADMIN_V2}/scheduled-tasks/cron-preview`, data, jsonConfig)
  },
  scheduledTaskHandlers() {
    return request.get<HandlerMetadata[]>(`${ADMIN_V2}/scheduled-task-handlers`)
  },
  scheduledTaskRunList(params?: QueryParams) {
    return request.get<ScheduledTaskPageData<ScheduledTaskRun>>(`${ADMIN_V2}/scheduled-task-runs`, { params })
  },
  scheduledTaskRunDetail(id: ID) {
    return request.get<ScheduledTaskRunDetail>(`${ADMIN_V2}/scheduled-task-runs/${encodePath(id)}`)
  },
  scheduledTaskRunRetry(id: ID) {
    return request.post<{ dispatchPending: boolean }, Record<string, never>>(`${ADMIN_V2}/scheduled-task-runs/${encodePath(id)}/retry`, {}, jsonConfig)
  },
  scheduledTaskRunCancel(id: ID) {
    return request.post(`${ADMIN_V2}/scheduled-task-runs/${encodePath(id)}/cancel`, {}, jsonConfig)
  },
  scheduledTaskWorkers() {
    return request.get<ScheduledTaskWorker[]>(`${ADMIN_V2}/scheduled-task-workers`)
  }
}
