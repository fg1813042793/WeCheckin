export function userAvatarInitial(...candidates: unknown[]): string {
  for (const candidate of candidates) {
    const text = String(candidate ?? '').trim()
    if (!text)
      continue

    const firstCharacter = Array.from(text)[0]
    if (firstCharacter)
      return firstCharacter.toLocaleUpperCase()
  }

  return 'U'
}
