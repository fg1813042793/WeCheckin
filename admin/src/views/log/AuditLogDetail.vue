<script setup lang="ts">
import { computed } from 'vue'
import { AdminDrawer } from '@/components/admin-ui'
import { hasPerm } from '@/utils/permission'
import type { AuditLogRecord } from '@/types/auditLog'
import { actionLabel, actorLabel, affectedCountLabel, logTime, moduleLabel, outcomeMeta, relatedLogRoute, requestedTargetsSummary, sourceLabel } from './auditLogHelpers'
import { auditSnapshotView } from './auditSnapshot'

const props = defineProps<{ modelValue: boolean; record: AuditLogRecord | null; loading: boolean; error: string }>()
defineEmits<{ 'update:modelValue': [value: boolean]; retry: [] }>()
const relatedRoute = computed(() => relatedLogRoute(props.record?.relatedType, props.record?.relatedId, hasPerm))
const requestedTargets = computed(() => requestedTargetsSummary(props.record))
const snapshot = computed(() => auditSnapshotView(props.record))
</script>

<template>
  <AdminDrawer :model-value="modelValue" title="审计日志详情" size="lg" @update:model-value="$emit('update:modelValue', $event)">
    <div v-if="loading" v-loading="loading" class="audit-detail-loading" aria-label="正在加载日志详情" />
    <div v-else-if="error" role="alert">
      <el-alert :title="error" type="error" :closable="false" show-icon />
      <el-button class="audit-retry" @click="$emit('retry')">重试</el-button>
    </div>
    <div v-else-if="record" class="audit-detail">
      <el-alert v-if="record.outcome === 'accepted'" title="此事件仅表示已受理，异步任务尚未确认完成。请结合关联运行记录核实最终结果。" type="info" :closable="false" show-icon />
      <el-descriptions title="事件与结果" :column="1" border>
        <el-descriptions-item label="日志 / 事件 ID">{{ record.id }} / {{ record.eventId || '—' }}</el-descriptions-item>
        <el-descriptions-item label="时间">{{ logTime(record._createTime) }}</el-descriptions-item>
        <el-descriptions-item label="来源 / 模块 / 动作">{{ sourceLabel(record.source) }} / {{ moduleLabel(record.module) }} / {{ actionLabel(record) }}</el-descriptions-item>
        <el-descriptions-item label="结果"><el-tag :type="outcomeMeta(record.outcome).type">{{ outcomeMeta(record.outcome).label }}</el-tag></el-descriptions-item>
        <el-descriptions-item label="内容">{{ record.content || '—' }}</el-descriptions-item>
        <el-descriptions-item label="原因代码">{{ record.reasonCode || '未记录' }}</el-descriptions-item>
        <el-descriptions-item label="原因">{{ record.reason || (record.outcome === 'success' ? '无' : '未记录') }}</el-descriptions-item>
      </el-descriptions>
      <el-descriptions title="操作者与目标" :column="1" border>
        <el-descriptions-item label="实际操作者">{{ actorLabel(record) }}</el-descriptions-item>
        <el-descriptions-item v-if="record.category === 'login'" label="尝试账号">{{ record.attemptedAccount || '—' }}<div class="audit-muted">尝试账号是登录输入，不代表已验证身份。</div></el-descriptions-item>
        <el-descriptions-item label="目标类型 / ID">{{ record.targetType || '—' }} / {{ record.targetId ?? '—' }}</el-descriptions-item>
        <el-descriptions-item label="目标名称">{{ record.targetName || '—' }}</el-descriptions-item>
        <el-descriptions-item label="目标 ID 列表">
          <div class="audit-target-ids" tabindex="0" role="region" aria-label="目标 ID 列表">{{ requestedTargets.text }}</div>
          <el-alert v-if="requestedTargets.notice" :title="requestedTargets.notice" type="warning" :closable="false" show-icon />
        </el-descriptions-item>
        <el-descriptions-item label="请求数量">{{ requestedTargets.countLabel }}<div class="audit-muted">请求数量不代表实际成功或影响数量。</div></el-descriptions-item>
        <el-descriptions-item label="影响数量">{{ affectedCountLabel(record.affectedCount) }}</el-descriptions-item>
      </el-descriptions>
      <el-descriptions title="请求信息" :column="1" border>
        <el-descriptions-item label="请求 ID">{{ record.requestId || '—' }}</el-descriptions-item>
        <el-descriptions-item label="方法 / 路径">{{ record.method || '—' }} {{ record.path || '—' }}</el-descriptions-item>
        <el-descriptions-item label="IP">{{ record.LOG_ADD_IP || '—' }}</el-descriptions-item>
        <el-descriptions-item label="耗时">{{ record.durationMs == null ? '未记录' : `${record.durationMs} ms` }}</el-descriptions-item>
        <el-descriptions-item label="User Agent">{{ record.userAgent || '—' }}</el-descriptions-item>
        <el-descriptions-item label="浏览器 / 设备 / 平台">{{ record.browser || '未记录' }} / {{ record.deviceType || '未记录' }} / {{ record.platformType || '未记录' }}</el-descriptions-item>
        <el-descriptions-item label="关联类型 / ID">
          {{ record.relatedType || '—' }} / {{ record.relatedId ?? '—' }}
          <router-link v-if="relatedRoute" class="audit-related-link" :to="relatedRoute.path">{{ relatedRoute.label }}</router-link>
        </el-descriptions-item>
      </el-descriptions>
      <section aria-label="变更快照">
        <h3>变更快照</h3>
        <p class="audit-snapshot-note">请求前后快照，可能包含并发更新，不代表均由本次操作者造成。</p>
        <p v-if="snapshot.hasRedactions" class="audit-snapshot-note">敏感值已脱敏；[已隐藏]、[已变更] 是保护标记，不能据此还原旧值或判断开关状态。“未记录”或“未提供”表示日志未包含该值。</p>
        <div class="audit-snapshot-status" role="status">
          <strong>{{ snapshot.title }}</strong>
          <p>{{ snapshot.description }}</p>
          <p v-if="snapshot.reason">采集原因：{{ snapshot.reason }}</p>
          <p v-if="snapshot.legacyNotice">{{ snapshot.legacyNotice }}</p>
        </div>
        <el-table v-if="snapshot.rows.length" :data="snapshot.rows" border>
          <el-table-column label="字段" min-width="200">
            <template #default="{ row }">
              <el-tooltip :content="row.field" placement="top"><span tabindex="0">{{ row.label }}</span></el-tooltip>
            </template>
          </el-table-column>
          <el-table-column label="请求前" min-width="200"><template #default="{ row }"><pre>{{ row.before }}</pre></template></el-table-column>
          <el-table-column label="请求后" min-width="200"><template #default="{ row }"><pre>{{ row.after }}</pre></template></el-table-column>
        </el-table>
        <el-empty v-else description="无可展示的字段差异，请参阅上述采集说明" :image-size="60" />
      </section>
    </div>
  </AdminDrawer>
</template>

<style scoped>
.audit-detail { display: grid; gap: var(--admin-ui-space-5); overflow-wrap: anywhere; }
.audit-detail-loading { min-height: 240px; }
.audit-detail :deep(.el-descriptions__table) { table-layout: fixed; }
.audit-detail :deep(.el-descriptions__label) { width: 130px; }
.audit-detail pre { white-space: pre-wrap; overflow-wrap: anywhere; font: inherit; margin: 0; max-height: 240px; overflow-y: auto; }
.audit-detail h3 { font-size: var(--admin-ui-font-size-section); margin: 0 0 var(--admin-ui-space-3); }
.audit-muted { color: var(--admin-ui-color-text-muted); font-size: var(--admin-ui-font-size-caption); }
.audit-target-ids { max-height: 160px; overflow-y: auto; white-space: pre-wrap; overflow-wrap: anywhere; }
.audit-snapshot-note { margin: 0 0 var(--admin-ui-space-3); color: var(--admin-ui-color-text-secondary); font-size: var(--admin-ui-font-size-caption); line-height: 1.6; }
.audit-snapshot-status { margin-bottom: var(--admin-ui-space-3); line-height: 1.6; }
.audit-snapshot-status p { margin: var(--admin-ui-space-1) 0 0; color: var(--admin-ui-color-text-secondary); }
.audit-related-link { display: block; color: var(--el-color-primary); }
.audit-retry { margin-top: var(--admin-ui-space-3); }
</style>
