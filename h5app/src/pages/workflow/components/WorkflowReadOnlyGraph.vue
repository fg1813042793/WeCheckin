<script setup lang="ts">
import type { WorkflowPublishedEdge, WorkflowPublishedNode } from '@/types/workflow'
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { alignWorkflowTerminalNodes } from '../workflow-graph-layout'

interface GraphNodeLayout {
  node: WorkflowPublishedNode
  x: number
  y: number
  width: number
  height: number
}

interface GraphPoint {
  x: number
  y: number
  side: 'top' | 'right' | 'bottom' | 'left'
}

const props = defineProps<{
  nodes: WorkflowPublishedNode[]
  edges: WorkflowPublishedEdge[]
}>()

const markerId = `workflow-graph-arrow-${Math.random().toString(36).slice(2, 9)}`
const canvasPadding = 64
const defaultNodeGap = 148
const viewportRef = ref<unknown>(null)
const viewportWidth = ref(0)
let resizeObserver: ResizeObserver | null = null

const positionedNodes = computed<GraphNodeLayout[]>(() => {
  const source = props.nodes || []
  const nodes = source.map((node, index) => {
    const size = nodeSize(node)
    return {
      node,
      width: size.width,
      height: size.height,
      x: Number.isFinite(Number(node.position?.x)) ? Number(node.position?.x) : 280,
      y: Number.isFinite(Number(node.position?.y)) ? Number(node.position?.y) : index * defaultNodeGap,
    }
  })
  return alignWorkflowTerminalNodes(nodes, props.edges || [])
})
const graphBounds = computed(() => {
  if (positionedNodes.value.length === 0)
    return { minX: 0, minY: 0, width: 0, height: 0 }
  const minX = Math.min(...positionedNodes.value.map(item => item.x))
  const minY = Math.min(...positionedNodes.value.map(item => item.y))
  const maxX = Math.max(...positionedNodes.value.map(item => item.x + item.width))
  const maxY = Math.max(...positionedNodes.value.map(item => item.y + item.height))
  return { minX, minY, width: maxX - minX, height: maxY - minY }
})
const canvasSize = computed(() => {
  const contentWidth = graphBounds.value.width
  const width = Math.max(viewportWidth.value, contentWidth + canvasPadding * 2)
  const height = Math.max(480, graphBounds.value.height + canvasPadding * 2)
  return { width: Math.ceil(width), height: Math.ceil(height) }
})
const graphNodes = computed<GraphNodeLayout[]>(() => {
  const bounds = graphBounds.value
  const horizontalOffset = (canvasSize.value.width - bounds.width) / 2 - bounds.minX
  return positionedNodes.value.map(item => ({
    ...item,
    x: item.x + horizontalOffset,
    y: item.y - bounds.minY + canvasPadding,
  }))
})

const nodeMap = computed(() => new Map(graphNodes.value.map(item => [item.node.id, item])))
const graphEdges = computed(() => {
  return (props.edges || []).map((edge) => {
    const source = nodeMap.value.get(edge.source)
    const target = nodeMap.value.get(edge.target)
    if (!source || !target)
      return null
    const start = nodeAnchor(source, edge.sourceHandle, true)
    const end = nodeAnchor(target, edge.targetHandle, false)
    return {
      edge,
      path: edgePath(start, end, source.node.type === 'start' || target.node.type === 'end'),
      labelX: (start.x + end.x) / 2,
      labelY: (start.y + end.y) / 2 - 8,
    }
  }).filter((item): item is NonNullable<typeof item> => Boolean(item))
})
const canvasStyle = computed(() => ({
  width: `${canvasSize.value.width}px`,
  height: `${canvasSize.value.height}px`,
}))

function nodeSize(node: WorkflowPublishedNode) {
  if (['start', 'end'].includes(node.type))
    return { width: 132, height: 72 }
  if (['exclusive', 'parallel'].includes(node.type))
    return { width: 148, height: 88 }
  return { width: 236, height: 96 }
}

function nodeStyle(item: GraphNodeLayout) {
  return {
    left: `${item.x}px`,
    top: `${item.y}px`,
    width: `${item.width}px`,
    height: `${item.height}px`,
  }
}

function normalizedHandle(value: string | undefined, fallback: GraphPoint['side']) {
  const handle = String(value || '').toLowerCase()
  if (handle.includes('top'))
    return 'top'
  if (handle.includes('right'))
    return 'right'
  if (handle.includes('left'))
    return 'left'
  if (handle.includes('bottom'))
    return 'bottom'
  return fallback
}

function nodeAnchor(item: GraphNodeLayout, handle: string | undefined, source: boolean): GraphPoint {
  const side = normalizedHandle(handle, source ? 'bottom' : 'top')
  if (side === 'top')
    return { x: item.x + item.width / 2, y: item.y, side }
  if (side === 'right')
    return { x: item.x + item.width, y: item.y + item.height / 2, side }
  if (side === 'left')
    return { x: item.x, y: item.y + item.height / 2, side }
  return { x: item.x + item.width / 2, y: item.y + item.height, side }
}

function controlPoint(point: GraphPoint, distance: number) {
  if (point.side === 'top')
    return { x: point.x, y: point.y - distance }
  if (point.side === 'right')
    return { x: point.x + distance, y: point.y }
  if (point.side === 'left')
    return { x: point.x - distance, y: point.y }
  return { x: point.x, y: point.y + distance }
}

function edgePath(start: GraphPoint, end: GraphPoint, straight = false) {
  if (straight)
    return `M ${start.x} ${start.y} L ${end.x} ${end.y}`
  const distance = Math.max(54, Math.min(150, Math.hypot(end.x - start.x, end.y - start.y) * 0.42))
  const first = controlPoint(start, distance)
  const second = controlPoint(end, distance)
  return `M ${start.x} ${start.y} C ${first.x} ${first.y}, ${second.x} ${second.y}, ${end.x} ${end.y}`
}

function nodeTypeLabel(node: WorkflowPublishedNode) {
  const labels: Record<string, string> = {
    start: '发起节点',
    end: '结束节点',
    approval: '审批节点',
    handle: '办理/填写',
    cc: '抄送/通知',
    notify: '通知节点',
    automation: '自动动作',
    timer: '定时/等待',
    exclusive: node.gatewayMode === 'join' ? '条件汇聚' : '条件分支',
    parallel: node.gatewayMode === 'join' ? '并行汇聚' : '并行分支',
  }
  return labels[node.type] || '流程节点'
}

function approvalModeLabel(mode = '') {
  return ({
    single: '单人审批',
    sequential: '依次审批',
    parallel: '并行审批',
    countersign: '会签审批',
  } as Record<string, string>)[mode] || ''
}

function assigneePrefix(node: WorkflowPublishedNode) {
  if (node.type === 'approval')
    return '审批人：'
  if (node.type === 'cc' || node.type === 'notify')
    return '接收人：'
  return '处理人：'
}

function viewportElement() {
  if (typeof HTMLElement === 'undefined')
    return null
  const target = viewportRef.value as HTMLElement | { $el?: unknown } | null
  if (target instanceof HTMLElement)
    return target
  return target?.$el instanceof HTMLElement ? target.$el : null
}

function updateViewportWidth() {
  const element = viewportElement()
  if (element)
    viewportWidth.value = Math.max(0, element.getBoundingClientRect().width)
}

onMounted(async () => {
  await nextTick()
  updateViewportWidth()
  const element = viewportElement()
  if (element && typeof ResizeObserver !== 'undefined') {
    resizeObserver = new ResizeObserver(updateViewportWidth)
    resizeObserver.observe(element)
  }
})

onBeforeUnmount(() => {
  resizeObserver?.disconnect()
  resizeObserver = null
})
</script>

<template>
  <view v-if="graphNodes.length === 0" class="workflow-graph__empty">
    <u-icon name="share" size="58" color="#c8c9cc" />
    <text>暂无流程图</text>
  </view>
  <scroll-view v-else ref="viewportRef" class="workflow-graph__viewport" scroll-x scroll-y>
    <view class="workflow-graph__canvas" :style="canvasStyle">
      <svg
        class="workflow-graph__edges"
        :width="canvasSize.width"
        :height="canvasSize.height"
        :viewBox="`0 0 ${canvasSize.width} ${canvasSize.height}`"
      >
        <defs>
          <marker :id="markerId" marker-width="8" marker-height="8" ref-x="7" ref-y="4" orient="auto" marker-units="strokeWidth">
            <path d="M 0 0 L 8 4 L 0 8 z" fill="#aeb8c5" />
          </marker>
        </defs>
        <g v-for="item in graphEdges" :key="item.edge.id">
          <path
            :d="item.path"
            fill="none"
            stroke="#aeb8c5"
            stroke-width="1.5"
            :marker-end="`url(#${markerId})`"
          />
          <text
            v-if="item.edge.name"
            :x="item.labelX"
            :y="item.labelY"
            fill="#687386"
            font-size="12"
            text-anchor="middle"
          >
            {{ item.edge.name }}
          </text>
        </g>
      </svg>

      <view
        v-for="item in graphNodes"
        :key="item.node.id"
        class="workflow-graph__node"
        :class="[`workflow-graph__node--${item.node.type}`]"
        :style="nodeStyle(item)"
      >
        <view class="workflow-graph__node-type">
          {{ nodeTypeLabel(item.node) }}
        </view>
        <text class="workflow-graph__node-name">
          {{ item.node.name || nodeTypeLabel(item.node) }}
        </text>
        <text v-if="item.node.assigneeDisplay" class="workflow-graph__node-assignee">
          {{ assigneePrefix(item.node) }}{{ item.node.assigneeDisplay }}
        </text>
        <text v-else-if="item.node.type === 'approval'" class="workflow-graph__node-assignee workflow-graph__node-assignee--empty">
          审批人：未配置
        </text>
        <text v-if="item.node.type === 'approval' && approvalModeLabel(item.node.approvalMode)" class="workflow-graph__node-mode">
          {{ approvalModeLabel(item.node.approvalMode) }}
        </text>
      </view>
    </view>
  </scroll-view>
</template>

<style lang="scss" scoped>
.workflow-graph__viewport {
  width: 100%;
  height: 100%;
  min-height: 480px;
  border: 1px solid #e4e9f1;
  border-radius: 6px;
  background-color: #f8fafc;
  background-image:
    linear-gradient(#e8edf4 1px, transparent 1px),
    linear-gradient(90deg, #e8edf4 1px, transparent 1px);
  background-size: 24px 24px;
  box-sizing: border-box;
}

.workflow-graph__canvas {
  position: relative;
}

.workflow-graph__edges {
  position: absolute;
  inset: 0;
  z-index: 1;
  overflow: visible;
  pointer-events: none;
}

.workflow-graph__node {
  position: absolute;
  z-index: 2;
  padding: 31px 13px 9px;
  border: 1px solid #d7dee8;
  border-radius: 5px;
  overflow: hidden;
  background: #ffffff;
  box-shadow: 0 4px 12px rgba(31, 41, 55, 0.09);
  box-sizing: border-box;
}

.workflow-graph__node-type {
  position: absolute;
  top: 0;
  right: 0;
  left: 0;
  height: 27px;
  padding: 0 11px;
  display: flex;
  align-items: center;
  background: #60749f;
  color: #ffffff;
  font-size: 11px;
  font-weight: 600;
}

.workflow-graph__node--approval .workflow-graph__node-type { background: #e78d36; }
.workflow-graph__node--handle .workflow-graph__node-type { background: #3b82c4; }
.workflow-graph__node--cc .workflow-graph__node-type { background: #7c5bb5; }
.workflow-graph__node--notify .workflow-graph__node-type { background: #0f766e; }
.workflow-graph__node--automation .workflow-graph__node-type { background: #566270; }
.workflow-graph__node--timer .workflow-graph__node-type { background: #b47a2d; }
.workflow-graph__node--exclusive .workflow-graph__node-type,
.workflow-graph__node--parallel .workflow-graph__node-type { background: #15966d; }
.workflow-graph__node--start .workflow-graph__node-type { background: #3282ce; }
.workflow-graph__node--end .workflow-graph__node-type { background: #697586; }

.workflow-graph__node-name,
.workflow-graph__node-assignee,
.workflow-graph__node-mode {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.workflow-graph__node-name {
  color: #273142;
  font-size: 13px;
  font-weight: 700;
  line-height: 20px;
}

.workflow-graph__node-assignee,
.workflow-graph__node-mode {
  margin-top: 3px;
  color: #687386;
  font-size: 11px;
  line-height: 16px;
}

.workflow-graph__node-assignee--empty {
  color: #c05621;
}

.workflow-graph__node-mode {
  color: #9aa4b2;
}

.workflow-graph__node--start,
.workflow-graph__node--end,
.workflow-graph__node--exclusive,
.workflow-graph__node--parallel {
  padding-right: 10px;
  padding-left: 10px;
  text-align: center;
}

.workflow-graph__empty {
  min-height: 420px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  color: #86909c;
  font-size: 13px;
}

@media screen and (max-width: 768px) {
  .workflow-graph__viewport {
    height: 580px;
    min-height: 580px;
  }
}
</style>
