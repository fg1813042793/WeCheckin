<script>
import { computed, h, ref, Teleport } from 'vue'
import { usePerformanceContext } from '../common/context'
import { HistoryReviewDetailModal } from '../common/components/ReviewDetailModal'
import { renderTablePagination, useTablePagination } from '../common/tablePagination'
import { renderHistoryFilters, renderHistoryTable } from '../common/tables'
import { renderPerformanceLoadingState, renderWorkbenchHead } from '../common/workbenchPage'

export default {
  name: 'HistoryPage',
  setup() {
    const ctx = usePerformanceContext()
    const historyDetailReview = ref(null)
    const historyFiltersCollapsed = ref(true)
    const historyFilterCount = computed(() => [
      ctx.historyFilters.year,
      ctx.historyFilters.month,
      ctx.historyFilters.grade
    ].filter(Boolean).length)
    const historyTableRows = computed(() => ctx.currentReviews.value
      .slice()
      .sort((a, b) => String(b.period || '').localeCompare(String(a.period || '')))
    )
    const historyPagination = useTablePagination(historyTableRows)

    function openHistoryDetail(review) {
      ctx.reviewTab.value = 'currentTargets'
      historyDetailReview.value = review
    }

    function toggleHistoryFilters() {
      historyFiltersCollapsed.value = !historyFiltersCollapsed.value
    }

    return () => {
      const head = renderWorkbenchHead(ctx)
      if (ctx.contentLoading.value) {
        return h('view', { class: 'workbench workbench-loading' }, [
          head,
          renderPerformanceLoadingState()
        ])
      }
      const historyRows = historyPagination.rows.value
      return h('view', { class: 'workbench history-performance-page' }, [
        head,
        h('view', { class: 'table-panel-stack history-table-panel-stack' }, [
          h('section', { class: 'panel table-panel' }, [
            renderHistoryFilters(ctx, historyFiltersCollapsed.value, historyFilterCount.value, toggleHistoryFilters),
            renderHistoryTable(ctx, historyRows, openHistoryDetail)
          ]),
          renderTablePagination(historyPagination)
        ]),
        historyDetailReview.value
          ? h(Teleport, { to: 'body' }, [
              h(HistoryReviewDetailModal, {
                review: historyDetailReview.value,
                onClose: () => { historyDetailReview.value = null }
              })
            ])
          : null
      ])
    }
  }
}
</script>
