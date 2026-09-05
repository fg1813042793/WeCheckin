import type { Component } from 'vue'
import NotificationHistoryPage from '@/pages/notifications/components/NotificationHistoryPage.vue'
import { NOTIFICATION_HISTORY_CONTENT_KEY } from '@/pages/notifications/notification-route-keys'
import PerformanceWorkbench from '@/pages/performance/components/PerformanceWorkbench.vue'
import { performanceContentRoutes } from '@/pages/performance/performance.routes'
import { resolveWorkflowContentComponent, workflowContentRoutes } from '@/pages/workflow/workflow.routes'

export interface AppContentRoute {
  component: Component
}

export const appContentRoutes: Record<string, AppContentRoute> = {
  dashboard: {
    component: PerformanceWorkbench,
  },
  [NOTIFICATION_HISTORY_CONTENT_KEY]: {
    component: NotificationHistoryPage,
  },
  ...Object.fromEntries(
    Object.entries(performanceContentRoutes).map(([key, component]) => [key, { component }]),
  ),
  ...Object.fromEntries(
    Object.entries(workflowContentRoutes).map(([key, component]) => [key, { component }]),
  ),
}

export function resolveAppContentComponent(key: string) {
  return appContentRoutes[key]?.component || resolveWorkflowContentComponent(key) || appContentRoutes.dashboard.component
}
