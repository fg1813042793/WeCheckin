import type { PerformanceReview, ReviewActionRequest, ReviewStatus } from '@/types/dingtalk-h5'

export const statusMeta: Record<string, { label: string, type: 'primary' | 'success' | 'warning' | 'error' | 'info', step: number }> = {
  draft: { label: '员工填写', type: 'warning', step: 0 },
  manager_review: { label: '上级评价', type: 'primary', step: 1 },
  hrbp_review: { label: 'HRBP评价', type: 'primary', step: 2 },
  employee_confirm: { label: '员工确认', type: 'warning', step: 3 },
  hr_final: { label: 'HRBP归档', type: 'primary', step: 4 },
  completed: { label: '已完成', type: 'success', step: 5 },
}

export const myPerformanceStatuses = ['draft', 'manager_review', 'hrbp_review', 'employee_confirm', 'hr_final']

export const myPerformanceStatusSet = new Set(myPerformanceStatuses)

export const withdrawPreviousStatusMap: Record<string, ReviewStatus> = {
  manager_review: 'draft',
  hrbp_review: 'manager_review',
  employee_confirm: 'hrbp_review',
  hr_final: 'employee_confirm',
}

function firstText(...values: unknown[]) {
  for (const value of values) {
    const text = String(value || '').trim()
    if (text) {
      return text
    }
  }
  return ''
}

export function resolveWithdrawTargetStatus(status: unknown): ReviewStatus | '' {
  const statusText = firstText(status)
  return withdrawPreviousStatusMap[statusText] || ''
}

export function normalizeReviewActionResult(review: PerformanceReview | void, action: string, data?: ReviewActionRequest) {
  if (!review || action !== 'withdraw') {
    return review
  }
  const sourceStatus = firstText(data?.fromStatus, data?.sourceStatus, review.status)
  const targetStatus = firstText(data?.targetStatus) || resolveWithdrawTargetStatus(sourceStatus)
  if (!targetStatus) {
    return review
  }
  return {
    ...review,
    status: targetStatus,
  }
}
