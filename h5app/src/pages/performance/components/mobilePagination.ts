import { reactive } from 'vue'

export const MOBILE_REVIEW_PAGE_SIZE = 20
export const PC_REVIEW_PAGE_SIZE = 100

export interface MobilePaginationState {
  page: number
  pageSize: number
  total: number
  loadingMore: boolean
}

export function createMobilePaginationState(pageSize = MOBILE_REVIEW_PAGE_SIZE) {
  return reactive<MobilePaginationState>({
    page: 1,
    pageSize,
    total: 0,
    loadingMore: false,
  })
}

export function resetMobilePagination(state: MobilePaginationState) {
  state.page = 1
  state.total = 0
  state.loadingMore = false
}

export function mobileReviewListParams(enabled: boolean, state: MobilePaginationState) {
  return {
    page: enabled ? state.page : 1,
    pageSize: enabled ? state.pageSize : PC_REVIEW_PAGE_SIZE,
  }
}

export function updateMobilePaginationTotal(state: MobilePaginationState, total: unknown, fallback = 0) {
  const nextTotal = Number(total)
  state.total = Number.isFinite(nextTotal) ? nextTotal : fallback
}

export function showMobileLoadMore(state: MobilePaginationState, loadedCount: number) {
  return loadedCount > 0 && state.total > loadedCount
}

export function showMobileNoMore(state: MobilePaginationState, loadedCount: number) {
  return loadedCount > 0 && state.total > state.pageSize && state.total <= loadedCount
}

export function appendUniqueReviewRows<T extends { id?: unknown, reviewNo?: unknown, employeeId?: unknown, period?: unknown }>(
  current: T[],
  next: T[],
) {
  const seen = new Set(current.map((item, index) => reviewRowKey(item, index)))
  const merged = current.slice()
  for (const item of next) {
    const key = reviewRowKey(item, merged.length)
    if (!seen.has(key)) {
      seen.add(key)
      merged.push(item)
    }
  }
  return merged
}

function reviewRowKey<T extends { id?: unknown, reviewNo?: unknown, employeeId?: unknown, period?: unknown }>(item: T, index: number) {
  return String(item.id || item.reviewNo || `${item.employeeId || 'employee'}-${item.period || 'period'}-${index}`)
}
