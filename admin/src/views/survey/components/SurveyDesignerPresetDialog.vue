<template>
  <el-dialog v-model="visible" title="保存自定义模版" width="360px" :close-on-click-modal="false">
    <el-form label-position="top" size="small">
      <el-form-item label="模版名称"><el-input v-model="name" placeholder="输入模版名称" maxlength="50" /></el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" @click="save">保存</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const emit = defineEmits<{ save: [name: string] }>()
const visible = ref(false)
const name = ref('')

function open() {
  name.value = ''
  visible.value = true
}
function save() {
  emit('save', name.value)
}
function close() { visible.value = false }

defineExpose({ open, close })
</script>
