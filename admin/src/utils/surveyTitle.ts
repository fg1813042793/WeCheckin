export function resolveSurveyTitle(title: string | undefined, displayTitle: unknown): string {
  const display = typeof displayTitle === 'string' ? displayTitle.trim() : ''
  return display || title || ''
}
