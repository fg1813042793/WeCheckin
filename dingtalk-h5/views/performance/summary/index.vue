<script>
import { computed, h, ref, Teleport, watch } from 'vue'
import { usePerformanceContext } from '../common/context'
import { HistoryReviewDetailModal } from '../common/components/ReviewDetailModal'
import { renderTablePagination, useTablePagination } from '../common/tablePagination'

const SUMMARY_TABLE_COLUMNS = [
  { key: 'employee', label: '员工', mobile: true },
  { key: 'department', label: '部门' },
  { key: 'position', label: '岗位' },
  { key: 'period', label: '考评月份', mobile: true },
  { key: 'status', label: '状态' },
  { key: 'objectiveScore', label: '目标得分' },
  { key: 'managerGrade', label: '上级分档' },
  { key: 'hrbpGrade', label: 'HRBP分档' },
  { key: 'employeeConfirm', label: '员工确认' },
  { key: 'finalGrade', label: '最终分档', mobile: true },
  { key: 'actions', label: '操作' }
]

export default {
  name: 'SummaryPage',
  setup() {
    const ctx = usePerformanceContext()
    const departmentDropdownOpen = ref(false)
    const departmentSearchKeyword = ref('')
    const summaryFiltersCollapsed = ref(true)
    const summaryDepartmentExpandedKeys = ref(new Set())
    const summaryDetailReview = ref(null)
    const summaryDepartmentTree = computed(() => buildSummaryDepartmentTree(ctx.state.users))
    const summaryDepartmentRows = computed(() => flattenSummaryDepartmentTree(filterSummaryDepartmentTree(summaryDepartmentTree.value, departmentSearchKeyword.value), summaryDepartmentExpandedKeys.value, departmentSearchKeyword.value))
    const selectedDepartmentLabels = computed(() => Array.isArray(ctx.summaryFilters.departmentNames) ? ctx.summaryFilters.departmentNames : [])
    const summaryTableRows = computed(() => ctx.summaryReviews.value)
    const summaryPagination = useTablePagination(summaryTableRows)
    const summaryFilterCount = computed(() => {
      return [
        String(ctx.summaryFilters.employeeName || '').trim(),
        selectedDepartmentLabels.value.length > 0,
        ctx.summaryFilters.period
      ].filter(Boolean).length
    })

    watch(summaryDepartmentTree, (nodes) => {
      const availableKeys = new Set(collectSummaryDepartmentKeys(nodes))
      const next = new Set()
      for (const key of summaryDepartmentExpandedKeys.value) {
        if (availableKeys.has(key)) next.add(key)
      }
      summaryDepartmentExpandedKeys.value = next
    }, { immediate: true })

    function renderDepartmentFilter() {
      const labels = selectedDepartmentLabels.value
      return h('view', { class: 'summary-department-filter' }, [
        h('button', {
          class: ['summary-department-trigger', departmentDropdownOpen.value ? 'active' : ''],
          onClick: () => {
            departmentDropdownOpen.value = !departmentDropdownOpen.value
          }
        }, [
          h('text', { class: ['summary-department-trigger-text', labels.length > 0 ? 'selected' : ''] }, labels.length > 0 ? labels.join('、') : '部门名称'),
          labels.length > 0 ? h('text', { class: 'summary-department-selected-count' }, `${labels.length} 项`) : null,
          h('text', { class: ['summary-department-arrow', departmentDropdownOpen.value ? 'expanded' : ''] })
        ]),
        departmentDropdownOpen.value ? h('view', { class: 'summary-department-panel' }, [
          h('input', {
            class: 'summary-department-search',
            value: departmentSearchKeyword.value,
            placeholder: '搜索部门名称',
            onInput: (event) => {
              departmentSearchKeyword.value = event.detail?.value ?? event.target.value
            }
          }),
          h('view', { class: 'summary-department-tree' }, summaryDepartmentRows.value.length > 0
            ? summaryDepartmentRows.value.map((row) => renderDepartmentRow(row))
            : h('view', { class: 'summary-department-empty' }, '暂无匹配部门')
          ),
          h('view', { class: 'summary-department-actions' }, [
            h('button', { class: 'dt-btn dt-btn-light small', onClick: clearSummaryDepartments }, '清空'),
            h('button', { class: 'dt-btn dt-btn-primary small', onClick: () => { departmentDropdownOpen.value = false } }, '确定')
          ])
        ]) : null
      ])
    }

    function renderSummaryPeriodFilter() {
      return h('picker', {
        class: 'summary-month-picker',
        mode: 'date',
        fields: 'month',
        value: ctx.summaryFilters.period,
        onChange: (event) => {
          ctx.summaryFilters.period = event.detail?.value || ''
        }
      }, [
        h('view', { class: ['field-input', 'summary-month-value', ctx.summaryFilters.period ? 'selected' : ''] }, [
          h('text', { class: 'summary-month-text' }, ctx.summaryFilters.period || '考评月份'),
          h('text', { class: 'summary-month-arrow' })
        ])
      ])
    }

    function renderDepartmentRow(row) {
      const state = summaryDepartmentCheckState(row)
      return h('view', {
        class: ['summary-department-row', `depth-${row.depth}`, row.expanded ? 'expanded' : 'collapsed']
      }, [
        h('button', {
          class: 'summary-department-node',
          onClick: () => toggleSummaryDepartmentExpand(row)
        }, [
          h('text', { class: ['summary-department-chevron', row.expandable ? (row.expanded ? 'expanded' : 'collapsed') : 'placeholder'] }),
          h('text', { class: 'summary-department-name' }, row.name),
          h('text', { class: 'summary-department-count' }, `${row.count} 人`)
        ]),
        h('button', {
          class: [
            'summary-department-check',
            state === 'checked' ? 'checked' : '',
            state === 'indeterminate' ? 'summary-department-check-indeterminate' : ''
          ],
          onClick: (event) => {
            event?.stopPropagation?.()
            toggleSummaryDepartment(row)
          }
        }, state === 'checked' ? '✓' : state === 'indeterminate' ? '-' : '')
      ])
    }

    function toggleSummaryFilters() {
      summaryFiltersCollapsed.value = !summaryFiltersCollapsed.value
      if (summaryFiltersCollapsed.value) {
        departmentDropdownOpen.value = false
      }
    }

    function summaryColumnClass(column) {
      return ['summary-col', `summary-col-${column.key}`, column.mobile ? 'summary-mobile-visible' : 'summary-mobile-hidden']
    }

    async function handleDeleteReview(event, review) {
      event?.stopPropagation?.()
      await ctx.deleteReview(review.id)
    }

    function handleViewReview(event, review) {
      event?.stopPropagation?.()
      summaryDetailReview.value = review
    }

    function renderSummaryCell(column, review) {
      switch (column.key) {
        case 'employee':
          return ctx.userName(review.employeeId)
        case 'department':
          return review.department
        case 'position':
          return summaryEmployeePosition(review)
        case 'period':
          return review.period
        case 'status':
          return h('text', { class: ctx.reviewStatusClass(review) }, ctx.reviewStatusText(review))
        case 'objectiveScore':
          return String(ctx.totalObjectiveScore(review))
        case 'managerGrade':
          return review.managerGrade || '-'
        case 'hrbpGrade':
          return review.hrbpGrade || '-'
        case 'finalGrade':
          if (review.employeeConfirmResult === 'disputed') return '-'
          return ctx.effectiveGrade(review) || '-'
        case 'employeeConfirm':
          return review.employeeConfirmResult === 'confirmed' ? '已确认' : review.employeeConfirmResult === 'disputed' ? '有异议' : '-'
        case 'actions':
          return h('view', { class: 'summary-action-buttons' }, [
            h('button', {
              class: 'dt-btn dt-btn-light small summary-view-btn',
              onClick: (event) => handleViewReview(event, review)
            }, '查看'),
            ctx.canDeleteReview() ? h('button', {
              class: 'dt-btn dt-btn-danger-light small',
              onClick: (event) => handleDeleteReview(event, review)
            }, '删除') : null
          ])
        default:
          return '-'
      }
    }

    function summaryEmployeePosition(review) {
      const user = ctx.state.users.find((item) => item.id === review.employeeId)
      return user?.position || '-'
    }

    return () => {
      const summaryColumns = SUMMARY_TABLE_COLUMNS
      const summaryRows = summaryPagination.rows.value

      return h('view', { class: 'summary-page' }, [
        h('view', { class: 'page-head' }, [
          h('view', { class: 'summary-page-head-copy' }, [
            h('text', { class: 'page-title' }, ctx.sectionTitle.value),
            h('text', { class: 'page-desc' }, '按人员和月份查看进度、分档并导出结果')
          ]),
          h('view', { class: 'head-actions' }, [
            ctx.canExportReviews() ? h('button', { class: 'dt-btn dt-btn-primary', onClick: ctx.exportSummary }, '导出当前筛选') : null,
            h('button', { class: 'dt-btn dt-btn-light', onClick: ctx.refreshData }, '刷新')
          ])
        ]),

        h('view', { class: 'table-panel-stack summary-table-panel-stack' }, [
          h('section', { class: 'panel table-panel' }, [
            h('view', { class: 'panel-head' }, [
              h('text', { class: 'panel-title' }, '汇总列表'),
              h('text', { class: 'count-pill' }, `${ctx.summaryReviews.value.length} / ${ctx.state.reviews.length}`)
            ]),
            h('view', { class: ['summary-filter-shell', summaryFiltersCollapsed.value ? 'collapsed' : 'expanded', departmentDropdownOpen.value ? 'dropdown-open' : ''] }, [
              h('button', { class: 'summary-filter-toggle', onClick: toggleSummaryFilters }, [
                h('text', { class: 'summary-filter-toggle-title' }, '筛选条件'),
                summaryFilterCount.value > 0 ? h('text', { class: 'summary-filter-count' }, `已选 ${summaryFilterCount.value}`) : null,
                h('text', { class: ['summary-filter-arrow', summaryFiltersCollapsed.value ? '' : 'expanded'] })
              ]),
              h('view', { class: 'filters summary-filters' }, [
                h('input', { class: 'field-input', value: ctx.summaryFilters.employeeName, placeholder: '员工姓名', onInput: (event) => { ctx.summaryFilters.employeeName = event.detail?.value ?? event.target.value } }),
                renderDepartmentFilter(),
                renderSummaryPeriodFilter(),
                h('button', { class: 'dt-btn dt-btn-light', onClick: ctx.resetFilters }, '重置')
              ])
            ]),
            h('view', { class: 'table-wrap' }, [
              h('table', { class: 'summary-table' }, [
                h('thead', {}, h('tr', {}, summaryColumns.map((column) => h('th', { class: summaryColumnClass(column) }, column.label)))),
                h('tbody', {}, summaryRows.length
                  ? summaryRows.map((review) => h('tr', {}, summaryColumns.map((column) =>
                      h('td', { class: summaryColumnClass(column) }, renderSummaryCell(column, review))
                    )))
                  : h('tr', {}, [
                      h('td', { class: 'summary-empty-row', colspan: summaryColumns.length }, '当前没有汇总记录')
                    ]))
              ])
            ])
          ]),
          renderTablePagination(summaryPagination)
        ]),
        summaryDetailReview.value
          ? h(Teleport, { to: 'body' }, [
              h(HistoryReviewDetailModal, {
                review: summaryDetailReview.value,
                onClose: () => { summaryDetailReview.value = null }
              })
            ])
          : null
      ])
    }

    function buildSummaryDepartmentTree(users = []) {
      const root = new Map()
      for (const user of users || []) {
        const levels = summaryDepartmentLevels(user)
        let currentMap = root
        let currentNode = null
        let path = ''
        for (const [index, level] of levels.entries()) {
          path = [path, level].filter(Boolean).join(' / ')
          currentNode = ensureSummaryDepartmentNode(currentMap, `${currentNode?.key || 'root'}/l${index + 1}:${level}`, level, path)
          currentNode.count += 1
          currentMap = currentNode.childMap
        }
      }
      return finalizeSummaryDepartmentNodes([...root.values()])
    }

    function ensureSummaryDepartmentNode(map, key, name, path) {
      if (!map.has(key)) {
        map.set(key, { key, name, path, count: 0, childMap: new Map(), children: [] })
      }
      return map.get(key)
    }

    function finalizeSummaryDepartmentNodes(nodes = []) {
      return nodes
        .sort((left, right) => left.name.localeCompare(right.name, 'zh-Hans-CN'))
        .map((node) => ({
          key: node.key,
          name: node.name,
          path: node.path,
          count: node.count,
          children: finalizeSummaryDepartmentNodes([...node.childMap.values()])
        }))
    }

    function filterSummaryDepartmentTree(nodes = [], keyword = '') {
      const search = String(keyword || '').trim().toLowerCase()
      if (!search) return nodes
      const result = []
      for (const node of nodes) {
        const selfMatched = [node.name, node.path].filter(Boolean).join(' ').toLowerCase().includes(search)
        const children = selfMatched ? node.children : filterSummaryDepartmentTree(node.children, search)
        if (selfMatched || children.length > 0) {
          result.push({ ...node, children })
        }
      }
      return result
    }

    function flattenSummaryDepartmentTree(nodes = [], expandedKeys, keyword = '', depth = 1) {
      const rows = []
      const forceExpanded = String(keyword || '').trim() !== ''
      for (const node of nodes) {
        const expandable = node.children.length > 0
        const expanded = forceExpanded || expandedKeys.has(node.key)
        rows.push({ type: 'department', key: node.key, depth, name: node.name, path: node.path, count: node.count, node, expandable, expanded })
        if (expanded) {
          rows.push(...flattenSummaryDepartmentTree(node.children, expandedKeys, keyword, depth + 1))
        }
      }
      return rows
    }

    function collectSummaryDepartmentKeys(nodes = []) {
      const keys = []
      for (const node of nodes) {
        if (node.children.length > 0) keys.push(node.key)
        keys.push(...collectSummaryDepartmentKeys(node.children))
      }
      return keys
    }

    function summaryDepartmentLevels(user = {}) {
      const parts = String(user.department || '').split('/').map((item) => item.trim()).filter(Boolean)
      const levels = [
        firstText(user.departmentLevel1, parts[0]),
        firstText(user.departmentLevel2, parts[1]),
        firstText(user.departmentLevel3, parts[2])
      ].filter(Boolean)
      return levels.length > 0 ? levels : ['未设置部门']
    }

    function firstText(...values) {
      for (const value of values) {
        const text = String(value || '').trim()
        if (text) return text
      }
      return ''
    }

    function summaryDepartmentPaths(row) {
      const paths = []
      collectSummaryDepartmentPaths(row?.node, paths)
      return paths
    }

    function collectSummaryDepartmentPaths(node, paths) {
      if (!node) return
      if (node.path) paths.push(node.path)
      for (const child of node.children || []) {
        collectSummaryDepartmentPaths(child, paths)
      }
    }

    function summaryDepartmentCheckState(row) {
      const paths = summaryDepartmentPaths(row)
      if (paths.length === 0) return 'unchecked'
      const selected = new Set(selectedDepartmentLabels.value)
      if (selected.has(row.path)) return 'checked'
      return paths.some((path) => selected.has(path)) ? 'indeterminate' : 'unchecked'
    }

    function toggleSummaryDepartment(row) {
      const selected = new Set(selectedDepartmentLabels.value)
      const paths = summaryDepartmentPaths(row)
      const checked = selected.has(row.path)
      for (const path of paths) {
        selected.delete(path)
      }
      if (!checked && row.path) {
        selected.add(row.path)
      }
      setSummaryDepartments([...selected])
    }

    function toggleSummaryDepartmentExpand(row) {
      if (!row.expandable) return
      const next = new Set(summaryDepartmentExpandedKeys.value)
      if (next.has(row.key)) {
        next.delete(row.key)
      } else {
        next.add(row.key)
      }
      summaryDepartmentExpandedKeys.value = next
    }

    function setSummaryDepartments(names = []) {
      const normalized = []
      const seen = new Set()
      for (const name of names) {
        const text = String(name || '').trim()
        if (!text || seen.has(text)) continue
        seen.add(text)
        normalized.push(text)
      }
      ctx.summaryFilters.departmentNames = normalized
      ctx.summaryFilters.departmentName = normalized.join(',')
    }

    function clearSummaryDepartments() {
      departmentSearchKeyword.value = ''
      setSummaryDepartments([])
    }
  }
}
</script>

<style>
/* 页面专属样式：从 styles/performance.css 拆分 */
.summary-page {
  width: 100%;
  max-width: 1480px;
  margin: 0 auto;
  display: grid;
  gap: 16px;
}

.summary-page .panel {
  overflow: visible;
}

.summary-department-filter {
  position: relative;
  min-width: 0;
  z-index: 6;
}

.summary-month-picker {
  width: 100%;
  min-width: 0;
}

.summary-month-value {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  cursor: pointer;
}

.summary-month-value.selected .summary-month-text {
  color: #1f2329;
  font-weight: 500;
}

.summary-month-text {
  min-width: 0;
  flex: 1 1 auto;
  color: #86909c;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.summary-month-arrow {
  position: relative;
  width: 14px;
  height: 14px;
  flex: 0 0 14px;
}

.summary-month-arrow::before {
  content: "";
  position: absolute;
  left: 4px;
  top: 4px;
  width: 6px;
  height: 6px;
  border-right: 1.5px solid #86909c;
  border-bottom: 1.5px solid #86909c;
  transform: rotate(45deg);
}

.summary-department-trigger {
  width: 100%;
  height: 32px;
  min-height: 32px;
  margin: 0;
  padding: 0 12px;
  border: 1px solid #e5e6eb;
  border-radius: 4px;
  display: flex;
  align-items: center;
  gap: 8px;
  background: #fff;
  color: #1f2329;
  line-height: 30px;
  text-align: left;
  transition: border-color 0.16s ease, box-shadow 0.16s ease, background 0.16s ease;
}

.summary-department-trigger.active,
.summary-department-trigger:hover {
  border-color: #1677ff;
  background: #fbfdff;
  box-shadow: 0 0 0 3px rgba(22, 119, 255, 0.08);
}

.summary-department-trigger-text {
  min-width: 0;
  flex: 1 1 auto;
  color: #86909c;
  font-size: 13px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.summary-department-trigger-text.selected {
  color: #1f2329;
  font-weight: 600;
}

.summary-department-selected-count {
  flex: 0 0 auto;
  padding: 2px 7px;
  border-radius: 999px;
  background: #eef6ff;
  color: #1677ff;
  font-size: 12px;
  font-weight: 700;
  line-height: 18px;
}

.summary-department-arrow {
  position: relative;
  width: 18px;
  height: 18px;
  flex: 0 0 18px;
  color: #86909c;
}

.summary-department-arrow::before {
  content: "";
  position: absolute;
  left: 6px;
  top: 5px;
  width: 6px;
  height: 6px;
  border-right: 1.5px solid currentColor;
  border-bottom: 1.5px solid currentColor;
  transform: rotate(45deg);
  transform-origin: center;
  transition: transform 0.16s ease, color 0.16s ease;
}

.summary-department-arrow.expanded::before {
  transform: rotate(225deg);
}

.summary-department-panel {
  position: absolute;
  top: calc(100% + 6px);
  left: 0;
  right: 0;
  z-index: 30;
  min-width: 320px;
  border: 1px solid #dbe7f7;
  border-radius: 8px;
  background: #fff;
  box-shadow: 0 16px 38px rgba(31, 35, 41, 0.14);
  overflow: hidden;
}

.summary-department-search {
  width: calc(100% - 24px);
  height: 34px;
  min-height: 34px;
  margin: 12px;
  padding: 0 12px;
  border: 1px solid #e5e6eb;
  border-radius: 6px;
  background: #f7f8fa;
  color: #1f2329;
  font-size: 13px;
  line-height: 32px;
}

.summary-department-tree {
  max-height: 280px;
  overflow: auto;
}

.summary-department-row {
  position: relative;
  min-width: 0;
  min-height: 38px;
  border-top: 1px solid #f2f3f5;
  display: grid;
  grid-template-columns: minmax(0, 1fr) 38px;
  align-items: center;
}

.summary-department-node {
  min-width: 0;
  min-height: 38px;
  margin: 0;
  padding: 0 10px 0 12px;
  border: 0;
  border-radius: 0;
  display: flex;
  align-items: center;
  gap: 8px;
  background: #fff;
  color: #1f2329;
  line-height: 1.4;
  text-align: left;
}

.summary-department-node:hover,
.summary-department-row:hover .summary-department-node {
  background: #f7fbff;
}

.summary-department-row.depth-2 .summary-department-node {
  padding-left: 28px;
}

.summary-department-row.depth-3 .summary-department-node {
  padding-left: 44px;
}

.summary-department-row.depth-4 .summary-department-node,
.summary-department-row.depth-5 .summary-department-node {
  padding-left: 60px;
}

.summary-department-chevron {
  position: relative;
  width: 16px;
  height: 16px;
  flex: 0 0 16px;
  border-radius: 4px;
  color: #8a94a6;
}

.summary-department-chevron::before {
  content: "";
  position: absolute;
  left: 5px;
  top: 4px;
  width: 6px;
  height: 6px;
  border-right: 1.5px solid currentColor;
  border-bottom: 1.5px solid currentColor;
  transform: rotate(-45deg);
  transform-origin: center;
  transition: transform 0.16s ease, color 0.16s ease;
}

.summary-department-chevron.expanded::before {
  transform: rotate(45deg);
}

.summary-department-chevron.placeholder::before {
  display: none;
}

.summary-department-name {
  min-width: 0;
  flex: 1 1 auto;
  color: #1f2329;
  font-size: 13px;
  font-weight: 700;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.summary-department-count {
  flex: 0 0 auto;
  color: #86909c;
  font-size: 12px;
}

.summary-department-check {
  width: 18px;
  height: 18px;
  min-height: 18px;
  margin: 0 10px;
  padding: 0;
  border: 1px solid #c9d3e2;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  justify-self: center;
  background: #fff;
  color: #fff;
  font-size: 12px;
  font-weight: 800;
  line-height: 16px;
}

.summary-department-check.checked,
.summary-department-check-indeterminate {
  border-color: #1677ff;
  background: #1677ff;
}

.summary-department-actions {
  padding: 10px 12px;
  border-top: 1px solid #eef1f6;
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  background: #fbfcff;
}

.summary-department-empty {
  padding: 28px 12px;
  color: #86909c;
  font-size: 13px;
  text-align: center;
}

.summary-page .table-wrap {
  position: relative;
  z-index: 1;
}

.summary-action-buttons {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  white-space: nowrap;
}

.summary-view-btn {
  min-width: 56px;
}

@media (max-width: 960px) {
  .summary-page {
    max-width: none;
    padding: 16px;
  }

  .summary-page-head-copy {
    display: none;
  }

  .summary-page > .page-head {
    justify-content: flex-start;
    margin-bottom: 12px;
  }

  .summary-page .summary-table {
    min-width: 0;
    table-layout: fixed;
  }

  .summary-page .summary-table .summary-mobile-hidden {
    display: none;
  }

  .summary-page .summary-table th,
  .summary-page .summary-table td {
    padding: 12px;
  }

  .summary-page .summary-col-employee {
    width: 30%;
  }

  .summary-page .summary-col-period {
    width: 34%;
  }

  .summary-page .summary-col-finalGrade {
    width: 36%;
  }

  .summary-page .summary-table td {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}
</style>
