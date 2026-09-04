export interface DepartmentEntity {
  department?: unknown
  departmentLevel1?: unknown
  departmentLevel2?: unknown
  departmentLevel3?: unknown
  departmentLevel4?: unknown
  departmentLevels?: unknown
}

function firstText(...values: Array<unknown>) {
  for (const value of values) {
    const text = String(value || '').trim()
    if (text) {
      return text
    }
  }
  return ''
}

function normalizeDepartmentLevels(levels: unknown[]) {
  return levels.map(item => String(item || '').trim()).filter(Boolean)
}

function splitDepartmentText(value: unknown) {
  const text = firstText(value)
  if (!text) {
    return []
  }
  return normalizeDepartmentLevels(text.split(' / '))
}

export function departmentLevelsFromEntity(entity?: DepartmentEntity | null) {
  if (!entity) {
    return []
  }
  if (Array.isArray(entity.departmentLevels)) {
    const levels = normalizeDepartmentLevels(entity.departmentLevels)
    if (levels.length > 0) {
      return levels
    }
  }
  const textParts = splitDepartmentText(entity.department)
  const levels = normalizeDepartmentLevels([
    firstText(entity.departmentLevel1, textParts[0]),
    firstText(entity.departmentLevel2, textParts[1]),
    firstText(entity.departmentLevel3, textParts[2]),
    firstText(entity.departmentLevel4, textParts[3]),
  ])
  return levels.length > 0 ? levels : textParts
}

export function departmentPathFromEntity(entity?: DepartmentEntity | null, fallback = '-') {
  const explicit = firstText(entity?.department)
  if (explicit) {
    return explicit
  }
  const path = departmentLevelsFromEntity(entity).join(' / ')
  return path || fallback
}

export function departmentLeafFromEntity(entity?: DepartmentEntity | null) {
  const levels = departmentLevelsFromEntity(entity)
  if (levels.length > 0) {
    return levels[levels.length - 1]
  }
  return firstText(entity?.department)
}
