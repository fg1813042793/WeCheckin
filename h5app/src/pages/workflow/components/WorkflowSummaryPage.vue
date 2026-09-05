<script setup lang="ts">
import type { WorkflowPublishedDefinition } from '@/types/workflow'
import { onMounted, ref, watch } from 'vue'
import { listWorkflowSummaryDefinitions } from '@/api/workflow'
import { useAppContentStore } from '@/stores'
import WorkflowSummarySection from './WorkflowSummarySection.vue'

const appContent = useAppContentStore()
const definitions = ref<WorkflowPublishedDefinition[]>([])
const loading = ref(false)

watch(
  () => appContent.refreshTick,
  () => {
    if (appContent.currentKey === 'workflow')
      void loadDefinitions()
  },
)

onMounted(() => void loadDefinitions())

async function loadDefinitions() {
  if (loading.value)
    return
  loading.value = true
  try {
    const response = await listWorkflowSummaryDefinitions()
    definitions.value = Array.isArray(response?.data) ? response.data : []
  }
  catch {
    uni.showToast({ title: '流程定义加载失败', icon: 'none' })
  }
  finally {
    loading.value = false
  }
}
</script>

<template>
  <view class="workflow-summary-page">
    <WorkflowSummarySection :definitions="definitions" />
  </view>
</template>

<style lang="scss" scoped>
.workflow-summary-page {
  width: 100%;
  min-height: 0;
  display: flex;
  flex: 1 1 auto;
  flex-direction: column;
  overflow: hidden;
  box-sizing: border-box;
}
</style>
