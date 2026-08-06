<script>
import { h } from 'vue'
import { usePerformanceContext } from '../common/context'
import { ReviewForm } from '../common/components/ReviewForm'
import {
  renderMyPerformanceSwitcher,
  renderPerformanceEmptyState,
  renderPerformanceLoadingState,
  renderWorkbenchHead
} from '../common/workbenchPage'

export default {
  name: 'MinePage',
  setup() {
    const ctx = usePerformanceContext()
    return () => {
      const head = renderWorkbenchHead(ctx, { showCreate: true, onCreate: () => ctx.openCreateReviewDialog() })
      if (ctx.contentLoading.value) {
        return h('view', { class: 'workbench workbench-loading' }, [
          head,
          renderPerformanceLoadingState()
        ])
      }
      return h('view', { class: 'workbench' }, [
        head,
        h('view', { class: 'review-detail-only' }, [
          renderMyPerformanceSwitcher(ctx),
          h('section', { class: 'panel detail-panel' }, ctx.selectedReview.value
            ? [h(ReviewForm, { review: ctx.selectedReview.value })]
            : [renderPerformanceEmptyState(ctx)])
        ])
      ])
    }
  }
}
</script>
