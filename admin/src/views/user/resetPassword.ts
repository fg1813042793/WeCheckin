import { ElMessage, ElMessageBox } from 'element-plus'
import { adminApi } from '../../api'
import type { AdminUser } from '../../api/types'
import { showRequestError } from '../../utils/request'

export async function resetPwd(row: Pick<AdminUser, 'id' | 'name'>) {
  try {
    await ElMessageBox.prompt(
      `请为用户「${row.name}」设置新密码。重置成功后，该用户需要重新登录。`,
      '重置密码',
      {
        inputType: 'password',
        inputPlaceholder: '请输入新密码',
        confirmButtonText: '确认重置',
        cancelButtonText: '取消',
        closeOnClickModal: false,
        inputValidator: value => {
          if (!value || !value.trim()) return '请输入新密码'
          if ([...value].length < 6) return '密码至少需要 6 位'
          if (new TextEncoder().encode(value).length > 72) return '密码不能超过 72 字节'
          return true
        },
        beforeClose: async (action, instance, done) => {
          if (instance.confirmButtonLoading) return
          if (action !== 'confirm') { done(); return }
          instance.confirmButtonLoading = true
          try {
            await adminApi.userResetPwd({ id: row.id, password: instance.inputValue ?? '' })
            done()
          } catch (error) {
            showRequestError(error, '密码重置失败，请稍后重试')
          } finally {
            instance.confirmButtonLoading = false
          }
        },
      },
    )
    ElMessage.success('密码已重置')
  } catch (error) {
    if (error !== 'cancel' && error !== 'close') showRequestError(error, '密码重置失败，请稍后重试')
  }
}
