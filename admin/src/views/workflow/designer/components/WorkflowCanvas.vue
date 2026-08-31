<template>
  <main class="workflow-canvas" @click="$emit('select', '')">
    <div class="canvas-toolbar">
      <div>
        <strong>流程设计</strong>
        <span>点击连线中的加号添加节点，点击节点进入配置</span>
      </div>
      <div class="canvas-toolbar__actions">
        <el-button-group>
          <el-button icon="Minus" title="缩小" :disabled="zoom <= 75" @click.stop="changeZoom(-10)" />
          <el-button class="zoom-value" @click.stop="resetZoom">{{ zoom }}%</el-button>
          <el-button icon="Plus" title="放大" :disabled="zoom >= 125" @click.stop="changeZoom(10)" />
        </el-button-group>
      </div>
    </div>
    <div class="canvas-scroll">
      <div class="canvas-stage" :style="{ zoom: `${zoom}%` }">
        <WorkflowSequence
          :sequence="tree"
          :selected-node-id="selectedNodeId"
          :readonly="readonly"
          @select="$emit('select', $event)"
          @insert="$emit('insert', $event)"
          @add-branch="$emit('add-branch', $event)"
        />
      </div>
    </div>
  </main>
</template>

<script lang="ts" setup>
import { computed, ref } from 'vue'
import type { WorkflowDraft, WorkflowNodeType } from '../../types'
import { buildWorkflowTree } from '../flowTree'
import WorkflowSequence from './WorkflowSequence.vue'

const props = defineProps<{ draft: WorkflowDraft; selectedNodeId: string; readonly?: boolean }>()
defineEmits<{
  select: [nodeId: string]
  insert: [payload: { edgeId: string; type: Extract<WorkflowNodeType, 'approval' | 'exclusive' | 'parallel'> }]
  'add-branch': [splitId: string]
}>()

const zoom = ref(100)
const tree = computed(() => buildWorkflowTree(props.draft))

function changeZoom(step: number) {
  zoom.value = Math.min(125, Math.max(75, zoom.value + step))
}

function resetZoom() {
  zoom.value = 100
}
</script>

<style scoped>
.workflow-canvas { min-width: 0; flex: 1; display: flex; flex-direction: column; background: #f4f6f8; }
.canvas-toolbar { display: flex; min-height: 55px; align-items: center; justify-content: space-between; padding: 0 18px; border-bottom: 1px solid #e2e6ec; background: #fff; }
.canvas-toolbar strong, .canvas-toolbar span { display: block; }
.canvas-toolbar strong { color: #263244; font-size: 14px; }
.canvas-toolbar > div > span { margin-top: 3px; color: #98a2b3; font-size: 11px; }
.canvas-toolbar__actions :deep(.el-button) { width: 34px; }
.canvas-toolbar__actions :deep(.zoom-value) { width: 54px; padding: 0; color: #667085; font-size: 11px; }
.canvas-scroll { flex: 1; overflow: auto; padding: 30px 36px 90px; }
.canvas-stage { width: max-content; min-width: 100%; margin: 0 auto; transform-origin: top center; }
@media (max-width: 900px) {
  .canvas-toolbar { padding: 0 12px; }
  .canvas-toolbar > div > span { display: none; }
  .canvas-scroll { padding: 24px 18px 70px; }
}
</style>
