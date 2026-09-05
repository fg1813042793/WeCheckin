<script setup lang="ts">
import type {
  WorkflowFieldAccessMap,
  WorkflowFormData,
  WorkflowInstanceDetail,
} from '@/types/workflow'
import { computed, ref, watch } from 'vue'
import { getWorkflowInstance } from '@/api/workflow'
import { useAppContentStore } from '@/stores'
import {
  initialWorkflowFormData,
  workflowFieldAccessMap,
} from '../workflow-form'
import { workflowFormDetailInstanceIdFromContentKey } from '../workflow-route-keys'
import { workflowInstanceStatusMeta } from '../workflow-status'
import WorkflowRuntimeForm from './WorkflowRuntimeForm.vue'

const props = defineProps<{
  contentKey: string
}>()

const appContent = useAppContentStore()
const detail = ref<WorkflowInstanceDetail | null>(null)
const formData = ref<WorkflowFormData>({})
const loading = ref(false)
const loadError = ref('')

const instanceId = computed(() => workflowFormDetailInstanceIdFromContentKey(props.contentKey))
const title = computed(() => (
  detail.value?.instance.definitionName
  || appContent.dynamicTab(props.contentKey)?.label
  || '表单详情'
))
const fieldAccess = computed<WorkflowFieldAccessMap>(() => (
  workflowFieldAccessMap(detail.value?.form || [], [], 'read')
))

watch(
  () => [props.contentKey, appContent.currentKey, appContent.refreshTick],
  () => {
    if (appContent.currentKey === props.contentKey)
      void loadDetail()
  },
  { immediate: true },
)

async function loadDetail() {
  if (!instanceId.value || loading.value)
    return
  loading.value = true
  loadError.value = ''
  try {
    const response = await getWorkflowInstance(instanceId.value)
    if (!response?.data) {
      loadError.value = '流程表单加载失败'
      return
    }
    detail.value = response.data
    formData.value = initialWorkflowFormData(response.data.form || [], response.data.formData || {})
  }
  catch {
    loadError.value = '流程表单加载失败，请稍后重试'
  }
  finally {
    loading.value = false
  }
}
</script>

<template>
  <view class="workflow-form-detail-page app-pc-control-scope">
    <view class="workflow-form-detail-page__header">
      <view class="workflow-form-detail-page__heading">
        <text class="workflow-form-detail-page__title">
          {{ title }}
        </text>
        <text v-if="detail" class="workflow-form-detail-page__meta">
          业务编号：{{ detail.instance.businessKey || '-' }} · 发起人：{{ detail.instance.starterName || '未知用户' }}
        </text>
      </view>
      <u-tag
        v-if="detail"
        :text="workflowInstanceStatusMeta(detail.instance.status).label"
        :type="workflowInstanceStatusMeta(detail.instance.status).type"
        size="mini"
      />
    </view>

    <view v-if="loading" class="workflow-form-detail-page__state">
      <u-loading mode="circle" size="24px" />
      <text>正在加载流程表单...</text>
    </view>
    <view v-else-if="loadError" class="workflow-form-detail-page__state workflow-form-detail-page__state--error">
      <u-icon name="info-circle" size="28px" color="#f56c6c" />
      <text>{{ loadError }}</text>
      <u-button size="small" type="primary" plain @click="loadDetail">
        重新加载
      </u-button>
    </view>
    <scroll-view v-else-if="detail" scroll-y class="workflow-form-detail-page__body">
      <view class="workflow-form-detail-page__content">
        <view class="workflow-form-detail-page__section-head">
          <text class="workflow-form-detail-page__section-title">
            申请表单
          </text>
          <text class="workflow-form-detail-page__section-desc">
            发起时提交的表单内容
          </text>
        </view>
        <WorkflowRuntimeForm
          v-model="formData"
          class="workflow-form-detail-page__form app-workflow-form app-pc-control-scope"
          :fields="detail.form || []"
          :field-access="fieldAccess"
          :readonly="true"
          readonly-appearance="plain"
        />
      </view>
    </scroll-view>
  </view>
</template>

<style lang="scss" scoped>
.workflow-form-detail-page {
  width: 100%;
  height: 100%;
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: #f6f8fb;
  color: #1f2329;
}

.workflow-form-detail-page__header {
  min-height: 72px;
  padding: 14px 24px;
  border-bottom: 1px solid #e5eaf3;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  box-sizing: border-box;
  background: #fff;
}

.workflow-form-detail-page__heading {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.workflow-form-detail-page__title {
  color: #1f2329;
  font-size: 18px;
  line-height: 1.35;
  font-weight: 700;
}

.workflow-form-detail-page__meta,
.workflow-form-detail-page__section-desc {
  color: #86909c;
  font-size: 12px;
  line-height: 1.5;
}

.workflow-form-detail-page__body {
  flex: 1;
  min-height: 0;
  width: min(var(--app-pc-content-max-width, 1080px), 100%);
  margin: 0 auto;
  padding: 20px 24px 32px;
  box-sizing: border-box;
}

.workflow-form-detail-page__content {
  width: 100%;
  padding: 20px;
  border: 1px solid #dfe5ee;
  border-radius: 6px;
  background: #fff;
  box-shadow: 0 2px 8px rgba(31, 35, 41, 0.05);
  box-sizing: border-box;
}

.workflow-form-detail-page__section-head {
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid #e5eaf3;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.workflow-form-detail-page__section-title {
  font-size: 16px;
  font-weight: 700;
}

.workflow-form-detail-page__form {
  width: 100%;
}

.workflow-form-detail-page__state {
  flex: 1;
  min-height: 240px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  color: #86909c;
}

.workflow-form-detail-page__state--error {
  flex-direction: column;
  color: #4e5969;
}

.workflow-form-detail-page__state :deep(.u-btn) {
  width: auto;
  min-width: 88px;
  margin: 4px 0 0;
}

@media screen and (max-width: 768px) {
  .workflow-form-detail-page__header {
    padding-left: 16px;
    padding-right: 16px;
  }

  .workflow-form-detail-page__body {
    padding: 16px 12px;
  }

  .workflow-form-detail-page__content {
    padding: 12px;
  }

  .workflow-form-detail-page__section-head {
    margin-bottom: 16px;
    padding-bottom: 12px;
  }
}
</style>
