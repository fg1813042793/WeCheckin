<script setup lang="ts">
import type {
  WorkflowHistorySummary,
  WorkflowInstanceSummary,
  WorkflowNodeProgressStatus,
  WorkflowNodeProgressSummary,
  WorkflowTaskSummary,
} from '@/types/workflow'
import { computed } from 'vue'
import { workflowNodeProgressStatusMeta, workflowTaskStatusMeta } from '../workflow-status'
import WorkflowImagePicker from './WorkflowImagePicker.vue'

const props = defineProps<{
  nodeProgress: WorkflowNodeProgressSummary[]
  tasks: WorkflowTaskSummary[]
  history: WorkflowHistorySummary[]
  instance: WorkflowInstanceSummary
}>()

const tasksByNode = computed(() => {
  const grouped: Record<string, WorkflowTaskSummary[]> = {}
  for (const task of props.tasks) {
    grouped[task.nodeId] ||= []
    grouped[task.nodeId].push(task)
  }
  for (const tasks of Object.values(grouped)) {
    tasks.sort((left, right) => left.sequence - right.sequence || left.id.localeCompare(right.id))
  }
  return grouped
})

const commentsByNode = computed(() => {
  const grouped: Record<string, WorkflowHistorySummary[]> = {}
  for (const event of props.history) {
    if (event.eventType !== 'instance_commented' || !event.nodeId)
      continue
    grouped[event.nodeId] ||= []
    grouped[event.nodeId].push(event)
  }
  for (const comments of Object.values(grouped)) {
    comments.sort((left, right) => right.eventTime - left.eventTime || left.id.localeCompare(right.id))
  }
  return grouped
})

const legacyProgress = computed<WorkflowNodeProgressSummary[]>(() => {
  return Object.entries(tasksByNode.value).map(([nodeId, tasks]) => ({
    nodeId,
    nodeName: tasks[0]?.nodeName || '审批节点',
    nodeType: tasks.some(task => task.action === 'submit') ? 'handle' : 'approval',
    status: legacyNodeStatus(tasks),
  }))
})

const visibleProgress = computed(() => {
  return props.nodeProgress.length > 0 ? props.nodeProgress : legacyProgress.value
})

function legacyNodeStatus(tasks: WorkflowTaskSummary[]): WorkflowNodeProgressStatus {
  if (tasks.some(task => task.status === 'pending' || task.status === 'waiting'))
    return 'processing'
  if (tasks.some(task => task.status === 'rejected'))
    return 'terminated'
  if (tasks.some(task => ['approved', 'completed', 'submitted'].includes(task.status)))
    return 'completed'
  if (tasks.length > 0 && tasks.every(task => task.status === 'cancelled'))
    return 'terminated'
  return 'not_started'
}

function nodeProgressStatusClass(status: WorkflowNodeProgressStatus) {
  const type = workflowNodeProgressStatusMeta(status, props.instance.status).type
  return `workflow-node-progress__item--${type}`
}

function nodeTypeLabel(node: WorkflowNodeProgressSummary) {
  if (node.nodeType === 'exclusive')
    return node.gatewayMode === 'join' ? '条件汇聚' : '条件分支'
  if (node.nodeType === 'parallel')
    return node.gatewayMode === 'join' ? '并行汇聚' : '并行分支'
  const labels: Record<string, string> = {
    start: '发起节点',
    approval: '审批节点',
    handle: '办理节点',
    cc: '抄送节点',
    notify: '通知节点',
    automation: '自动动作',
    timer: '定时等待',
    end: '结束节点',
  }
  return labels[node.nodeType] || node.nodeType || '流程节点'
}

function actionLabel(action = '') {
  return {
    approve: '通过',
    reject: '驳回',
    return: '退回',
    submit: '提交',
  }[action] || action
}

function taskHandlerName(task: WorkflowTaskSummary) {
  const name = task.handledBy ? task.handledByName : task.assigneeName
  return String(name || '').trim() || '未知用户'
}

function commentActorName(comment: WorkflowHistorySummary) {
  return String(comment.actorName || '').trim() || '未知用户'
}

function formatTime(timestamp?: number) {
  if (!timestamp)
    return '-'
  return new Date(timestamp).toLocaleString('zh-CN', { hour12: false })
}
</script>

<template>
  <view class="workflow-node-progress">
    <view v-if="visibleProgress.length === 0" class="workflow-node-progress__empty">
      暂无节点进度
    </view>
    <view
      v-for="node in visibleProgress"
      :key="node.nodeId"
      class="workflow-node-progress__item"
      :class="nodeProgressStatusClass(node.status)"
    >
      <view class="workflow-node-progress__rail">
        <view class="workflow-node-progress__dot" />
      </view>
      <view class="workflow-node-progress__content">
        <view class="workflow-node-progress__heading">
          <view class="workflow-node-progress__title-wrap">
            <text class="workflow-node-progress__title">
              {{ node.nodeName || nodeTypeLabel(node) }}
            </text>
            <text class="workflow-node-progress__type">
              {{ nodeTypeLabel(node) }}
            </text>
          </view>
          <u-tag
            :text="workflowNodeProgressStatusMeta(node.status, instance.status).label"
            :type="workflowNodeProgressStatusMeta(node.status, instance.status).type"
            custom-class="workflow-node-progress__status-tag"
            size="mini"
          />
        </view>

        <text v-if="node.nodeType === 'start'" class="workflow-node-progress__summary">
          发起人：{{ instance.starterName || '未知用户' }} · {{ formatTime(instance.startTime) }}
        </text>
        <text v-else-if="node.nodeType === 'end' && node.status === 'completed'" class="workflow-node-progress__summary">
          完成于 {{ formatTime(instance.endTime) }}
        </text>

        <view v-if="tasksByNode[node.nodeId]?.length" class="workflow-node-progress__tasks">
          <view v-for="task in tasksByNode[node.nodeId]" :key="task.id" class="workflow-node-progress__task">
            <view class="workflow-node-progress__task-heading">
              <text class="workflow-node-progress__task-user">
                处理人：{{ taskHandlerName(task) }}
              </text>
              <u-tag
                :text="workflowTaskStatusMeta(task.status).label"
                :type="workflowTaskStatusMeta(task.status).type"
                custom-class="workflow-node-progress__status-tag"
                size="mini"
              />
            </view>
            <text class="workflow-node-progress__task-meta">
              {{ task.handledAt ? formatTime(task.handledAt) : '等待处理' }}
            </text>
            <text v-if="task.action || task.comment" class="workflow-node-progress__task-comment">
              {{ actionLabel(task.action) }}{{ task.comment ? `：${task.comment}` : '' }}
            </text>
            <WorkflowImagePicker
              v-if="task.images?.length"
              class="workflow-node-progress__task-images"
              :model-value="task.images"
              readonly
            />
          </view>
        </view>

        <view v-if="commentsByNode[node.nodeId]?.length" class="workflow-node-progress__comments">
          <view
            v-for="comment in commentsByNode[node.nodeId]"
            :key="comment.id"
            class="workflow-node-progress__comment"
          >
            <view class="workflow-node-progress__comment-heading">
              <view class="workflow-node-progress__comment-author">
                <u-icon name="chat" size="14px" color="#0f766e" />
                <text>评论人：{{ commentActorName(comment) }}</text>
              </view>
              <text class="workflow-node-progress__comment-time">
                {{ formatTime(comment.eventTime) }}
              </text>
            </view>
            <text v-if="comment.message" class="workflow-node-progress__comment-message">
              {{ comment.message }}
            </text>
            <WorkflowImagePicker
              v-if="comment.images?.length"
              class="workflow-node-progress__comment-images"
              :model-value="comment.images"
              readonly
            />
          </view>
        </view>
      </view>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.workflow-node-progress {
  min-width: 0;
}

.workflow-node-progress__empty {
  padding: 20px 0;
  color: #86909c;
  font-size: 13px;
  text-align: center;
}

.workflow-node-progress__item {
  position: relative;
  min-width: 0;
  padding-bottom: 22px;
  display: grid;
  grid-template-columns: 18px minmax(0, 1fr);
  column-gap: 10px;
}

.workflow-node-progress__item:last-child {
  padding-bottom: 0;
}

.workflow-node-progress__rail {
  position: relative;
  width: 18px;
}

.workflow-node-progress__rail::after {
  position: absolute;
  top: 17px;
  bottom: -22px;
  left: 8px;
  width: 2px;
  background: #e5e6eb;
  content: '';
}

.workflow-node-progress__item:last-child .workflow-node-progress__rail::after {
  display: none;
}

.workflow-node-progress__dot {
  position: absolute;
  top: 4px;
  left: 2px;
  width: 12px;
  height: 12px;
  border: 2px solid #fff;
  border-radius: 50%;
  background: #c9cdd4;
  box-shadow: 0 0 0 1px #c9cdd4;
  box-sizing: border-box;
}

.workflow-node-progress__item--success .workflow-node-progress__dot {
  background: #00a870;
  box-shadow: 0 0 0 1px #7bd8bb;
}

.workflow-node-progress__item--warning .workflow-node-progress__dot {
  background: #f59e0b;
  box-shadow: 0 0 0 2px #fde68a;
}

.workflow-node-progress__item--info .workflow-node-progress__dot {
  background: #c9cdd4;
  box-shadow: 0 0 0 1px #c9cdd4;
}

.workflow-node-progress__item--error .workflow-node-progress__dot {
  background: #f53f3f;
  box-shadow: 0 0 0 1px #ffb8b8;
}

.workflow-node-progress__item--success .workflow-node-progress__rail::after {
  background: #a7e3cf;
}

.workflow-node-progress__content {
  min-width: 0;
}

.workflow-node-progress__heading,
.workflow-node-progress__task-heading {
  min-width: 0;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 10px;
}

.workflow-node-progress__title-wrap {
  min-width: 0;
  display: grid;
  gap: 3px;
}

.workflow-node-progress__title {
  color: #1f2329;
  font-size: 14px;
  font-weight: 700;
  line-height: 1.45;
  overflow-wrap: anywhere;
}

.workflow-node-progress__type,
.workflow-node-progress__summary,
.workflow-node-progress__task-meta,
.workflow-node-progress__task-comment {
  display: block;
  color: #86909c;
  font-size: 12px;
  line-height: 1.55;
}

.workflow-node-progress__summary {
  margin-top: 7px;
  color: #4e5969;
}

.workflow-node-progress__tasks {
  margin-top: 9px;
  border-top: 1px solid #edf0f5;
}

.workflow-node-progress__task {
  min-width: 0;
  padding: 10px 0 0;
}

.workflow-node-progress__task + .workflow-node-progress__task {
  margin-top: 10px;
  border-top: 1px dashed #e5e6eb;
}

.workflow-node-progress__task-user {
  min-width: 0;
  color: #4e5969;
  font-size: 12px;
  line-height: 1.5;
  overflow-wrap: anywhere;
}

.workflow-node-progress__task-meta,
.workflow-node-progress__task-comment {
  margin-top: 4px;
}

.workflow-node-progress__task-images {
  margin-top: 8px;
}

.workflow-node-progress__comments {
  margin-top: 10px;
  padding-top: 10px;
  border-top: 1px solid #edf0f5;
}

.workflow-node-progress__comment {
  min-width: 0;
  padding-left: 10px;
  border-left: 2px solid #a7e3cf;
}

.workflow-node-progress__comment + .workflow-node-progress__comment {
  margin-top: 12px;
}

.workflow-node-progress__comment-heading {
  min-width: 0;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 10px;
}

.workflow-node-progress__comment-author {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 5px;
  color: #4e5969;
  font-size: 12px;
  line-height: 1.5;
  overflow-wrap: anywhere;
}

.workflow-node-progress__comment-time {
  flex: 0 0 auto;
  color: #86909c;
  font-size: 12px;
  line-height: 1.5;
  white-space: nowrap;
}

.workflow-node-progress__comment-message {
  display: block;
  margin-top: 5px;
  color: #1f2329;
  font-size: 13px;
  line-height: 1.6;
  overflow-wrap: anywhere;
  white-space: pre-wrap;
}

.workflow-node-progress__comment-images {
  margin-top: 8px;
}

:deep(.workflow-node-progress__status-tag) {
  width: auto !important;
  height: 24px;
  min-height: 24px;
  padding: 0 8px !important;
  border-radius: 4px !important;
  font-size: 12px !important;
  line-height: 22px !important;
  white-space: nowrap;
  box-sizing: border-box;
}

.workflow-node-progress__task-comment {
  color: #4e5969;
  overflow-wrap: anywhere;
}
</style>
