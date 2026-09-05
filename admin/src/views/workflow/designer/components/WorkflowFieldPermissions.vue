<template>
  <div class="permission-designer">
    <header class="permission-heading">
      <div>
        <strong>字段权限</strong>
        <p>按流程节点控制字段隐藏、只读或可编辑。审批节点只有设为可编辑的字段才能提交修改。</p>
      </div>
      <div class="permission-legend">
        <span><i class="hidden" />隐藏</span>
        <span><i class="read" />只读</span>
        <span><i class="write" />可编辑</span>
      </div>
    </header>

    <el-empty v-if="permissionFieldEntries.length === 0" :image-size="92" description="请先在表单设计中添加业务字段" />
    <el-empty v-else-if="permissionNodes.length === 0" :image-size="92" description="流程中暂无可配置字段权限的节点" />

    <div v-else class="permission-table-wrap">
      <table class="permission-table">
        <thead>
          <tr>
            <th class="field-column field-column-header">表单字段</th>
            <th v-for="node in permissionNodes" :key="node.id">
              <span class="node-type">{{ node.type === 'start' ? '发起节点' : node.type === 'handle' ? '办理节点' : '审批节点' }}</span>
              <strong>{{ node.name }}</strong>
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="entry in permissionFieldEntries" :key="entry.field.key">
            <th class="field-column field-column-body">
              <span v-if="entry.group" class="field-group-title">{{ entry.group.label }}</span>
              <strong>{{ entry.field.label }}</strong>
              <span>{{ entry.field.key }}</span>
            </th>
            <td v-for="node in permissionNodes" :key="node.id">
              <div class="permission-cell">
                <el-select
                  :model-value="fieldAccess(node, entry.field)"
                  :disabled="readonly"
                  :class="`access-${fieldAccess(node, entry.field)}`"
                  @change="updateFieldAccess(node, entry.field, $event)"
                >
                  <el-option label="隐藏" value="hidden" />
                  <el-option label="只读" value="read" />
                  <el-option v-if="entry.field.type !== 'calculation'" label="可编辑" value="write" />
                </el-select>
                <span v-if="entry.field.type === 'calculation'" class="permission-cell__hint">计算字段由系统生成，不能编辑</span>
                <div v-if="entry.field.type === 'detail_list'" class="row-action-slot">
                  <el-checkbox-group
                    v-if="(node.type === 'approval' || node.type === 'handle') && fieldAccess(node, entry.field) === 'write'"
                    :model-value="fieldActions(node, entry.field.key)"
                    :disabled="readonly"
                    class="row-action-options"
                    @change="updateFieldActions(node, entry.field.key, $event)"
                  >
                    <el-checkbox value="add">新增行</el-checkbox>
                    <el-checkbox value="delete">删除行</el-checkbox>
                  </el-checkbox-group>
                </div>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { computed } from 'vue'
import type { WorkflowDetailRowAction, WorkflowDraft, WorkflowFieldAccess, WorkflowFieldPermission, WorkflowFormField, WorkflowNode } from '../../types'
import { workflowDataFieldEntries } from '../../formLayout'
import { workflowPermissionNodes } from '../flowTree'

const props = defineProps<{ draft: WorkflowDraft; readonly?: boolean }>()
const emit = defineEmits<{ change: [] }>()

const permissionNodes = computed(() => workflowPermissionNodes(props.draft))
const permissionFieldEntries = computed(() => workflowDataFieldEntries(props.draft.form))

function defaultAccess(node: WorkflowNode, field?: WorkflowFormField): WorkflowFieldAccess {
  return field?.type === 'calculation' ? 'read' : node.type === 'start' ? 'write' : 'read'
}

function fieldAccess(node: WorkflowNode, field: WorkflowFormField): WorkflowFieldAccess {
  const access = node.formPermissions?.find(item => item.field === field.key)?.access || defaultAccess(node, field)
  return field.type === 'calculation' && access === 'write' ? 'read' : access
}

function fieldPermission(node: WorkflowNode, fieldKey: string): WorkflowFieldPermission | undefined {
  return node.formPermissions?.find(item => item.field === fieldKey)
}

function fieldActions(node: WorkflowNode, fieldKey: string): WorkflowDetailRowAction[] {
  return (fieldPermission(node, fieldKey)?.actions || []).filter((action): action is WorkflowDetailRowAction => action === 'add' || action === 'delete')
}

function ensureFieldPermission(node: WorkflowNode, fieldKey: string, access: WorkflowFieldAccess = defaultAccess(node)): WorkflowFieldPermission {
  node.formPermissions ||= []
  let permission = fieldPermission(node, fieldKey)
  if (!permission) {
    permission = { field: fieldKey, access }
    node.formPermissions.push(permission)
  }
  return permission
}

function updateFieldAccess(node: WorkflowNode, field: WorkflowFormField, access: WorkflowFieldAccess) {
  if (props.readonly) return
  const permission = ensureFieldPermission(node, field.key, defaultAccess(node, field))
  permission.access = field.type === 'calculation' && access === 'write' ? 'read' : access
  if (permission.access !== 'write') delete permission.actions
  emit('change')
}

function updateFieldActions(node: WorkflowNode, fieldKey: string, actions: unknown) {
  if (props.readonly) return
  const permission = ensureFieldPermission(node, fieldKey)
  permission.access = 'write'
  permission.actions = Array.isArray(actions)
    ? actions.filter((action): action is WorkflowDetailRowAction => action === 'add' || action === 'delete')
    : []
  emit('change')
}
</script>

<style scoped>
.permission-designer { width: 100%; min-height: 100%; overflow: hidden; background: #fff; }
.permission-heading { display: flex; align-items: center; justify-content: space-between; gap: 24px; min-height: 72px; padding: 0 20px; border-bottom: 1px solid #e2e8f0; }
.permission-heading strong { color: #1f2937; font-size: 15px; }
.permission-heading p { margin: 5px 0 0; color: #8492a6; font-size: 12px; }
.permission-legend { display: flex; flex: 0 0 auto; gap: 16px; color: #64748b; font-size: 11px; }
.permission-legend span { display: inline-flex; align-items: center; gap: 5px; }
.permission-legend i { width: 8px; height: 8px; border-radius: 50%; }
.permission-legend .hidden { background: #94a3b8; }
.permission-legend .read { background: #f59e0b; }
.permission-legend .write { background: #10b981; }
.permission-table-wrap { height: calc(100% - 73px); overflow: auto; padding: 0 20px 24px; }
.permission-table { width: max(100%, 780px); margin-top: 16px; border-spacing: 0; border-collapse: separate; border: 1px solid #dfe6ee; border-radius: 8px; background: #fff; }
.permission-table th, .permission-table td { min-width: 190px; padding: 12px; border-right: 1px solid #e8edf3; border-bottom: 1px solid #e8edf3; text-align: left; vertical-align: top; }
.permission-table tr:last-child th, .permission-table tr:last-child td { border-bottom: 0; }
.permission-table th:last-child, .permission-table td:last-child { border-right: 0; }
.permission-table thead th { position: sticky; top: 0; z-index: 2; color: #475569; background: #f8fafc; }
.permission-table .field-column { position: sticky; left: 0; z-index: 1; min-width: 220px; }
.permission-table thead .field-column { z-index: 3; border-bottom-color: #cbd5e1; background: #eef2f7; box-shadow: inset 0 -1px 0 #cbd5e1; }
.permission-table tbody .field-column { background: #fbfcfe; }
.permission-table th strong { display: block; color: #273548; font-size: 13px; }
.permission-table th span { display: block; margin-top: 4px; color: #94a3b8; font-size: 11px; font-weight: 400; }
.permission-table th .node-type { margin: 0 0 5px; color: #1677ff; }
.permission-table th .field-group-title { margin: 0 0 6px; color: #1677ff; font-weight: 600; }
.permission-cell { display: flex; min-width: 0; flex-direction: column; gap: 8px; }
.permission-cell__hint { color: #8492a6; font-size: 11px; line-height: 1.5; }
.permission-table :deep(.el-select) { width: 100%; }
.permission-table :deep(.access-hidden .el-select__wrapper) { color: #64748b; background: #f1f5f9; }
.permission-table :deep(.access-read .el-select__wrapper) { color: #b45309; background: #fffbeb; }
.permission-table :deep(.access-write .el-select__wrapper) { color: #047857; background: #ecfdf5; }
.row-action-slot { min-height: 24px; }
.row-action-options { display: flex; flex-wrap: wrap; gap: 4px 10px; }
.row-action-options :deep(.el-checkbox) { height: 24px; margin-right: 0; color: #64748b; font-size: 12px; }
</style>
