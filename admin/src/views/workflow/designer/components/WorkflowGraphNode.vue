<template>
  <div
    class="workflow-graph-node"
    :class="{ 'workflow-graph-node--gateway': data.node.gatewayMode }"
  >
    <Handle
      v-for="anchor in anchors"
      :id="anchor.id"
      :key="anchor.id"
      class="workflow-anchor"
      :class="[
        `workflow-anchor--${anchor.id}`,
        {
          'workflow-anchor--connected': data.connectedHandles.includes(anchor.id),
          'workflow-anchor--editing': data.connectable,
        },
      ]"
      type="source"
      :position="anchor.position"
      :connectable="data.connectable"
    />

    <div
      v-if="data.node.gatewayMode"
      class="gateway-node"
      :class="[`gateway-node--${data.node.type}`, { selected: data.selected }]"
    >
      <button
        class="gateway-node__main"
        type="button"
        :title="data.node.name"
        @click.stop="data.onSelect(data.node.id)"
      >
        <span class="gateway-node__content">
          <b>{{ data.node.name }}</b>
          <small>{{ data.node.gatewayMode === 'split' ? '分支' : '汇聚' }}</small>
        </span>
      </button>
      <button
        v-if="data.node.gatewayMode === 'split'"
        class="gateway-node__add nodrag nopan"
        type="button"
        :disabled="data.readonly"
        :title="data.node.type === 'exclusive' ? '添加条件分支' : '添加并行分支'"
        @click.stop="data.onAddBranch(data.node.id)"
      >
        <el-icon><Plus /></el-icon>
      </button>
    </div>

    <div v-else class="workflow-graph-node__card" @click.stop="data.onSelect(data.node.id)">
      <WorkflowNodeCard
        :node="data.node"
        :selected="data.selected"
        @select="data.onSelect"
      />
    </div>

  </div>
</template>

<script lang="ts" setup>
import { Handle, Position } from '@vue-flow/core'
import type { NodeProps } from '@vue-flow/core'
import { Plus } from '@element-plus/icons-vue'
import type { WorkflowEdgeHandle, WorkflowNode } from '../../types'
import WorkflowNodeCard from './WorkflowNodeCard.vue'

const anchors = [
  { id: 'top', position: Position.Top },
  { id: 'right', position: Position.Right },
  { id: 'bottom', position: Position.Bottom },
  { id: 'left', position: Position.Left },
] as const

defineProps<NodeProps<{
  node: WorkflowNode
  selected: boolean
  readonly: boolean
  connectable: boolean
  connectedHandles: WorkflowEdgeHandle[]
  onSelect: (nodeId: string) => void
  onAddBranch: (nodeId: string) => void
}>>()
</script>

<style scoped>
.workflow-graph-node { position: relative; display: grid; min-width: 120px; place-items: center; }
.workflow-graph-node--gateway { width: 136px; height: 136px; min-width: 136px; }
.workflow-graph-node__card { width: max-content; }
.gateway-node { position: relative; display: grid; width: 136px; height: 136px; place-items: center; }
.gateway-node__main { display: grid; width: 96px; height: 96px; box-sizing: border-box; place-items: center; padding: 0; border: 2px solid #27a56c; color: #158a58; background: #fff; box-shadow: 0 4px 12px rgb(31 41 55 / 8%); cursor: pointer; transform: rotate(45deg); transition: border-color .16s, background .16s, box-shadow .16s; }
.gateway-node__main:hover, .gateway-node.selected .gateway-node__main { border-color: #158a58; background: #f3fff9; box-shadow: 0 6px 18px rgb(21 138 88 / 16%); }
.gateway-node--parallel .gateway-node__main { border-color: #4b8ffb; color: #2878d0; }
.gateway-node--parallel .gateway-node__main:hover, .gateway-node--parallel.selected .gateway-node__main { border-color: #2878d0; background: #f4f8ff; box-shadow: 0 6px 18px rgb(40 120 208 / 16%); }
.gateway-node__content { display: grid; width: 72px; gap: 4px; place-items: center; text-align: center; transform: rotate(-45deg); }
.gateway-node__content b { display: -webkit-box; max-width: 72px; overflow: hidden; font-size: 12px; font-weight: 600; line-height: 16px; overflow-wrap: anywhere; -webkit-box-orient: vertical; -webkit-line-clamp: 2; }
.gateway-node__content small { font-size: 10px; line-height: 12px; opacity: .68; }
.gateway-node__add { position: absolute; top: 50%; right: -36px; display: grid; width: 25px; height: 25px; place-items: center; padding: 0; border: 0; border-radius: 50%; color: #fff; background: #25a86f; box-shadow: 0 2px 7px rgb(37 168 111 / 26%); cursor: pointer; transform: translateY(-50%); }
.gateway-node--parallel .gateway-node__add { background: #4b8ffb; box-shadow: 0 2px 7px rgb(75 143 251 / 26%); }
.gateway-node__add:hover { background: #168b5a; }
.gateway-node--parallel .gateway-node__add:hover { background: #2878d0; }
.gateway-node__add:disabled { cursor: not-allowed; opacity: .45; }
.workflow-graph-node--gateway > .workflow-anchor { z-index: 4; }
.workflow-anchor { opacity: 0; width: 9px; height: 9px; border: 2px solid #fff; background: #8f9daf; box-shadow: 0 1px 4px rgb(31 41 55 / 16%); transition: width .14s, height .14s, background .14s, opacity .14s; }
.workflow-anchor--connected, .workflow-anchor--editing, .workflow-graph-node:hover .workflow-anchor { opacity: 1; }
.workflow-anchor:hover, .workflow-anchor.connecting { width: 13px; height: 13px; background: #4b8ffb; }
</style>
