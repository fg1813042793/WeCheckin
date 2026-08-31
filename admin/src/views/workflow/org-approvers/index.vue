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
        <section class="org-panel">
          <div class="org-panel__header">
            <strong>部门</strong>
          </div>
          <el-input
            v-model="deptKeyword"
            placeholder="搜索部门"
            clearable
            size="small"
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
              @node-click="onDepartmentClick"
            />
          </div>
        </section>

        <section class="org-panel">
          <div class="org-panel__header">
            <strong>身份</strong>
          </div>
          <el-table
            v-loading="identityLoading"
            :data="identityOptions"
            stripe
            height="420"
            highlight-current-row
            :row-class-name="identityRowClassName"
            @row-click="onIdentityRowClick"
          >
            <el-table-column prop="name" label="名称" min-width="120" show-overflow-tooltip />
            <el-table-column prop="code" label="编码" min-width="150" show-overflow-tooltip />
          </el-table>
        </section>

        <section class="org-panel org-panel--wide">
          <div class="org-panel__header org-panel__header--assignment">
            <div class="assignment-scope">
              <strong>处理人</strong>
              <span>{{ activeDepartmentName || '-' }}</span>
              <span>{{ activeIdentity?.name || '-' }}</span>
            </div>
            <div class="assignment-actions">
              <el-button icon="Refresh" :loading="assignmentsLoading" :disabled="!scopeReady" @click="loadAssignments" />
              <el-button v-if="canEdit" type="primary" icon="Check" :loading="saving" :disabled="!scopeReady" @click="saveAssignments">保存</el-button>
            </div>
          </div>

          <div v-if="scopeReady" class="assignment-body">
            <div class="assignment-picker">
              <el-input
                v-model="userKeyword"
                placeholder="搜索用户/手机号/部门"
                clearable
                size="small"
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
                <strong>已选处理人</strong>
                <el-tag size="small" type="info">{{ selectedUserRows.length }}</el-tag>
              </div>
              <el-table :data="selectedUserRows" stripe height="360">
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
          <el-empty v-else description="请选择部门和身份" />
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
  departmentId: number
  departmentName?: string
  identityCode: string
  userId: number
  userName?: string
  sort?: number
}

const canEdit = computed(() => hasPerm('admin:menu:workflow:org-approver:edit'))
const deptTreeRef = ref()
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
const activeDepartmentId = ref(0)
const activeIdentityCode = ref('')
const selectedUserIds = ref<number[]>([])
const deptKeyword = ref('')
const userKeyword = ref('')

const bootstrapLoading = computed(() => deptLoading.value || identityLoading.value || userLoading.value)
const identityOptions = computed(() => identities.value.filter((item) => Number(item.status ?? 1) === 1))
const activeIdentity = computed(() => identityOptions.value.find((item) => item.code === activeIdentityCode.value))
const scopeReady = computed(() => activeDepartmentId.value > 0 && !!activeIdentityCode.value)
const activeDepartmentName = computed(() => departmentName(activeDepartmentId.value))
const orgApproverUserTreeData = computed(() => buildOrgApproverUserTree(deptTreeData.value, userOptions.value))
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

function normalizeNumberIDs(ids: unknown[]) {
  return Array.from(new Set((ids || []).map((id) => Number(id)).filter(Boolean)))
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

function deptPathText(deptIds: number[]) {
  const pathMap = deptPathInfoMap()
  return normalizeNumberIDs(deptIds || [])
    .map((id) => (pathMap.get(id) || [String(id)]).join('|'))
    .join('、') || '-'
}

function departmentName(id: number) {
  if (!id) return ''
  const path = deptPathInfoMap().get(Number(id))
  return path?.[path.length - 1] || String(id)
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

function normalizeFilterText(value: unknown) {
  return String(value || '').trim().toLowerCase()
}

function filterDeptTree() {
  deptTreeRef.value?.filter(normalizeFilterText(deptKeyword.value))
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
      departmentId: activeDepartmentId.value,
      identityCode: activeIdentityCode.value,
    })
    const rows = Array.isArray(response.data) ? response.data : []
    assignments.value = rows
    selectedUserIds.value = normalizeNumberIDs(rows.sort((a: OrgApproverAssignment, b: OrgApproverAssignment) => Number(a.sort || 0) - Number(b.sort || 0)).map((row: OrgApproverAssignment) => row.userId))
    await nextTick()
    setOrgApproverUserTreeKeys()
  } finally {
    assignmentsLoading.value = false
  }
}

async function saveAssignments() {
  if (!scopeReady.value) {
    ElMessage.warning('请选择部门和身份')
    return
  }
  saving.value = true
  try {
    await adminApi.workflowOrgApproverAssignmentsSave({
      departmentId: activeDepartmentId.value,
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

.admin-toolbar h2 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #1f2937;
}

.org-approver-layout {
  display: grid;
  grid-template-columns: minmax(220px, 280px) minmax(240px, 300px) minmax(420px, 1fr);
  gap: 16px;
  align-items: stretch;
}

.org-panel {
  min-width: 0;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  padding: 14px;
  background: #ffffff;
}

.org-panel--wide {
  min-height: 500px;
}

.org-panel__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  min-height: 32px;
  margin-bottom: 12px;
}

.org-panel__header--assignment {
  align-items: flex-start;
}

.assignment-scope {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  min-width: 0;
}

.assignment-scope strong {
  color: #1f2937;
}

.assignment-scope span {
  max-width: 180px;
  padding: 3px 8px;
  border-radius: 6px;
  background: #f3f4f6;
  color: #4b5563;
  font-size: 12px;
  line-height: 20px;
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
  height: 420px;
  overflow: auto;
  border: 1px solid #eef0f3;
  border-radius: 6px;
  padding: 8px;
}

.org-tree-wrap--users {
  height: 360px;
}

.assignment-body {
  display: grid;
  grid-template-columns: minmax(240px, 1fr) minmax(260px, 360px);
  gap: 16px;
}

.assignment-picker,
.selected-users {
  min-width: 0;
}

.selected-users__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 32px;
  margin-bottom: 10px;
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
  background: #f0f9ff !important;
}

@media (max-width: 1180px) {
  .org-approver-layout {
    grid-template-columns: 1fr;
  }

  .assignment-body {
    grid-template-columns: 1fr;
  }

  .org-tree-wrap,
  .org-tree-wrap--users {
    height: 320px;
  }
}
</style>
