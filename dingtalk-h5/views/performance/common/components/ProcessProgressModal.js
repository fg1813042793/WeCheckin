import { h } from 'vue'
import { usePerformanceContext } from '../context'
import { currentAssignee, flowProgressRows } from '../reviewFlow'

export const ProcessProgressModal = {
  props: ['review'],
  emits: ['close'],
  setup(props, { emit }) {
    const ctx = usePerformanceContext()

    return () => {
      const review = props.review
      const currentStep = ctx.statusMeta[review.status]?.step || 0
      const progressRows = flowProgressRows(ctx, review)
      const currentProgress = progressRows[currentStep] || progressRows[0]
      return h('view', { class: 'process-modal-mask', onClick: () => emit('close') }, [
        h('view', { class: 'process-modal', onClick: (event) => event.stopPropagation() }, [
          h('view', { class: 'process-modal-head' }, [
            h('view', {}, [
              h('text', { class: 'process-modal-title' }, '流程进度'),
              h('text', { class: 'process-modal-subtitle' }, `当前：${ctx.reviewStatusText(review)} · 处理人 ${currentAssignee(ctx, review)}`)
            ]),
            h('button', { class: 'process-modal-close', onClick: () => emit('close') }, '×')
          ]),
          h('view', { class: 'process-current-card compact' }, [
            h('text', { class: 'process-current-index' }, currentProgress.indexText),
            h('view', { class: 'process-current-main' }, [
              h('view', { class: 'process-current-title-line' }, [
                h('text', { class: 'process-current-label' }, '当前节点'),
                h('text', { class: 'process-current-title' }, currentProgress.title),
                h('text', { class: 'process-current-progress' }, currentProgress.progressText)
              ]),
              h('view', { class: 'process-current-meta' }, [
                h('text', { class: ctx.reviewStatusClass(review) }, ctx.reviewStatusText(review)),
                h('text', { class: ['process-state-badge', currentProgress.stateClass] }, currentProgress.stateLabel),
                h('text', { class: 'process-current-handler' }, `处理人 ${currentAssignee(ctx, review)}`)
              ])
            ])
          ]),
          h('view', { class: 'process-timeline' }, progressRows.map((step) =>
            h('view', { class: ['process-step-row', step.state] }, [
              h('text', { class: 'process-step-index' }, step.indexText),
              h('view', { class: 'process-step-main' }, [
                h('view', { class: 'process-step-title-line' }, [
                  h('text', { class: 'process-step-title' }, step.title),
                  h('text', { class: 'process-step-role' }, step.role),
                  h('text', { class: ['process-state-badge', step.stateClass] }, step.stateLabel)
                ]),
                h('text', { class: 'process-step-desc' }, step.detail),
                step.reason ? h('view', { class: 'process-step-reason' }, [
                  h('text', { class: 'process-step-reason-label' }, step.reasonLabel || '理由'),
                  h('text', { class: 'process-step-reason-text' }, step.reason)
                ]) : null,
                h('text', { class: 'process-step-help' }, step.desc)
              ])
            ])
          )),
          h('view', { class: 'process-modal-actions' }, [
            h('button', { class: 'dt-btn dt-btn-primary', onClick: () => emit('close') }, '知道了')
          ])
        ])
      ])
    }
  }
}
