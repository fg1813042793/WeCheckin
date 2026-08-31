<template>
  <div class="workflow-sequence">
    <template v-for="item in sequence.items" :key="item.kind === 'node' ? item.node.id : item.split.id">
      <template v-if="item.kind === 'node'">
        <WorkflowNodeCard :node="item.node" :selected="item.node.id === selectedNodeId" @select="$emit('select', $event)" />
        <FlowInsertButton v-if="item.outgoingEdgeId" :edge-id="item.outgoingEdgeId" :disabled="readonly" @insert="$emit('insert', $event)" />
      </template>

      <template v-else>
        <section class="branch-group" :class="`branch-group--${item.split.type}`">
          <button
            class="branch-group__label"
            type="button"
            :disabled="readonly"
            :title="item.split.type === 'exclusive' ? '添加条件分支' : '添加并行分支'"
            @click.stop="$emit('add-branch', item.split.id)"
          >
            <el-icon><Plus /></el-icon>
            {{ item.split.type === 'exclusive' ? '添加条件' : '添加并行分支' }}
          </button>
          <div class="branch-group__lanes" :style="{ '--branch-count': item.branches.length }">
            <div v-for="(branch, index) in item.branches" :key="branch.edge.id" class="branch-lane">
              <button class="branch-card" type="button" :class="{ selected: item.split.id === selectedNodeId }" @click.stop="$emit('select', item.split.id)">
                <span class="branch-card__header">
                  <b>{{ branch.edge.name || `分支${index + 1}` }}</b>
                  <small>优先级 {{ index + 1 }}</small>
                </span>
                <span class="branch-card__summary">{{ conditionSummary(branch.edge, item.split.type) }}</span>
              </button>
              <FlowInsertButton :edge-id="branch.entryEdgeId" :disabled="readonly" @insert="$emit('insert', $event)" />
              <WorkflowSequence
                :sequence="branch.sequence"
                :selected-node-id="selectedNodeId"
                :readonly="readonly"
                @select="$emit('select', $event)"
                @insert="$emit('insert', $event)"
                @add-branch="$emit('add-branch', $event)"
              />
            </div>
          </div>
          <span class="branch-group__join" />
        </section>
        <FlowInsertButton v-if="item.outgoingEdgeId" :edge-id="item.outgoingEdgeId" :disabled="readonly" @insert="$emit('insert', $event)" />
      </template>
    </template>
  </div>
</template>

<script lang="ts" setup>
import { Plus } from '@element-plus/icons-vue'
import type { WorkflowEdge, WorkflowNodeType } from '../../types'
import type { WorkflowTreeSequence } from '../flowTree'
import FlowInsertButton from './FlowInsertButton.vue'
import WorkflowNodeCard from './WorkflowNodeCard.vue'

defineOptions({ name: 'WorkflowSequence' })
defineProps<{ sequence: WorkflowTreeSequence; selectedNodeId: string; readonly?: boolean }>()
defineEmits<{
  select: [nodeId: string]
  insert: [payload: { edgeId: string; type: Extract<WorkflowNodeType, 'approval' | 'exclusive' | 'parallel'> }]
  'add-branch': [splitId: string]
}>()

function conditionSummary(edge: WorkflowEdge, type: WorkflowNodeType) {
  if (type === 'parallel') return '与其他分支同时执行'
  if (edge.default) return '其他条件未命中时进入'
  if (!edge.condition?.field) return '请设置条件'
  const operator = ({ eq: '=', ne: '!=', gt: '>', gte: '>=', lt: '<', lte: '<=' } as Record<string, string>)[edge.condition.operator] || edge.condition.operator
  return `${edge.condition.field} ${operator} ${String(edge.condition.value ?? '')}`
}
</script>

<style scoped>
.workflow-sequence { display: flex; flex-direction: column; align-items: center; }
.branch-group { position: relative; min-width: 620px; padding-top: 34px; }
.branch-group__label { position: absolute; top: 0; left: 50%; z-index: 2; display: inline-flex; height: 27px; align-items: center; gap: 4px; padding: 0 10px; border: 1px solid #bfe7cf; border-radius: 14px; color: #17905c; background: #fff; font-size: 11px; cursor: pointer; transform: translateX(-50%); }
.branch-group--parallel .branch-group__label { border-color: #bdd7fa; color: #2878d0; }
.branch-group__label:disabled { cursor: not-allowed; opacity: .55; }
.branch-group__lanes { position: relative; display: grid; grid-template-columns: repeat(var(--branch-count), minmax(260px, 1fr)); padding: 24px 42px 38px; }
.branch-group__lanes::before, .branch-group__lanes::after { position: absolute; left: 50%; width: calc(100% - 84px); height: 1px; content: ''; background: #c8d0da; transform: translateX(-50%); }
.branch-group__lanes::before { top: 0; }
.branch-group__lanes::after { bottom: 0; }
.branch-lane { position: relative; display: flex; min-width: 260px; flex-direction: column; align-items: center; padding: 0 24px; }
.branch-lane::before, .branch-lane::after { position: absolute; left: 50%; width: 1px; content: ''; background: #c8d0da; }
.branch-lane::before { top: -24px; height: 24px; }
.branch-lane::after { bottom: -38px; height: 38px; }
.branch-card { display: block; width: 210px; min-height: 76px; padding: 11px 12px; border: 1px solid #dce2e9; border-radius: 5px; color: #344054; background: #fff; text-align: left; box-shadow: 0 3px 9px rgb(31 41 55 / 7%); cursor: pointer; transition: border-color .18s, box-shadow .18s; }
.branch-card:hover, .branch-card.selected { border-color: #70c89b; box-shadow: 0 5px 13px rgb(31 41 55 / 10%); }
.branch-group--parallel .branch-card:hover, .branch-group--parallel .branch-card.selected { border-color: #78adf3; }
.branch-card__header { display: flex; align-items: center; justify-content: space-between; gap: 10px; }
.branch-card__header b { color: #16935d; font-size: 12px; font-weight: 600; }
.branch-group--parallel .branch-card__header b { color: #2878d0; }
.branch-card__header small { color: #a0a8b5; font-size: 10px; white-space: nowrap; }
.branch-card__summary { display: block; margin-top: 12px; overflow: hidden; color: #596579; font-size: 11px; text-overflow: ellipsis; white-space: nowrap; }
.branch-group__join { position: absolute; bottom: 0; left: 50%; width: 8px; height: 8px; border-radius: 50%; background: #c8d0da; transform: translate(-50%, 50%); }
</style>
