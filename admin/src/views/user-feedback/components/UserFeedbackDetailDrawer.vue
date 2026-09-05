<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { AdminDrawer } from '@/components/admin-ui'
import { adminApi } from '@/api'
import type { UserFeedbackDetail } from '@/types/userFeedback'
import { showRequestError } from '@/utils/request'
import { canHandleUserFeedback, userFeedbackStatusMeta } from '../userFeedbackStatus'
import UserFeedbackTimeline from './UserFeedbackTimeline.vue'

const props = withDefaults(defineProps<{
  modelValue: boolean
  feedbackId?: number | null
  canHandle?: boolean
  formatTime: (timestamp: number) => string
}>(), {
  feedbackId: null,
  canHandle: false,
})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'request-status-update': [detail: UserFeedbackDetail]
  'detail-change': [detail: UserFeedbackDetail | null]
}>()

const detail = ref<UserFeedbackDetail | null>(null)
const loading = ref(false)
const error = ref('')
let detailRequestSequence = 0

const visible = computed({
  get: () => props.modelValue,
  set: value => emit('update:modelValue', value),
})
const drawerTitle = computed(() => detail.value ? `反馈详情 · ${detail.value.feedbackNo}` : '反馈详情')

function personLabel(name: string | undefined, id: number | undefined, fallback: string) {
  if (name) return name
  if (id) return `#${id}`
  return fallback
}

async function loadDetail() {
  const feedbackId = props.feedbackId
  if (!visible.value || !feedbackId) return
  const sequence = ++detailRequestSequence
  loading.value = true
  error.value = ''
  try {
    const response = await adminApi.userFeedbackDetail(feedbackId)
    if (sequence !== detailRequestSequence) return
    detail.value = response.data
    emit('detail-change', response.data)
  } catch (requestError) {
    if (sequence !== detailRequestSequence) return
    detail.value = null
    emit('detail-change', null)
    error.value = '反馈详情加载失败，请重试'
    showRequestError(requestError, '反馈详情加载失败')
  } finally {
    if (sequence === detailRequestSequence) loading.value = false
  }
}

function requestStatusUpdate() {
  if (!canHandleUserFeedback()) {
    ElMessage.warning('当前账号没有反馈处理权限')
    return
  }
  if (!detail.value) return
  emit('request-status-update', detail.value)
}

function handleClosed() {
  detailRequestSequence += 1
  detail.value = null
  error.value = ''
  emit('detail-change', null)
}

watch([() => props.modelValue, () => props.feedbackId], ([isVisible, feedbackId], previous) => {
  if (!isVisible || !feedbackId) return
  if (!previous || previous[1] !== feedbackId) {
    detail.value = null
    emit('detail-change', null)
  }
  void loadDetail()
})

defineExpose({ refresh: loadDetail })
</script>

<template>
  <AdminDrawer
    v-model="visible"
    :title="drawerTitle"
    size="lg"
    :show-footer="Boolean(canHandle && detail && !loading && !error)"
    confirm-text="更新状态"
    append-to-body
    @confirm="requestStatusUpdate"
    @closed="handleClosed"
  >
    <div v-loading="loading" class="feedback-detail">
      <el-alert v-if="error" :title="error" type="error" show-icon :closable="false">
        <template #default>
          <el-button link type="primary" @click="loadDetail">重新加载</el-button>
        </template>
      </el-alert>
      <template v-else-if="detail">
        <section class="feedback-detail__summary" aria-label="反馈概况">
          <div class="feedback-detail__identity">
            <div>
              <strong>{{ detail.feedbackNo }}</strong>
              <span>版本 {{ detail.version }}</span>
            </div>
            <el-tag :type="userFeedbackStatusMeta(detail.status).type">
              {{ userFeedbackStatusMeta(detail.status).label }}
            </el-tag>
          </div>
          <el-descriptions :column="2" border>
            <el-descriptions-item label="提交人">
              {{ personLabel(detail.submitterName, detail.submitterId, '未知用户') }}
            </el-descriptions-item>
            <el-descriptions-item label="处理人">
              {{ personLabel(detail.handlerName, detail.handlerId, '未分配') }}
            </el-descriptions-item>
            <el-descriptions-item label="提交时间">{{ formatTime(detail.createdAt) }}</el-descriptions-item>
            <el-descriptions-item label="最近活动">{{ formatTime(detail.lastActivityAt) }}</el-descriptions-item>
            <el-descriptions-item label="图片数量">{{ detail.imageCount }}</el-descriptions-item>
            <el-descriptions-item label="更新时间">{{ formatTime(detail.updatedAt) }}</el-descriptions-item>
          </el-descriptions>
        </section>

        <section class="feedback-detail__timeline" aria-label="反馈时间线">
          <header>
            <h3>反馈记录</h3>
            <span>共 {{ detail.messages.length }} 条</span>
          </header>
          <UserFeedbackTimeline :messages="detail.messages" :format-time="formatTime" />
        </section>
      </template>
    </div>

  </AdminDrawer>
</template>

<style scoped>
.feedback-detail {
  min-height: 320px;
}

.feedback-detail__summary,
.feedback-detail__timeline {
  min-width: 0;
}

.feedback-detail__identity {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--admin-ui-space-3);
  margin-bottom: var(--admin-ui-space-4);
}

.feedback-detail__identity > div {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.feedback-detail__identity strong {
  overflow-wrap: anywhere;
  color: var(--admin-ui-color-text);
  font-size: var(--admin-ui-font-size-section);
}

.feedback-detail__identity span,
.feedback-detail__timeline header span {
  color: var(--admin-ui-color-text-muted);
  font-size: var(--admin-ui-font-size-caption);
}

.feedback-detail__timeline {
  margin-top: var(--admin-ui-space-6);
  padding-top: var(--admin-ui-space-5);
  border-top: 1px solid var(--admin-ui-color-border);
}

.feedback-detail__timeline header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--admin-ui-space-3);
  margin-bottom: var(--admin-ui-space-4);
}

.feedback-detail__timeline h3 {
  margin: 0;
  color: var(--admin-ui-color-text);
  font-size: var(--admin-ui-font-size-section);
  letter-spacing: 0;
}

@media (max-width: 767px) {
  .feedback-detail :deep(.el-descriptions__body .el-descriptions__table) {
    table-layout: fixed;
  }
}
</style>
