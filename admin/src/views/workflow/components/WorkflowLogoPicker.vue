<template>
  <div class="workflow-logo-picker">
    <div class="workflow-logo-picker__preview">
      <img v-if="previewUrl" :src="previewUrl" alt="流程 Logo 预览" />
      <el-icon v-else><Picture /></el-icon>
    </div>
    <div class="workflow-logo-picker__actions">
      <el-upload
        :auto-upload="false"
        :show-file-list="false"
        :disabled="disabled"
        accept="image/png,image/jpeg,image/webp"
        :on-change="handleChange"
      >
        <el-button icon="Upload" :disabled="disabled">{{ previewUrl ? '更换图片' : '选择图片' }}</el-button>
      </el-upload>
      <el-button v-if="previewUrl" icon="Delete" type="danger" plain :disabled="disabled" @click="clearLogo">移除</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import type { UploadFile } from 'element-plus'

const props = withDefaults(defineProps<{
  imageUrl?: string
  file?: File | null
  disabled?: boolean
}>(), {
  imageUrl: '',
  file: null,
  disabled: false,
})

const emit = defineEmits<{
  'update:file': [file: File | null]
  clear: []
}>()

const objectUrl = ref('')
const previewUrl = computed(() => objectUrl.value || props.imageUrl)

function revokeObjectUrl() {
  if (!objectUrl.value) return
  URL.revokeObjectURL(objectUrl.value)
  objectUrl.value = ''
}

watch(() => props.file, (file) => {
  revokeObjectUrl()
  if (file) objectUrl.value = URL.createObjectURL(file)
}, { immediate: true })

function handleChange(uploadFile: UploadFile) {
  const file = uploadFile.raw
  if (!file) return
  if (!['image/png', 'image/jpeg', 'image/webp'].includes(file.type)) {
    ElMessage.warning('流程 Logo 仅支持 PNG、JPG、JPEG 或 WebP 格式')
    return
  }
  if (file.size > 2 * 1024 * 1024) {
    ElMessage.warning('流程 Logo 不能超过 2MB')
    return
  }
  emit('update:file', file)
}

function clearLogo() {
  emit('update:file', null)
  emit('clear')
}

onBeforeUnmount(revokeObjectUrl)
</script>

<style scoped>
.workflow-logo-picker { display: flex; align-items: center; gap: 14px; }
.workflow-logo-picker__preview { display: grid; place-items: center; width: 72px; height: 72px; flex: 0 0 72px; overflow: hidden; border: 1px solid #dcdfe6; border-radius: 8px; color: #909399; background: #f5f7fa; font-size: 26px; }
.workflow-logo-picker__preview img { width: 100%; height: 100%; object-fit: cover; }
.workflow-logo-picker__actions { display: flex; flex-wrap: wrap; gap: 8px; }
</style>
