<template>
  <div class="admin-page workflow-org-approver-page">
    <el-card class="admin-card" shadow="never">
      <div class="admin-toolbar">
        <div class="admin-toolbar__left">
          <h2>组织审批身份设置</h2>
        </div>
        <div class="admin-toolbar__right">
          <el-button circle icon="Refresh" title="刷新" :loading="bootstrapLoading" @click="reloadAll" />
        </div>
      </div>

      <div class="org-approver-layout">
        <section class="org-panel org-panel--departments">
          <div class="org-panel__header">
            <div class="panel-title">
              <span class="panel-title__index">1</span>
              <strong>为谁配置</strong>
            </div>
            <span class="panel-title__count">{{ activeSubjectType === 'department' ? `${departmentCount} 个部门` : `${activeSubjectUserOptions.length} 名人员` }}</span>
          </div>
          <div class="org-panel__content">
            <el-radio-group v-model="activeSubjectType" class="subject-type-switch" @change="onSubjectTypeChange">
              <el-radio-button value="department">部门内员工</el-radio-button>
              <el-radio-button value="user">指定员工</el-radio-button>
            </el-radio-group>
            <template v-if="activeSubjectType === 'department'">
              <el-input
                v-model="deptKeyword"
                prefix-icon="Search"
                placeholder="搜索部门"
                clearable
                @input="filterDeptTree"
              />
              <div class="org-tree-wrap">
                <el-tree
                  ref="deptTreeRef"
                  v-loading="deptLoading"
                  :data="deptTreeData"
                  :props="{ label: 'name', children: 'children' }"
                  :filter-node-method="filterDeptNode"
                  node-key="id"
                  highlight-current
                  default-expand-all
                  empty-text="暂无部门"
                  @node-click="onDepartmentClick"
                />
              </div>
            </template>
            <template v-else>
              <el-input
                v-model="subjectUserKeyword"
                prefix-icon="Search"
                placeholder="搜索指定员工/手机号/部门"
                clearable
                @input="filterSubjectUserTree"
              />
              <div class="org-tree-wrap">
                <el-tree
                  ref="subjectUserTreeRef"
                  v-loading="userLoading"
                  :data="subjectUserTreeData"
                  :props="{ label: 'label', children: 'children', disabled: 'disabled' }"
                  :filter-node-method="filterOrgApproverUserNode"
                  node-key="key"
                  highlight-current
                  default-expand-all
                  empty-text="暂无可选员工"
                  @node-click="onSubjectUserClick"
                  class="subject-user-tree"
                >
                  <template #default="{ node, data }">
                    <span :class="['org-user-node', data.type === 'user' ? 'is-user' : 'is-dept']">
                      <span class="org-user-node__label">{{ node.label }}</span>
                      <span v-if="data.type === 'user' && data.mobile" class="org-user-node__meta">{{ data.mobile }}</span>
                    </span>
                  </template>
                </el-tree>
              </div>
            </template>
          </div>
        </section>

        <section class="org-panel org-panel--identities">
          <div class="org-panel__header">
            <div class="panel-title">
              <span class="panel-title__index">2</span>
              <strong>配置什么身份</strong>
            </div>
            <span class="panel-title__count">{{ identityOptions.length }} 项</span>
          </div>
          <div class="identity-table-wrap">
            <el-table
              v-loading="identityLoading"
              :data="identityOptions"
              height="100%"
              highlight-current-row
              empty-text="暂无可配置身份"
              :row-class-name="identityRowClassName"
              @row-click="onIdentityRowClick"
            >
              <el-table-column prop="name" label="名称" min-width="112" show-overflow-tooltip>
                <template #default="{ row }">
                  <span class="identity-name">{{ row.name }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="code" label="编码" min-width="136" show-overflow-tooltip>
                <template #default="{ row }">
                  <span class="identity-code">{{ row.code }}</span>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </section>

        <section class="org-panel org-panel--wide">
          <div class="org-panel__header org-panel__header--assignment">
            <div class="assignment-scope">
              <div class="panel-title">
                <span class="panel-title__index">3</span>
                <strong>由谁担任该身份</strong>
              </div>
              <div class="assignment-summary">
                <span class="assignment-summary__text" :title="assignmentSummary">{{ assignmentSummary }}</span>
                <el-tag v-if="scopeReady" size="small" :type="activeSubjectType === 'department' ? 'info' : 'warning'" effect="plain">
                  {{ activeSubjectType === 'department' ? '部门通用规则' : '优先于部门规则' }}
                </el-tag>
              </div>
            </div>
            <div class="assignment-actions">
              <el-button circle icon="Refresh" title="刷新处理人" :loading="assignmentsLoading" :disabled="!scopeReady" @click="loadAssignments" />
              <el-button v-if="canEdit" type="primary" icon="Check" :loading="saving" :disabled="!scopeReady" @click="saveAssignments">保存担任人员</el-button>
            </div>
          </div>

          <div v-if="scopeReady" class="assignment-body">
            <div class="assignment-picker">
              <div class="candidate-users__header">
                <strong>选择担任人员</strong>
                <span>{{ userOptions.length }} 人</span>
              </div>
              <el-input
                v-model="userKeyword"
                prefix-icon="Search"
                placeholder="搜索担任人员/手机号/部门"
                clearable
                @input="filterOrgApproverUserTree"
              />
              <div class="org-tree-wrap org-tree-wrap--users">
                <el-tree
                  ref="orgApproverUserTreeRef"
                  v-loading="userLoading || assignmentsLoading"
                  :data="orgApproverUserTreeData"
                  :props="{ label: 'label', children: 'children', disabled: 'disabled' }"
                  :filter-node-method="filterOrgApproverUserNode"
                  show-checkbox
                  check-on-click-node
                  node-key="key"
                  :default-checked-keys="orgApproverCheckedNodeKeys"
                  :default-expanded-keys="defaultExpandedUserTreeKeys"
                  empty-text="暂无可选用户"
                  @check="onOrgApproverUserCheck"
                  class="org-approver-user-tree"
                >
                  <template #default="{ node, data }">
                    <span :class="['org-user-node', data.type === 'user' ? 'is-user' : 'is-dept']">
                      <span class="org-user-node__label">{{ node.label }}</span>
                      <span v-if="data.type === 'user' && data.mobile" class="org-user-node__meta">{{ data.mobile }}</span>
                    </span>
                  </template>
                </el-tree>
              </div>
            </div>

            <div class="selected-users">
              <div class="selected-users__header">
                <strong>当前担任人员</strong>
                <span>{{ selectedUserRows.length }} 人</span>
              </div>
              <el-table :data="selectedUserRows" height="100%" empty-text="尚未指定担任人员" class="selected-users__table">
                <el-table-column type="index" label="顺序" width="60" />
                <el-table-column prop="name" label="用户" min-width="100" show-overflow-tooltip />
                <el-table-column prop="mobile" label="手机号" min-width="120" show-overflow-tooltip />
                <el-table-column label="操作" width="128">
                  <template #default="{ $index, row }">
                    <div class="selected-user-actions">
                      <el-button circle size="small" icon="ArrowUp" title="上移" :disabled="$index === 0" @click="moveSelectedUser($index, -1)" />
                      <el-button circle size="small" icon="ArrowDown" title="下移" :disabled="$index === selectedUserRows.length - 1" @click="moveSelectedUser($index, 1)" />
                      <el-button circle size="small" type="danger" plain icon="Delete" title="移除" @click="removeSelectedUser(row.id)" />
                    </div>
                  </template>
                </el-table-column>
              </el-table>
            </div>
          </div>
          <el-empty v-else description="请先选择配置对象和审批身份" />
        </section>
      </div>
    </el-card>
  </div>
</template>

<script lang="ts" setup>
import { computed, nextTick, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { adminApi } from '../../../api'
import { hasPerm } from '../../../utils/permission'
import type { WorkflowAssigneeUser, WorkflowOrgApproverIdentity } from '../types'

type DepartmentNode = {
  id: number
  name: string
  children?: DepartmentNode[]
}

type OrgApproverTreeNode = {
  key: string
  type: 'dept' | 'user'
  id?: number
  label: string
  mobile?: string
  pathText?: string
  disabled?: boolean
  children?: OrgApproverTreeNode[]
}

type OrgApproverAssignment = {
  subjectType: 'department' | 'user'
  subjectId: number
  subjectName?: string
  departmentId?: number
  identityCode: string
  userId: number
  userName?: string
  sort?: number
}

const canEdit = computed(() => hasPerm('admin:menu:workflow:org-approver:edit'))
const deptTreeRef = ref()
const subjectUserTreeRef = ref()
const orgApproverUserTreeRef = ref()
const deptLoading = ref(false)
const identityLoading = ref(false)
const userLoading = ref(false)
const assignmentsLoading = ref(false)
const saving = ref(false)
const deptTreeData = ref<DepartmentNode[]>([])
const identities = ref<WorkflowOrgApproverIdentity[]>([])
const userOptions = ref<WorkflowAssigneeUser[]>([])
const assignments = ref<OrgApproverAssignment[]>([])
const activeSubjectType = ref<'department' | 'user'>('department')
const activeDepartmentId = ref(0)
const activeSubjectUserId = ref(0)
const activeIdentityCode = ref('')
const selectedUserIds = ref<number[]>([])
const deptKeyword = ref('')
const subjectUserKeyword = ref('')
const userKeyword = ref('')

const bootstrapLoading = computed(() => deptLoading.value || identityLoading.value || userLoading.value)
const identityOptions = computed(() => identities.value.filter((item) => Number(item.status ?? 1) === 1))
const activeIdentity = computed(() => identityOptions.value.find((item) => item.code === activeIdentityCode.value))
const activeSubjectId = computed(() => activeSubjectType.value === 'department' ? activeDepartmentId.value : activeSubjectUserId.value)
const scopeReady = computed(() => activeSubjectId.value > 0 && !!activeIdentityCode.value)
const departmentCount = computed(() => countDepartments(deptTreeData.value))
const activeDepartmentPath = computed(() => (deptPathInfoMap().get(activeDepartmentId.value) || []).join(' / '))
const activeSubjectUserOptions = computed(() => userOptions.value.filter((user) => Number(user.status ?? 1) === 1))
const activeSubjectUser = computed(() => activeSubjectUserOptions.value.find((user) => Number(user.id) === activeSubjectUserId.value))
const activeSubjectPath = computed(() => activeSubjectType.value === 'department'
  ? activeDepartmentPath.value
  : activeSubjectUser.value?.name || '')
const orgApproverUserTreeData = computed(() => buildOrgApproverUserTree(deptTreeData.value, userOptions.value))
const subjectUserTreeData = computed(() => buildOrgApproverUserTree(deptTreeData.value, activeSubjectUserOptions.value))
const defaultExpandedUserTreeKeys = computed(() => {
  const deptIDs = activeSubjectType.value === 'department'
    ? [activeDepartmentId.value]
    : normalizeNumberIDs(activeSubjectUser.value?.deptIds || [])
  return Array.from(new Set(deptIDs.flatMap((id) => deptPathIDMap().get(id) || []).map((id) => `dept-${id}`)))
})
const orgApproverCheckedNodeKeys = computed(() => checkedOrgApproverUserNodeKeys(selectedUserIds.value, orgApproverUserTreeData.value))
const userById = computed(() => {
  const result = new Map<number, WorkflowAssigneeUser>()
  for (const user of userOptions.value) result.set(Number(user.id), user)
  return result
})
const assignmentNameByUserId = computed(() => {
  const result = new Map<number, string>()
  for (const row of assignments.value) {
    if (row.userId && row.userName) result.set(Number(row.userId), row.userName)
  }
  return result
})
const selectedUserRows = computed(() => selectedUserIds.value.map((id) => {
  const user = userById.value.get(Number(id))
  return {
    id,
    name: user?.name || assignmentNameByUserId.value.get(Number(id)) || String(id),
    mobile: user?.mobile || '',
  }
}))
const assignmentSummary = computed(() => {
  if (!scopeReady.value) return '选择配置对象和审批身份后，在这里指定担任人员'
  const subject = activeSubjectType.value === 'department'
    ? `“${activeSubjectPath.value}”部门内员工`
    : `员工“${activeSubjectPath.value}”`
  const identity = activeIdentity.value?.name || activeIdentityCode.value
  const names = selectedUserRows.value.map((row) => row.name)
  if (names.length === 0) return `${subject}的“${identity}”尚未指定担任人员`
  const assignees = names.length > 2 ? `${names.slice(0, 2).join('、')}等 ${names.length} 人` : names.join('、')
  return `${subject}的“${identity}”由 ${assignees} 担任`
})

function normalizeNumberIDs(ids: unknown[]) {
  return Array.from(new Set((ids || []).map((id) => Number(id)).filter(Boolean)))
}

function countDepartments(nodes: DepartmentNode[]): number {
  return (nodes || []).reduce((total, node) => total + 1 + countDepartments(node.children || []), 0)
}

function deptPathInfoMap() {
  const result = new Map<number, string[]>()
  function walk(nodes: DepartmentNode[], parentPath: string[] = []) {
    for (const node of nodes || []) {
      const id = Number(node.id)
      if (!id) continue
      const name = String(node.name || id)
      const path = [...parentPath, name]
      result.set(id, path)
      if (Array.isArray(node.children) && node.children.length > 0) walk(node.children, path)
    }
  }
  walk(deptTreeData.value)
  return result
}

function deptPathIDMap() {
  const result = new Map<number, number[]>()
  function walk(nodes: DepartmentNode[], parentPath: number[] = []) {
    for (const node of nodes || []) {
      const id = Number(node.id)
      if (!id) continue
      const path = [...parentPath, id]
      result.set(id, path)
      if (Array.isArray(node.children) && node.children.length > 0) walk(node.children, path)
    }
  }
  walk(deptTreeData.value)
  return result
}

function deptPathText(deptIds: number[]) {
  const pathMap = deptPathInfoMap()
  return normalizeNumberIDs(deptIds || [])
    .map((id) => (pathMap.get(id) || [String(id)]).join('|'))
    .join('、') || '-'
}

function firstDepartment(nodes: DepartmentNode[]): DepartmentNode | null {
  for (const node of nodes || []) {
    if (Number(node.id)) return node
    const child = firstDepartment(node.children || [])
    if (child) return child
  }
  return null
}

function buildOrgApproverUserNode(user: WorkflowAssigneeUser, scopeKey: string): OrgApproverTreeNode {
  const userId = Number(user.id)
  return {
    key: `user:${scopeKey}:${userId}`,
    type: 'user',
    id: userId,
    label: user.name || String(userId),
    mobile: user.mobile || '',
    pathText: deptPathText(user.deptIds || []),
  }
}

function buildOrgApproverUserTree(depts: DepartmentNode[], users: WorkflowAssigneeUser[]): OrgApproverTreeNode[] {
  const usersByDept: Record<number, WorkflowAssigneeUser[]> = {}
  const usersWithoutDept: WorkflowAssigneeUser[] = []
  for (const user of users || []) {
    const deptIds = normalizeNumberIDs(user.deptIds || [])
    if (deptIds.length === 0) {
      usersWithoutDept.push(user)
      continue
    }
    for (const deptId of deptIds) {
      if (!usersByDept[deptId]) usersByDept[deptId] = []
      usersByDept[deptId].push(user)
    }
  }

  function buildDeptNode(dept: DepartmentNode): OrgApproverTreeNode | null {
    const deptId = Number(dept.id)
    if (!deptId) return null
    const childDeptNodes = (dept.children || []).map((item) => buildDeptNode(item)).filter(Boolean) as OrgApproverTreeNode[]
    const userNodes = (usersByDept[deptId] || []).map((user) => buildOrgApproverUserNode(user, `dept-${deptId}`))
    const children = [...childDeptNodes, ...userNodes]
    if (children.length === 0) return null
    return {
      key: `dept-${deptId}`,
      type: 'dept',
      id: deptId,
      label: dept.name || String(deptId),
      disabled: true,
      children,
    }
  }

  const tree = (depts || []).map((dept) => buildDeptNode(dept)).filter(Boolean) as OrgApproverTreeNode[]
  if (usersWithoutDept.length > 0) {
    tree.push({
      key: 'dept-unassigned',
      type: 'dept',
      label: '未分配部门',
      disabled: true,
      children: usersWithoutDept.map((user) => buildOrgApproverUserNode(user, 'unassigned')),
    })
  }
  const visibleUserIds = new Set<number>()
  function collectUserIds(items: OrgApproverTreeNode[]) {
    for (const item of items || []) {
      if (item.type === 'user' && item.id) visibleUserIds.add(Number(item.id))
      if (Array.isArray(item.children) && item.children.length > 0) collectUserIds(item.children)
    }
  }
  collectUserIds(tree)
  const unmatchedUsers = (users || []).filter((user) => !visibleUserIds.has(Number(user.id)))
  if (unmatchedUsers.length > 0) {
    tree.push({
      key: 'dept-unmatched',
      type: 'dept',
      label: '未匹配部门',
      disabled: true,
      children: unmatchedUsers.map((user) => buildOrgApproverUserNode(user, 'unmatched')),
    })
  }
  return tree
}

function orgApproverUserIdFromNodeKey(key: string) {
  if (!key.startsWith('user:')) return 0
  const parts = key.split(':')
  return Number(parts[parts.length - 1]) || 0
}

function orgApproverUserIdsFromCheckedKeys(keys: string[]) {
  return normalizeNumberIDs((keys || []).map((key) => orgApproverUserIdFromNodeKey(String(key))).filter(Boolean))
}

function checkedOrgApproverUserNodeKeys(userIds: number[], nodes: OrgApproverTreeNode[]) {
  const selected = new Set(normalizeNumberIDs(userIds || []))
  const keys: string[] = []
  function walk(items: OrgApproverTreeNode[]) {
    for (const item of items || []) {
      if (item.type === 'user' && item.id && selected.has(Number(item.id))) keys.push(item.key)
      if (Array.isArray(item.children) && item.children.length > 0) walk(item.children)
    }
  }
  walk(nodes)
  return keys
}

function firstOrgApproverUserNodeKey(userId: number, nodes: OrgApproverTreeNode[]): string {
  for (const item of nodes || []) {
    if (item.type === 'user' && Number(item.id) === Number(userId)) return item.key
    if (item.children?.length) {
      const childKey = firstOrgApproverUserNodeKey(userId, item.children)
      if (childKey) return childKey
    }
  }
  return ''
}

function normalizeFilterText(value: unknown) {
  return String(value || '').trim().toLowerCase()
}

function filterDeptTree() {
  deptTreeRef.value?.filter(normalizeFilterText(deptKeyword.value))
}

function filterSubjectUserTree() {
  subjectUserTreeRef.value?.filter(normalizeFilterText(subjectUserKeyword.value))
}

function filterOrgApproverUserTree() {
  orgApproverUserTreeRef.value?.filter(normalizeFilterText(userKeyword.value))
}

function filterDeptNode(keyword: string, data: DepartmentNode) {
  const value = normalizeFilterText(keyword)
  if (!value) return true
  return normalizeFilterText(data?.name).includes(value)
}

function filterOrgApproverUserNode(keyword: string, data: OrgApproverTreeNode) {
  const value = normalizeFilterText(keyword)
  if (!value) return true
  return [data?.label, data?.mobile, data?.pathText].some((item) => normalizeFilterText(item).includes(value))
}

function setOrgApproverUserTreeKeys() {
  orgApproverUserTreeRef.value?.setCheckedKeys(orgApproverCheckedNodeKeys.value)
}

function expandActiveUserTreePath() {
  for (const key of defaultExpandedUserTreeKeys.value) {
    orgApproverUserTreeRef.value?.getNode?.(key)?.expand?.()
  }
}

function onOrgApproverUserCheck() {
  nextTick(() => {
    const checkedKeys = orgApproverUserTreeRef.value?.getCheckedKeys?.() || []
    const checkedIds = orgApproverUserIdsFromCheckedKeys(checkedKeys)
    const checkedSet = new Set(checkedIds)
    const keptIds = selectedUserIds.value.filter((id) => checkedSet.has(Number(id)))
    const addedIds = checkedIds.filter((id) => !keptIds.includes(Number(id)))
    selectedUserIds.value = normalizeNumberIDs([...keptIds, ...addedIds])
    nextTick(() => setOrgApproverUserTreeKeys())
  })
}

function moveSelectedUser(index: number, offset: number) {
  const target = index + offset
  if (target < 0 || target >= selectedUserIds.value.length) return
  const nextIds = [...selectedUserIds.value]
  const current = nextIds[index]
  nextIds[index] = nextIds[target]
  nextIds[target] = current
  selectedUserIds.value = nextIds
}

function removeSelectedUser(id: number) {
  selectedUserIds.value = selectedUserIds.value.filter((item) => Number(item) !== Number(id))
  nextTick(() => setOrgApproverUserTreeKeys())
}

function identityRowClassName({ row }: { row: WorkflowOrgApproverIdentity }) {
  return row.code === activeIdentityCode.value ? 'is-active-identity' : ''
}

async function onDepartmentClick(data: DepartmentNode) {
  const id = Number(data?.id)
  if (!id || id === activeDepartmentId.value) return
  activeDepartmentId.value = id
  await loadAssignments()
}

async function onSubjectTypeChange() {
  assignments.value = []
  selectedUserIds.value = []
  await nextTick()
  if (activeSubjectType.value === 'department') {
    deptTreeRef.value?.setCurrentKey(activeDepartmentId.value || undefined)
  } else {
    const key = firstOrgApproverUserNodeKey(activeSubjectUserId.value, subjectUserTreeData.value)
    subjectUserTreeRef.value?.setCurrentKey(key || undefined)
  }
  await loadAssignments()
}

async function onSubjectUserClick(data: OrgApproverTreeNode) {
  const id = data.type === 'user' ? Number(data.id) : 0
  if (!id || id === activeSubjectUserId.value) return
  activeSubjectUserId.value = id
  await loadAssignments()
}

async function onIdentityRowClick(row: WorkflowOrgApproverIdentity) {
  if (!row?.code || row.code === activeIdentityCode.value) return
  activeIdentityCode.value = row.code
  await loadAssignments()
}

async function loadDepartments() {
  deptLoading.value = true
  try {
    const response = await adminApi.deptTree()
    deptTreeData.value = Array.isArray(response.data) ? response.data : []
  } finally {
    deptLoading.value = false
  }
}

async function loadIdentities() {
  identityLoading.value = true
  try {
    const response = await adminApi.workflowOrgApproverIdentities()
    identities.value = Array.isArray(response.data) ? response.data : []
  } finally {
    identityLoading.value = false
  }
}

async function loadUsers() {
  userLoading.value = true
  try {
    const response = await adminApi.userList({ page: 1, pageSize: 9999 })
    userOptions.value = Array.isArray(response.data?.list) ? response.data.list : []
  } finally {
    userLoading.value = false
  }
}

async function loadAssignments() {
  if (!scopeReady.value) {
    assignments.value = []
    selectedUserIds.value = []
    return
  }
  assignmentsLoading.value = true
  try {
    const response = await adminApi.workflowOrgApproverAssignments({
      subjectType: activeSubjectType.value,
      subjectId: activeSubjectId.value,
      identityCode: activeIdentityCode.value,
    })
    const rows = Array.isArray(response.data) ? response.data : []
    assignments.value = rows
    selectedUserIds.value = normalizeNumberIDs(rows.sort((a: OrgApproverAssignment, b: OrgApproverAssignment) => Number(a.sort || 0) - Number(b.sort || 0)).map((row: OrgApproverAssignment) => row.userId))
    await nextTick()
    setOrgApproverUserTreeKeys()
    expandActiveUserTreePath()
  } finally {
    assignmentsLoading.value = false
  }
}

async function saveAssignments() {
  if (!scopeReady.value) {
    ElMessage.warning('请先选择配置对象和审批身份')
    return
  }
  saving.value = true
  try {
    await adminApi.workflowOrgApproverAssignmentsSave({
      subjectType: activeSubjectType.value,
      subjectId: activeSubjectId.value,
      identityCode: activeIdentityCode.value,
      userIds: selectedUserIds.value,
    })
    ElMessage.success('保存成功')
    await loadAssignments()
  } finally {
    saving.value = false
  }
}

async function reloadAll() {
  await Promise.all([loadDepartments(), loadIdentities(), loadUsers()])
  const firstDept = activeDepartmentId.value ? null : firstDepartment(deptTreeData.value)
  if (firstDept) activeDepartmentId.value = Number(firstDept.id)
  if (!activeIdentityCode.value && identityOptions.value.length > 0) activeIdentityCode.value = identityOptions.value[0].code
  await nextTick()
  if (activeDepartmentId.value) deptTreeRef.value?.setCurrentKey(activeDepartmentId.value)
  if (activeSubjectUserId.value) {
    const key = firstOrgApproverUserNodeKey(activeSubjectUserId.value, subjectUserTreeData.value)
    subjectUserTreeRef.value?.setCurrentKey(key || undefined)
  }
  await loadAssignments()
}

onMounted(() => {
  reloadAll()
})
</script>

<style scoped>
.workflow-org-approver-page {
  min-height: 100%;
}

.admin-card {
  min-height: calc(100vh - 170px);
}

.admin-card :deep(.el-card__body) {
  display: flex;
  flex-direction: column;
  min-height: calc(100vh - 212px);
}

.admin-toolbar h2 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #1f2937;
}

.org-approver-layout {
  display: grid;
  grid-template-columns: minmax(230px, 260px) minmax(240px, 280px) minmax(560px, 1fr);
  gap: 16px;
  flex: 1;
  align-items: stretch;
  min-height: 560px;
}

.org-panel {
  display: flex;
  flex-direction: column;
  min-width: 0;
  min-height: 0;
  overflow: hidden;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  background: #ffffff;
}

.org-panel--departments { grid-area: departments; }
.org-panel--identities { grid-area: identities; }
.org-panel--wide { grid-area: assignment; }

.org-panel__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  min-height: 52px;
  padding: 0 16px;
  border-bottom: 1px solid #edf0f3;
  background: #fbfcfe;
}

.org-panel__header--assignment {
  min-height: 62px;
}

.org-panel__content {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-height: 0;
  padding: 14px;
}

.subject-type-switch {
  display: flex;
  width: 100%;
  margin-bottom: 12px;
}

.subject-type-switch :deep(.el-radio-button) {
  flex: 1;
}

.subject-type-switch :deep(.el-radio-button__inner) {
  width: 100%;
}

.panel-title {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.panel-title strong {
  color: #1f2937;
  font-size: 15px;
}

.panel-title__index {
  display: inline-grid;
  width: 22px;
  height: 22px;
  place-items: center;
  border: 1px solid #b8d8ff;
  border-radius: 50%;
  background: #eef6ff;
  color: #337ecc;
  font-size: 12px;
  font-weight: 600;
}

.panel-title__count,
.candidate-users__header span,
.selected-users__header span {
  color: #8b95a5;
  font-size: 12px;
}

.assignment-scope {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  min-width: 0;
}

.assignment-summary {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  min-width: 0;
  color: #5f6b7a;
  font-size: 13px;
}

.assignment-summary__text {
  max-width: 520px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.assignment-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.org-tree-wrap {
  margin-top: 10px;
  flex: 1;
  min-height: 0;
  overflow: auto;
  border: 1px solid #eef0f3;
  border-radius: 6px;
  padding: 8px;
  background: #ffffff;
}

.identity-table-wrap {
  flex: 1;
  min-height: 0;
}

.identity-name {
  color: #303846;
  font-weight: 500;
}

.identity-code {
  color: #6b7280;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 12px;
}

.assignment-body {
  display: grid;
  grid-template-columns: minmax(320px, 1.08fr) minmax(320px, 0.92fr);
  flex: 1;
  min-height: 0;
  padding: 16px;
}

.assignment-picker,
.selected-users {
  display: flex;
  flex-direction: column;
  min-width: 0;
  min-height: 0;
}

.assignment-picker {
  padding-right: 16px;
}

.selected-users {
  padding-left: 16px;
  border-left: 1px solid #edf0f3;
}

.candidate-users__header,
.selected-users__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 32px;
  flex-shrink: 0;
  margin-bottom: 10px;
}

.candidate-users__header strong,
.selected-users__header strong {
  color: #303846;
  font-size: 14px;
}

.selected-users__table {
  flex: 1;
  min-height: 0;
}

.org-user-node {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.org-user-node__label {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.org-user-node__meta {
  color: #909399;
  font-size: 12px;
}

.selected-user-actions {
  display: flex;
  align-items: center;
  gap: 4px;
}

:deep(.is-active-identity td) {
  background: #eef6ff !important;
}

:deep(.is-active-identity td:first-child) {
  box-shadow: inset 3px 0 0 #409eff;
}

:deep(.el-tree-node__content) {
  min-width: max-content;
  height: 34px;
  border-radius: 4px;
}

:deep(.el-tree--highlight-current .el-tree-node.is-current > .el-tree-node__content) {
  background: #eaf3ff;
  color: #1769aa;
  font-weight: 600;
}

:deep(.identity-table-wrap .el-table__row) {
  cursor: pointer;
}

.org-panel--wide :deep(.el-empty) {
  flex: 1;
  justify-content: center;
}

@media (max-width: 1360px) {
  .org-approver-layout {
    grid-template-areas:
      "departments identities"
      "assignment assignment";
    grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  }

  .org-panel--departments,
  .org-panel--identities {
    min-height: 360px;
  }

  .org-panel--wide {
    min-height: 560px;
  }
}

@media (min-width: 1361px) {
  .org-approver-layout {
    grid-template-areas: "departments identities assignment";
  }
}

@media (max-width: 760px) {
  .admin-card,
  .admin-card :deep(.el-card__body) {
    min-height: auto;
  }

  .org-approver-layout {
    grid-template-areas:
      "departments"
      "identities"
      "assignment";
    grid-template-columns: minmax(0, 1fr);
  }

  .assignment-body {
    grid-template-columns: 1fr;
    gap: 18px;
  }

  .assignment-picker,
  .selected-users {
    min-height: 320px;
    padding: 0;
  }

  .selected-users {
    padding-top: 18px;
    border-top: 1px solid #edf0f3;
    border-left: 0;
  }

  .org-panel__header--assignment {
    align-items: flex-start;
    padding-top: 12px;
    padding-bottom: 12px;
  }

  .assignment-summary {
    width: 100%;
  }

  .assignment-summary__text {
    max-width: calc(100% - 112px);
  }
}
</style>
