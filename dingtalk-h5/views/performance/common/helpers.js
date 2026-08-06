import { menuPageKeys } from './constants'

export function menuContentKey(key) {
  if (String(key || '').startsWith('performance:')) {
    return String(key).replace('performance:', '')
  }
  return key
}

export function menuIcon(item = {}) {
  const key = menuContentKey(item.key)
  return item.icon || key || item.key
}

export function menuLabel(item = {}) {
  return firstText(item.label, item.name, item.title, item.menuName) || titleForContent(menuContentKey(item.key))
}

export function firstText(...values) {
  for (const value of values) {
    const text = String(value || '').trim()
    if (text) return text
  }
  return ''
}

export function defaultAppConfig() {
  return {
    appTitle: '钉钉H5应用',
    appName: '钉钉H5应用',
    logoText: 'H5',
    logoUrl: '',
    appUrl: ''
  }
}

export function normalizeAppConfig(config = {}) {
  const fallback = defaultAppConfig()
  const appName = firstText(config.appName, config.appTitle, fallback.appName)
  const appTitle = firstText(config.appTitle, appName, fallback.appTitle)
  const logoText = firstText(config.logoText, fallback.logoText).slice(0, 4)
  const logoUrl = firstText(config.logoUrl, config.logoURL)
  const appUrl = firstText(config.appUrl)
  return {
    appTitle,
    appName,
    logoText,
    logoUrl,
    appUrl
  }
}

export function titleForContent(view) {
  const titles = {
    dashboard: '工作台',
    performance: '绩效管理',
    mine: '我的绩效',
    history: '历史绩效',
    manager: '上级评价',
    hrbp: 'HRBP评价',
    summary: 'HRBP汇总',
    org: '流程执行',
    template: '绩效模版'
  }
  return titles[view] || '工作台'
}

export function normalizeMenuTree(menus = [], isRoot = true) {
  if (!Array.isArray(menus)) return []
  const items = menus
    .filter((item) => item && menuPageKeys.has(item.key))
    .map((item) => ({
      ...item,
      label: menuLabel(item),
      children: normalizeMenuTree(item.children || [], false)
    }))
  const performanceMenu = items.find((item) => item.key === 'performance')
  if (performanceMenu && performanceMenu.children.length === 0) {
    performanceMenu.children = items
      .filter((item) => String(item.key || '').startsWith('performance:'))
      .map((item) => ({ ...item, children: [] }))
  }
  return isRoot ? items.filter((item) => !String(item.key || '').startsWith('performance:')) : items
}

export function flattenMenuTree(menus = []) {
  return menus.flatMap((item) => [item, ...flattenMenuTree(item.children || [])])
}

export function normalizeSearchKeyword(value) {
  return String(value || '').trim().toLowerCase()
}

export function createTargetUserMeta(user) {
  return [user.position, user.department].filter(Boolean).join(' · ') || user.id
}

export function createTargetUserMatchesKeyword(user, keyword) {
  const text = [
    user.id,
    user.name,
    user.account,
    user.mobile,
    user.phone,
    user.position,
    user.department,
    user.departmentLevel1,
    user.departmentLevel2,
    user.departmentLevel3,
    createTargetUserMeta(user)
  ].filter(Boolean).join(' ').toLowerCase()
  const terms = keyword.split(/\s+/).filter(Boolean)
  return terms.every((term) => text.includes(term))
}

export function buildCreateTargetUserTree(users = []) {
  const root = new Map()
  for (const user of users || []) {
    const levels = createTargetDepartmentLevels(user)
    let currentMap = root
    let currentNode = null
    for (const [index, level] of levels.entries()) {
      const parentKey = currentNode?.key || 'root'
      currentNode = ensureCreateTargetNode(currentMap, `${parentKey}/l${index + 1}:${level}`, level)
      currentNode.count += 1
      if (user.id) currentNode.userIds.push(user.id)
      currentMap = currentNode.childMap
    }
    if (currentNode) {
      currentNode.users.push(user)
    }
  }
  return finalizeCreateTargetNodes([...root.values()])
}

function ensureCreateTargetNode(map, key, name) {
  if (!map.has(key)) {
    map.set(key, { key, name, count: 0, userIds: [], childMap: new Map(), children: [], users: [] })
  }
  return map.get(key)
}

function finalizeCreateTargetNodes(nodes = []) {
  return nodes
    .sort((left, right) => left.name.localeCompare(right.name, 'zh-CN'))
    .map((node) => ({
      key: node.key,
      name: node.name,
      count: node.count,
      userIds: [...new Set(node.userIds.filter(Boolean))],
      children: finalizeCreateTargetNodes([...node.childMap.values()]),
      users: node.users.slice().sort((left, right) => userSortText(left).localeCompare(userSortText(right), 'zh-CN'))
    }))
}

export function flattenCreateTargetTree(nodes = [], expandedKeys, depth = 1, forceExpanded = false) {
  const rows = []
  for (const node of nodes) {
    const hasChildren = node.children.length > 0 || node.users.length > 0
    const expanded = forceExpanded || expandedKeys.has(node.key)
    rows.push({
      type: 'department',
      key: node.key,
      depth,
      name: node.name,
      count: node.count,
      userIds: node.userIds,
      expandable: hasChildren,
      expanded
    })
    if (!expanded) continue
    for (const user of node.users) {
      rows.push({ type: 'employee', key: `${node.key}/user:${user.id}`, depth: depth + 1, user })
    }
    rows.push(...flattenCreateTargetTree(node.children, expandedKeys, depth + 1, forceExpanded))
  }
  return rows
}

function createTargetDepartmentLevels(user) {
  const parts = String(user.department || '').split('/').map((item) => item.trim()).filter(Boolean)
  const levels = [
    firstText(user.departmentLevel1, parts[0]),
    firstText(user.departmentLevel2, parts[1]),
    firstText(user.departmentLevel3, parts[2])
  ].filter(Boolean)
  return levels.length > 0 ? levels : ['未设置部门']
}

function userSortText(user) {
  return [user.departmentLevel1, user.departmentLevel2, user.departmentLevel3, user.name, user.id].filter(Boolean).join('\x00')
}

export function unique(items) {
  return [...new Set(items.filter(Boolean))]
}

export function currentMonth() {
  const now = new Date()
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
}

export function nextMonthFromPeriod(period) {
  const [yearText, monthText] = String(period || '').split('-')
  const year = Number(yearText)
  const month = Number(monthText)
  if (!year || month < 1 || month > 12) return nextMonth()
  const date = new Date(year, month, 1)
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}`
}

export function nextMonth() {
  const now = new Date()
  now.setMonth(now.getMonth() + 1)
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
}
