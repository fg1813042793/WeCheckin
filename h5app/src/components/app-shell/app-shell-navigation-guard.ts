export interface UnsavedNavigationOptions {
  activeKey: string
  targetKey: string
  guardUnsavedChanges?: boolean
  hasUnsavedChanges: (key: string) => boolean
  confirmLeave: () => Promise<boolean>
  navigate: () => void
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
