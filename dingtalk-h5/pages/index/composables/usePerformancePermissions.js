import {
  reviewActionApiPermissions,
  reviewActionButtonPermissions
} from '../../../views/performance/common/constants'

export function usePerformancePermissions({
  state,
  flatMenuItems,
  sameUserId
}) {
  function hasApiPermission(key) {
    if (!state.apiPermissionReady) return false
    return state.apiPermissionKeys.includes(key)
  }

  function hasButtonPermission(key) {
    if (!state.buttonPermissionReady) return false
    return state.buttonPermissionKeys.includes(key)
  }

  function hasMenuPermission(key) {
    return flatMenuItems.value.some((item) => item.permissionKey === key)
  }

  function canCreateReview() {
    return hasButtonPermission('dingtalk_h5:button:review:create') &&
      hasApiPermission('dingtalk_h5:api:review:create')
  }

  function canDeleteReview() {
    return hasButtonPermission('dingtalk_h5:button:review:delete') &&
      hasApiPermission('dingtalk_h5:api:review:delete')
  }

  function canExportReviews() {
    return hasButtonPermission('dingtalk_h5:button:review:export') &&
      hasApiPermission('dingtalk_h5:api:review:export')
  }

  function canEditTemplate() {
    return hasButtonPermission('dingtalk_h5:button:template:edit') &&
      hasApiPermission('dingtalk_h5:api:template:save')
  }

  function canEditUsers() {
    return hasButtonPermission('dingtalk_h5:button:user:config') &&
      hasApiPermission('dingtalk_h5:api:user:edit')
  }

  function canPerformReviewAction(action) {
    const apiPermission = reviewActionApiPermissions[action]
    const buttonPermission = reviewActionButtonPermissions[action]
    return Boolean(apiPermission && buttonPermission && hasApiPermission(apiPermission) && hasButtonPermission(buttonPermission))
  }

  function canSelf(review) {
    return sameUserId(review?.employeeId, state.user?.id) &&
      review.status === 'draft' &&
      (canPerformReviewAction('save-self') || canPerformReviewAction('submit-self'))
  }

  function canManager(review) {
    return sameUserId(review?.managerId, state.user?.id) &&
      review.status === 'manager_review' &&
      canPerformReviewAction('submit-manager')
  }

  function isHrbpActor(review) {
    if (!review) return false
    if (review.hrbpReviewerId) return sameUserId(review.hrbpReviewerId, state.user?.id)
    return sameUserId(review.hrbpId, state.user?.id)
  }

  function canHrbpHandle(review) {
    return review?.status === 'hrbp_review' &&
      isHrbpActor(review) &&
      canPerformReviewAction('submit-hrbp')
  }

  function canEmployeeConfirm(review) {
    return sameUserId(review?.employeeId, state.user?.id) &&
      review.status === 'employee_confirm' &&
      (canPerformReviewAction('confirm-result') || canPerformReviewAction('dispute-result'))
  }

  function canFinal(review) {
    if (!review || review.status !== 'hr_final') return false
    if (!canPerformReviewAction('finalize')) return false
    if (review.hrbpReviewerId) return sameUserId(review.hrbpReviewerId, state.user?.id)
    return sameUserId(review.hrbpId, state.user?.id)
  }

  function canWithdraw(review) {
    if (!canPerformReviewAction('withdraw')) return false
    if (!review) return false
    if (review.status === 'manager_review' && sameUserId(review.employeeId, state.user?.id)) return true
    if (review.status === 'hrbp_review' && sameUserId(review.managerId, state.user?.id)) return true
    if (review.status === 'employee_confirm' && isHrbpActor(review)) return true
    if (review.status === 'hr_final' && sameUserId(review.employeeId, state.user?.id) && !review.finalGrade) return true
    if (review.status === 'hr_final' && canFinal(review) && !review.finalGrade) return true
    return false
  }

  function canEditObjectiveDimension(review) {
    return canSelf(review) && !review.objectiveSourceReviewId
  }

  function canEditNextObjectives(review) {
    return canSelf(review) &&
      hasButtonPermission('dingtalk_h5:button:review:next_objective_edit')
  }

  function canAddNextObjective(review) {
    return canEditNextObjectives(review) &&
      hasButtonPermission('dingtalk_h5:button:review:next_objective_add')
  }

  function canDeleteNextObjective(review) {
    return canEditNextObjectives(review) &&
      hasButtonPermission('dingtalk_h5:button:review:next_objective_delete')
  }

  return {
    canAddNextObjective,
    canCreateReview,
    canDeleteNextObjective,
    canDeleteReview,
    canEditNextObjectives,
    canEditObjectiveDimension,
    canEditTemplate,
    canEditUsers,
    canEmployeeConfirm,
    canExportReviews,
    canFinal,
    canHrbpHandle,
    canManager,
    canPerformReviewAction,
    canSelf,
    canWithdraw,
    hasApiPermission,
    hasButtonPermission,
    hasMenuPermission
  }
}
