import { h } from 'vue'
import { usePerformanceContext } from '../context'
import { ReviewForm } from './ReviewForm'

export const HistoryReviewDetailModal = {
  props: ['review'],
  emits: ['close'],
  setup(props, { emit }) {
    const ctx = usePerformanceContext()

    return () => h('view', { class: 'history-detail-modal-mask', onClick: () => emit('close') }, [
      h('view', { class: 'history-detail-modal', onClick: (event) => event.stopPropagation() }, [
        h('view', { class: 'history-detail-modal-head' }, [
          h('view', {}, [
            h('text', { class: 'history-detail-modal-title' }, '绩效详情'),
            h('text', { class: 'history-detail-modal-subtitle' }, `${ctx.userName(props.review.employeeId)} · ${props.review.period || '-'} 月度考评`)
          ]),
          h('button', { class: 'process-modal-close', onClick: () => emit('close') }, '×')
        ]),
        h('view', { class: 'history-detail-modal-body' }, [
          h(ReviewForm, { review: props.review, readonly: true })
        ])
      ])
    ])
  }
}
