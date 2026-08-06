<script>
import { computed, h, ref, Teleport } from 'vue'
import { usePerformanceContext } from '../common/context'
import { HistoryReviewDetailModal } from '../common/components/ReviewDetailModal'
import { ReviewForm } from '../common/components/ReviewForm'
import { renderTablePagination, useTablePagination } from '../common/tablePagination'
import {
  renderHrbpReviewTable,
  renderHrbpReviewedFilters,
  reviewMatchesHrbpReviewedFilters
} from '../common/tables'
import { renderPerformanceLoadingState, renderWorkbenchHead } from '../common/workbenchPage'

export default {
  name: 'HrbpReviewPage',
  setup() {
    const ctx = usePerformanceContext()
    const hrbpDetailReview = ref(null)
    const hrbpReviewedFiltersCollapsed = ref(true)
    const hrbpReviewedFilterCount = computed(() => [
      ctx.hrbpReviewedFilters.employeeName,
      ctx.hrbpReviewedFilters.period,
      ctx.hrbpReviewedFilters.grade
    ].filter(Boolean).length)
    const hrbpTableRows = computed(() => {
      const pendingRows = ctx.currentReviews.value.filter((review) => review.status === 'hrbp_review')
      const reviewedRows = ctx.currentReviews.value
        .filter((review) => ['employee_confirm', 'hr_final', 'completed'].includes(review.status))
        .filter((review) => reviewMatchesHrbpReviewedFilters(ctx, review))
      return ctx.hrbpReviewTab.value === 'pending' ? pendingRows : reviewedRows
    })
    const hrbpPagination = useTablePagination(hrbpTableRows)

    function openHrbpReviewRow(review) {
      if (ctx.hrbpReviewTab.value === 'pending') {
        hrbpDetailReview.value = null
        ctx.selectReview(review.id)
        ctx.reviewTab.value = 'hrbp'
        return
      }
      ctx.selectedReviewId.value = ''
      ctx.reviewTab.value = 'currentTargets'
      hrbpDetailReview.value = review
    }

    function toggleHrbpReviewedFilters() {
      hrbpReviewedFiltersCollapsed.value = !hrbpReviewedFiltersCollapsed.value
    }

    return () => {
      const head = renderWorkbenchHead(ctx)
      if (ctx.contentLoading.value) {
        return h('view', { class: 'workbench workbench-loading' }, [
          head,
          renderPerformanceLoadingState()
        ])
      }
      const selectedHrbpReview = ctx.selectedReviewId.value ? ctx.selectedReview.value : null
      if (selectedHrbpReview) {
        return h('view', { class: 'workbench hrbp-review-detail-page' }, [
          head,
          h('view', { class: 'detail-page-toolbar' }, [
            h('button', { class: 'dt-btn dt-btn-light', onClick: () => { ctx.selectedReviewId.value = '' } }, '返回列表')
          ]),
          h('section', { class: 'panel detail-panel hrbp-review-detail-panel' }, [
            h(ReviewForm, { review: selectedHrbpReview })
          ])
        ])
      }
      const rows = hrbpPagination.rows.value
      return h('view', { class: 'workbench hrbp-review-page' }, [
        head,
        h('view', { class: 'hrbp-review-tabs-bar' }, [
          h('view', { class: 'hrbp-review-tabs' }, [
            h('button', { class: ['hrbp-review-tab', ctx.hrbpReviewTab.value === 'pending' ? 'active' : ''], onClick: () => ctx.switchHrbpReviewTab('pending') }, '待评'),
            h('button', { class: ['hrbp-review-tab', ctx.hrbpReviewTab.value === 'reviewed' ? 'active' : ''], onClick: () => ctx.switchHrbpReviewTab('reviewed') }, '已评')
          ])
        ]),
        h('view', { class: 'table-panel-stack hrbp-table-panel-stack' }, [
          h('section', { class: 'panel table-panel' }, [
            renderHrbpReviewedFilters(ctx, hrbpReviewedFiltersCollapsed.value, hrbpReviewedFilterCount.value, toggleHrbpReviewedFilters),
            renderHrbpReviewTable(ctx, rows, openHrbpReviewRow)
          ]),
          renderTablePagination(hrbpPagination)
        ]),
        hrbpDetailReview.value
          ? h(Teleport, { to: 'body' }, [
              h(HistoryReviewDetailModal, {
                review: hrbpDetailReview.value,
                onClose: () => { hrbpDetailReview.value = null }
              })
            ])
          : null
      ])
    }
  }
}
</script>
