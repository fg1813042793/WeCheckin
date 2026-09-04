import type { RegisteredAppPage } from '@/config/app-navigation'

function appContentRoute(key: string) {
  return `/pages/index/index?view=${encodeURIComponent(key)}`
}

export const workflowRootNavItem = {
  key: 'workflow',
  label: '流程审批',
  icon: 'checkmark-circle',
}

export const workflowMenuPages: RegisteredAppPage[] = [
  {
    key: 'workflow',
    contentKey: 'workflow',
    route: appContentRoute('workflow'),
    title: '流程审批',
    description: '发起和处理通用审批流程',
    icon: 'checkmark-circle',
    rootKey: 'workflow',
  },
]
