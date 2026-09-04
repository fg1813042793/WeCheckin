export interface ApiEnvelope<T = unknown> {
  code?: number
  msg?: string
  message?: string
  data?: T
  [key: string]: unknown
}

export interface AppConfig {
  appTitle: string
  appName: string
  logoText: string
  logoUrl: string
  appUrl: string
}

export interface PublicConfigPayload extends Partial<AppConfig> {
  appConfig?: Partial<AppConfig>
  corpId?: string
}

export interface DingTalkMenu {
  key: string
  label?: string
  name?: string
  title?: string
  menuName?: string
  icon?: string
  children?: DingTalkMenu[]
  [key: string]: unknown
}

export interface DingTalkUser {
  id: string
  workflowActorId?: string
  name: string
  account?: string
  avatar?: string
  avatarUrl?: string
  department?: string
  departmentLevel1?: string
  departmentLevel2?: string
  departmentLevel3?: string
  departmentLevel4?: string
  departmentLevels?: string[]
  position?: string
  status?: number
  userStatus?: number
  [key: string]: unknown
}

export interface AuthSessionPayload {
  token?: string
  user?: DingTalkUser
  userInfo?: DingTalkUser
  menus?: DingTalkMenu[]
  apiPermissionKeys?: string[]
  buttonPermissionKeys?: string[]
  apiPermissionReady?: boolean
  buttonPermissionReady?: boolean
  permissionVersion?: number
  appConfig?: Partial<AppConfig>
  appTitle?: string
  appName?: string
  applicationName?: string
  logoText?: string
  logoUrl?: string
  logoURL?: string
  appUrl?: string
}

export interface LoginRequest {
  name: string
  password: string
}

export interface SsoLoginRequest {
  corpId: string
  authCode: string
}

export interface BindSelfRequest {
  bindTicket: string
  account: string
  password: string
}

export interface BindRequiredPayload extends Partial<AuthSessionPayload> {
  bindTicket?: string
  corpId?: string
  dingTalkUserIdMasked?: string
  unionIdMasked?: string
  expiresIn?: number
}

export interface WorkbenchStat {
  key: string
  label: string
  value: number
}

export type ReviewStatus
  = | 'draft'
    | 'manager_review'
    | 'hrbp_review'
    | 'employee_confirm'
    | 'hr_final'
    | 'completed'
    | string

export interface PerformanceReview {
  id: string
  reviewNo?: string
  employeeId?: string
  employeeName?: string
  employee?: string
  managerId?: string
  managerName?: string
  hrbpId?: string
  hrbpName?: string
  hrbpReviewerId?: string
  hrbpReviewerName?: string
  period?: string
  status?: ReviewStatus
  department?: string
  departmentLevel1?: string
  departmentLevel2?: string
  departmentLevel3?: string
  departmentLevel4?: string
  departmentLevels?: string[]
  position?: string
  objectiveScore?: number | string
  managerScore?: number | string
  hrbpScore?: number | string
  finalGrade?: string
  employeeConfirmResult?: string
  currentAssigneeId?: string
  currentAssigneeName?: string
  assigneeId?: string
  currentHandlerId?: string
  currentHandlerUserId?: string
  createdAt?: string
  updatedAt?: string
  [key: string]: unknown
}

export interface PerformanceReviewListPayload {
  list?: PerformanceReview[]
  rows?: PerformanceReview[]
  items?: PerformanceReview[]
  total?: number
}

export interface PerformanceUser {
  id: string
  name: string
  password?: string
  position?: string
  department?: string
  departmentLevel1?: string
  departmentLevel2?: string
  departmentLevel3?: string
  departmentLevel4?: string
  departmentLevels?: string[]
  managerId?: string
  hrbpId?: string
  status?: number
  userStatus?: number
  [key: string]: unknown
}

export interface PerformanceTemplate {
  gradeLevels?: Array<{ grade: string, label?: string, min?: number, max?: number }>
  [key: string]: unknown
}

export interface ReviewActionRequest {
  action: string
  remark?: string
  [key: string]: unknown
}

export interface CreateReviewRequest {
  employeeId?: string
  employeeIds?: string[]
  period: string
  nextPeriod?: string
  [key: string]: unknown
}

export interface CreateReviewPayload {
  list?: PerformanceReview[]
  failed?: unknown[]
  id?: string
  [key: string]: unknown
}

export interface UpdateProfileRequest {
  account?: string
  name?: string
  avatar?: string
  currentPassword?: string
  [key: string]: unknown
}

export interface ChangePasswordRequest {
  currentPassword?: string
  oldPassword?: string
  newPassword: string
  confirmPassword?: string
}

export interface AvatarUploadPayload {
  url?: string
  avatar?: string
  path?: string
  [key: string]: unknown
}

export interface DingTalkHttpError {
  code?: number
  msg?: string
  message?: string
  data?: unknown
  statusCode?: number
  [key: string]: unknown
}
