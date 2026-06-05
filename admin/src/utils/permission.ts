const STORAGE_KEY = 'admin_perms'
let cachedPerms: string[] | null = null

export function getPerms(): string[] {
  if (cachedPerms !== null) return cachedPerms
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return []
    const parsed = JSON.parse(raw)
    cachedPerms = Array.isArray(parsed) ? parsed : []
  } catch {
    cachedPerms = []
  }
  return cachedPerms
}

export function setPerms(perms: string[] | null | undefined) {
  cachedPerms = (perms || [])
  localStorage.setItem(STORAGE_KEY, JSON.stringify(cachedPerms))
}

export function clearPerms() {
  cachedPerms = null
  localStorage.removeItem(STORAGE_KEY)
}

export function hasPerm(perm: string): boolean {
  return getPerms().includes(perm)
}
