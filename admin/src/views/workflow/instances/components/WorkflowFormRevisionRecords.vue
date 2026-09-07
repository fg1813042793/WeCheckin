<template>
  <div class="revision-records">
    <el-empty v-if="revisions.length === 0" :image-size="60" description="暂无修订记录" />
    <el-collapse v-else v-model="expandedRevisionIds" class="revision-collapse">
      <el-collapse-item v-for="revision in revisions" :key="revision.id" :name="revision.id">
        <template #title>
          <div class="revision-title">
            <span class="revision-title__main">{{ revision.sourceNodeName || revision.sourceNodeId || '表单修订' }}</span>
            <span class="revision-title__meta">{{ requesterDisplay(revision) }} · {{ formatTime(revision.createdAt) }}</span>
            <el-tag :type="revisionStatusMeta(revision.status).type" size="small">
              {{ revisionStatusMeta(revision.status).label }}
            </el-tag>
          </div>
        </template>

        <dl class="revision-meta">
          <div><dt>修订人</dt><dd>{{ requesterDisplay(revision) }}</dd></div>
          <div><dt>来源节点</dt><dd>{{ revision.sourceNodeName || revision.sourceNodeId || '-' }}</dd></div>
          <div><dt>生效方式</dt><dd>{{ revision.mode === 'direct' ? '直接生效' : '下游确认' }}</dd></div>
          <div><dt>表单版本</dt><dd>v{{ revision.baseFormRevision }} → {{ revision.appliedFormRevision ? `v${revision.appliedFormRevision}` : '待生效' }}</dd></div>
          <div class="revision-meta__wide"><dt>修改原因</dt><dd>{{ revision.reason || '暂无填写' }}</dd></div>
        </dl>

        <section class="revision-section">
          <h4>字段差异</h4>
          <el-table :data="fieldDiffRows(revision)" size="small" border empty-text="暂无字段差异">
            <el-table-column prop="label" label="字段" min-width="130" />
            <el-table-column label="修改前" min-width="210">
              <template #default="{ row }"><span class="readonly-value">{{ displayValue(row.before) }}</span></template>
            </el-table-column>
            <el-table-column label="修改后" min-width="210">
              <template #default="{ row }"><span class="readonly-value">{{ displayValue(row.after) }}</span></template>
            </el-table-column>
          </el-table>
        </section>

        <section class="revision-section">
          <h4>确认记录</h4>
          <el-empty v-if="revision.tasks.length === 0" :image-size="44" description="直接生效，无需确认" />
          <el-timeline v-else class="revision-timeline">
            <el-timeline-item
              v-for="task in revision.tasks"
              :key="task.id"
              :timestamp="formatTime(task.handledAt)"
              :type="revisionTaskStatusMeta(task.status).type"
            >
              <div class="revision-task-heading">
                <strong>{{ task.nodeName || task.nodeId }}</strong>
                <el-tag :type="revisionTaskStatusMeta(task.status).type" size="small" effect="plain">
                  {{ revisionTaskStatusMeta(task.status).label }}
                </el-tag>
              </div>
              <p>处理人：{{ taskUserDisplay(task) }}</p>
              <p v-if="task.comment">处理意见：{{ task.comment }}</p>
            </el-timeline-item>
          </el-timeline>
        </section>

        <section class="revision-section">
          <h4>通知投递</h4>
          <el-table :data="revisionNotifications(revision.id)" size="small" border empty-text="暂无通知投递">
            <el-table-column label="接收人" min-width="110">
              <template #default="{ row }">{{ userDisplay(row.recipientUserName, row.recipientUserId) }}</template>
            </el-table-column>
            <el-table-column label="事件" min-width="130">
              <template #default="{ row }">{{ notificationKindLabel(row.kind) }}</template>
            </el-table-column>
            <el-table-column label="渠道" width="100">
              <template #default="{ row }">{{ row.channel === 'in_app' ? '站内通知' : '钉钉通知' }}</template>
            </el-table-column>
            <el-table-column label="状态" width="95">
              <template #default="{ row }">{{ notificationStatusLabel(row.status) }}</template>
            </el-table-column>
            <el-table-column label="发送时间" width="170">
              <template #default="{ row }">{{ formatTime(row.sentAt) }}</template>
            </el-table-column>
          </el-table>
        </section>
      </el-collapse-item>
    </el-collapse>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type {
  WorkflowFormField,
  WorkflowFormRevisionDetail,
  WorkflowFormRevisionStatus,
  WorkflowFormRevisionTaskStatus,
  WorkflowFormRevisionTaskSummary,
  WorkflowNotificationKind,
  WorkflowNotificationRecord,
  WorkflowNotificationStatus,
} from '../../types'
import { workflowDataFields } from '../../formLayout'

const props = defineProps<{
  revisions: WorkflowFormRevisionDetail[]
  fields: WorkflowFormField[]
  userNames: Record<string, string>
  notifications: WorkflowNotificationRecord[]
}>()

const expandedRevisionIds = ref<string[]>([])
const statusTypes = {
  pending: 'warning', applied: 'success', rejected: 'danger', cancelled: 'info', conflict: 'danger',
} as const
const statusLabels = {
  pending: '待确认', applied: '已生效', rejected: '已驳回', cancelled: '已取消', conflict: '版本冲突',
} as const
const taskStatusTypes = {
  waiting: 'info', pending: 'warning', approved: 'success', rejected: 'danger', cancelled: 'info',
} as const
const taskStatusLabels = {
  waiting: '未开始', pending: '待确认', approved: '已通过', rejected: '已驳回', cancelled: '已取消',
} as const
const notificationLabels: Partial<Record<WorkflowNotificationKind, string>> = {
  instance_form_revision_requested: '修订待确认',
  instance_form_revision_approved: '修订已生效',
  instance_form_revision_rejected: '修订已驳回',
  instance_form_revision_cancelled: '修订已取消',
  instance_form_revision_conflict: '修订版本冲突',
}

function revisionStatusMeta(status: WorkflowFormRevisionStatus) {
  return { label: statusLabels[status] || status, type: statusTypes[status] || 'info' }
}

function revisionTaskStatusMeta(status: WorkflowFormRevisionTaskStatus) {
  return { label: taskStatusLabels[status] || status, type: taskStatusTypes[status] || 'info' }
}

function fieldDiffRows(revision: WorkflowFormRevisionDetail) {
  const labels = new Map(workflowDataFields(props.fields).map(field => [field.key, field.label || field.key]))
  return Object.keys(revision.patch || {}).sort().map(key => ({
    key,
    label: labels.get(key) || key,
    before: revision.beforeFormData?.[key],
    after: revision.proposedFormData?.[key],
  }))
}

function revisionNotifications(revisionId: string) {
  return props.notifications.filter(item => item.payload?.sourceType === 'workflow_form_revision' && item.payload?.sourceId === revisionId)
}

function requesterDisplay(revision: WorkflowFormRevisionDetail) {
  return userDisplay(revision.requesterName || props.userNames[revision.requesterId], revision.requesterId)
}

function taskUserDisplay(task: WorkflowFormRevisionTaskSummary) {
  const id = task.handledBy || task.assigneeId
  return userDisplay(props.userNames[id] || task.assigneeName, id)
}

function userDisplay(name?: string, id?: string) {
  return name?.trim() || id?.trim() || '-'
}

function displayValue(value: unknown) {
  if (value === undefined || value === null || value === '') return '暂无填写'
  if (typeof value === 'boolean') return value ? '是' : '否'
  if (typeof value === 'string' || typeof value === 'number') return String(value)
  try {
    return JSON.stringify(value)
  } catch {
    return String(value)
  }
}

function formatTime(timestamp: number) {
  if (!timestamp) return '-'
  return new Date(timestamp).toLocaleString('zh-CN', { hour12: false })
}

function notificationKindLabel(kind: WorkflowNotificationKind) {
  return notificationLabels[kind] || kind
}

function notificationStatusLabel(status: WorkflowNotificationStatus) {
  return { pending: '待投递', sending: '投递中', sent: '已发送', failed: '待重试', dead: '已停止' }[status] || status
}
</script>

<style scoped>
.revision-collapse { border-top: 0; }
.revision-title { display: flex; width: 100%; min-width: 0; align-items: center; gap: 10px; padding-right: 12px; }
.revision-title__main { flex: 0 1 auto; overflow: hidden; color: #303133; font-weight: 600; text-overflow: ellipsis; white-space: nowrap; }
.revision-title__meta { flex: 1 1 auto; min-width: 0; overflow: hidden; color: #909399; font-size: 12px; text-overflow: ellipsis; white-space: nowrap; }
.revision-meta { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); margin: 0 0 18px; border-top: 1px solid #ebeef5; border-left: 1px solid #ebeef5; }
.revision-meta > div { display: grid; grid-template-columns: 92px minmax(0, 1fr); border-right: 1px solid #ebeef5; border-bottom: 1px solid #ebeef5; }
.revision-meta dt,
.revision-meta dd { margin: 0; padding: 9px 12px; line-height: 1.5; }
.revision-meta dt { background: #f5f7fa; color: #606266; font-weight: 600; }
.revision-meta dd { min-width: 0; overflow-wrap: anywhere; color: #303133; }
.revision-meta__wide { grid-column: 1 / -1; }
.revision-section + .revision-section { margin-top: 20px; }
.revision-section h4 { margin: 0 0 10px; color: #303133; font-size: 14px; }
.readonly-value { display: block; min-height: 22px; color: #303133; line-height: 1.6; overflow-wrap: anywhere; white-space: pre-wrap; }
.revision-timeline { padding-top: 6px; }
.revision-timeline p { margin: 4px 0 0; color: #606266; }
.revision-task-heading { display: flex; align-items: center; justify-content: space-between; gap: 12px; }
@media (max-width: 767px) {
  .revision-meta { grid-template-columns: 1fr; }
  .revision-meta__wide { grid-column: auto; }
}
</style>
