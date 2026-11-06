import { inject, onScopeDispose } from 'vue'
import type { InjectionKey } from 'vue'
import { useRoute } from 'vue-router'
import type { RouteLocationNormalizedLoaded } from 'vue-router'

export interface VisitedTab {
  title: string
  path: string
  fullPath: string
  name: string
  customTitle?: string
}

type TabRoute = Pick<RouteLocationNormalizedLoaded, 'path' | 'fullPath' | 'meta'> & { name?: RouteLocationNormalizedLoaded['name'] | null }
type UpdateTabTitle = (fullPath: string, title: string) => void
export const ADMIN_ROUTE_TAB_TITLE_KEY: InjectionKey<UpdateTabTitle> = Symbol('admin-route-tab-title')
export const ADMIN_ROUTE_TAB_REPLACE_KEY: InjectionKey<(fromFullPath: string, target: TabRoute, title: string) => void> = Symbol('admin-route-tab-replace')

function defaultTabTitle(path: string, title: string, fullPath: string): string {
  const workflowID = path.match(/^\/workflow\/definitions\/(\d+)\/designer$/)?.[1]
  if (workflowID) return `流程设计器 #${workflowID}`
  if (path === '/survey/designer' || path === '/exam/designer') {
    const kind = path === '/survey/designer' ? '问卷' : '考试'
    const id = fullPath.match(/[?&]id=([1-9]\d*)(?:&|#|$)/)?.[1]
    return id ? `${kind}设计器 #${id}` : `新建${kind}`
  }
  return title
}

export function formDesignerTabTitle(kind: 'survey' | 'exam', id: number, name?: string): string {
  const label = kind === 'survey' ? '问卷' : '考试'
  const title = typeof name === 'string' ? name.trim() : ''
  if (!id) return `新建${label}`
  return title ? `${title} · ${label}#${id}` : `${label}设计器 #${id}`
}

export function workflowDesignerTabTitle(id: number, name?: string): string {
  const title = typeof name === 'string' ? name.trim() : ''
  return title ? `${title} · #${id}` : `流程设计器 #${id}`
}

export function normalizeVisitedTabs(tabs: VisitedTab[]): VisitedTab[] {
  const seen = new Set<string>()
  return tabs.filter(tab => {
    if (seen.has(tab.fullPath)) return false
    seen.add(tab.fullPath)
    return true
  }).map(tab => ({ ...tab, title: tab.customTitle || defaultTabTitle(tab.path, tab.title, tab.fullPath) }))
}

export function routeToVisitedTab(target: TabRoute, previous?: VisitedTab): VisitedTab | null {
  if (target.path === '/login' || target.path === '/forbidden') return null
  const title = String(target.meta?.title || '')
  if (!title) return null
  const customTitle = previous?.fullPath === target.fullPath ? previous.customTitle : undefined
  return {
    title: customTitle || defaultTabTitle(target.path, title, target.fullPath),
    path: target.path,
    fullPath: target.fullPath,
    name: String(target.name || target.path),
    ...(customTitle ? { customTitle } : {}),
  }
}

export function updateVisitedTabTitle(tabs: VisitedTab[], fullPath: string, title: string): boolean {
  const tab = tabs.find(item => item.fullPath === fullPath)
  const value = title.trim()
  if (!tab || !value || (tab.title === value && tab.customTitle === value)) return false
  tab.title = value
  tab.customTitle = value
  return true
}

export function replaceVisitedTab(tabs: VisitedTab[], fromFullPath: string, target: TabRoute, title: string): boolean {
  const index = tabs.findIndex(tab => tab.fullPath === fromFullPath)
  const next = routeToVisitedTab(target)
  if (index < 0 || !next || !title.trim()) return false
  next.title = title.trim()
  next.customTitle = next.title
  const updated = tabs.flatMap((tab, position) => position === index ? [next] : tab.fullPath === next.fullPath ? [] : [tab])
  tabs.splice(0, tabs.length, ...updated)
  return true
}

export function useAdminRouteTabTitle(): (title: string) => void {
  const update = inject(ADMIN_ROUTE_TAB_TITLE_KEY, undefined)
  // useRoute is shared/reactive even in deactivated KeepAlive pages. Capture the
  // page's own identity so a delayed response cannot rename another open tab.
  const fullPath = useRoute().fullPath
  let disposed = false
  onScopeDispose(() => { disposed = true })
  return title => {
    if (!disposed) update?.(fullPath, title)
  }
}
