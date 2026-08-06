import { h, ref, Teleport } from 'vue'
import {
  addCurrentObjective,
  addNextObjective,
  confirmRemoveCurrentObjective,
  confirmRemoveNextObjective,
  hasHrbpEvaluation,
  hasManagerEvaluation,
  hrbpReviewTitleMeta,
  managerReviewTitleMeta,
  readInputValue,
  renderValueRubricList,
  reviewGradeBadge,
  textareaAutoHeightStyle,
  estimateTextareaRows,
  valueRubricItems
} from './reviewFormHelpers'
import { usePerformanceContext } from '../context'
import { historyActionParts } from '../reviewFlow'

const ValueStandardModal = {
  props: ['standard'],
  emits: ['close'],
  setup(props, { emit }) {
    return () => {
      const standard = props.standard || {}
      const rubrics = Array.isArray(standard.rubrics) ? standard.rubrics : []
      return h(Teleport, { to: 'body' }, [
        h('view', { class: 'value-standard-modal-mask', onClick: () => emit('close') }, [
          h('view', { class: 'value-standard-modal', onClick: (event) => event.stopPropagation() }, [
            h('view', { class: 'value-standard-modal-head' }, [
              h('view', { class: 'value-standard-title-copy' }, [
                h('text', { class: 'value-standard-modal-title' }, '评分标准'),
                h('text', { class: 'value-standard-modal-subtitle' }, standard.name || '-')
              ]),
              h('button', { class: 'process-modal-close', onClick: () => emit('close') }, '×')
            ]),
            standard.definition ? h('text', { class: 'value-standard-definition' }, standard.definition) : null,
            h('view', { class: 'value-score-guide value-standard-guide' }, [
              renderValueRubricList(rubrics)
            ]),
            h('view', { class: 'value-standard-actions' }, [
              h('button', { class: 'dt-btn dt-btn-primary', onClick: () => emit('close') }, '知道了')
            ])
          ])
        ])
      ])
    }
  }
}

export const CurrentObjectiveSection = {
  props: ['review', 'editableSelf', 'readonly'],
  setup(props) {
    const ctx = usePerformanceContext()
    const showDeleteActions = ref(false)

    function toggleDeleteActions() {
      showDeleteActions.value = !showDeleteActions.value
    }

    return () => {
      const editableObjectives = !props.readonly && ctx.canEditObjectiveDimension(props.review)
      const objectives = Array.isArray(props.review.objectives) ? props.review.objectives : []

      return h('view', {}, [
        h('section', { class: 'form-section' }, [
          h('view', { class: 'section-title' }, [
            h('view', { class: 'section-title-main' }, [
              h('text', {}, '本月目标'),
              h('text', { class: 'count-pill' }, `合计 ${ctx.totalObjectiveScore(props.review)}`)
            ]),
            editableObjectives ? h('view', { class: 'section-actions' }, [
              h('button', { class: 'dt-btn dt-btn-light small', onClick: () => addCurrentObjective(props.review) }, '增加目标'),
              h('button', {
                class: ['dt-btn small objective-delete-toggle', showDeleteActions.value ? 'active' : 'dt-btn-light'],
                onClick: toggleDeleteActions
              }, showDeleteActions.value ? '隐藏删除' : '显示删除')
            ]) : null
          ]),
          objectives.length
            ? h('view', { class: 'objective-list' }, objectives.map((item, index) => h('view', { class: 'objective-card', key: `objective-${index}` }, [
              h('view', { class: 'objective-head' }, [
                h('text', { class: 'objective-title' }, `目标 ${index + 1}`),
                h('view', { class: 'objective-head-actions' }, [
                  h('text', { class: 'score-badge' }, `${ctx.objectiveScore(item)} 分`),
                  showDeleteActions.value && editableObjectives ? h('button', {
                    class: 'dt-btn dt-btn-danger-light small objective-delete-btn',
                    onClick: () => confirmRemoveCurrentObjective(props.review, index, () => {
                      if (!props.review.objectives?.length) showDeleteActions.value = false
                    })
                  }, '删除') : null
                ])
              ]),
              h('view', { class: 'field-block field-block-wide' }, [
                h('text', { class: 'field-label' }, '目标描述'),
                h('textarea', { class: 'field-textarea', value: item.target, disabled: !editableObjectives, placeholder: '绩效目标', onInput: (event) => { item.target = readInputValue(event) } })
              ]),
              h('view', { class: 'objective-fields' }, [
                h('view', { class: 'field-block' }, [
                  h('text', { class: 'field-label' }, '权重'),
                  h('input', {
                    class: 'field-input',
                    type: 'number',
                    min: 0,
                    max: 100,
                    step: 1,
                    value: item.weight,
                    disabled: !editableObjectives,
                    placeholder: '权重%',
                    onInput: (event) => { item.weight = Number(readInputValue(event)) }
                  })
                ]),
                h('view', { class: 'field-block' }, [
                  h('text', { class: 'field-label' }, '完成度'),
                  h('input', {
                    class: 'field-input',
                    type: 'number',
                    min: 0,
                    max: 100,
                    step: 1,
                    value: item.completion,
                    disabled: !props.editableSelf,
                    placeholder: '完成%',
                    onInput: (event) => { item.completion = readInputValue(event) }
                  })
                ])
              ]),
              h('view', { class: 'field-block field-block-wide' }, [
                h('text', { class: 'field-label' }, '达成结果'),
                h('textarea', { class: 'field-textarea', value: item.result, disabled: !props.editableSelf, placeholder: '达成结果自评', onInput: (event) => { item.result = readInputValue(event) } })
              ])
            ])))
            : h('view', { class: 'objective-empty' }, editableObjectives ? '暂无目标，点击增加目标开始填写' : '暂无目标')
        ])
      ])
    }
  }
}

export const SelfSummarySection = {
  props: ['review', 'editableSelf'],
  setup(props) {
    return () => h('section', { class: 'form-section' }, [
      h('view', { class: 'section-title' }, [h('text', {}, '思考总结')]),
      h('textarea', {
        class: 'field-textarea self-summary-textarea',
        value: props.review.selfSummary,
        rows: estimateTextareaRows(props.review.selfSummary),
        style: textareaAutoHeightStyle(props.review.selfSummary),
        disabled: !props.editableSelf,
        placeholder: '填写思考总结',
        onInput: (event) => { props.review.selfSummary = readInputValue(event) }
      })
    ])
  }
}

export const NextObjectiveSection = {
  props: ['review', 'editable'],
  setup(props) {
    const ctx = usePerformanceContext()
    const showDeleteActions = ref(false)

    function toggleDeleteActions() {
      showDeleteActions.value = !showDeleteActions.value
    }

    return () => {
      const nextObjectives = Array.isArray(props.review.nextObjectives) ? props.review.nextObjectives : []
      const editableNextObjectives = props.editable && ctx.canEditNextObjectives(props.review)
      const canAddNextObjective = props.editable && ctx.canAddNextObjective(props.review)
      const canDeleteNextObjective = props.editable && ctx.canDeleteNextObjective(props.review)

      return h('section', { class: 'form-section' }, [
        h('view', { class: 'section-title' }, [
          h('view', { class: 'section-title-main' }, [
            h('text', {}, '下月目标'),
            h('text', { class: 'count-pill' }, `合计 ${nextObjectives.reduce((total, item) => total + (Number(item.weight) || 0), 0)}`)
          ]),
          canAddNextObjective || canDeleteNextObjective ? h('view', { class: 'section-actions' }, [
            canAddNextObjective ? h('button', { class: 'dt-btn dt-btn-light small', onClick: () => addNextObjective(props.review) }, '增加目标') : null,
            canDeleteNextObjective ? h('button', {
              class: ['dt-btn small objective-delete-toggle', showDeleteActions.value ? 'active' : 'dt-btn-light'],
              onClick: toggleDeleteActions
            }, showDeleteActions.value ? '隐藏删除' : '显示删除') : null
          ]) : null
        ]),
        nextObjectives.length
          ? h('view', { class: 'objective-list' }, nextObjectives.map((item, index) => h('view', { class: 'objective-card next', key: `next-objective-${index}` }, [
            h('view', { class: 'objective-head' }, [
              h('text', { class: 'objective-title' }, `下月目标 ${index + 1}`),
              showDeleteActions.value && canDeleteNextObjective ? h('button', {
                class: 'dt-btn dt-btn-danger-light small objective-delete-btn',
                onClick: () => confirmRemoveNextObjective(props.review, index, () => {
                  if (!props.review.nextObjectives?.length) showDeleteActions.value = false
                })
              }, '删除') : null
            ]),
            h('view', { class: 'field-block field-block-wide' }, [
              h('text', { class: 'field-label' }, '目标描述'),
              h('textarea', { class: 'field-textarea', value: item.target, disabled: !editableNextObjectives, placeholder: '下月绩效目标', onInput: (event) => { item.target = readInputValue(event) } })
            ]),
            h('view', { class: 'objective-fields objective-fields-next' }, [
              h('view', { class: 'field-block' }, [
                h('text', { class: 'field-label' }, '权重'),
                h('input', {
                  class: 'field-input',
                  type: 'number',
                  min: 0,
                  max: 100,
                  step: 1,
                  value: item.weight,
                  disabled: !editableNextObjectives,
                  placeholder: '权重%',
                  onInput: (event) => { item.weight = Number(readInputValue(event)) }
                })
              ])
            ])
          ])))
          : h('view', { class: 'objective-empty next-objective-empty' }, [
            h('view', { class: 'next-objective-empty-mark' }, '+'),
            h('text', { class: 'next-objective-empty-title' }, '暂无下月目标'),
            h('text', { class: 'next-objective-empty-desc' }, canAddNextObjective ? '点击增加目标，填写下月计划和权重' : '当前暂无可查看的下月目标'),
            canAddNextObjective ? h('button', { class: 'dt-btn dt-btn-primary small next-objective-empty-action', onClick: () => addNextObjective(props.review) }, '增加目标') : null
          ])
      ])
    }
  }
}

export const ValueSection = {
  props: ['review', 'field', 'title', 'editable'],
  setup(props) {
    const ctx = usePerformanceContext()
    const activeStandard = ref(null)

    function openStandard(event, payload) {
      event?.stopPropagation?.()
      activeStandard.value = payload
    }

    return () => h('section', { class: 'form-section' }, [
      h('view', { class: 'section-title' }, [
        h('text', {}, props.title),
        h('text', { class: 'count-pill' }, `总分 ${ctx.valueTotal(props.review, props.field)}`)
      ]),
      h('view', { class: 'value-grid value-list' }, props.review.values.map((item) => {
        const tpl = ctx.state.template?.values?.find((value) => value.id === item.id) || {}
        const valueName = tpl.name || item.name || item.id
        const valueDefinition = tpl.definition || item.definition || ''
        const rubrics = valueRubricItems(tpl, item)
        return h('view', { class: 'value-card' }, [
          h('view', { class: 'value-title-row' }, [
            h('text', { class: 'value-name' }, valueName),
            rubrics.length ? h('button', {
              class: 'value-score-tag',
              onClick: (event) => openStandard(event, {
                name: valueName,
                definition: valueDefinition,
                rubrics
              })
            }, '查看评分标准') : null
          ]),
          valueDefinition ? h('text', { class: 'value-desc' }, valueDefinition) : null,
          h('input', { class: 'field-input', type: 'number', value: item[props.field], disabled: !props.editable, placeholder: '0-50', onInput: (event) => { item[props.field] = readInputValue(event) } })
        ])
      })),
      activeStandard.value ? h(ValueStandardModal, { standard: activeStandard.value, onClose: () => { activeStandard.value = null } }) : null
    ])
  }
}

export const ManagerSection = {
  props: ['review', 'editable'],
  setup(props) {
    const ctx = usePerformanceContext()

    return () => {
      const managerPendingText = '上级待评'
      const managerPending = !hasManagerEvaluation(props.review)

      return h('section', { class: 'form-section review-evaluation-section manager-review-section' }, [
        h('view', { class: 'section-title' }, [
          h('view', { class: 'section-title-main' }, [
            h('text', {}, '上级评价'),
            h('text', { class: 'section-meta' }, managerReviewTitleMeta(ctx, props.review))
          ]),
          reviewGradeBadge('上级分档', props.review.managerGrade)
        ]),
        h('view', { class: props.editable ? 'form-grid manager-review-grid manager-review-block' : 'manager-review-readonly manager-review-block' }, [
          props.editable ? h('view', { class: 'field-block manager-grade-field' }, [
            h('text', { class: 'field-label' }, '上级分档'),
            h('select', { class: 'field-select', value: props.review.managerGrade, onChange: (event) => { props.review.managerGrade = event.target.value } }, ctx.gradeOptions().map((grade) => h('option', { value: grade }, grade || '未选择')))
          ]) : null,
          h('view', { class: 'field-block field-block-wide manager-comment-field' }, [
            props.editable ? h('text', { class: 'field-label' }, '评价内容') : null,
            h('textarea', { class: 'field-textarea', value: props.review.managerComment, disabled: !props.editable, placeholder: managerPending && !props.editable ? managerPendingText : '填写上级评价', onInput: (event) => { props.review.managerComment = readInputValue(event) } })
          ])
        ]),
        h(ValueSection, { review: props.review, field: 'manager', title: '上级价值观评分', editable: props.editable })
      ])
    }
  }
}

export const HrbpSection = {
  props: ['review', 'editable'],
  setup(props) {
    const ctx = usePerformanceContext()

    return () => {
      const hrbpPendingText = 'HRBP待评'
      const hrbpPending = !hasHrbpEvaluation(props.review)
      const gradeMismatch = props.review.managerGrade && props.review.hrbpGrade && props.review.managerGrade !== props.review.hrbpGrade
      const gradeMismatchNotice = () => gradeMismatch
        ? h('view', { class: 'notice danger hrbp-grade-notice' }, `上级分档为 ${props.review.managerGrade}，HRBP分档为 ${props.review.hrbpGrade}，双方不一致时不能提交。`)
        : null

      return h('section', { class: 'form-section review-evaluation-section hrbp-review-section' }, [
        h('view', { class: 'section-title' }, [
          h('view', { class: 'section-title-main' }, [
            h('text', {}, 'HRBP评价'),
            h('text', { class: 'section-meta' }, hrbpReviewTitleMeta(ctx, props.review))
          ]),
          reviewGradeBadge('HRBP分档', props.review.hrbpGrade)
        ]),
        h('view', { class: props.editable ? 'form-grid hrbp-review-grid hrbp-review-block' : 'manager-review-readonly hrbp-review-block' }, [
          props.editable ? h('view', { class: 'field-block hrbp-grade-field' }, [
            h('text', { class: 'field-label' }, 'HRBP分档'),
            h('view', { class: 'hrbp-grade-row' }, [
              h('select', { class: 'field-select', value: props.review.hrbpGrade, onChange: (event) => { props.review.hrbpGrade = event.target.value } }, ctx.gradeOptions().map((grade) => h('option', { value: grade }, grade || '未选择'))),
              gradeMismatchNotice()
            ])
          ]) : null,
          h('view', { class: 'field-block field-block-wide hrbp-comment-field' }, [
            props.editable ? h('text', { class: 'field-label' }, '评价内容') : null,
            h('textarea', { class: 'field-textarea', value: props.review.hrbpComment, disabled: !props.editable, placeholder: hrbpPending && !props.editable ? hrbpPendingText : '填写 HRBP 评价', onInput: (event) => { props.review.hrbpComment = readInputValue(event) } })
          ])
        ]),
        !props.editable ? gradeMismatchNotice() : null,
        h(ValueSection, { review: props.review, field: 'hrbp', title: 'HRBP价值观评分', editable: props.editable })
      ])
    }
  }
}

export const EmployeeConfirmSection = {
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

export const FinalSection = {
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

export const HistorySection = {
  props: ['review'],
  setup(props) {
    return () => h('section', { class: 'form-section' }, [
      h('view', { class: 'section-title' }, [h('text', {}, '流转记录')]),
      h('view', { class: 'history-list' }, (props.review.history || []).slice().reverse().map((item) => {
        const actionParts = historyActionParts(item.action)
        return h('view', { class: 'history-row' }, [
          h('text', {}, actionParts.title || item.action),
          actionParts.reason ? h('text', { class: 'history-row-reason' }, `理由：${actionParts.reason}`) : null,
          h('text', { class: 'muted' }, `${item.by} · ${item.at ? new Date(item.at).toLocaleString() : ''}`)
        ])
      }))
    ])
  }
}
