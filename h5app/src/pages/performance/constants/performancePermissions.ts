export type ReviewActionKey
  = | 'save-self'
    | 'submit-self'
    | 'submit-manager'
    | 'submit-hrbp'
    | 'confirm-result'
    | 'dispute-result'
    | 'withdraw'
    | 'return-employee'
    | 'return-manager'
    | 'return-hrbp'
    | 'finalize'

export const reviewActionApiPermissions: Record<ReviewActionKey, string> = {
  'save-self': 'dingtalk_h5:api:review:self_save',
  'submit-self': 'dingtalk_h5:api:review:self_submit',
  'submit-manager': 'dingtalk_h5:api:review:manager_submit',
  'submit-hrbp': 'dingtalk_h5:api:review:hrbp_submit',
  'confirm-result': 'dingtalk_h5:api:review:confirm',
  'dispute-result': 'dingtalk_h5:api:review:dispute',
  'withdraw': 'dingtalk_h5:api:review:withdraw',
  'return-employee': 'dingtalk_h5:api:review:return_employee',
  'return-manager': 'dingtalk_h5:api:review:return_manager',
  'return-hrbp': 'dingtalk_h5:api:review:return_hrbp',
  'finalize': 'dingtalk_h5:api:review:finalize',
}

export const reviewActionButtonPermissions: Record<ReviewActionKey, string> = {
  'save-self': 'dingtalk_h5:button:review:self_save',
  'submit-self': 'dingtalk_h5:button:review:self_submit',
  'submit-manager': 'dingtalk_h5:button:review:manager_submit',
  'submit-hrbp': 'dingtalk_h5:button:review:hrbp_submit',
  'confirm-result': 'dingtalk_h5:button:review:confirm',
  'dispute-result': 'dingtalk_h5:button:review:dispute',
  'withdraw': 'dingtalk_h5:button:review:withdraw',
  'return-employee': 'dingtalk_h5:button:review:return_employee',
  'return-manager': 'dingtalk_h5:button:review:return_manager',
  'return-hrbp': 'dingtalk_h5:button:review:return_hrbp',
  'finalize': 'dingtalk_h5:button:review:finalize',
}
