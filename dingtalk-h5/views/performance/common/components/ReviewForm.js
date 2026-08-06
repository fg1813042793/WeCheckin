import { h, ref } from 'vue'
import { usePerformanceContext } from '../context'
import { ProcessProgressModal } from './ProcessProgressModal'
import {
  CurrentObjectiveSection,
  HrbpSection,
  ManagerSection,
  NextObjectiveSection,
  SelfSummarySection,
  ValueSection
} from './ReviewFormSections'
import { currentAssignee, flowProgressRows } from '../reviewFlow'

function reviewDetailTitle(ctx, review) {
  const title = `${review.period || '-'} 月度考评`
  if (ctx.contentView.value === 'mine') return title
  return `${ctx.userName(review.employeeId)} · ${title}`
}

const reviewFormTabs = [
  ['currentTargets', '本月目标'],
  ['selfSummary', '思考总结'],
  ['selfValues', '价值观自评'],
  ['manager', '上级评价'],
  ['hrbp', 'HRBP评价'],
  ['nextTargets', '下月目标']
].map(([key, label]) => ({ key, label }))

function normalizeReviewFormTab(ctx) {
  if (ctx.reviewTab.value === 'current') ctx.reviewTab.value = 'currentTargets'
  if (ctx.reviewTab.value === 'next') ctx.reviewTab.value = 'nextTargets'
  if (!reviewFormTabs.some((item) => item.key === ctx.reviewTab.value)) {
    ctx.reviewTab.value = reviewFormTabs[0].key
  }
  return ctx.reviewTab.value
}

function renderReviewFormPane(ctx, review, editableSelf, editableManager, editableHrbp, readonly = false) {
  switch (ctx.reviewTab.value) {
    case 'selfSummary':
      return h(SelfSummarySection, { review, editableSelf })
    case 'selfValues':
      return h(ValueSection, { review, field: 'self', title: '价值观自评', editable: editableSelf })
    case 'manager':
      return h(ManagerSection, { review, editable: editableManager })
    case 'hrbp':
      return h(HrbpSection, { review, editable: editableHrbp })
    case 'nextTargets':
      return h(NextObjectiveSection, { review, editable: editableSelf })
    case 'currentTargets':
    default:
      return h(CurrentObjectiveSection, { review, editableSelf, readonly })
  }
}

export const ReviewForm = {
  props: ['review', 'readonly'],
  setup(props) {
    const ctx = usePerformanceContext()
    const processVisible = ref(false)

    return () => {
      const review = props.review
      const readonly = props.readonly === true
      const editableSelf = !readonly && ctx.canSelf(review)
      const editableManager = !readonly && ctx.canManager(review)
      const editableHrbp = !readonly && ctx.canHrbpHandle(review)
      const editableConfirm = !readonly && ctx.canEmployeeConfirm(review)
      const editableFinal = !readonly && ctx.canFinal(review)
      const showReturnHrbp = editableFinal && review.status === 'hr_final' && ctx.contentView.value === 'mine'
      const showWithdraw = ctx.canWithdraw(review)
      const activeFormTab = normalizeReviewFormTab(ctx)
      const currentStep = ctx.statusMeta[review.status]?.step || 0
      const progressRows = flowProgressRows(ctx, review)
      const currentProgress = progressRows[currentStep] || progressRows[0]

      return h('view', { class: 'review-form' }, [
        h('view', { class: 'detail-head' }, [
          h('view', {}, [
            h('view', { class: 'detail-title-row' }, [
              h('text', { class: 'detail-title' }, reviewDetailTitle(ctx, review)),
              h('button', { class: 'dt-btn dt-btn-light process-title-btn', onClick: () => { processVisible.value = true } }, '查看流程进度')
            ]),
            ctx.contentView.value === 'mine'
              ? null
              : h('text', { class: 'detail-subtitle' }, `${review.nextPeriod} 目标 · 当前处理人 ${currentAssignee(ctx, review)}`)
          ]),
          ctx.contentView.value === 'mine'
            ? null
            : h('text', { class: ctx.reviewStatusClass(review) }, ctx.reviewStatusText(review))
        ]),
        h('view', { class: 'process-summary' }, [
          h('view', { class: 'process-summary-main' }, [
            h('text', { class: 'process-kicker' }, '当前流程状态'),
            h('view', { class: 'process-status-line' }, [
              h('text', { class: ['process-state-badge', currentProgress.stateClass] }, currentProgress.stateLabel),
              h('text', { class: ctx.reviewStatusClass(review) }, ctx.reviewStatusText(review)),
              h('text', { class: 'process-handler' }, `当前处理人 ${currentAssignee(ctx, review)}`)
            ]),
            h('text', { class: 'process-desc' }, `${currentProgress.progressText} · ${currentProgress.detail}`)
          ]),
          h('button', { class: 'dt-btn dt-btn-light process-help-btn', onClick: () => { processVisible.value = true } }, '查看流程进度')
        ]),
        processVisible.value ? h(ProcessProgressModal, { review, onClose: () => { processVisible.value = false } }) : null,
        h('view', { class: 'review-tabs' }, reviewFormTabs.map((item) =>
          h('button', { class: ['review-tab', activeFormTab === item.key ? 'active' : ''], onClick: () => { ctx.reviewTab.value = item.key } }, item.label)
        )),
        h('view', { class: 'review-form-pane' }, [
          renderReviewFormPane(ctx, review, editableSelf, editableManager, editableHrbp, readonly)
        ]),
        readonly ? null : h('view', { class: 'action-bar' }, [
          editableSelf ? h('button', { class: 'dt-btn dt-btn-light', onClick: () => ctx.performReviewAction('save-self', '已保存') }, '保存') : null,
          editableSelf ? h('button', { class: 'dt-btn dt-btn-primary', onClick: () => ctx.performReviewAction('submit-self', '已提交给上级') }, '提交自评') : null,
          editableManager ? h('button', { class: 'dt-btn dt-btn-danger-light', onClick: () => ctx.returnReview('return-employee', '退回员工修改') }, '退回员工') : null,
          editableManager ? h('button', { class: 'dt-btn dt-btn-primary', onClick: () => ctx.performReviewAction('submit-manager', '已提交给 HRBP') }, '提交给HRBP') : null,
          editableHrbp ? h('button', { class: 'dt-btn dt-btn-danger-light', onClick: () => ctx.returnReview('return-manager', '退回上级修改') }, '退回上级') : null,
          editableHrbp ? h('button', { class: 'dt-btn dt-btn-primary', onClick: () => ctx.performReviewAction('submit-hrbp', '已提交给员工确认') }, '提交给员工') : null,
          editableConfirm ? h('button', { class: 'dt-btn dt-btn-primary', onClick: () => ctx.performReviewAction('confirm-result', '已确认') }, '确认结果') : null,
          editableConfirm ? h('button', { class: 'dt-btn dt-btn-danger-light', onClick: () => ctx.performReviewAction('dispute-result', '已提出异议') }, '提出异议') : null,
          showReturnHrbp ? h('button', { class: 'dt-btn dt-btn-danger-light', onClick: () => ctx.returnReview('return-hrbp', '退回 HRBP 修改') }, '退回HRBP') : null,
          showWithdraw ? h('button', { class: 'dt-btn dt-btn-light', onClick: ctx.withdrawReview }, '撤回提交') : null,
          editableFinal ? h('button', { class: 'dt-btn dt-btn-primary', onClick: () => ctx.performReviewAction('finalize', '已归档') }, '归档') : null
        ])
      ])
    }
  }
}
