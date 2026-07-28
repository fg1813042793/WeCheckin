import { h } from 'vue'
import { usePerformanceContext } from './context'

export default {
  name: 'SummaryView',
  setup() {
    const ctx = usePerformanceContext()

    return () => h('view', { class: 'summary-page' }, [
      h('view', { class: 'page-head' }, [
        h('view', {}, [
          h('text', { class: 'page-title' }, ctx.sectionTitle.value),
          h('text', { class: 'page-desc' }, '按人员和月份查看进度、分档并导出结果')
        ]),
        h('view', { class: 'head-actions' }, [
          h('button', { class: 'dt-btn dt-btn-primary', onClick: ctx.exportSummary }, '导出当前筛选'),
          h('button', { class: 'dt-btn dt-btn-light', onClick: ctx.refreshData }, '刷新')
        ])
      ]),

      ['admin', 'hrbp', 'hrbp_manager'].includes(ctx.state.user.role)
        ? h('section', { class: 'panel create-panel' }, [
            h('view', { class: 'panel-head' }, [h('text', { class: 'panel-title' }, '新建考评单')]),
            h('view', { class: 'create-form' }, [
              h('select', { class: 'field-select', value: ctx.newReview.employeeId, onChange: (event) => { ctx.newReview.employeeId = event.target.value } }, ctx.reviewTargetUsers.value.map((user) => h('option', { value: user.id }, ctx.userOptionText(user)))),
              h('input', { class: 'field-input', type: 'month', value: ctx.newReview.period, onInput: (event) => { ctx.newReview.period = event.detail?.value ?? event.target.value } }),
              h('input', { class: 'field-input', type: 'month', value: ctx.newReview.nextPeriod, onInput: (event) => { ctx.newReview.nextPeriod = event.detail?.value ?? event.target.value } }),
              h('button', { class: 'dt-btn dt-btn-primary', onClick: ctx.createReview }, '创建')
            ])
          ])
        : null,

      h('section', { class: 'panel' }, [
        h('view', { class: 'panel-head' }, [
          h('text', { class: 'panel-title' }, '汇总列表'),
          h('text', { class: 'count-pill' }, `${ctx.summaryReviews.value.length} / ${ctx.state.reviews.length}`)
        ]),
        h('view', { class: 'filters' }, [
          h('input', { class: 'field-input', value: ctx.summaryFilters.keyword, placeholder: '员工、部门、上级', onInput: (event) => { ctx.summaryFilters.keyword = event.detail?.value ?? event.target.value } }),
          h('input', { class: 'field-input', type: 'month', value: ctx.summaryFilters.period, onInput: (event) => { ctx.summaryFilters.period = event.detail?.value ?? event.target.value } }),
          h('input', { class: 'field-input', type: 'month', value: ctx.summaryFilters.nextPeriod, onInput: (event) => { ctx.summaryFilters.nextPeriod = event.detail?.value ?? event.target.value } }),
          h('select', { class: 'field-select', value: ctx.summaryFilters.status, onChange: (event) => { ctx.summaryFilters.status = event.target.value } }, [''].concat(Object.keys(ctx.statusMeta)).map((status) => h('option', { value: status }, status ? ctx.statusText(status) : '全部状态'))),
          h('select', { class: 'field-select', value: ctx.summaryFilters.department, onChange: (event) => { ctx.summaryFilters.department = event.target.value } }, [''].concat(ctx.departments.value).map((item) => h('option', { value: item }, item || '全部部门'))),
          h('select', { class: 'field-select', value: ctx.summaryFilters.managerId, onChange: (event) => { ctx.summaryFilters.managerId = event.target.value } }, [''].concat(ctx.managerIds.value).map((id) => h('option', { value: id }, id ? ctx.userName(id) : '全部上级'))),
          h('select', { class: 'field-select', value: ctx.summaryFilters.hrbpId, onChange: (event) => { ctx.summaryFilters.hrbpId = event.target.value } }, [''].concat(ctx.hrbpIds.value).map((id) => h('option', { value: id }, id ? ctx.userName(id) : '全部HRBP'))),
          h('select', { class: 'field-select', value: ctx.summaryFilters.grade, onChange: (event) => { ctx.summaryFilters.grade = event.target.value } }, [''].concat(ctx.grades.value).map((grade) => h('option', { value: grade }, grade || '全部分档'))),
          h('button', { class: 'dt-btn dt-btn-light', onClick: ctx.resetFilters }, '重置')
        ]),
        h('view', { class: 'table-wrap' }, [
          h('table', { class: 'summary-table' }, [
            h('thead', {}, h('tr', {}, ['员工', '部门', '考评月份', '目标月份', '状态', '目标得分', '上级分档', 'HRBP分档', '最终分档', '员工确认', ctx.state.user.role === 'admin' ? '操作' : ''].filter(Boolean).map((cell) => h('th', {}, cell)))),
            h('tbody', {}, ctx.summaryReviews.value.map((review) => h('tr', {}, [
              h('td', {}, ctx.userName(review.employeeId)),
              h('td', {}, review.department),
              h('td', {}, review.period),
              h('td', {}, review.nextPeriod),
              h('td', {}, h('text', { class: ctx.statusClass(review.status) }, ctx.statusText(review.status))),
              h('td', {}, String(ctx.totalObjectiveScore(review))),
              h('td', {}, review.managerGrade || '-'),
              h('td', {}, review.hrbpGrade || '-'),
              h('td', {}, ctx.effectiveGrade(review) || '-'),
              h('td', {}, review.employeeConfirmResult === 'confirmed' ? '已确认' : review.employeeConfirmResult === 'disputed' ? '有异议' : '-'),
              ctx.state.user.role === 'admin' ? h('td', {}, h('button', { class: 'dt-btn dt-btn-danger-light small', onClick: () => ctx.deleteReview(review.id) }, '删除')) : null
            ])))
          ])
        ])
      ])
    ])
  }
}
