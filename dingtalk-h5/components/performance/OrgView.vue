<template>
  <view class="org-page">
    <view class="page-head">
      <view>
        <text class="page-title">{{ ctx.sectionTitle.value }}</text>
        <text class="page-desc">按权限查看流程人员、直属上级和 HRBP 关系</text>
      </view>
      <button class="dt-btn dt-btn-light" @click="ctx.refreshData">刷新</button>
    </view>

    <view class="org-toolbar">
      <view class="org-stat">
        <text class="org-stat-value">{{ filteredUsers.length }}</text>
        <text class="org-stat-label">可见人员</text>
      </view>
      <view class="org-stat">
        <text class="org-stat-value">{{ managerConfiguredCount }}</text>
        <text class="org-stat-label">已配置上级</text>
      </view>
      <view class="org-stat">
        <text class="org-stat-value">{{ hrbpConfiguredCount }}</text>
        <text class="org-stat-label">已配置HRBP</text>
      </view>
      <view class="org-stat">
        <text class="org-stat-value">{{ positionConfiguredCount }}</text>
        <text class="org-stat-label">已设置岗位</text>
      </view>
    </view>

    <section class="panel org-list-panel">
      <view class="panel-head">
        <view>
          <text class="panel-title">人员列表</text>
          <text class="org-scope-hint">已按权限过滤，数据来自用户表</text>
        </view>
        <text class="count-pill">{{ filteredUsers.length }} / {{ ctx.state.users.length }} 人</text>
      </view>

      <view class="org-filter-shell" :class="{ collapsed: orgFiltersCollapsed, expanded: !orgFiltersCollapsed, 'dropdown-open': orgDepartmentDropdownOpen }">
        <button class="org-filter-toggle" @click="toggleOrgFilters">
          <text class="org-filter-toggle-title">搜索条件</text>
          <text v-if="orgFilterCount > 0" class="org-filter-count">已选 {{ orgFilterCount }}</text>
          <text :class="['org-filter-arrow', orgFiltersCollapsed ? '' : 'expanded']"></text>
        </button>

        <view class="org-filter-bar">
          <input
            class="field-input org-filter-input"
            :value="orgFilters.employeeKeyword"
            placeholder="搜索员工姓名/账号"
            @input="handleOrgEmployeeKeywordInput"
          />

          <view class="org-department-filter" :class="{ active: orgDepartmentDropdownOpen }">
            <button
              class="org-department-trigger"
              :class="{ active: orgDepartmentDropdownOpen }"
              @click="toggleOrgDepartmentDropdown"
            >
              <text class="org-department-trigger-text" :class="{ selected: selectedOrgDepartmentLabels.length > 0 }">
                {{ selectedOrgDepartmentLabels.length > 0 ? selectedOrgDepartmentLabels.join('、') : '部门名称' }}
              </text>
              <text v-if="selectedOrgDepartmentLabels.length > 0" class="org-department-selected-count">{{ selectedOrgDepartmentLabels.length }} 项</text>
              <text :class="['org-department-arrow', orgDepartmentDropdownOpen ? 'expanded' : '']"></text>
            </button>

            <view v-if="orgDepartmentDropdownOpen" class="org-department-panel">
              <input
                class="org-department-search"
                :value="orgDepartmentSearchKeyword"
                placeholder="搜索部门名称"
                @input="handleOrgDepartmentSearchInput"
              />
              <view class="org-department-select-tree">
                <view
                  v-for="row in orgFilterDepartmentRows"
                  :key="row.key"
                  :class="['org-department-option-row', `depth-${row.depth}`, row.expanded ? 'expanded' : 'collapsed']"
                >
                  <button class="org-department-option" @click="toggleOrgFilterDepartmentExpand(row)">
                    <text :class="['org-department-option-chevron', row.expandable ? (row.expanded ? 'expanded' : 'collapsed') : 'placeholder']"></text>
                    <text class="org-department-option-name">{{ row.name }}</text>
                    <text class="org-department-option-count">{{ row.count }} 人</text>
                  </button>
                  <button
                    :class="['org-department-check', orgDepartmentCheckState(row) === 'checked' ? 'checked' : '', orgDepartmentCheckState(row) === 'indeterminate' ? 'indeterminate' : '']"
                    @click.stop="toggleOrgFilterDepartment(row)"
                  >
                    {{ orgDepartmentCheckState(row) === 'checked' ? '✓' : orgDepartmentCheckState(row) === 'indeterminate' ? '-' : '' }}
                  </button>
                </view>
                <view v-if="orgFilterDepartmentRows.length === 0" class="org-department-empty">暂无匹配部门</view>
              </view>
              <view class="org-department-actions">
                <button class="dt-btn dt-btn-light small" @click="clearOrgFilterDepartments">清空</button>
                <button class="dt-btn dt-btn-primary small" @click="orgDepartmentDropdownOpen = false">确定</button>
              </view>
            </view>
          </view>

          <select class="field-select" v-model="orgFilters.managerStatus">
            <option value="">全部上级状态</option>
            <option value="set">已设置上级</option>
            <option value="unset">未设置上级</option>
          </select>
          <select class="field-select" v-model="orgFilters.hrbpStatus">
            <option value="">全部HRBP状态</option>
            <option value="set">已设置HRBP</option>
            <option value="unset">未设置HRBP</option>
          </select>
          <button class="dt-btn dt-btn-light" @click="resetOrgFilters">重置</button>
        </view>
      </view>

      <view v-if="departmentTree.length === 0" class="empty org-empty">
        {{ ctx.state.users.length === 0 ? '当前没有可见人员' : '当前没有符合条件的人员' }}
      </view>

      <view v-else class="org-tree-table">
        <view class="org-tree-table-head">
          <text>部门名称</text>
        </view>

        <view class="org-tree-list">
          <view
            v-for="row in treeRows"
            :key="row.key"
            :class="['org-tree-row', row.type, `depth-${row.depth}`]"
          >
            <view
              v-if="row.type === 'department'"
              :class="['org-department-row', row.expandable ? 'expandable' : 'leaf', row.expanded ? 'expanded' : 'collapsed']"
              @click="toggleDepartment(row)"
            >
              <view class="org-department-title">
                <text :class="['org-tree-chevron', row.expandable ? (row.expanded ? 'expanded' : 'collapsed') : 'placeholder']"></text>
                <text class="org-department-name">{{ row.name }}</text>
              </view>
              <text class="count-pill muted">{{ row.count }} 人</text>
            </view>

            <view v-else-if="row.type === 'employee'" class="org-employee-list">
              <view class="org-employee-row">
                <view class="org-person-cell">
                  <view class="org-avatar">{{ initials(row.user.name || row.user.id) }}</view>
                  <view class="org-person-main">
                    <view class="org-person-title-line">
                      <text class="user-name">{{ row.user.name || row.user.id }}</text>
                      <text class="org-position-inline">{{ row.user.position || '未设置岗位' }}</text>
                    </view>
                    <text class="org-account">{{ row.user.id }}</text>
                    <view class="org-flow-summary org-flow-summary-mobile">
                      <view>
                        <text>上级</text>
                        <strong>{{ ctx.userName(row.user.managerId) }}</strong>
                      </view>
                      <view>
                        <text>HRBP</text>
                        <strong>{{ ctx.userName(row.user.hrbpId) }}</strong>
                      </view>
                    </view>
                  </view>
                </view>

                <view class="org-employee-meta">
                  <text>{{ row.user.position || '未设置岗位' }}</text>
                  <text>{{ departmentText(row.user) }}</text>
                </view>

                <view class="org-flow-summary">
                  <view>
                    <text class="org-flow-label-desktop">直属上级</text>
                    <text class="org-flow-label-mobile">上级</text>
                    <strong>{{ ctx.userName(row.user.managerId) }}</strong>
                  </view>
                  <view>
                    <text>HRBP</text>
                    <strong>{{ ctx.userName(row.user.hrbpId) }}</strong>
                  </view>
                </view>

                <button
                  v-if="canEditUsers"
                  class="dt-btn dt-btn-light small org-config-btn"
                  @click="openConfig(row.user)"
                >
                  配置
                </button>
              </view>
            </view>
          </view>
        </view>
      </view>
    </section>

    <view v-if="selectedUser" class="org-config-modal-mask" @click.self="closeConfig">
      <view class="org-config-modal">
        <view class="org-config-modal-head">
          <view>
            <text class="org-config-title">配置流程审批人</text>
            <text class="org-config-subtitle">{{ selectedUser.name || selectedUser.id }} · {{ departmentText(selectedUser) }}</text>
          </view>
          <button class="process-modal-close" @click="closeConfig">×</button>
        </view>

        <view class="org-config-user">
          <view class="org-avatar">{{ initials(selectedUser.name || selectedUser.id) }}</view>
          <view>
            <text class="user-name">{{ selectedUser.name || selectedUser.id }}</text>
            <text class="org-account">{{ [selectedUser.id, selectedUser.position || '未设置岗位'].filter(Boolean).join(' · ') }}</text>
          </view>
        </view>

        <view class="org-config-form">
          <view class="org-form-field relation-picker-field" :class="{ active: activeRelationPicker === 'manager' }">
            <text>直属上级</text>
            <button
              class="relation-picker-trigger"
              :class="{ active: activeRelationPicker === 'manager' }"
              @click="toggleRelationPicker('manager')"
            >
              <text class="relation-picker-value">{{ relationPickerText(configForm.managerId, '无直属上级') }}</text>
              <text :class="['relation-picker-arrow', activeRelationPicker === 'manager' ? 'expanded' : '']"></text>
            </button>
            <view v-if="activeRelationPicker === 'manager'" class="relation-picker-panel">
              <button class="relation-picker-clear" @click="selectRelationUser('manager', '')">无直属上级</button>
              <view class="relation-picker-tree">
                <view
                  v-for="row in managerPickerRows"
                  :key="`manager-${row.key}`"
                  :class="['relation-picker-row', row.type, `depth-${row.depth}`]"
                >
                  <button
                    v-if="row.type === 'department'"
                    class="relation-picker-dept"
                    @click="toggleRelationDepartment('manager', row)"
                  >
                    <text :class="['relation-picker-chevron', row.expandable ? (row.expanded ? 'expanded' : 'collapsed') : 'placeholder']"></text>
                    <text class="relation-picker-dept-name">{{ row.name }}</text>
                    <text class="relation-picker-count">{{ row.count }} 人</text>
                  </button>
                  <button
                    v-else
                    class="relation-picker-user"
                    :class="{ selected: configForm.managerId === row.user.id }"
                    @click="selectRelationUser('manager', row.user.id)"
                  >
                    <text class="relation-picker-radio" :class="{ checked: configForm.managerId === row.user.id }"></text>
                    <view class="relation-picker-user-main">
                      <text class="relation-picker-user-name">{{ row.user.name || row.user.id }}</text>
                      <text class="relation-picker-user-meta">{{ [row.user.position, departmentText(row.user)].filter(Boolean).join(' · ') }}</text>
                    </view>
                  </button>
                </view>
              </view>
            </view>
          </view>
          <view class="org-form-field relation-picker-field" :class="{ active: activeRelationPicker === 'hrbp' }">
            <text>HRBP</text>
            <button
              class="relation-picker-trigger"
              :class="{ active: activeRelationPicker === 'hrbp' }"
              @click="toggleRelationPicker('hrbp')"
            >
              <text class="relation-picker-value">{{ relationPickerText(configForm.hrbpId, '无HRBP') }}</text>
              <text :class="['relation-picker-arrow', activeRelationPicker === 'hrbp' ? 'expanded' : '']"></text>
            </button>
            <view v-if="activeRelationPicker === 'hrbp'" class="relation-picker-panel">
              <button class="relation-picker-clear" @click="selectRelationUser('hrbp', '')">无HRBP</button>
              <view class="relation-picker-tree">
                <view
                  v-for="row in hrbpPickerRows"
                  :key="`hrbp-${row.key}`"
                  :class="['relation-picker-row', row.type, `depth-${row.depth}`]"
                >
                  <button
                    v-if="row.type === 'department'"
                    class="relation-picker-dept"
                    @click="toggleRelationDepartment('hrbp', row)"
                  >
                    <text :class="['relation-picker-chevron', row.expandable ? (row.expanded ? 'expanded' : 'collapsed') : 'placeholder']"></text>
                    <text class="relation-picker-dept-name">{{ row.name }}</text>
                    <text class="relation-picker-count">{{ row.count }} 人</text>
                  </button>
                  <button
                    v-else
                    class="relation-picker-user"
                    :class="{ selected: configForm.hrbpId === row.user.id }"
                    @click="selectRelationUser('hrbp', row.user.id)"
                  >
                    <text class="relation-picker-radio" :class="{ checked: configForm.hrbpId === row.user.id }"></text>
                    <view class="relation-picker-user-main">
                      <text class="relation-picker-user-name">{{ row.user.name || row.user.id }}</text>
                      <text class="relation-picker-user-meta">{{ [row.user.position, departmentText(row.user)].filter(Boolean).join(' · ') }}</text>
                    </view>
                  </button>
                </view>
              </view>
            </view>
          </view>
        </view>

        <view class="process-modal-actions">
          <button class="dt-btn dt-btn-light" @click="closeConfig">取消</button>
          <button class="dt-btn dt-btn-primary" @click="saveConfig">保存配置</button>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { computed, reactive, ref, watch } from 'vue'
import { usePerformanceContext } from './context'

const ctx = usePerformanceContext()

const selectedUser = ref(null)
const orgFilters = reactive({
  employeeKeyword: '',
  departmentNames: [],
  managerStatus: '',
  hrbpStatus: ''
})
const configForm = reactive({
  id: '',
  name: '',
  password: '',
  position: '',
  department: '',
  departmentLevel1: '',
  departmentLevel2: '',
  departmentLevel3: '',
  managerId: '',
  hrbpId: ''
})

const canEditUsers = computed(() => ctx.canEditUsers())
const filteredUsers = computed(() => ctx.state.users.filter(matchesOrgFilters))
const departmentTree = computed(() => buildDepartmentTree(filteredUsers.value))
const expandedDepartmentKeys = ref(new Set())
const orgFiltersCollapsed = ref(true)
const orgDepartmentDropdownOpen = ref(false)
const orgDepartmentSearchKeyword = ref('')
const orgFilterDepartmentExpandedKeys = ref(new Set())
const orgFilterCount = computed(() => [
  normalizeText(orgFilters.employeeKeyword),
  selectedOrgDepartmentLabels.value.length > 0,
  orgFilters.managerStatus,
  orgFilters.hrbpStatus
].filter(Boolean).length)
const orgHasActiveFilters = computed(() => Boolean(
  normalizeText(orgFilters.employeeKeyword) ||
  selectedOrgDepartmentLabels.value.length > 0 ||
  orgFilters.managerStatus ||
  orgFilters.hrbpStatus
))
const activeRelationPicker = ref('')
const managerPickerExpandedKeys = ref(new Set())
const hrbpPickerExpandedKeys = ref(new Set())
const treeRows = computed(() => flattenDepartmentTree(departmentTree.value, expandedDepartmentKeys.value))
const orgFilterDepartmentTree = computed(() => buildDepartmentTree(ctx.state.users))
const orgFilterDepartmentRows = computed(() => flattenDepartmentSelectionTree(filterDepartmentTree(orgFilterDepartmentTree.value, orgDepartmentSearchKeyword.value), orgFilterDepartmentExpandedKeys.value, orgDepartmentSearchKeyword.value))
const selectedOrgDepartmentLabels = computed(() => Array.isArray(orgFilters.departmentNames) ? orgFilters.departmentNames : [])
const hrbpOptions = computed(() => ctx.state.users.filter((user) => user.id && user.id !== configForm.id))
const managerPickerTree = computed(() => buildDepartmentTree(managerOptions(configForm.id)))
const hrbpPickerTree = computed(() => buildDepartmentTree(hrbpOptions.value))
const managerPickerRows = computed(() => flattenDepartmentTree(managerPickerTree.value, managerPickerExpandedKeys.value))
const hrbpPickerRows = computed(() => flattenDepartmentTree(hrbpPickerTree.value, hrbpPickerExpandedKeys.value))
const managerConfiguredCount = computed(() => filteredUsers.value.filter((user) => user.managerId).length)
const hrbpConfiguredCount = computed(() => filteredUsers.value.filter((user) => user.hrbpId).length)
const positionConfiguredCount = computed(() => filteredUsers.value.filter((user) => hasConfiguredPosition(user)).length)

watch(departmentTree, (nodes) => syncExpandedDepartments(nodes), { immediate: true })
watch(orgFilterDepartmentTree, (nodes) => syncOrgFilterDepartmentState(nodes), { immediate: true })
watch(managerPickerTree, (nodes) => syncRelationExpandedKeys(managerPickerExpandedKeys, nodes), { immediate: true })
watch(hrbpPickerTree, (nodes) => syncRelationExpandedKeys(hrbpPickerExpandedKeys, nodes), { immediate: true })

function buildDepartmentTree(users = []) {
  const root = new Map()
  for (const user of users || []) {
    const levels = realDepartmentLevels(user)
    let currentMap = root
    let currentNode = null
    for (const [index, level] of levels.entries()) {
      currentNode = ensureTreeNode(currentMap, `${currentNode?.key || 'root'}/l${index + 1}:${level}`, level)
      currentNode.count += 1
      currentMap = currentNode.childMap
    }
    if (currentNode) {
      currentNode.users.push(user)
    }
  }
  return finalizeTreeNodes([...root.values()])
}

function ensureTreeNode(map, key, name) {
  if (!map.has(key)) {
    map.set(key, { key, name, count: 0, childMap: new Map(), children: [], users: [] })
  }
  return map.get(key)
}

function finalizeTreeNodes(nodes) {
  return nodes
    .sort((left, right) => left.name.localeCompare(right.name, 'zh-Hans-CN'))
    .map((node) => ({
      key: node.key,
      name: node.name,
      count: node.count,
      children: finalizeTreeNodes([...node.childMap.values()]),
      users: sortUsers(node.users)
    }))
}

function flattenDepartmentTree(nodes = [], expandedKeys, depth = 1, parentPath = '') {
  const rows = []
  for (const node of nodes) {
    const path = [parentPath, node.name].filter(Boolean).join(' / ')
    const hasChildren = node.children.length > 0 || node.users.length > 0
    rows.push({
      type: 'department',
      key: node.key,
      depth,
      name: node.name,
      path,
      count: node.count,
      expandable: hasChildren,
      hasChildren: node.children.length > 0 || node.users.length > 0,
      expanded: expandedKeys.has(node.key)
    })
    if (!expandedKeys.has(node.key)) continue
    for (const user of node.users) {
      rows.push({ type: 'employee', key: `${node.key}/user:${user.id}`, depth: depth + 1, user })
    }
    rows.push(...flattenDepartmentTree(node.children, expandedKeys, depth + 1, path))
  }
  return rows
}

function filterDepartmentTree(nodes = [], keyword = '') {
  const search = normalizeText(keyword)
  if (!search) return nodes
  const result = []
  for (const node of nodes) {
    const selfMatched = normalizeText([node.name, node.path].filter(Boolean).join(' ')).includes(search)
    const children = selfMatched ? node.children : filterDepartmentTree(node.children, keyword)
    if (selfMatched || children.length > 0) {
      result.push({ ...node, children })
    }
  }
  return result
}

function flattenDepartmentSelectionTree(nodes = [], expandedKeys, keyword = '', depth = 1, parentPath = '') {
  const rows = []
  const forceExpand = Boolean(normalizeText(keyword))
  for (const node of nodes) {
    const path = [parentPath, node.name].filter(Boolean).join(' / ')
    const expandable = node.children.length > 0
    const expanded = forceExpand || expandedKeys.has(node.key)
    rows.push({
      type: 'department',
      key: node.key,
      depth,
      name: node.name,
      path,
      count: node.count,
      expandable,
      expanded
    })
    if (expanded) {
      rows.push(...flattenDepartmentSelectionTree(node.children, expandedKeys, keyword, depth + 1, path))
    }
  }
  return rows
}

function matchesOrgFilters(user) {
  const keyword = normalizeText(orgFilters.employeeKeyword)
  if (keyword) {
    const haystack = normalizeText([
      user.name,
      user.id,
      user.account,
      user.position,
      departmentText(user)
    ].filter(Boolean).join(' '))
    if (!haystack.includes(keyword)) return false
  }
  if (!matchesOrgDepartmentFilter(user)) return false
  if (orgFilters.managerStatus === 'set' && !hasConfiguredManager(user)) return false
  if (orgFilters.managerStatus === 'unset' && hasConfiguredManager(user)) return false
  if (orgFilters.hrbpStatus === 'set' && !hasConfiguredHrbp(user)) return false
  if (orgFilters.hrbpStatus === 'unset' && hasConfiguredHrbp(user)) return false
  return true
}

function matchesOrgDepartmentFilter(user) {
  const names = Array.isArray(orgFilters.departmentNames) ? orgFilters.departmentNames : []
  if (names.length === 0) return true
  const department = departmentText(user)
  return names.some((name) => department === name || department.startsWith(`${name} / `))
}

function hasConfiguredPosition(user) {
  return Boolean(firstText(user.position))
}

function hasConfiguredManager(user) {
  return Boolean(firstText(user.managerId))
}

function hasConfiguredHrbp(user) {
  return Boolean(firstText(user.hrbpId))
}

function handleOrgEmployeeKeywordInput(event) {
  orgFilters.employeeKeyword = eventValue(event)
}

function handleOrgDepartmentSearchInput(event) {
  orgDepartmentSearchKeyword.value = eventValue(event)
}

function eventValue(event) {
  return event?.detail?.value ?? event?.target?.value ?? ''
}

function normalizeText(value) {
  return String(value || '').trim().toLowerCase()
}

function toggleOrgFilters() {
  orgFiltersCollapsed.value = !orgFiltersCollapsed.value
  if (orgFiltersCollapsed.value) {
    orgDepartmentDropdownOpen.value = false
  }
}

function toggleOrgDepartmentDropdown() {
  orgDepartmentDropdownOpen.value = !orgDepartmentDropdownOpen.value
}

function toggleOrgFilterDepartmentExpand(row) {
  if (!row.expandable) return
  const next = new Set(orgFilterDepartmentExpandedKeys.value)
  if (next.has(row.key)) {
    next.delete(row.key)
  } else {
    next.add(row.key)
  }
  orgFilterDepartmentExpandedKeys.value = next
}

function toggleOrgFilterDepartment(row) {
  const next = new Set(selectedOrgDepartmentLabels.value)
  if (next.has(row.path)) {
    next.delete(row.path)
  } else {
    next.add(row.path)
  }
  orgFilters.departmentNames = [...next].sort((left, right) => left.localeCompare(right, 'zh-Hans-CN'))
}

function orgDepartmentCheckState(row) {
  if (selectedOrgDepartmentLabels.value.includes(row.path)) return 'checked'
  const prefix = `${row.path} / `
  return selectedOrgDepartmentLabels.value.some((name) => name.startsWith(prefix)) ? 'indeterminate' : ''
}

function clearOrgFilterDepartments() {
  orgFilters.departmentNames = []
  orgDepartmentSearchKeyword.value = ''
}

function resetOrgFilters() {
  Object.assign(orgFilters, {
    employeeKeyword: '',
    departmentNames: [],
    managerStatus: '',
    hrbpStatus: ''
  })
  orgDepartmentSearchKeyword.value = ''
  orgDepartmentDropdownOpen.value = false
}

function syncOrgFilterDepartmentState(nodes = []) {
  const availableKeys = new Set(collectDepartmentKeys(nodes))
  const availablePaths = new Set(collectDepartmentPaths(nodes))
  orgFilters.departmentNames = selectedOrgDepartmentLabels.value.filter((name) => availablePaths.has(name))
  const next = new Set()
  for (const key of orgFilterDepartmentExpandedKeys.value) {
    if (availableKeys.has(key)) next.add(key)
  }
  orgFilterDepartmentExpandedKeys.value = next
}

function collectDepartmentPaths(nodes = [], parentPath = '') {
  const paths = []
  for (const node of nodes) {
    const path = [parentPath, node.name].filter(Boolean).join(' / ')
    paths.push(path)
    paths.push(...collectDepartmentPaths(node.children, path))
  }
  return paths
}

function syncExpandedDepartments(nodes = []) {
  const availableKeys = new Set(collectDepartmentKeys(nodes))
  const next = new Set()
  if (orgHasActiveFilters.value) {
    for (const key of availableKeys) next.add(key)
    expandedDepartmentKeys.value = next
    return
  }
  for (const key of expandedDepartmentKeys.value) {
    if (availableKeys.has(key)) next.add(key)
  }
  expandedDepartmentKeys.value = next
}

function collectDepartmentKeys(nodes = []) {
  const keys = []
  for (const node of nodes) {
    if (node.children.length > 0 || node.users.length > 0) {
      keys.push(node.key)
    }
    keys.push(...collectDepartmentKeys(node.children))
  }
  return keys
}

function toggleDepartment(row) {
  if (!row.expandable) return
  const next = new Set(expandedDepartmentKeys.value)
  if (next.has(row.key)) {
    next.delete(row.key)
  } else {
    next.add(row.key)
  }
  expandedDepartmentKeys.value = next
}

function relationExpandedRef(type) {
  return type === 'manager' ? managerPickerExpandedKeys : hrbpPickerExpandedKeys
}

function toggleRelationPicker(type) {
  activeRelationPicker.value = activeRelationPicker.value === type ? '' : type
}

function toggleRelationDepartment(type, row) {
  if (!row.expandable) return
  const target = relationExpandedRef(type)
  const next = new Set(target.value)
  if (next.has(row.key)) {
    next.delete(row.key)
  } else {
    next.add(row.key)
  }
  target.value = next
}

function selectRelationUser(type, id) {
  if (type === 'manager') {
    configForm.managerId = id
  } else {
    configForm.hrbpId = id
  }
  activeRelationPicker.value = ''
}

function syncRelationExpandedKeys(target, nodes = []) {
  const availableKeys = new Set(collectDepartmentKeys(nodes))
  const next = new Set()
  for (const key of target.value) {
    if (availableKeys.has(key)) next.add(key)
  }
  target.value = next
}

function sortUsers(users = []) {
  return [...users].sort((left, right) => {
    const a = [left.name, left.id].filter(Boolean).join('\x00')
    const b = [right.name, right.id].filter(Boolean).join('\x00')
    return a.localeCompare(b, 'zh-Hans-CN')
  })
}

function realDepartmentLevels(user) {
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

function managerOptions(currentId) {
  return ctx.state.users.filter((user) => user.id && user.id !== currentId)
}

function departmentText(user) {
  return [user.departmentLevel1, user.departmentLevel2, user.departmentLevel3].filter(Boolean).join(' / ') || user.department || '未设置部门'
}

function userOptionText(user) {
  return [user.name || user.id, user.position, departmentText(user)].filter(Boolean).join(' · ')
}

function relationPickerText(id, emptyText) {
  if (!id) return emptyText
  const user = ctx.state.users.find((item) => item.id === id)
  return user ? userOptionText(user) : id
}

function initials(value) {
  return String(value || '人').slice(0, 1).toUpperCase()
}

function openConfig(user) {
  selectedUser.value = user
  Object.assign(configForm, {
    id: user.id,
    name: user.name || user.id,
    password: '',
    position: user.position || '',
    department: user.department || '',
    departmentLevel1: user.departmentLevel1 || '',
    departmentLevel2: user.departmentLevel2 || '',
    departmentLevel3: user.departmentLevel3 || '',
    managerId: user.managerId || '',
    hrbpId: user.hrbpId || ''
  })
  activeRelationPicker.value = ''
  managerPickerExpandedKeys.value = new Set()
  hrbpPickerExpandedKeys.value = new Set()
}

function closeConfig() {
  selectedUser.value = null
  activeRelationPicker.value = ''
}

async function saveConfig() {
  await ctx.saveUser(configForm)
  closeConfig()
}
</script>
