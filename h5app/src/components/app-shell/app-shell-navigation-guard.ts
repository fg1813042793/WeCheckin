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

export function createUnsavedNavigationCoordinator() {
  let navigationToken = 0
  let pendingConfirmation: Promise<boolean> | null = null

  return {
    async navigate(options: UnsavedNavigationOptions) {
      const token = ++navigationToken
      const shouldConfirm = options.guardUnsavedChanges !== false
        && options.targetKey !== options.activeKey
        && options.hasUnsavedChanges(options.activeKey)

      if (shouldConfirm) {
        if (!pendingConfirmation) {
          const confirmation = Promise.resolve()
            .then(() => options.confirmLeave())
            .then(Boolean, () => false)
          pendingConfirmation = confirmation
          void confirmation.finally(() => {
            if (pendingConfirmation === confirmation)
              pendingConfirmation = null
          })
        }

        const confirmed = await pendingConfirmation
        if (!confirmed || token !== navigationToken)
          return false
      }
      else if (token !== navigationToken) {
        return false
      }

      options.navigate()
      return true
    },
  }
}

const sharedUnsavedNavigationCoordinator = createUnsavedNavigationCoordinator()

export function navigateWithUnsavedGuard(options: UnsavedNavigationOptions) {
  return sharedUnsavedNavigationCoordinator.navigate(options)
}
