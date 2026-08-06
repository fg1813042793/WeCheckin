export function currentAssignee(ctx, review) {
  if (review.status === 'draft') return ctx.userName(review.employeeId)
  if (review.status === 'manager_review') return ctx.userName(review.managerId)
  if (review.status === 'hrbp_review') return ctx.userName(review.hrbpReviewerId || review.hrbpId)
  if (review.status === 'employee_confirm') return ctx.userName(review.employeeId)
  if (review.status === 'hr_final') return ctx.userName(review.hrbpReviewerId || review.hrbpId)
  return '已归档'
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

export function historyActionParts(action) {
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

function reviewDisputeReasonForStep(review, step) {
  if (step.status !== 'hrbp_review') return ''
  if (review.employeeConfirmResult !== 'disputed') return ''
  return String(review.employeeConfirmComment || '').trim()
}

function historyDetailMeta(history, time) {
  return [history?.by || 'system', time].filter(Boolean).join(' · ')
}

export function flowProgressRows(ctx, review) {
  const currentStep = ctx.statusMeta[review.status]?.step || 0
  return flowSteps.map((step, index) => {
    const history = latestHistoryForStep(review, step)
    const isCompletedFinal = review.status === 'completed' && index === currentStep
    const state = index < currentStep || isCompletedFinal ? 'done' : index === currentStep ? 'active' : 'pending'
    const latestAction = String(review.latestAction || '').trim()
    const shouldUseLatestAction = state === 'active' && (ctx.returnStatusTextFromAction(latestAction) || latestAction.includes('员工提出异议'))
    const historyAction = history?.action || (shouldUseLatestAction ? latestAction : '')
    const time = formatHistoryTime(history?.at)
    const actionParts = historyActionParts(historyAction)
    const returnStatusText = state === 'active' ? ctx.returnStatusTextFromAction(actionParts.title || historyAction) : ''
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
    const detailMeta = history ? historyDetailMeta(history, time) : ''
    const detail = historyAction
      ? [actionParts.title || historyAction, detailMeta].filter(Boolean).join(' · ')
      : disputeReason && state === 'active'
        ? `员工提出异议，等待 ${stepActorName(ctx, review, step)} 处理`
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
      stateClass,
      stateLabel,
      returnStatusText,
      detail,
      reason,
      reasonLabel,
      history
    }
  })
}

