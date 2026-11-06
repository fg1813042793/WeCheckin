import type { PublicActivityItem, PublicActivityState, PublicPortalAdminItem, PublicPortalItemInput, PublicPortalItemUpdate, PublicPortalKind, PublicPortalSettings } from '@/types/publicPortal'

export const PORTAL_EDIT_PERMISSION = 'admin:menu:public-portal:edit'
export const kindLabels: Record<PublicPortalKind, string> = { survey: '问卷', exam: '考试' }
export const stateLabels: Record<PublicActivityState, string> = { upcoming: '未开始', active: '进行中', ended: '已结束' }

export function activityCategory(kind: PublicPortalKind, category?: string): string {
  return category?.trim() || kindLabels[kind]
}

export function activityFilters(items: PublicActivityItem[]): Array<{ value: string; label: string }> {
  const labels = new Set(items.map(item => activityCategory(item.kind, item.category)))
  return [{ value: 'all', label: '全部' }, ...Array.from(labels, label => ({ value: `category:${label}`, label }))]
}

export function activityIdentity(kind: PublicPortalKind, sourceId: number, category?: string): string {
  return `${activityCategory(kind, category)}编号：${sourceId}`
}

export function safeTargetUrl(value: string): string {
  if (!value || /[\u0000-\u0020\u007f\\]/.test(value)) return ''
  try {
    const url = new URL(value)
    return ['http:', 'https:'].includes(url.protocol) && !url.username && !url.password ? url.href : ''
  } catch { return '' }
}

export function safeLogoUrl(value: string): string {
  if (/^\/(?!\/)/.test(value) && !/[\u0000-\u0020\u007f\\]/.test(value)) return value
  return safeTargetUrl(value)
}

function browserOrigin(): string {
  return typeof window === 'undefined' ? '' : window.location.origin
}

function safeSiteOrigin(origin: string): string {
  const safe = safeTargetUrl(origin)
  if (!safe || !/^https?:\/\/[^/?#@]+\/?$/i.test(origin)) return ''
  const url = new URL(safe)
  return url.pathname === '/' ? url.origin : ''
}

export function publicActivitiesUrl(origin = browserOrigin()): string {
  const site = safeSiteOrigin(origin)
  return site ? `${site}/public/activities` : ''
}

export function publicActivityUrl(value: string, origin = browserOrigin()): string {
  const site = safeSiteOrigin(origin)
  if (!site) return ''
  // Preserve only a literal filling route from legacy HTTP(S) links. Never trust
  // their host or URL normalization of credentials, dot segments or escapes.
  let path = value
  if (!value.startsWith('/')) {
    if (!safeTargetUrl(value)) return ''
    path = value.match(/^https?:\/\/[^/?#@]+(\/.*)$/i)?.[1] || ''
  }
  if (!/^\/(?:sf1?|ef1?)\/[1-9]\d*$/.test(path)) return ''
  return new URL(path, site).href
}

export function isPreviewQuery(value: unknown): boolean { return value === '1' }

export function canEditItem(allowed: boolean, item: PublicPortalAdminItem): boolean { return allowed && item.canManage && item.unavailableReason !== '来源已删除' }

export function canRemoveItem(allowed: boolean, item: PublicPortalAdminItem): boolean { return allowed && item.canManage }

export function canToggleItem(allowed: boolean, item: PublicPortalAdminItem): boolean { return canEditItem(allowed, item) && (item.enabled || item.available) }

export function itemWritePayload(item: PublicPortalAdminItem): PublicPortalItemUpdate {
  return { kind: item.kind, sourceId: item.sourceId, title: item.title.trim(), description: item.description.trim(), enabled: item.enabled, sort: item.sort, version: item.version }
}

export function settingsWritePayload(settings: PublicPortalSettings): PublicPortalSettings {
  return {
    enabled: settings.enabled, title: settings.title.trim(), description: settings.description.trim(),
    logoUrl: settings.logoUrl.trim(), showEnded: settings.showEnded, version: settings.version,
  }
}

export function settingsError(settings: PublicPortalSettings): string {
  if (!settings.title.trim()) return '请输入页面标题'
  if (Array.from(settings.title.trim()).length > 100) return '页面标题不能超过 100 字'
  if (Array.from(settings.description).length > 1000) return '页面简介不能超过 1000 字'
  if (settings.logoUrl && !safeLogoUrl(settings.logoUrl.trim())) return 'Logo 必须是站内绝对路径或 HTTP(S) 地址'
  return ''
}

export function itemInputError(item: PublicPortalItemInput): string {
  if (!Number.isSafeInteger(item.sourceId) || item.sourceId <= 0) return '请选择已发布、公开且可管理的问卷或考试'
  if (Array.from(item.title.trim()).length > 255) return '展示标题不能超过 255 字'
  if (Array.from(item.description).length > 1000) return '展示简介不能超过 1000 字'
  if (!Number.isInteger(item.sort) || item.sort < 0 || item.sort > 99999) return '排序必须为 0 至 99999 的整数'
  return ''
}

export function canOpenActivity(item: PublicActivityItem): boolean { return item.state === 'active' && Boolean(publicActivityUrl(item.url)) }

export function filterActivities(items: PublicActivityItem[], category: string, keyword: string): PublicActivityItem[] {
  const term = keyword.trim().toLocaleLowerCase()
  return items.filter(item => (category === 'all' || `category:${activityCategory(item.kind, item.category)}` === category)
    && `${item.title}\n${item.managementTitle || ''}\n${item.description}\n${activityIdentity(item.kind, item.sourceId, item.category)}\n${activityIdentity(item.kind, item.sourceId)}\n#${item.sourceId}`.toLocaleLowerCase().includes(term))
}

export function formatActivityTime(milliseconds: number): string {
  if (!milliseconds) return '不限'
  const date = new Date(milliseconds)
  if (!Number.isFinite(date.getTime())) return '时间不可用'
  return new Intl.DateTimeFormat('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', hour12: false }).format(date)
}

export function errorText(error: unknown, fallback = '操作失败，请重试'): string {
  if (error instanceof Error) return error.message || fallback
  if (error && typeof error === 'object' && 'msg' in error && typeof error.msg === 'string') return error.msg || fallback
  return fallback
}

export function safeQrFilename(title: string, kind: PublicPortalKind, sourceId: number): string {
  const name = Array.from(title.replace(/[<>:"/\\|?*\u0000-\u001f\u007f]/g, '_').replace(/^\.+|[. ]+$/g, '').trim()).slice(0, 70).join('') || '活动'
  return `${kind}-${sourceId}-${name}.png`
}

export async function copyPortalLink(url: string): Promise<void> {
  if (!safeTargetUrl(url)) throw new Error('链接无效，无法复制')
  if (!navigator.clipboard?.writeText) throw new Error('当前浏览器不支持复制，请手动复制链接')
  try { await navigator.clipboard.writeText(url) } catch { throw new Error('复制失败，请检查剪贴板权限或手动复制链接') }
}
