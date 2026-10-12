<template>
  <main
    ref="canvasRoot"
    class="workflow-canvas"
    tabindex="0"
    @keydown.delete.prevent="deleteSelectedEdge"
    @keydown.backspace.prevent="deleteSelectedEdge"
  >
    <div class="canvas-toolbar">
      <div>
        <strong>流程设计</strong>
        <span>拖动节点调整布局，拖动画布平移视图，点击连线中的加号添加节点</span>
      </div>
      <div class="canvas-toolbar__actions">
        <el-button-group>
          <el-button icon="Minus" title="缩小" :disabled="zoomPercent <= 35" @click.stop="zoomOutCanvas" />
          <el-button class="zoom-value" title="适应画布" @click.stop="fitCanvas">{{ zoomPercent }}%</el-button>
          <el-button icon="Plus" title="放大" :disabled="zoomPercent >= 160" @click.stop="zoomInCanvas" />
        </el-button-group>
        <el-button icon="Aim" title="适应画布" @click.stop="fitCanvas" />
        <el-button icon="Operation" title="自动整理布局" :disabled="readonly" @click.stop="applyAutoLayout" />
      </div>
    </div>
    <VueFlow
      id="workflow-designer"
      class="workflow-flow"
      :nodes="canvasNodes"
      :edges="canvasEdges"
      :node-types="nodeTypes"
      :edge-types="edgeTypes"
      :nodes-draggable="!readonly"
      :nodes-connectable="!readonly"
      :edges-updatable="!readonly"
      :connection-mode="ConnectionMode.Loose"
      :elements-selectable="!readonly"
      :delete-key-code="null"
      :min-zoom="0.35"
      :max-zoom="1.6"
      :fit-view-on-init="false"
      :pan-on-drag="true"
      :zoom-on-double-click="false"
      @pane-click="clearCanvasSelection"
      @node-drag-stop="persistNodePosition"
      @connect-start="connecting = true"
      @connect="createConnection"
      @connect-end="connecting = false"
      @edge-click="selectEdge"
      @edge-update-start="beginEdgeUpdate"
      @edge-update="persistEdgeConnection"
      @edge-update-end="finishEdgeUpdate"
      @nodes-initialized="focusStart"
    />
  </main>
</template>

<script lang="ts" setup>
import { computed, markRaw, nextTick, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import {
  ConnectionMode,
  MarkerType,
  Position,
  VueFlow,
  useVueFlow,
} from '@vue-flow/core'
import type { Connection, Edge, EdgeMouseEvent, EdgeUpdateEvent, Node, NodeDragEvent } from '@vue-flow/core'
import '@vue-flow/core/dist/style.css'
import '@vue-flow/core/dist/theme-default.css'
import type { WorkflowDraft, WorkflowEdgeHandle, WorkflowInsertableNodeType } from '../../types'
import { connectWorkflowEdge, reconnectWorkflowEdge } from '../graph'
import { layoutWorkflow, WORKFLOW_NODE_WIDTH } from '../layout'
import WorkflowCurveEdge from './WorkflowCurveEdge.vue'
import WorkflowGraphNode from './WorkflowGraphNode.vue'

const props = defineProps<{ draft: WorkflowDraft; selectedNodeId: string; readonly?: boolean }>()
const emit = defineEmits<{
  select: [nodeId: string]
  insert: [payload: { edgeId: string; type: WorkflowInsertableNodeType }]
  'add-branch': [splitId: string]
  'delete-edge': [edgeId: string]
  change: []
}>()

const nodeTypes = { workflow: markRaw(WorkflowGraphNode) }
const edgeTypes = { workflow: markRaw(WorkflowCurveEdge) }
const { fitView, setCenter, viewport, zoomIn, zoomOut } = useVueFlow('workflow-designer')
const canvasRoot = ref<HTMLElement | null>(null)
const hasInitialViewport = ref(false)
const updatingEdgeId = ref('')
const selectedEdgeId = ref('')
const connecting = ref(false)
const zoomPercent = computed(() => Math.round(viewport.value.zoom * 100))
const topologySignature = computed(() => JSON.stringify({
  nodes: props.draft.nodes.map(node => `${node.id}:${node.type}:${node.gatewayMode || ''}`),
  edges: props.draft.edges.map(edge => `${edge.id}:${edge.source}:${edge.target}`),
}))

const canvasNodes = computed<Node[]>(() => {
  const automatic = layoutWorkflow(props.draft)
  const connectedHandles = new Map<string, Set<WorkflowEdgeHandle>>()
  const connectHandle = (nodeId: string, handle: WorkflowEdgeHandle) => {
    const handles = connectedHandles.get(nodeId) || new Set<WorkflowEdgeHandle>()
    handles.add(handle)
    connectedHandles.set(nodeId, handles)
  }
  props.draft.edges.forEach((edge) => {
    connectHandle(edge.source, edge.sourceHandle || 'bottom')
    connectHandle(edge.target, edge.targetHandle || 'top')
  })
  return props.draft.nodes.map(node => ({
    id: node.id,
    type: 'workflow',
    position: node.position || automatic.get(node.id) || { x: 0, y: 0 },
    sourcePosition: Position.Bottom,
    targetPosition: Position.Top,
    draggable: !props.readonly,
    dragHandle: '.workflow-node-drag-handle',
    selectable: false,
    data: {
      node,
      selected: node.id === props.selectedNodeId,
      readonly: Boolean(props.readonly),
      connectable: !props.readonly,
      showAnchors: Boolean(updatingEdgeId.value) || connecting.value || node.id === props.selectedNodeId,
      connectedHandles: [...(connectedHandles.get(node.id) || [])],
      onSelect: (nodeId: string) => emit('select', nodeId),
      onAddBranch: (nodeId: string) => emit('add-branch', nodeId),
    },
  }))
})

const canvasEdges = computed<Edge[]>(() => props.draft.edges.map(edge => ({
  id: edge.id,
  source: edge.source,
  target: edge.target,
  sourceHandle: edge.sourceHandle || 'bottom',
  targetHandle: edge.targetHandle || 'top',
  type: 'workflow',
  markerEnd: MarkerType.ArrowClosed,
  selected: edge.id === selectedEdgeId.value,
  selectable: !props.readonly,
  updatable: !props.readonly,
  events: { click: selectEdge },
  data: {
    edge,
    selected: edge.id === selectedEdgeId.value,
    readonly: Boolean(props.readonly),
    onInsert: (payload: { edgeId: string; type: WorkflowInsertableNodeType }) => emit('insert', payload),
    onDelete: (edgeId: string) => requestDeleteEdge(edgeId),
  },
})))

watch(topologySignature, (_current, previous) => {
  if (selectedEdgeId.value && !props.draft.edges.some(edge => edge.id === selectedEdgeId.value)) selectedEdgeId.value = ''
  if (!props.draft.nodes.some(node => !node.position)) return
  assignLayoutPositions(previous !== undefined)
}, { immediate: true })

function persistNodePosition(event: NodeDragEvent) {
  if (props.readonly) return
  event.nodes.forEach((dragged) => {
    const node = props.draft.nodes.find(item => item.id === dragged.id)
    if (!node) return
    node.position = {
      x: Math.round(dragged.position.x),
      y: Math.round(dragged.position.y),
    }
  })
  emit('change')
}

function clearCanvasSelection() {
  selectedEdgeId.value = ''
  emit('select', '')
}

function selectEdge(event: EdgeMouseEvent) {
  if (props.readonly) return
  event.event.stopPropagation()
  selectedEdgeId.value = event.edge.id
  emit('select', '')
  canvasRoot.value?.focus({ preventScroll: true })
}

function createConnection(connection: Connection) {
  if (props.readonly) return
  const created = connectWorkflowEdge(props.draft, connection)
  if (!created) {
    ElMessage.warning('无法创建该连线，请检查开始、结束节点或重复连接')
    return
  }
  selectedEdgeId.value = created.id
  emit('change')
}

function beginEdgeUpdate(event: EdgeMouseEvent) {
  if (props.readonly) return
  updatingEdgeId.value = event.edge.id
  selectedEdgeId.value = event.edge.id
}

function persistEdgeConnection(event: EdgeUpdateEvent) {
  if (props.readonly) return
  if (!reconnectWorkflowEdge(props.draft, event.edge.id, event.connection)) {
    ElMessage.warning('无法改接该连线，请检查开始、结束节点或重复连接')
    return
  }
  selectedEdgeId.value = event.edge.id
  emit('change')
}

function finishEdgeUpdate() {
  updatingEdgeId.value = ''
}

function requestDeleteEdge(edgeId: string) {
  if (props.readonly || !props.draft.edges.some(edge => edge.id === edgeId)) return
  emit('delete-edge', edgeId)
}

function deleteSelectedEdge() {
  if (!selectedEdgeId.value) return
  requestDeleteEdge(selectedEdgeId.value)
}

function assignLayoutPositions(notify = true) {
  const positions = layoutWorkflow(props.draft)
  props.draft.nodes.forEach((node) => {
    const position = positions.get(node.id)
    if (position) node.position = position
  })
  if (notify) emit('change')
  if (notify) void nextTick(fitCanvas)
}

function applyAutoLayout() {
  if (props.readonly) return
  assignLayoutPositions()
}

function fitCanvas() {
  void fitView({ padding: 0.2, minZoom: 0.35, maxZoom: 1.15, duration: 220 })
}

function focusStart() {
  if (hasInitialViewport.value) return
  const start = canvasNodes.value.find(node => node.data?.node?.type === 'start')
  if (!start) return
  hasInitialViewport.value = true
  void setCenter(
    start.position.x + WORKFLOW_NODE_WIDTH / 2,
    start.position.y + 315,
    { zoom: 0.82 },
  )
}

function zoomInCanvas() {
  void zoomIn({ duration: 140 })
}

function zoomOutCanvas() {
  void zoomOut({ duration: 140 })
}
</script>

<style scoped>
.workflow-canvas { min-width: 0; flex: 1; display: flex; flex-direction: column; background: #f4f6f8; }
.workflow-canvas:focus { outline: none; }
.canvas-toolbar { display: flex; min-height: 55px; align-items: center; justify-content: space-between; padding: 0 18px; border-bottom: 1px solid #e2e6ec; background: #fff; }
.canvas-toolbar strong, .canvas-toolbar span { display: block; }
.canvas-toolbar strong { color: #263244; font-size: 14px; }
.canvas-toolbar > div > span { margin-top: 3px; color: #98a2b3; font-size: 11px; }
.canvas-toolbar__actions { display: flex; align-items: center; gap: 8px; }
.canvas-toolbar__actions :deep(.el-button) { width: 34px; }
.canvas-toolbar__actions :deep(.zoom-value) { width: 54px; padding: 0; color: #667085; font-size: 11px; }
.workflow-flow { min-height: 0; flex: 1; background-color: #f4f6f8; background-image: linear-gradient(#e7ebf0 1px, transparent 1px), linear-gradient(90deg, #e7ebf0 1px, transparent 1px); background-size: 24px 24px; }
.workflow-flow :deep(.vue-flow__pane) { cursor: grab; }
.workflow-flow :deep(.vue-flow__pane.dragging) { cursor: grabbing; }
.workflow-flow :deep(.vue-flow__node-workflow) { cursor: grab; }
.workflow-flow :deep(.vue-flow__node-workflow.dragging) { cursor: grabbing; }
.workflow-flow :deep(.vue-flow__edge-path) { transition: stroke .16s, stroke-width .16s; }
.workflow-flow :deep(.vue-flow__edge:hover .vue-flow__edge-path) { stroke: #7e91aa !important; stroke-width: 2px !important; }
.workflow-flow :deep(.vue-flow__edgeupdater) { cursor: crosshair; }
.workflow-flow :deep(.vue-flow__edgeupdater:hover) { fill: rgb(75 143 251 / 28%); stroke: #4b8ffb; }
@media (max-width: 900px) {
  .canvas-toolbar { padding: 0 12px; }
  .canvas-toolbar > div > span { display: none; }
  .canvas-toolbar__actions { gap: 4px; }
}
</style>
