import { computed, h, ref, Teleport } from 'vue'
import { usePerformanceContext } from './context'
import { renderTablePagination, useTablePagination } from './tablePagination'

const HISTORY_TABLE_COLUMNS = [
  { key: 'period', label: '考评月份', mobile: true },
  { key: 'status', label: '状态', mobile: true },
  { key: 'objectiveScore', label: '目标得分' },
  { key: 'managerGrade', label: '上级分档' },
  { key: 'hrbpGrade', label: 'HRBP分档' },
  { key: 'finalGrade', label: '最终分档', mobile: true }
]

const HRBP_REVIEW_TABLE_COLUMNS = [
  { key: 'employee', label: '员工', mobile: true },
  { key: 'department', label: '部门' },
  { key: 'position', label: '岗位' },
  { key: 'period', label: '考评月份', mobile: true },
  { key: 'nextPeriod', label: '目标月份' },
  { key: 'status', label: '状态', mobile: true },
  { key: 'objectiveScore', label: '目标得分' },
  { key: 'managerGrade', label: '上级分档' },
  { key: 'hrbpGrade', label: 'HRBP分档' },
  { key: 'finalGrade', label: '最终分档' },
  { key: 'actions', label: '操作' }
]

const MANAGER_REVIEW_TABLE_COLUMNS = [
  { key: 'employee', label: '员工', mobile: true },
  { key: 'position', label: '岗位' },
  { key: 'period', label: '考评月份', mobile: true },
  { key: 'status', label: '状态', mobile: true },
  { key: 'objectiveScore', label: '目标得分' },
  { key: 'managerGrade', label: '上级分档' },
  { key: 'hrbpGrade', label: 'HRBP分档' },
  { key: 'finalGrade', label: '最终分档' },
  { key: 'handler', label: '当前处理人' },
  { key: 'actions', label: '操作' }
]

function readInputValue(event) {
  return event.detail?.value ?? event.target.value
}

function createObjective() {
  return {
    target: '',
    weight: 0,
    completion: '',
    result: ''
  }
}

function createNextObjective() {
  return {
    target: '',
    weight: 0
  }
}

function addCurrentObjective(review) {
  if (!Array.isArray(review.objectives)) review.objectives = []
  review.objectives.push(createObjective())
}

function addNextObjective(review) {
  if (!Array.isArray(review.nextObjectives)) review.nextObjectives = []
  review.nextObjectives.push(createNextObjective())
}

function removeCurrentObjective(review, index) {
  if (!Array.isArray(review.objectives)) return
  review.objectives.splice(index, 1)
}

function removeNextObjective(review, index) {
  if (!Array.isArray(review.nextObjectives)) return
  review.nextObjectives.splice(index, 1)
}

function confirmObjectiveDelete(index, label = '目标') {
  const content = `确认删除${label} ${index + 1}？删除后需要保存才会生效。`
  return new Promise((resolve) => {
    if (typeof uni !== 'undefined' && uni.showModal) {
      uni.showModal({
        title: `删除${label}`,
        content,
        confirmText: '删除',
        confirmColor: '#e34d59',
        success: (res) => resolve(Boolean(res.confirm)),
        fail: () => resolve(false)
      })
      return
    }
    if (typeof window !== 'undefined' && window.confirm) {
      resolve(window.confirm(content))
      return
    }
    resolve(false)
  })
}

async function confirmRemoveCurrentObjective(review, index, onRemoved) {
  const confirmed = await confirmObjectiveDelete(index)
  if (!confirmed) return
  removeCurrentObjective(review, index)
  if (onRemoved) onRemoved()
}

async function confirmRemoveNextObjective(review, index, onRemoved) {
  const confirmed = await confirmObjectiveDelete(index, '下月目标')
  if (!confirmed) return
  removeNextObjective(review, index)
  if (onRemoved) onRemoved()
}

function currentAssignee(ctx, review) {
  if (review.status === 'draft') return ctx.userName(review.employeeId)
  if (review.status === 'manager_review') return ctx.userName(review.managerId)
  if (review.status === 'hrbp_review') return ctx.userName(review.hrbpReviewerId || review.hrbpId)
  if (review.status === 'employee_confirm') return ctx.userName(review.employeeId)
  if (review.status === 'hr_final') return ctx.userName(review.hrbpReviewerId || review.hrbpId)
  return '已归档'
}

function reviewDetailTitle(ctx, review) {
  const title = `${review.period || '-'} 月度考评`
  if (ctx.contentView.value === 'mine') return title
  return `${ctx.userName(review.employeeId)} · ${title}`
}

function myPerformanceActionText(review) {
  if (review.status === 'manager_review') return '等待上级评价'
  if (review.status === 'hrbp_review') return '等待HRBP评价'
  if (review.status === 'employee_confirm') return '等待确认结果'
  if (review.status === 'hr_final') return '等待HRBP归档'
  return '等待填写自评'
}

function renderMyPerformanceSwitcher(ctx) {
  const rows = ctx.currentReviews.value
  if (ctx.contentView.value !== 'mine' || rows.length <= 1) return null
  return h('section', { class: 'panel my-performance-switcher' }, [
    h('view', { class: 'my-performance-switcher-head' }, [
      h('text', { class: 'my-performance-switcher-title' }, '进行中绩效'),
      h('text', { class: 'count-pill' }, `${rows.length} 条`)
    ]),
    h('view', { class: 'my-performance-switcher-list' }, rows.map((review) =>
      h('button', {
        class: ['my-performance-switcher-item', ctx.selectedReview.value?.id === review.id ? 'active' : ''],
        onClick: () => ctx.selectReview(review.id)
      }, [
        h('view', { class: 'my-performance-switcher-main' }, [
          h('text', { class: 'my-performance-switcher-name' }, `${review.period || '-'} 月度考评`),
          h('text', { class: 'my-performance-switcher-desc' }, myPerformanceActionText(review))
        ]),
        h('text', { class: ctx.statusClass(review.status) }, ctx.statusText(review.status))
      ])
    ))
  ])
}

function historyColumnClass(column) {
  return ['history-col', `history-col-${column.key}`, column.mobile ? 'history-mobile-visible' : 'history-mobile-hidden']
}

function renderHistoryCell(ctx, column, review) {
  switch (column.key) {
    case 'period':
      return review.period || '-'
    case 'status':
      return h('text', { class: ctx.statusClass(review.status) }, ctx.statusText(review.status))
    case 'objectiveScore':
      return String(ctx.totalObjectiveScore(review))
    case 'managerGrade':
      return review.managerGrade || '-'
    case 'hrbpGrade':
      return review.hrbpGrade || '-'
    case 'finalGrade':
      return ctx.effectiveGrade(review) || '-'
    default:
      return '-'
  }
}

function hrbpReviewColumnClass(column) {
  return [
    'hrbp-review-col',
    `hrbp-review-col-${column.key}`,
    column.mobile ? 'hrbp-review-mobile-visible' : 'hrbp-review-mobile-hidden',
    column.key === 'actions' ? 'table-actions-cell' : ''
  ]
}

function renderHrbpReviewCell(ctx, column, review, onOpenRow) {
  switch (column.key) {
    case 'employee':
      return ctx.userName(review.employeeId)
    case 'department':
      return review.department || '-'
    case 'position':
      return employeePosition(ctx, review)
    case 'period':
      return review.period || '-'
    case 'nextPeriod':
      return review.nextPeriod || '-'
    case 'status':
      return h('text', { class: ctx.statusClass(review.status) }, ctx.statusText(review.status))
    case 'objectiveScore':
      return String(ctx.totalObjectiveScore(review))
    case 'managerGrade':
      return review.managerGrade || '-'
    case 'hrbpGrade':
      return review.hrbpGrade || '-'
    case 'finalGrade':
      return ctx.effectiveGrade(review) || '-'
    case 'actions':
      return h('button', {
        class: ['dt-btn', ctx.canHrbpHandle(review) ? 'dt-btn-primary' : 'dt-btn-light', 'small', 'table-action-btn'],
        onClick: (event) => {
          event?.stopPropagation?.()
          onOpenRow(review)
        }
      }, ctx.canHrbpHandle(review) ? '评价' : '查看')
    default:
      return '-'
  }
}

function managerReviewColumnClass(column) {
  return ['manager-review-col', `manager-review-col-${column.key}`, column.mobile ? 'manager-review-mobile-visible' : 'manager-review-mobile-hidden']
}

function employeePosition(ctx, review) {
  const user = ctx.state.users.find((item) => item.id === review.employeeId)
  return user?.position || '-'
}

function renderManagerReviewCell(ctx, column, review, onOpenRow) {
  switch (column.key) {
    case 'employee':
      return ctx.userName(review.employeeId)
    case 'position':
      return employeePosition(ctx, review)
    case 'period':
      return review.period || '-'
    case 'status':
      return h('text', { class: ctx.statusClass(review.status) }, ctx.statusText(review.status))
    case 'objectiveScore':
      return String(ctx.totalObjectiveScore(review))
    case 'managerGrade':
      return review.managerGrade || '-'
    case 'hrbpGrade':
      return review.hrbpGrade || '-'
    case 'finalGrade':
      return ctx.effectiveGrade(review) || '-'
    case 'handler':
      return currentAssignee(ctx, review)
    case 'actions':
      return h('button', {
        class: ['dt-btn', ctx.canManager(review) ? 'dt-btn-primary' : 'dt-btn-light', 'small', 'table-action-btn'],
        onClick: (event) => {
          event?.stopPropagation?.()
          onOpenRow(review)
        }
      }, ctx.canManager(review) ? '评价' : '查看')
    default:
      return '-'
  }
}

function hasManagerEvaluation(review) {
  if (!review) return false
  if (String(review.managerGrade || '').trim()) return true
  if (String(review.managerComment || '').trim()) return true
  return (review.values || []).some((item) => String(item.manager ?? '').trim())
}

function hasHrbpEvaluation(review) {
  if (!review) return false
  if (String(review.hrbpGrade || '').trim()) return true
  if (String(review.hrbpComment || '').trim()) return true
  return (review.values || []).some((item) => String(item.hrbp ?? '').trim())
}

function managerReviewTitleMeta(ctx, review) {
  const name = ctx.userName(review?.managerId)
  return name && name !== '无' ? `上级：${name}` : '上级未设置'
}

function hrbpReviewTitleMeta(ctx, review) {
  const name = ctx.userName(review?.hrbpReviewerId || review?.hrbpId)
  return name && name !== '无' ? `HRBP：${name}` : 'HRBP未设置'
}

function reviewGradeBadge(label, grade) {
  const value = String(grade || '').trim()
  if (!value) return null
  return h('text', { class: 'review-grade-badge' }, `${label}：${value}`)
}

function valueRubricItems(tpl, item) {
  const source = Array.isArray(tpl?.rubric) && tpl.rubric.length > 0 ? tpl.rubric : item?.rubric
  return (Array.isArray(source) ? source : []).filter((rubric) =>
    String(rubric?.label || '').trim() ||
    String(rubric?.score ?? '').trim() ||
    String(rubric?.description || '').trim()
  )
}

function valueRubricScoreText(score) {
  const value = Number(score)
  if (!Number.isFinite(value)) return '-'
  return Number.isInteger(value) ? String(value) : String(value.toFixed(1)).replace(/\.0$/, '')
}

function renderValueRubricList(rubrics) {
  if (!rubrics.length) return null
  return h('view', { class: 'value-score-guide-list' }, rubrics.map((rubric) => {
    const label = String(rubric?.label || '').trim() || '未命名'
    const description = String(rubric?.description || '').trim()
    return h('view', { class: 'value-score-guide-item' }, [
      h('text', { class: 'value-score-guide-score' }, `${valueRubricScoreText(rubric?.score)}分`),
      h('view', { class: 'value-score-guide-copy' }, [
        h('text', { class: 'value-score-guide-name' }, label),
        description ? h('text', { class: 'value-score-guide-desc' }, description) : null
      ])
    ])
  }))
}

const reviewFormTabs = [
  ['currentTargets', '本月目标'],
  ['selfSummary', '思考总结'],
  ['selfValues', '价值观自评'],
  ['manager', '上级评价'],
  ['hrbp', 'HRBP评价'],
  ['nextTargets', '下月目标']
].map(([key, label]) => ({ key, label }))

function normalizeReviewFormTab(ctx, review) {
  if (ctx.reviewTab.value === 'current') ctx.reviewTab.value = 'currentTargets'
  if (ctx.reviewTab.value === 'next') ctx.reviewTab.value = 'nextTargets'
  if (!reviewFormTabs.some((item) => item.key === ctx.reviewTab.value)) {
    ctx.reviewTab.value = reviewFormTabs[0].key
  }
  return ctx.reviewTab.value
}

function renderReviewFormPane(ctx, review, editableSelf, editableManager, editableHrbp, readonly = false) {
  switch (ctx.reviewTab.value) {
    case 'selfSummary':
      return h(SelfSummarySection, { review, editableSelf })
    case 'selfValues':
      return h(ValueSection, { review, field: 'self', title: '价值观自评', editable: editableSelf })
    case 'manager':
      return h(ManagerSection, { review, editable: editableManager })
    case 'hrbp':
      return h(HrbpSection, { review, editable: editableHrbp })
    case 'nextTargets':
      return h(NextObjectiveSection, { review, editable: editableSelf })
    case 'currentTargets':
    default:
      return h(CurrentObjectiveSection, { review, editableSelf, readonly })
  }
}

const flowSteps = [
  {
    status: 'draft',
    title: '员工填写',
    role: '员工',
    desc: '填写当月目标完成度、达成结果自评、思考总结，并确认下月目标。',
    actor: 'employee',
    historyKeywords: ['创建考评单', '保存员工自评', '提交员工自评', '退回员工修改', '撤销员工自评提交']
  },
  {
    status: 'manager_review',
    title: '上级评价',
    role: '直属上级',
    desc: '审核员工自评内容，填写上级评价、价值观评分和建议分档。',
    actor: 'manager',
    historyKeywords: ['提交上级评价', '退回上级修改', '撤销上级评价提交']
  },
  {
    status: 'hrbp_review',
    title: 'HRBP评价',
    role: 'HRBP',
    desc: '复核绩效材料，填写 HRBP 评价和分档；如有问题可退回上级修改。',
    actor: 'hrbp',
    historyKeywords: ['提交 HRBP 评价', '员工提出异议', '撤销 HRBP 评价提交', '退回 HRBP 修改']
  },
  {
    status: 'employee_confirm',
    title: '员工确认',
    role: '员工',
    desc: '查看评价结果并确认；如存在异议，可填写说明后提交反馈。',
    actor: 'employee',
    historyKeywords: ['员工确认结果', '员工提出异议']
  },
  {
    status: 'hr_final',
    title: 'HRBP归档',
    role: 'HRBP',
    desc: '处理员工确认或异议，确认最终分档和归档备注。',
    actor: 'hrbp',
    historyKeywords: ['HRBP 归档', '退回 HRBP 修改', '员工确认结果']
  },
  {
    status: 'completed',
    title: '完成',
    role: '系统归档',
    desc: '考评单归档完成，结果进入汇总统计。',
    actor: 'system',
    historyKeywords: ['HRBP 归档']
  }
]

function stepActorName(ctx, review, step) {
  if (step.actor === 'employee') return ctx.userName(review.employeeId)
  if (step.actor === 'manager') return ctx.userName(review.managerId)
  if (step.actor === 'hrbp') return ctx.userName(review.hrbpReviewerId || review.hrbpId)
  return '系统'
}

function formatHistoryTime(at) {
  if (!at) return ''
  const date = new Date(at)
  if (Number.isNaN(date.getTime())) return ''
  const pad = (value) => String(value).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`
}

function latestHistoryForStep(review, step) {
  const histories = Array.isArray(review.history) ? review.history : []
  return histories.slice().reverse().find((item) =>
    step.historyKeywords.some((keyword) => String(item.action || '').includes(keyword))
  )
}

function historyMatchesStep(history, step) {
  return step.historyKeywords.some((keyword) => String(history?.action || '').includes(keyword))
}

function historyActionParts(action) {
  const text = String(action || '').trim()
  const fullSeparatorIndex = text.indexOf('：')
  const halfSeparatorIndex = text.indexOf(':')
  const separatorIndex = fullSeparatorIndex >= 0
    ? fullSeparatorIndex
    : halfSeparatorIndex
  if (separatorIndex < 0) return { title: text, reason: '' }
  return {
    title: text.slice(0, separatorIndex).trim(),
    reason: text.slice(separatorIndex + 1).trim()
  }
}

function historyActionNeedsReason(title) {
  const text = String(title || '').trim()
  return text.startsWith('撤销') || text.startsWith('撤回') || text.startsWith('退回') || text.includes('异议')
}

function missingHistoryReasonText(title) {
  const text = String(title || '').trim()
  if (text.startsWith('撤销')) return '未记录撤销理由'
  if (text.startsWith('撤回')) return '未记录撤回理由'
  if (text.startsWith('退回')) return '未记录退回原因'
  if (text.includes('异议')) return '未记录异议原因'
  return ''
}

function historyReasonLabel(title) {
  const text = String(title || '').trim()
  if (text.startsWith('撤销')) return '撤销理由'
  if (text.startsWith('撤回')) return '撤回理由'
  if (text.startsWith('退回')) return '退回原因'
  if (text.includes('异议')) return '异议原因'
  return '理由'
}

function historyReasonForStep(review, step, selectedHistory, actionParts) {
  if (actionParts.reason) return actionParts.reason
  const histories = Array.isArray(review.history) ? review.history : []
  const selectedTitle = actionParts.title || selectedHistory?.action || ''
  const sameActionWithReason = histories.slice().reverse().find((item) => {
    if (!historyMatchesStep(item, step)) return false
    const parts = historyActionParts(item.action)
    return parts.reason && (!selectedTitle || parts.title === selectedTitle)
  })
  if (sameActionWithReason) return historyActionParts(sameActionWithReason.action).reason
  return historyActionNeedsReason(selectedTitle) ? missingHistoryReasonText(selectedTitle) : ''
}

function historyDetailMeta(history, time) {
  return [history?.by || 'system', time].filter(Boolean).join(' · ')
}

function flowProgressRows(ctx, review) {
  const currentStep = ctx.statusMeta[review.status]?.step || 0
  return flowSteps.map((step, index) => {
    const history = latestHistoryForStep(review, step)
    const isCompletedFinal = review.status === 'completed' && index === currentStep
    const state = index < currentStep || isCompletedFinal ? 'done' : index === currentStep ? 'active' : 'pending'
    const stateLabel = state === 'done' ? '已完成' : state === 'active' ? '进行中' : '待处理'
    const time = formatHistoryTime(history?.at)
    const actionParts = historyActionParts(history?.action)
    const reason = historyReasonForStep(review, step, history, actionParts)
    const reasonLabel = historyReasonLabel(actionParts.title || history?.action)
    const detailMeta = historyDetailMeta(history, time)
    const detail = history
      ? [actionParts.title || history.action, detailMeta].filter(Boolean).join(' · ')
      : state === 'active'
        ? `等待 ${stepActorName(ctx, review, step)} 处理`
        : state === 'done'
          ? '已进入下一流程节点'
          : step.actor === 'system'
            ? '待归档完成'
            : `待 ${step.role} 处理`
    return {
      ...step,
      index,
      indexText: String(index + 1).padStart(2, '0'),
      progressText: `${String(index + 1).padStart(2, '0')}/${String(flowSteps.length).padStart(2, '0')}`,
      actorName: stepActorName(ctx, review, step),
      state,
      stateLabel,
      detail,
      reason,
      reasonLabel,
      history
    }
  })
}

function emptyReviewCopy(ctx) {
  const view = ctx.contentView.value
  if (view === 'manager') {
    return {
      title: '暂无待处理绩效单',
      desc: '员工提交自评后，会在这里进入上级评价队列。',
      steps: ['等待员工提交', '上级评价', '继续流转']
    }
  }
  if (view === 'dashboard') {
    return {
      title: '暂无待办绩效单',
      desc: '有新的填写、评价或确认任务时，会自动汇总到这里。',
      steps: ['员工填写', '上级评价', 'HRBP归档']
    }
  }
  if (view === 'mine') {
    return {
      title: '暂无进行中绩效单',
      desc: '属于你的绩效单只要流程未完成，就会在这里展示。',
      steps: ['等待创建', '流程流转', '完成归档']
    }
  }
  return {
    title: '本月还没有绩效单',
    desc: '绩效单创建后，可以在这里填写目标完成度、思考总结和自评内容。',
    steps: ['等待创建', '填写自评', '提交流转']
  }
}

function renderPerformanceEmptyState(ctx) {
  const copy = emptyReviewCopy(ctx)
  return h('view', { class: 'performance-empty' }, [
    h('view', { class: 'performance-empty-visual' }, [
      h('view', { class: 'performance-empty-icon' }, '绩'),
      h('view', { class: 'performance-empty-line wide' }),
      h('view', { class: 'performance-empty-line' }),
      h('view', { class: 'performance-empty-dots' }, [
        h('text'),
        h('text'),
        h('text')
      ])
    ]),
    h('text', { class: 'performance-empty-title' }, copy.title),
    h('text', { class: 'performance-empty-desc' }, copy.desc),
    h('view', { class: 'performance-empty-steps' }, copy.steps.map((item, index) =>
      h('text', { class: 'performance-empty-step' }, `${index + 1}. ${item}`)
    )),
    h('button', { class: 'dt-btn dt-btn-light performance-empty-action', onClick: ctx.refreshData }, '刷新看看')
  ])
}

function renderPerformanceLoadingState() {
  return h('section', { class: 'panel performance-loading-panel' }, [
    h('view', { class: 'performance-loading-head' }, [
      h('view', { class: 'performance-loading-title' }),
      h('view', { class: 'performance-loading-action' })
    ]),
    h('view', { class: 'performance-loading-table' }, Array.from({ length: 6 }).map((_, rowIndex) =>
      h('view', { class: 'performance-loading-row' }, Array.from({ length: 6 }).map((__, cellIndex) =>
        h('view', {
          class: [
            'performance-loading-cell',
            cellIndex === 0 ? 'wide' : '',
            rowIndex === 0 ? 'head' : ''
          ]
        })
      ))
    ))
  ])
}

function workbenchPageDesc(ctx) {
  if (ctx.state.view === 'dashboard') return '处理需要你跟进的绩效事项'
  if (ctx.contentLoading.value) return '加载中...'
  if (ctx.contentView.value === 'mine') return `共 ${ctx.currentReviews.value.length} 条进行中`
  return `共 ${Number(ctx.state.reviewListTotal || ctx.currentReviews.value.length)} 条记录`
}

function workbenchTodoMeta(ctx, review) {
  const period = review.period || '-'
  const employee = ctx.userName(review.employeeId)
  if (ctx.canSelf(review)) {
    return {
      title: `${period} 月度考评待填写`,
      desc: '填写本月目标完成情况、思考总结、价值观自评和下月目标。',
      action: '去填写'
    }
  }
  if (ctx.canManager(review)) {
    return {
      title: `${employee} 的 ${period} 月度考评待上级评价`,
      desc: '审核员工自评，填写上级评价、价值观评分和建议分档。',
      action: '去评价'
    }
  }
  if (ctx.canHrbpHandle(review)) {
    return {
      title: `${employee} 的 ${period} 月度考评待 HRBP 评价`,
      desc: '复核绩效材料，填写 HRBP 评价和分档。',
      action: '去评价'
    }
  }
  if (ctx.canEmployeeConfirm(review)) {
    return {
      title: `${period} 月度考评待确认`,
      desc: '查看评价结果，确认无误或提交异议说明。',
      action: '去确认'
    }
  }
  if (ctx.canFinal(review)) {
    return {
      title: `${employee} 的 ${period} 月度考评待归档`,
      desc: '处理员工确认或异议，确认最终分档并完成归档。',
      action: '去归档'
    }
  }
  return {
    title: `${period} 月度考评待处理`,
    desc: '查看当前绩效流程并处理待办。',
    action: '查看'
  }
}

function renderWorkbenchTodoList(ctx) {
  const todos = ctx.queueReviews()
  return h('section', { class: 'panel workbench-todo-panel' }, [
    h('view', { class: 'workbench-todo-head' }, [
      h('view', {}, [
        h('text', { class: 'panel-title' }, '我的待办'),
        h('text', { class: 'workbench-todo-desc' }, '点击待办可直接进入对应处理页面')
      ]),
      h('text', { class: 'count-pill' }, `${todos.length} 条`)
    ]),
    todos.length
      ? h('view', { class: 'workbench-todo-list' }, todos.map((review) => {
          const meta = workbenchTodoMeta(ctx, review)
          return h('button', { class: 'workbench-todo-item', onClick: () => ctx.openWorkbenchTodo(review) }, [
            h('view', { class: 'workbench-todo-main' }, [
              h('view', { class: 'workbench-todo-title-row' }, [
                h('text', { class: 'workbench-todo-title' }, meta.title),
                h('text', { class: ctx.statusClass(review.status) }, ctx.statusText(review.status))
              ]),
              h('text', { class: 'workbench-todo-meta' }, [
                review.department ? `${review.department} · ` : '',
                `当前处理人 ${currentAssignee(ctx, review)}`
              ].join('')),
              h('text', { class: 'workbench-todo-desc-line' }, meta.desc)
            ]),
            h('text', { class: 'workbench-todo-action' }, meta.action)
          ])
        }))
      : h('view', { class: 'workbench-todo-empty' }, [
          h('text', { class: 'workbench-todo-empty-title' }, '当前没有待办'),
          h('text', { class: 'workbench-todo-empty-desc' }, '需要你填写、评价、确认或归档的绩效单会显示在这里。')
        ])
  ])
}

const WorkbenchView = {
  name: 'WorkbenchView',
  setup() {
    const ctx = usePerformanceContext()
    const historyDetailReview = ref(null)
    const hrbpDetailReview = ref(null)
    const managerDetailReview = ref(null)
    const historyTableRows = computed(() => ctx.currentReviews.value
      .slice()
      .sort((a, b) => String(b.period || '').localeCompare(String(a.period || '')))
    )
    const historyPagination = useTablePagination(historyTableRows)
    const hrbpTableRows = computed(() => {
      const pendingRows = ctx.currentReviews.value.filter((review) => review.status === 'hrbp_review')
      const reviewedRows = ctx.currentReviews.value.filter((review) => ['employee_confirm', 'hr_final', 'completed'].includes(review.status))
      return ctx.hrbpReviewTab.value === 'pending' ? pendingRows : reviewedRows
    })
    const hrbpPagination = useTablePagination(hrbpTableRows)
    const managerTableRows = computed(() => {
      const pendingRows = ctx.currentReviews.value.filter((review) => review.status === 'manager_review')
      const reviewedRows = ctx.currentReviews.value.filter((review) => ['hrbp_review', 'employee_confirm', 'hr_final', 'completed'].includes(review.status))
      return ctx.managerReviewTab.value === 'pending' ? pendingRows : reviewedRows
    })
    const managerPagination = useTablePagination(managerTableRows)

    function openHistoryDetail(review) {
      ctx.reviewTab.value = 'currentTargets'
      historyDetailReview.value = review
    }

    function openHrbpReviewRow(review) {
      if (ctx.hrbpReviewTab.value === 'pending') {
        hrbpDetailReview.value = null
        ctx.selectReview(review.id)
        ctx.reviewTab.value = 'hrbp'
        return
      }
      ctx.selectedReviewId.value = ''
      ctx.reviewTab.value = 'currentTargets'
      hrbpDetailReview.value = review
    }

    function openManagerReviewRow(review) {
      if (ctx.managerReviewTab.value === 'pending') {
        managerDetailReview.value = null
        ctx.selectReview(review.id)
        ctx.reviewTab.value = 'manager'
        return
      }
      ctx.selectedReviewId.value = ''
      ctx.reviewTab.value = 'currentTargets'
      managerDetailReview.value = review
    }

    return () => {
      const head = h('view', { class: 'page-head' }, [
        h('view', {}, [
          h('text', { class: 'page-title' }, ctx.sectionTitle.value),
          h('text', { class: 'page-desc' }, workbenchPageDesc(ctx))
        ]),
        h('view', { class: 'head-actions' }, [
          ctx.contentView.value === 'mine' && ctx.canCreateReview()
            ? h('button', { class: 'dt-btn dt-btn-primary', onClick: ctx.openCreateReviewDialog }, '创建')
            : null,
          h('button', { class: 'dt-btn dt-btn-light', onClick: ctx.refreshData }, '刷新')
        ])
      ])

      if (ctx.contentLoading.value) {
        return h('view', { class: 'workbench workbench-loading' }, [
          head,
          renderPerformanceLoadingState()
        ])
      }

      if (ctx.state.view === 'dashboard') {
        return h('view', { class: 'workbench' }, [
          head,
          renderWorkbenchTodoList(ctx)
        ])
      }

      if (ctx.contentView.value === 'hrbp') {
        const rows = hrbpPagination.rows.value
        const selectedHrbpReview = ctx.selectedReviewId.value ? ctx.selectedReview.value : null
        if (selectedHrbpReview) {
          return h('view', { class: 'workbench hrbp-review-detail-page' }, [
            head,
            h('view', { class: 'detail-page-toolbar' }, [
              h('button', { class: 'dt-btn dt-btn-light', onClick: () => { ctx.selectedReviewId.value = '' } }, '返回列表')
            ]),
            h('section', { class: 'panel detail-panel hrbp-review-detail-panel' }, [
              h(ReviewForm, { review: selectedHrbpReview })
            ])
          ])
        }
        return h('view', { class: 'workbench hrbp-review-page' }, [
          head,
          h('view', { class: 'hrbp-review-tabs-bar' }, [
            h('view', { class: 'hrbp-review-tabs' }, [
              h('button', { class: ['hrbp-review-tab', ctx.hrbpReviewTab.value === 'pending' ? 'active' : ''], onClick: () => ctx.switchHrbpReviewTab('pending') }, '待评'),
              h('button', { class: ['hrbp-review-tab', ctx.hrbpReviewTab.value === 'reviewed' ? 'active' : ''], onClick: () => ctx.switchHrbpReviewTab('reviewed') }, '已评')
            ])
          ]),
          h('view', { class: 'table-panel-stack hrbp-table-panel-stack' }, [
            h('section', { class: 'panel table-panel' }, [
              h('view', { class: 'hrbp-table-wrap table-wrap' }, [
                h('table', { class: 'hrbp-review-table summary-table' }, [
                  h('thead', {}, h('tr', {}, HRBP_REVIEW_TABLE_COLUMNS.map((column) => h('th', { class: hrbpReviewColumnClass(column) }, column.label)))),
                  h('tbody', {}, rows.length
                    ? rows.map((review) => h('tr', { class: 'hrbp-review-clickable-row', onClick: () => openHrbpReviewRow(review) }, HRBP_REVIEW_TABLE_COLUMNS.map((column) =>
                        h('td', { class: hrbpReviewColumnClass(column) }, renderHrbpReviewCell(ctx, column, review, openHrbpReviewRow))
                      )))
                    : h('tr', {}, [
                        h('td', { class: 'hrbp-empty-row', colspan: HRBP_REVIEW_TABLE_COLUMNS.length }, ctx.hrbpReviewTab.value === 'pending' ? '当前没有待评记录' : '当前没有已评记录')
                      ]))
                ])
              ])
            ]),
            renderTablePagination(hrbpPagination)
          ]),
          hrbpDetailReview.value
            ? h(Teleport, { to: 'body' }, [
                h(HistoryReviewDetailModal, {
                  review: hrbpDetailReview.value,
                  onClose: () => { hrbpDetailReview.value = null }
                })
              ])
            : null
        ])
      }

      if (ctx.contentView.value === 'manager') {
        const rows = managerPagination.rows.value
        const selectedManagerReview = ctx.selectedReviewId.value ? ctx.selectedReview.value : null
        if (selectedManagerReview) {
          return h('view', { class: 'workbench manager-review-detail-page' }, [
            head,
            h('view', { class: 'detail-page-toolbar' }, [
              h('button', { class: 'dt-btn dt-btn-light', onClick: () => { ctx.selectedReviewId.value = '' } }, '返回列表')
            ]),
            h('section', { class: 'panel detail-panel manager-review-detail-panel' }, [
              h(ReviewForm, { review: selectedManagerReview })
            ])
          ])
        }
        return h('view', { class: 'workbench hrbp-review-page manager-review-page' }, [
          head,
          h('view', { class: 'hrbp-review-tabs-bar' }, [
            h('view', { class: 'hrbp-review-tabs' }, [
              h('button', { class: ['hrbp-review-tab', ctx.managerReviewTab.value === 'pending' ? 'active' : ''], onClick: () => ctx.switchManagerReviewTab('pending') }, '待评'),
              h('button', { class: ['hrbp-review-tab', ctx.managerReviewTab.value === 'reviewed' ? 'active' : ''], onClick: () => ctx.switchManagerReviewTab('reviewed') }, '已评')
            ])
          ]),
          h('view', { class: 'table-panel-stack manager-table-panel-stack' }, [
            h('section', { class: 'panel table-panel' }, [
              h('view', { class: 'hrbp-table-wrap table-wrap' }, [
                h('table', { class: 'hrbp-review-table summary-table' }, [
                  h('thead', {}, h('tr', {}, MANAGER_REVIEW_TABLE_COLUMNS.map((column) => h('th', { class: managerReviewColumnClass(column) }, column.label)))),
                  h('tbody', {}, rows.length
                    ? rows.map((review) => h('tr', { class: 'manager-review-clickable-row', onClick: () => openManagerReviewRow(review) }, MANAGER_REVIEW_TABLE_COLUMNS.map((column) =>
                        h('td', { class: [...managerReviewColumnClass(column), column.key === 'actions' ? 'table-actions-cell' : ''] }, renderManagerReviewCell(ctx, column, review, openManagerReviewRow))
                      )))
                    : h('tr', {}, [
                        h('td', { class: 'hrbp-empty-row', colspan: MANAGER_REVIEW_TABLE_COLUMNS.length }, ctx.managerReviewTab.value === 'pending' ? '当前没有待评记录' : '当前没有已评记录')
                      ]))
                ])
              ])
            ]),
            renderTablePagination(managerPagination)
          ]),
          managerDetailReview.value
            ? h(Teleport, { to: 'body' }, [
                h(HistoryReviewDetailModal, {
                  review: managerDetailReview.value,
                  onClose: () => { managerDetailReview.value = null }
                })
              ])
            : null
        ])
      }

      if (ctx.contentView.value === 'history') {
        const historyRows = historyPagination.rows.value
        return h('view', { class: 'workbench history-performance-page' }, [
          head,
          h('view', { class: 'table-panel-stack history-table-panel-stack' }, [
            h('section', { class: 'panel table-panel' }, [
              h('view', { class: 'history-table-wrap table-wrap' }, [
                h('table', { class: 'history-performance-table summary-table' }, [
                  h('thead', {}, h('tr', {}, HISTORY_TABLE_COLUMNS.map((column) => h('th', { class: historyColumnClass(column) }, column.label)))),
                  h('tbody', {}, historyRows.length
                    ? historyRows.map((review) => h('tr', { class: 'history-clickable-row', onClick: () => openHistoryDetail(review) }, HISTORY_TABLE_COLUMNS.map((column) =>
                        h('td', { class: historyColumnClass(column) }, renderHistoryCell(ctx, column, review))
                      )))
                    : h('tr', {}, [
                        h('td', { class: 'history-empty-row', colspan: HISTORY_TABLE_COLUMNS.length }, '当前没有历史记录')
                      ]))
                ])
              ])
            ]),
            renderTablePagination(historyPagination)
          ]),
          historyDetailReview.value
            ? h(Teleport, { to: 'body' }, [
                h(HistoryReviewDetailModal, {
                  review: historyDetailReview.value,
                  onClose: () => { historyDetailReview.value = null }
                })
              ])
            : null
        ])
      }

      return h('view', { class: 'workbench' }, [
        head,
        h('view', { class: 'review-detail-only' }, [
          renderMyPerformanceSwitcher(ctx),
          h('section', { class: 'panel detail-panel' }, ctx.selectedReview.value
            ? [h(ReviewForm, { review: ctx.selectedReview.value })]
            : [renderPerformanceEmptyState(ctx)])
        ])
      ])
    }
  }
}

const ReviewForm = {
  props: ['review', 'readonly'],
  setup(props) {
    const ctx = usePerformanceContext()
    const processVisible = ref(false)

    return () => {
      const review = props.review
      const readonly = props.readonly === true
      const editableSelf = !readonly && ctx.canSelf(review)
      const editableManager = !readonly && ctx.canManager(review)
      const editableHrbp = !readonly && ctx.canHrbpHandle(review)
      const editableConfirm = !readonly && ctx.canEmployeeConfirm(review)
      const editableFinal = !readonly && ctx.canFinal(review)
      const activeFormTab = normalizeReviewFormTab(ctx, review)
      const currentStep = ctx.statusMeta[review.status]?.step || 0
      const progressRows = flowProgressRows(ctx, review)
      const currentProgress = progressRows[currentStep] || progressRows[0]
      return h('view', { class: 'review-form' }, [
        h('view', { class: 'detail-head' }, [
          h('view', {}, [
            h('view', { class: 'detail-title-row' }, [
              h('text', { class: 'detail-title' }, reviewDetailTitle(ctx, review)),
              h('button', { class: 'dt-btn dt-btn-light process-title-btn', onClick: () => { processVisible.value = true } }, '查看流程进度')
            ]),
            ctx.contentView.value === 'mine'
              ? null
              : h('text', { class: 'detail-subtitle' }, `${review.nextPeriod} 目标 · 当前处理人 ${currentAssignee(ctx, review)}`)
          ]),
          ctx.contentView.value === 'mine'
            ? null
            : h('text', { class: ctx.statusClass(review.status) }, ctx.statusText(review.status))
        ]),
        h('view', { class: 'process-summary' }, [
          h('view', { class: 'process-summary-main' }, [
            h('text', { class: 'process-kicker' }, '当前流程状态'),
            h('view', { class: 'process-status-line' }, [
              h('text', { class: ['process-state-badge', currentProgress.state] }, currentProgress.stateLabel),
              h('text', { class: ctx.statusClass(review.status) }, ctx.statusText(review.status)),
              h('text', { class: 'process-handler' }, `当前处理人 ${currentAssignee(ctx, review)}`)
            ]),
            h('text', { class: 'process-desc' }, `${currentProgress.progressText} · ${currentProgress.detail}`)
          ]),
          h('button', { class: 'dt-btn dt-btn-light process-help-btn', onClick: () => { processVisible.value = true } }, '查看流程进度')
        ]),
        processVisible.value ? h(ProcessModal, { review, onClose: () => { processVisible.value = false } }) : null,
        h('view', { class: 'review-tabs' }, reviewFormTabs.map((item) =>
          h('button', { class: ['review-tab', activeFormTab === item.key ? 'active' : ''], onClick: () => { ctx.reviewTab.value = item.key } }, item.label)
        )),
        h('view', { class: 'review-form-pane' }, [
          renderReviewFormPane(ctx, review, editableSelf, editableManager, editableHrbp, readonly)
        ]),
        readonly ? null : h('view', { class: 'action-bar' }, [
          editableSelf ? h('button', { class: 'dt-btn dt-btn-light', onClick: () => ctx.performReviewAction('save-self', '已保存') }, '保存') : null,
          editableSelf ? h('button', { class: 'dt-btn dt-btn-primary', onClick: () => ctx.performReviewAction('submit-self', '已提交给上级') }, '提交自评') : null,
          editableManager ? h('button', { class: 'dt-btn dt-btn-danger-light', onClick: () => ctx.returnReview('return-employee', '退回员工修改') }, '退回员工') : null,
          editableManager ? h('button', { class: 'dt-btn dt-btn-primary', onClick: () => ctx.performReviewAction('submit-manager', '已提交给 HRBP') }, '提交给HRBP') : null,
          editableHrbp ? h('button', { class: 'dt-btn dt-btn-danger-light', onClick: () => ctx.returnReview('return-manager', '退回上级修改') }, '退回上级') : null,
          editableHrbp ? h('button', { class: 'dt-btn dt-btn-primary', onClick: () => ctx.performReviewAction('submit-hrbp', '已提交给员工确认') }, '提交给员工') : null,
          editableConfirm ? h('button', { class: 'dt-btn dt-btn-primary', onClick: () => ctx.performReviewAction('confirm-result', '已确认') }, '确认结果') : null,
          editableConfirm ? h('button', { class: 'dt-btn dt-btn-danger-light', onClick: () => ctx.performReviewAction('dispute-result', '已提出异议') }, '提出异议') : null,
          editableFinal && review.status === 'hr_final' ? h('button', { class: 'dt-btn dt-btn-danger-light', onClick: () => ctx.returnReview('return-hrbp', '退回 HRBP 修改') }, '退回HRBP') : null,
          editableFinal ? h('button', { class: 'dt-btn dt-btn-primary', onClick: () => ctx.performReviewAction('finalize', '已归档') }, '归档') : null,
          ctx.canWithdraw(review) ? h('button', { class: 'dt-btn dt-btn-light', onClick: ctx.withdrawReview }, '撤回提交') : null
        ])
      ])
    }
  }
}

const HistoryReviewDetailModal = {
  props: ['review'],
  emits: ['close'],
  setup(props, { emit }) {
    const ctx = usePerformanceContext()

    return () => h('view', { class: 'history-detail-modal-mask', onClick: () => emit('close') }, [
      h('view', { class: 'history-detail-modal', onClick: (event) => event.stopPropagation() }, [
        h('view', { class: 'history-detail-modal-head' }, [
          h('view', {}, [
            h('text', { class: 'history-detail-modal-title' }, '绩效详情'),
            h('text', { class: 'history-detail-modal-subtitle' }, `${ctx.userName(props.review.employeeId)} · ${props.review.period || '-'} 月度考评`)
          ]),
          h('button', { class: 'process-modal-close', onClick: () => emit('close') }, '×')
        ]),
        h('view', { class: 'history-detail-modal-body' }, [
          h(ReviewForm, { review: props.review, readonly: true })
        ])
      ])
    ])
  }
}

const ProcessModal = {
  props: ['review'],
  emits: ['close'],
  setup(props, { emit }) {
    const ctx = usePerformanceContext()

    return () => {
      const review = props.review
      const currentStep = ctx.statusMeta[review.status]?.step || 0
      const progressRows = flowProgressRows(ctx, review)
      const currentProgress = progressRows[currentStep] || progressRows[0]
      return h('view', { class: 'process-modal-mask', onClick: () => emit('close') }, [
        h('view', { class: 'process-modal', onClick: (event) => event.stopPropagation() }, [
          h('view', { class: 'process-modal-head' }, [
            h('view', {}, [
              h('text', { class: 'process-modal-title' }, '流程进度'),
              h('text', { class: 'process-modal-subtitle' }, `当前：${ctx.statusText(review.status)} · 处理人 ${currentAssignee(ctx, review)}`)
            ]),
            h('button', { class: 'process-modal-close', onClick: () => emit('close') }, '×')
          ]),
          h('view', { class: 'process-current-card compact' }, [
            h('text', { class: 'process-current-index' }, currentProgress.indexText),
            h('view', { class: 'process-current-main' }, [
              h('view', { class: 'process-current-title-line' }, [
                h('text', { class: 'process-current-label' }, '当前节点'),
                h('text', { class: 'process-current-title' }, currentProgress.title),
                h('text', { class: 'process-current-progress' }, currentProgress.progressText)
              ]),
              h('view', { class: 'process-current-meta' }, [
                h('text', { class: ctx.statusClass(review.status) }, ctx.statusText(review.status)),
                h('text', { class: ['process-state-badge', currentProgress.state] }, currentProgress.stateLabel),
                h('text', { class: 'process-current-handler' }, `处理人 ${currentAssignee(ctx, review)}`)
              ])
            ])
          ]),
          h('view', { class: 'process-timeline' }, progressRows.map((step) =>
            h('view', { class: ['process-step-row', step.state] }, [
              h('text', { class: 'process-step-index' }, step.indexText),
              h('view', { class: 'process-step-main' }, [
                h('view', { class: 'process-step-title-line' }, [
                  h('text', { class: 'process-step-title' }, step.title),
                  h('text', { class: 'process-step-role' }, step.role),
                  h('text', { class: ['process-state-badge', step.state] }, step.stateLabel)
                ]),
                h('text', { class: 'process-step-desc' }, step.detail),
                step.reason ? h('view', { class: 'process-step-reason' }, [
                  h('text', { class: 'process-step-reason-label' }, step.reasonLabel || '理由'),
                  h('text', { class: 'process-step-reason-text' }, step.reason)
                ]) : null,
                h('text', { class: 'process-step-help' }, step.desc)
              ])
            ])
          )),
          h('view', { class: 'process-modal-actions' }, [
            h('button', { class: 'dt-btn dt-btn-primary', onClick: () => emit('close') }, '知道了')
          ])
        ])
      ])
    }
  }
}

const ValueStandardModal = {
  props: ['standard'],
  emits: ['close'],
  setup(props, { emit }) {
    return () => {
      const standard = props.standard || {}
      const rubrics = Array.isArray(standard.rubrics) ? standard.rubrics : []
      return h(Teleport, { to: 'body' }, [
        h('view', { class: 'value-standard-modal-mask', onClick: () => emit('close') }, [
          h('view', { class: 'value-standard-modal', onClick: (event) => event.stopPropagation() }, [
            h('view', { class: 'value-standard-modal-head' }, [
              h('view', { class: 'value-standard-title-copy' }, [
                h('text', { class: 'value-standard-modal-title' }, '评分标准'),
                h('text', { class: 'value-standard-modal-subtitle' }, standard.name || '-')
              ]),
              h('button', { class: 'process-modal-close', onClick: () => emit('close') }, '×')
            ]),
            standard.definition ? h('text', { class: 'value-standard-definition' }, standard.definition) : null,
            h('view', { class: 'value-score-guide value-standard-guide' }, [
              renderValueRubricList(rubrics)
            ]),
            h('view', { class: 'value-standard-actions' }, [
              h('button', { class: 'dt-btn dt-btn-primary', onClick: () => emit('close') }, '知道了')
            ])
          ])
        ])
      ])
    }
  }
}

const CurrentObjectiveSection = {
  props: ['review', 'editableSelf', 'readonly'],
  setup(props) {
    const ctx = usePerformanceContext()
    const showDeleteActions = ref(false)

    function toggleDeleteActions() {
      showDeleteActions.value = !showDeleteActions.value
    }

    return () => {
      const editableObjectives = !props.readonly && ctx.canEditObjectiveDimension(props.review)
      const objectives = Array.isArray(props.review.objectives) ? props.review.objectives : []

      return h('view', {}, [
        h('section', { class: 'form-section' }, [
          h('view', { class: 'section-title' }, [
            h('view', { class: 'section-title-main' }, [
              h('text', {}, '本月目标'),
              h('text', { class: 'count-pill' }, `合计 ${ctx.totalObjectiveScore(props.review)}`)
            ]),
            editableObjectives ? h('view', { class: 'section-actions' }, [
              h('button', { class: 'dt-btn dt-btn-light small', onClick: () => addCurrentObjective(props.review) }, '增加目标'),
              h('button', {
                class: ['dt-btn small objective-delete-toggle', showDeleteActions.value ? 'active' : 'dt-btn-light'],
                onClick: toggleDeleteActions
              }, showDeleteActions.value ? '隐藏删除' : '显示删除')
            ]) : null
          ]),
          objectives.length
            ? h('view', { class: 'objective-list' }, objectives.map((item, index) => h('view', { class: 'objective-card', key: `objective-${index}` }, [
              h('view', { class: 'objective-head' }, [
                h('text', { class: 'objective-title' }, `目标 ${index + 1}`),
                h('view', { class: 'objective-head-actions' }, [
                  h('text', { class: 'score-badge' }, `${ctx.objectiveScore(item)} 分`),
                  showDeleteActions.value && editableObjectives ? h('button', {
                    class: 'dt-btn dt-btn-danger-light small objective-delete-btn',
                    onClick: () => confirmRemoveCurrentObjective(props.review, index, () => {
                      if (!props.review.objectives?.length) showDeleteActions.value = false
                    })
                  }, '删除') : null
                ])
              ]),
              h('view', { class: 'field-block field-block-wide' }, [
                h('text', { class: 'field-label' }, '目标描述'),
                h('textarea', { class: 'field-textarea', value: item.target, disabled: !editableObjectives, placeholder: '绩效目标', onInput: (event) => { item.target = readInputValue(event) } })
              ]),
              h('view', { class: 'objective-fields' }, [
                h('view', { class: 'field-block' }, [
                  h('text', { class: 'field-label' }, '权重'),
                  h('input', { class: 'field-input', type: 'number', value: item.weight, disabled: !editableObjectives, placeholder: '权重%', onInput: (event) => { item.weight = Number(readInputValue(event)) } })
                ]),
                h('view', { class: 'field-block' }, [
                  h('text', { class: 'field-label' }, '完成度'),
                  h('input', { class: 'field-input', type: 'number', value: item.completion, disabled: !props.editableSelf, placeholder: '完成%', onInput: (event) => { item.completion = readInputValue(event) } })
                ])
              ]),
              h('view', { class: 'field-block field-block-wide' }, [
                h('text', { class: 'field-label' }, '达成结果'),
                h('textarea', { class: 'field-textarea', value: item.result, disabled: !props.editableSelf, placeholder: '达成结果自评', onInput: (event) => { item.result = readInputValue(event) } })
              ])
            ])))
            : h('view', { class: 'objective-empty' }, editableObjectives ? '暂无目标，点击增加目标开始填写' : '暂无目标')
        ])
      ])
    }
  }
}

const SelfSummarySection = {
  props: ['review', 'editableSelf'],
  setup(props) {
    return () => h('section', { class: 'form-section' }, [
      h('view', { class: 'section-title' }, [h('text', {}, '思考总结')]),
      h('textarea', { class: 'field-textarea', value: props.review.selfSummary, disabled: !props.editableSelf, placeholder: '填写思考总结', onInput: (event) => { props.review.selfSummary = readInputValue(event) } })
    ])
  }
}

const NextObjectiveSection = {
  props: ['review', 'editable'],
  setup(props) {
    const ctx = usePerformanceContext()
    const showDeleteActions = ref(false)

    function toggleDeleteActions() {
      showDeleteActions.value = !showDeleteActions.value
    }

    return () => {
      const nextObjectives = Array.isArray(props.review.nextObjectives) ? props.review.nextObjectives : []
      const editableNextObjectives = props.editable && ctx.canEditNextObjectives(props.review)
      const canAddNextObjective = props.editable && ctx.canAddNextObjective(props.review)
      const canDeleteNextObjective = props.editable && ctx.canDeleteNextObjective(props.review)

      return h('section', { class: 'form-section' }, [
        h('view', { class: 'section-title' }, [
          h('view', { class: 'section-title-main' }, [
            h('text', {}, '下月目标'),
            h('text', { class: 'count-pill' }, `合计 ${nextObjectives.reduce((total, item) => total + (Number(item.weight) || 0), 0)}`)
          ]),
          canAddNextObjective || canDeleteNextObjective ? h('view', { class: 'section-actions' }, [
            canAddNextObjective ? h('button', { class: 'dt-btn dt-btn-light small', onClick: () => addNextObjective(props.review) }, '增加目标') : null,
            canDeleteNextObjective ? h('button', {
              class: ['dt-btn small objective-delete-toggle', showDeleteActions.value ? 'active' : 'dt-btn-light'],
              onClick: toggleDeleteActions
            }, showDeleteActions.value ? '隐藏删除' : '显示删除') : null
          ]) : null
        ]),
        nextObjectives.length
          ? h('view', { class: 'objective-list' }, nextObjectives.map((item, index) => h('view', { class: 'objective-card next', key: `next-objective-${index}` }, [
            h('view', { class: 'objective-head' }, [
              h('text', { class: 'objective-title' }, `下月目标 ${index + 1}`),
              showDeleteActions.value && canDeleteNextObjective ? h('button', {
                class: 'dt-btn dt-btn-danger-light small objective-delete-btn',
                onClick: () => confirmRemoveNextObjective(props.review, index, () => {
                  if (!props.review.nextObjectives?.length) showDeleteActions.value = false
                })
              }, '删除') : null
            ]),
            h('view', { class: 'field-block field-block-wide' }, [
              h('text', { class: 'field-label' }, '目标描述'),
              h('textarea', { class: 'field-textarea', value: item.target, disabled: !editableNextObjectives, placeholder: '下月绩效目标', onInput: (event) => { item.target = readInputValue(event) } })
            ]),
            h('view', { class: 'objective-fields objective-fields-next' }, [
              h('view', { class: 'field-block' }, [
                h('text', { class: 'field-label' }, '权重'),
                h('input', { class: 'field-input', type: 'number', value: item.weight, disabled: !editableNextObjectives, placeholder: '权重%', onInput: (event) => { item.weight = Number(readInputValue(event)) } })
              ])
            ])
          ])))
          : h('view', { class: 'objective-empty next-objective-empty' }, [
            h('view', { class: 'next-objective-empty-mark' }, '+'),
            h('text', { class: 'next-objective-empty-title' }, '暂无下月目标'),
            h('text', { class: 'next-objective-empty-desc' }, canAddNextObjective ? '点击增加目标，填写下月计划和权重' : '当前暂无可查看的下月目标'),
            canAddNextObjective ? h('button', { class: 'dt-btn dt-btn-primary small next-objective-empty-action', onClick: () => addNextObjective(props.review) }, '增加目标') : null
          ])
      ])
    }
  }
}

const ValueSection = {
  props: ['review', 'field', 'title', 'editable'],
  setup(props) {
    const ctx = usePerformanceContext()
    const activeStandard = ref(null)

    function openStandard(event, payload) {
      event?.stopPropagation?.()
      activeStandard.value = payload
    }

    return () => h('section', { class: 'form-section' }, [
      h('view', { class: 'section-title' }, [
        h('text', {}, props.title),
        h('text', { class: 'count-pill' }, `总分 ${ctx.valueTotal(props.review, props.field)}`)
      ]),
      h('view', { class: 'value-grid value-list' }, props.review.values.map((item) => {
        const tpl = ctx.state.template?.values?.find((value) => value.id === item.id) || {}
        const valueName = tpl.name || item.name || item.id
        const valueDefinition = tpl.definition || item.definition || ''
        const rubrics = valueRubricItems(tpl, item)
        return h('view', { class: 'value-card' }, [
          h('view', { class: 'value-title-row' }, [
            h('text', { class: 'value-name' }, valueName),
            rubrics.length ? h('button', {
              class: 'value-score-tag',
              onClick: (event) => openStandard(event, {
                name: valueName,
                definition: valueDefinition,
                rubrics
              })
            }, '评分标准') : null
          ]),
          valueDefinition ? h('text', { class: 'value-desc' }, valueDefinition) : null,
          h('input', { class: 'field-input', type: 'number', value: item[props.field], disabled: !props.editable, placeholder: '0-50', onInput: (event) => { item[props.field] = readInputValue(event) } })
        ])
      })),
      activeStandard.value ? h(ValueStandardModal, { standard: activeStandard.value, onClose: () => { activeStandard.value = null } }) : null
    ])
  }
}

const ManagerSection = {
  props: ['review', 'editable'],
  setup(props) {
    const ctx = usePerformanceContext()

    return () => {
      const managerPendingText = '上级待评'
      const managerPending = !hasManagerEvaluation(props.review)

      return h('section', { class: 'form-section review-evaluation-section manager-review-section' }, [
        h('view', { class: 'section-title' }, [
          h('view', { class: 'section-title-main' }, [
            h('text', {}, '上级评价'),
            h('text', { class: 'section-meta' }, managerReviewTitleMeta(ctx, props.review))
          ]),
          reviewGradeBadge('上级分档', props.review.managerGrade)
        ]),
        h('view', { class: props.editable ? 'form-grid manager-review-grid manager-review-block' : 'manager-review-readonly manager-review-block' }, [
          props.editable ? h('view', { class: 'field-block manager-grade-field' }, [
            h('text', { class: 'field-label' }, '上级分档'),
            h('select', { class: 'field-select', value: props.review.managerGrade, onChange: (event) => { props.review.managerGrade = event.target.value } }, ctx.gradeOptions().map((grade) => h('option', { value: grade }, grade || '未选择')))
          ]) : null,
          h('view', { class: 'field-block field-block-wide manager-comment-field' }, [
            props.editable ? h('text', { class: 'field-label' }, '评价内容') : null,
            h('textarea', { class: 'field-textarea', value: props.review.managerComment, disabled: !props.editable, placeholder: managerPending && !props.editable ? managerPendingText : '填写上级评价', onInput: (event) => { props.review.managerComment = readInputValue(event) } })
          ])
        ]),
        h(ValueSection, { review: props.review, field: 'manager', title: '上级价值观评分', editable: props.editable })
      ])
    }
  }
}

const HrbpSection = {
  props: ['review', 'editable'],
  setup(props) {
    const ctx = usePerformanceContext()

    return () => {
      const hrbpPendingText = 'HRBP待评'
      const hrbpPending = !hasHrbpEvaluation(props.review)
      const gradeMismatch = props.review.managerGrade && props.review.hrbpGrade && props.review.managerGrade !== props.review.hrbpGrade
      const gradeMismatchNotice = () => gradeMismatch
        ? h('view', { class: 'notice danger hrbp-grade-notice' }, `上级分档为 ${props.review.managerGrade}，HRBP分档为 ${props.review.hrbpGrade}，双方不一致时不能提交。`)
        : null

      return h('section', { class: 'form-section review-evaluation-section hrbp-review-section' }, [
        h('view', { class: 'section-title' }, [
          h('view', { class: 'section-title-main' }, [
            h('text', {}, 'HRBP评价'),
            h('text', { class: 'section-meta' }, hrbpReviewTitleMeta(ctx, props.review))
          ]),
          reviewGradeBadge('HRBP分档', props.review.hrbpGrade)
        ]),
        h('view', { class: props.editable ? 'form-grid hrbp-review-grid hrbp-review-block' : 'manager-review-readonly hrbp-review-block' }, [
          props.editable ? h('view', { class: 'field-block hrbp-grade-field' }, [
            h('text', { class: 'field-label' }, 'HRBP分档'),
            h('view', { class: 'hrbp-grade-row' }, [
              h('select', { class: 'field-select', value: props.review.hrbpGrade, onChange: (event) => { props.review.hrbpGrade = event.target.value } }, ctx.gradeOptions().map((grade) => h('option', { value: grade }, grade || '未选择'))),
              gradeMismatchNotice()
            ])
          ]) : null,
          h('view', { class: 'field-block field-block-wide hrbp-comment-field' }, [
            props.editable ? h('text', { class: 'field-label' }, '评价内容') : null,
            h('textarea', { class: 'field-textarea', value: props.review.hrbpComment, disabled: !props.editable, placeholder: hrbpPending && !props.editable ? hrbpPendingText : '填写 HRBP 评价', onInput: (event) => { props.review.hrbpComment = readInputValue(event) } })
          ])
        ]),
        !props.editable ? gradeMismatchNotice() : null,
        h(ValueSection, { review: props.review, field: 'hrbp', title: 'HRBP价值观评分', editable: props.editable })
      ])
    }
  }
}

const EmployeeConfirmSection = {
  props: ['review', 'editable'],
  setup(props) {
    return () => h('section', { class: 'form-section' }, [
      h('view', { class: 'section-title' }, [
        h('text', {}, '员工确认'),
        h('text', { class: 'count-pill' }, props.review.employeeConfirmResult === 'confirmed' ? '已确认' : props.review.employeeConfirmResult === 'disputed' ? '有异议' : '待确认')
      ]),
      h('textarea', { class: 'field-textarea', value: props.review.employeeConfirmComment, disabled: !props.editable, placeholder: '确认可简单说明；如提出异议，请填写原因。', onInput: (event) => { props.review.employeeConfirmComment = readInputValue(event) } })
    ])
  }
}

const FinalSection = {
  props: ['review', 'editable'],
  setup(props) {
    const ctx = usePerformanceContext()

    return () => h('section', { class: 'form-section' }, [
      h('view', { class: 'section-title' }, [h('text', {}, 'HRBP归档')]),
      h('view', { class: 'form-grid' }, [
        h('select', { class: 'field-select', value: ctx.effectiveGrade(props.review), disabled: !props.editable, onChange: (event) => { props.review.finalGrade = event.target.value } }, ctx.gradeOptions().map((grade) => h('option', { value: grade }, grade || '未选择'))),
        h('textarea', { class: 'field-textarea', value: props.review.finalNote, disabled: !props.editable, placeholder: 'HRBP备注', onInput: (event) => { props.review.finalNote = readInputValue(event) } })
      ])
    ])
  }
}

const HistorySection = {
  props: ['review'],
  setup(props) {
    return () => h('section', { class: 'form-section' }, [
      h('view', { class: 'section-title' }, [h('text', {}, '流转记录')]),
      h('view', { class: 'history-list' }, (props.review.history || []).slice().reverse().map((item) => {
        const actionParts = historyActionParts(item.action)
        return h('view', { class: 'history-row' }, [
          h('text', {}, actionParts.title || item.action),
          actionParts.reason ? h('text', { class: 'history-row-reason' }, `理由：${actionParts.reason}`) : null,
          h('text', { class: 'muted' }, `${item.by} · ${item.at ? new Date(item.at).toLocaleString() : ''}`)
        ])
      }))
    ])
  }
}

export default WorkbenchView
