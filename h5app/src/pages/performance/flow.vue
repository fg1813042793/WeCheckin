<script setup lang="ts">
import type { PerformanceUser } from '@/types/dingtalk-h5'
import { computed, reactive, ref, watch } from 'vue'
import { listUsers, updateUser } from '@/api/dingtalk-h5'
import { useDingtalkAuthStore } from '@/stores'
import { useAppContentStore } from '@/stores/appContent'
import { departmentLevelsFromEntity, departmentPathFromEntity } from '@/utils/departments'
import PerformanceAdaptiveSelect from './components/PerformanceAdaptiveSelect.vue'

interface FlowUserNode {
  key: string
  name: string
  count: number
  userIds: string[]
  childMap?: Map<string, FlowUserNode>
  children: FlowUserNode[]
  users: PerformanceUser[]
}

interface FlowUserRow {
  type: 'department' | 'employee'
  key: string
  depth: number
  name?: string
  count?: number
  userIds?: string[]
  expandable?: boolean
  expanded?: boolean
  user?: PerformanceUser
}

type RelationPickerType = 'manager' | 'hrbp'

interface SelectOption {
  value: string
  label: string
}

function firstText(...values: Array<unknown>) {
  for (const value of values) {
    const text = String(value || '').trim()
    if (text) {
      return text
    }
  }
  return ''
}

function sameUserId(left: unknown, right: unknown) {
  const leftText = String(left || '').trim()
  const rightText = String(right || '').trim()
  return Boolean(leftText && rightText && leftText === rightText)
}

function readInputValue(event: unknown) {
  const candidate = event as { detail?: { value?: unknown }, target?: { value?: unknown } }
  return String(candidate?.detail?.value ?? candidate?.target?.value ?? event ?? '')
}

function resolveMobilePage() {
  try {
    const info = uni.getSystemInfoSync()
    const width = Number(info.windowWidth || info.screenWidth || 0)
    const deviceType = String(info.deviceType || '').toLowerCase()
    const platform = String(info.platform || '').toLowerCase()
    return (width > 0 && width <= 768) || deviceType === 'phone' || (['android', 'ios'].includes(platform) && width <= 1024)
  }
  catch {
    return false
  }
}

const auth = useDingtalkAuthStore()
const appContent = useAppContentStore()
const users = ref<PerformanceUser[]>([])
const loading = ref(false)
const isMobilePage = ref(resolveMobilePage())
const orgExpandedDeptKeys = ref(new Set<string>())
const managerPickerExpandedKeys = ref(new Set<string>())
const hrbpPickerExpandedKeys = ref(new Set<string>())
const activeRelationPicker = ref<RelationPickerType | ''>('')
const managerRelationKeyword = ref('')
const hrbpRelationKeyword = ref('')
const selectedOrgUser = ref<PerformanceUser | null>(null)
const orgFiltersCollapsed = ref(resolveMobilePage())
const orgFilters = reactive({
  employeeKeyword: '',
  managerStatus: '',
  hrbpStatus: '',
})
const orgConfigForm = reactive<Partial<PerformanceUser> & { id: string }>({
  id: '',
  name: '',
  password: '',
  position: '',
  department: '',
  departmentLevel1: '',
  departmentLevel2: '',
  departmentLevel3: '',
  departmentLevel4: '',
  managerId: '',
  hrbpId: '',
})
const listTitle = '流程制定'
const listDesc = '按权限查看流程人员、直属上级和 HRBP 关系'
const emptyTitle = '当前没有可见人员'
const filteredUsers = computed(() => {
  const keyword = orgFilters.employeeKeyword.trim().toLowerCase()
  return users.value.filter((user) => {
    const haystack = [
      user.name,
      user.id,
      user.position,
      user.department,
      user.departmentLevel1,
      user.departmentLevel2,
      user.departmentLevel3,
      user.departmentLevel4,
      ...(Array.isArray(user.departmentLevels) ? user.departmentLevels : []),
    ].filter(Boolean).join(' ').toLowerCase()
    if (keyword && !haystack.includes(keyword)) {
      return false
    }
    if (orgFilters.managerStatus === 'set' && !user.managerId) {
      return false
    }
    if (orgFilters.managerStatus === 'unset' && user.managerId) {
      return false
    }
    if (orgFilters.hrbpStatus === 'set' && !user.hrbpId) {
      return false
    }
    if (orgFilters.hrbpStatus === 'unset' && user.hrbpId) {
      return false
    }
    return true
  })
})
const orgStats = computed(() => [
  ['可见人员', filteredUsers.value.length],
  ['已配置上级', filteredUsers.value.filter(user => user.managerId).length],
  ['已配置HRBP', filteredUsers.value.filter(user => user.hrbpId).length],
  ['已设置岗位', filteredUsers.value.filter(user => user.position).length],
])
const orgFilterCount = computed(() => [
  orgFilters.employeeKeyword.trim(),
  orgFilters.managerStatus,
  orgFilters.hrbpStatus,
].filter(Boolean).length)
const orgDepartmentTree = computed(() => buildFlowUserTree(filteredUsers.value))
const orgSearchActive = computed(() => Boolean(orgFilters.employeeKeyword.trim()))
const effectiveOrgExpandedDeptKeys = computed(() => {
  if (!orgSearchActive.value) {
    return orgExpandedDeptKeys.value
  }
  return new Set(collectFlowUserKeys(orgDepartmentTree.value))
})
const orgTreeRows = computed(() => {
  return flattenFlowUserTree(orgDepartmentTree.value, effectiveOrgExpandedDeptKeys.value)
})
const visibleOrgTreeRows = computed(() => {
  return orgTreeRows.value
})
const orgTreeCountText = computed(() => `${visibleOrgTreeRows.value.length} 行`)
const managerPickerTree = computed(() => buildFlowUserTree(users.value.filter(user => user.id)))
const hrbpPickerTree = computed(() => buildFlowUserTree(users.value.filter(user => user.id)))
const managerPickerRows = computed(() => {
  return filteredRelationPickerRows('manager')
})
const hrbpPickerRows = computed(() => {
  return filteredRelationPickerRows('hrbp')
})
const managerStatusOptions: SelectOption[] = [
  { value: '', label: '全部上级状态' },
  { value: 'set', label: '已设置上级' },
  { value: 'unset', label: '未设置上级' },
]
const hrbpStatusOptions: SelectOption[] = [
  { value: '', label: '全部HRBP状态' },
  { value: 'set', label: '已设置HRBP' },
  { value: 'unset', label: '未设置HRBP' },
]

async function loadFlowPageData() {
  loading.value = true
  try {
    const res = await listUsers()
    users.value = (res.data || []).map(user => ({ ...user, password: '' }))
  }
  finally {
    loading.value = false
  }
}

function departmentText(user: PerformanceUser) {
  return departmentPathFromEntity(user)
}

function flowDepartmentLevels(user: PerformanceUser) {
  const levels = departmentLevelsFromEntity(user)
  return levels.length > 0 ? levels : ['未设置部门']
}

function ensureFlowUserNode(map: Map<string, FlowUserNode>, key: string, name: string) {
  if (!map.has(key)) {
    map.set(key, {
      key,
      name,
      count: 0,
      userIds: [],
      childMap: new Map(),
      children: [],
      users: [],
    })
  }
  return map.get(key) as FlowUserNode
}

function finalizeFlowUserNodes(nodes: FlowUserNode[]): FlowUserNode[] {
  return nodes
    .sort((left, right) => left.name.localeCompare(right.name, 'zh-CN'))
    .map(node => ({
      key: node.key,
      name: node.name,
      count: node.count,
      userIds: [...new Set(node.userIds.filter(Boolean))],
      children: finalizeFlowUserNodes([...(node.childMap?.values() || [])]),
      users: node.users.slice().sort((left, right) => [left.department, left.name, left.id].filter(Boolean).join('\0').localeCompare([right.department, right.name, right.id].filter(Boolean).join('\0'), 'zh-CN')),
    }))
}

function buildFlowUserTree(users: PerformanceUser[] = []) {
  const root = new Map<string, FlowUserNode>()
  for (const user of users) {
    const levels = flowDepartmentLevels(user)
    let currentMap = root
    let currentNode: FlowUserNode | null = null
    for (const [index, level] of levels.entries()) {
      const parentKey = currentNode?.key || 'root'
      currentNode = ensureFlowUserNode(currentMap, `${parentKey}/l${index + 1}:${level}`, level)
      currentNode.count += 1
      if (user.id) {
        currentNode.userIds.push(user.id)
      }
      currentMap = currentNode.childMap as Map<string, FlowUserNode>
    }
    if (currentNode) {
      currentNode.users.push(user)
    }
  }
  return finalizeFlowUserNodes([...root.values()])
}

function flattenFlowUserTree(nodes: FlowUserNode[], expandedKeys: Set<string>, depth = 1): FlowUserRow[] {
  const rows: FlowUserRow[] = []
  for (const node of nodes) {
    const hasChildren = node.children.length > 0 || node.users.length > 0
    const expanded = expandedKeys.has(node.key)
    rows.push({
      type: 'department',
      key: node.key,
      depth,
      name: node.name,
      count: node.count,
      userIds: node.userIds,
      expandable: hasChildren,
      expanded,
    })
    if (!expanded) {
      continue
    }
    for (const user of node.users) {
      rows.push({ type: 'employee', key: `${node.key}/user:${user.id}`, depth: depth + 1, user })
    }
    rows.push(...flattenFlowUserTree(node.children, expandedKeys, depth + 1))
  }
  return rows
}

function collectFlowUserKeys(nodes: FlowUserNode[]) {
  const keys: string[] = []
  for (const node of nodes) {
    if (node.children.length > 0 || node.users.length > 0) {
      keys.push(node.key)
    }
    keys.push(...collectFlowUserKeys(node.children))
  }
  return keys
}

function relationPickerSearchKeyword(type: RelationPickerType) {
  const keyword = type === 'manager' ? managerRelationKeyword.value : hrbpRelationKeyword.value
  return keyword.trim().toLowerCase()
}

function relationPickerSearchUser(user: PerformanceUser, keyword: string) {
  if (!keyword) {
    return true
  }
  const haystack = [
    user.name,
    user.id,
    user.account,
    user.userid,
  ].filter(Boolean).join(' ').toLowerCase()
  return haystack.includes(keyword)
}

function filterRelationPickerTree(nodes: FlowUserNode[], keyword: string): FlowUserNode[] {
  if (!keyword) {
    return nodes
  }
  return nodes.flatMap((node) => {
    const children = filterRelationPickerTree(node.children, keyword)
    const matchedUsers = node.users.filter(user => relationPickerSearchUser(user, keyword))
    const userIds = [
      ...matchedUsers.map(user => user.id),
      ...children.flatMap(child => child.userIds),
    ].filter(Boolean)
    const count = userIds.length
    if (count <= 0) {
      return []
    }
    return [{
      key: node.key,
      name: node.name,
      count,
      userIds,
      children,
      users: matchedUsers,
    }]
  })
}

function filteredRelationPickerRows(type: RelationPickerType) {
  const keyword = relationPickerSearchKeyword(type)
  const tree = type === 'manager' ? managerPickerTree.value : hrbpPickerTree.value
  const filteredTree = filterRelationPickerTree(tree, keyword)
  const expandedKeys = keyword
    ? new Set(collectFlowUserKeys(filteredTree))
    : relationExpandedRef(type).value
  return flattenFlowUserTree(filteredTree, expandedKeys)
}

function resetRelationPickerSearch(type?: RelationPickerType) {
  if (!type || type === 'manager') {
    managerRelationKeyword.value = ''
  }
  if (!type || type === 'hrbp') {
    hrbpRelationKeyword.value = ''
  }
}

function resetOrgFilters() {
  orgFilters.employeeKeyword = ''
  orgFilters.managerStatus = ''
  orgFilters.hrbpStatus = ''
}

function setOrgEmployeeKeyword(value: unknown) {
  orgFilters.employeeKeyword = String(value ?? '')
}

function toggleOrgDepartment(row: FlowUserRow) {
  if (!row.expandable) {
    return
  }
  const next = new Set(orgExpandedDeptKeys.value)
  if (next.has(row.key)) {
    next.delete(row.key)
  }
  else {
    next.add(row.key)
  }
  orgExpandedDeptKeys.value = next
}

function relationPickerText(id: unknown, emptyText: string) {
  const key = String(id || '').trim()
  if (!key) {
    return emptyText
  }
  const user = users.value.find(item => item.id === key)
  return user ? (user.name || user.id) : key
}

function userName(id: unknown) {
  const key = String(id || '').trim()
  if (!key) {
    return '-'
  }
  const matched = users.value.find(user => user.id === key)
  if (matched?.name) {
    return matched.name
  }
  if (sameUserId(auth.user?.id, key)) {
    return auth.user?.name || key
  }
  return key
}

function upsertUser(user: PerformanceUser) {
  if (!user.id) {
    return
  }
  const index = users.value.findIndex(item => item.id === user.id)
  if (index >= 0) {
    users.value[index] = { ...users.value[index], ...user, password: '' }
    return
  }
  users.value.push({ ...user, password: '' })
}

function relationExpandedRef(type: RelationPickerType) {
  return type === 'manager' ? managerPickerExpandedKeys : hrbpPickerExpandedKeys
}

function toggleRelationPicker(type: RelationPickerType) {
  if (activeRelationPicker.value === type) {
    activeRelationPicker.value = ''
    resetRelationPickerSearch(type)
    return
  }
  activeRelationPicker.value = type
  resetRelationPickerSearch(type)
}

function relationPickerTriggerClass(type: RelationPickerType) {
  return ['relation-picker-trigger', activeRelationPicker.value === type ? 'active' : ''].filter(Boolean).join(' ')
}

function toggleRelationDepartment(type: RelationPickerType, row: FlowUserRow) {
  if (!row.expandable || relationPickerSearchKeyword(type)) {
    return
  }
  const target = relationExpandedRef(type)
  const next = new Set(target.value)
  if (next.has(row.key)) {
    next.delete(row.key)
  }
  else {
    next.add(row.key)
  }
  target.value = next
}

function relationPickerUserClass(type: RelationPickerType, user: PerformanceUser) {
  const selectedId = type === 'manager' ? orgConfigForm.managerId : orgConfigForm.hrbpId
  return ['relation-picker-user', selectedId === user.id ? 'selected' : ''].filter(Boolean).join(' ')
}

function relationPickerRowStyle(row: FlowUserRow) {
  const indent = Math.max(0, row.depth - 1) * 16
  return { '--relation-picker-indent': `${indent}px` }
}

function selectRelationUser(type: RelationPickerType, id: string) {
  if (type === 'manager') {
    orgConfigForm.managerId = id
  }
  else {
    orgConfigForm.hrbpId = id
  }
  activeRelationPicker.value = ''
  resetRelationPickerSearch(type)
}

function openOrgConfig(user: PerformanceUser) {
  selectedOrgUser.value = user
  Object.assign(orgConfigForm, {
    id: user.id,
    name: user.name || user.id,
    password: '',
    position: user.position || '',
    department: user.department || '',
    departmentLevel1: user.departmentLevel1 || '',
    departmentLevel2: user.departmentLevel2 || '',
    departmentLevel3: user.departmentLevel3 || '',
    departmentLevel4: user.departmentLevel4 || '',
    departmentLevels: Array.isArray(user.departmentLevels) ? user.departmentLevels : [],
    managerId: user.managerId || '',
    hrbpId: user.hrbpId || '',
  })
  activeRelationPicker.value = ''
  resetRelationPickerSearch()
  managerPickerExpandedKeys.value = new Set()
  hrbpPickerExpandedKeys.value = new Set()
}

function closeOrgConfig() {
  selectedOrgUser.value = null
  activeRelationPicker.value = ''
  resetRelationPickerSearch()
}

async function saveOrgConfig() {
  if (!orgConfigForm.id) {
    return
  }
  const res = await updateUser(orgConfigForm.id, { ...orgConfigForm, id: orgConfigForm.id })
  if (Array.isArray(res.data?.users)) {
    users.value = res.data.users.map(user => ({ ...user, password: '' }))
  }
  else if (res.data?.user) {
    upsertUser(res.data.user)
  }
  else {
    upsertUser(orgConfigForm as PerformanceUser)
  }
  closeOrgConfig()
  uni.showToast({
    title: '已保存配置',
    icon: 'success',
  })
}

watch(orgDepartmentTree, (nodes) => {
  const availableKeys = new Set(collectFlowUserKeys(nodes))
  const next = new Set<string>()
  for (const key of orgExpandedDeptKeys.value) {
    if (availableKeys.has(key)) {
      next.add(key)
    }
  }
  orgExpandedDeptKeys.value = next
}, { immediate: true })

watch(
  () => [appContent.currentKey, appContent.refreshTick],
  () => {
    if (appContent.currentKey === 'performance:org') {
      void loadFlowPageData()
    }
  },
  { immediate: true },
)

watch(
  () => [orgFilters.employeeKeyword, orgFilters.managerStatus, orgFilters.hrbpStatus],
  () => {
    selectedOrgUser.value = null
  },
)
</script>

<template>
  <view class="performance-page">
    <view class="page-head">
      <view v-if="!isMobilePage" class="page-head__copy">
        <text class="page-title">
          {{ listTitle }}
        </text>
        <text class="page-desc">
          {{ listDesc }}
        </text>
      </view>
    </view>

    <view v-if="loading" class="panel performance-loading-panel">
      <u-loading mode="circle" />
      <text>加载中...</text>
    </view>

    <template v-else>
      <view class="org-toolbar">
        <view v-for="[label, value] in orgStats" :key="label" class="org-stat">
          <text class="org-stat-value">
            {{ value }}
          </text>
          <text class="org-stat-label">
            {{ label }}
          </text>
        </view>
      </view>
      <view class="panel org-list-panel">
        <view class="panel-head">
          <view>
            <text class="panel-title">
              流程人员
            </text>
            <text class="org-scope-hint">
              已按权限过滤，数据来自用户表
            </text>
          </view>
          <text class="count-pill">
            {{ orgTreeCountText }}
          </text>
        </view>
        <view class="summary-filter-shell org-filter-shell" :class="{ collapsed: orgFiltersCollapsed, expanded: !orgFiltersCollapsed }">
          <u-button custom-class="summary-filter-toggle" @click="orgFiltersCollapsed = !orgFiltersCollapsed">
            <text class="summary-filter-toggle-title">
              筛选条件
            </text>
            <text v-if="orgFilterCount > 0" class="summary-filter-count">
              已选 {{ orgFilterCount }}
            </text>
            <text class="summary-filter-arrow" :class="{ expanded: !orgFiltersCollapsed }" />
          </u-button>
          <view class="filters org-filter-bar">
            <u-input
              :model-value="orgFilters.employeeKeyword"
              custom-class="field-input"
              :border="true"
              placeholder="搜索员工姓名/账号"
              @input="setOrgEmployeeKeyword(readInputValue($event))"
            />
            <PerformanceAdaptiveSelect
              v-model="orgFilters.managerStatus"
              custom-class="field-select"
              title="筛选上级状态"
              :options="managerStatusOptions"
              :border="true"
            />
            <PerformanceAdaptiveSelect
              v-model="orgFilters.hrbpStatus"
              custom-class="field-select"
              title="筛选HRBP状态"
              :options="hrbpStatusOptions"
              :border="true"
            />
            <u-button custom-class="dt-btn dt-btn-light" @click="resetOrgFilters">
              重置
            </u-button>
          </view>
        </view>
        <view v-if="orgTreeRows.length === 0" class="table-empty-cell org-empty">
          {{ emptyTitle }}
        </view>
        <view v-else class="org-tree-list">
          <view
            v-for="row in visibleOrgTreeRows"
            :key="row.key"
            class="org-tree-row"
            :class="[row.type, `depth-${row.depth}`]"
          >
            <view v-if="row.type === 'department'" class="org-department-row" :class="{ expanded: row.expanded }" @click="toggleOrgDepartment(row)">
              <view class="org-department-title">
                <text class="relation-picker-chevron" :class="{ expanded: row.expanded, placeholder: !row.expandable }" />
                <text class="org-department-name">
                  {{ row.name }}
                </text>
              </view>
              <text class="count-pill muted">
                {{ row.count || 0 }} 人
              </text>
            </view>
            <view v-else-if="row.user" class="org-employee-row">
              <view class="org-user-main">
                <text class="org-user-name">
                  {{ row.user.name || row.user.id }}
                </text>
                <text class="org-user-meta">
                  {{ departmentText(row.user) }} · {{ row.user.position || '未设置岗位' }}
                </text>
              </view>
              <view class="org-flow-summary">
                <view>
                  <text>直属上级</text>
                  <strong>{{ userName(row.user.managerId) }}</strong>
                </view>
                <view>
                  <text>HRBP</text>
                  <strong>{{ userName(row.user.hrbpId) }}</strong>
                </view>
              </view>
              <u-button
                v-if="auth.hasButtonPermission('dingtalk_h5:button:user:config')"
                custom-class="dt-btn dt-btn-light small org-config-btn"
                @click="openOrgConfig(row.user)"
              >
                配置
              </u-button>
            </view>
          </view>
        </view>
      </view>

      <view v-if="selectedOrgUser" class="org-config-modal-mask" @click.self="closeOrgConfig">
        <view class="org-config-modal">
          <view class="org-config-modal-head">
            <view>
              <text class="org-config-title">
                配置流程审批人
              </text>
              <text class="org-config-subtitle">
                {{ selectedOrgUser.name || selectedOrgUser.id }} · {{ departmentText(selectedOrgUser) }}
              </text>
            </view>
            <u-button custom-class="process-modal-close" @click="closeOrgConfig">
              ×
            </u-button>
          </view>

          <view class="org-config-user">
            <view class="org-avatar">
              {{ firstText(selectedOrgUser.name, selectedOrgUser.id).slice(0, 1).toUpperCase() }}
            </view>
            <view>
              <text class="org-user-name">
                {{ selectedOrgUser.name || selectedOrgUser.id }}
              </text>
              <!-- <text class="org-user-meta">
                {{ [selectedOrgUser.id, selectedOrgUser.position || '未设置岗位'].filter(Boolean).join(' · ') }}
              </text> -->
            </view>
          </view>

          <view class="org-config-form">
            <view class="org-form-field relation-picker-field" :class="{ active: activeRelationPicker === 'manager' }">
              <text>直属上级</text>
              <u-button :custom-class="relationPickerTriggerClass('manager')" @click="toggleRelationPicker('manager')">
                <text class="relation-picker-value">
                  {{ relationPickerText(orgConfigForm.managerId, '无直属上级') }}
                </text>
                <text class="relation-picker-arrow" :class="{ expanded: activeRelationPicker === 'manager' }" />
              </u-button>
              <view v-if="activeRelationPicker === 'manager'" class="relation-picker-panel">
                <view class="relation-picker-search-wrap">
                  <u-input
                    v-model="managerRelationKeyword"
                    custom-class="relation-picker-search"
                    placeholder="搜索用户名"
                    :border="true"
                    :clearable="true"
                  />
                </view>
                <u-button custom-class="relation-picker-clear" @click="selectRelationUser('manager', '')">
                  无直属上级
                </u-button>
                <view v-if="managerPickerRows.length > 0" class="relation-picker-tree">
                  <view v-for="row in managerPickerRows" :key="`manager-${row.key}`" class="relation-picker-row" :class="[row.type, `depth-${row.depth}`]" :style="relationPickerRowStyle(row)">
                    <u-button v-if="row.type === 'department'" custom-class="relation-picker-dept" @click="toggleRelationDepartment('manager', row)">
                      <text class="relation-picker-chevron" :class="{ expanded: row.expanded, placeholder: !row.expandable }" />
                      <text class="relation-picker-dept-name">
                        {{ row.name }}
                      </text>
                      <text class="relation-picker-count">
                        {{ row.count || 0 }} 人
                      </text>
                    </u-button>
                    <u-button v-else-if="row.user" :custom-class="relationPickerUserClass('manager', row.user)" @click="selectRelationUser('manager', row.user.id)">
                      <text class="relation-picker-radio" :class="{ checked: orgConfigForm.managerId === row.user.id }" />
                      <view class="relation-picker-user-main">
                        <text class="relation-picker-user-name">
                          {{ row.user.name || row.user.id }}
                        </text>
                      </view>
                    </u-button>
                  </view>
                </view>
                <view v-else class="relation-picker-empty">
                  没有匹配的用户
                </view>
              </view>
            </view>

            <view class="org-form-field relation-picker-field" :class="{ active: activeRelationPicker === 'hrbp' }">
              <text>HRBP</text>
              <u-button :custom-class="relationPickerTriggerClass('hrbp')" @click="toggleRelationPicker('hrbp')">
                <text class="relation-picker-value">
                  {{ relationPickerText(orgConfigForm.hrbpId, '无HRBP') }}
                </text>
                <text class="relation-picker-arrow" :class="{ expanded: activeRelationPicker === 'hrbp' }" />
              </u-button>
              <view v-if="activeRelationPicker === 'hrbp'" class="relation-picker-panel">
                <view class="relation-picker-search-wrap">
                  <u-input
                    v-model="hrbpRelationKeyword"
                    custom-class="relation-picker-search"
                    placeholder="搜索用户名"
                    :border="true"
                    :clearable="true"
                  />
                </view>
                <u-button custom-class="relation-picker-clear" @click="selectRelationUser('hrbp', '')">
                  无HRBP
                </u-button>
                <view v-if="hrbpPickerRows.length > 0" class="relation-picker-tree">
                  <view v-for="row in hrbpPickerRows" :key="`hrbp-${row.key}`" class="relation-picker-row" :class="[row.type, `depth-${row.depth}`]" :style="relationPickerRowStyle(row)">
                    <u-button v-if="row.type === 'department'" custom-class="relation-picker-dept" @click="toggleRelationDepartment('hrbp', row)">
                      <text class="relation-picker-chevron" :class="{ expanded: row.expanded, placeholder: !row.expandable }" />
                      <text class="relation-picker-dept-name">
                        {{ row.name }}
                      </text>
                      <text class="relation-picker-count">
                        {{ row.count || 0 }} 人
                      </text>
                    </u-button>
                    <u-button v-else-if="row.user" :custom-class="relationPickerUserClass('hrbp', row.user)" @click="selectRelationUser('hrbp', row.user.id)">
                      <text class="relation-picker-radio" :class="{ checked: orgConfigForm.hrbpId === row.user.id }" />
                      <view class="relation-picker-user-main">
                        <text class="relation-picker-user-name">
                          {{ row.user.name || row.user.id }}
                        </text>
                      </view>
                    </u-button>
                  </view>
                </view>
                <view v-else class="relation-picker-empty">
                  没有匹配的用户
                </view>
              </view>
            </view>
          </view>

          <view class="process-modal-actions">
            <u-button custom-class="dt-btn dt-btn-light" @click="closeOrgConfig">
              取消
            </u-button>
            <u-button custom-class="dt-btn dt-btn-primary" @click="saveOrgConfig">
              保存配置
            </u-button>
          </view>
        </view>
      </view>
    </template>
  </view>
</template>

<style lang="scss" scoped src="./components/performance-page.scss"></style>
