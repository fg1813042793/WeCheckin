<script>
import { computed, h, ref, Teleport } from 'vue'
import { usePerformanceContext } from '../common/context'
import { HistoryReviewDetailModal } from '../common/components/ReviewDetailModal'
import { ReviewForm } from '../common/components/ReviewForm'
import { renderTablePagination, useTablePagination } from '../common/tablePagination'
import {
  renderManagerReviewTable,
  renderManagerReviewedFilters,
  reviewMatchesManagerReviewedFilters
} from '../common/tables'
import { renderPerformanceLoadingState, renderWorkbenchHead } from '../common/workbenchPage'

export default {
  name: 'ManagerReviewPage',
  setup() {
    const ctx = usePerformanceContext()
    const managerDetailReview = ref(null)
    const managerReviewedFiltersCollapsed = ref(true)
    const managerReviewedFilterCount = computed(() => [
      ctx.managerReviewedFilters.employeeName,
      ctx.managerReviewedFilters.period,
      ctx.managerReviewedFilters.objectiveScore
    ].filter(Boolean).length)
    const managerTableRows = computed(() => {
      const pendingRows = ctx.currentReviews.value.filter((review) => review.status === 'manager_review')
      const reviewedRows = ctx.currentReviews.value
        .filter((review) => ['hrbp_review', 'employee_confirm', 'hr_final', 'completed'].includes(review.status))
        .filter((review) => reviewMatchesManagerReviewedFilters(ctx, review))
      return ctx.managerReviewTab.value === 'pending' ? pendingRows : reviewedRows
    })
    const managerPagination = useTablePagination(managerTableRows)

    function openManagerReviewRow(review) {
      if (ctx.managerReviewTab.value === 'pending') {
        managerDetailReview.value = null
        ctx.selectReview(review.id)
        ctx.reviewTab.value = 'manager'
        return
      }
      ctx.selectedReviewId.value = ''
      ctx.reviewTab.value = 'currentTargets'
      managerDetailReview.value = review
    }

    function toggleManagerReviewedFilters() {
      managerReviewedFiltersCollapsed.value = !managerReviewedFiltersCollapsed.value
    }

    return () => {
      const head = renderWorkbenchHead(ctx)
      if (ctx.contentLoading.value) {
        return h('view', { class: 'workbench workbench-loading' }, [
          head,
          renderPerformanceLoadingState()
        ])
      }
      const selectedManagerReview = ctx.selectedReviewId.value ? ctx.selectedReview.value : null
      if (selectedManagerReview) {
        return h('view', { class: 'workbench manager-review-detail-page' }, [
          head,
          h('view', { class: 'detail-page-toolbar' }, [
            h('button', { class: 'dt-btn dt-btn-light', onClick: () => { ctx.selectedReviewId.value = '' } }, '返回列表')
          ]),
          h('section', { class: 'panel detail-panel manager-review-detail-panel' }, [
            h(ReviewForm, { review: selectedManagerReview })
          ])
        ])
      }
      const rows = managerPagination.rows.value
      return h('view', { class: 'workbench hrbp-review-page manager-review-page' }, [
        head,
        h('view', { class: 'hrbp-review-tabs-bar' }, [
          h('view', { class: 'hrbp-review-tabs' }, [
            h('button', { class: ['hrbp-review-tab', ctx.managerReviewTab.value === 'pending' ? 'active' : ''], onClick: () => ctx.switchManagerReviewTab('pending') }, '待评'),
            h('button', { class: ['hrbp-review-tab', ctx.managerReviewTab.value === 'reviewed' ? 'active' : ''], onClick: () => ctx.switchManagerReviewTab('reviewed') }, '已评')
          ])
        ]),
        h('view', { class: 'table-panel-stack manager-table-panel-stack' }, [
          h('section', { class: 'panel table-panel' }, [
            renderManagerReviewedFilters(ctx, managerReviewedFiltersCollapsed.value, managerReviewedFilterCount.value, toggleManagerReviewedFilters),
            renderManagerReviewTable(ctx, rows, openManagerReviewRow)
          ]),
          renderTablePagination(managerPagination)
        ]),
        managerDetailReview.value
          ? h(Teleport, { to: 'body' }, [
              h(HistoryReviewDetailModal, {
                review: managerDetailReview.value,
                onClose: () => { managerDetailReview.value = null }
              })
            ])
          : null
      ])
    }
  }
}
</script>
