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

    <OrgUserConfigDialog
      v-if="selectedUser"
      :selected-user="selectedUser"
      :config-form="configForm"
      :active-relation-picker="activeRelationPicker"
      :manager-picker-rows="managerPickerRows"
      :hrbp-picker-rows="hrbpPickerRows"
      @close="closeConfig"
      @save="saveConfig"
      @toggle-relation-picker="toggleRelationPicker"
      @toggle-relation-department="toggleRelationDepartment"
      @select-relation-user="selectRelationUser"
    />
  </view>
</template>

<script setup>
import { computed, reactive, ref, watch } from 'vue'
import OrgUserConfigDialog from './components/OrgUserConfigDialog.vue'
import { usePerformanceContext } from '../common/context'
import {
  buildDepartmentTree,
  collectDepartmentKeys,
  collectDepartmentPaths,
  departmentText,
  eventValue,
  filterDepartmentTree,
  firstText,
  flattenDepartmentSelectionTree,
  flattenDepartmentTree,
  initials,
  normalizeText
} from './composables/useOrgDirectory'

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
const hrbpOptions = computed(() => ctx.state.users.filter((user) => user.id))
const managerPickerTree = computed(() => buildDepartmentTree(managerOptions()))
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

function managerOptions() {
  return ctx.state.users.filter((user) => user.id)
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

<style>
/* 页面专属样式：从 styles/performance.css 拆分 */
.org-avatar {
  width: 28px;
  height: 28px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #1677ff;
  color: #fff;
  font-weight: 700;
}

.org-page {
  width: 100%;
  max-width: 1480px;
  margin: 0 auto;
  display: grid;
  gap: 16px;
}

.org-table-shell {
  overflow: auto;
}

.org-table {
  width: 100%;
  min-width: 980px;
  border-collapse: collapse;
}

.org-table th,
.org-table td {
  padding: 12px;
  border-top: 1px solid #f2f3f5;
  color: #4e5969;
  font-size: 13px;
  text-align: left;
  vertical-align: middle;
}

.org-table th {
  position: sticky;
  top: 0;
  z-index: 1;
  color: #1f2329;
  background: #fbfcff;
  font-weight: 700;
}

.org-table {
  min-width: 1400px;
}

.org-table td {
  background: #fff;
}

.org-toolbar {
  display: grid;
  grid-template-columns: repeat(4, minmax(140px, 1fr));
  gap: 12px;
}

.org-stat {
  padding: 16px;
  border: 1px solid #e5e6eb;
  border-radius: 14px;
  background: #fff;
  box-shadow: 0 8px 24px rgba(31, 35, 41, 0.04);
}

.org-stat-value {
  display: block;
  color: #1677ff;
  font-size: 28px;
  font-weight: 800;
}

.org-stat-label {
  display: block;
  margin-top: 4px;
  color: #86909c;
  font-size: 13px;
}

.org-create-grid {
  padding: 16px;
  display: grid;
  grid-template-columns: repeat(5, minmax(140px, 1fr));
  gap: 12px;
  align-items: center;
}

.org-wide-field {
  grid-column: span 2;
}

.org-person-cell {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}

.org-avatar {
  flex: 0 0 auto;
  width: 34px;
  height: 34px;
  border-radius: 10px;
  font-size: 13px;
}

.org-person-main {
  display: grid;
  gap: 4px;
  min-width: 0;
}

.org-person-title-line {
  min-width: 0;
  display: block;
}

.org-position-inline {
  display: none;
  flex: 0 1 auto;
  min-width: 0;
  max-width: 120px;
  padding: 2px 6px;
  border-radius: 999px;
  background: #f2f6ff;
  color: #4e5969;
  font-size: 11px;
  font-weight: 700;
  line-height: 16px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.org-flow-label-mobile {
  display: none;
}

.org-account {
  color: #86909c;
  font-size: 12px;
}

.org-list-panel {
  overflow: visible;
}

.org-scope-hint {
  display: block;
  margin-top: 4px;
  color: #86909c;
  font-size: 12px;
  font-weight: 500;
}

.org-filter-shell {
  position: relative;
  z-index: 5;
  overflow: visible;
}

.org-filter-shell.dropdown-open {
  z-index: 45;
}

.org-filter-toggle {
  display: none;
}

.org-filter-bar {
  position: relative;
  z-index: 5;
  padding: 14px;
  border-top: 1px solid #f2f3f5;
  border-bottom: 1px solid #f2f3f5;
  display: grid;
  grid-template-columns: minmax(180px, 1.2fr) minmax(220px, 1.4fr) minmax(150px, 0.8fr) minmax(150px, 0.8fr) auto;
  gap: 10px;
  align-items: center;
  background: #fff;
}

.org-filter-input {
  min-width: 0;
}

.org-department-filter {
  position: relative;
  min-width: 0;
  z-index: 6;
}

.org-department-filter.active {
  z-index: 40;
}

.org-department-trigger {
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

.org-department-trigger.active,
.org-department-trigger:hover {
  border-color: #1677ff;
  background: #fbfdff;
  box-shadow: 0 0 0 3px rgba(22, 119, 255, 0.08);
}

.org-department-trigger-text {
  min-width: 0;
  flex: 1 1 auto;
  color: #86909c;
  font-size: 14px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.org-department-trigger-text.selected {
  color: #1f2329;
  font-weight: 500;
}

.org-department-selected-count {
  flex: 0 0 auto;
  padding: 1px 7px;
  border-radius: 999px;
  background: #eef6ff;
  color: #1677ff;
  font-size: 12px;
  font-weight: 700;
  line-height: 18px;
}

.org-department-arrow {
  position: relative;
  width: 18px;
  height: 18px;
  flex: 0 0 18px;
  color: #86909c;
}

.org-department-arrow::before {
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

.org-department-arrow.expanded::before {
  transform: rotate(225deg);
}

.org-department-panel {
  position: absolute;
  top: calc(100% + 6px);
  left: 0;
  right: 0;
  z-index: 40;
  min-width: 340px;
  border: 1px solid #dbe7f7;
  border-radius: 8px;
  background: #fff;
  box-shadow: 0 16px 38px rgba(31, 35, 41, 0.14);
  overflow: hidden;
}

.org-department-search {
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

.org-department-select-tree {
  max-height: 280px;
  overflow: auto;
  overscroll-behavior: contain;
}

.org-department-option-row {
  position: relative;
  min-width: 0;
  min-height: 38px;
  border-top: 1px solid #f2f3f5;
  display: grid;
  grid-template-columns: minmax(0, 1fr) 38px;
  align-items: center;
}

.org-department-option:hover,
.org-department-option-row:hover .org-department-option {
  background: #f7fbff;
}

.org-department-option {
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

.org-department-option-row.depth-2 .org-department-option {
  padding-left: 28px;
}

.org-department-option-row.depth-3 .org-department-option {
  padding-left: 44px;
}

.org-department-option-row.depth-4 .org-department-option,
.org-department-option-row.depth-5 .org-department-option {
  padding-left: 60px;
}

.org-department-option-chevron {
  position: relative;
  width: 16px;
  height: 16px;
  flex: 0 0 16px;
  border-radius: 4px;
  color: #8a94a6;
}

.org-department-option-chevron::before {
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

.org-department-option-chevron.expanded::before {
  transform: rotate(45deg);
}

.org-department-option-chevron.placeholder::before {
  display: none;
}

.org-department-option-name {
  min-width: 0;
  flex: 1 1 auto;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 13px;
  font-weight: 700;
}

.org-department-option-count {
  flex: 0 0 auto;
  color: #86909c;
  font-size: 12px;
}

.org-department-check {
  justify-self: center;
  width: 18px;
  height: 18px;
  min-height: 18px;
  margin: 0 10px;
  padding: 0;
  border: 1px solid #c9d3e2;
  border-radius: 4px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: #fff;
  color: #fff;
  font-size: 12px;
  font-weight: 800;
  line-height: 16px;
}

.org-department-check.checked,
.org-department-check.indeterminate {
  border-color: #1677ff;
  background: #1677ff;
}

.org-department-empty {
  padding: 18px 8px;
  color: #86909c;
  font-size: 13px;
  text-align: center;
}

.org-department-actions {
  padding: 10px 12px;
  border-top: 1px solid #eef1f6;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
  background: #fbfcff;
}

.org-empty {
  padding: 48px 16px;
}

.org-tree-table {
  margin: 14px;
  border: 1px solid #e5e6eb;
  border-radius: 4px;
  background: #fff;
  overflow: hidden;
}

.org-tree-table-head {
  min-height: 38px;
  padding: 0 14px;
  border-bottom: 1px solid #e5e6eb;
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  align-items: center;
  background: #fff;
  color: #86909c;
  font-size: 13px;
  font-weight: 700;
}

.org-tree-list {
  display: grid;
  gap: 0;
  background: #fff;
}

.org-tree-row {
  position: relative;
  min-width: 0;
}

.org-department-row {
  min-height: 40px;
  padding: 0 14px;
  border-bottom: 1px solid #edf0f5;
  border-radius: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  background: #fafafa;
  cursor: default;
}

.org-department-row.expandable {
  cursor: pointer;
}

.org-department-row.expandable:hover {
  background: #f7fbff;
}

.org-department-title {
  min-width: 0;
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.org-tree-row.depth-2 .org-department-title {
  padding-left: 28px;
}

.org-tree-row.depth-3 .org-department-title {
  padding-left: 56px;
}

.org-tree-row.depth-4 .org-department-title {
  padding-left: 84px;
}

.org-tree-row.depth-5 .org-department-title {
  padding-left: 112px;
}

.org-tree-chevron {
  position: relative;
  width: 18px;
  height: 18px;
  flex: 0 0 18px;
  border-radius: 4px;
  color: #8a94a6;
}

.org-tree-chevron::before {
  content: "";
  position: absolute;
  left: 6px;
  top: 5px;
  width: 6px;
  height: 6px;
  border-right: 1.5px solid currentColor;
  border-bottom: 1.5px solid currentColor;
  transform: rotate(-45deg);
  transform-origin: center;
  transition: transform 0.16s ease, color 0.16s ease;
}

.org-tree-chevron.expanded::before {
  transform: rotate(45deg);
}

.org-tree-chevron.placeholder::before {
  display: none;
}

.org-department-row.expandable:hover .org-tree-chevron {
  background: #eef6ff;
  color: #1677ff;
}

.org-department-name {
  min-width: 0;
  color: #1f2329;
  font-size: 14px;
  font-weight: 700;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.org-employee-list {
  display: grid;
  background: #fff;
}

.org-employee-row {
  min-width: 0;
  min-height: 50px;
  padding: 8px 14px;
  border-width: 0 0 1px;
  border-style: solid;
  border-color: #edf0f5;
  border-radius: 0;
  display: grid;
  grid-template-columns: minmax(260px, 1fr) minmax(360px, 1fr) minmax(220px, 0.9fr) 76px;
  align-items: center;
  gap: 14px;
  background: #fff;
}

.org-employee-row:hover {
  background: #f7fbff;
}

.org-tree-row.depth-2 .org-person-cell {
  padding-left: 28px;
}

.org-tree-row.depth-3 .org-person-cell {
  padding-left: 56px;
}

.org-tree-row.depth-4 .org-person-cell {
  padding-left: 84px;
}

.org-tree-row.depth-5 .org-person-cell {
  padding-left: 112px;
}

.org-employee-meta {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 8px;
  color: #4e5969;
  font-size: 12px;
}

.org-employee-meta text {
  min-width: 0;
  max-width: 160px;
  padding: 3px 7px;
  border-radius: 4px;
  background: #f5f7fa;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.org-flow-summary {
  min-width: 0;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.org-flow-summary-mobile {
  display: none;
}

.org-flow-summary view {
  min-width: 0;
  display: grid;
  gap: 3px;
}

.org-flow-summary text {
  color: #86909c;
  font-size: 12px;
}

.org-flow-summary strong {
  color: #1f2329;
  font-size: 13px;
  font-weight: 700;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.org-mobile-meta {
  display: grid;
  gap: 6px;
  margin-top: 12px;
  color: #4e5969;
  font-size: 13px;
}

@media (max-width: 960px) {
  .org-page {
    max-width: none;
    padding: 16px;
  }

  .org-toolbar {
    grid-template-columns: repeat(2, 1fr);
  }

  .org-filter-shell {
    border-top: 1px solid #f2f3f5;
  }

  .org-filter-toggle {
    width: 100%;
    height: 44px;
    min-height: 44px;
    margin: 0;
    padding: 0 16px;
    border: 0;
    border-radius: 0;
    border-bottom: 1px solid #f2f3f5;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    background: #fff;
    color: #1f2329;
    line-height: 44px;
    text-align: left;
  }

  .org-filter-toggle-title {
    min-width: 0;
    flex: 1 1 auto;
    font-size: 14px;
    font-weight: 800;
  }

  .org-filter-count {
    flex: 0 0 auto;
    padding: 2px 8px;
    border-radius: 999px;
    background: #eef6ff;
    color: #1677ff;
    font-size: 12px;
    font-weight: 700;
    line-height: 18px;
  }

  .org-filter-arrow {
    position: relative;
    width: 18px;
    height: 18px;
    flex: 0 0 18px;
    color: #86909c;
    transition: transform 0.18s cubic-bezier(0.22, 1, 0.36, 1);
  }

  .org-filter-arrow::before {
    content: "";
    position: absolute;
    left: 6px;
    top: 5px;
    width: 6px;
    height: 6px;
    border-right: 1.5px solid currentColor;
    border-bottom: 1.5px solid currentColor;
    transform: rotate(45deg);
  }

  .org-filter-arrow.expanded {
    transform: rotate(180deg);
  }

  .org-filter-shell.collapsed .org-filter-bar {
    display: none;
  }

  .org-filter-shell.expanded .org-filter-bar {
    animation: summaryFilterIn 0.18s cubic-bezier(0.22, 1, 0.36, 1);
  }

  .org-filter-bar {
    border-top: 0;
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .org-filter-bar > .dt-btn {
    justify-self: flex-start;
  }

  .org-create-grid {
    grid-template-columns: 1fr;
  }

  .org-wide-field {
    grid-column: auto;
  }

  .org-table-shell {
    display: none;
  }

  .org-page > .page-head,
  .org-toolbar,
  .org-scope-hint {
    display: none;
  }

  .org-page {
    padding-top: 12px;
  }

  .org-filter-bar {
    padding: 12px;
    grid-template-columns: minmax(0, 1fr);
  }

  .org-filter-bar > .dt-btn {
    width: 100%;
  }

  .org-department-panel {
    position: static;
    min-width: 0;
    width: 100%;
    margin-top: 6px;
    box-shadow: none;
  }

  .org-department-select-tree {
    max-height: 220px;
  }

  .org-tree-table {
    margin: 12px;
  }

  .org-tree-table-head {
    grid-template-columns: minmax(0, 1fr);
  }

  .org-tree-table-head text:nth-child(n + 2) {
    display: none;
  }

  .org-department-row {
    grid-template-columns: minmax(0, 1fr) auto;
    gap: 10px;
  }

  .org-tree-row.depth-2 .org-department-title {
    padding-left: 18px;
  }

  .org-tree-row.depth-3 .org-department-title {
    padding-left: 32px;
  }

  .org-tree-row.depth-4 .org-department-title,
  .org-tree-row.depth-5 .org-department-title {
    padding-left: 46px;
  }

  .org-employee-row {
    grid-template-columns: minmax(0, 1fr) auto;
    grid-template-areas:
    "person action";
    align-items: start;
    gap: 8px 10px;
    padding: 10px 12px;
  }

  .org-employee-row .org-person-cell {
    grid-area: person;
  }

  .org-employee-row .org-avatar {
    width: 34px;
    height: 34px;
    border-radius: 10px;
    font-size: 13px;
  }

  .org-employee-row .org-person-main {
    gap: 2px;
  }

  .org-employee-row .org-person-title-line {
    display: flex;
    align-items: center;
    gap: 6px;
    max-width: 100%;
  }

  .org-employee-row .user-name {
    display: block;
    flex: 0 1 auto;
    min-width: 0;
    font-size: 14px;
    line-height: 18px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .org-employee-row .org-account {
    display: none;
  }

  .org-employee-row .org-position-inline {
    display: inline-flex;
    align-items: center;
    max-width: 96px;
    padding: 1px 6px;
    font-size: 11px;
    line-height: 16px;
  }

  .org-employee-meta {
    display: none;
  }

  .org-employee-meta text {
    max-width: 100%;
    min-height: 20px;
    padding: 2px 6px;
    border-radius: 5px;
    font-size: 11px;
    line-height: 16px;
  }

  .org-tree-row.depth-2 .org-person-cell {
    padding-left: 18px;
  }

  .org-tree-row.depth-3 .org-person-cell {
    padding-left: 32px;
  }

  .org-tree-row.depth-4 .org-person-cell,
  .org-tree-row.depth-5 .org-person-cell {
    padding-left: 46px;
  }

  .org-employee-row > .org-flow-summary {
    display: none;
  }

  .org-employee-row .org-flow-summary-mobile {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 5px;
    margin-top: 4px;
  }

  .org-flow-summary view {
    display: inline-flex;
    align-items: center;
    min-width: 0;
    max-width: 100%;
    padding: 1px 6px;
    border-radius: 999px;
    background: #f2f6ff;
    gap: 3px;
  }

  .org-flow-summary text {
    font-size: 11px;
    line-height: 16px;
  }

  .org-flow-label-desktop {
    display: none;
  }

  .org-flow-label-mobile {
    display: inline;
  }

  .org-flow-summary strong {
    min-width: 0;
    max-width: 88px;
    color: #4e5969;
    font-size: 11px;
    line-height: 16px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}
</style>
