import { h } from 'vue'

export function myPerformanceActionText(review) {
  if (review.status === 'manager_review') return '等待上级评价'
  if (review.status === 'hrbp_review') return '等待HRBP评价'
  if (review.status === 'employee_confirm') return '等待确认结果'
  if (review.status === 'hr_final') return '等待HRBP归档'
  return '等待填写自评'
}

export function workbenchPageDesc(ctx) {
  if (ctx.state.view === 'dashboard') return '处理需要你跟进的绩效事项'
  if (ctx.contentLoading.value) return '加载中...'
  if (ctx.contentView.value === 'mine') return `共 ${ctx.currentReviews.value.length} 条进行中`
  return `共 ${Number(ctx.state.reviewListTotal || ctx.currentReviews.value.length)} 条记录`
}

export function renderWorkbenchHead(ctx, options = {}) {
  const showCreate = options.showCreate && ctx.canCreateReview()
  const onCreate = options.onCreate || ctx.openCreateReviewDialog
  return h('view', { class: 'page-head' }, [
    h('view', {}, [
      h('text', { class: 'page-title' }, ctx.sectionTitle.value),
      h('text', { class: 'page-desc' }, workbenchPageDesc(ctx))
    ]),
    h('view', { class: 'head-actions' }, [
      showCreate ? h('button', { class: 'dt-btn dt-btn-primary', onClick: onCreate }, '创建') : null,
      h('button', { class: 'dt-btn dt-btn-light', onClick: ctx.refreshData }, '刷新')
    ])
  ])
}

function emptyReviewCopy(ctx) {
  const view = ctx.contentView.value
  if (view === 'manager') {
    return {
      title: '暂无待处理绩效单',
      desc: '员工提交自评后，会在这里进入上级评价队列。',
      steps: ['等待员工提交', '上级评价', '继续流转']
    }
  }
  if (view === 'dashboard') {
    return {
      title: '暂无待办绩效单',
      desc: '有新的填写、评价或确认任务时，会自动汇总到这里。',
      steps: ['员工填写', '上级评价', 'HRBP归档']
    }
  }
  if (view === 'mine') {
    return {
      title: '暂无进行中绩效单',
      desc: '属于你的绩效单只要流程未完成，就会在这里展示。',
      steps: ['等待创建', '流程流转', '完成归档']
    }
  }
  return {
    title: '本月还没有绩效单',
    desc: '绩效单创建后，可以在这里填写目标完成度、思考总结和自评内容。',
    steps: ['等待创建', '填写自评', '提交流转']
  }
}

export function renderPerformanceEmptyState(ctx) {
  const copy = emptyReviewCopy(ctx)
  return h('view', { class: 'performance-empty' }, [
    h('view', { class: 'performance-empty-visual' }, [
      h('view', { class: 'performance-empty-icon' }, '绩'),
      h('view', { class: 'performance-empty-line wide' }),
      h('view', { class: 'performance-empty-line' }),
      h('view', { class: 'performance-empty-dots' }, [
        h('text'),
        h('text'),
        h('text')
      ])
    ]),
    h('text', { class: 'performance-empty-title' }, copy.title),
    h('text', { class: 'performance-empty-desc' }, copy.desc),
    h('view', { class: 'performance-empty-steps' }, copy.steps.map((item, index) =>
      h('text', { class: 'performance-empty-step' }, `${index + 1}. ${item}`)
    )),
    h('button', { class: 'dt-btn dt-btn-light performance-empty-action', onClick: ctx.refreshData }, '刷新看看')
  ])
}

export function renderPerformanceLoadingState() {
  return h('section', { class: 'panel performance-loading-panel' }, [
    h('view', { class: 'performance-loading-head' }, [
      h('view', { class: 'performance-loading-title' }),
      h('view', { class: 'performance-loading-action' })
    ]),
    h('view', { class: 'performance-loading-table' }, Array.from({ length: 6 }).map((_, rowIndex) =>
      h('view', { class: 'performance-loading-row' }, Array.from({ length: 6 }).map((__, cellIndex) =>
        h('view', {
          class: [
            'performance-loading-cell',
            cellIndex === 0 ? 'wide' : '',
            rowIndex === 0 ? 'head' : ''
          ]
        })
      ))
    ))
  ])
}

export function renderMyPerformanceSwitcher(ctx) {
  const rows = ctx.currentReviews.value
  if (ctx.contentView.value !== 'mine' || rows.length <= 1) return null
  return h('section', { class: 'panel my-performance-switcher' }, [
    h('view', { class: 'my-performance-switcher-head' }, [
      h('text', { class: 'my-performance-switcher-title' }, '进行中绩效'),
      h('text', { class: 'count-pill' }, `${rows.length} 条`)
    ]),
    h('view', { class: 'my-performance-switcher-list' }, rows.map((review) =>
      h('button', {
        class: ['my-performance-switcher-item', ctx.selectedReview.value?.id === review.id ? 'active' : ''],
        onClick: () => ctx.selectReview(review.id)
      }, [
        h('view', { class: 'my-performance-switcher-main' }, [
          h('text', { class: 'my-performance-switcher-name' }, `${review.period || '-'} 月度考评`),
          h('text', { class: 'my-performance-switcher-desc' }, myPerformanceActionText(review))
        ]),
        h('text', { class: ctx.reviewStatusClass(review) }, ctx.reviewStatusText(review))
      ])
    ))
  ])
}
