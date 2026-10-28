<script setup lang="ts">
import { computed, nextTick, onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Check, OfficeBuilding, User } from '@element-plus/icons-vue'
import { adminApi } from '../../../api'
import { hasPerm } from '../../../utils/permission'
import type { WorkflowAssigneeUser, WorkflowOrgApproverIdentity } from '../types'

type DepartmentNode = { id: number; name: string; children?: DepartmentNode[] }
type TreeNode = { key: string; type: 'department' | 'user'; id?: number; label: string; position?: string; disabled?: boolean; children?: TreeNode[] }
type Assignment = { id: number; subjectType: 'department' | 'user'; subjectId: number; identityCode: string; userId: number; userName?: string; sort?: number }

const canEdit = computed(() => hasPerm('admin:menu:workflow:org-approver:edit'))
const loading = ref(false)
const assignmentsLoading = ref(false)
const saving = ref(false)
const departments = ref<DepartmentNode[]>([])
const users = ref<WorkflowAssigneeUser[]>([])
const identities = ref<WorkflowOrgApproverIdentity[]>([])
const assignments = ref<Assignment[]>([])
const selectedKey = ref('')
const treeKeyword = ref('')
const settingKeyword = ref('')
const settingVisible = ref(false)
const subjectTreeRef = ref()
const settingTreeRef = ref()
const settingForm = ref({ identityCode: '', userId: 0 })

const identityOptions = computed(() => identities.value.filter(item => Number(item.status ?? 1) === 1))
const subjectTree = computed(() => buildTree(departments.value, users.value, false))
const filteredTree = computed(() => filterTree(subjectTree.value, treeKeyword.value))
const treeExpandedKeys = computed(() => {
  const keys: string[] = []
  function collect(nodes: TreeNode[]) {
    for (const node of nodes) {
      if (node.children?.length) {
        keys.push(node.key)
        collect(node.children)
      }
    }
  }
  if (treeKeyword.value.trim()) collect(filteredTree.value)
  return keys
})
const settingTree = computed(() => buildTree(departments.value, users.value, true))
const selectedSubject = computed(() => findNode(subjectTree.value, selectedKey.value))
const selectedTitle = computed(() => selectedSubject.value?.type === 'user' ? '员工：' + selectedSubject.value.label : (selectedSubject.value?.label || '请选择对象') + '部门')

function ids(value: unknown[]) { return Array.from(new Set((value || []).map(item => Number(item)).filter(Boolean))) }
function findNode(nodes: TreeNode[], key: string): TreeNode | null { for (const node of nodes) { if (node.key === key) return node; const found = node.children?.length ? findNode(node.children, key) : null; if (found) return found } return null }
function filterTree(nodes: TreeNode[], raw: string): TreeNode[] { const keyword = raw.trim().toLowerCase(); if (!keyword) return nodes; return nodes.map(node => { const children = node.children?.length ? filterTree(node.children, raw) : []; return node.label.toLowerCase().includes(keyword) || node.position?.toLowerCase().includes(keyword) || children.length ? { ...node, children } : null }).filter(Boolean) as TreeNode[] }
function buildTree(deptNodes: DepartmentNode[], userList: WorkflowAssigneeUser[], disableDepartments: boolean): TreeNode[] {
  const byDept = new Map<number, WorkflowAssigneeUser[]>()
  const withoutDept: WorkflowAssigneeUser[] = []
  for (const user of userList) { const deptIds = ids(user.deptIds || []); if (!deptIds.length) withoutDept.push(user); for (const deptId of deptIds) byDept.set(deptId, [...(byDept.get(deptId) || []), user]) }
  function visit(node: DepartmentNode): TreeNode {
    const usersInDept = (byDept.get(Number(node.id)) || []).map(user => ({ key: 'user-' + user.id, type: 'user' as const, id: Number(user.id), label: user.name || String(user.id), position: user.positionName || '未设置岗位' }))
    return { key: 'dept-' + node.id, type: 'department', id: Number(node.id), label: node.name, disabled: disableDepartments, children: [...(node.children || []).map(visit), ...usersInDept] }
  }
  const result = deptNodes.map(visit)
  if (withoutDept.length) result.push({ key: 'dept-unassigned', type: 'department', label: '未分配部门', disabled: disableDepartments, children: withoutDept.map(user => ({ key: 'user-' + user.id, type: 'user' as const, id: Number(user.id), label: user.name || String(user.id), position: user.positionName || '未设置岗位' })) })
  return result
}
function identityName(code: string) { return identityOptions.value.find(item => item.code === code)?.name || code }
function userName(id: number) { return users.value.find(item => Number(item.id) === Number(id))?.name || String(id) }
function userPosition(id: number) { return users.value.find(item => Number(item.id) === Number(id))?.positionName || '未设置岗位' }
function selectSubject(node: TreeNode) { if (!node.id) return; selectedKey.value = node.key; loadAssignments() }
function filterSettingTree() { settingTreeRef.value?.filter?.(settingKeyword.value.trim().toLowerCase()) }
function filterSettingNode(keyword: string, data: TreeNode) { const value = keyword.trim().toLowerCase(); return !value || data.label.toLowerCase().includes(value) || !!data.position?.toLowerCase().includes(value) }
function handleSettingCheck(data: TreeNode) { if (data.type === 'user' && data.id) { settingForm.value.userId = Number(data.id); settingTreeRef.value?.setCheckedKeys?.([data.key]) } }
function openSetting(row?: Assignment) { settingForm.value = { identityCode: row?.identityCode || identityOptions.value[0]?.code || '', userId: row?.userId || 0 }; settingKeyword.value = ''; settingVisible.value = true; nextTick(() => { settingTreeRef.value?.setCheckedKeys?.(settingForm.value.userId ? ['user-' + settingForm.value.userId] : []) }) }
async function saveSetting() { if (!selectedSubject.value?.id || !settingForm.value.identityCode || !settingForm.value.userId) { ElMessage.warning('请选择审批身份和负责人'); return } saving.value = true; try { await adminApi.workflowOrgApproverAssignmentsSave({ subjectType: selectedSubject.value.type, subjectId: selectedSubject.value.id, identityCode: settingForm.value.identityCode, userIds: [settingForm.value.userId] }); ElMessage.success('审批负责人已保存'); settingVisible.value = false; await loadAssignments() } finally { saving.value = false } }
async function removeAssignment(row: Assignment) { try { await ElMessageBox.confirm('确定移除“' + identityName(row.identityCode) + '”的审批负责人？', '移除审批负责人', { type: 'warning' }) } catch { return } await adminApi.workflowOrgApproverAssignmentsSave({ subjectType: row.subjectType, subjectId: row.subjectId, identityCode: row.identityCode, userIds: [] }); ElMessage.success('已移除审批负责人'); await loadAssignments() }
async function loadAssignments() { if (!selectedSubject.value?.id) { assignments.value = []; return } assignmentsLoading.value = true; try { const response = await adminApi.workflowOrgApproverAssignments({ subjectType: selectedSubject.value.type, subjectId: selectedSubject.value.id }); assignments.value = Array.isArray(response.data) ? response.data : [] } finally { assignmentsLoading.value = false } }
async function reload() { loading.value = true; try { const [deptResponse, identityResponse, userResponse] = await Promise.all([adminApi.deptTree(), adminApi.workflowOrgApproverIdentities(), adminApi.userList({ page: 1, pageSize: 9999 })]); departments.value = Array.isArray(deptResponse.data) ? deptResponse.data : []; identities.value = Array.isArray(identityResponse.data) ? identityResponse.data : []; users.value = Array.isArray(userResponse.data?.list) ? userResponse.data.list : []; if (!selectedKey.value) selectedKey.value = subjectTree.value[0]?.key || ''; await nextTick(); subjectTreeRef.value?.setCurrentKey?.(selectedKey.value); await loadAssignments() } finally { loading.value = false } }
onMounted(reload)
</script>

<template>
  <div class="admin-page org-approver-page">
    <section class="org-approver-header"><div><h1>组织审批身份设置</h1><p>点击左侧部门或员工，查看并配置对应的审批负责人。</p></div><el-button text icon="Refresh" title="刷新" :loading="loading" @click="reload" /></section>
    <section class="org-approver-workspace">
      <aside class="org-subject-panel">
        <header class="org-panel-heading"><div><strong>组织架构</strong><span>部门 / 员工</span></div></header>
        <div class="org-subject-toolbar"><el-input v-model="treeKeyword" prefix-icon="Search" placeholder="搜索部门、员工或岗位" clearable /></div>
        <el-tree ref="subjectTreeRef" v-loading="loading" :key="treeKeyword" class="org-subject-tree" :data="filteredTree" :default-expanded-keys="treeExpandedKeys" :props="{ label: 'label', children: 'children' }" node-key="key" highlight-current empty-text="暂无组织数据" @node-click="selectSubject">
          <template #default="{ node, data }"><span class="org-subject-node" :class="data.type === 'user' ? 'is-user' : 'is-department'"><el-icon><User v-if="data.type === 'user'" /><OfficeBuilding v-else /></el-icon><span>{{ node.label }}</span><small v-if="data.position">{{ data.position }}</small></span></template>
        </el-tree>
      </aside>
      <main class="org-owner-panel">
        <header class="org-owner-header"><div><span class="org-owner-kicker">当前对象</span><h2>{{ selectedTitle }}</h2><p>{{ selectedSubject?.type === 'department' ? '部门内员工的审批身份负责人' : '员工的专属审批身份负责人' }}</p></div><el-button v-if="canEdit" type="primary" icon="Setting" @click="openSetting">设置审批负责人</el-button></header>
        <div class="org-owner-summary"><span class="org-summary-icon"><el-icon><Check /></el-icon></span><span>已配置 <strong>{{ assignments.length }}</strong> 个审批身份</span><span class="org-summary-divider" /><span>{{ selectedSubject?.type === 'department' ? '部门通用规则' : '员工专属规则' }}</span></div>
        <el-table v-loading="assignmentsLoading" :data="assignments" stripe class="org-owner-table" empty-text="暂未设置审批负责人">
          <el-table-column label="审批身份" min-width="180"><template #default="{ row }">{{ identityName(row.identityCode) }}</template></el-table-column>
          <el-table-column label="负责人" min-width="160"><template #default="{ row }"><strong>{{ row.userName || userName(row.userId) }}</strong></template></el-table-column>
          <el-table-column label="岗位" width="150"><template #default="{ row }">{{ userPosition(row.userId) }}</template></el-table-column>
          <el-table-column label="操作" width="180" fixed="right"><template #default="{ row }"><el-button link type="primary" icon="Edit" @click="openSetting(row)">调整</el-button><el-button v-if="canEdit" link type="danger" icon="Delete" @click="removeAssignment(row)">移除</el-button></template></el-table-column>
        </el-table>
      </main>
    </section>
    <el-dialog class="org-setting-dialog" v-model="settingVisible" title="设置审批负责人" width="min(680px, 94vw)" destroy-on-close><div class="org-dialog-context"><strong>{{ selectedTitle }}</strong><span>选择审批身份，再选择该身份对应的员工。</span></div><el-form label-position="top"><el-form-item label="审批身份" required><el-select v-model="settingForm.identityCode" placeholder="请选择审批身份" style="width:100%"><el-option v-for="item in identityOptions" :key="item.code" :label="item.name" :value="item.code" /></el-select></el-form-item><el-form-item label="负责人" required><el-input v-model="settingKeyword" prefix-icon="Search" placeholder="搜索员工、岗位或部门" clearable @input="filterSettingTree" /><el-tree ref="settingTreeRef" class="org-setting-tree" :data="settingTree" :props="{ label: 'label', children: 'children', disabled: 'disabled' }" node-key="key" show-checkbox check-strictly :filter-node-method="filterSettingNode" empty-text="暂无可选员工" @check="handleSettingCheck"><template #default="{ node, data }"><span class="org-setting-node"><span>{{ node.label }}</span><small v-if="data.position">{{ data.position }}</small></span></template></el-tree></el-form-item></el-form><template #footer><el-button @click="settingVisible = false">取消</el-button><el-button v-if="canEdit" type="primary" icon="Check" :loading="saving" @click="saveSetting">保存设置</el-button></template></el-dialog>
  </div>
</template>

<style scoped>
.org-approver-page { display: flex; min-height: 100%; flex-direction: column; gap: 16px; }
.org-approver-header { display: flex; align-items: center; justify-content: space-between; gap: 16px; padding: 18px 20px; border: 1px solid var(--admin-ui-color-border); border-radius: 8px; background: #fff; }
.org-approver-header h1 { margin: 0; color: var(--admin-ui-color-text); font-size: 20px; line-height: 28px; }
.org-approver-header p { margin: 4px 0 0; color: var(--admin-ui-color-text-secondary); font-size: 13px; }
.org-approver-workspace { display: grid; grid-template-columns: 300px minmax(0, 1fr); min-height: 680px; overflow: hidden; border: 1px solid var(--admin-ui-color-border); border-radius: 8px; background: #fff; }
.org-subject-panel { display: flex; min-width: 0; flex-direction: column; border-right: 1px solid #e7ebf0; }
.org-panel-heading { display: flex; align-items: center; justify-content: space-between; min-height: 56px; padding: 0 16px; border-bottom: 1px solid #edf0f4; background: #fbfcfe; }
.org-panel-heading div { display: flex; align-items: center; gap: 10px; }
.org-panel-heading strong { color: var(--admin-ui-color-text); font-size: 14px; }
.org-panel-heading span { color: var(--admin-ui-color-text-muted); font-size: 11px; }
.org-subject-toolbar { padding: 12px; border-bottom: 1px solid #edf0f4; }
.org-subject-toolbar :deep(.el-input__wrapper.is-focus) { box-shadow: 0 0 0 1px var(--admin-ui-color-border) inset !important; }
.org-subject-tree { flex: 1; min-height: 0; padding: 8px; overflow: auto; }
.org-subject-node { display: flex; min-width: 0; align-items: center; gap: 7px; }
.org-subject-node .el-icon { flex: 0 0 auto; color: #738096; }
.org-subject-node.is-user .el-icon { color: #6d8fb5; }
.org-subject-node > span { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.org-subject-node small { margin-left: 6px; overflow: hidden; color: #9aa5b3; font-size: 10px; text-overflow: ellipsis; white-space: nowrap; }
.org-owner-panel { display: flex; min-width: 0; flex-direction: column; }
.org-owner-header { display: flex; align-items: center; justify-content: space-between; gap: 16px; min-height: 86px; padding: 16px 20px; border-bottom: 1px solid #edf0f4; background: #fbfcfe; }
.org-owner-kicker { color: var(--admin-ui-color-primary); font-size: 11px; font-weight: 650; }
.org-owner-header h2 { margin: 3px 0 0; color: var(--admin-ui-color-text); font-size: 18px; }
.org-owner-header p { margin: 4px 0 0; color: var(--admin-ui-color-text-muted); font-size: 12px; }
.org-owner-summary { display: flex; align-items: center; gap: 8px; min-height: 48px; padding: 0 20px; border-bottom: 1px solid #edf0f4; color: var(--admin-ui-color-text-secondary); font-size: 12px; }
.org-summary-icon { display: grid; width: 20px; height: 20px; place-items: center; border-radius: 50%; background: #eaf8f1; color: var(--admin-ui-color-success); }
.org-summary-divider { width: 1px; height: 14px; margin: 0 4px; background: #dfe4ea; }
.org-owner-table { flex: 1; min-height: 0; }
.org-owner-table :deep(.el-table__cell) { padding: 14px 0; }
.org-dialog-context { display: flex; flex-direction: column; gap: 5px; margin-bottom: 18px; padding: 12px; border-radius: 6px; background: #f7faff; }
.org-dialog-context strong { color: var(--admin-ui-color-text); font-size: 14px; }
.org-dialog-context span { color: var(--admin-ui-color-text-secondary); font-size: 12px; }
.org-setting-tree { box-sizing: border-box; width: 100%; max-height: 360px; margin-top: 10px; overflow: auto; border: 1px solid #edf0f4; border-radius: 6px; }
.org-setting-tree :deep(.el-tree-node__content) { width: 100%; }
.org-setting-node { display: flex; align-items: center; gap: 8px; width: 100%; min-width: 0; }
.org-setting-node > span { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.org-setting-node small { flex: 0 1 180px; margin-left: 6px; overflow: hidden; color: var(--admin-ui-color-text-muted); font-size: 11px; text-overflow: ellipsis; white-space: nowrap; }
@media (max-width: 900px) { .org-approver-workspace { grid-template-columns: 240px minmax(0, 1fr); } }
@media (max-width: 680px) { .org-approver-workspace { grid-template-columns: 1fr; } .org-subject-panel { max-height: 420px; border-right: 0; border-bottom: 1px solid #e7ebf0; } .org-owner-header { align-items: flex-start; flex-direction: column; } }
</style>
