<template>
  <el-drawer v-model="drawerVisible" title="题目预览" size="520px">
    <div v-if="row" class="preview-panel">
      <el-descriptions :column="1" border>
        <el-descriptions-item label="ID">{{ row.id }}</el-descriptions-item>
        <el-descriptions-item label="标题">{{ questionTitle(row) }}</el-descriptions-item>
        <el-descriptions-item label="题型">{{ typeName(row.type) }}</el-descriptions-item>
        <el-descriptions-item label="分类">{{ row.category || '-' }}</el-descriptions-item>
        <el-descriptions-item label="标签">{{ row.tags || '-' }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatTime(row.addTime) }}</el-descriptions-item>
      </el-descriptions>
      <div class="preview-schema-title">题目配置</div>
      <pre class="preview-schema">{{ prettySchema(row.schema) }}</pre>
    </div>
  </el-drawer>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  visible: boolean
  row: any | null
  questionTitle: (row: any) => string
  typeName: (type: string) => string
  formatTime: (value: number) => string
  prettySchema: (schema: string) => string
}>()

const emit = defineEmits<{
  (event: 'update:visible', value: boolean): void
}>()

const drawerVisible = computed({
  get: () => props.visible,
  set: (value: boolean) => emit('update:visible', value)
})
</script>

<style scoped>
.preview-panel {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.preview-schema-title {
  color: #344054;
  font-size: 13px;
  font-weight: 650;
}

.preview-schema {
  overflow: auto;
  max-height: 420px;
  margin: 0;
  padding: 12px;
  border: 1px solid var(--bank-border);
  border-radius: 8px;
  background: #f8fafc;
  color: #344054;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
  font-size: 12px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-word;
}
</style>
