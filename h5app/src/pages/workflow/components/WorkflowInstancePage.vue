<script setup lang="ts">
import { computed, ref } from 'vue'
import { useAppContentStore } from '@/stores'
import { workflowInstanceIdFromContentKey } from '../workflow-route-keys'
import WorkflowDetailPanel from './WorkflowDetailPanel.vue'

const props = defineProps<{
  contentKey: string
}>()

const appContent = useAppContentStore()
const detailVisible = ref(true)
const instanceId = computed(() => workflowInstanceIdFromContentKey(props.contentKey))
const detailTitle = computed(() => appContent.dynamicTab(props.contentKey)?.label || '流程详情')

function handleVisibleChange(visible: boolean) {
  detailVisible.value = visible
  if (!visible)
    appContent.requestCloseTab(props.contentKey)
}

function handleChanged() {
  appContent.switchContent('workflow')
  appContent.removeDynamicTab(props.contentKey)
  appContent.requestRefresh()
}
</script>

<template>
  <view class="workflow-instance-page">
    <WorkflowDetailPanel
      :model-value="detailVisible"
      :instance-id="instanceId"
      :display-title="detailTitle"
      presentation="history-page"
      comment-action
      @update:model-value="handleVisibleChange"
      @changed="handleChanged"
    />
  </view>
</template>

<style lang="scss" scoped>
.workflow-instance-page {
  width: 100%;
  height: 100%;
  min-height: 0;
}
</style>
