import { h } from 'vue'
import { currentAssignee } from './reviewFlow'

const HISTORY_TABLE_COLUMNS = [
  { key: 'period', label: '考评月份', mobile: true },
  { key: 'status', label: '状态', mobile: true },
  { key: 'objectiveScore', label: '目标得分' },
  { key: 'managerGrade', label: '上级分档' },
  { key: 'hrbpGrade', label: 'HRBP分档' },
  { key: 'finalGrade', label: '最终分档', mobile: true },
  { key: 'actions', label: '操作', mobile: true }
]

const HRBP_REVIEW_TABLE_COLUMNS = [
  { key: 'employee', label: '员工', mobile: true },
  { key: 'department', label: '部门' },
  { key: 'position', label: '岗位' },
  { key: 'period', label: '考评月份', mobile: true },
  { key: 'nextPeriod', label: '目标月份' },
  { key: 'status', label: '状态', mobile: true },
  { key: 'objectiveScore', label: '目标得分' },
  { key: 'managerGrade', label: '上级分档' },
  { key: 'hrbpGrade', label: 'HRBP分档' },
  { key: 'finalGrade', label: '最终分档' },
  { key: 'actions', label: '操作' }
]

const MANAGER_REVIEW_TABLE_COLUMNS = [
  { key: 'employee', label: '员工', mobile: true },
  { key: 'position', label: '岗位' },
  { key: 'period', label: '考评月份', mobile: true },
  { key: 'status', label: '状态', mobile: true },
  { key: 'objectiveScore', label: '目标得分' },
  { key: 'managerGrade', label: '上级分档' },
  { key: 'hrbpGrade', label: 'HRBP分档' },
  { key: 'finalGrade', label: '最终分档' },
  { key: 'handler', label: '当前处理人' },
  { key: 'actions', label: '操作' }
]

function readInputValue(event) {
  return event.detail?.value ?? event.target.value
}

function historyColumnClass(column) {
  return [
    'history-col',
    `history-col-${column.key}`,
    column.mobile ? 'history-mobile-visible' : 'history-mobile-hidden',
    column.key === 'actions' ? 'table-actions-cell' : ''
  ]
}

function reviewColumnClass(prefix, column) {
  return [
    `${prefix}-col`,
    `${prefix}-col-${column.key}`,
    column.mobile ? `${prefix}-mobile-visible` : `${prefix}-mobile-hidden`,
    column.key === 'actions' ? 'table-actions-cell' : ''
  ]
}

function employeePosition(ctx, review) {
  const user = ctx.state.users.find((item) => item.id === review.employeeId)
  return user?.position || '-'
}

function renderHistoryCell(ctx, column, review, onOpenRow) {
  switch (column.key) {
    case 'period':
      return review.period || '-'
    case 'status':
      return h('text', { class: ctx.reviewStatusClass(review) }, ctx.reviewStatusText(review))
    case 'objectiveScore':
      return String(ctx.totalObjectiveScore(review))
    case 'managerGrade':
      return review.managerGrade || '-'
    case 'hrbpGrade':
      return review.hrbpGrade || '-'
    case 'finalGrade':
      return ctx.effectiveGrade(review) || '-'
    case 'actions':
      return h('button', {
        class: 'dt-btn dt-btn-light small table-action-btn history-action-btn',
        onClick: (event) => {
          event?.stopPropagation?.()
          onOpenRow(review)
        }
      }, '查看')
    default:
      return '-'
  }
}

function renderHrbpReviewCell(ctx, column, review, onOpenRow) {
  switch (column.key) {
    case 'employee':
      return ctx.userName(review.employeeId)
    case 'department':
      return review.department || '-'
    case 'position':
      return employeePosition(ctx, review)
    case 'period':
      return review.period || '-'
    case 'nextPeriod':
      return review.nextPeriod || '-'
    case 'status':
      return h('text', { class: ctx.reviewStatusClass(review) }, ctx.reviewStatusText(review))
    case 'objectiveScore':
      return String(ctx.totalObjectiveScore(review))
    case 'managerGrade':
      return review.managerGrade || '-'
    case 'hrbpGrade':
      return review.hrbpGrade || '-'
    case 'finalGrade':
      return ctx.effectiveGrade(review) || '-'
    case 'actions':
      return h('button', {
        class: ['dt-btn', ctx.canHrbpHandle(review) ? 'dt-btn-primary' : 'dt-btn-light', 'small', 'table-action-btn'],
        onClick: (event) => {
          event?.stopPropagation?.()
          onOpenRow(review)
        }
      }, ctx.canHrbpHandle(review) ? '评价' : '查看')
    default:
      return '-'
  }
}

function renderManagerReviewCell(ctx, column, review, onOpenRow) {
  switch (column.key) {
    case 'employee':
      return ctx.userName(review.employeeId)
    case 'position':
      return employeePosition(ctx, review)
    case 'period':
      return review.period || '-'
    case 'status':
      return h('text', { class: ctx.reviewStatusClass(review) }, ctx.reviewStatusText(review))
    case 'objectiveScore':
      return String(ctx.totalObjectiveScore(review))
    case 'managerGrade':
      return review.managerGrade || '-'
    case 'hrbpGrade':
      return review.hrbpGrade || '-'
    case 'finalGrade':
      return ctx.effectiveGrade(review) || '-'
    case 'handler':
      return currentAssignee(ctx, review)
    case 'actions':
      return h('button', {
        class: ['dt-btn', ctx.canManager(review) ? 'dt-btn-primary' : 'dt-btn-light', 'small', 'table-action-btn'],
        onClick: (event) => {
          event?.stopPropagation?.()
          onOpenRow(review)
        }
      }, ctx.canManager(review) ? '评价' : '查看')
    default:
      return '-'
  }
}

export function reviewMatchesHrbpReviewedFilters(ctx, review) {
  const employeeName = String(ctx.hrbpReviewedFilters.employeeName || '').trim().toLowerCase()
  if (employeeName) {
    const employeeText = [
      ctx.userName(review.employeeId),
      review.employeeName,
      review.employeeId
    ].filter(Boolean).join(' ').toLowerCase()
    if (!employeeText.includes(employeeName)) return false
  }
  const period = String(ctx.hrbpReviewedFilters.period || '').trim()
  if (period && review.period !== period) return false
  const grade = String(ctx.hrbpReviewedFilters.grade || '').trim()
  if (grade && ctx.effectiveGrade(review) !== grade) return false
  return true
}

export function reviewMatchesManagerReviewedFilters(ctx, review) {
  const employeeName = String(ctx.managerReviewedFilters.employeeName || '').trim().toLowerCase()
  if (employeeName) {
    const employeeText = [
      ctx.userName(review.employeeId),
      review.employeeName,
      review.employeeId
    ].filter(Boolean).join(' ').toLowerCase()
    if (!employeeText.includes(employeeName)) return false
  }
  const period = String(ctx.managerReviewedFilters.period || '').trim()
  if (period && review.period !== period) return false
  const objectiveScore = String(ctx.managerReviewedFilters.objectiveScore || '').trim()
  if (objectiveScore) {
    const expectedScore = Number(objectiveScore)
    const actualScore = Number(ctx.totalObjectiveScore(review))
    if (!Number.isFinite(expectedScore) || !Number.isFinite(actualScore)) return false
    if (Math.abs(actualScore - expectedScore) > 0.0001) return false
  }
  return true
}

export function renderHistoryFilters(ctx, collapsed, selectedCount, onToggle) {
  return h('view', {
    class: ['summary-filter-shell', 'history-filter-shell', collapsed ? 'collapsed' : 'expanded']
  }, [
    h('button', { class: 'summary-filter-toggle', onClick: onToggle }, [
      h('text', { class: 'summary-filter-toggle-title' }, '筛选条件'),
      selectedCount > 0 ? h('text', { class: 'summary-filter-count' }, `已选 ${selectedCount}`) : null,
      h('text', { class: ['summary-filter-arrow', collapsed ? '' : 'expanded'] })
    ]),
    h('view', { class: 'filters summary-filters history-filters' }, [
      h('select', {
        class: 'field-select',
        value: ctx.historyFilters.year,
        onChange: (event) => ctx.setHistoryFilter('year', readInputValue(event))
      }, [
        h('option', { value: '' }, '全部年份'),
        ...ctx.historyYearOptions.value.map((year) => h('option', { value: year }, `${year}年`))
      ]),
      h('select', {
        class: 'field-select',
        value: ctx.historyFilters.month,
        onChange: (event) => ctx.setHistoryFilter('month', readInputValue(event))
      }, [
        h('option', { value: '' }, '全部月份'),
        ...ctx.historyMonthOptions.map((month) => h('option', { value: month }, `${Number(month)}月`))
      ]),
      h('select', {
        class: 'field-select',
        value: ctx.historyFilters.grade,
        onChange: (event) => ctx.setHistoryFilter('grade', readInputValue(event))
      }, [
        h('option', { value: '' }, '全部分档'),
        ...ctx.grades.value.map((grade) => h('option', { value: grade }, grade))
      ]),
      h('button', { class: 'dt-btn dt-btn-light history-filter-reset-btn', onClick: ctx.resetHistoryFilters }, '重置')
    ])
  ])
}

export function renderHrbpReviewedFilters(ctx, collapsed, selectedCount, onToggle) {
  if (ctx.hrbpReviewTab.value !== 'reviewed') return null
  return h('view', {
    class: ['summary-filter-shell', 'hrbp-reviewed-filter-shell', collapsed ? 'collapsed' : 'expanded']
  }, [
    h('button', { class: 'summary-filter-toggle', onClick: onToggle }, [
      h('text', { class: 'summary-filter-toggle-title' }, '筛选条件'),
      selectedCount > 0 ? h('text', { class: 'summary-filter-count' }, `已选 ${selectedCount}`) : null,
      h('text', { class: ['summary-filter-arrow', collapsed ? '' : 'expanded'] })
    ]),
    h('view', { class: 'filters summary-filters hrbp-reviewed-filters' }, [
      h('input', {
        class: 'field-input',
        value: ctx.hrbpReviewedFilters.employeeName,
        placeholder: '员工姓名/账号',
        onInput: (event) => ctx.setHrbpReviewedFilter('employeeName', readInputValue(event))
      }),
      h('input', {
        class: 'field-input',
        type: 'month',
        value: ctx.hrbpReviewedFilters.period,
        placeholder: '年月份',
        onInput: (event) => ctx.setHrbpReviewedFilter('period', readInputValue(event)),
        onChange: (event) => ctx.setHrbpReviewedFilter('period', readInputValue(event))
      }),
      h('select', {
        class: 'field-select',
        value: ctx.hrbpReviewedFilters.grade,
        onChange: (event) => ctx.setHrbpReviewedFilter('grade', readInputValue(event))
      }, [
        h('option', { value: '' }, '全部最终分档'),
        ...ctx.grades.value.map((grade) => h('option', { value: grade }, grade))
      ]),
      h('button', { class: 'dt-btn dt-btn-primary hrbp-reviewed-search-btn', onClick: ctx.searchHrbpReviewedReviews }, '查询'),
      h('button', { class: 'dt-btn dt-btn-light hrbp-reviewed-reset-btn', onClick: ctx.resetHrbpReviewedFilters }, '重置')
    ])
  ])
}

export function renderManagerReviewedFilters(ctx, collapsed, selectedCount, onToggle) {
  if (ctx.managerReviewTab.value !== 'reviewed') return null
  return h('view', {
    class: ['summary-filter-shell', 'manager-reviewed-filter-shell', collapsed ? 'collapsed' : 'expanded']
  }, [
    h('button', { class: 'summary-filter-toggle', onClick: onToggle }, [
      h('text', { class: 'summary-filter-toggle-title' }, '筛选条件'),
      selectedCount > 0 ? h('text', { class: 'summary-filter-count' }, `已选 ${selectedCount}`) : null,
      h('text', { class: ['summary-filter-arrow', collapsed ? '' : 'expanded'] })
    ]),
    h('view', { class: 'filters summary-filters manager-reviewed-filters' }, [
      h('input', {
        class: 'field-input',
        value: ctx.managerReviewedFilters.employeeName,
        placeholder: '员工姓名/账号',
        onInput: (event) => ctx.setManagerReviewedFilter('employeeName', readInputValue(event))
      }),
      h('input', {
        class: 'field-input',
        type: 'month',
        value: ctx.managerReviewedFilters.period,
        placeholder: '年月份',
        onInput: (event) => ctx.setManagerReviewedFilter('period', readInputValue(event)),
        onChange: (event) => ctx.setManagerReviewedFilter('period', readInputValue(event))
      }),
      h('input', {
        class: 'field-input',
        type: 'number',
        value: ctx.managerReviewedFilters.objectiveScore,
        placeholder: '最终得分',
        onInput: (event) => ctx.setManagerReviewedFilter('objectiveScore', readInputValue(event))
      }),
      h('button', { class: 'dt-btn dt-btn-primary manager-reviewed-search-btn', onClick: ctx.searchManagerReviewedReviews }, '查询'),
      h('button', { class: 'dt-btn dt-btn-light manager-reviewed-reset-btn', onClick: ctx.resetManagerReviewedFilters }, '重置')
    ])
  ])
}

export function renderHistoryTable(ctx, rows, onOpenRow) {
  return h('view', { class: 'history-table-wrap table-wrap' }, [
    h('table', { class: 'history-performance-table summary-table' }, [
      h('thead', {}, h('tr', {}, HISTORY_TABLE_COLUMNS.map((column) =>
        h('th', { class: historyColumnClass(column) }, column.label)
      ))),
      h('tbody', {}, rows.length
        ? rows.map((review) => h('tr', { class: 'history-clickable-row', onClick: () => onOpenRow(review) }, HISTORY_TABLE_COLUMNS.map((column) =>
            h('td', { class: historyColumnClass(column) }, renderHistoryCell(ctx, column, review, onOpenRow))
          )))
        : h('tr', {}, [
            h('td', { class: 'history-empty-row', colspan: HISTORY_TABLE_COLUMNS.length }, '当前没有历史记录')
          ]))
    ])
  ])
}

export function renderHrbpReviewTable(ctx, rows, onOpenRow) {
  return h('view', { class: 'hrbp-table-wrap table-wrap' }, [
    h('table', { class: 'hrbp-review-table summary-table' }, [
      h('thead', {}, h('tr', {}, HRBP_REVIEW_TABLE_COLUMNS.map((column) =>
        h('th', { class: reviewColumnClass('hrbp-review', column) }, column.label)
      ))),
      h('tbody', {}, rows.length
        ? rows.map((review) => h('tr', { class: 'hrbp-review-clickable-row', onClick: () => onOpenRow(review) }, HRBP_REVIEW_TABLE_COLUMNS.map((column) =>
            h('td', { class: reviewColumnClass('hrbp-review', column) }, renderHrbpReviewCell(ctx, column, review, onOpenRow))
          )))
        : h('tr', {}, [
            h('td', { class: 'hrbp-empty-row', colspan: HRBP_REVIEW_TABLE_COLUMNS.length }, ctx.hrbpReviewTab.value === 'pending' ? '当前没有待评记录' : '当前没有已评记录')
          ]))
    ])
  ])
}

export function renderManagerReviewTable(ctx, rows, onOpenRow) {
  return h('view', { class: 'hrbp-table-wrap table-wrap' }, [
    h('table', { class: 'hrbp-review-table summary-table' }, [
      h('thead', {}, h('tr', {}, MANAGER_REVIEW_TABLE_COLUMNS.map((column) =>
        h('th', { class: reviewColumnClass('manager-review', column) }, column.label)
      ))),
      h('tbody', {}, rows.length
        ? rows.map((review) => h('tr', { class: 'manager-review-clickable-row', onClick: () => onOpenRow(review) }, MANAGER_REVIEW_TABLE_COLUMNS.map((column) =>
            h('td', { class: reviewColumnClass('manager-review', column) }, renderManagerReviewCell(ctx, column, review, onOpenRow))
          )))
        : h('tr', {}, [
            h('td', { class: 'hrbp-empty-row', colspan: MANAGER_REVIEW_TABLE_COLUMNS.length }, ctx.managerReviewTab.value === 'pending' ? '当前没有待评记录' : '当前没有已评记录')
          ]))
    ])
  ])
}
