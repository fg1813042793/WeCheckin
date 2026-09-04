<script setup lang="ts">
import type { ReviewActionKey } from '../constants/performancePermissions'
import type { PerformanceReview, PerformanceTemplate, PerformanceUser, ReviewActionRequest } from '@/types/dingtalk-h5'
import { computed, onBeforeUnmount, reactive, ref, watch } from 'vue'
import { useDingtalkAuthStore } from '@/stores'
import { reviewActionApiPermissions, reviewActionButtonPermissions } from '../constants/performancePermissions'
import { resolveWithdrawTargetStatus, statusMeta } from '../constants/performanceStatus'
import PerformanceAdaptiveSelect from './PerformanceAdaptiveSelect.vue'

type FormRecord = Record<string, any>
type ReviewActionScope = 'mine' | 'manager' | 'hrbp' | 'readonly'
type ActionTone = 'default' | 'primary' | 'success' | 'warning' | 'error'
type ValueScoreField = 'self' | 'manager' | 'hrbp'

interface ReviewObjective extends FormRecord {
  target: string
  weight: number | string
  completion?: number | string
  result?: string
}

interface ReviewValueScore extends FormRecord {
  id: string
  name: string
  definition?: string
  rubric?: ReviewValueRubric[]
  self?: number | string
  manager?: number | string
  hrbp?: number | string
}

interface ReviewValueRubric extends FormRecord {
  score?: number | string
  label?: string
  description?: string
}

interface ReviewActionItem {
  action: ReviewActionKey
  label: string
  type: ActionTone
  plain?: boolean
}

interface ReviewActionConfirmCopy {
  title: string
  content: string
  confirmText: string
  confirmColor?: string
}

interface ReviewActionConfirmResult {
  confirmed: boolean
  remark: string
  finalGrade?: string
}

interface ReviewHistoryRecord extends FormRecord {
  action?: string
  by?: string
  at?: string | number
}

interface SelectOption {
  value: string
  label: string
  type?: 'group'
}

interface FlowStepDefinition {
  status: string
  title: string
  role: string
  desc: string
  actor: 'employee' | 'manager' | 'hrbp' | 'system'
  historyKeywords: string[]
}

interface FlowProgressRow extends FlowStepDefinition {
  index: number
  indexText: string
  progressText: string
  actorName: string
  state: 'done' | 'active' | 'pending'
  stateClass: 'done' | 'active' | 'pending' | 'returned'
  stateLabel: string
  returnStatusText: string
  detail: string
  reason: string
  reasonLabel: string
  history?: ReviewHistoryRecord
}

const props = defineProps<{
  review: PerformanceReview | null
  users?: PerformanceUser[]
  template?: PerformanceTemplate | null
  grades?: string[]
  detailLoading?: boolean
  actionScope?: ReviewActionScope
  submitReviewAction?: (id: string, action: string, data: ReviewActionRequest) => Promise<PerformanceReview | void>
}>()

const emit = defineEmits<{
  'action-success': [review?: PerformanceReview | void]
}>()

function sameUserId(left: unknown, right: unknown) {
  const leftText = String(left || '').trim()
  const rightText = String(right || '').trim()
  return Boolean(leftText && rightText && leftText === rightText)
}

function firstText(...values: Array<unknown>) {
  for (const value of values) {
    const text = String(value || '').trim()
    if (text) {
      return text
    }
  }
  return ''
}

function baseReviewStatusText(review: PerformanceReview) {
  return statusMeta[String(review.status || '')]?.label || String(review.status || '未知')
}

function reviewStatusType(review: PerformanceReview) {
  return statusMeta[String(review.status || '')]?.type || 'info'
}

function currentAssignee(review: PerformanceReview, resolveUserName: (id: unknown) => string) {
  const explicitAssignee = firstText(review.currentAssigneeName)
  if (explicitAssignee && review.status !== 'completed') {
    return explicitAssignee
  }
  if (review.status === 'draft') {
    return firstText(review.employeeName, resolveUserName(review.employeeId))
  }
  if (review.status === 'manager_review') {
    return firstText(review.managerName, resolveUserName(review.managerId))
  }
  if (review.status === 'hrbp_review') {
    return firstText(review.hrbpReviewerName, review.hrbpName, resolveUserName(review.hrbpReviewerId || review.hrbpId))
  }
  if (review.status === 'employee_confirm') {
    return firstText(review.employeeName, resolveUserName(review.employeeId))
  }
  if (review.status === 'hr_final') {
    return firstText(review.hrbpReviewerName, review.hrbpName, resolveUserName(review.hrbpReviewerId || review.hrbpId))
  }
  return '已归档'
}

const auth = useDingtalkAuthStore()
const activeFormTab = ref('currentTargets')
const processVisible = ref(false)
const actionConfirmVisible = ref(false)
const actionConfirmAction = ref<ReviewActionKey | null>(null)
const actionConfirmRemark = ref('')
const actionConfirmFinalGrade = ref('')
const actionConfirmFinalNote = ref('')
const currentObjectiveDeleteVisible = ref(false)
const nextObjectiveDeleteVisible = ref(false)
const valueRubricVisible = ref(false)
const activeValueRubric = ref<ReviewValueScore | null>(null)
const localReview = reactive<FormRecord>({})
const actionConfirmCloseResetDelay = 220
let actionConfirmResolver: ((result: ReviewActionConfirmResult) => void) | null = null
let actionConfirmResetTimer: ReturnType<typeof setTimeout> | null = null
let processPageScrollLocked = false
let processPageScrollTop = 0
let previousBodyOverflow = ''
let previousBodyPosition = ''
let previousBodyTop = ''
let previousBodyWidth = ''
let previousHtmlOverflow = ''

const actionButtonCustomStyle = {
  display: 'inline-flex',
  flex: '0 0 auto',
  width: 'auto',
  minWidth: '152rpx',
  margin: '0',
}

const miniButtonCustomStyle = {
  display: 'inline-flex',
  flex: '0 0 auto',
  width: 'auto',
  minWidth: 'auto',
  height: '52rpx',
  minHeight: '52rpx',
  margin: '0',
  padding: '0 18rpx',
  lineHeight: '50rpx',
}

const formTabButtonCustomStyle = {
  display: 'inline-flex',
  width: 'auto',
  minWidth: 'auto',
  height: '60rpx',
  minHeight: '60rpx',
  margin: '0 12rpx 0 0',
  padding: '0 24rpx',
  lineHeight: '58rpx',
}

const processCloseButtonCustomStyle = {
  display: 'inline-flex',
  width: '68rpx',
  minWidth: '68rpx',
  height: '68rpx',
  minHeight: '68rpx',
  margin: '0',
  padding: '0',
  lineHeight: '68rpx',
}

const fieldInputCustomStyle = {
  width: '100%',
  minHeight: '70rpx',
  height: '70rpx',
  boxSizing: 'border-box',
  fontSize: '24rpx',
}

const nativeTextareaTag = 'textarea'
const OBJECTIVE_NUMBER_MIN = 0
const OBJECTIVE_WEIGHT_MAX = 100
const OBJECTIVE_COMPLETION_MAX = 300
const VALUE_SCORE_MIN = 0
const VALUE_SCORE_MAX = 50
const REVIEW_TEXTAREA_LINE_HEIGHT_RPX = 38.4
const REVIEW_TEXTAREA_VERTICAL_PADDING_RPX = 36
const SELF_SUMMARY_MIN_ROWS = 4
const SELF_SUMMARY_MAX_ROWS = 12
const SELF_SUMMARY_MOBILE_CHARS_PER_ROW = 22
const SELF_SUMMARY_DESKTOP_CHARS_PER_ROW = 92

const reviewFormTabs = [
  { key: 'currentTargets', label: '本月目标' },
  { key: 'selfSummary', label: '思考总结' },
  { key: 'selfValues', label: '价值观自评' },
  { key: 'manager', label: '上级评价' },
  { key: 'hrbp', label: 'HRBP评价' },
  { key: 'nextTargets', label: '下月目标' },
]

const employeeConfirmHeaderActionSet = new Set<ReviewActionKey>()
const reviewActionScopeActions: Record<ReviewActionScope, Set<ReviewActionKey>> = {
  mine: new Set(['save-self', 'submit-self', 'confirm-result', 'dispute-result', 'withdraw']),
  manager: new Set(['return-employee', 'submit-manager']),
  hrbp: new Set(['return-manager', 'submit-hrbp', 'return-hrbp', 'finalize']),
  readonly: new Set(),
}

const flowStepDefinitions: FlowStepDefinition[] = [
  {
    status: 'draft',
    title: '员工填写',
    role: '员工',
    desc: '填写当月目标完成度、达成结果自评、思考总结，并确认下月目标。',
    actor: 'employee',
    historyKeywords: ['创建考评单', '保存员工自评', '提交员工自评', '退回员工修改', '撤销员工自评提交'],
  },
  {
    status: 'manager_review',
    title: '上级评价',
    role: '直属上级',
    desc: '审核员工自评内容，填写上级评价、价值观评分和建议分档。',
    actor: 'manager',
    historyKeywords: ['提交上级评价', '退回上级修改', '撤销上级评价提交'],
  },
  {
    status: 'hrbp_review',
    title: 'HRBP评价',
    role: 'HRBP',
    desc: '复核绩效材料，填写 HRBP 评价和分档；如有问题可退回上级修改。',
    actor: 'hrbp',
    historyKeywords: ['提交 HRBP 评价', '员工提出异议', '撤销 HRBP 评价提交', '退回 HRBP 修改'],
  },
  {
    status: 'employee_confirm',
    title: '员工确认',
    role: '员工',
    desc: '查看评价结果并确认；如存在异议，可填写说明后提交反馈。',
    actor: 'employee',
    historyKeywords: ['员工确认结果', '员工提出异议'],
  },
  {
    status: 'hr_final',
    title: 'HRBP归档',
    role: 'HRBP',
    desc: '处理员工确认或异议，确认最终分档和归档备注。',
    actor: 'hrbp',
    historyKeywords: ['HRBP 归档', '退回 HRBP 修改', '员工确认结果'],
  },
  {
    status: 'completed',
    title: '完成',
    role: '系统归档',
    desc: '考评单归档完成，结果进入汇总统计。',
    actor: 'system',
    historyKeywords: ['HRBP 归档'],
  },
]

const reviewActionConfirmCopy: Partial<Record<ReviewActionKey, ReviewActionConfirmCopy>> = {
  'submit-self': { title: '提交自评', content: '确认提交当前绩效？提交后将进入上级评价流程。', confirmText: '提交' },
  'submit-manager': { title: '提交给HRBP', content: '确认提交给 HRBP？提交后将进入 HRBP 评价流程。', confirmText: '提交' },
  'submit-hrbp': { title: '提交给员工', content: '确认提交给员工确认？提交后员工将查看并确认绩效结果。', confirmText: '提交' },
  'confirm-result': { title: '确认结果', content: '确认绩效结果无误？确认后将进入 HRBP 归档流程。', confirmText: '确认' },
  'dispute-result': { title: '提出异议', content: '确认提交异议？提交后将返回 HRBP 处理。', confirmText: '提交', confirmColor: '#e34d59' },
  'return-employee': { title: '退回员工', content: '确认退回员工修改？退回后员工可重新编辑自评内容。', confirmText: '退回', confirmColor: '#e34d59' },
  'return-manager': { title: '退回上级', content: '确认退回上级修改？退回后上级可重新调整评价。', confirmText: '退回', confirmColor: '#e34d59' },
  'return-hrbp': { title: '退回HRBP', content: '确认退回 HRBP 修改？退回后 HRBP 可重新处理评价。', confirmText: '退回', confirmColor: '#e34d59' },
  'withdraw': { title: '撤回提交', content: '确认撤回当前提交？撤回后流程将返回上一节点。', confirmText: '撤回', confirmColor: '#e34d59' },
  'finalize': { title: '绩效归档', content: '确认归档绩效结果？归档后流程将完成。', confirmText: '归档' },
}

const currentObjectives = computed<ReviewObjective[]>(() => localReview.objectives || [])
const nextObjectives = computed<ReviewObjective[]>(() => localReview.nextObjectives || [])
const values = computed<ReviewValueScore[]>(() => localReview.values || [])
const gradeOptions = computed(() => ['', ...(props.grades || [])])
const gradeGroupDefinitions = [
  { label: '优秀', grades: ['A+', 'A-'] },
  { label: '良好', grades: ['B+', 'B-'] },
  { label: '及格', grades: ['C+', 'C'] },
  { label: '较差', grades: ['C-'] },
  { label: '糟糕', grades: ['D+', 'D-', 'E+', 'E-'] },
]
const gradeGroupLabels = new Set(gradeGroupDefinitions.map(group => group.label))

function isGradeValue(value: string) {
  return /^[A-Z][+-]?$/.test(value)
}

function createGroupedGradeSelectOptions(sourceGrades: unknown[]) {
  const uniqueGrades: string[] = []
  for (const sourceGrade of sourceGrades) {
    const grade = firstText(sourceGrade)
    if (!grade || gradeGroupLabels.has(grade) || !isGradeValue(grade) || uniqueGrades.includes(grade)) {
      continue
    }
    uniqueGrades.push(grade)
  }

  const options: SelectOption[] = [{ value: '', label: '未选择' }]
  const groupedGrades = new Set<string>()
  for (const group of gradeGroupDefinitions) {
    const grades = group.grades.filter(grade => uniqueGrades.includes(grade))
    if (!grades.length) {
      continue
    }
    options.push({ value: `__grade_group_${group.label}`, label: group.label, type: 'group' })
    for (const grade of grades) {
      groupedGrades.add(grade)
      options.push({ value: grade, label: grade })
    }
  }

  const otherGrades = uniqueGrades.filter(grade => !groupedGrades.has(grade))
  if (otherGrades.length) {
    options.push({ value: '__grade_group_other', label: '其他', type: 'group' })
    for (const grade of otherGrades) {
      options.push({ value: grade, label: grade })
    }
  }

  return options
}

function createPlainGradeSelectOptions(sourceGrades: unknown[]) {
  const options: SelectOption[] = [{ value: '', label: '未选择' }]
  const uniqueGrades: string[] = []
  for (const sourceGrade of sourceGrades) {
    const grade = firstText(sourceGrade)
    if (!grade || !isGradeValue(grade) || uniqueGrades.includes(grade)) {
      continue
    }
    uniqueGrades.push(grade)
    options.push({ value: grade, label: grade })
  }
  return options
}

const reviewStatus = computed(() => String(props.review?.status || ''))
const editableSelf = computed(() => (props.actionScope || 'readonly') === 'mine' && reviewStatus.value === 'draft')
const editableManager = computed(() => (props.actionScope || 'readonly') === 'manager' && reviewStatus.value === 'manager_review')
const editableHrbp = computed(() => (props.actionScope || 'readonly') === 'hrbp' && reviewStatus.value === 'hrbp_review')
const canMutateSelfReview = computed(() => (
  editableSelf.value
  && (
    auth.hasButtonPermission(reviewActionButtonPermissions['save-self'])
    || auth.hasButtonPermission(reviewActionButtonPermissions['submit-self'])
  )
))
const canEditObjectiveDimension = computed(() => canMutateSelfReview.value)
const canAddCurrentObjective = computed(() => canMutateSelfReview.value)
const canDeleteCurrentObjective = computed(() => canMutateSelfReview.value)
const canToggleCurrentObjectiveDelete = computed(() => canDeleteCurrentObjective.value && currentObjectives.value.length > 0)
const showCurrentObjectiveDelete = computed(() => canDeleteCurrentObjective.value && currentObjectiveDeleteVisible.value)
const canEditNextObjectives = computed(() => (
  editableSelf.value
  && auth.hasButtonPermission('dingtalk_h5:button:review:next_objective_edit')
))
const canAddNextObjective = computed(() => (
  canEditNextObjectives.value
  && auth.hasButtonPermission('dingtalk_h5:button:review:next_objective_add')
))
const canDeleteNextObjective = computed(() => (
  canEditNextObjectives.value
  && auth.hasButtonPermission('dingtalk_h5:button:review:next_objective_delete')
))
const canToggleNextObjectiveDelete = computed(() => canDeleteNextObjective.value && nextObjectives.value.length > 0)
const showNextObjectiveDelete = computed(() => canDeleteNextObjective.value && nextObjectiveDeleteVisible.value)
const currentObjectiveTotal = computed(() => totalObjectiveScore(localReview))
const nextObjectiveTotal = computed(() => roundScore(nextObjectives.value.reduce((sum, item) => sum + numberValue(item.weight), 0)))
const gradeMismatch = computed(() => Boolean(localReview.managerGrade && localReview.hrbpGrade && localReview.managerGrade !== localReview.hrbpGrade))
const gradeSelectOptions = computed<SelectOption[]>(() => createGroupedGradeSelectOptions(gradeOptions.value))
const plainGradeSelectOptions = computed<SelectOption[]>(() => createPlainGradeSelectOptions(gradeOptions.value))
const flowProgressRows = computed<FlowProgressRow[]>(() => props.review ? createFlowProgressRows(props.review) : [])
const currentProgress = computed<FlowProgressRow | null>(() => {
  const review = props.review
  const currentStep = statusMeta[String(review?.status || '')]?.step ?? 0
  return flowProgressRows.value[currentStep] || flowProgressRows.value[0] || null
})
const processSummary = computed(() => {
  const review = props.review
  const progress = currentProgress.value
  return {
    statusText: review ? reviewStatusText(review) : '-',
    handler: review ? currentAssignee(review, userName) : '-',
    desc: progress ? `${progress.progressText} · ${progress.detail}` : '-',
  }
})

const actionItems = computed<ReviewActionItem[]>(() => {
  const review = props.review
  if (!review) {
    return []
  }
  const statusActions = props.actionScope === 'mine'
    ? mineActionsForStatus(reviewStatus.value)
    : actionsForStatus(reviewStatus.value)
  return statusActions.filter(item => canPerformReviewAction(item.action))
})
const headerActionItems = computed(() => (
  reviewStatus.value === 'employee_confirm'
    ? actionItems.value.filter(item => employeeConfirmHeaderActionSet.has(item.action))
    : []
))
const footerActionItems = computed(() => (
  actionItems.value.filter((item) => {
    if (reviewStatus.value === 'employee_confirm') {
      return !employeeConfirmHeaderActionSet.has(item.action)
    }
    return true
  })
))
const actionConfirmCopy = computed(() => (
  actionConfirmAction.value ? reviewActionConfirmCopy[actionConfirmAction.value] || null : null
))
const actionConfirmNeedsInput = computed(() => (
  actionConfirmAction.value ? actionNeedsRemarkInput(actionConfirmAction.value) : false
))
const actionConfirmNeedsFinalizeFields = computed(() => actionConfirmAction.value === 'finalize')
const actionConfirmInputRequired = computed(() => (
  actionConfirmAction.value ? actionRequiresRemarkInput(actionConfirmAction.value) : false
))
const actionConfirmInputLabel = computed(() => (
  actionConfirmAction.value ? actionRemarkLabel(actionConfirmAction.value) : ''
))
const actionConfirmInputPlaceholder = computed(() => (
  actionConfirmAction.value ? actionRemarkPlaceholder(actionConfirmAction.value) : ''
))
const actionConfirmButtonType = computed<ActionTone>(() => {
  const action = actionConfirmAction.value
  if (action === 'confirm-result' || action === 'finalize') {
    return 'success'
  }
  if (action === 'dispute-result' || action === 'withdraw' || (action?.startsWith('return-') ?? false)) {
    return 'error'
  }
  return 'primary'
})
const activeValueRubricItems = computed(() => valueRubricItems(activeValueRubric.value))

function setProcessPageScrollLocked(locked: boolean) {
  if (typeof document === 'undefined' || typeof window === 'undefined' || processPageScrollLocked === locked) {
    return
  }

  const body = document.body
  const html = document.documentElement
  if (!body || !html) {
    return
  }

  if (locked) {
    processPageScrollTop = window.pageYOffset || html.scrollTop || body.scrollTop || 0
    previousBodyOverflow = body.style.overflow
    previousBodyPosition = body.style.position
    previousBodyTop = body.style.top
    previousBodyWidth = body.style.width
    previousHtmlOverflow = html.style.overflow
    html.style.overflow = 'hidden'
    body.style.overflow = 'hidden'
    body.style.position = 'fixed'
    body.style.top = `-${processPageScrollTop}px`
    body.style.width = '100%'
    processPageScrollLocked = true
    return
  }

  html.style.overflow = previousHtmlOverflow
  body.style.overflow = previousBodyOverflow
  body.style.position = previousBodyPosition
  body.style.top = previousBodyTop
  body.style.width = previousBodyWidth
  processPageScrollLocked = false
  window.scrollTo(0, processPageScrollTop)
}

watch(
  () => props.review,
  (review) => {
    syncLocalReview(review)
    activeFormTab.value = defaultFormTabForReview(review)
    processVisible.value = false
    closeValueRubric()
    cancelActionConfirm()
    currentObjectiveDeleteVisible.value = false
    nextObjectiveDeleteVisible.value = false
  },
  { immediate: true },
)

watch(
  () => props.template,
  () => {
    if (props.review && values.value.length === 0) {
      localReview.values = normalizeValues([], templateValues())
    }
  },
)

watch(processVisible, (visible) => {
  setProcessPageScrollLocked(visible)
})

onBeforeUnmount(() => {
  setProcessPageScrollLocked(false)
})

function createFlowProgressRows(review: PerformanceReview): FlowProgressRow[] {
  const currentStep = statusMeta[String(review.status || '')]?.step ?? 0
  return flowStepDefinitions.map((step, index) => {
    const history = latestHistoryForStep(review, step)
    const isCompletedFinal = review.status === 'completed' && index === currentStep
    const state = index < currentStep || isCompletedFinal ? 'done' : index === currentStep ? 'active' : 'pending'
    const latestAction = String(review.latestAction || '').trim()
    const shouldUseLatestAction = state === 'active' && (returnStatusTextFromAction(latestAction) || latestAction.includes('员工提出异议'))
    const historyAction = String(history?.action || (shouldUseLatestAction ? latestAction : '')).trim()
    const actionParts = historyActionParts(historyAction)
    const returnStatusText = state === 'active' ? returnStatusTextFromAction(actionParts.title || historyAction) : ''
    const stateLabel = returnStatusText ? '已退回' : state === 'done' ? '已完成' : state === 'active' ? '进行中' : '待处理'
    const stateClass = returnStatusText ? 'returned' : state
    const disputeReason = reviewDisputeReasonForStep(review, step)
    const historyReason = historyReasonForStep(review, step, history, actionParts)
    const reason = disputeReason && (!historyReason || historyReason === missingHistoryReasonText('员工提出异议'))
      ? disputeReason
      : historyReason
    const reasonLabel = disputeReason && reason === disputeReason
      ? '异议原因'
      : historyReasonLabel(actionParts.title || history?.action)
    const detailMeta = history ? historyDetailMeta(history) : ''
    const detail = historyAction
      ? [actionParts.title || historyAction, detailMeta].filter(Boolean).join(' · ')
      : disputeReason && state === 'active'
        ? `员工提出异议，等待 ${stepActorName(review, step)} 处理`
        : state === 'active'
          ? `等待 ${stepActorName(review, step)} 处理`
          : state === 'done'
            ? '已进入下一流程节点'
            : step.actor === 'system'
              ? '待归档完成'
              : `待 ${step.role} 处理`

    return {
      ...step,
      index,
      indexText: String(index + 1).padStart(2, '0'),
      progressText: `${String(index + 1).padStart(2, '0')}/${String(flowStepDefinitions.length).padStart(2, '0')}`,
      actorName: stepActorName(review, step),
      state,
      stateClass,
      stateLabel,
      returnStatusText,
      detail,
      reason,
      reasonLabel,
      history,
    }
  })
}

function stepActorName(review: PerformanceReview, step: FlowStepDefinition) {
  if (step.actor === 'employee') {
    return userName(review.employeeId)
  }
  if (step.actor === 'manager') {
    return userName(review.managerId)
  }
  if (step.actor === 'hrbp') {
    return userName(review.hrbpReviewerId || review.hrbpId)
  }
  return '系统'
}

function userName(id: unknown) {
  const key = String(id || '').trim()
  if (!key) {
    return '-'
  }
  const matched = (props.users || []).find(user => user.id === key)
  if (matched?.name) {
    return matched.name
  }
  if (sameUserId(auth.user?.id, key)) {
    return auth.user?.name || key
  }
  return key
}

function reviewHistories(review: PerformanceReview): ReviewHistoryRecord[] {
  return Array.isArray(review.history) ? review.history.filter(isRecord) as ReviewHistoryRecord[] : []
}

function latestHistoryForStep(review: PerformanceReview, step: FlowStepDefinition) {
  return reviewHistories(review).slice().reverse().find(item =>
    step.historyKeywords.some(keyword => String(item.action || '').includes(keyword)),
  )
}

function historyMatchesStep(history: ReviewHistoryRecord | undefined, step: FlowStepDefinition) {
  return step.historyKeywords.some(keyword => String(history?.action || '').includes(keyword))
}

function historyActionParts(action: unknown) {
  const text = String(action || '').trim()
  const fullSeparatorIndex = text.indexOf('：')
  const halfSeparatorIndex = text.indexOf(':')
  const separatorIndex = fullSeparatorIndex >= 0 ? fullSeparatorIndex : halfSeparatorIndex
  if (separatorIndex < 0) {
    return { title: text, reason: '' }
  }
  return {
    title: text.slice(0, separatorIndex).trim(),
    reason: text.slice(separatorIndex + 1).trim(),
  }
}

function historyActionNeedsReason(title: unknown) {
  const text = String(title || '').trim()
  return text.startsWith('撤销') || text.startsWith('撤回') || text.startsWith('退回') || text.includes('异议')
}

function missingHistoryReasonText(title: unknown) {
  const text = String(title || '').trim()
  if (text.startsWith('撤销')) {
    return '未记录撤销理由'
  }
  if (text.startsWith('撤回')) {
    return '未记录撤回理由'
  }
  if (text.startsWith('退回')) {
    return '未记录退回原因'
  }
  if (text.includes('异议')) {
    return '未记录异议原因'
  }
  return ''
}

function historyReasonLabel(title: unknown) {
  const text = String(title || '').trim()
  if (text.startsWith('撤销')) {
    return '撤销理由'
  }
  if (text.startsWith('撤回')) {
    return '撤回理由'
  }
  if (text.startsWith('退回')) {
    return '退回原因'
  }
  if (text.includes('异议')) {
    return '异议原因'
  }
  return '理由'
}

function historyReasonForStep(review: PerformanceReview, step: FlowStepDefinition, selectedHistory: ReviewHistoryRecord | undefined, actionParts: { title: string, reason: string }) {
  if (actionParts.reason) {
    return actionParts.reason
  }
  const selectedTitle = actionParts.title || selectedHistory?.action || ''
  const sameActionWithReason = reviewHistories(review).slice().reverse().find((item) => {
    if (!historyMatchesStep(item, step)) {
      return false
    }
    const parts = historyActionParts(item.action)
    return parts.reason && (!selectedTitle || parts.title === selectedTitle)
  })
  if (sameActionWithReason) {
    return historyActionParts(sameActionWithReason.action).reason
  }
  return historyActionNeedsReason(selectedTitle) ? missingHistoryReasonText(selectedTitle) : ''
}

function reviewDisputeReasonForStep(review: PerformanceReview, step: FlowStepDefinition) {
  if (step.status !== 'hrbp_review') {
    return ''
  }
  if (review.employeeConfirmResult !== 'disputed') {
    return ''
  }
  return String(review.employeeConfirmComment || '').trim()
}

function historyDetailMeta(history: ReviewHistoryRecord) {
  return [history.by || 'system', formatHistoryTime(history.at)].filter(Boolean).join(' · ')
}

function formatHistoryTime(at: unknown) {
  if (!at) {
    return ''
  }
  const date = new Date(at as string | number)
  if (Number.isNaN(date.getTime())) {
    return ''
  }
  const pad = (value: number) => String(value).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`
}

function latestReviewAction(review: PerformanceReview) {
  const latestAction = String(review.latestAction || '').trim()
  if (latestAction) {
    return latestAction
  }
  const histories = reviewHistories(review)
  return String(histories[histories.length - 1]?.action || '').trim()
}

function returnStatusTextFromAction(action: unknown) {
  const title = String(action || '').split(/[：:]/)[0].trim()
  if (title.startsWith('退回员工')) {
    return '退回员工'
  }
  if (title.startsWith('退回上级')) {
    return '退回上级'
  }
  if (title.startsWith('退回 HRBP') || title.startsWith('退回HRBP')) {
    return '退回HRBP'
  }
  return ''
}

function reviewStatusText(review: PerformanceReview) {
  return returnStatusTextFromAction(latestReviewAction(review)) || baseReviewStatusText(review)
}

function statusChipClass(review: PerformanceReview) {
  if (returnStatusTextFromAction(latestReviewAction(review))) {
    return 'status-chip--returned'
  }
  return `status-chip--${reviewStatusType(review)}`
}

function syncLocalReview(review: PerformanceReview | null) {
  for (const key of Object.keys(localReview)) {
    delete localReview[key]
  }
  if (!review) {
    return
  }
  Object.assign(localReview, cloneData(review))
  localReview.objectives = normalizeObjectives(firstArray(localReview.objectives, localReview.currentObjectives, localReview.targets), 'current')
  localReview.nextObjectives = normalizeObjectives(firstArray(localReview.nextObjectives, localReview.nextTargets), 'next')
  localReview.values = normalizeValues(firstArray(localReview.values, localReview.valueScores), templateValues())
}

function cloneData<T>(value: T): T {
  return JSON.parse(JSON.stringify(value))
}

function firstArray(...valuesToCheck: unknown[]) {
  for (const value of valuesToCheck) {
    if (Array.isArray(value)) {
      return value
    }
  }
  return []
}

function isRecord(value: unknown): value is FormRecord {
  return Boolean(value && typeof value === 'object')
}

function templateValues() {
  const template = props.template
  return isRecord(template) && Array.isArray(template.values) ? template.values as FormRecord[] : []
}

function normalizeObjectives(items: unknown[], kind: 'current' | 'next'): ReviewObjective[] {
  return items.map((item) => {
    const source = isRecord(item) ? item : {}
    if (kind === 'next') {
      return {
        ...source,
        target: String(source.target || ''),
        weight: source.weight ?? 0,
      }
    }
    return {
      ...source,
      target: String(source.target || ''),
      weight: source.weight ?? 0,
      completion: source.completion ?? '',
      result: String(source.result || ''),
    }
  })
}

function normalizeValues(items: unknown[], templateItems: FormRecord[]): ReviewValueScore[] {
  const sourceItems = items.length > 0 ? items : templateItems
  return sourceItems.map((item, index) => {
    const source = isRecord(item) ? item : {}
    const template = templateItems.find(value => String(value.id || '') === String(source.id || '')) || {}
    return {
      ...template,
      ...source,
      id: String(source.id || template.id || `value-${index + 1}`),
      name: String(source.name || template.name || `价值观 ${index + 1}`),
      definition: String(source.definition || template.definition || ''),
      self: source.self ?? '',
      manager: source.manager ?? '',
      hrbp: source.hrbp ?? '',
    }
  })
}

function createObjective(): ReviewObjective {
  return {
    target: '',
    weight: 0,
    completion: '',
    result: '',
  }
}

function createNextObjective(): ReviewObjective {
  return {
    target: '',
    weight: 0,
  }
}

function addCurrentObjective() {
  if (!canAddCurrentObjective.value) {
    return
  }
  currentObjectives.value.push(createObjective())
}

function toggleCurrentObjectiveDelete() {
  if (!canToggleCurrentObjectiveDelete.value) {
    return
  }
  currentObjectiveDeleteVisible.value = !currentObjectiveDeleteVisible.value
}

function addNextObjective() {
  nextObjectives.value.push(createNextObjective())
}

function toggleNextObjectiveDelete() {
  nextObjectiveDeleteVisible.value = !nextObjectiveDeleteVisible.value
}

function confirmRemoveCurrentObjective(index: number) {
  if (!canDeleteCurrentObjective.value) {
    return
  }
  confirmObjectiveDelete(index, '目标', () => {
    currentObjectives.value.splice(index, 1)
    if (currentObjectives.value.length === 0) {
      currentObjectiveDeleteVisible.value = false
    }
  })
}

function confirmRemoveNextObjective(index: number) {
  confirmObjectiveDelete(index, '下月目标', () => {
    nextObjectives.value.splice(index, 1)
    if (nextObjectives.value.length === 0) {
      nextObjectiveDeleteVisible.value = false
    }
  })
}

function confirmObjectiveDelete(index: number, label: string, onConfirm: () => void) {
  uni.showModal({
    title: `删除${label}`,
    content: `确认删除${label} ${index + 1}？删除后需要保存才会生效。`,
    confirmText: '删除',
    confirmColor: '#e34d59',
    success: (res) => {
      if (res.confirm) {
        onConfirm()
      }
    },
  })
}

function readInputValue(event: unknown) {
  if (typeof event === 'string' || typeof event === 'number') {
    return event
  }
  const inputEvent = event as { detail?: { value?: unknown }, target?: { value?: unknown } }
  return inputEvent.detail?.value ?? inputEvent.target?.value ?? ''
}

function fieldText(value: unknown) {
  return value === null || value === undefined ? '' : String(value)
}

function viewportWidth() {
  try {
    const info = uni.getSystemInfoSync()
    return Number(info.windowWidth || info.screenWidth || 0)
  }
  catch {
    return 0
  }
}

function visualTextLength(text: string) {
  let length = 0
  for (const char of text) {
    length += char.charCodeAt(0) <= 0xFF ? 0.55 : 1
  }
  return length
}

function estimateTextareaRows(value: unknown, charsPerRow: number) {
  const text = fieldText(value).replace(/\r\n/g, '\n')
  if (!text.trim()) {
    return SELF_SUMMARY_MIN_ROWS
  }
  const rows = text.split('\n').reduce((total, line) => {
    return total + Math.max(1, Math.ceil(visualTextLength(line) / charsPerRow))
  }, 0)
  return Math.min(SELF_SUMMARY_MAX_ROWS, Math.max(SELF_SUMMARY_MIN_ROWS, rows))
}

function textareaEventText(event: unknown) {
  return fieldText(readInputValue(event))
}

function setTextareaValue(record: FormRecord, field: string, event: unknown) {
  record[field] = textareaEventText(event)
}

function sanitizeNumberInput(event: unknown) {
  const text = fieldText(readInputValue(event)).trim()
  if (!text) {
    return ''
  }
  const numericText = text.replace(/[^\d.]/g, '')
  const dotIndex = numericText.indexOf('.')
  if (dotIndex === -1) {
    return numericText
  }
  return `${numericText.slice(0, dotIndex + 1)}${numericText.slice(dotIndex + 1).replace(/\./g, '')}`
}

function setNumberField(record: FormRecord, field: string, event: unknown) {
  record[field] = sanitizeNumberInput(event)
}

const selfSummaryTextareaRows = computed(() => {
  const width = viewportWidth()
  const charsPerRow = width > 0 && width <= 768
    ? SELF_SUMMARY_MOBILE_CHARS_PER_ROW
    : SELF_SUMMARY_DESKTOP_CHARS_PER_ROW
  return estimateTextareaRows(localReview.selfSummary, charsPerRow)
})

const selfSummaryTextareaStyle = computed<Record<string, string>>(() => {
  const heightRpx = Math.ceil(selfSummaryTextareaRows.value * REVIEW_TEXTAREA_LINE_HEIGHT_RPX + REVIEW_TEXTAREA_VERTICAL_PADDING_RPX)
  const maxHeightRpx = Math.ceil(SELF_SUMMARY_MAX_ROWS * REVIEW_TEXTAREA_LINE_HEIGHT_RPX + REVIEW_TEXTAREA_VERTICAL_PADDING_RPX)
  return {
    '--review-textarea-height': `${heightRpx}rpx`,
    '--review-textarea-max-height': `${maxHeightRpx}rpx`,
  }
})

function reviewFormTabClass(key: string) {
  return ['review-form-tab', activeFormTab.value === key ? 'active' : ''].filter(Boolean).join(' ')
}

function defaultFormTabForReview(review: PerformanceReview | null) {
  if (review?.status === 'hr_final') {
    return 'hrbp'
  }
  return 'currentTargets'
}

function miniButtonClass(extraClass = '') {
  return ['dt-mini-btn', extraClass].filter(Boolean).join(' ')
}

function currentObjectiveDeleteButtonClass() {
  return miniButtonClass(currentObjectiveDeleteVisible.value ? 'dt-mini-btn--danger' : '')
}

function nextObjectiveDeleteButtonClass() {
  return miniButtonClass(nextObjectiveDeleteVisible.value ? 'dt-mini-btn--danger' : '')
}

function setValueScore(item: ReviewValueScore, field: ValueScoreField, event: unknown) {
  item[field] = sanitizeNumberInput(event)
}

function valueRubricItems(item: ReviewValueScore | null) {
  return Array.isArray(item?.rubric) ? item.rubric : []
}

function hasValueRubric(item: ReviewValueScore) {
  return valueRubricItems(item).length > 0
}

function openValueRubric(item: ReviewValueScore) {
  if (!hasValueRubric(item)) {
    return
  }
  activeValueRubric.value = item
  valueRubricVisible.value = true
}

function closeValueRubric() {
  valueRubricVisible.value = false
  activeValueRubric.value = null
}

function numberValue(value: unknown) {
  if (value === null || value === undefined || String(value).trim() === '') {
    return 0
  }
  const number = Number(value)
  return Number.isFinite(number) ? number : 0
}

function roundScore(value: number) {
  return Math.round(value * 10) / 10
}

function objectiveScore(item: ReviewObjective) {
  const completion = Number(item.completion)
  const weight = Number(item.weight)
  if (!Number.isFinite(completion) || !Number.isFinite(weight)) {
    return 0
  }
  return roundScore(weight * completion / 100)
}

function totalObjectiveScore(review: FormRecord) {
  return roundScore((review.objectives || []).reduce((sum: number, item: ReviewObjective) => sum + objectiveScore(item), 0))
}

function valueTotal(field: ValueScoreField) {
  const nums = values.value.map(item => Number(item[field])).filter(Number.isFinite)
  if (nums.length === 0) {
    return '-'
  }
  return roundScore(nums.reduce((sum, item) => sum + item, 0))
}

function effectiveGrade(review: FormRecord) {
  return String(review.finalGrade || review.hrbpGrade || '')
}

function hasFilledRequiredValue(value: unknown) {
  if (value === null || value === undefined) {
    return false
  }
  if (typeof value === 'number') {
    return Number.isFinite(value)
  }
  return String(value).trim() !== ''
}

function isNumericText(value: unknown) {
  if (typeof value === 'number') {
    return Number.isFinite(value)
  }
  const text = String(value || '').trim()
  if (!text) {
    return false
  }
  let digitCount = 0
  let dotCount = 0
  for (const char of text) {
    if (char === '.') {
      dotCount += 1
      if (dotCount > 1) {
        return false
      }
      continue
    }
    if (char < '0' || char > '9') {
      return false
    }
    digitCount += 1
  }
  return digitCount > 0 && Number.isFinite(Number(text))
}

function numberRangeValidationMessage(value: unknown, label: string, min: number, max: number, required = false) {
  if (value === null || value === undefined || String(value).trim() === '') {
    return required ? `${label}必须是数字` : ''
  }
  if (!isNumericText(value)) {
    return `${label}必须是数字`
  }
  const number = Number(value)
  if (number < min || number > max) {
    return `${label}必须在 ${min}-${max} 之间`
  }
  return ''
}

function optionalPercentValidationMessage(value: unknown, label: string) {
  return numberRangeValidationMessage(value, label, OBJECTIVE_NUMBER_MIN, OBJECTIVE_COMPLETION_MAX)
}

function objectiveWeightValidationMessage(items: ReviewObjective[], label: string) {
  if (!items.length) {
    return ''
  }
  let total = 0
  for (let index = 0; index < items.length; index += 1) {
    const message = numberRangeValidationMessage(
      items[index]?.weight,
      `${label} ${index + 1} 的权重`,
      OBJECTIVE_NUMBER_MIN,
      OBJECTIVE_WEIGHT_MAX,
      true,
    )
    if (message) {
      return message
    }
    total += Number(items[index]?.weight)
  }
  if (total > 100) {
    return `${label}权重合计不能大于 100`
  }
  return ''
}

function validateSelfObjectiveNumbers(review: FormRecord) {
  return objectiveWeightValidationMessage(review.objectives || [], '本月目标')
    || currentObjectiveCompletionValidationMessage(review.objectives || [])
    || objectiveWeightValidationMessage(review.nextObjectives || [], '下月目标')
}

function currentObjectiveCompletionValidationMessage(items: ReviewObjective[]) {
  for (let index = 0; index < items.length; index += 1) {
    const message = optionalPercentValidationMessage(items[index]?.completion, `本月目标 ${index + 1} 的完成度`)
    if (message) {
      return message
    }
  }
  return ''
}

function valueScoreValidationMessage(value: unknown, label: string) {
  return numberRangeValidationMessage(value, label, VALUE_SCORE_MIN, VALUE_SCORE_MAX)
}

function validateValueScores(items: ReviewValueScore[], field: ValueScoreField, label: string) {
  for (let index = 0; index < items.length; index += 1) {
    const message = valueScoreValidationMessage(items[index]?.[field], `${label} ${index + 1}`)
    if (message) {
      return message
    }
  }
  return ''
}

function validateReviewNumbers(action: ReviewActionKey) {
  if (action === 'save-self' || action === 'submit-self') {
    return validateSelfObjectiveNumbers(localReview)
      || validateValueScores(localReview.values || [], 'self', '价值观自评')
  }
  if (action === 'submit-manager') {
    return validateValueScores(localReview.values || [], 'manager', '上级价值观评分')
  }
  if (action === 'submit-hrbp') {
    return validateValueScores(localReview.values || [], 'hrbp', 'HRBP价值观评分')
  }
  return ''
}

function hasRequiredCurrentObjectives(review: FormRecord) {
  const objectives: ReviewObjective[] = review.objectives || []
  return objectives.length > 0 && objectives.every(item => (
    String(item.target || '').trim() !== ''
    && hasFilledRequiredValue(item.completion)
    && String(item.result || '').trim() !== ''
  ))
}

function hasRequiredSelfValues(review: FormRecord) {
  const reviewValues: ReviewValueScore[] = review.values || []
  return reviewValues.length > 0 && reviewValues.every(item => hasFilledRequiredValue(item.self))
}

function hasRequiredNextObjectives(review: FormRecord) {
  const reviewNextObjectives: ReviewObjective[] = review.nextObjectives || []
  return reviewNextObjectives.length > 0 && reviewNextObjectives.every(item => String(item.target || '').trim() !== '')
}

function validateSelfSubmitReview(review: FormRecord) {
  const missing: string[] = []
  if (!hasRequiredCurrentObjectives(review)) {
    missing.push('本月目标')
  }
  if (String(review.selfSummary || '').trim() === '') {
    missing.push('思考总结')
  }
  if (!hasRequiredSelfValues(review)) {
    missing.push('价值观自评')
  }
  if (!hasRequiredNextObjectives(review)) {
    missing.push('下月目标')
  }
  return missing.length ? `请完善：${missing.join('、')}` : ''
}

function hasRequiredManagerValues(review: FormRecord) {
  const reviewValues: ReviewValueScore[] = review.values || []
  return reviewValues.length > 0 && reviewValues.every(item => hasFilledRequiredValue(item.manager))
}

function validateManagerSubmitReview(review: FormRecord) {
  const missing: string[] = []
  if (!hasFilledRequiredValue(review.managerGrade)) {
    missing.push('上级分档')
  }
  if (String(review.managerComment || '').trim() === '') {
    missing.push('评价内容')
  }
  if (!hasRequiredManagerValues(review)) {
    missing.push('上级价值观评分')
  }
  return missing.length ? `请完善：${missing.join('、')}` : ''
}

function hasRequiredHrbpValues(review: FormRecord) {
  const reviewValues: ReviewValueScore[] = review.values || []
  return reviewValues.length > 0 && reviewValues.every(item => hasFilledRequiredValue(item.hrbp))
}

function validateHrbpSubmitReview(review: FormRecord) {
  const missing: string[] = []
  if (!hasFilledRequiredValue(review.hrbpGrade)) {
    missing.push('HRBP分档')
  }
  if (String(review.hrbpComment || '').trim() === '') {
    missing.push('评价内容')
  }
  if (!hasRequiredHrbpValues(review)) {
    missing.push('HRBP价值观评分')
  }
  if (missing.length) {
    return { message: `请完善：${missing.join('、')}`, modal: false }
  }
  const managerGrade = String(review.managerGrade || '').trim()
  const hrbpGrade = String(review.hrbpGrade || '').trim()
  if (!managerGrade) {
    return { title: '分档不一致', message: '上级分档为空，不能提交给员工确认。', modal: true }
  }
  if (hrbpGrade && hrbpGrade !== managerGrade) {
    return { title: '分档不一致', message: `HRBP分档需与上级分档一致，当前上级分档为「${managerGrade}」，HRBP分档为「${hrbpGrade}」。`, modal: true }
  }
  return null
}

function validateAction(action: ReviewActionKey) {
  if (action === 'save-self') {
    return validateReviewNumbers(action)
  }
  if (action === 'submit-self') {
    return validateReviewNumbers(action) || validateSelfSubmitReview(localReview)
  }
  if (action === 'submit-manager') {
    return validateReviewNumbers(action) || validateManagerSubmitReview(localReview)
  }
  if (action === 'submit-hrbp') {
    const validation = validateReviewNumbers(action) || validateHrbpSubmitReview(localReview)
    return validation
  }
  return ''
}

function normalizePayloadNumber(value: unknown) {
  if (typeof value === 'number') {
    return Number.isFinite(value) ? value : ''
  }
  const text = fieldText(value).trim()
  if (!text) {
    return ''
  }
  return isNumericText(text) ? Number(text) : text
}

function normalizePayloadObjective(item: ReviewObjective) {
  return {
    ...item,
    weight: normalizePayloadNumber(item.weight),
    completion: normalizePayloadNumber(item.completion),
  }
}

function normalizePayloadNextObjective(item: ReviewObjective) {
  return {
    ...item,
    weight: normalizePayloadNumber(item.weight),
  }
}

function normalizePayloadValueScore(item: ReviewValueScore) {
  return {
    ...item,
    self: normalizePayloadNumber(item.self),
    manager: normalizePayloadNumber(item.manager),
    hrbp: normalizePayloadNumber(item.hrbp),
  }
}

function reviewPayload(review: FormRecord = localReview) {
  return {
    objectives: (review.objectives || []).map(normalizePayloadObjective),
    nextObjectives: (review.nextObjectives || []).map(normalizePayloadNextObjective),
    values: (review.values || []).map(normalizePayloadValueScore),
    selfSummary: review.selfSummary,
    managerComment: review.managerComment,
    managerGrade: review.managerGrade,
    hrbpComment: review.hrbpComment,
    hrbpGrade: review.hrbpGrade,
    employeeConfirmResult: review.employeeConfirmResult,
    employeeConfirmComment: review.employeeConfirmComment,
    finalGrade: review.finalGrade,
    finalNote: review.finalNote,
  }
}

function canPerformReviewAction(action: ReviewActionKey) {
  const scopedActions = reviewActionScopeActions[props.actionScope || 'readonly']
  return scopedActions.has(action)
    && auth.hasApiPermission(reviewActionApiPermissions[action])
    && auth.hasButtonPermission(reviewActionButtonPermissions[action])
}

function mineActionsForStatus(status: string): ReviewActionItem[] {
  if (status === 'draft' || status === 'employee_confirm') {
    return actionsForStatus(status)
  }
  if (status === 'manager_review') {
    return [
      { action: 'withdraw', label: '撤回提交', type: 'warning', plain: true },
    ]
  }
  return []
}

function actionsForStatus(status: string): ReviewActionItem[] {
  if (status === 'draft') {
    return [
      { action: 'save-self', label: '保存', type: 'default' },
      { action: 'submit-self', label: '提交自评', type: 'primary' },
    ]
  }
  if (status === 'manager_review') {
    return [
      { action: 'return-employee', label: '退回员工', type: 'error', plain: true },
      { action: 'submit-manager', label: '提交给HRBP', type: 'primary' },
    ]
  }
  if (status === 'hrbp_review') {
    return [
      { action: 'return-manager', label: '退回上级', type: 'error', plain: true },
      { action: 'submit-hrbp', label: '提交给员工', type: 'primary' },
    ]
  }
  if (status === 'employee_confirm') {
    return [
      { action: 'confirm-result', label: '确认结果', type: 'success' },
      { action: 'dispute-result', label: '提出异议', type: 'error', plain: true },
    ]
  }
  if (status === 'hr_final') {
    return [
      { action: 'return-hrbp', label: '退回HRBP', type: 'error', plain: true },
      { action: 'finalize', label: '确认归档', type: 'success' },
    ]
  }
  return []
}

function actionNeedsRemark(action: ReviewActionKey) {
  return action === 'withdraw' || action.startsWith('return-')
}

function actionNeedsRemarkInput(action: ReviewActionKey) {
  return action === 'confirm-result' || action === 'dispute-result' || actionNeedsRemark(action)
}

function actionRequiresRemarkInput(action: ReviewActionKey) {
  return action === 'dispute-result' || actionNeedsRemark(action)
}

function actionRemarkLabel(action: ReviewActionKey) {
  if (action === 'confirm-result') {
    return '确认说明'
  }
  if (action === 'dispute-result') {
    return '异议原因'
  }
  return action === 'withdraw' ? '撤回理由' : '退回原因'
}

function actionRemarkPlaceholder(action: ReviewActionKey) {
  if (action === 'confirm-result') {
    return '请输入确认说明（可选）'
  }
  return `请输入${actionRemarkLabel(action)}`
}

function defaultActionRemark(action: ReviewActionKey) {
  if (action === 'confirm-result' || action === 'dispute-result') {
    return String(localReview.employeeConfirmComment || '').trim()
  }
  return ''
}

function resetActionConfirmFields() {
  actionConfirmAction.value = null
  actionConfirmRemark.value = ''
  actionConfirmFinalGrade.value = ''
  actionConfirmFinalNote.value = ''
}

function clearActionConfirmResetTimer() {
  if (!actionConfirmResetTimer) {
    return
  }
  clearTimeout(actionConfirmResetTimer)
  actionConfirmResetTimer = null
}

function scheduleActionConfirmReset() {
  clearActionConfirmResetTimer()
  actionConfirmResetTimer = setTimeout(() => {
    actionConfirmResetTimer = null
    resetActionConfirmFields()
  }, actionConfirmCloseResetDelay)
}

function closeActionConfirm(result: ReviewActionConfirmResult) {
  const resolver = actionConfirmResolver
  actionConfirmResolver = null
  actionConfirmVisible.value = false
  scheduleActionConfirmReset()
  resolver?.(result)
}

function cancelActionConfirm() {
  closeActionConfirm({ confirmed: false, remark: '' })
}

function submitActionConfirm() {
  const action = actionConfirmAction.value
  if (!action) {
    closeActionConfirm({ confirmed: false, remark: '' })
    return
  }
  if (action === 'finalize') {
    const finalGrade = String(actionConfirmFinalGrade.value || '').trim()
    const finalNote = String(actionConfirmFinalNote.value || '').trim()
    if (!finalGrade) {
      showToast('请选择最终分档')
      return
    }
    closeActionConfirm({ confirmed: true, remark: finalNote, finalGrade })
    return
  }
  const remark = String(actionConfirmRemark.value || '').trim()
  if (actionRequiresRemarkInput(action) && !remark) {
    showToast(`请填写${actionRemarkLabel(action)}`)
    return
  }
  closeActionConfirm({ confirmed: true, remark })
}

function openActionConfirm(action: ReviewActionKey): Promise<ReviewActionConfirmResult> {
  const copy = reviewActionConfirmCopy[action]
  if (!copy) {
    return Promise.resolve({ confirmed: true, remark: '' })
  }
  if (actionConfirmResolver) {
    closeActionConfirm({ confirmed: false, remark: '' })
  }
  clearActionConfirmResetTimer()
  actionConfirmAction.value = action
  actionConfirmRemark.value = defaultActionRemark(action)
  actionConfirmFinalGrade.value = action === 'finalize' ? effectiveGrade(localReview) : ''
  actionConfirmFinalNote.value = action === 'finalize' ? String(localReview.finalNote || '').trim() : ''
  actionConfirmVisible.value = true
  return new Promise<ReviewActionConfirmResult>((resolve) => {
    actionConfirmResolver = resolve
  })
}

function confirmReviewAction(action: ReviewActionKey): Promise<ReviewActionConfirmResult> {
  return openActionConfirm(action)
}

function showValidationModal(title: string, content: string) {
  return new Promise<boolean>((resolve) => {
    uni.showModal({
      title,
      content,
      showCancel: false,
      confirmText: '知道了',
      confirmColor: '#1677ff',
      success: () => resolve(true),
      fail: () => resolve(false),
    })
  })
}

function showToast(title: string) {
  uni.showToast({
    title,
    icon: 'none',
  })
}

async function runAction(action: ReviewActionKey) {
  const review = props.review
  if (!review?.id) {
    return
  }
  if (!canPerformReviewAction(action)) {
    showToast('无权限操作')
    return
  }

  const validation = validateAction(action)
  if (validation) {
    if (typeof validation === 'object' && validation.modal) {
      await showValidationModal(validation.title || '提示', validation.message)
    }
    else {
      showToast(typeof validation === 'string' ? validation : validation.message)
    }
    return
  }

  const confirmedAction = await confirmReviewAction(action)
  if (!confirmedAction.confirmed) {
    return
  }

  const remark = confirmedAction.remark
  if (action === 'confirm-result') {
    localReview.employeeConfirmResult = 'confirmed'
    localReview.employeeConfirmComment = remark
  }
  if (action === 'dispute-result') {
    localReview.employeeConfirmResult = 'disputed'
    localReview.employeeConfirmComment = remark
  }
  if (action === 'finalize') {
    localReview.finalGrade = confirmedAction.finalGrade || effectiveGrade(localReview)
    localReview.finalNote = remark
  }
  const payload: ReviewActionRequest = {
    action,
    remark,
    ...reviewPayload(),
  }
  if (actionNeedsRemark(action)) {
    payload.returnReason = remark
  }
  if (action === 'withdraw') {
    const withdrawTargetStatus = resolveWithdrawTargetStatus(review.status)
    payload.fromStatus = review.status
    payload.targetStatus = withdrawTargetStatus
  }

  if (!props.submitReviewAction) {
    showToast('当前页面不支持该操作')
    return
  }

  const updatedReview = await props.submitReviewAction(review.id, action, payload)
  emit('action-success', updatedReview)
  uni.showToast({
    title: action === 'save-self' ? '已保存' : '处理成功',
    icon: 'success',
  })
}
</script>

<template>
  <view class="review-detail-shell" :class="{ 'review-detail-shell--with-actions': review && footerActionItems.length }">
    <view class="review-form detail detail--floating-card">
      <view v-if="!review" class="empty">
        <u-icon name="file-text" size="72" color="#c8c9cc" />
        <text class="empty__title">
          暂无绩效单
        </text>
        <text class="empty__desc">
          当前视图还没有可展示的绩效记录。
        </text>
      </view>

      <template v-else>
        <view class="detail__head">
          <view class="detail__title-block">
            <view class="detail-title-row">
              <text class="detail__title">
                {{ review.period || '-' }} 月度考评
              </text>
              <view class="mobile-process-button">
                <u-button
                  custom-class="dt-mini-btn process-title-button"
                  :custom-style="miniButtonCustomStyle"
                  @click="processVisible = true"
                >
                  查看流程进度
                </u-button>
              </view>
            </view>
            <text class="detail__desc">
              {{ review.employeeName || userName(review.employeeId) }} · 当前处理人 {{ currentAssignee(review, userName) }}
            </text>
          </view>
          <view class="detail__head-actions">
            <view class="detail-status-tag">
              <u-tag :text="reviewStatusText(review)" :type="reviewStatusType(review)" mode="light" />
            </view>
            <view v-if="headerActionItems.length" class="detail-top-actions">
              <view
                v-for="item in headerActionItems"
                :key="item.action"
                class="detail-top-action-wrap"
              >
                <u-button
                  custom-class="dt-mini-btn detail-top-action"
                  :type="item.type"
                  :plain="item.plain"
                  :loading="detailLoading"
                  :custom-style="miniButtonCustomStyle"
                  @click="runAction(item.action)"
                >
                  {{ item.label }}
                </u-button>
              </view>
            </view>
          </view>
        </view>

        <view class="process-summary">
          <view class="process-summary-main">
            <text class="process-kicker">
              当前流程状态
            </text>
            <view class="process-status-line">
              <text class="status-chip" :class="[statusChipClass(review)]">
                {{ processSummary.statusText }}
              </text>
              <text v-if="currentProgress" class="process-state-badge" :class="[currentProgress.stateClass]">
                {{ currentProgress.stateLabel }}
              </text>
              <text class="process-handler">
                当前处理人 {{ processSummary.handler }}
              </text>
            </view>
            <text class="process-desc">
              {{ processSummary.desc }}
            </text>
          </view>
          <u-button
            custom-class="dt-mini-btn process-help-btn"
            :custom-style="miniButtonCustomStyle"
            @click="processVisible = true"
          >
            查看流程进度
          </u-button>
        </view>

        <view v-if="processVisible && currentProgress" class="process-modal-mask" @click="processVisible = false" @touchmove.stop.prevent>
          <view class="process-modal" @click.stop @touchmove.stop>
            <view class="process-modal-head">
              <view class="process-modal-title-copy">
                <text class="process-modal-title">
                  流程进度
                </text>
                <text class="process-modal-subtitle">
                  当前：{{ reviewStatusText(review) }} · 处理人 {{ processSummary.handler }}
                </text>
              </view>
              <u-button
                custom-class="process-modal-close"
                :custom-style="processCloseButtonCustomStyle"
                @click="processVisible = false"
              >
                ×
              </u-button>
            </view>

            <view class="process-current-card compact">
              <text class="process-current-index">
                {{ currentProgress.indexText }}
              </text>
              <view class="process-current-main">
                <view class="process-current-title-line">
                  <text class="process-current-label">
                    当前节点
                  </text>
                  <text class="process-current-title">
                    {{ currentProgress.title }}
                  </text>
                  <text class="process-current-progress">
                    {{ currentProgress.progressText }}
                  </text>
                </view>
                <view class="process-current-meta">
                  <text class="status-chip" :class="[statusChipClass(review)]">
                    {{ reviewStatusText(review) }}
                  </text>
                  <text class="process-state-badge" :class="[currentProgress.stateClass]">
                    {{ currentProgress.stateLabel }}
                  </text>
                  <text class="process-current-handler">
                    处理人 {{ processSummary.handler }}
                  </text>
                </view>
              </view>
            </view>

            <view class="process-timeline">
              <view v-for="step in flowProgressRows" :key="step.status" class="process-step-row" :class="step.state">
                <text class="process-step-index">
                  {{ step.indexText }}
                </text>
                <view class="process-step-main">
                  <view class="process-step-title-line">
                    <text class="process-step-title">
                      {{ step.title }}
                    </text>
                    <text class="process-step-role">
                      {{ step.role }}
                    </text>
                    <text class="process-state-badge" :class="[step.stateClass]">
                      {{ step.stateLabel }}
                    </text>
                  </view>
                  <text class="process-step-desc">
                    {{ step.detail }}
                  </text>
                  <view v-if="step.reason" class="process-step-reason">
                    <text class="process-step-reason-label">
                      {{ step.reasonLabel }}
                    </text>
                    <text class="process-step-reason-text">
                      {{ step.reason }}
                    </text>
                  </view>
                  <text class="process-step-help">
                    {{ step.desc }}
                  </text>
                </view>
              </view>
            </view>

            <view class="process-modal-actions">
              <u-button
                custom-class="dt-mini-btn dt-mini-btn--primary"
                :custom-style="miniButtonCustomStyle"
                @click="processVisible = false"
              >
                知道了
              </u-button>
            </view>
          </view>
        </view>

        <!-- <view class="field-list">
        <view v-for="[label, value] in fields" :key="label" class="field">
          <text class="field__label">
            {{ label }}
          </text>
          <text class="field__value">
            {{ value }}
          </text>
        </view>
      </view> -->

        <scroll-view class="review-form-tabs" scroll-x>
          <u-button
            v-for="item in reviewFormTabs"
            :key="item.key"
            :custom-class="reviewFormTabClass(item.key)"
            :custom-style="formTabButtonCustomStyle"
            @click="activeFormTab = item.key"
          >
            {{ item.label }}
          </u-button>
        </scroll-view>

        <view class="review-form-pane">
          <view v-if="activeFormTab === 'currentTargets'" class="form-section current-targets">
            <view class="section-title current-objective-title">
              <view class="section-title-main">
                <text>本月目标</text>
                <text class="count-pill">
                  合计 {{ currentObjectiveTotal }}
                </text>
              </view>
              <view v-if="canAddCurrentObjective || canToggleCurrentObjectiveDelete" class="section-title-actions">
                <u-button
                  v-if="canAddCurrentObjective"
                  custom-class="dt-mini-btn"
                  :custom-style="miniButtonCustomStyle"
                  @click="addCurrentObjective"
                >
                  增加目标
                </u-button>
                <u-button
                  v-if="canToggleCurrentObjectiveDelete"
                  :custom-class="currentObjectiveDeleteButtonClass()"
                  :custom-style="miniButtonCustomStyle"
                  @click="toggleCurrentObjectiveDelete"
                >
                  {{ currentObjectiveDeleteVisible ? '隐藏删除' : '显示删除' }}
                </u-button>
              </view>
            </view>

            <view v-if="currentObjectives.length" class="objective-list objective-list--separated">
              <view v-for="(item, index) in currentObjectives" :key="`objective-${index}`" class="objective-card">
                <view class="objective-head">
                  <text class="objective-title">
                    目标 {{ index + 1 }}
                  </text>
                  <view class="objective-head-actions">
                    <text class="score-badge">
                      {{ objectiveScore(item) }} 分
                    </text>
                    <u-button
                      v-if="showCurrentObjectiveDelete"
                      custom-class="dt-mini-btn dt-mini-btn--danger"
                      :custom-style="miniButtonCustomStyle"
                      @click="confirmRemoveCurrentObjective(index)"
                    >
                      删除
                    </u-button>
                  </view>
                </view>

                <view class="field-block field-block-wide">
                  <text class="form-field-label">
                    目标描述
                  </text>
                  <!-- H5 PC 需要浏览器原生 textarea 的 resize 行为，uView Pro/uni 包装层只能移动内部手柄，不能带动外层布局。 -->
                  <component
                    :is="nativeTextareaTag"
                    :value="fieldText(item.target)"
                    class="native-review-textarea field-textarea next-objective-textarea"
                    :disabled="!canEditObjectiveDimension"
                    placeholder="绩效目标"
                    @input="setTextareaValue(item, 'target', $event)"
                  />
                </view>

                <view class="objective-fields">
                  <view class="field-block">
                    <text class="form-field-label">
                      权重
                    </text>
                    <u-input
                      :model-value="fieldText(item.weight)"
                      custom-class="field-input"
                      type="digit"
                      :border="true"
                      :clearable="false"
                      :disabled="!canEditObjectiveDimension"
                      :custom-style="fieldInputCustomStyle"
                      placeholder="权重%"
                      @input="setNumberField(item, 'weight', $event)"
                    />
                  </view>
                  <view class="field-block">
                    <text class="form-field-label">
                      完成度
                    </text>
                    <u-input
                      :model-value="fieldText(item.completion)"
                      custom-class="field-input"
                      type="digit"
                      :border="true"
                      :clearable="false"
                      :disabled="!editableSelf"
                      :custom-style="fieldInputCustomStyle"
                      placeholder="完成%"
                      @input="setNumberField(item, 'completion', $event)"
                    />
                  </view>
                </view>

                <view class="field-block field-block-wide">
                  <text class="form-field-label">
                    达成结果
                  </text>
                  <component
                    :is="nativeTextareaTag"
                    :value="fieldText(item.result)"
                    class="native-review-textarea field-textarea"
                    :disabled="!editableSelf"
                    placeholder="达成结果自评"
                    @input="setTextareaValue(item, 'result', $event)"
                  />
                </view>
              </view>
            </view>

            <view v-else class="objective-empty">
              {{ canEditObjectiveDimension ? '暂无目标，点击增加目标开始填写' : '暂无目标' }}
            </view>
          </view>

          <view v-else-if="activeFormTab === 'selfSummary'" class="form-section self-summary">
            <view class="section-title self-summary__title">
              <text>思考总结</text>
            </view>
            <component
              :is="nativeTextareaTag"
              :value="fieldText(localReview.selfSummary)"
              class="native-review-textarea field-textarea self-summary-textarea"
              :rows="selfSummaryTextareaRows"
              :style="selfSummaryTextareaStyle"
              :disabled="!editableSelf"
              placeholder="填写思考总结"
              @input="setTextareaValue(localReview, 'selfSummary', $event)"
            />
          </view>

          <view v-else-if="activeFormTab === 'selfValues'" class="form-section self-values value-review">
            <view class="section-title">
              <view class="section-title-main">
                <text>价值观自评</text>
                <text class="count-pill">
                  总分 {{ valueTotal('self') }}
                </text>
              </view>
            </view>
            <view v-if="values.length" class="value-grid">
              <view v-for="item in values" :key="item.id" class="value-card">
                <view class="value-title-row">
                  <text class="value-name">
                    {{ item.name }}
                  </text>
                  <u-button
                    v-if="hasValueRubric(item)"
                    custom-class="dt-mini-btn value-standard-button"
                    :custom-style="miniButtonCustomStyle"
                    @click="openValueRubric(item)"
                  >
                    查看评分标准
                  </u-button>
                </view>
                <text v-if="item.definition" class="value-desc">
                  {{ item.definition }}
                </text>
                <u-input
                  custom-class="field-input"
                  type="digit"
                  :model-value="fieldText(item.self)"
                  :border="true"
                  :clearable="false"
                  :disabled="!editableSelf"
                  :custom-style="fieldInputCustomStyle"
                  placeholder="0-50"
                  @input="setValueScore(item, 'self', $event)"
                />
              </view>
            </view>
            <view v-else class="objective-empty">
              暂无价值观评分项
            </view>
          </view>

          <view v-else-if="activeFormTab === 'manager'" class="form-section manager-review">
            <view class="section-title">
              <view class="section-title-main">
                <text>上级评价</text>
                <text class="section-meta">
                  上级：{{ userName(review.managerId) }}
                </text>
              </view>
              <text v-if="localReview.managerGrade" class="review-grade-badge">
                上级分档：{{ localReview.managerGrade }}
              </text>
            </view>
            <view class="form-grid">
              <view class="field-block">
                <text class="form-field-label">
                  上级分档
                </text>
                <PerformanceAdaptiveSelect
                  v-model="localReview.managerGrade"
                  custom-class="field-input field-select"
                  title="选择分档"
                  :options="gradeSelectOptions"
                  :mobile-options="plainGradeSelectOptions"
                  :border="true"
                  :disabled="!editableManager"
                  :custom-style="fieldInputCustomStyle"
                  placeholder="未选择"
                />
              </view>
              <view class="field-block field-block-wide">
                <text class="form-field-label">
                  评价内容
                </text>
                <component
                  :is="nativeTextareaTag"
                  :value="fieldText(localReview.managerComment)"
                  class="native-review-textarea field-textarea"
                  :disabled="!editableManager"
                  placeholder="填写上级评价"
                  @input="setTextareaValue(localReview, 'managerComment', $event)"
                />
              </view>
            </view>
            <view class="section-title section-title--sub">
              <view class="section-title-main">
                <text>上级价值观评分</text>
                <text class="count-pill">
                  总分 {{ valueTotal('manager') }}
                </text>
              </view>
            </view>
            <view v-if="values.length" class="value-grid manager-value-review">
              <view v-for="item in values" :key="`manager-${item.id}`" class="value-card">
                <view class="value-title-row">
                  <text class="value-name">
                    {{ item.name }}
                  </text>
                  <u-button
                    v-if="hasValueRubric(item)"
                    custom-class="dt-mini-btn value-standard-button"
                    :custom-style="miniButtonCustomStyle"
                    @click="openValueRubric(item)"
                  >
                    查看评分规则
                  </u-button>
                </view>
                <u-input
                  custom-class="field-input"
                  type="digit"
                  :model-value="fieldText(item.manager)"
                  :border="true"
                  :clearable="false"
                  :disabled="!editableManager"
                  :custom-style="fieldInputCustomStyle"
                  placeholder="0-50"
                  @input="setValueScore(item, 'manager', $event)"
                />
              </view>
            </view>
            <view v-else class="objective-empty">
              暂无价值观评分项
            </view>
          </view>

          <view v-else-if="activeFormTab === 'hrbp'" class="form-section hrbp-review">
            <view class="section-title">
              <view class="section-title-main">
                <text>HRBP评价</text>
                <text class="section-meta">
                  HRBP：{{ userName(localReview.hrbpReviewerId || review.hrbpId) }}
                </text>
              </view>
              <text v-if="localReview.hrbpGrade" class="review-grade-badge">
                HRBP分档：{{ localReview.hrbpGrade }}
              </text>
            </view>
            <view v-if="gradeMismatch" class="notice danger">
              上级分档为 {{ localReview.managerGrade }}，HRBP分档为 {{ localReview.hrbpGrade }}，双方不一致时不能提交。
            </view>
            <view class="form-grid">
              <view class="field-block">
                <text class="form-field-label">
                  HRBP分档
                </text>
                <PerformanceAdaptiveSelect
                  v-model="localReview.hrbpGrade"
                  custom-class="field-input field-select"
                  title="选择分档"
                  :options="gradeSelectOptions"
                  :border="true"
                  :disabled="!editableHrbp"
                  :custom-style="fieldInputCustomStyle"
                  placeholder="未选择"
                />
              </view>
              <view class="field-block field-block-wide">
                <text class="form-field-label">
                  评价内容
                </text>
                <component
                  :is="nativeTextareaTag"
                  :value="fieldText(localReview.hrbpComment)"
                  class="native-review-textarea field-textarea"
                  :disabled="!editableHrbp"
                  placeholder="填写 HRBP 评价"
                  @input="setTextareaValue(localReview, 'hrbpComment', $event)"
                />
              </view>
            </view>
            <view class="section-title section-title--sub">
              <view class="section-title-main">
                <text>HRBP价值观评分</text>
                <text class="count-pill">
                  总分 {{ valueTotal('hrbp') }}
                </text>
              </view>
            </view>
            <view v-if="values.length" class="value-grid hrbp-value-review">
              <view v-for="item in values" :key="`hrbp-${item.id}`" class="value-card">
                <view class="value-title-row">
                  <text class="value-name">
                    {{ item.name }}
                  </text>
                  <u-button
                    v-if="hasValueRubric(item)"
                    custom-class="dt-mini-btn value-standard-button"
                    :custom-style="miniButtonCustomStyle"
                    @click="openValueRubric(item)"
                  >
                    查看评分规则
                  </u-button>
                </view>
                <u-input
                  custom-class="field-input"
                  type="digit"
                  :model-value="fieldText(item.hrbp)"
                  :border="true"
                  :clearable="false"
                  :disabled="!editableHrbp"
                  :custom-style="fieldInputCustomStyle"
                  placeholder="0-50"
                  @input="setValueScore(item, 'hrbp', $event)"
                />
              </view>
            </view>
            <view v-else class="objective-empty">
              暂无价值观评分项
            </view>
          </view>

          <view v-else-if="activeFormTab === 'nextTargets'" class="form-section next-targets">
            <view class="section-title next-objective-title">
              <view class="section-title-main">
                <text>下月目标</text>
                <text class="count-pill">
                  合计 {{ nextObjectiveTotal }}
                </text>
              </view>
              <view v-if="canAddNextObjective || canToggleNextObjectiveDelete" class="section-title-actions">
                <u-button
                  v-if="canAddNextObjective"
                  custom-class="dt-mini-btn"
                  :custom-style="miniButtonCustomStyle"
                  @click="addNextObjective"
                >
                  增加目标
                </u-button>
                <u-button
                  v-if="canToggleNextObjectiveDelete"
                  :custom-class="nextObjectiveDeleteButtonClass()"
                  :custom-style="miniButtonCustomStyle"
                  @click="toggleNextObjectiveDelete"
                >
                  {{ nextObjectiveDeleteVisible ? '隐藏删除' : '显示删除' }}
                </u-button>
              </view>
            </view>

            <view v-if="nextObjectives.length" class="objective-list objective-list--separated">
              <view v-for="(item, index) in nextObjectives" :key="`next-objective-${index}`" class="objective-card">
                <view class="objective-head">
                  <text class="objective-title">
                    下月目标 {{ index + 1 }}
                  </text>
                  <u-button
                    v-if="showNextObjectiveDelete"
                    custom-class="dt-mini-btn dt-mini-btn--danger"
                    :custom-style="miniButtonCustomStyle"
                    @click="confirmRemoveNextObjective(index)"
                  >
                    删除
                  </u-button>
                </view>
                <view class="field-block field-block-wide">
                  <text class="form-field-label">
                    目标描述
                  </text>
                  <component
                    :is="nativeTextareaTag"
                    :value="fieldText(item.target)"
                    class="native-review-textarea field-textarea next-objective-textarea"
                    :disabled="!canEditNextObjectives"
                    placeholder="下月绩效目标"
                    @input="setTextareaValue(item, 'target', $event)"
                  />
                </view>
                <view class="field-block">
                  <text class="form-field-label">
                    权重
                  </text>
                  <u-input
                    :model-value="fieldText(item.weight)"
                    custom-class="field-input"
                    type="digit"
                    :border="true"
                    :clearable="false"
                    :disabled="!canEditNextObjectives"
                    :custom-style="fieldInputCustomStyle"
                    placeholder="权重%"
                    @input="setNumberField(item, 'weight', $event)"
                  />
                </view>
              </view>
            </view>
            <view v-else class="objective-empty next-objective-empty">
              <text class="next-objective-empty-title">
                暂无下月目标
              </text>
              <text class="next-objective-empty-desc">
                {{ canAddNextObjective ? '点击增加目标，填写下月计划和权重' : '当前暂无可查看的下月目标' }}
              </text>
            </view>
          </view>
        </view>
      </template>
    </view>

    <u-popup
      v-model="valueRubricVisible"
      class="value-rubric-popup"
      mode="center"
      custom-class="app-pc-control-scope"
      width="720rpx"
      border-radius="12"
      z-index="10080"
      :safe-area-inset-bottom="true"
      @close="closeValueRubric"
    >
      <view class="value-rubric-modal">
        <view class="value-rubric-modal__head">
          <view class="value-rubric-modal__copy">
            <text class="value-rubric-modal__title">
              评分标准
            </text>
            <text class="value-rubric-modal__desc">
              {{ activeValueRubric?.name || '价值观' }}
            </text>
          </view>
          <u-button custom-class="value-rubric-modal__close app-icon-button" @click="closeValueRubric">
            <u-icon name="close" size="18" color="#4e5969" />
          </u-button>
        </view>
        <scroll-view class="value-rubric-modal__body" scroll-y>
          <view v-if="activeValueRubric?.definition" class="value-rubric-modal__definition">
            {{ activeValueRubric.definition }}
          </view>
          <view v-if="activeValueRubricItems.length" class="value-rubric-list">
            <view
              v-for="(rubric, rubricIndex) in activeValueRubricItems"
              :key="`value-rubric-${activeValueRubric?.id || 'active'}-${rubricIndex}`"
              class="value-rubric-item"
            >
              <text class="value-rubric-score">
                {{ rubric.score || 0 }}分
              </text>
              <view class="value-rubric-copy">
                <text class="value-rubric-label">
                  {{ rubric.label || '未命名' }}
                </text>
                <text v-if="rubric.description" class="value-rubric-desc">
                  {{ rubric.description }}
                </text>
              </view>
            </view>
          </view>
          <view v-else class="objective-empty">
            暂无评分标准
          </view>
        </scroll-view>
        <view class="value-rubric-modal__actions">
          <u-button
            custom-class="dt-mini-btn dt-mini-btn--primary"
            :custom-style="miniButtonCustomStyle"
            @click="closeValueRubric"
          >
            知道了
          </u-button>
        </view>
      </view>
    </u-popup>

    <u-modal
      v-model="actionConfirmVisible"
      custom-class="action-confirm-modal app-pc-control-scope"
      :show-title="false"
      :show-confirm-button="false"
      :show-cancel-button="false"
      :mask-close-able="false"
      width="640rpx"
    >
      <view class="action-confirm-content">
        <text class="action-confirm-title">
          {{ actionConfirmCopy?.title || '确认操作' }}
        </text>
        <text class="action-confirm-desc">
          {{ actionConfirmCopy?.content || '确认执行当前操作？' }}
        </text>
        <view v-if="actionConfirmNeedsFinalizeFields" class="action-confirm-field">
          <view class="field-block">
            <text class="form-field-label">
              最终分档
            </text>
            <PerformanceAdaptiveSelect
              v-model="actionConfirmFinalGrade"
              custom-class="field-input field-select"
              title="选择分档"
              :options="gradeSelectOptions"
              :border="true"
              :custom-style="fieldInputCustomStyle"
              placeholder="未选择"
            />
          </view>
          <view class="field-block field-block-wide">
            <text class="form-field-label">
              HRBP备注
            </text>
            <component
              :is="nativeTextareaTag"
              :value="fieldText(actionConfirmFinalNote)"
              class="native-review-textarea field-textarea action-confirm-textarea"
              placeholder="请输入归档备注（可选）"
              @input="actionConfirmFinalNote = textareaEventText($event)"
            />
          </view>
        </view>
        <view v-else-if="actionConfirmNeedsInput" class="action-confirm-field">
          <view class="action-confirm-label-row">
            <text class="form-field-label">
              {{ actionConfirmInputLabel }}
            </text>
            <text v-if="!actionConfirmInputRequired" class="action-confirm-optional">
              可选
            </text>
          </view>
          <component
            :is="nativeTextareaTag"
            :value="fieldText(actionConfirmRemark)"
            class="native-review-textarea field-textarea action-confirm-textarea"
            :placeholder="actionConfirmInputPlaceholder"
            @input="actionConfirmRemark = textareaEventText($event)"
          />
        </view>
        <view class="action-confirm-actions">
          <u-button
            custom-class="dt-mini-btn action-confirm-cancel"
            :custom-style="miniButtonCustomStyle"
            @click="cancelActionConfirm"
          >
            取消
          </u-button>
          <u-button
            custom-class="dt-mini-btn action-confirm-submit"
            :type="actionConfirmButtonType"
            :loading="detailLoading"
            :custom-style="miniButtonCustomStyle"
            @click="submitActionConfirm"
          >
            {{ actionConfirmCopy?.confirmText || '确认' }}
          </u-button>
        </view>
      </view>
    </u-modal>

    <view v-if="review && footerActionItems.length" class="action-bar actions actions--left actions-outside-card">
      <view
        v-for="item in footerActionItems"
        :key="item.action"
        class="action-button-wrap"
      >
        <u-button
          custom-class="action-button"
          :type="item.type"
          :plain="item.plain"
          :loading="detailLoading"
          :custom-style="actionButtonCustomStyle"
          @click="runAction(item.action)"
        >
          {{ item.label }}
        </u-button>
      </view>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.review-detail-shell {
  width: 100%;
  max-width: 100%;
  min-width: 0;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  gap: 18rpx;
  overflow-x: hidden;
}

.review-detail-shell--with-actions {
  min-height: 0;
}

.detail {
  width: 100%;
  max-width: 100%;
  min-width: 0;
  box-sizing: border-box;
  overflow: hidden;
  border-radius: 8rpx;
  background: $u-bg-white;
}

.review-detail-shell--with-actions .detail {
  flex: none;
  min-height: 0;
  display: block;
}

.review-detail-shell--with-actions .detail__head,
.review-detail-shell--with-actions .process-summary,
.review-detail-shell--with-actions .review-form-tabs {
  flex: 0 0 auto;
}

.detail--floating-card {
  border: 0;
  box-shadow: 0 16rpx 44rpx rgba(31, 35, 41, 0.08);
}

.empty {
  min-height: 360rpx;
  padding: 56rpx 32rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 14rpx;
  text-align: center;
}

.empty__title {
  color: $u-main-color;
  font-size: 30rpx;
  font-weight: 700;
}

.empty__desc {
  color: $u-tips-color;
  font-size: 24rpx;
}

.detail__head {
  padding: 28rpx;
  border-bottom: 1rpx solid $u-border-color;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 20rpx;
}

.detail__title-block {
  min-width: 0;
}

.detail-title-row {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 16rpx;
}

.mobile-process-button {
  display: none;
}

.detail__head-actions {
  flex: 0 0 auto;
  display: flex;
  align-items: flex-start;
  justify-content: flex-end;
  gap: 14rpx;
  flex-wrap: wrap;
}

.detail-status-tag {
  flex: 0 0 auto;
}

.detail-top-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 12rpx;
  flex-wrap: wrap;
}

.detail-top-action-wrap {
  flex: 0 0 auto;
  display: inline-flex;
}

.detail__title,
.detail__desc {
  display: block;
}

.detail__title {
  color: $u-main-color;
  font-size: 32rpx;
  font-weight: 700;
}

.detail__desc {
  margin-top: 6rpx;
  color: $u-tips-color;
  font-size: 24rpx;
  word-break: break-all;
}

.process-summary {
  margin: 24rpx 28rpx 0;
  padding: 22rpx 24rpx;
  border: 1rpx solid #dbeafe;
  border-radius: 8rpx;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24rpx;
  background: #f7fbff;
}

.process-summary-main {
  min-width: 0;
  display: flex;
  flex: 1 1 auto;
  flex-direction: column;
  gap: 12rpx;
}

.process-kicker {
  display: block;
  color: $u-tips-color;
  font-size: 23rpx;
  font-weight: 700;
}

.process-status-line,
.process-current-title-line,
.process-current-meta,
.process-step-title-line {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 14rpx;
  flex-wrap: wrap;
}

.process-handler,
.process-desc,
.process-step-desc,
.process-step-help,
.process-modal-subtitle,
.process-current-handler {
  display: block;
  color: #4e5969;
  font-size: 24rpx;
  line-height: 1.5;
}

.process-desc {
  color: #5f6f86;
}

.status-chip,
.process-state-badge {
  min-height: 40rpx;
  padding: 0 16rpx;
  border-radius: 999rpx;
  display: inline-flex;
  align-items: center;
  background: #f2f3f5;
  color: #86909c;
  font-size: 22rpx;
  font-weight: 700;
  line-height: 40rpx;
}

.status-chip--primary,
.status-chip--info,
.process-state-badge.active {
  background: #eaf3ff;
  color: #1677ff;
}

.status-chip--warning {
  background: #fff7e6;
  color: #d46b08;
}

.status-chip--success,
.process-state-badge.done {
  background: #e8ffea;
  color: #178742;
}

.status-chip--error,
.status-chip--returned,
.process-state-badge.returned {
  background: #fff1f0;
  color: #c42a2a;
}

.process-state-badge.pending {
  background: #f2f3f5;
  color: #86909c;
}

.process-help-btn {
  flex: 0 0 auto;
  max-width: 100%;
}

.process-help-btn :deep(.u-button__text),
:deep(.process-help-btn .u-button__text),
.process-help-btn text,
:deep(.process-help-btn text) {
  min-width: 0;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  justify-content: center !important;
}

.process-modal-mask {
  position: fixed;
  inset: 0;
  z-index: 10090;
  padding: 48rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(15, 23, 42, 0.42);
  backdrop-filter: blur(8rpx);
  overflow: hidden;
  overscroll-behavior: contain;
  touch-action: none;
}

.process-modal {
  width: min(1280rpx, 100%);
  max-height: calc(100vh - 96rpx);
  border-radius: 36rpx;
  overflow: auto;
  background: #fff;
  box-shadow: 0 56rpx 160rpx rgba(15, 23, 42, 0.2);
  overscroll-behavior: contain;
  touch-action: pan-y;
  -webkit-overflow-scrolling: touch;
}

.process-modal-head {
  padding: 40rpx 40rpx 28rpx;
  border-bottom: 1rpx solid $u-border-color;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 32rpx;
}

.process-modal-title-copy {
  min-width: 0;
}

.process-modal-title {
  display: block;
  color: $u-main-color;
  font-size: 36rpx;
  font-weight: 800;
}

.process-modal-subtitle {
  margin-top: 6rpx;
}

.process-modal-close,
:deep(.process-modal-close) {
  width: 68rpx;
  height: 68rpx;
  min-height: 68rpx;
  margin: 0;
  padding: 0;
  border: 0;
  border-radius: 20rpx;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: #f2f3f5;
  color: #4e5969;
  font-size: 40rpx;
  line-height: 68rpx;
}

.process-modal-close::after,
:deep(.process-modal-close)::after {
  border: 0;
}

.process-current-card {
  margin: 32rpx 40rpx 8rpx;
  padding: 24rpx 28rpx;
  border-radius: 28rpx;
  background: #f6f9ff;
}

.process-current-card.compact {
  padding: 28rpx 32rpx;
  border: 1rpx solid #dbeafe;
  display: grid;
  grid-template-columns: 76rpx minmax(0, 1fr);
  align-items: center;
  gap: 24rpx;
  background: linear-gradient(180deg, #f8fbff 0%, #ffffff 100%);
  box-shadow: 0 16rpx 48rpx rgba(22, 119, 255, 0.08);
}

.process-current-index {
  width: 76rpx;
  height: 76rpx;
  min-height: 76rpx;
  border-radius: 24rpx;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: #1677ff;
  color: #fff;
  font-size: 28rpx;
  font-weight: 800;
  line-height: 76rpx;
  box-shadow: 0 20rpx 36rpx rgba(22, 119, 255, 0.18);
}

.process-current-main {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 14rpx;
}

.process-current-label {
  flex: 0 0 auto;
  color: $u-tips-color;
  font-size: 22rpx;
  font-weight: 700;
}

.process-current-title {
  min-width: 0;
  color: $u-main-color;
  font-size: 30rpx;
  font-weight: 800;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.process-current-progress {
  flex: 0 0 auto;
  min-height: 40rpx;
  padding: 0 16rpx;
  border-radius: 999rpx;
  background: #eef5ff;
  color: #1677ff;
  font-size: 22rpx;
  font-weight: 800;
  line-height: 40rpx;
}

.process-timeline {
  padding: 8rpx 40rpx 24rpx;
  display: grid;
}

.process-step-row {
  position: relative;
  display: grid;
  grid-template-columns: 68rpx minmax(0, 1fr);
  gap: 24rpx;
  padding: 24rpx 0;
  border-top: 1rpx solid $u-border-color;

  &:first-child {
    border-top: 0;
  }

  &::before {
    content: "";
    position: absolute;
    left: 32rpx;
    top: 96rpx;
    bottom: -24rpx;
    width: 1rpx;
    background: #dbe3ef;
  }

  &:last-child::before {
    display: none;
  }
}

.process-step-index {
  position: relative;
  z-index: 1;
  width: 68rpx;
  height: 68rpx;
  border-radius: 24rpx;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: #f2f3f5;
  color: #86909c;
  font-size: 24rpx;
  font-weight: 800;
}

.process-step-row.done .process-step-index {
  background: #e8ffea;
  color: #178742;
}

.process-step-row.active .process-step-index {
  background: #1677ff;
  color: #fff;
}

.process-step-main {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 10rpx;
}

.process-step-title {
  color: $u-main-color;
  font-size: 28rpx;
  font-weight: 800;
}

.process-step-role {
  min-height: 40rpx;
  padding: 0 16rpx;
  border-radius: 999rpx;
  display: inline-flex;
  align-items: center;
  background: #f2f3f5;
  color: #86909c;
  font-size: 22rpx;
  line-height: 40rpx;
}

.process-step-row.active .process-step-title {
  color: #1677ff;
}

.process-step-row.done .process-step-title {
  color: #178742;
}

.process-step-row.active .process-step-role {
  background: #eaf3ff;
  color: #1677ff;
}

.process-step-row.done .process-step-role {
  background: #e8ffea;
  color: #178742;
}

.process-step-desc {
  color: $u-main-color;
  font-weight: 600;
}

.process-step-row.pending .process-step-desc,
.process-step-row.pending .process-step-help {
  color: $u-tips-color;
}

.process-step-reason {
  padding: 16rpx 20rpx;
  border: 1rpx solid #ffe1c2;
  border-radius: 16rpx;
  display: flex;
  flex-direction: column;
  gap: 8rpx;
  background: #fff8f0;
}

.process-step-reason-label,
.process-step-reason-text {
  display: block;
}

.process-step-reason-label {
  color: #d46b08;
  font-size: 22rpx;
  font-weight: 700;
}

.process-step-reason-text {
  color: $u-main-color;
  font-size: 24rpx;
  line-height: 1.6;
  word-break: break-word;
}

.process-modal-actions {
  padding: 0 40rpx 40rpx;
  display: flex;
  justify-content: flex-end;
}

.action-confirm-content {
  padding: 34rpx 34rpx 30rpx;
  display: flex;
  flex-direction: column;
  gap: 18rpx;
}

.action-confirm-title {
  color: $u-main-color;
  font-size: 32rpx;
  font-weight: 800;
  line-height: 1.35;
}

.action-confirm-desc {
  color: #4e5969;
  font-size: 24rpx;
  line-height: 1.6;
}

.action-confirm-field {
  display: flex;
  flex-direction: column;
  gap: 10rpx;
}

.action-confirm-label-row {
  display: flex;
  align-items: center;
  gap: 10rpx;
}

.action-confirm-optional {
  min-height: 34rpx;
  padding: 0 12rpx;
  border-radius: 999rpx;
  background: #f2f3f5;
  color: $u-tips-color;
  font-size: 21rpx;
  line-height: 34rpx;
}

.action-confirm-actions {
  padding-top: 4rpx;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 14rpx;
  flex-wrap: wrap;
}

.field-list {
  padding: 0 28rpx;
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  border-top: 1rpx solid $u-border-color;
}

.field {
  min-height: 78rpx;
  padding: 12rpx 18rpx 12rpx 0;
  border-bottom: 1rpx solid $u-border-color;
  display: flex;
  flex-direction: column;
  gap: 4rpx;
}

.field__label {
  color: $u-tips-color;
  font-size: 22rpx;
}

.field__value {
  min-width: 0;
  color: $u-main-color;
  font-size: 24rpx;
  word-break: break-all;
}

.review-form-tabs {
  width: 100%;
  max-width: 100%;
  min-height: 82rpx;
  height: 82rpx;
  padding: 22rpx 28rpx 0;
  box-sizing: border-box;
  display: block;
  overflow-x: auto;
  overflow-y: hidden;
  white-space: nowrap;
}

.review-form-tabs :deep(.uni-scroll-view-content) {
  width: max-content;
  min-width: 100%;
  display: inline-flex;
  align-items: center;
}

.review-form-tab,
:deep(.review-form-tab) {
  flex: 0 0 auto;
  height: 60rpx;
  margin-right: 12rpx;
  padding: 0 24rpx;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1rpx solid $u-border-color;
  border-radius: 8rpx;
  background: $u-bg-white;
  color: $u-content-color;
  font-size: 24rpx;
  line-height: 58rpx;
}

.review-form-tab::after,
:deep(.review-form-tab)::after {
  border: 0;
}

.review-form-tab.active,
:deep(.review-form-tab.active) {
  border-color: #b7d4ff;
  background: #eff6ff;
  color: #1677ff;
  font-weight: 700;
}

:deep(.review-form-tab.active .u-button__text) {
  color: #1677ff;
}

.review-form-pane {
  padding: 24rpx 28rpx 28rpx;
}

.review-detail-shell--with-actions .review-form-pane {
  flex: none;
  min-height: 0;
  overflow-y: visible;
  -webkit-overflow-scrolling: touch;
}

@media screen and (min-width: 769px) and (hover: hover) and (pointer: fine) {
  .review-detail-shell--with-actions .review-form-pane {
    min-height: 560px;
  }
}

.form-section {
  padding: 0;
}

.self-summary__title {
  margin-bottom: 16rpx;
}

.section-title {
  min-height: 54rpx;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
  color: $u-main-color;
  font-size: 28rpx;
  font-weight: 700;
}

.section-title--sub {
  margin-top: 24rpx;
  padding-top: 24rpx;
  border-top: 1rpx solid $u-border-color;
}

.section-title-main {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 12rpx;
  flex-wrap: wrap;
}

.section-title-actions {
  margin-left: auto;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 12rpx;
  flex-wrap: wrap;
}

.section-meta {
  color: $u-tips-color;
  font-size: 22rpx;
  font-weight: 400;
}

.count-pill,
.review-grade-badge,
.score-badge {
  min-height: 40rpx;
  padding: 0 16rpx;
  border-radius: 999rpx;
  display: inline-flex;
  align-items: center;
  background: rgba(var(--u-type-primary-rgb, 41, 121, 255), 0.08);
  color: var(--u-type-primary);
  font-size: 22rpx;
  font-weight: 600;
}

.objective-list {
  display: flex;
  flex-direction: column;
}

.objective-list--separated {
  margin-top: 12rpx;
}

.objective-list--separated .objective-card {
  padding: 22rpx 10rpx 24rpx;
}

.objective-list--separated .objective-card + .objective-card {
  border-top: 1rpx solid $u-border-color;
}

.value-card {
  padding: 20rpx;
  border: 1rpx solid $u-border-color;
  border-radius: 8rpx;
  background: #fff;
}

.objective-head,
.objective-head-actions,
.value-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
}

.objective-title,
.value-name {
  color: $u-main-color;
  font-size: 25rpx;
  font-weight: 700;
}

.objective-fields,
.form-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 18rpx;
}

.field-block {
  margin-top: 18rpx;
  display: flex;
  flex-direction: column;
  gap: 10rpx;
}

.field-block-wide {
  grid-column: 1 / -1;
}

.form-field-label {
  color: $u-content-color;
  font-size: 23rpx;
  font-weight: 600;
}

.field-input,
.field-select,
:deep(.field-input),
:deep(.field-select) {
  width: 100%;
  box-sizing: border-box;
  border: 1rpx solid $u-border-color;
  border-radius: 8rpx;
  background: $u-bg-white;
  color: $u-main-color;
  font-size: 24rpx;
}

.field-input,
.field-select,
:deep(.field-input),
:deep(.field-select) {
  height: 70rpx;
  padding: 0 20rpx;
}

.native-review-textarea {
  --review-textarea-height: 128rpx;
  --review-textarea-line-height: 38.4rpx;
  --review-textarea-max-height: none;
  display: block;
  width: 100%;
  min-height: var(--review-textarea-height);
  height: var(--review-textarea-height);
  max-height: var(--review-textarea-max-height);
  box-sizing: border-box;
  padding: 18rpx 20rpx;
  border: 1rpx solid $u-border-color;
  border-radius: 8rpx;
  appearance: auto;
  background: $u-bg-white;
  color: $u-main-color;
  font-size: 24rpx;
  line-height: var(--review-textarea-line-height);
  overflow: auto !important;
  resize: vertical !important;
  outline: none;
}

.native-review-textarea::-webkit-resizer {
  width: 18rpx;
  height: 18rpx;
  border-radius: 0 0 6rpx;
  background:
    linear-gradient(135deg, transparent 0 58%, $u-tips-color 58% 64%, transparent 64% 72%, $u-tips-color 72% 78%, transparent 78%);
  opacity: 1;
}

@media screen and (max-width: 768px), (hover: none) and (pointer: coarse) {
  .native-review-textarea::-webkit-resizer {
    width: 12rpx;
    height: 12rpx;
    background-size: 12rpx 12rpx;
    background-position: right 3rpx bottom 3rpx;
    background-repeat: no-repeat;
  }
}

// uni-app H5 may render the native textarea component as a uni-textarea host.
// Resize belongs to the host so the whole field grows, while the inner textarea does not render a second handle.
:deep(uni-textarea.native-review-textarea) {
  padding: 0;
}

:deep(uni-textarea.native-review-textarea .uni-textarea-wrapper) {
  min-height: inherit;
  height: 100%;
}

:deep(uni-textarea.native-review-textarea .uni-textarea-textarea) {
  width: 100%;
  min-height: inherit;
  height: 100%;
  box-sizing: border-box;
  padding: 18rpx 20rpx;
  border: 0;
  outline: none;
  color: $u-main-color;
  font-size: 24rpx !important;
  line-height: 1.6 !important;
  overflow: auto;
  resize: none;
}

:deep(.field-input .u-input__input),
:deep(.field-select .u-input__input) {
  color: $u-main-color;
  font-size: 24rpx !important;
}

:deep(.field-input .u-input__input),
:deep(.field-select .u-input__input) {
  line-height: 70rpx !important;
}

.self-summary-textarea {
  --review-textarea-height: 220rpx;
}

.action-confirm-textarea {
  --review-textarea-height: 160rpx;
}

.next-objective-textarea {
  min-height: var(--review-textarea-height);
}

@media (hover: hover) and (pointer: fine) {
  :deep(.action-confirm-modal .u-mode-center-box) {
    width: min(520px, calc(100vw - 96px)) !important;
  }

  :deep(.action-confirm-modal .u-model) {
    width: 100%;
  }

  .native-review-textarea {
    min-height: var(--review-textarea-height);
    height: var(--review-textarea-height);
    font-size: 24rpx !important;
    line-height: 1.6 !important;
    overflow: auto !important;
    resize: vertical !important;
  }

  :deep(uni-textarea.native-review-textarea .uni-textarea-textarea) {
    resize: none;
  }
}

.field-input[disabled],
.field-select[disabled],
.field-textarea[disabled],
:deep(.field-input.u-input--disabled),
:deep(.field-select.u-input--disabled),
:deep(.field-input input:disabled),
:deep(.field-select input:disabled),
:deep(.field-input .u-input__input:disabled),
:deep(.field-select .u-input__input:disabled),
:deep(.field-input .uni-input-input:disabled),
:deep(.field-select .uni-input-input:disabled) {
  background: #f7f9fc;
  color: $u-main-color;
  -webkit-text-fill-color: $u-main-color;
  opacity: 1;
}

:deep(uni-textarea.native-review-textarea[disabled]),
:deep(uni-textarea.native-review-textarea[disabled] .uni-textarea-textarea) {
  background: #f7f9fc;
  color: $u-main-color;
  -webkit-text-fill-color: $u-main-color;
  opacity: 1;
}

.dt-mini-btn,
:deep(.dt-mini-btn) {
  width: auto;
  height: 52rpx;
  margin: 0;
  padding: 0 18rpx;
  border: 1rpx solid rgba(var(--u-type-primary-rgb, 41, 121, 255), 0.24);
  border-radius: 8rpx;
  box-sizing: border-box;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: rgba(var(--u-type-primary-rgb, 41, 121, 255), 0.08);
  color: var(--u-type-primary);
  font-size: 22rpx;
  line-height: 50rpx;
  white-space: nowrap;
}

.dt-mini-btn::after,
:deep(.dt-mini-btn)::after {
  border: 0;
}

.dt-mini-btn--danger,
:deep(.dt-mini-btn--danger) {
  border-color: rgba(227, 77, 89, 0.28);
  background: rgba(227, 77, 89, 0.08);
  color: #e34d59;
}

.dt-mini-btn--primary,
:deep(.dt-mini-btn--primary) {
  border-color: #1677ff;
  background: #1677ff;
  color: #fff;
}

.objective-empty {
  min-height: 180rpx;
  margin-top: 16rpx;
  border: 1rpx dashed $u-border-color;
  border-radius: 8rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  color: $u-tips-color;
  font-size: 24rpx;
  text-align: center;
}

.next-objective-empty {
  flex-direction: column;
  gap: 8rpx;
}

.next-objective-empty-title {
  color: $u-content-color;
  font-weight: 700;
}

.next-objective-empty-desc {
  color: $u-tips-color;
  font-size: 22rpx;
}

.value-grid {
  margin-top: 16rpx;
  display: grid;
  grid-template-columns: 1fr;
  gap: 18rpx;
}

.value-card {
  display: flex;
  flex-direction: column;
  gap: 12rpx;
}

.value-desc {
  color: $u-tips-color;
  font-size: 22rpx;
  line-height: 1.5;
}

.value-rubric-list {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
}

.value-rubric-item {
  padding: 18rpx 0;
  border-bottom: 1rpx solid $u-border-color;
  display: flex;
  gap: 18rpx;
  color: $u-content-color;
  font-size: 24rpx;
  line-height: 1.65;

  &:last-child {
    border-bottom: 0;
  }
}

.value-rubric-score {
  flex: 0 0 auto;
  min-width: 64rpx;
  color: var(--u-type-primary);
  font-weight: 800;
}

.value-rubric-copy {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 6rpx;
  word-break: break-word;
}

.value-rubric-label {
  color: $u-main-color;
  font-weight: 700;
}

.value-rubric-desc {
  color: $u-content-color;
  font-weight: 400;
}

.value-standard-button,
:deep(.value-standard-button) {
  flex: 0 0 auto;
  height: 46rpx;
  min-height: 46rpx;
  padding: 0 16rpx;
  border: 1rpx solid rgba(var(--u-type-primary-rgb, 41, 121, 255), 0.22);
  border-radius: 6rpx;
  background: rgba(var(--u-type-primary-rgb, 41, 121, 255), 0.06);
  color: var(--u-type-primary);
  font-size: 22rpx;
  line-height: 44rpx;
}

.value-standard-button::after,
:deep(.value-standard-button)::after {
  display: none;
}

.value-review .value-grid {
  gap: 0;
}

.value-review .value-card {
  padding: 24rpx 0 28rpx;
  border: 0;
  border-radius: 0;
  background: transparent;
}

.value-review .value-card + .value-card {
  border-top: 1rpx solid $u-border-color;
}

.manager-value-review,
.hrbp-value-review {
  gap: 0;
}

.manager-value-review .value-card,
.hrbp-value-review .value-card {
  padding: 24rpx 0 28rpx;
  border: 0;
  border-radius: 0;
  background: transparent;
}

.manager-value-review .value-card + .value-card,
.hrbp-value-review .value-card + .value-card {
  border-top: 1rpx solid $u-border-color;
}

.value-rubric-popup {
  z-index: 10080;
}

.value-rubric-modal {
  width: 720rpx;
  max-width: calc(100vw - 48rpx);
  max-height: 78vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: #fff;
}

.value-rubric-modal__head {
  flex: 0 0 auto;
  padding: 28rpx 32rpx 22rpx;
  border-bottom: 1rpx solid $u-border-color;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20rpx;
}

.value-rubric-modal__copy {
  min-width: 0;
  display: grid;
  gap: 6rpx;
}

.value-rubric-modal__title {
  color: $u-main-color;
  font-size: 32rpx;
  font-weight: 800;
  line-height: 1.3;
}

.value-rubric-modal__desc {
  color: $u-tips-color;
  font-size: 24rpx;
  line-height: 1.4;
}

.value-rubric-modal__close,
:deep(.value-rubric-modal__close) {
  width: 56rpx;
  height: 56rpx;
  min-height: 56rpx;
  margin: 0;
  padding: 0;
  border: 0;
  border-radius: 50%;
  background: #f2f3f5;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.value-rubric-modal__close::after,
:deep(.value-rubric-modal__close)::after {
  display: none;
}

.value-rubric-modal__body {
  flex: 1 1 auto;
  min-height: 0;
  max-height: 52vh;
  padding: 24rpx 32rpx;
  box-sizing: border-box;
}

.value-rubric-modal__definition {
  margin-bottom: 16rpx;
  padding: 16rpx 18rpx;
  border-radius: 8rpx;
  background: #f7f9fc;
  color: $u-tips-color;
  font-size: 24rpx;
  line-height: 1.55;
}

.value-rubric-modal__actions {
  flex: 0 0 auto;
  padding: 18rpx 32rpx 26rpx;
  border-top: 1rpx solid $u-border-color;
  display: flex;
  justify-content: flex-end;
  background: #fff;
}

@media screen and (min-width: 769px) and (hover: hover) and (pointer: fine) {
  .value-rubric-popup :deep(.u-mode-center-box) {
    width: min(680px, calc(100vw - 96px)) !important;
  }

  .value-rubric-modal {
    width: 100%;
    max-width: none;
  }
}

.notice {
  margin-top: 12rpx;
  padding: 16rpx 18rpx;
  border-radius: 8rpx;
  font-size: 23rpx;
  line-height: 1.5;
}

.notice.danger {
  border: 1rpx solid rgba(227, 77, 89, 0.22);
  background: rgba(227, 77, 89, 0.08);
  color: #b4232f;
}

.actions {
  width: 100%;
  padding: 0;
  display: flex;
  align-items: center;
  align-content: flex-start;
  flex-wrap: wrap;
  gap: 16rpx;
  justify-content: flex-start;
}

.action-button-wrap {
  flex: 0 0 auto;
  width: auto;
  display: inline-flex;
  align-items: center;
}

.action-button-wrap :deep(.action-button.u-btn),
.action-button-wrap :deep(.u-btn),
.action-button-wrap :deep(button) {
  flex: 0 0 auto;
  width: auto !important;
  min-width: 152rpx;
  margin: 0 !important;
}

.actions--left {
  justify-content: flex-start !important;
}

.actions-outside-card {
  align-self: stretch;
  margin-top: 4rpx;
}

.review-detail-shell--with-actions .actions-outside-card {
  position: static;
  bottom: auto;
  z-index: auto;
  margin-top: 0;
  min-height: 96rpx;
  padding: 18rpx 0 28rpx;
  border-top: 1rpx solid rgba(229, 230, 235, 0.72);
  box-sizing: border-box;
  background: #fff;
  box-shadow: none;
  overflow: visible;
}

@media screen and (max-width: 768px), (hover: none) and (pointer: coarse) {
  .review-detail-shell {
    gap: 16rpx;
  }

  .review-detail-shell--with-actions {
    height: auto;
    min-height: 0;
  }

  .review-detail-shell--with-actions .detail {
    flex: none;
    min-height: 0;
    display: block;
    overflow: visible;
  }

  .review-detail-shell--with-actions .review-form-pane {
    flex: none;
    min-height: 0;
    overflow-y: visible;
  }

  .detail {
    width: 100%;
    max-width: 100%;
    min-width: 0;
    border-radius: 12rpx;
  }

  .detail--floating-card {
    box-shadow: 0 12rpx 36rpx rgba(31, 35, 41, 0.08);
  }

  .detail__head {
    padding: 24rpx;
    gap: 14rpx;
    align-items: stretch;
    flex-direction: column;
  }

  .detail-title-row {
    align-items: center;
    flex-direction: row;
    justify-content: space-between;
    gap: 12rpx;
  }

  .detail__title {
    flex: 1 1 auto;
    min-width: 0;
  }

  .mobile-process-button {
    width: auto;
    flex: 0 0 auto;
    display: flex;
    justify-content: flex-end;
  }

  .detail__head-actions {
    width: 100%;
    justify-content: space-between;
  }

  .detail-status-tag {
    display: none;
  }

  .detail-top-actions {
    justify-content: flex-start;
  }

  .detail__title {
    font-size: 30rpx;
    line-height: 1.35;
  }

  .detail__desc {
    font-size: 23rpx;
    line-height: 1.5;
  }

  .process-summary {
    display: none;
  }

  .review-form-tabs {
    min-height: 78rpx;
    height: 78rpx;
    padding: 18rpx 24rpx 0;
    overflow-x: auto;
    overflow-y: hidden;
    white-space: nowrap;
  }

  .review-form-tabs :deep(.uni-scroll-view-content) {
    width: max-content;
    min-width: 100%;
    display: inline-flex;
    align-items: center;
    flex-wrap: nowrap;
    gap: 8rpx;
  }

  .review-form-tab,
  :deep(.review-form-tab) {
    flex: 0 0 auto;
    min-width: 148rpx;
    height: 56rpx;
    margin-right: 0;
    padding: 0 22rpx;
    font-size: 23rpx;
    line-height: 54rpx;
  }

  :deep(.review-form-tab .u-button__text) {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .review-form-pane {
    padding: 20rpx 24rpx 24rpx;
  }

  .section-title {
    min-height: auto;
    align-items: flex-start;
    flex-direction: column;
    gap: 12rpx;
  }

  .section-title-main {
    width: 100%;
  }

  .section-title-actions {
    width: 100%;
    margin-left: 0;
    justify-content: flex-start;
    gap: 10rpx;
  }

  .current-objective-title,
  .next-objective-title {
    align-items: center;
    flex-direction: row;
    justify-content: space-between;
    gap: 12rpx;
  }

  .current-objective-title .section-title-main,
  .next-objective-title .section-title-main {
    width: auto;
    min-width: 0;
    flex: 1 1 auto;
  }

  .current-objective-title .section-title-actions,
  .next-objective-title .section-title-actions {
    width: auto;
    margin-left: auto;
    flex: 0 0 auto;
    justify-content: flex-end;
  }

  .objective-list--separated {
    margin-top: 10rpx;
  }

  .objective-list--separated .objective-card {
    padding: 22rpx 0 24rpx;
  }

  .objective-head {
    align-items: center;
    flex-wrap: nowrap;
  }

  .objective-head-actions {
    margin-left: auto;
    gap: 10rpx;
    flex: 0 0 auto;
    flex-wrap: nowrap;
    justify-content: flex-end;
  }

  .field-block {
    margin-top: 16rpx;
    gap: 8rpx;
  }

  .field-input,
  .field-select,
  .field-textarea,
  :deep(.field-input),
  :deep(.field-select),
  :deep(.field-textarea) {
    min-width: 0;
  }

  .process-modal-mask {
    padding: calc(76rpx + env(safe-area-inset-top)) 28rpx 28rpx;
    align-items: flex-start;
  }

  .process-modal {
    width: 100%;
    max-height: calc(100dvh - 160rpx);
    border-radius: 28rpx;
  }

  .process-modal-head,
  .process-current-card,
  .process-timeline,
  .process-modal-actions {
    margin-left: 0;
    margin-right: 0;
    padding-left: 28rpx;
    padding-right: 28rpx;
  }

  .field-list,
  .objective-fields,
  .form-grid,
  .value-grid {
    grid-template-columns: 1fr;
  }

  .actions-outside-card {
    padding: 0 24rpx;
    box-sizing: border-box;
  }

  .review-detail-shell--with-actions .actions-outside-card {
    position: static;
    bottom: auto;
    z-index: auto;
    min-height: 0;
    padding: 16rpx 24rpx 6rpx;
  }
}
</style>
