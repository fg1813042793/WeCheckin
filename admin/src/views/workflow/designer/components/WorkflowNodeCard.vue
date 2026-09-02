<template>
  <button
    v-if="node.type === 'start'"
    type="button"
    class="start-node"
    :class="{ selected }"
    :aria-label="node.name || '流程开始'"
    @click.stop="$emit('select', node.id)"
  >
    <span class="start-node__event" aria-hidden="true">
      <el-icon><VideoPlay /></el-icon>
    </span>
    <b>{{ node.name || '流程开始' }}</b>
  </button>
  <button
    v-else-if="node.type === 'end'"
    type="button"
    class="end-node"
    :class="{ selected }"
    :aria-label="node.name || '流程结束'"
    @click.stop="$emit('select', node.id)"
  >
    <span class="end-node__event" aria-hidden="true">
      <el-icon><Check /></el-icon>
    </span>
    <b>{{ node.name || '流程结束' }}</b>
  </button>
  <button
    v-else
    type="button"
    class="workflow-node-card"
    :class="[`workflow-node-card--${node.type}`, { selected }]"
    @click.stop="$emit('select', node.id)"
  >
    <span class="workflow-node-card__header">
      <span>{{ nodeTypeLabel }}</span>
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
import { ArrowRight, Check, VideoPlay } from '@element-plus/icons-vue'
import type { WorkflowNode } from '../../types'

const props = defineProps<{ node: WorkflowNode; selected?: boolean }>()
defineEmits<{ select: [nodeId: string] }>()

const nodeTypeLabel = computed(() => {
  if (props.node.type === 'handle') return '办理/填写'
  if (props.node.type === 'cc') return '抄送'
  if (props.node.type === 'notify') return '通知'
  if (props.node.type === 'automation') return '自动动作'
  if (props.node.type === 'timer') return '定时/等待'
  return '审批人'
})

const description = computed(() => {
  if (props.node.type === 'handle') return [assigneeDescription(props.node) || '请配置办理人', notificationChannelSummary(props.node)].filter(Boolean).join(' · ')
  if (props.node.type === 'cc') return [assigneeDescription(props.node) || '请配置抄送人', notificationChannelSummary(props.node)].filter(Boolean).join(' · ')
  if (props.node.type === 'notify') return [assigneeDescription(props.node) || '请配置接收人', notificationChannelSummary(props.node)].filter(Boolean).join(' · ')
  if (props.node.type === 'automation') return `写入 ${Object.keys(props.node.automation?.variables || {}).length} 个变量`
  if (props.node.type === 'timer') return formatDelay(props.node.timer?.delaySeconds || 0)
  const mode = ({ single: '单人审批', sequential: '依次审批', parallel: '并行审批', countersign: '会签审批' } as Record<string, string>)[props.node.approvalMode || '']
  return [assigneeDescription(props.node), mode, notificationChannelSummary(props.node)].filter(Boolean).join(' · ') || '请配置审批规则'
})

function notificationChannelSummary(node: WorkflowNode) {
  if (!node.notification?.enabled) return '未开启通知'
  const labels = node.notification.channels.map(channel => channel === 'in_app' ? '站内' : channel === 'dingtalk_oa' ? '钉钉 OA' : channel)
  return labels.length ? `${labels.join('+')}通知` : '请配置通知渠道'
}

function assigneeDescription(node: WorkflowNode) {
  return ({
    initiator: '发起人',
    user: '指定用户',
    role: '指定角色',
    department_leader: '部门负责人',
    manager: '直属上级',
    variable: '流程变量',
    org_identity: '组织审批身份'
  } as Record<string, string>)[node.assignee?.type || ''] || ''
}

function formatDelay(seconds: number) {
  if (seconds >= 86400 && seconds % 86400 === 0) return `等待 ${seconds / 86400} 天`
  if (seconds >= 3600 && seconds % 3600 === 0) return `等待 ${seconds / 3600} 小时`
  if (seconds >= 60 && seconds % 60 === 0) return `等待 ${seconds / 60} 分钟`
  return seconds > 0 ? `等待 ${seconds} 秒` : '请配置等待时长'
}
</script>

<style scoped>
.workflow-node-card { display: block; width: 238px; padding: 0; overflow: hidden; border: 1px solid #d9dee7; border-radius: 5px; color: #344054; background: #fff; text-align: left; box-shadow: 0 3px 9px rgb(31 41 55 / 8%); cursor: pointer; transition: border-color .18s, box-shadow .18s, transform .18s; }
.workflow-node-card:hover { border-color: #9dbff7; box-shadow: 0 6px 15px rgb(31 41 55 / 11%); transform: translateY(-1px); }
.workflow-node-card.selected { border-color: #4b8ffb; box-shadow: 0 0 0 2px rgb(75 143 251 / 13%), 0 6px 15px rgb(31 41 55 / 10%); }
.workflow-node-card__header { display: flex; height: 30px; align-items: center; justify-content: space-between; padding: 0 11px; color: #fff; background: #60749f; font-size: 12px; }
.workflow-node-card--approval .workflow-node-card__header { background: #ef9b48; }
.workflow-node-card--handle .workflow-node-card__header { background: #3b82c4; }
.workflow-node-card--cc .workflow-node-card__header { background: #7c5bb5; }
.workflow-node-card--notify .workflow-node-card__header { background: #0f766e; }
.workflow-node-card--automation .workflow-node-card__header { background: #566270; }
.workflow-node-card--timer .workflow-node-card__header { background: #b47a2d; }
.workflow-node-card__body { display: block; min-height: 58px; padding: 10px 12px; }
.workflow-node-card__body b, .workflow-node-card__body small { display: block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.workflow-node-card__body b { font-size: 13px; font-weight: 600; }
.workflow-node-card__body small { margin-top: 7px; color: #8a94a3; font-size: 11px; }
.start-node, .end-node { display: flex; width: 120px; flex-direction: column; align-items: center; gap: 8px; padding: 0; border: 0; color: #667085; background: transparent; font-family: inherit; cursor: pointer; }
.start-node__event { display: grid; width: 44px; height: 44px; box-sizing: border-box; place-items: center; border: 2px solid #3282ce; border-radius: 50%; color: #2877bd; background: #f4f9ff; box-shadow: 0 3px 10px rgb(31 41 55 / 9%); transition: border-color .18s, color .18s, background .18s, box-shadow .18s, transform .18s; }
.start-node__event .el-icon { font-size: 17px; }
.end-node__event { position: relative; display: grid; width: 44px; height: 44px; box-sizing: border-box; place-items: center; border: 3px solid #697586; border-radius: 50%; color: #697586; background: #fff; box-shadow: 0 3px 10px rgb(31 41 55 / 10%); transition: border-color .18s, color .18s, box-shadow .18s, transform .18s; }
.end-node__event::before { position: absolute; inset: 4px; border: 1px solid #a4adba; border-radius: 50%; content: ''; transition: border-color .18s; }
.end-node__event .el-icon { z-index: 1; font-size: 16px; stroke-width: 2.5; }
.start-node b, .end-node b { max-width: 120px; overflow: hidden; font-size: 13px; font-weight: 600; line-height: 18px; text-overflow: ellipsis; white-space: nowrap; }
.start-node:hover .start-node__event { border-color: #1668ad; color: #1668ad; background: #edf6ff; box-shadow: 0 5px 13px rgb(50 130 206 / 18%); transform: translateY(-1px); }
.start-node.selected .start-node__event { border-color: #1668ad; color: #1668ad; box-shadow: 0 0 0 3px rgb(50 130 206 / 14%), 0 5px 13px rgb(31 41 55 / 10%); }
.end-node:hover .end-node__event { border-color: #4b8ffb; color: #3478dc; box-shadow: 0 5px 13px rgb(75 143 251 / 18%); transform: translateY(-1px); }
.end-node:hover .end-node__event::before, .end-node.selected .end-node__event::before { border-color: #8bb7f6; }
.end-node.selected .end-node__event { border-color: #4b8ffb; color: #3478dc; box-shadow: 0 0 0 3px rgb(75 143 251 / 13%), 0 5px 13px rgb(31 41 55 / 10%); }
.start-node:focus-visible, .end-node:focus-visible { outline: none; }
.start-node:focus-visible .start-node__event { border-color: #1668ad; box-shadow: 0 0 0 3px rgb(50 130 206 / 18%), 0 5px 13px rgb(31 41 55 / 10%); }
.end-node:focus-visible .end-node__event { border-color: #4b8ffb; box-shadow: 0 0 0 3px rgb(75 143 251 / 18%), 0 5px 13px rgb(31 41 55 / 10%); }
</style>
