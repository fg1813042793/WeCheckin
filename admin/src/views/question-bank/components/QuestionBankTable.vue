<template>
  <el-table :data="list" v-loading="loading" stripe style="width:100%" empty-text="暂无题目">
    <el-table-column prop="id" label="ID" width="76" />
    <el-table-column label="题目" min-width="260" show-overflow-tooltip>
      <template #default="{ row }">
        <div class="bank-question-title">{{ questionTitle(row) }}</div>
        <div v-if="row.tags || row.category" class="bank-question-meta">
          <el-tag v-if="row.category" size="small" round>{{ row.category }}</el-tag>
          <span v-for="tag in tagList(row.tags)" :key="tag" class="bank-tag">{{ tag }}</span>
        </div>
      </template>
    </el-table-column>
    <el-table-column label="题型" width="150">
      <template #default="{ row }">
        <span class="bank-type-cell">
          <question-icon :type="row.type" class="bank-type-icon" />
          <span>{{ typeName(row.type) }}</span>
        </span>
      </template>
    </el-table-column>
    <el-table-column label="来源" width="100">
      <template #default>{{ activeScope === 'survey' ? '问卷' : '考试' }}</template>
    </el-table-column>
    <el-table-column label="状态" width="90">
      <template #default="{ row }">
        <el-tag size="small" :type="row.status === 1 ? 'success' : 'info'" round>{{ row.status === 1 ? '启用' : '停用' }}</el-tag>
      </template>
    </el-table-column>
    <el-table-column label="创建时间" width="170">
      <template #default="{ row }">{{ formatTime(row.addTime) }}</template>
    </el-table-column>
    <el-table-column label="操作" width="210" fixed="right">
      <template #default="{ row }">
        <div class="admin-table-actions">
          <el-button size="small" @click="$emit('preview', row)">预览</el-button>
          <el-button size="small" type="primary" @click="$emit('edit', row)">编辑</el-button>
          <el-button size="small" type="danger" plain @click="$emit('delete', row)">删除</el-button>
        </div>
      </template>
    </el-table-column>
  </el-table>

  <div class="admin-pagination">
    <el-pagination
      v-model:current-page="currentPage"
      v-model:page-size="pageSizeModel"
      :page-sizes="[10,20,50,100]"
      :total="total"
      layout="total,sizes,prev,pager,next"
      @current-change="$emit('load')"
      @size-change="$emit('size-change')"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import QuestionIcon from '../../survey/formkit/QuestionIcon.vue'
import type { BankScope } from '../utils/importExport'

const props = defineProps<{
  list: any[]
  loading: boolean
  activeScope: BankScope
  page: number
  pageSize: number
  total: number
  questionTitle: (row: any) => string
  tagList: (tags: string) => string[]
  typeName: (type: string) => string
  formatTime: (value: number) => string
}>()

const emit = defineEmits<{
  (event: 'update:page', value: number): void
  (event: 'update:pageSize', value: number): void
  (event: 'load'): void
  (event: 'size-change'): void
  (event: 'preview', row: any): void
  (event: 'edit', row: any): void
  (event: 'delete', row: any): void
}>()

const currentPage = computed({
  get: () => props.page,
  set: (value: number) => emit('update:page', value)
})

const pageSizeModel = computed({
  get: () => props.pageSize,
  set: (value: number) => emit('update:pageSize', value)
})
</script>

<style scoped>
.bank-question-title {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: #1f2937;
  font-weight: 600;
}

.bank-question-meta {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 5px;
  min-width: 0;
}

.bank-tag {
  overflow: hidden;
  max-width: 96px;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: #64748b;
  font-size: 12px;
}

.bank-type-cell {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  min-width: 0;
}

.bank-type-icon {
  width: 16px;
  height: 16px;
  color: #94a3b8;
}
</style>
