<template>
  <el-dialog v-model="visible" title="上传到题库" width="420px" :close-on-click-modal="false">
    <el-form label-position="top" size="small">
      <el-form-item label="题库分类">
        <el-select v-model="category" placeholder="选择或输入分类" filterable allow-create clearable class="full-width">
          <el-option v-for="item in categories" :key="item" :label="item" :value="item" />
        </el-select>
      </el-form-item>
      <el-form-item label="标签"><el-input v-model="tags" placeholder="用逗号分隔" /></el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" :loading="saving" @click="emit('confirm', { category, tags })">上传</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { SurveyBankUploadPayload } from '../survey-designer-types'

const emit = defineEmits<{ confirm: [payload: SurveyBankUploadPayload] }>()
const visible = ref(false)
const saving = ref(false)
const category = ref('')
const tags = ref('')
const categories = ref<string[]>([])

function open(initialCategories: string[] = []) {
  category.value = ''
  tags.value = ''
  categories.value = initialCategories
  visible.value = true
}
function setCategories(items: string[]) { categories.value = items }
function setSaving(value: boolean) { saving.value = value }
function close() { visible.value = false }

defineExpose({ open, setCategories, setSaving, close })
</script>

<style scoped>
.full-width { width:100%; }
</style>
