export const COMPLETION_REDIRECT_DELAY_MS = 2000

type RedirectScheduler = (callback: () => void, delay: number) => ReturnType<typeof setTimeout>

export function scheduleCompletionRedirect(
  url: unknown,
  navigate: (target: string) => void,
  schedule: RedirectScheduler = (callback, delay) => setTimeout(callback, delay),
): ReturnType<typeof setTimeout> | null {
  const target = typeof url === 'string' ? url.trim() : ''
  if (!target) return null
  return schedule(() => navigate(target), COMPLETION_REDIRECT_DELAY_MS)
}
