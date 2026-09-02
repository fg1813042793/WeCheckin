<template>
  <BaseEdge
    :id="id"
    :path="edgePath"
    :marker-start="markerStart"
    :marker-end="markerEnd"
    :interaction-width="22"
    :style="edgeStyle"
  />
  <EdgeLabelRenderer>
    <div
      class="workflow-curve-edge__label nodrag nopan"
      :style="{ transform: `translate(-50%, -50%) translate(${labelX}px, ${labelY}px)` }"
    >
      <span v-if="branchLabel" class="workflow-curve-edge__branch" :title="branchLabel">{{ branchLabel }}</span>
      <FlowInsertButton
        compact
        :edge-id="data.edge.id"
        :disabled="data.readonly"
        @insert="data.onInsert"
      />
    </div>
  </EdgeLabelRenderer>
</template>

<script lang="ts" setup>
import { computed } from 'vue'
import { BaseEdge, EdgeLabelRenderer, getSimpleBezierPath } from '@vue-flow/core'
import type { CSSProperties } from 'vue'
import type { EdgeProps } from '@vue-flow/core'
import type { WorkflowEdge, WorkflowInsertableNodeType } from '../../types'
import FlowInsertButton from './FlowInsertButton.vue'

const props = defineProps<EdgeProps<{
  edge: WorkflowEdge
  readonly: boolean
  onInsert: (payload: { edgeId: string; type: WorkflowInsertableNodeType }) => void
}>>()

const path = computed(() => getSimpleBezierPath({
  sourceX: props.sourceX,
  sourceY: props.sourceY,
  sourcePosition: props.sourcePosition,
  targetX: props.targetX,
  targetY: props.targetY,
  targetPosition: props.targetPosition,
}))
const edgePath = computed(() => path.value[0])
const labelX = computed(() => path.value[1])
const labelY = computed(() => path.value[2])
const edgeStyle = computed<CSSProperties>(() => ({
  ...props.style,
  stroke: props.selected ? '#4b8ffb' : '#aeb8c5',
  strokeWidth: props.selected ? 2 : 1.5,
}))
const branchLabel = computed(() => {
  const edge = props.data.edge
  if (!edge.name && !edge.default) return ''
  return edge.default ? `${edge.name || '默认分支'} · 默认` : edge.name || ''
})
</script>

<style scoped>
.workflow-curve-edge__label { position: absolute; z-index: 4; display: flex; flex-direction: column; align-items: center; pointer-events: all; }
.workflow-curve-edge__branch { display: block; max-width: 150px; margin-bottom: 5px; padding: 3px 7px; overflow: hidden; border: 1px solid #d9e0e8; border-radius: 4px; color: #526174; background: rgb(255 255 255 / 94%); box-shadow: 0 2px 6px rgb(31 41 55 / 7%); font-size: 10px; text-overflow: ellipsis; white-space: nowrap; }
</style>
