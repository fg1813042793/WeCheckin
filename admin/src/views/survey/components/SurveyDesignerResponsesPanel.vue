<template>
  <div class="setting-wrapper">
    <div class="setting-scroll">
      <div class="setting-group">
        <div class="group-title">数据统计</div>
        <div v-if="!surveyId" class="save-tip">请先保存并发布问卷</div>
        <div v-else class="responses-content">
          <div class="responses-toolbar">
            <el-button size="small" type="primary" @click="showPendingFeature('添加数据')">添加数据</el-button>
            <el-button size="small" @click="showPendingFeature('编辑数据')">编辑数据</el-button>
            <el-button size="small" type="danger" @click="batchDeleteResponses">删除数据</el-button>
            <el-button size="small" @click="showPendingFeature('导出数据')">导出数据</el-button>
            <el-button size="small" @click="showPendingFeature('导入数据')">导入数据</el-button>
            <el-button size="small" @click="showPendingFeature('回收站')">回收站</el-button>
            <el-button class="refresh-button" size="small" title="刷新数据" @click="refreshResponses">
              <el-icon><Refresh /></el-icon>
            </el-button>
          </div>
          <div class="responses-search">
            <el-input v-model="responseKeyword" placeholder="搜索提交人..." size="small" clearable @input="onSearch" />
          </div>
          <el-table ref="dataTableRef" :data="responseRows" border size="small" max-height="500px" @selection-change="onSelectionChange">
            <el-table-column type="selection" width="40" fixed />
            <el-table-column
              v-for="column in dataColumns"
              :key="column.key"
              :label="column.label"
              :width="column.width"
              :min-width="column.minWidth"
              :show-overflow-tooltip="column.showOverflowTooltip"
              :fixed="column.fixed"
            >
              <template #default="{ row }">
                <span v-if="column.prop">{{ row[column.prop] ?? '-' }}</span>
                <span v-else-if="column.key === 'nickname'">{{ row.nickname || '匿名' }}</span>
                <span v-else-if="column.key === 'startTime'">{{ formatSurveyResponseTime(row.startTime) }}</span>
                <span v-else-if="column.key === 'submitTime'">{{ formatSurveyResponseTime(row.submitTime) }}</span>
                <span v-else-if="column.key === 'duration'">{{ row.duration ? `${row.duration}s` : '-' }}</span>
                <span v-else-if="column.key === 'ip'">{{ row.ip || '-' }}</span>
                <span v-else-if="column.key === 'browser'">{{ row.browser || '-' }}</span>
                <span v-else-if="column.key === 'deviceType'">{{ row.deviceType || '-' }}</span>
                <span v-else-if="column.key === 'platformType'">{{ row.platformType || '-' }}</span>
                <span v-else-if="column.key === 'isAutoSubmit'">
                  <el-tag v-if="row.isAutoSubmit === 1" size="small" type="warning">自动</el-tag>
                  <span v-else>-</span>
                </span>
                <span v-else-if="column.qId">{{ formatSurveyAnswer(row.answers?.[column.qId], column.qType) }}</span>
                <el-button v-else-if="column.key === 'actions'" text size="small" type="danger" @click="deleteResponse(row.id)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
          <div v-if="!responseRows.length" class="responses-empty">暂无提交数据</div>
          <el-pagination
            class="responses-pagination"
            size="small"
            layout="total, sizes, prev, pager, next"
            :total="responseTotal"
            :page-size="responsePageSize"
            :current-page="responsePage"
            :page-sizes="[10, 20, 50, 100]"
            @current-change="onPageChange"
            @size-change="onPageSizeChange"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { Refresh } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import type { SurveyDesignerQuestion } from '../survey-designer-types'
import { formatSurveyAnswer, formatSurveyResponseTime, useSurveyResponses } from '../composables/useSurveyResponses'

const props = defineProps<{
  surveyId: number
  questions: SurveyDesignerQuestion[]
  active: boolean
}>()

const {
  responseRows,
  responseTotal,
  responsePage,
  responsePageSize,
  responseKeyword,
  dataTableRef,
  dataColumns,
  loadResponses,
  onSearch,
  onPageChange,
  onPageSizeChange,
  onSelectionChange,
  deleteResponse,
  batchDeleteResponses,
  refreshResponses,
} = useSurveyResponses({
  getSurveyId: () => props.surveyId,
  getQuestions: () => props.questions,
})

watch(
  () => [props.active, props.surveyId] as const,
  ([active, surveyId]) => { if (active && surveyId) void loadResponses(1) },
  { immediate: true },
)

function showPendingFeature(name: string) {
  ElMessage.info(`${name}功能待实现`)
}
</script>

<style scoped>
.setting-wrapper { height:100%; overflow-y:auto; background:var(--designer-bg); }
.setting-scroll { display:block; padding:20px 24px; }
.setting-group { padding:18px 20px; overflow:hidden; background:#fff; border:1px solid var(--designer-border); border-radius:10px; box-shadow:0 8px 18px rgba(15,23,42,.04); }
.group-title { margin-bottom:14px; color:#344054; font-size:14px; font-weight:650; }
.save-tip { padding:60px 0; color:#98a2b3; font-size:14px; text-align:center; }
.responses-content { padding:12px; }
.responses-toolbar { display:flex; flex-wrap:wrap; align-items:center; gap:8px; margin-bottom:12px; }
.refresh-button { margin-left:auto; }
.responses-search { display:flex; align-items:center; gap:8px; margin-bottom:12px; }
.responses-search :deep(.el-input) { width:200px; }
.responses-empty { padding:40px 0; color:#aaa; font-size:14px; text-align:center; }
.responses-pagination { margin-top:12px; justify-content:flex-start; }
:deep(.el-table .cell), :deep(.el-table th.el-table__cell > .cell) { overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
</style>
