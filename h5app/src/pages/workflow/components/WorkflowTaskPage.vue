<script setup lang="ts">
import { computed, ref } from 'vue'
import { useAppContentStore } from '@/stores'
import {
  workflowTaskIdFromContentKey,
  workflowTaskInstanceIdFromContentKey,
} from '../workflow-route-keys'
import WorkflowDetailPanel from './WorkflowDetailPanel.vue'

const props = defineProps<{
  contentKey: string
}>()

const appContent = useAppContentStore()
const detailVisible = ref(true)
const taskId = computed(() => workflowTaskIdFromContentKey(props.contentKey))
const instanceId = computed(() => workflowTaskInstanceIdFromContentKey(props.contentKey))
const taskTitle = computed(() => appContent.dynamicTab(props.contentKey)?.label || '流程办理')

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
  <view class="workflow-task-page">
    <WorkflowDetailPanel
      :model-value="detailVisible"
      :instance-id="instanceId"
      :task-id="taskId"
      :display-title="taskTitle"
      presentation="page"
      comment-action
      @update:model-value="handleVisibleChange"
      @changed="handleChanged"
    />
  </view>
</template>

<style lang="scss" scoped>
.workflow-task-page {
  width: 100%;
  height: 100%;
  min-height: 0;
}
</style>
