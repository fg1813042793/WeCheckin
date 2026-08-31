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

    <el-empty v-if="draft.form.length === 0" :image-size="92" description="请先在表单设计中添加字段" />
    <el-empty v-else-if="permissionNodes.length === 0" :image-size="92" description="流程中暂无可配置字段权限的节点" />

    <div v-else class="permission-table-wrap">
      <table class="permission-table">
        <thead>
          <tr>
            <th class="field-column">表单字段</th>
            <th v-for="node in permissionNodes" :key="node.id">
              <span class="node-type">{{ node.type === 'start' ? '发起节点' : '审批节点' }}</span>
              <strong>{{ node.name }}</strong>
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="field in draft.form" :key="field.key">
            <th class="field-column">
              <strong>{{ field.label }}</strong>
              <span>{{ field.key }}</span>
            </th>
            <td v-for="node in permissionNodes" :key="node.id">
              <el-select
                :model-value="fieldAccess(node, field.key)"
                :disabled="readonly"
                :class="`access-${fieldAccess(node, field.key)}`"
                @change="updateFieldAccess(node, field.key, $event)"
              >
                <el-option label="隐藏" value="hidden" />
                <el-option label="只读" value="read" />
                <el-option label="可编辑" value="write" />
              </el-select>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { computed } from 'vue'
import type { WorkflowDraft, WorkflowFieldAccess, WorkflowNode } from '../../types'

const props = defineProps<{ draft: WorkflowDraft; readonly?: boolean }>()
const emit = defineEmits<{ change: [] }>()

const permissionNodes = computed(() => props.draft.nodes.filter(node => node.type === 'start' || node.type === 'approval'))

function defaultAccess(node: WorkflowNode): WorkflowFieldAccess {
  return node.type === 'start' ? 'write' : 'read'
}

function fieldAccess(node: WorkflowNode, fieldKey: string): WorkflowFieldAccess {
  return node.formPermissions?.find(item => item.field === fieldKey)?.access || defaultAccess(node)
}

function updateFieldAccess(node: WorkflowNode, fieldKey: string, access: WorkflowFieldAccess) {
  if (props.readonly) return
  node.formPermissions ||= []
  const permission = node.formPermissions.find(item => item.field === fieldKey)
  if (permission) permission.access = access
  else node.formPermissions.push({ field: fieldKey, access })
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
.permission-table-wrap { height: calc(100% - 73px); overflow: auto; padding: 16px 20px 24px; }
.permission-table { width: max(100%, 780px); border-spacing: 0; border-collapse: separate; border: 1px solid #dfe6ee; border-radius: 8px; background: #fff; }
.permission-table th, .permission-table td { min-width: 190px; padding: 12px; border-right: 1px solid #e8edf3; border-bottom: 1px solid #e8edf3; text-align: left; }
.permission-table tr:last-child th, .permission-table tr:last-child td { border-bottom: 0; }
.permission-table th:last-child, .permission-table td:last-child { border-right: 0; }
.permission-table thead th { position: sticky; top: 0; z-index: 2; color: #475569; background: #f8fafc; }
.permission-table .field-column { position: sticky; left: 0; z-index: 1; min-width: 220px; background: #fbfcfe; }
.permission-table thead .field-column { z-index: 3; }
.permission-table th strong { display: block; color: #273548; font-size: 13px; }
.permission-table th span { display: block; margin-top: 4px; color: #94a3b8; font-size: 11px; font-weight: 400; }
.permission-table th .node-type { margin: 0 0 5px; color: #1677ff; }
.permission-table :deep(.el-select) { width: 100%; }
.permission-table :deep(.access-hidden .el-select__wrapper) { color: #64748b; background: #f1f5f9; }
.permission-table :deep(.access-read .el-select__wrapper) { color: #b45309; background: #fffbeb; }
.permission-table :deep(.access-write .el-select__wrapper) { color: #047857; background: #ecfdf5; }
</style>
