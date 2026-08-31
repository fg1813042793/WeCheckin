<template>
  <aside class="node-inspector" :class="{ readonly }">

    <el-empty v-if="!selectedNode" :image-size="72" description="选择节点后在此配置" />

    <div v-else class="inspector-content">
      <section class="inspector-section">
        <label class="field-label">节点名称</label>
        <el-input v-model="selectedNode.name" maxlength="40" :disabled="readonly" @input="$emit('change')" />
      </section>

      <template v-if="selectedNode.type === 'approval'">
        <section class="inspector-section">
          <h4>审批规则</h4>
          <label class="field-label">审批方式</label>
          <el-select :model-value="selectedNode.approvalMode" :disabled="readonly" style="width: 100%" @change="updateApprovalMode">
            <el-option label="单人审批" value="single" />
            <el-option label="依次审批" value="sequential" />
            <el-option label="并行审批" value="parallel" />
            <el-option label="会签审批" value="countersign" />
          </el-select>
          <label class="field-label spacing">处理人来源</label>
          <el-select :model-value="selectedNode.assignee?.type" :disabled="readonly" style="width: 100%" @change="updateAssigneeType">
            <el-option label="指定用户" value="user" />
            <el-option label="指定角色" value="role" />
            <el-option label="部门负责人" value="department_leader" />
            <el-option label="直属上级" value="manager" />
            <el-option label="流程变量" value="variable" />
          </el-select>
          <label class="field-label spacing">处理人标识</label>
          <el-input
            :model-value="selectedNode.assignee?.value"
            :disabled="readonly"
            :placeholder="assigneePlaceholder(selectedNode.assignee?.type)"
            @input="updateAssigneeValue"
          />
          <template v-if="selectedNode.approvalMode === 'countersign'">
            <label class="field-label spacing">通过比例</label>
            <el-input-number
              :model-value="selectedNode.completionRate ?? 100"
              :min="1"
              :max="100"
              :step="5"
              :disabled="readonly"
              controls-position="right"
              style="width: 100%"
              @change="updateCompletionRate"
            />
          </template>
        </section>
      </template>

      <template v-if="selectedNode.gatewayMode === 'split'">
        <section class="inspector-section">
          <div class="section-title-row">
            <h4>分支配置</h4>
            <el-button size="small" type="primary" plain icon="Plus" :disabled="readonly" @click="$emit('add-branch', selectedNode.id)">新增分支</el-button>
          </div>
          <div v-for="(branch, index) in branchEdges" :key="branch.id" class="branch-editor">
            <div class="branch-editor__heading">
              <b>分支 {{ index + 1 }}</b>
              <el-button
                link
                type="danger"
                :disabled="readonly || branchEdges.length <= 2"
                @click="$emit('delete-branch', selectedNode.id, branch.id)"
              >删除</el-button>
            </div>
            <label class="field-label">分支名称</label>
            <el-input :model-value="branch.name" size="small" :disabled="readonly" @input="updateBranchName(branch, $event)" />
            <template v-if="selectedNode.type === 'exclusive'">
              <el-checkbox :model-value="branch.default === true" :disabled="readonly" class="default-check" @change="updateDefaultBranch(branch, $event)">默认分支</el-checkbox>
              <template v-if="!branch.default">
                <label class="field-label">条件字段</label>
                <el-input :model-value="branch.condition?.field" size="small" :disabled="readonly" placeholder="例如：amount" @input="updateConditionField(branch, $event)" />
                <div class="condition-row">
                  <el-select :model-value="branch.condition?.operator || 'eq'" size="small" :disabled="readonly" @change="updateConditionOperator(branch, $event)">
                    <el-option label="等于" value="eq" />
                    <el-option label="不等于" value="ne" />
                    <el-option label="大于" value="gt" />
                    <el-option label="大于等于" value="gte" />
                    <el-option label="小于" value="lt" />
                    <el-option label="小于等于" value="lte" />
                  </el-select>
                  <el-input :model-value="String(branch.condition?.value ?? '')" size="small" :disabled="readonly" placeholder="条件值" @input="updateConditionValue(branch, $event)" />
                </div>
              </template>
            </template>
          </div>
        </section>
      </template>

      <section v-if="selectedNode.gatewayMode === 'join'" class="inspector-section join-note">
        <el-icon><Connection /></el-icon>
        <span>该节点由对应分支自动汇聚。删除此节点会同时删除整组分支。</span>
      </section>

      <div v-if="canDelete" class="inspector-footer">
        <el-button type="danger" plain icon="Delete" @click="$emit('delete', selectedNode.id)">删除节点</el-button>
      </div>
    </div>
  </aside>
</template>

<script lang="ts" setup>
import { computed } from 'vue'
import type { ApprovalMode, AssigneeType, ConditionOperator, WorkflowDraft, WorkflowEdge } from '../../types'

const props = defineProps<{ draft: WorkflowDraft; selectedNodeId: string; readonly?: boolean }>()
const emit = defineEmits<{
  change: []
  delete: [nodeId: string]
  'delete-branch': [splitId: string, edgeId: string]
  'add-branch': [splitId: string]
}>()

const selectedNode = computed(() => props.draft.nodes.find(item => item.id === props.selectedNodeId))
const branchEdges = computed(() => selectedNode.value?.gatewayMode === 'split'
  ? props.draft.edges.filter(item => item.source === selectedNode.value?.id)
  : [])
const canDelete = computed(() => !props.readonly && selectedNode.value && !['start', 'end'].includes(selectedNode.value.type))

function ensureAssignee() {
  if (!selectedNode.value) return undefined
  selectedNode.value.assignee ||= { type: 'manager', value: 'direct_manager' }
  return selectedNode.value.assignee
}

function updateApprovalMode(value: ApprovalMode) {
  if (!selectedNode.value) return
  selectedNode.value.approvalMode = value
  if (value !== 'countersign') delete selectedNode.value.completionRate
  else selectedNode.value.completionRate ||= 100
  emit('change')
}

function updateAssigneeType(value: AssigneeType) {
  const assignee = ensureAssignee()
  if (!assignee) return
  assignee.type = value
  assignee.value = value === 'manager' ? 'direct_manager' : value === 'department_leader' ? 'current_department' : ''
  emit('change')
}

function updateAssigneeValue(value: string) {
  const assignee = ensureAssignee()
  if (!assignee) return
  assignee.value = value
  emit('change')
}

function updateCompletionRate(value: number | undefined) {
  if (!selectedNode.value) return
  selectedNode.value.completionRate = Number(value || 100)
  emit('change')
}

function assigneePlaceholder(type?: AssigneeType) {
  return ({
    user: '输入用户 ID，多个用逗号分隔',
    role: '输入角色编码',
    department_leader: '例如：current_department',
    manager: '例如：direct_manager',
    variable: '输入流程变量名',
  } as Record<AssigneeType, string>)[type || 'manager']
}

function updateEdge(edge: WorkflowEdge, key: 'name', value: unknown) {
  edge[key] = String(value || '')
  emit('change')
}

function updateBranchName(edge: WorkflowEdge, value: string) {
  updateEdge(edge, 'name', value)
}

function ensureCondition(edge: WorkflowEdge) {
  edge.condition ||= { field: '', operator: 'eq', value: '' }
  return edge.condition
}

function updateCondition(edge: WorkflowEdge, key: 'field' | 'operator' | 'value', value: unknown) {
  const condition = ensureCondition(edge)
  if (key === 'operator') condition.operator = value as ConditionOperator
  else if (key === 'field') condition.field = String(value || '')
  else condition.value = String(value ?? '')
  emit('change')
}

function updateConditionField(edge: WorkflowEdge, value: string) {
  updateCondition(edge, 'field', value)
}

function updateConditionOperator(edge: WorkflowEdge, value: ConditionOperator) {
  updateCondition(edge, 'operator', value)
}

function updateConditionValue(edge: WorkflowEdge, value: string) {
  updateCondition(edge, 'value', value)
}

function updateDefaultBranch(edge: WorkflowEdge, value: string | number | boolean) {
  setDefaultBranch(edge, Boolean(value))
}

function setDefaultBranch(edge: WorkflowEdge, checked: boolean) {
  if (!selectedNode.value) return
  if (checked) {
    branchEdges.value.forEach(item => { item.default = item.id === edge.id })
    delete edge.condition
  } else {
    edge.default = false
    ensureCondition(edge)
  }
  emit('change')
}
</script>

<style scoped>
.node-inspector { width: 100%; min-height: 100%; background: #fff; }
.inspector-content { padding: 0 2px 24px; }
.inspector-section { padding: 18px 0; border-bottom: 1px solid #edf0f5; }
.inspector-section h4 { margin: 0 0 14px; color: #334155; font-size: 13px; }
.field-label { display: block; margin-bottom: 7px; color: #64748b; font-size: 12px; font-weight: 500; }
.field-label.spacing { margin-top: 14px; }
.section-title-row { display: flex; align-items: center; justify-content: space-between; margin-bottom: 12px; }
.section-title-row h4 { margin: 0; }
.branch-editor { margin-bottom: 12px; padding: 12px; border: 1px solid #e5eaf0; border-radius: 7px; background: #fafbfd; }
.branch-editor:last-child { margin-bottom: 0; }
.branch-editor__heading { display: flex; align-items: center; justify-content: space-between; margin-bottom: 8px; color: #334155; font-size: 12px; }
.default-check { margin: 9px 0; }
.condition-row { display: grid; grid-template-columns: 105px 1fr; gap: 7px; margin-top: 7px; }
.join-note { display: flex; align-items: flex-start; gap: 8px; color: #64748b; font-size: 12px; line-height: 1.6; }
.join-note .el-icon { margin-top: 3px; color: #0f766e; }
.inspector-footer { padding-top: 18px; }
</style>
