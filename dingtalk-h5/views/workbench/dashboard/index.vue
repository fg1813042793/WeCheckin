<script>
import { h } from 'vue'
import { usePerformanceContext } from '../../performance/common/context'
import { currentAssignee } from '../../performance/common/reviewFlow'
import { renderPerformanceLoadingState, renderWorkbenchHead } from '../../performance/common/workbenchPage'

function workbenchTodoMeta(ctx, review) {
  const period = review.period || '-'
  const employee = ctx.userName(review.employeeId)
  const hasDispute = review.employeeConfirmResult === 'disputed'
  if (ctx.canSelf(review)) {
    return {
      title: `${period} 月度考评待填写`,
      desc: '填写本月目标完成情况、思考总结、价值观自评和下月目标。',
      action: '去填写'
    }
  }
  if (ctx.canManager(review)) {
    return {
      title: `${employee} 的 ${period} 月度考评待上级评价`,
      desc: '审核员工自评，填写上级评价、价值观评分和建议分档。',
      action: '去评价'
    }
  }
  if (ctx.canHrbpHandle(review)) {
    return {
      title: `${employee} 的 ${period} 月度考评待 HRBP 评价`,
      desc: hasDispute ? '员工已提出异议，请复核评价结果并处理分档。' : '复核绩效材料，填写 HRBP 评价和分档。',
      action: '去评价'
    }
  }
  if (ctx.canEmployeeConfirm(review)) {
    return {
      title: `${period} 月度考评待确认`,
      desc: '查看评价结果，确认无误或提交异议说明。',
      action: '去确认'
    }
  }
  if (ctx.canFinal(review)) {
    return {
      title: `${employee} 的 ${period} 月度考评待归档`,
      desc: '处理员工确认或异议，确认最终分档并完成归档。',
      action: '去归档'
    }
  }
  return {
    title: `${period} 月度考评待处理`,
    desc: '查看当前绩效流程并处理待办。',
    action: '查看'
  }
}

function workbenchTodoResultBadge(review) {
  if (review.employeeConfirmResult === 'disputed') {
    return {
      text: '员工有异议',
      className: 'workbench-todo-result-disputed'
    }
  }
  if (review.employeeConfirmResult === 'confirmed') {
    return {
      text: '员工已确认',
      className: 'workbench-todo-result-confirmed'
    }
  }
  return null
}

function renderWorkbenchTodoBadges(ctx, review) {
  const resultBadge = workbenchTodoResultBadge(review)
  const badges = [
    h('text', { class: ctx.reviewStatusClass(review) }, ctx.reviewStatusText(review))
  ]
  if (resultBadge) {
    badges.push(h('text', { class: ['workbench-todo-result-badge', resultBadge.className] }, resultBadge.text))
  }
  return h('view', { class: 'workbench-todo-badges' }, badges)
}

function renderWorkbenchTodoList(ctx) {
  const todos = ctx.queueReviews()
  return h('section', { class: 'panel workbench-todo-panel' }, [
    h('view', { class: 'workbench-todo-head' }, [
      h('view', {}, [
        h('text', { class: 'panel-title' }, '我的待办'),
        h('text', { class: 'workbench-todo-desc' }, '点击待办可直接进入对应处理页面')
      ]),
      h('text', { class: 'count-pill' }, `${todos.length} 条`)
    ]),
    todos.length
      ? h('view', { class: 'workbench-todo-list' }, todos.map((review) => {
          const meta = workbenchTodoMeta(ctx, review)
          return h('button', { class: 'workbench-todo-item', onClick: () => ctx.openWorkbenchTodo(review) }, [
            h('view', { class: 'workbench-todo-main' }, [
              h('view', { class: 'workbench-todo-title-row' }, [
                h('text', { class: 'workbench-todo-title' }, meta.title),
                renderWorkbenchTodoBadges(ctx, review)
              ]),
              h('text', { class: 'workbench-todo-meta' }, [
                review.department ? `${review.department} · ` : '',
                `当前处理人 ${currentAssignee(ctx, review)}`
              ].join('')),
              h('text', { class: 'workbench-todo-desc-line' }, meta.desc)
            ]),
            h('text', { class: 'workbench-todo-action' }, meta.action)
          ])
        }))
      : h('view', { class: 'workbench-todo-empty' }, [
          h('text', { class: 'workbench-todo-empty-title' }, '当前没有待办'),
          h('text', { class: 'workbench-todo-empty-desc' }, '需要你填写、评价、确认或归档的绩效单会显示在这里。')
        ])
  ])
}

export default {
  name: 'DashboardPage',
  setup() {
    const ctx = usePerformanceContext()
    return () => {
      const head = renderWorkbenchHead(ctx)
      if (ctx.contentLoading.value) {
        return h('view', { class: 'workbench workbench-loading' }, [
          head,
          renderPerformanceLoadingState()
        ])
      }
      return h('view', { class: 'workbench' }, [
        head,
        renderWorkbenchTodoList(ctx)
      ])
    }
  }
}
</script>

<style>
/* 页面专属样式：从 styles/performance.css 拆分 */
.workbench-todo-panel {
  width: min(100%, 720px);
  justify-self: start;
  overflow: hidden;
}

.workbench-todo-head {
  min-height: 56px;
  padding: 0 16px;
  border-bottom: 1px solid #edf0f5;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
}

.workbench-todo-desc,
.workbench-todo-desc-line,
.workbench-todo-meta {
  display: block;
  color: #86909c;
  font-size: 13px;
}

.workbench-todo-desc {
  margin-top: 3px;
}

.workbench-todo-list {
  display: grid;
}

.workbench-todo-item {
  min-height: 76px;
  padding: 14px 16px;
  border: 0;
  border-bottom: 1px solid #edf0f5;
  border-radius: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  background: #fff;
  text-align: left;
  cursor: pointer;
}

.workbench-todo-item:last-child {
  border-bottom: 0;
}

.workbench-todo-item:hover {
  background: #f7fbff;
}

.workbench-todo-main {
  min-width: 0;
  display: grid;
  gap: 6px;
}

.workbench-todo-title-row {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 10px;
}

.workbench-todo-badges {
  flex: 0 0 auto;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.workbench-todo-title {
  min-width: 0;
  overflow: hidden;
  color: #1f2329;
  font-size: 15px;
  font-weight: 700;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.workbench-todo-result-badge {
  display: inline-flex;
  align-items: center;
  padding: 2px 8px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 700;
  line-height: 18px;
  white-space: nowrap;
}

.workbench-todo-result-disputed {
  background: #fff1f0;
  color: #c42a2a;
}

.workbench-todo-result-confirmed {
  background: #e8fff3;
  color: #0f8f54;
}

.workbench-todo-action {
  flex: 0 0 auto;
  padding: 0 12px;
  border-radius: 4px;
  background: #eaf3ff;
  color: #1677ff;
  font-size: 13px;
  font-weight: 700;
  line-height: 30px;
}

.workbench-todo-empty {
  min-height: 180px;
  padding: 36px 16px;
  display: grid;
  place-items: center;
  align-content: center;
  gap: 8px;
  text-align: center;
}

.workbench-todo-empty-title {
  color: #1f2329;
  font-size: 16px;
  font-weight: 700;
}

.workbench-todo-empty-desc {
  color: #86909c;
  font-size: 13px;
}

.workbench-loading {
  min-height: 420px;
}

@media (max-width: 960px) {
  .workbench-todo-head {
    min-height: 50px;
    padding: 0 12px;
  }
  .workbench-todo-item {
    min-height: 0;
    padding: 14px 12px;
    align-items: stretch;
    flex-direction: column;
    gap: 10px;
  }
  .workbench-todo-title-row {
    align-items: flex-start;
    flex-direction: column;
    gap: 6px;
  }
  .workbench-todo-title {
    width: 100%;
    white-space: normal;
  }
  .workbench-todo-action {
    justify-self: start;
    align-self: flex-start;
  }
}
</style>
