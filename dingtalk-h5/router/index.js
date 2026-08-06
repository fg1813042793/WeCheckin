import { menuPageKeys } from '../views/performance/common/constants'
import { reviewFormTabKeys } from '../views/performance/common/reviewTabs'
import DashboardPage from '../views/workbench/dashboard/index.vue'
import FlowConfigPage from '../views/performance/flow-config/index.vue'
import HrbpReviewPage from '../views/performance/hrbp-review/index.vue'
import HistoryPage from '../views/performance/history/index.vue'
import ManagerReviewPage from '../views/performance/manager-review/index.vue'
import MinePage from '../views/performance/mine/index.vue'
import SummaryPage from '../views/performance/summary/index.vue'
import TemplatePage from '../views/performance/parameters/index.vue'

const menuPageComponents = {
  dashboard: DashboardPage,
  mine: MinePage,
  history: HistoryPage,
  manager: ManagerReviewPage,
  hrbp: HrbpReviewPage,
  summary: SummaryPage,
  org: FlowConfigPage,
  template: TemplatePage
}

export function resolveMenuPageComponent(contentView) {
  return menuPageComponents[contentView] || DashboardPage
}

export function normalizeReviewDeepLinkView(value) {
  const raw = String(value || '').trim()
  if (!raw) return ''
  if (menuPageKeys.has(raw)) return raw
  const performanceKey = `performance:${raw}`
  return menuPageKeys.has(performanceKey) ? performanceKey : ''
}

export function normalizeReviewDeepLinkTab(value) {
  const raw = String(value || '').trim()
  return reviewFormTabKeys.has(raw) ? raw : ''
}
