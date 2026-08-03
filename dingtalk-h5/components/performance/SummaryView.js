import { computed, h, ref, watch } from 'vue'
import { usePerformanceContext } from './context'
import { renderTablePagination, useTablePagination } from './tablePagination'

const SUMMARY_TABLE_COLUMNS = [
  { key: 'employee', label: '员工', mobile: true },
  { key: 'department', label: '部门' },
  { key: 'position', label: '岗位' },
  { key: 'period', label: '考评月份', mobile: true },
  { key: 'status', label: '状态' },
  { key: 'objectiveScore', label: '目标得分' },
  { key: 'managerGrade', label: '上级分档' },
  { key: 'hrbpGrade', label: 'HRBP分档' },
  { key: 'finalGrade', label: '最终分档', mobile: true },
  { key: 'employeeConfirm', label: '员工确认' },
  { key: 'actions', label: '操作' }
]

export default {
  name: 'SummaryView',
  setup() {
    const ctx = usePerformanceContext()
    const departmentDropdownOpen = ref(false)
    const departmentSearchKeyword = ref('')
    const summaryFiltersCollapsed = ref(true)
    const summaryDepartmentExpandedKeys = ref(new Set())
    const summaryDepartmentTree = computed(() => buildSummaryDepartmentTree(ctx.state.users))
    const summaryDepartmentRows = computed(() => flattenSummaryDepartmentTree(filterSummaryDepartmentTree(summaryDepartmentTree.value, departmentSearchKeyword.value), summaryDepartmentExpandedKeys.value, departmentSearchKeyword.value))
    const selectedDepartmentLabels = computed(() => Array.isArray(ctx.summaryFilters.departmentNames) ? ctx.summaryFilters.departmentNames : [])
    const summaryTableRows = computed(() => ctx.summaryReviews.value)
    const summaryPagination = useTablePagination(summaryTableRows)
    const summaryFilterCount = computed(() => {
      return [
        String(ctx.summaryFilters.employeeName || '').trim(),
        selectedDepartmentLabels.value.length > 0,
        ctx.summaryFilters.period,
        ctx.summaryFilters.status
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
          return h('text', { class: ctx.statusClass(review.status) }, ctx.statusText(review.status))
        case 'objectiveScore':
          return String(ctx.totalObjectiveScore(review))
        case 'managerGrade':
          return review.managerGrade || '-'
        case 'hrbpGrade':
          return review.hrbpGrade || '-'
        case 'finalGrade':
          return ctx.effectiveGrade(review) || '-'
        case 'employeeConfirm':
          return review.employeeConfirmResult === 'confirmed' ? '已确认' : review.employeeConfirmResult === 'disputed' ? '有异议' : '-'
        case 'actions':
          return h('button', {
            class: 'dt-btn dt-btn-danger-light small',
            onClick: (event) => handleDeleteReview(event, review)
          }, '删除')
        default:
          return '-'
      }
    }

    function summaryEmployeePosition(review) {
      const user = ctx.state.users.find((item) => item.id === review.employeeId)
      return user?.position || '-'
    }

    return () => {
      const summaryColumns = SUMMARY_TABLE_COLUMNS.filter((column) => column.key !== 'actions' || ctx.canDeleteReview())
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
                h('select', { class: 'field-select', value: ctx.summaryFilters.status, onChange: (event) => { ctx.summaryFilters.status = event.target.value } }, [''].concat(Object.keys(ctx.statusMeta)).map((status) => h('option', { value: status }, status ? ctx.statusText(status) : '全部状态'))),
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
        ])
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
