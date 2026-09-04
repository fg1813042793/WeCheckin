import type { Component } from 'vue'
import FlowPage from './flow.vue'
import HistoryPage from './history.vue'
import HrbpPage from './hrbp.vue'
import MinePage from './mine.vue'
import ParametersPage from './parameters.vue'
import ReviewPage from './review.vue'
import SummaryPage from './summary.vue'

export const performanceContentRoutes: Record<string, Component> = {
  'performance:mine': MinePage,
  'performance:history': HistoryPage,
  'performance:manager': ReviewPage,
  'performance:hrbp': HrbpPage,
  'performance:summary': SummaryPage,
  'performance:org': FlowPage,
  'performance:template': ParametersPage,
}
