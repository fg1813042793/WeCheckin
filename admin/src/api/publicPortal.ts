import type {
  PublicActivities, PublicActivityItem, PublicPortalConfig, PublicPortalSettings,
  PublicPortalItemInput, PublicPortalItemUpdate, PublicPortalOptions, PublicPortalOptionsQuery, PublicPortalAdminItem, PublicPortalDeleteResponse,
} from '@/types/publicPortal'
import { publicActivityUrl, settingsWritePayload } from '@/views/public-portal/portalModel'

const ADMIN_PATH = '/api/v2/admin/public-portal'
const jsonBody = {
  headers: { 'Content-Type': 'application/json' },
  transformRequest: [(body: unknown) => JSON.stringify(body)],
}

// Lazy import keeps the anonymous entry independent of admin interceptors/session handling.
export const publicPortalApi = {
  async get() {
    const { default: request } = await import('@/utils/request')
    return request.get<PublicPortalConfig>(ADMIN_PATH)
  },
  async saveSettings(data: PublicPortalSettings) {
    const { default: request } = await import('@/utils/request')
    return request.put<PublicPortalSettings, PublicPortalSettings>(`${ADMIN_PATH}/settings`, settingsWritePayload(data), jsonBody)
  },
  async options(params: PublicPortalOptionsQuery, signal?: AbortSignal) {
    const { default: request } = await import('@/utils/request')
    return request.get<PublicPortalOptions>(`${ADMIN_PATH}/options`, { params, signal })
  },
  async create(data: PublicPortalItemInput) {
    const { default: request } = await import('@/utils/request')
    return request.post<PublicPortalAdminItem, PublicPortalItemInput>(`${ADMIN_PATH}/items`, data, jsonBody)
  },
  async update(id: number, data: PublicPortalItemUpdate) {
    const { default: request } = await import('@/utils/request')
    return request.put<PublicPortalAdminItem, PublicPortalItemUpdate>(`${ADMIN_PATH}/items/${id}`, data, jsonBody)
  },
  async remove(id: number, version: number) {
    const { default: request } = await import('@/utils/request')
    return request.delete<PublicPortalDeleteResponse>(`${ADMIN_PATH}/items/${id}`, { params: { version } })
  },
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null
}

function isPublicItem(value: unknown): value is PublicActivityItem {
  if (!isRecord(value)) return false
  return ['id', 'sourceId', 'startTime', 'endTime'].every(key => typeof value[key] === 'number' && Number.isFinite(value[key]))
    && ['title', 'description', 'url'].every(key => typeof value[key] === 'string')
    && (value.managementTitle === undefined || typeof value.managementTitle === 'string')
    && (value.category === undefined || typeof value.category === 'string')
    && (value.kind === 'survey' || value.kind === 'exam')
    && ['upcoming', 'active', 'ended'].includes(String(value.state))
    && typeof value.loginRequired === 'boolean' && typeof value.passwordRequired === 'boolean'
}

function isPublicActivities(value: unknown): value is PublicActivities {
  return isRecord(value) && typeof value.enabled === 'boolean'
    && ['title', 'description', 'logoUrl'].every(key => typeof value[key] === 'string')
    && Array.isArray(value.items) && value.items.every(isPublicItem)
}

async function fetchPortalData(path: string, token: string, signal: AbortSignal | undefined, apiBase: string): Promise<unknown> {
  const headers = new Headers({ Accept: 'application/json' })
  if (token) headers.set('Authorization', token)
  const response = await fetch(`${apiBase.replace(/\/+$/, '')}${path}`, {
    method: 'GET', credentials: 'omit', cache: 'no-store', headers, signal,
  })
  let payload: unknown
  try { payload = await response.json() } catch { throw new Error('活动接口响应异常，请稍后重试') }
  if (!isRecord(payload) || typeof payload.code !== 'number' || typeof payload.msg !== 'string') {
    throw new Error('活动接口响应格式错误，请稍后重试')
  }
  if (!response.ok || payload.code !== 0) {
    throw new Error(payload.msg || (response.status === 401 ? '请先管理员登录后再预览' : '活动加载失败，请重试'))
  }
  return payload.data
}

async function fetchActivities(path: string, token: string, signal: AbortSignal | undefined, apiBase: string): Promise<PublicActivities> {
  const data = await fetchPortalData(path, token, signal, apiBase)
  if (!isPublicActivities(data)) throw new Error('活动接口响应格式错误，请稍后重试')
  return { ...data, items: data.items.map(item => ({ ...item, url: publicActivityUrl(item.url) })) }
}

export function fetchPublicActivities(signal?: AbortSignal, apiBase = import.meta.env.VITE_API_BASE || ''): Promise<PublicActivities> {
  return fetchActivities('/api/v2/public/activities', '', signal, apiBase)
}

export async function fetchPublicPortalPreview(token: string, signal?: AbortSignal, apiBase = import.meta.env.VITE_API_BASE || ''): Promise<PublicActivities> {
  if (!token) throw new Error('预览需要管理员登录，请登录管理后台后重试')
  return fetchActivities(`${ADMIN_PATH}/preview`, token, signal, apiBase)
}
