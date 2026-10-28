<script setup lang="ts">
import { computed, ref } from 'vue'
import { ArrowDown, ArrowRight, ArrowUp, Bell, Check, Delete, Grid, RefreshRight, Search, Setting, User } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

type SubjectType = 'department' | 'user'
interface OrgNode { key: string; name: string; type: SubjectType; level: number; mobile?: string; position?: string; children?: OrgNode[] }
interface Person { id: number; name: string; mobile: string; position: string; department: string }
interface Assignment { id: number; subjectKey: string; identity: string; employeeId: number }

const people: Person[] = [
  { id: 101, name: 'Foster', mobile: '138****2601', position: '运维', department: '研发部 / 运维组' },
  { id: 102, name: 'Lip', mobile: '138****1228', position: 'Java开发', department: '研发部 / Java开发组' },
  { id: 103, name: 'Steven', mobile: '137****8391', position: '业务负责人', department: 'M/H业务 / 综合部' },
  { id: 104, name: 'Nick', mobile: '136****4412', position: '产品经理', department: '产品部' },
  { id: 105, name: 'Amy', mobile: '135****9082', position: '综合部负责人', department: 'M/H业务 / 综合部' },
  { id: 106, name: 'David', mobile: '158****7710', position: '研发经理', department: '研发部 / Java开发组' },
]
const identityOptions = ['部门负责人', '主管', '组长', 'HRBP', '财务对接人']
const tree = ref<OrgNode[]>([
  { key: 'dept:company', name: '米视科技', type: 'department', level: 0, children: [
    { key: 'dept:rd', name: '研发部', type: 'department', level: 1, children: [
      { key: 'user:101', name: 'Foster', type: 'user', level: 2, mobile: '138****2601', position: '运维' },
      { key: 'user:102', name: 'Lip', type: 'user', level: 2, mobile: '138****1228', position: 'Java开发' },
      { key: 'user:106', name: 'David', type: 'user', level: 2, mobile: '158****7710', position: '研发经理' },
    ] },
    { key: 'dept:mh', name: 'M/H业务', type: 'department', level: 1, children: [
      { key: 'user:103', name: 'Steven', type: 'user', level: 2, mobile: '137****8391', position: '业务负责人' },
      { key: 'user:105', name: 'Amy', type: 'user', level: 2, mobile: '135****9082', position: '综合部负责人' },
    ] },
    { key: 'dept:product', name: '产品部', type: 'department', level: 1, children: [
      { key: 'user:104', name: 'Nick', type: 'user', level: 2, mobile: '136****4412', position: '产品经理' },
    ] },
  ] },
])
const selectedKey = ref('dept:company')
const expandedKeys = ref(new Set<string>())
const treeKeyword = ref('')
const dialogVisible = ref(false)
const saving = ref(false)
const form = ref({ identity: '', employeeId: 0 })
const settingTreeRef = ref()
const settingUserKeyword = ref('')
const assignments = ref<Assignment[]>([
  { id: 1, subjectKey: 'dept:company', identity: '部门负责人', employeeId: 101 },
  { id: 2, subjectKey: 'dept:rd', identity: '主管', employeeId: 106 },
  { id: 3, subjectKey: 'user:101', identity: 'HRBP', employeeId: 105 },
])
const selectedNode = computed(() => findNode(tree.value, selectedKey.value) || tree.value[0])
const filteredTree = computed(() => tree.value.map(filterNode).filter(Boolean) as OrgNode[])
const selectedAssignments = computed(() => assignments.value.filter(item => item.subjectKey === selectedKey.value))
const selectedTitle = computed(() => selectedNode.value.type === 'department' ? selectedNode.value.name + '部门' : '员工：' + selectedNode.value.name)
const settingUserTree = computed(() => [{
  key: 'setting-users', name: '全部员工', type: 'department' as const, level: 0, children: people
    .filter(person => !settingUserKeyword.value || `${person.name} ${person.position} ${person.department}`.toLowerCase().includes(settingUserKeyword.value.trim().toLowerCase()))
    .map(person => ({ key: `setting-user:${person.id}`, name: person.name, type: 'user' as const, level: 1, position: person.position })),
}])

function findNode(nodes: OrgNode[], key: string): OrgNode | null {
  for (const node of nodes) {
    if (node.key === key) return node
    if (node.children?.length) {
      const found = findNode(node.children, key)
      if (found) return found
    }
  }
  return null
}
function filterNode(node: OrgNode): OrgNode | null {
  const keyword = treeKeyword.value.trim().toLowerCase()
  if (!keyword) return node
  const children = node.children?.map(filterNode).filter(Boolean) as OrgNode[] | undefined
  if (node.name.toLowerCase().includes(keyword) || node.position?.toLowerCase().includes(keyword) || children?.length) return { ...node, children }
  return null
}
function toggleNode(node: OrgNode) {
  if (!node.children?.length) return
  const next = new Set(expandedKeys.value)
  if (next.has(node.key)) next.delete(node.key)
  else next.add(node.key)
  expandedKeys.value = next
}
function selectNode(node: OrgNode) { selectedKey.value = node.key }
function openSettings() { form.value = { identity: '', employeeId: 0 }; settingUserKeyword.value = ''; dialogVisible.value = true }
function editAssignment(row: Assignment) { form.value = { identity: row.identity, employeeId: row.employeeId }; settingUserKeyword.value = ''; dialogVisible.value = true }
function selectSettingUser(person: Person) { form.value.employeeId = person.id; settingTreeRef.value?.setCheckedKeys?.([`setting-user:${person.id}`]) }
function handleSettingCheck(_node: unknown, checked: { checkedKeys?: unknown[] }) {
  const key = checked.checkedKeys?.[0]
  const person = people.find(item => `setting-user:${item.id}` === key)
  if (person) selectSettingUser(person)
}
function isSettingNodeDisabled(data: OrgNode) { return data.type === 'department' }
function removeAssignment(row: Assignment) {
  assignments.value = assignments.value.filter(item => item.id !== row.id)
  ElMessage.success('已移除审批负责人')
}
function saveAssignment() {
  if (!form.value.identity || !form.value.employeeId) {
    ElMessage.warning('请选择审批身份和负责人')
    return
  }
  saving.value = true
  const current = assignments.value.find(item => item.subjectKey === selectedKey.value && item.identity === form.value.identity)
  if (current) current.employeeId = form.value.employeeId
  else assignments.value.push({ id: Date.now(), subjectKey: selectedKey.value, identity: form.value.identity, employeeId: form.value.employeeId })
  saving.value = false
  dialogVisible.value = false
  ElMessage.success('审批负责人已保存')
}
function personName(id: number) { return people.find(item => item.id === id)?.name || String(id) }
function personPosition(id: number) { return people.find(item => item.id === id)?.position || '未设置岗位' }
</script>

<template>
  <div class="preview-app org-preview">
    <aside class="preview-sidebar">
      <a class="preview-brand" href="/design-preview.html"><span class="brand-mark">W</span><span class="brand-copy">WeCheckin<small>管理工作台</small></span></a>
      <nav><span class="nav-section-label">工作空间</span><a href="/design-preview.html"><el-icon><Grid /></el-icon><span>控制台</span></a><a href="/user-preview.html"><el-icon><User /></el-icon><span>用户管理</span></a><a class="active" href="/org-approver-preview.html"><el-icon><Bell /></el-icon><span>组织审批身份</span></a></nav>
      <div class="sidebar-footer"><span class="avatar">A</span><span class="profile-copy">管理员<small>Admin workspace</small></span></div>
    </aside>
    <div class="preview-workspace">
      <header class="workspace-header"><div class="breadcrumb"><span>流程管理</span><el-icon>›</el-icon><strong>组织审批身份设置</strong></div><span class="preview-indicator"><i />样式预览 · 演示数据</span></header>
      <div class="workspace-tabs"><span>控制台</span><span>流程管理</span><span class="current">组织审批身份设置</span></div>
      <main class="org-manage-main">
        <header class="org-manage-heading"><div><h1>组织审批身份设置</h1><p>点击左侧部门或员工，查看并配置对应的审批负责人。</p></div><el-button text :icon="RefreshRight" aria-label="刷新" title="刷新" /></header>
        <section class="org-manage-shell">
          <aside class="org-tree-panel">
            <header><strong>组织架构</strong><span>部门 / 员工</span></header>
            <div class="org-tree-toolbar"><el-input v-model="treeKeyword" :prefix-icon="Search" placeholder="搜索部门或员工" clearable /></div>
            <div class="org-tree-list">
              <template v-for="node in filteredTree" :key="node.key">
                <button class="tree-node" :class="{ active: selectedKey === node.key }" :style="{ paddingLeft: (8 + node.level * 20) + 'px' }" @click="selectNode(node)"><span v-if="node.children?.length" class="tree-toggle" @click.stop="toggleNode(node)">{{ expandedKeys.has(node.key) ? '⌄' : '›' }}</span><span v-else class="tree-toggle-placeholder" /><span class="tree-node-icon"><Grid v-if="node.type === 'department'" /><User v-else /></span><span class="tree-node-name">{{ node.name }}</span><small v-if="node.position">{{ node.position }}</small></button>
                <template v-if="node.children?.length && expandedKeys.has(node.key)">
                  <template v-for="child in node.children" :key="child.key">
                    <button class="tree-node" :class="{ active: selectedKey === child.key }" :style="{ paddingLeft: (8 + child.level * 20) + 'px' }" @click="selectNode(child)"><span v-if="child.children?.length" class="tree-toggle" @click.stop="toggleNode(child)">{{ expandedKeys.has(child.key) ? '⌄' : '›' }}</span><span v-else class="tree-toggle-placeholder" /><span class="tree-node-icon"><Grid v-if="child.type === 'department'" /><User v-else /></span><span class="tree-node-name">{{ child.name }}</span><small v-if="child.position">{{ child.position }}</small></button>
                    <template v-if="child.children?.length && expandedKeys.has(child.key)"><button v-for="leaf in child.children" :key="leaf.key" class="tree-node" :class="{ active: selectedKey === leaf.key }" :style="{ paddingLeft: (8 + leaf.level * 20) + 'px' }" @click="selectNode(leaf)"><span class="tree-toggle-placeholder" /><span class="tree-node-icon"><User /></span><span class="tree-node-name">{{ leaf.name }}</span><small>{{ leaf.position }}</small></button></template>
                  </template>
                </template>
              </template>
            </div>
          </aside>
          <section class="org-owner-panel">
            <header class="org-owner-heading"><div><span class="owner-step">当前对象</span><h2>{{ selectedTitle }}</h2><p>{{ selectedNode.type === 'department' ? '部门内员工的审批身份负责人' : selectedNode.name + ' 的审批身份负责人' }}</p></div><el-button type="primary" :icon="Setting" @click="openSettings">设置审批负责人</el-button></header>
            <div class="owner-summary"><span class="summary-icon"><Check /></span><span>已配置 <strong>{{ selectedAssignments.length }}</strong> 个审批身份</span><span class="summary-separator" /><span>{{ selectedNode.type === 'department' ? '部门通用规则' : '员工专属规则' }}</span></div>
            <el-table :data="selectedAssignments" class="owner-table" empty-text="暂未设置审批负责人" stripe>
              <el-table-column prop="identity" label="审批身份" min-width="180" />
              <el-table-column label="负责人" min-width="160"><template #default="{ row }"><strong>{{ personName(row.employeeId) }}</strong></template></el-table-column>
              <el-table-column label="岗位" width="150"><template #default="{ row }">{{ personPosition(row.employeeId) }}</template></el-table-column>
              <el-table-column label="操作" width="180" fixed="right"><template #default="{ row }"><el-button link type="primary" :icon="Setting" @click="editAssignment(row)">调整</el-button><el-button link type="danger" :icon="Delete" @click="removeAssignment(row)">移除</el-button></template></el-table-column>
            </el-table>
          </section>
        </section>
      </main>
    </div>
    <el-dialog class="org-setting-dialog" v-model="dialogVisible" title="设置审批负责人" width="min(680px, 94vw)" destroy-on-close><div class="owner-dialog-context"><strong>{{ selectedTitle }}</strong><span>为当前对象配置审批身份与负责人</span></div><el-form label-position="top"><el-form-item label="审批身份" required><el-select v-model="form.identity" placeholder="请选择审批身份" style="width:100%"><el-option v-for="identity in identityOptions" :key="identity" :label="identity" :value="identity" /></el-select></el-form-item><el-form-item label="负责人" required><el-input v-model="settingUserKeyword" :prefix-icon="Search" placeholder="搜索员工、岗位或部门" clearable /><el-tree ref="settingTreeRef" class="org-setting-tree" :data="settingUserTree" :props="{ label: 'name', children: 'children', disabled: isSettingNodeDisabled }" node-key="key" show-checkbox check-strictly empty-text="暂无可选员工" @check="handleSettingCheck"><template #default="{ node, data }"><span class="org-setting-node"><span>{{ node.label }}</span><small v-if="data.position">{{ data.position }}</small></span></template></el-tree></el-form-item></el-form><template #footer><el-button @click="dialogVisible = false">取消</el-button><el-button type="primary" :icon="Check" :loading="saving" @click="saveAssignment">保存设置</el-button></template></el-dialog>
  </div>
</template>
