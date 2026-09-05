import type { RegisteredAppPage } from '@/config/app-navigation'
import { FEEDBACK_CONTENT_KEY } from './feedback-route-keys'

export const FEEDBACK_MENU_PERMISSION_KEY = 'dingtalk_h5:menu:feedback'

function appContentRoute(key: string) {
  return `/pages/index/index?view=${encodeURIComponent(key)}`
}

export const feedbackRootNavItem = {
  key: 'feedback',
  label: '我的反馈',
  icon: 'chat',
  permissionKey: FEEDBACK_MENU_PERMISSION_KEY,
}

export const feedbackMenuPages: RegisteredAppPage[] = [
  {
    key: 'feedback',
    contentKey: FEEDBACK_CONTENT_KEY,
    route: appContentRoute(FEEDBACK_CONTENT_KEY),
    title: '我的反馈',
    description: '提交和查看自己的反馈',
    icon: 'chat',
    permissionKey: 'dingtalk_h5:menu:feedback',
    rootKey: 'feedback',
  },
]
