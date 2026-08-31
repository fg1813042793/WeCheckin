<template>
  <div v-if="node.type === 'end'" class="end-node">
    <span />
    <b>{{ node.name || '流程结束' }}</b>
  </div>
  <button
    v-else
    type="button"
    class="workflow-node-card"
    :class="[`workflow-node-card--${node.type}`, { selected }]"
    @click.stop="$emit('select', node.id)"
  >
    <span class="workflow-node-card__header">
      <span>{{ node.type === 'start' ? '发起人' : '审批人' }}</span>
      <el-icon><ArrowRight /></el-icon>
    </span>
    <span class="workflow-node-card__body">
      <b>{{ node.name }}</b>
      <small>{{ description }}</small>
    </span>
  </button>
</template>

<script lang="ts" setup>
import { computed } from 'vue'
import { ArrowRight } from '@element-plus/icons-vue'
import type { WorkflowNode } from '../../types'

const props = defineProps<{ node: WorkflowNode; selected?: boolean }>()
defineEmits<{ select: [nodeId: string] }>()

const description = computed(() => {
  if (props.node.type === 'start') return '所有人均可发起'
  const mode = ({ single: '单人审批', sequential: '依次审批', parallel: '并行审批', countersign: '会签审批' } as Record<string, string>)[props.node.approvalMode || '']
  const assignee = ({
    user: '指定用户',
    role: '指定角色',
    department_leader: '部门负责人',
    manager: '直属上级',
    variable: '流程变量',
    org_identity: '组织审批身份'
  } as Record<string, string>)[props.node.assignee?.type || '']
  return [assignee, mode].filter(Boolean).join(' · ') || '请配置审批规则'
})
</script>

<style scoped>
.workflow-node-card { display: block; width: 238px; padding: 0; overflow: hidden; border: 1px solid #d9dee7; border-radius: 5px; color: #344054; background: #fff; text-align: left; box-shadow: 0 3px 9px rgb(31 41 55 / 8%); cursor: pointer; transition: border-color .18s, box-shadow .18s, transform .18s; }
.workflow-node-card:hover { border-color: #9dbff7; box-shadow: 0 6px 15px rgb(31 41 55 / 11%); transform: translateY(-1px); }
.workflow-node-card.selected { border-color: #4b8ffb; box-shadow: 0 0 0 2px rgb(75 143 251 / 13%), 0 6px 15px rgb(31 41 55 / 10%); }
.workflow-node-card__header { display: flex; height: 30px; align-items: center; justify-content: space-between; padding: 0 11px; color: #fff; background: #60749f; font-size: 12px; }
.workflow-node-card--approval .workflow-node-card__header { background: #ef9b48; }
.workflow-node-card__body { display: block; min-height: 58px; padding: 10px 12px; }
.workflow-node-card__body b, .workflow-node-card__body small { display: block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.workflow-node-card__body b { font-size: 13px; font-weight: 600; }
.workflow-node-card__body small { margin-top: 7px; color: #8a94a3; font-size: 11px; }
.end-node { display: flex; flex-direction: column; align-items: center; gap: 7px; color: #7c8798; font-size: 12px; }
.end-node span { width: 9px; height: 9px; border-radius: 50%; background: #c7ced8; }
.end-node b { font-weight: 500; }
</style>
