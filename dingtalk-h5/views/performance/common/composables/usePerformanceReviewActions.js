import { computed, reactive } from 'vue'
import { reviewAction as submitReviewAction } from '../../../../api/performance/common/reviews'
import { reviewActionConfirmCopy } from '../constants'

export function usePerformanceReviewActions({
  selectedReview,
  canPerformReviewAction,
  loadReviews,
  updateReview,
  toast
}) {
  const withdrawDialog = reactive({ visible: false, loading: false, reason: '' })
  const returnDialog = reactive({
    visible: false,
    loading: false,
    action: '',
    title: '退回',
    desc: '',
    reasonLabel: '退回原因',
    placeholder: '请填写退回原因',
    reason: ''
  })
  const disputeDialog = reactive({ visible: false, loading: false, reason: '' })

  const withdrawReasonLength = computed(() => [...String(withdrawDialog.reason || '')].length)
  const returnReasonLength = computed(() => [...String(returnDialog.reason || '')].length)
  const disputeReasonLength = computed(() => [...String(disputeDialog.reason || '')].length)

  function reviewPayload(review) {
    return {
      objectives: review.objectives,
      nextObjectives: review.nextObjectives,
      values: review.values,
      selfSummary: review.selfSummary,
      managerComment: review.managerComment,
      managerGrade: review.managerGrade,
      hrbpComment: review.hrbpComment,
      hrbpGrade: review.hrbpGrade,
      employeeConfirmComment: review.employeeConfirmComment,
      finalGrade: review.finalGrade,
      finalNote: review.finalNote
    }
  }

  function hasFilledRequiredValue(value) {
    if (value === null || value === undefined) return false
    if (typeof value === 'number') return Number.isFinite(value)
    return String(value).trim() !== ''
  }

  function optionalPercentValidationMessage(value, label) {
    if (value === null || value === undefined || String(value).trim() === '') return ''
    const num = Number(value)
    if (!Number.isFinite(num) || num < 0 || num > 100) {
      return `${label}必须在 0-100 之间`
    }
    return ''
  }

  function objectiveWeightValue(item) {
    if (item?.weight === null || item?.weight === undefined || String(item.weight).trim() === '') return 0
    return Number(item.weight)
  }

  function objectiveWeightValidationMessage(items, label) {
    if (!Array.isArray(items) || !items.length) return ''
    let total = 0
    for (let index = 0; index < items.length; index += 1) {
      const value = objectiveWeightValue(items[index])
      if (!Number.isFinite(value) || value < 0 || value > 100) {
        return `${label} ${index + 1} 的权重必须在 0-100 之间`
      }
      total += value
    }
    if (total > 100) {
      return `${label}权重合计不能大于 100`
    }
    return ''
  }

  function currentObjectiveCompletionValidationMessage(items) {
    if (!Array.isArray(items) || !items.length) return ''
    for (let index = 0; index < items.length; index += 1) {
      const message = optionalPercentValidationMessage(items[index]?.completion, `本月目标 ${index + 1} 的完成度`)
      if (message) return message
    }
    return ''
  }

  function validateSelfObjectiveNumbers(review) {
    const currentObjectives = Array.isArray(review?.objectives) ? review.objectives : []
    const nextObjectives = Array.isArray(review?.nextObjectives) ? review.nextObjectives : []
    return objectiveWeightValidationMessage(currentObjectives, '本月目标') ||
      currentObjectiveCompletionValidationMessage(currentObjectives) ||
      objectiveWeightValidationMessage(nextObjectives, '下月目标')
  }

  function hasRequiredCurrentObjectives(review) {
    const objectives = Array.isArray(review?.objectives) ? review.objectives : []
    if (!objectives.length) return false
    return objectives.every((item) =>
      String(item?.target || '').trim() !== '' &&
      hasFilledRequiredValue(item?.completion) &&
      String(item?.result || '').trim() !== ''
    )
  }

  function hasRequiredSelfValues(review) {
    const values = Array.isArray(review?.values) ? review.values : []
    if (!values.length) return false
    return values.every((item) => hasFilledRequiredValue(item?.self))
  }

  function hasRequiredNextObjectives(review) {
    const nextObjectives = Array.isArray(review?.nextObjectives) ? review.nextObjectives : []
    if (!nextObjectives.length) return false
    return nextObjectives.every((item) => String(item?.target || '').trim() !== '')
  }

  function selfSubmitRequiredMessage(review) {
    const missing = []
    if (!hasRequiredCurrentObjectives(review)) missing.push('本月目标')
    if (String(review?.selfSummary || '').trim() === '') missing.push('思考总结')
    if (!hasRequiredSelfValues(review)) missing.push('价值观自评')
    if (!hasRequiredNextObjectives(review)) missing.push('下月目标')
    return missing.length ? `请完善：${missing.join('、')}` : ''
  }

  function validateSelfSubmitReview(review) {
    return selfSubmitRequiredMessage(review)
  }

  function hasRequiredManagerValues(review) {
    const values = Array.isArray(review?.values) ? review.values : []
    if (!values.length) return false
    return values.every((item) => hasFilledRequiredValue(item?.manager))
  }

  function managerSubmitRequiredMessage(review) {
    const missing = []
    if (!hasFilledRequiredValue(review?.managerGrade)) missing.push('上级分档')
    if (String(review?.managerComment || '').trim() === '') missing.push('评价内容')
    if (!hasRequiredManagerValues(review)) missing.push('上级价值观评分')
    return missing.length ? `请完善：${missing.join('、')}` : ''
  }

  function validateManagerSubmitReview(review) {
    return managerSubmitRequiredMessage(review)
  }

  function hasRequiredHrbpValues(review) {
    const values = Array.isArray(review?.values) ? review.values : []
    if (!values.length) return false
    return values.every((item) => hasFilledRequiredValue(item?.hrbp))
  }

  function hrbpSubmitRequiredMessage(review) {
    const missing = []
    if (!hasFilledRequiredValue(review?.hrbpGrade)) missing.push('HRBP分档')
    if (String(review?.hrbpComment || '').trim() === '') missing.push('评价内容')
    if (!hasRequiredHrbpValues(review)) missing.push('HRBP价值观评分')
    return missing.length ? `请完善：${missing.join('、')}` : ''
  }

  function hrbpSubmitGradeMismatchMessage(review) {
    const managerGrade = String(review?.managerGrade || '').trim()
    const hrbpGrade = String(review?.hrbpGrade || '').trim()
    if (!managerGrade) return '上级分档为空，不能提交给员工确认。'
    if (hrbpGrade && hrbpGrade !== managerGrade) {
      return `HRBP分档需与上级分档一致，当前上级分档为「${managerGrade}」，HRBP分档为「${hrbpGrade}」。`
    }
    return ''
  }

  function validateHrbpSubmitReview(review) {
    const requiredMessage = hrbpSubmitRequiredMessage(review)
    if (requiredMessage) return { message: requiredMessage, modal: false }
    const gradeMessage = hrbpSubmitGradeMismatchMessage(review)
    if (gradeMessage) return { title: '分档不一致', message: gradeMessage, modal: true }
    return null
  }

  async function confirmReviewAction(action) {
    const copy = reviewActionConfirmCopy[action]
    if (!copy) return true
    return new Promise((resolve) => {
      if (typeof uni !== 'undefined' && uni.showModal) {
        uni.showModal({
          title: copy.title,
          content: copy.content,
          confirmText: copy.confirmText || '确定',
          confirmColor: copy.confirmColor || '#1677ff',
          success: (res) => resolve(Boolean(res.confirm)),
          fail: () => resolve(false)
        })
        return
      }
      if (typeof window !== 'undefined' && window.confirm) {
        resolve(window.confirm(copy.content))
        return
      }
      resolve(false)
    })
  }

  async function showValidationModal(title, content) {
    return new Promise((resolve) => {
      if (typeof uni !== 'undefined' && uni.showModal) {
        uni.showModal({
          title,
          content,
          showCancel: false,
          confirmText: '知道了',
          confirmColor: '#1677ff',
          success: () => resolve(true),
          fail: () => resolve(false)
        })
        return
      }
      if (typeof window !== 'undefined' && window.alert) {
        window.alert(content)
        resolve(true)
        return
      }
      resolve(false)
    })
  }

  async function performReviewAction(action, successText) {
    if (!selectedReview.value) return
    if (!canPerformReviewAction(action)) {
      toast('无权限操作')
      return
    }
    if (action === 'save-self' || action === 'submit-self') {
      const objectiveValidationMessage = validateSelfObjectiveNumbers(selectedReview.value)
      if (objectiveValidationMessage) {
        toast(objectiveValidationMessage)
        return
      }
    }
    if (action === 'submit-self') {
      const validationMessage = validateSelfSubmitReview(selectedReview.value)
      if (validationMessage) {
        toast(validationMessage)
        return
      }
    }
    if (action === 'submit-manager') {
      const managerValidationMessage = validateManagerSubmitReview(selectedReview.value)
      if (managerValidationMessage) {
        toast(managerValidationMessage)
        return
      }
    }
    if (action === 'submit-hrbp') {
      const hrbpValidationMessage = validateHrbpSubmitReview(selectedReview.value)
      if (hrbpValidationMessage) {
        if (hrbpValidationMessage.modal) {
          await showValidationModal(hrbpValidationMessage.title || '提示', hrbpValidationMessage.message)
        } else {
          toast(hrbpValidationMessage.message || hrbpValidationMessage)
        }
        return
      }
    }
    if (action === 'dispute-result') {
      openDisputeDialog()
      return
    }
    const confirmed = await confirmReviewAction(action)
    if (!confirmed) return
    const res = await submitReviewAction(selectedReview.value.id, action, reviewPayload(selectedReview.value))
    updateReview(res.data)
    await loadReviews()
    toast(successText)
  }

  async function returnReview(action, label) {
    if (!selectedReview.value) return
    if (!canPerformReviewAction(action)) {
      toast('无权限操作')
      return
    }
    openReturnDialog(action, label)
  }

  function returnDialogCopy(action, label) {
    const titleMap = {
      'return-employee': '退回员工',
      'return-manager': '退回上级',
      'return-hrbp': '退回 HRBP'
    }
    const descMap = {
      'return-employee': '退回后员工可重新编辑自评内容，流程记录会保存退回原因。',
      'return-manager': '退回后上级可重新调整评价内容，流程记录会保存退回原因。',
      'return-hrbp': '退回后 HRBP 可重新处理评价内容，流程记录会保存退回原因。'
    }
    return {
      title: titleMap[action] || label || '退回',
      desc: descMap[action] || '退回后流程会返回上一节点，流程记录会保存退回原因。',
      reasonLabel: '退回原因',
      placeholder: '请填写退回原因'
    }
  }

  function openReturnDialog(action, label) {
    const copy = returnDialogCopy(action, label)
    returnDialog.visible = true
    returnDialog.loading = false
    returnDialog.action = action
    returnDialog.title = copy.title
    returnDialog.desc = copy.desc
    returnDialog.reasonLabel = copy.reasonLabel
    returnDialog.placeholder = copy.placeholder
    returnDialog.reason = ''
  }

  function closeReturnDialog() {
    if (returnDialog.loading) return
    returnDialog.visible = false
    returnDialog.action = ''
    returnDialog.reason = ''
  }

  async function submitReturnReview() {
    if (!selectedReview.value) {
      closeReturnDialog()
      return
    }
    const action = returnDialog.action
    if (!canPerformReviewAction(action)) {
      toast('无权限操作')
      return
    }
    const reason = String(returnDialog.reason || '').trim()
    if (!reason) {
      toast('请填写退回原因')
      return
    }
    returnDialog.loading = true
    try {
      const res = await submitReviewAction(selectedReview.value.id, action, {
        ...reviewPayload(selectedReview.value),
        returnReason: reason
      })
      updateReview(res.data)
      await loadReviews()
      returnDialog.loading = false
      closeReturnDialog()
      toast('已退回')
    } catch (error) {
      toast(error?.msg || '退回失败')
    } finally {
      returnDialog.loading = false
    }
  }

  function openDisputeDialog() {
    if (!selectedReview.value) return
    if (!canPerformReviewAction('dispute-result')) {
      toast('无权限操作')
      return
    }
    disputeDialog.reason = String(selectedReview.value.employeeConfirmComment || '').trim()
    disputeDialog.loading = false
    disputeDialog.visible = true
  }

  function closeDisputeDialog() {
    if (disputeDialog.loading) return
    disputeDialog.visible = false
    disputeDialog.reason = ''
  }

  async function submitDisputeReview() {
    if (!selectedReview.value) {
      closeDisputeDialog()
      return
    }
    if (!canPerformReviewAction('dispute-result')) {
      toast('无权限操作')
      return
    }
    const reason = String(disputeDialog.reason || '').trim()
    if (!reason) {
      toast('请填写异议原因')
      return
    }
    disputeDialog.loading = true
    try {
      const res = await submitReviewAction(selectedReview.value.id, 'dispute-result', {
        ...reviewPayload(selectedReview.value),
        employeeConfirmComment: reason
      })
      updateReview(res.data)
      await loadReviews()
      disputeDialog.loading = false
      closeDisputeDialog()
      toast('已提出异议')
    } catch (error) {
      toast(error?.msg || '提交异议失败')
    } finally {
      disputeDialog.loading = false
    }
  }

  function openWithdrawDialog() {
    if (!selectedReview.value) return
    if (!canPerformReviewAction('withdraw')) {
      toast('无权限操作')
      return
    }
    withdrawDialog.reason = ''
    withdrawDialog.loading = false
    withdrawDialog.visible = true
  }

  function closeWithdrawDialog() {
    if (withdrawDialog.loading) return
    withdrawDialog.visible = false
    withdrawDialog.reason = ''
  }

  async function withdrawReview() {
    openWithdrawDialog()
  }

  async function submitWithdrawReview() {
    if (!selectedReview.value) {
      closeWithdrawDialog()
      return
    }
    if (!canPerformReviewAction('withdraw')) {
      toast('无权限操作')
      return
    }
    const reason = String(withdrawDialog.reason || '').trim()
    if (!reason) {
      toast('请填写撤回理由')
      return
    }
    withdrawDialog.loading = true
    try {
      const res = await submitReviewAction(selectedReview.value.id, 'withdraw', {
        returnReason: reason
      })
      updateReview(res.data)
      await loadReviews()
      withdrawDialog.loading = false
      closeWithdrawDialog()
      toast('已撤回')
    } catch (error) {
      toast(error?.msg || '撤回失败')
    } finally {
      withdrawDialog.loading = false
    }
  }

  return {
    withdrawDialog,
    withdrawReasonLength,
    closeWithdrawDialog,
    withdrawReview,
    submitWithdrawReview,
    returnDialog,
    returnReasonLength,
    closeReturnDialog,
    returnReview,
    submitReturnReview,
    disputeDialog,
    disputeReasonLength,
    closeDisputeDialog,
    submitDisputeReview,
    confirmReviewAction,
    performReviewAction
  }
}
