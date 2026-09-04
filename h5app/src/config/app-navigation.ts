import type { DingTalkMenu } from '@/types/dingtalk-h5'
import type { AppNavItem } from '@/types/navigation'
import { performanceMenuPages, performanceRootNavItem } from '@/pages/performance/performance.menu'
import { workflowMenuPages, workflowRootNavItem } from '@/pages/workflow/workflow.menu'
import { resolveAppMenuIcon } from './app-icons'

export type AppContentView
  = | 'dashboard'
    | 'mine'
    | 'history'
    | 'manager'
    | 'hrbp'
    | 'summary'
    | 'org'
    | 'template'
    | 'workflow'

export interface RegisteredAppPage {
  key: string
  contentKey: AppContentView
  route: string
  title: string
  description: string
  icon: string
  parentKey?: string
  rootKey: string
}

interface AppRootNavItem {
  key: string
  label: string
  icon: string
}

const dashboardPage: RegisteredAppPage = {
  key: 'dashboard',
  contentKey: 'dashboard',
  route: '/pages/index/index',
  title: '工作台',
  description: '查看当前绩效待办和统计信息',
  icon: 'home',
  rootKey: 'dashboard',
}

const appRootNavItems: AppRootNavItem[] = [
  {
    key: 'dashboard',
    label: dashboardPage.title,
    icon: dashboardPage.icon,
  },
  performanceRootNavItem,
  workflowRootNavItem,
]

export const appRegisteredPages: RegisteredAppPage[] = [
  dashboardPage,
  ...performanceMenuPages,
  ...workflowMenuPages,
]

const registeredPageMap = new Map(appRegisteredPages.map(page => [page.key, page]))
const contentPageMap = new Map(appRegisteredPages.map(page => [page.contentKey, page]))
const rootNavMap = new Map(appRootNavItems.map(item => [item.key, item]))
const rootMenuKeys = new Set(appRootNavItems.map(item => item.key))
const registeredMenuKeys = new Set([
  ...appRootNavItems.map(item => item.key),
  ...appRegisteredPages.map(page => page.key),
])

function firstText(...values: Array<unknown>) {
  for (const value of values) {
    const text = String(value || '').trim()
    if (text) {
      return text
    }
  }
  return ''
}

function safeDecodeURIComponent(value: string) {
  try {
    return decodeURIComponent(value)
  }
  catch {
    return value
  }
}

function normalizeRoutePath(route: string) {
  const rawRoute = String(route || '').trim()
  const routeWithoutHash = rawRoute.split('#')[0] || rawRoute
  const routePath = routeWithoutHash.split('?')[0] || routeWithoutHash
  return `/${routePath.replace(/^\/+/, '')}`.replace(/\.vue$/, '')
}

function routeQueryValue(route: string, key: string) {
  const rawRoute = String(route || '')
  const queryIndex = rawRoute.indexOf('?')
  if (queryIndex < 0) {
    return ''
  }
  const queryText = rawRoute.slice(queryIndex + 1).split('#')[0] || ''
  for (const segment of queryText.split('&')) {
    const [rawKey, rawValue = ''] = segment.split('=')
    if (safeDecodeURIComponent(rawKey || '') === key) {
      return safeDecodeURIComponent(rawValue)
    }
  }
  return ''
}

function menuLabel(item: DingTalkMenu) {
  return firstText(item.label, item.name, item.title, item.menuName)
    || registeredPageMap.get(item.key)?.title
    || rootNavMap.get(item.key)?.label
    || item.key
}

function menuIcon(item: DingTalkMenu) {
  return resolveAppMenuIcon(
    item.icon,
    registeredPageMap.get(item.key)?.icon || rootNavMap.get(item.key)?.icon || 'grid',
  )
}

function menuPath(key: string) {
  return registeredPageMap.get(key)?.route
}

function registeredChildrenFor(parentKey: string) {
  return appRegisteredPages.filter(page => page.parentKey === parentKey)
}

function isAllowedMenuAtDepth(key: string, parentKey = '') {
  if (!registeredMenuKeys.has(key)) {
    return false
  }
  if (!parentKey) {
    return rootMenuKeys.has(key) || registeredPageMap.has(key)
  }
  if (rootMenuKeys.has(key)) {
    return false
  }
  return registeredPageMap.get(key)?.parentKey === parentKey
}

export function normalizeAppMenus(menus: DingTalkMenu[], parentKey = ''): AppNavItem[] {
  return menus
    .filter(item => item && isAllowedMenuAtDepth(item.key, parentKey))
    .map((item) => {
      const registeredChildren = registeredChildrenFor(item.key)
      const children = normalizeAppMenus(item.children || [], item.key)
      const childKeys = new Set(children.map(child => child.key))
      const mergedChildren = [
        ...children,
        ...registeredChildren
          .filter(page => !childKeys.has(page.key) && menus.some(menu => menu.key === page.key))
          .map(page => ({
            key: page.key,
            label: page.title,
            icon: page.icon,
            path: page.route,
            children: [],
          })),
      ]

      return {
        key: item.key,
        label: menuLabel(item),
        icon: menuIcon(item),
        path: menuPath(item.key),
        children: mergedChildren,
      }
    })
}

export function flattenAppNav(items: AppNavItem[]): AppNavItem[] {
  return items.flatMap(item => [item, ...flattenAppNav(item.children || [])])
}

export function createAppNavItems(items: AppNavItem[]) {
  return appRootNavItems
    .map((root) => {
      const existingRoot = items.find(item => item.key === root.key)
      const rootPage = registeredPageMap.get(root.key)
      const fallbackChildren = items.filter(item => registeredPageMap.get(item.key)?.rootKey === root.key && item.key !== root.key)
      const children = existingRoot?.children.length ? existingRoot.children : fallbackChildren

      return {
        allowed: children.length > 0 || Boolean(rootPage && existingRoot),
        key: root.key,
        label: existingRoot?.label || root.label,
        icon: resolveAppMenuIcon(existingRoot?.icon, root.icon),
        path: existingRoot?.path || menuPath(root.key),
        children,
      }
    })
    .filter(item => item.allowed)
    .map(({ allowed, ...item }) => item)
}

export function resolveAppPage(value: string) {
  return registeredPageMap.get(value) || contentPageMap.get(value as AppContentView) || appRegisteredPages[0]
}

export function resolveAppContentView(value: string): AppContentView {
  return resolveAppPage(value).contentKey
}

export function appPageTitle(value: string) {
  return resolveAppPage(value).title
}

export function appPageRoute(value: string) {
  return resolveAppPage(value).route
}

export function resolveAppPageByRoute(route: string) {
  const routeView = routeQueryValue(route, 'view')
  if (routeView) {
    return resolveAppPage(routeView)
  }
  const normalizedRoute = normalizeRoutePath(route)
  return appRegisteredPages.find(page => normalizeRoutePath(page.route) === normalizedRoute) || appRegisteredPages[0]
}

export function isRegisteredAppMenuKey(value: string) {
  return registeredMenuKeys.has(value)
}

export function useRegisteredAppPages() {
  return appRegisteredPages
}
