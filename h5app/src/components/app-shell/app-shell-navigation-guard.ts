export interface UnsavedNavigationOptions {
  activeKey: string
  targetKey: string
  guardUnsavedChanges?: boolean
  hasUnsavedChanges: (key: string) => boolean
  confirmLeave: () => Promise<boolean>
  navigate: () => void
}

export interface UnsavedNavigationCopy {
  title: string
  content: string
  confirm: string
  cancel: string
}

export function confirmUnsavedNavigation(copy: UnsavedNavigationCopy) {
  return new Promise<boolean>((resolve) => {
    uni.showModal({
      title: copy.title,
      content: copy.content,
      confirmText: copy.confirm,
      cancelText: copy.cancel,
      success: result => resolve(Boolean(result.confirm)),
      fail: () => resolve(false),
    })
  })
}

export async function navigateWithUnsavedGuard(options: UnsavedNavigationOptions) {
  const shouldConfirm = options.guardUnsavedChanges !== false
    && options.targetKey !== options.activeKey
    && options.hasUnsavedChanges(options.activeKey)

  if (shouldConfirm && !await options.confirmLeave())
    return false

  options.navigate()
  return true
}
