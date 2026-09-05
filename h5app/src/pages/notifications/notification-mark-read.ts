export interface NotificationMarkReadOptions {
  requireExplicitSuccess?: boolean
  request: () => Promise<unknown>
  commit: () => void
  fail: () => void
}

export async function runNotificationMarkRead(options: NotificationMarkReadOptions) {
  try {
    const response = await options.request()
    if (options.requireExplicitSuccess && !response) {
      options.fail()
      return false
    }
    options.commit()
    return true
  }
  catch {
    options.fail()
    return false
  }
}
