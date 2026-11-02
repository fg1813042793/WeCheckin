<script setup lang="ts">
import { ref, watch } from 'vue'
import { AdminDialog } from '@/components/admin-ui'
import BusinessNotificationStyles from './BusinessNotificationStyles.vue'
import NotificationTemplateManager from './NotificationTemplateManager.vue'
const props = defineProps<{ modelValue: boolean; canEdit: boolean; canSend: boolean; canSendDingTalk: boolean }>()
const emit = defineEmits<{ 'update:modelValue': [value: boolean]; sent: [channel: 'in_app' | 'dingtalk'] }>()
const activeTab = ref('business')
watch(() => props.modelValue, visible => { if (visible) activeTab.value = 'business' })
</script>

<template>
  <AdminDialog :model-value="modelValue" title="消息样式" width="min(1080px, 96vw)" @update:model-value="emit('update:modelValue', $event)">
    <el-tabs v-if="modelValue" v-model="activeTab">
      <el-tab-pane name="business" label="业务消息样式" lazy>
        <BusinessNotificationStyles :can-edit="canEdit" :can-send="canSend" :can-send-ding-talk="canSendDingTalk" @sent="emit('sent', $event)" />
      </el-tab-pane>
      <el-tab-pane name="templates" label="发送模板" lazy>
        <NotificationTemplateManager :can-edit="canEdit" />
      </el-tab-pane>
    </el-tabs>
    <template #footer><el-button @click="emit('update:modelValue', false)">关闭</el-button></template>
  </AdminDialog>
</template>
