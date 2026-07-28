import { h, ref } from 'vue'
import { usePerformanceContext } from './context'

function readInputValue(event) {
  return event.detail?.value ?? event.target.value
}

function currentAssignee(ctx, review) {
  if (review.status === 'draft') return ctx.userName(review.employeeId)
  if (review.status === 'manager_review') return ctx.userName(review.managerId)
  if (review.status === 'hrbp_review') return ctx.userName(review.hrbpReviewerId || review.hrbpId)
  if (review.status === 'employee_confirm') return ctx.userName(review.employeeId)
  if (review.status === 'hr_final') return ctx.userName(review.hrbpReviewerId || review.hrbpId)
  return '已归档'
}

const flowSteps = [
  {
    status: 'draft',
    title: '员工填写',
    role: '员工',
    desc: '填写当月目标完成度、达成结果自评、思考总结，并确认下月目标。',
    actor: 'employee',
    historyKeywords: ['创建考评单', '保存员工自评', '提交员工自评', '退回员工修改', '撤销员工自评提交']
  },
  {
    status: 'manager_review',
    title: '上级评价',
    role: '直属上级',
    desc: '审核员工自评内容，填写上级评价、价值观评分和建议分档。',
    actor: 'manager',
    historyKeywords: ['提交上级评价', '退回上级修改', '撤销上级评价提交']
  },
  {
    status: 'hrbp_review',
    title: 'HRBP评价',
    role: 'HRBP',
    desc: '复核绩效材料，填写 HRBP 评价和分档；如有问题可退回上级修改。',
    actor: 'hrbp',
    historyKeywords: ['提交 HRBP 评价', '员工提出异议', '撤销 HRBP 评价提交', '退回 HRBP 修改']
  },
  {
    status: 'employee_confirm',
    title: '员工确认',
    role: '员工',
    desc: '查看评价结果并确认；如存在异议，可填写说明后提交反馈。',
    actor: 'employee',
    historyKeywords: ['员工确认结果', '员工提出异议']
  },
  {
    status: 'hr_final',
    title: 'HRBP归档',
    role: 'HRBP',
    desc: '处理员工确认或异议，确认最终分档和归档备注。',
    actor: 'hrbp',
    historyKeywords: ['HRBP 归档', '退回 HRBP 修改', '员工确认结果']
  },
  {
    status: 'completed',
    title: '完成',
    role: '系统归档',
    desc: '考评单归档完成，结果进入汇总统计。',
    actor: 'system',
    historyKeywords: ['HRBP 归档']
  }
]

function stepActorName(ctx, review, step) {
  if (step.actor === 'employee') return ctx.userName(review.employeeId)
  if (step.actor === 'manager') return ctx.userName(review.managerId)
  if (step.actor === 'hrbp') return ctx.userName(review.hrbpReviewerId || review.hrbpId)
  return '系统'
}

function formatHistoryTime(at) {
  if (!at) return ''
  const date = new Date(at)
  if (Number.isNaN(date.getTime())) return ''
  return date.toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

function latestHistoryForStep(review, step) {
  const histories = Array.isArray(review.history) ? review.history : []
  return histories.slice().reverse().find((item) =>
    step.historyKeywords.some((keyword) => String(item.action || '').includes(keyword))
  )
}

function flowProgressRows(ctx, review) {
  const currentStep = ctx.statusMeta[review.status]?.step || 0
  return flowSteps.map((step, index) => {
    const history = latestHistoryForStep(review, step)
    const isCompletedFinal = review.status === 'completed' && index === currentStep
    const state = index < currentStep || isCompletedFinal ? 'done' : index === currentStep ? 'active' : 'pending'
    const stateLabel = state === 'done' ? '已完成' : state === 'active' ? '进行中' : '待处理'
    const time = formatHistoryTime(history?.at)
    const detail = history
      ? `${history.action} · ${history.by || 'system'}${time ? ` · ${time}` : ''}`
      : state === 'active'
        ? `等待 ${stepActorName(ctx, review, step)} 处理`
        : state === 'done'
          ? '已进入下一流程节点'
          : step.actor === 'system'
            ? '待归档完成'
            : `待 ${step.role} 处理`
    return {
      ...step,
      index,
      indexText: String(index + 1).padStart(2, '0'),
      progressText: `${String(index + 1).padStart(2, '0')}/${String(flowSteps.length).padStart(2, '0')}`,
      actorName: stepActorName(ctx, review, step),
      state,
      stateLabel,
      detail,
      history
    }
  })
}

const WorkbenchView = {
  name: 'WorkbenchView',
  setup() {
    const ctx = usePerformanceContext()

    return () => {
      const head = h('view', { class: 'page-head' }, [
        h('view', {}, [
          h('text', { class: 'page-title' }, ctx.sectionTitle.value),
          h('text', { class: 'page-desc' }, ctx.state.view === 'dashboard'
            ? `${ctx.roleText(ctx.state.user.role)}视角，查看当前绩效概况`
            : `${ctx.roleText(ctx.state.user.role)}视角，共 ${ctx.currentReviews.value.length} 条记录`)
        ]),
        h('button', { class: 'dt-btn dt-btn-light', onClick: ctx.refreshData }, '刷新')
      ])

      if (ctx.state.view === 'dashboard') {
        return h('view', { class: 'workbench' }, [
          head,
          h('view', { class: 'stats-grid workbench-stats-grid' }, ctx.workbenchCards.value.map(([key, label, value]) =>
            h('view', { class: ['stat-card', 'static', `stat-card-${key}`] }, [
              h('text', { class: 'stat-label' }, label),
              h('text', { class: 'stat-value' }, String(value))
            ])
          ))
        ])
      }

      return h('view', { class: 'workbench' }, [
        head,
        h('view', { class: 'review-grid' }, [
          h('section', { class: 'panel list-panel' }, [
            h('view', { class: 'panel-head' }, [
              h('text', { class: 'panel-title' }, ctx.sectionTitle.value),
              h('text', { class: 'count-pill' }, String(ctx.currentReviews.value.length))
            ]),
            h('view', { class: 'review-list' }, ctx.currentReviews.value.length
              ? ctx.currentReviews.value.map((review) => h('button', {
                  class: ['review-row', ctx.selectedReview.value?.id === review.id ? 'active' : ''],
                  onClick: () => ctx.selectReview(review.id)
                }, [
                  h('view', { class: 'review-row-top' }, [
                    h('text', { class: 'review-name' }, ctx.userName(review.employeeId)),
                    h('text', { class: ctx.statusClass(review.status) }, ctx.statusText(review.status))
                  ]),
                  h('text', { class: 'review-meta' }, `${review.department} · ${review.period} 考评 / ${review.nextPeriod} 目标`),
                  h('text', { class: 'review-meta' }, `目标得分 ${ctx.totalObjectiveScore(review)} · 当前 ${currentAssignee(ctx, review)}`)
                ]))
              : h('view', { class: 'empty' }, '当前没有记录'))
          ]),
          h('section', { class: 'panel detail-panel' }, ctx.selectedReview.value
            ? [h(ReviewForm, { review: ctx.selectedReview.value })]
            : [h('view', { class: 'empty' }, '选择一张考评单')])
        ])
      ])
    }
  }
}

const ReviewForm = {
  props: ['review'],
  setup(props) {
    const ctx = usePerformanceContext()
    const processVisible = ref(false)

    return () => {
      const review = props.review
      const editableSelf = ctx.canSelf(review)
      const editableManager = ctx.canManager(review)
      const editableHrbp = ctx.canHrbpHandle(review)
      const editableConfirm = ctx.canEmployeeConfirm(review)
      const editableFinal = ctx.canFinal(review)
      const activeNext = ctx.reviewTab.value === 'next'
      const currentStep = ctx.statusMeta[review.status]?.step || 0
      const progressRows = flowProgressRows(ctx, review)
      const currentProgress = progressRows[currentStep] || progressRows[0]
      return h('view', { class: 'review-form' }, [
        h('view', { class: 'detail-head' }, [
          h('view', {}, [
            h('text', { class: 'detail-title' }, `${ctx.userName(review.employeeId)} · ${review.period} 月度考评`),
            h('text', { class: 'detail-subtitle' }, `${review.nextPeriod} 目标 · 当前处理人 ${currentAssignee(ctx, review)}`)
          ]),
          h('text', { class: ctx.statusClass(review.status) }, ctx.statusText(review.status))
        ]),
        h('view', { class: 'process-summary' }, [
          h('view', { class: 'process-summary-main' }, [
            h('text', { class: 'process-kicker' }, '当前流程状态'),
            h('view', { class: 'process-status-line' }, [
              h('text', { class: ctx.statusClass(review.status) }, ctx.statusText(review.status)),
              h('text', { class: ['process-state-badge', currentProgress.state] }, currentProgress.stateLabel),
              h('text', { class: 'process-handler' }, `当前处理人 ${currentAssignee(ctx, review)}`)
            ]),
            h('text', { class: 'process-desc' }, `${currentProgress.progressText} · ${currentProgress.detail}`)
          ]),
          h('button', { class: 'dt-btn dt-btn-light process-help-btn', onClick: () => { processVisible.value = true } }, '查看流程进度')
        ]),
        processVisible.value ? h(ProcessModal, { review, onClose: () => { processVisible.value = false } }) : null,
        h('view', { class: 'review-tabs' }, [
          h('button', { class: ['review-tab', activeNext ? '' : 'active'], onClick: () => { ctx.reviewTab.value = 'current' } }, '当月绩效'),
          h('button', { class: ['review-tab', activeNext ? 'active' : ''], onClick: () => { ctx.reviewTab.value = 'next' } }, '下月目标')
        ]),
        activeNext
          ? h(NextObjectiveSection, { review, editable: editableSelf })
          : h(CurrentPerformanceSection, { review, editableSelf }),
        review.status !== 'draft' ? h(ManagerSection, { review, editable: editableManager }) : null,
        ['hrbp_review', 'employee_confirm', 'hr_final', 'completed'].includes(review.status) ? h(HrbpSection, { review, editable: editableHrbp }) : null,
        ['employee_confirm', 'hr_final', 'completed'].includes(review.status) ? h(EmployeeConfirmSection, { review, editable: editableConfirm }) : null,
        ['hr_final', 'completed'].includes(review.status) ? h(FinalSection, { review, editable: editableFinal }) : null,
        h(HistorySection, { review }),
        h('view', { class: 'action-bar' }, [
          editableSelf ? h('button', { class: 'dt-btn dt-btn-light', onClick: () => ctx.performReviewAction('save-self', '已保存') }, '保存') : null,
          editableSelf ? h('button', { class: 'dt-btn dt-btn-primary', onClick: () => ctx.performReviewAction('submit-self', '已提交给上级') }, '提交自评') : null,
          editableManager ? h('button', { class: 'dt-btn dt-btn-danger-light', onClick: () => ctx.returnReview('return-employee', '退回员工修改') }, '退回员工') : null,
          editableManager ? h('button', { class: 'dt-btn dt-btn-primary', onClick: () => ctx.performReviewAction('submit-manager', '已提交给 HRBP') }, '提交给HRBP') : null,
          editableHrbp ? h('button', { class: 'dt-btn dt-btn-danger-light', onClick: () => ctx.returnReview('return-manager', '退回上级修改') }, '退回上级') : null,
          editableHrbp ? h('button', { class: 'dt-btn dt-btn-primary', onClick: () => ctx.performReviewAction('submit-hrbp', '已提交给员工确认') }, '提交给员工') : null,
          editableConfirm ? h('button', { class: 'dt-btn dt-btn-primary', onClick: () => ctx.performReviewAction('confirm-result', '已确认') }, '确认结果') : null,
          editableConfirm ? h('button', { class: 'dt-btn dt-btn-danger-light', onClick: () => ctx.performReviewAction('dispute-result', '已提出异议') }, '提出异议') : null,
          editableFinal && review.status === 'hr_final' ? h('button', { class: 'dt-btn dt-btn-danger-light', onClick: () => ctx.returnReview('return-hrbp', '退回 HRBP 修改') }, '退回HRBP') : null,
          editableFinal ? h('button', { class: 'dt-btn dt-btn-primary', onClick: () => ctx.performReviewAction('finalize', '已归档') }, '归档') : null,
          ctx.canWithdraw(review) ? h('button', { class: 'dt-btn dt-btn-light', onClick: ctx.withdrawReview }, '撤销提交') : null
        ])
      ])
    }
  }
}

const ProcessModal = {
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
              h('text', { class: 'process-modal-subtitle' }, `当前：${ctx.statusText(review.status)} · 处理人 ${currentAssignee(ctx, review)}`)
            ]),
            h('button', { class: 'process-modal-close', onClick: () => emit('close') }, '×')
          ]),
          h('view', { class: 'process-current-card' }, [
            h('text', { class: 'process-kicker' }, '当前流程状态'),
            h('view', { class: 'process-status-line' }, [
              h('text', { class: ctx.statusClass(review.status) }, ctx.statusText(review.status)),
              h('text', { class: ['process-state-badge', currentProgress.state] }, currentProgress.stateLabel),
              h('text', { class: 'process-handler' }, currentProgress.progressText)
            ])
          ]),
          h('view', { class: 'process-timeline' }, progressRows.map((step) =>
            h('view', { class: ['process-step-row', step.state] }, [
              h('text', { class: 'process-step-index' }, step.indexText),
              h('view', { class: 'process-step-main' }, [
                h('view', { class: 'process-step-title-line' }, [
                  h('text', { class: 'process-step-title' }, step.title),
                  h('text', { class: 'process-step-role' }, step.role),
                  h('text', { class: ['process-state-badge', step.state] }, step.stateLabel)
                ]),
                h('text', { class: 'process-step-desc' }, step.detail),
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

const CurrentPerformanceSection = {
  props: ['review', 'editableSelf'],
  setup(props) {
    const ctx = usePerformanceContext()

    return () => h('view', {}, [
      h('section', { class: 'form-section' }, [
        h('view', { class: 'section-title' }, [
          h('text', {}, '本月目标考评'),
          h('text', { class: 'count-pill' }, `合计 ${ctx.totalObjectiveScore(props.review)}`)
        ]),
        h('view', { class: 'objective-list' }, props.review.objectives.map((item, index) => h('view', { class: 'objective-card' }, [
          h('view', { class: 'objective-head' }, [
            h('text', { class: 'objective-title' }, `目标 ${index + 1}`),
            h('text', { class: 'score-badge' }, `${ctx.objectiveScore(item)} 分`)
          ]),
          h('view', { class: 'field-block field-block-wide' }, [
            h('text', { class: 'field-label' }, '目标描述'),
            h('textarea', { class: 'field-textarea', value: item.target, disabled: !ctx.canEditObjectiveDimension(props.review), placeholder: '绩效目标', onInput: (event) => { item.target = readInputValue(event) } })
          ]),
          h('view', { class: 'objective-fields' }, [
            h('view', { class: 'field-block' }, [
              h('text', { class: 'field-label' }, '权重'),
              h('input', { class: 'field-input', type: 'number', value: item.weight, disabled: !ctx.canEditObjectiveDimension(props.review), placeholder: '权重%', onInput: (event) => { item.weight = Number(readInputValue(event)) } })
            ]),
            h('view', { class: 'field-block' }, [
              h('text', { class: 'field-label' }, '完成度'),
              h('input', { class: 'field-input', type: 'number', value: item.completion, disabled: !props.editableSelf, placeholder: '完成%', onInput: (event) => { item.completion = readInputValue(event) } })
            ])
          ]),
          h('view', { class: 'field-block field-block-wide' }, [
            h('text', { class: 'field-label' }, '达成结果'),
            h('textarea', { class: 'field-textarea', value: item.result, disabled: !props.editableSelf, placeholder: '达成结果自评', onInput: (event) => { item.result = readInputValue(event) } })
          ])
        ])))
      ]),
      h('section', { class: 'form-section' }, [
        h('view', { class: 'section-title' }, [h('text', {}, '思考与总结')]),
        h('textarea', { class: 'field-textarea', value: props.review.selfSummary, disabled: !props.editableSelf, placeholder: '填写本月思考与总结', onInput: (event) => { props.review.selfSummary = readInputValue(event) } })
      ]),
      h(ValueSection, { review: props.review, field: 'self', title: '价值观自评', editable: props.editableSelf })
    ])
  }
}

const NextObjectiveSection = {
  props: ['review', 'editable'],
  setup(props) {
    return () => h('section', { class: 'form-section' }, [
      h('view', { class: 'section-title' }, [h('text', {}, '下月目标')]),
      h('view', { class: 'objective-list' }, props.review.nextObjectives.map((item, index) => h('view', { class: 'objective-card next' }, [
        h('view', { class: 'objective-head' }, [
          h('text', { class: 'objective-title' }, `下月目标 ${index + 1}`)
        ]),
        h('view', { class: 'field-block field-block-wide' }, [
          h('text', { class: 'field-label' }, '目标描述'),
          h('textarea', { class: 'field-textarea', value: item.target, disabled: !props.editable, placeholder: '下月绩效目标', onInput: (event) => { item.target = readInputValue(event) } })
        ]),
        h('view', { class: 'objective-fields objective-fields-next' }, [
          h('view', { class: 'field-block' }, [
            h('text', { class: 'field-label' }, '权重'),
            h('input', { class: 'field-input', type: 'number', value: item.weight, disabled: !props.editable, placeholder: '权重%', onInput: (event) => { item.weight = Number(readInputValue(event)) } })
          ])
        ])
      ])))
    ])
  }
}

const ValueSection = {
  props: ['review', 'field', 'title', 'editable'],
  setup(props) {
    const ctx = usePerformanceContext()

    return () => h('section', { class: 'form-section' }, [
      h('view', { class: 'section-title' }, [
        h('text', {}, props.title),
        h('text', { class: 'count-pill' }, `总分 ${ctx.valueTotal(props.review, props.field)}`)
      ]),
      h('view', { class: 'value-grid' }, props.review.values.map((item) => {
        const tpl = ctx.state.template?.values?.find((value) => value.id === item.id) || { name: item.id, definition: '' }
        return h('view', { class: 'value-card' }, [
          h('text', { class: 'value-name' }, tpl.name),
          h('text', { class: 'value-desc' }, tpl.definition),
          h('input', { class: 'field-input', type: 'number', value: item[props.field], disabled: !props.editable, placeholder: '0-50', onInput: (event) => { item[props.field] = readInputValue(event) } })
        ])
      }))
    ])
  }
}

const ManagerSection = {
  props: ['review', 'editable'],
  setup(props) {
    const ctx = usePerformanceContext()

    return () => h('section', { class: 'form-section' }, [
      h('view', { class: 'section-title' }, [h('text', {}, '上级评价')]),
      h('view', { class: 'form-grid' }, [
        h('select', { class: 'field-select', value: props.review.managerGrade, disabled: !props.editable, onChange: (event) => { props.review.managerGrade = event.target.value } }, ctx.gradeOptions().map((grade) => h('option', { value: grade }, grade || '未选择'))),
        h('textarea', { class: 'field-textarea', value: props.review.managerComment, disabled: !props.editable, placeholder: '填写上级评价', onInput: (event) => { props.review.managerComment = readInputValue(event) } })
      ]),
      h(ValueSection, { review: props.review, field: 'manager', title: '上级价值观评分', editable: props.editable })
    ])
  }
}

const HrbpSection = {
  props: ['review', 'editable'],
  setup(props) {
    const ctx = usePerformanceContext()

    return () => h('section', { class: 'form-section' }, [
      h('view', { class: 'section-title' }, [h('text', {}, 'HRBP评价')]),
      h('view', { class: 'form-grid' }, [
        h('select', { class: 'field-select', value: props.review.hrbpGrade, disabled: !props.editable, onChange: (event) => { props.review.hrbpGrade = event.target.value } }, ctx.gradeOptions().map((grade) => h('option', { value: grade }, grade || '未选择'))),
        h('textarea', { class: 'field-textarea', value: props.review.hrbpComment, disabled: !props.editable, placeholder: '填写 HRBP 评价', onInput: (event) => { props.review.hrbpComment = readInputValue(event) } })
      ]),
      props.review.managerGrade && props.review.hrbpGrade && props.review.managerGrade !== props.review.hrbpGrade
        ? h('view', { class: 'notice danger' }, `上级分档为 ${props.review.managerGrade}，HRBP分档为 ${props.review.hrbpGrade}，双方不一致时不能提交。`)
        : null,
      h(ValueSection, { review: props.review, field: 'hrbp', title: 'HRBP价值观评分', editable: props.editable })
    ])
  }
}

const EmployeeConfirmSection = {
  props: ['review', 'editable'],
  setup(props) {
    return () => h('section', { class: 'form-section' }, [
      h('view', { class: 'section-title' }, [
        h('text', {}, '员工确认'),
        h('text', { class: 'count-pill' }, props.review.employeeConfirmResult === 'confirmed' ? '已确认' : props.review.employeeConfirmResult === 'disputed' ? '有异议' : '待确认')
      ]),
      h('textarea', { class: 'field-textarea', value: props.review.employeeConfirmComment, disabled: !props.editable, placeholder: '确认可简单说明；如提出异议，请填写原因。', onInput: (event) => { props.review.employeeConfirmComment = readInputValue(event) } })
    ])
  }
}

const FinalSection = {
  props: ['review', 'editable'],
  setup(props) {
    const ctx = usePerformanceContext()

    return () => h('section', { class: 'form-section' }, [
      h('view', { class: 'section-title' }, [h('text', {}, 'HRBP归档')]),
      h('view', { class: 'form-grid' }, [
        h('select', { class: 'field-select', value: ctx.effectiveGrade(props.review), disabled: !props.editable, onChange: (event) => { props.review.finalGrade = event.target.value } }, ctx.gradeOptions().map((grade) => h('option', { value: grade }, grade || '未选择'))),
        h('textarea', { class: 'field-textarea', value: props.review.finalNote, disabled: !props.editable, placeholder: 'HRBP备注', onInput: (event) => { props.review.finalNote = readInputValue(event) } })
      ])
    ])
  }
}

const HistorySection = {
  props: ['review'],
  setup(props) {
    return () => h('section', { class: 'form-section' }, [
      h('view', { class: 'section-title' }, [h('text', {}, '流转记录')]),
      h('view', { class: 'history-list' }, (props.review.history || []).slice().reverse().map((item) => h('view', { class: 'history-row' }, [
        h('text', {}, item.action),
        h('text', { class: 'muted' }, `${item.by} · ${item.at ? new Date(item.at).toLocaleString() : ''}`)
      ])))
    ])
  }
}

export default WorkbenchView
