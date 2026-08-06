export const statusMeta = {
  draft: { label: '员工填写', tone: 'warning', step: 0 },
  manager_review: { label: '上级评价', tone: 'purple', step: 1 },
  hrbp_review: { label: 'HRBP评价', tone: 'blue', step: 2 },
  employee_confirm: { label: '员工确认', tone: 'warning', step: 3 },
  hr_final: { label: 'HRBP归档', tone: 'orange', step: 4 },
  completed: { label: '已完成', tone: 'green', step: 5 }
}

export const myPerformanceStatuses = ['draft', 'manager_review', 'hrbp_review', 'employee_confirm', 'hr_final']
export const myPerformanceStatusSet = new Set(myPerformanceStatuses)

export const reviewActionApiPermissions = {
  'save-self': 'dingtalk_h5:api:review:self_save',
  'submit-self': 'dingtalk_h5:api:review:self_submit',
  'submit-manager': 'dingtalk_h5:api:review:manager_submit',
  'submit-hrbp': 'dingtalk_h5:api:review:hrbp_submit',
  'confirm-result': 'dingtalk_h5:api:review:confirm',
  'dispute-result': 'dingtalk_h5:api:review:dispute',
  withdraw: 'dingtalk_h5:api:review:withdraw',
  'return-employee': 'dingtalk_h5:api:review:return_employee',
  'return-manager': 'dingtalk_h5:api:review:return_manager',
  'return-hrbp': 'dingtalk_h5:api:review:return_hrbp',
  finalize: 'dingtalk_h5:api:review:finalize'
}

export const reviewActionButtonPermissions = {
  'save-self': 'dingtalk_h5:button:review:self_save',
  'submit-self': 'dingtalk_h5:button:review:self_submit',
  'submit-manager': 'dingtalk_h5:button:review:manager_submit',
  'submit-hrbp': 'dingtalk_h5:button:review:hrbp_submit',
  'confirm-result': 'dingtalk_h5:button:review:confirm',
  'dispute-result': 'dingtalk_h5:button:review:dispute',
  withdraw: 'dingtalk_h5:button:review:withdraw',
  'return-employee': 'dingtalk_h5:button:review:return_employee',
  'return-manager': 'dingtalk_h5:button:review:return_manager',
  'return-hrbp': 'dingtalk_h5:button:review:return_hrbp',
  finalize: 'dingtalk_h5:button:review:finalize'
}

export const reviewActionConfirmCopy = {
  'submit-self': {
    title: '提交自评',
    content: '确认提交当前绩效？提交后将进入上级评价流程。',
    confirmText: '提交'
  },
  'submit-manager': {
    title: '提交给HRBP',
    content: '确认提交给 HRBP？提交后将进入 HRBP 评价流程。',
    confirmText: '提交'
  },
  'submit-hrbp': {
    title: '提交给员工',
    content: '确认提交给员工确认？提交后员工将查看并确认绩效结果。',
    confirmText: '提交'
  },
  'confirm-result': {
    title: '确认结果',
    content: '确认绩效结果无误？确认后将进入 HRBP 归档流程。',
    confirmText: '确认'
  },
  'dispute-result': {
    title: '提出异议',
    content: '确认提交异议？提交后将返回 HRBP 处理。',
    confirmText: '提交',
    confirmColor: '#e34d59'
  },
  'return-employee': {
    title: '退回员工',
    content: '确认退回员工修改？退回后员工可重新编辑自评内容。',
    confirmText: '退回',
    confirmColor: '#e34d59'
  },
  'return-manager': {
    title: '退回上级',
    content: '确认退回上级修改？退回后上级可重新调整评价。',
    confirmText: '退回',
    confirmColor: '#e34d59'
  },
  'return-hrbp': {
    title: '退回HRBP',
    content: '确认退回 HRBP 修改？退回后 HRBP 可重新处理评价。',
    confirmText: '退回',
    confirmColor: '#e34d59'
  },
  finalize: {
    title: '绩效归档',
    content: '确认归档绩效结果？归档后流程将完成。',
    confirmText: '归档'
  },
  'delete-review': {
    title: '删除考评单',
    content: '确认删除这张考评单？删除后将不再展示。',
    confirmText: '删除',
    confirmColor: '#e34d59'
  }
}

export const menuPageKeys = new Set([
  'dashboard',
  'performance',
  'performance:mine',
  'performance:history',
  'performance:manager',
  'performance:hrbp',
  'performance:summary',
  'performance:org',
  'performance:template'
])

export const historyMonthOptions = Array.from({ length: 12 }, (_, index) => String(index + 1).padStart(2, '0'))
