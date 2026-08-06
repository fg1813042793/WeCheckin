import { firstText } from '../../common/helpers'

export { firstText }

export function buildDepartmentTree(users = []) {
  const root = new Map()
  for (const user of users || []) {
    const levels = realDepartmentLevels(user)
    let currentMap = root
    let currentNode = null
    for (const [index, level] of levels.entries()) {
      currentNode = ensureTreeNode(currentMap, `${currentNode?.key || 'root'}/l${index + 1}:${level}`, level)
      currentNode.count += 1
      currentMap = currentNode.childMap
    }
    if (currentNode) {
      currentNode.users.push(user)
    }
  }
  return finalizeTreeNodes([...root.values()])
}

function ensureTreeNode(map, key, name) {
  if (!map.has(key)) {
    map.set(key, { key, name, count: 0, childMap: new Map(), children: [], users: [] })
  }
  return map.get(key)
}

function finalizeTreeNodes(nodes) {
  return nodes
    .sort((left, right) => left.name.localeCompare(right.name, 'zh-Hans-CN'))
    .map((node) => ({
      key: node.key,
      name: node.name,
      count: node.count,
      children: finalizeTreeNodes([...node.childMap.values()]),
      users: sortUsers(node.users)
    }))
}

export function flattenDepartmentTree(nodes = [], expandedKeys, depth = 1, parentPath = '') {
  const rows = []
  for (const node of nodes) {
    const path = [parentPath, node.name].filter(Boolean).join(' / ')
    const hasChildren = node.children.length > 0 || node.users.length > 0
    rows.push({
      type: 'department',
      key: node.key,
      depth,
      name: node.name,
      path,
      count: node.count,
      expandable: hasChildren,
      hasChildren,
      expanded: expandedKeys.has(node.key)
    })
    if (!expandedKeys.has(node.key)) continue
    for (const user of node.users) {
      rows.push({ type: 'employee', key: `${node.key}/user:${user.id}`, depth: depth + 1, user })
    }
    rows.push(...flattenDepartmentTree(node.children, expandedKeys, depth + 1, path))
  }
  return rows
}

export function filterDepartmentTree(nodes = [], keyword = '') {
  const search = normalizeText(keyword)
  if (!search) return nodes
  const result = []
  for (const node of nodes) {
    const selfMatched = normalizeText([node.name, node.path].filter(Boolean).join(' ')).includes(search)
    const children = selfMatched ? node.children : filterDepartmentTree(node.children, keyword)
    if (selfMatched || children.length > 0) {
      result.push({ ...node, children })
    }
  }
  return result
}

export function flattenDepartmentSelectionTree(nodes = [], expandedKeys, keyword = '', depth = 1, parentPath = '') {
  const rows = []
  const forceExpand = Boolean(normalizeText(keyword))
  for (const node of nodes) {
    const path = [parentPath, node.name].filter(Boolean).join(' / ')
    const expandable = node.children.length > 0
    const expanded = forceExpand || expandedKeys.has(node.key)
    rows.push({
      type: 'department',
      key: node.key,
      depth,
      name: node.name,
      path,
      count: node.count,
      expandable,
      expanded
    })
    if (expanded) {
      rows.push(...flattenDepartmentSelectionTree(node.children, expandedKeys, keyword, depth + 1, path))
    }
  }
  return rows
}

export function eventValue(event) {
  return event?.detail?.value ?? event?.target?.value ?? ''
}

export function normalizeText(value) {
  return String(value || '').trim().toLowerCase()
}

export function collectDepartmentPaths(nodes = [], parentPath = '') {
  const paths = []
  for (const node of nodes) {
    const path = [parentPath, node.name].filter(Boolean).join(' / ')
    paths.push(path)
    paths.push(...collectDepartmentPaths(node.children, path))
  }
  return paths
}

export function collectDepartmentKeys(nodes = []) {
  const keys = []
  for (const node of nodes) {
    if (node.children.length > 0 || node.users.length > 0) {
      keys.push(node.key)
    }
    keys.push(...collectDepartmentKeys(node.children))
  }
  return keys
}

export function sortUsers(users = []) {
  return [...users].sort((left, right) => {
    const a = [left.name, left.id].filter(Boolean).join('\x00')
    const b = [right.name, right.id].filter(Boolean).join('\x00')
    return a.localeCompare(b, 'zh-Hans-CN')
  })
}

export function realDepartmentLevels(user) {
  const parts = String(user.department || '').split('/').map((item) => item.trim()).filter(Boolean)
  const levels = [
    firstText(user.departmentLevel1, parts[0]),
    firstText(user.departmentLevel2, parts[1]),
    firstText(user.departmentLevel3, parts[2])
  ].filter(Boolean)
  return levels.length > 0 ? levels : ['未设置部门']
}

export function departmentText(user) {
  return [user.departmentLevel1, user.departmentLevel2, user.departmentLevel3].filter(Boolean).join(' / ') || user.department || '未设置部门'
}

export function userOptionText(user) {
  return [user.name || user.id, user.position, departmentText(user)].filter(Boolean).join(' · ')
}

export function initials(value) {
  return String(value || '人').slice(0, 1).toUpperCase()
}
