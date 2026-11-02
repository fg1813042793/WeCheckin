import { computed, onBeforeUnmount, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { adminApi } from '@/api'
import type { UserFeedbackDetail, UserFeedbackSummary } from '@/types/userFeedback'
import { showRequestError } from '@/utils/request'
import { canHandleUserFeedback } from '../userFeedbackStatus'

interface FeedbackActionDialog {
  open: (detail: UserFeedbackDetail) => void
  visible: boolean
}

type FeedbackAction = 'status' | 'notify'

export function useUserFeedbackActions() {
  const statusDialogRef = ref<FeedbackActionDialog>()
  const notifyDialogRef = ref<FeedbackActionDialog>()
  const actionLoading = ref<{ id: string; action: FeedbackAction } | null>(null)
  const actionError = ref('')
  const canHandle = computed(() => canHandleUserFeedback())
  let sequence = 0

  function cancelPendingAction() {
    sequence += 1
    actionLoading.value = null
    actionError.value = ''
  }

  async function openAction(feedback: UserFeedbackSummary, action: FeedbackAction) {
    if (!canHandleUserFeedback()) {
      ElMessage.warning('当前账号没有反馈处理权限')
      return
    }
    if (statusDialogRef.value?.visible || notifyDialogRef.value?.visible) return
    if (actionLoading.value?.id === feedback.id && actionLoading.value.action === action) return
    const current = ++sequence
    actionLoading.value = { id: feedback.id, action }
    actionError.value = ''
    try {
      const response = await adminApi.userFeedbackDetail(feedback.id)
      if (current !== sequence) return
      if (!canHandleUserFeedback()) {
        ElMessage.warning('当前账号没有反馈处理权限')
        return
      }
      const dialog = action === 'status' ? statusDialogRef.value : notifyDialogRef.value
      dialog?.open(response.data)
    } catch (error) {
      if (current !== sequence) return
      actionError.value = `反馈 ${feedback.feedbackNo} 详情加载失败，请重试`
      showRequestError(error, '反馈详情加载失败')
    } finally {
      if (current === sequence) actionLoading.value = null
    }
  }

  onBeforeUnmount(cancelPendingAction)

  return {
    statusDialogRef, notifyDialogRef, canHandle, actionLoading, actionError, cancelPendingAction,
    openStatusDialog: (feedback: UserFeedbackSummary) => openAction(feedback, 'status'),
    openNotifyDialog: (feedback: UserFeedbackSummary) => openAction(feedback, 'notify'),
  }
}
