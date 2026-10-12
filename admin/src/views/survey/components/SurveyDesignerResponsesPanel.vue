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
            <span class="responses-quick-actions">
              <el-button size="small" @click="openSurveyPreview"><el-icon><Document /></el-icon>查看问卷详情</el-button>
              <el-button size="small" type="success" plain @click="openStatReport"><el-icon><View /></el-icon>查看统计报表</el-button>
            </span>
            <el-button class="refresh-button" size="small" title="刷新数据" @click="refreshResponses">
              <el-icon><Refresh /></el-icon>
            </el-button>
          </div>
          <div class="responses-search">
            <el-input v-model="responseKeyword" placeholder="按提交人搜索" size="small" clearable @input="onSearch" />
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
                <div v-else-if="column.key === 'actions'" class="response-actions">
                  <el-button text size="small" type="primary" @click="openResponseDetail(row.id)">查看详情</el-button>
                  <el-button text size="small" type="success" @click="openResponseStatistics(row.id)"><el-icon><DataAnalysis /></el-icon>统计结果</el-button>
                  <el-button text size="small" type="danger" @click="deleteResponse(row.id)">删除</el-button>
                </div>
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

    <el-dialog v-model="responseDetail.visible" :title="detailMode === 'statistics' ? '本份答卷统计结果' : '答卷详情'" width="520px" destroy-on-close>
      <el-skeleton v-if="responseDetail.loading" :rows="6" animated />
      <template v-else-if="responseDetail.response">
        <template v-if="detailMode === 'statistics'">
          <div class="response-stat-summary">
            <span>提交人：{{ responseDetail.response.nickname || '匿名' }}</span>
            <span>统计项：{{ responseStatisticTotal }} 项</span>
          </div>
          <el-table v-if="responseStatisticRows.length" :data="responseStatisticRows" border stripe size="small" class="response-stat-table">
            <el-table-column prop="value" :label="responseStatisticLabel" width="180" show-overflow-tooltip />
            <el-table-column prop="count" label="次数" width="90" align="center" />
            <el-table-column prop="percentage" label="占比" width="100" align="center" />
          </el-table>
          <el-empty v-else description="此份答卷没有可统计的选择题结果" :image-size="60" />
        </template>
        <template v-else>
          <el-descriptions :column="2" border size="small">
            <el-descriptions-item label="提交人">{{ responseDetail.response.nickname || '匿名' }}</el-descriptions-item>
            <el-descriptions-item label="用户 ID">{{ responseDetail.response.userId || '-' }}</el-descriptions-item>
            <el-descriptions-item label="开始时间">{{ formatSurveyResponseTime(responseDetail.response.startTime) }}</el-descriptions-item>
            <el-descriptions-item label="提交时间">{{ formatSurveyResponseTime(responseDetail.response.submitTime) }}</el-descriptions-item>
            <el-descriptions-item label="答题时长">{{ responseDetail.response.duration ? `${responseDetail.response.duration}s` : '-' }}</el-descriptions-item>
            <el-descriptions-item label="IP 地址">{{ responseDetail.response.ip || '-' }}</el-descriptions-item>
            <el-descriptions-item label="浏览器">{{ responseDetail.response.browser || '-' }}</el-descriptions-item>
            <el-descriptions-item label="设备类型">{{ responseDetail.response.deviceType || '-' }}</el-descriptions-item>
            <el-descriptions-item label="平台类型">{{ responseDetail.response.platformType || '-' }}</el-descriptions-item>
          </el-descriptions>
          <el-table :data="responseAnswerRows" border stripe size="small" class="response-detail-table">
            <el-table-column prop="number" label="#" width="55" />
            <el-table-column prop="title" label="题目" min-width="260" show-overflow-tooltip />
            <el-table-column prop="value" label="答案" min-width="220" show-overflow-tooltip />
          </el-table>
        </template>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { DataAnalysis, Document, Refresh, View } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'
import type { SurveyDesignerQuestion } from '../survey-designer-types'
import { formatSurveyAnswer, formatSurveyResponseTime, useSurveyResponses } from '../composables/useSurveyResponses'

const props = defineProps<{
  surveyId: number
  questions: SurveyDesignerQuestion[]
  resultConfig: { statType?: string }
  active: boolean
}>()

const router = useRouter()
const detailMode = ref<'detail' | 'statistics'>('detail')

const {
  responseRows,
  responseTotal,
  responsePage,
  responsePageSize,
  responseKeyword,
  responseDetail,
  dataTableRef,
  dataColumns,
  loadResponses,
  onSearch,
  onPageChange,
  onPageSizeChange,
  onSelectionChange,
  viewResponse,
  deleteResponse,
  batchDeleteResponses,
  refreshResponses,
} = useSurveyResponses({
  getSurveyId: () => props.surveyId,
  getQuestions: () => props.questions,
})

const responseAnswerRows = computed(() => {
  const answers = responseDetail.value.answers
  return props.questions
    .filter(question => !['divider', 'description', 'pagination', 'questionSet'].includes(question.type))
    .map((question, index) => {
      const value = formatSurveyAnswer(answers[question.id], question.type)
      return {
        number: index + 1,
        title: stripHtml(question.title),
        value: value === '-' ? '未填写' : value,
      }
    })
})

const responseStatisticRows = computed(() => {
  const counts = new Map<string, { value: string; label: string; count: number }>()
  const answers = responseDetail.value.answers
  const statType = props.resultConfig?.statType || 'value'
  const statisticTypes = new Set(['radio', 'checkbox', 'select', 'picker', 'judge', 'rating', 'nps', 'score'])

  function addAnswer(question: SurveyDesignerQuestion, raw: unknown) {
    if (raw === null || raw === undefined || raw === '') return
    const option = findOption(question, raw)
    const optionValue = String(option?.value ?? option?.id ?? raw)
    const label = stripHtml(String(option?.label ?? option?.value ?? raw))
    const value = statType === 'count'
      ? label
      : statType === 'score'
        ? String(option?.score ?? raw)
        : optionValue
    const existing = counts.get(value)
    if (existing) existing.count += 1
    else counts.set(value, { value, label, count: 1 })
  }

  for (const question of props.questions) {
    if (!statisticTypes.has(question.type)) continue
    const answer = answers[question.id]
    if (question.type === 'checkbox' && Array.isArray(answer)) {
      answer.forEach(value => addAnswer(question, value))
    } else {
      addAnswer(question, answer)
    }
  }

  const total = Array.from(counts.values()).reduce((sum, item) => sum + item.count, 0)
  return Array.from(counts.values())
    .sort((left, right) => right.count - left.count || left.value.localeCompare(right.value))
    .map(item => ({ ...item, percentage: `${((item.count / total) * 100).toFixed(1)}%` }))
})

const responseStatisticLabel = computed(() => {
  const statType = props.resultConfig?.statType || 'value'
  if (statType === 'count') return '统计项'
  if (statType === 'score') return '分值'
  return '选项值'
})

const responseStatisticTotal = computed(() => responseStatisticRows.value.reduce((sum, item) => sum + item.count, 0))

watch(
  () => [props.active, props.surveyId] as const,
  ([active, surveyId]) => { if (active && surveyId) void loadResponses(1) },
  { immediate: true },
)

function showPendingFeature(name: string) {
  ElMessage.info(`${name}功能待实现`)
}

function openResponseDetail(id: number) {
  detailMode.value = 'detail'
  void viewResponse(id)
}

function openResponseStatistics(id: number) {
  detailMode.value = 'statistics'
  void viewResponse(id)
}

function openSurveyPreview() {
  openPage(`/sf1/${props.surveyId}`)
}

function openStatReport() {
  openPage('/survey/stat-report')
}

function openPage(path: string) {
  const query = path.startsWith('/survey/') ? { surveyId: String(props.surveyId) } : undefined
  const href = router.resolve({ path, query }).href
  window.open(href, '_blank', 'noopener,noreferrer')
}

function stripHtml(value: string) {
  const container = document.createElement('div')
  container.innerHTML = value || ''
  return container.textContent || container.innerText || ''
}

function findOption(question: SurveyDesignerQuestion, raw: unknown) {
  const options = Array.isArray(question.props?.options) ? question.props.options : []
  return options.find((option: any) => String(option.value ?? option.id ?? option.label) === String(raw) || String(option.label) === String(raw))
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
.responses-quick-actions { display:flex; flex-wrap:wrap; gap:8px; margin-left:4px; }
.refresh-button { margin-left:auto; }
.responses-search { display:flex; align-items:center; gap:8px; margin-bottom:12px; }
.responses-search :deep(.el-input) { width:200px; }
.responses-empty { padding:40px 0; color:#aaa; font-size:14px; text-align:center; }
.responses-pagination { margin-top:12px; justify-content:flex-start; }
:deep(.el-table .cell), :deep(.el-table th.el-table__cell > .cell) { overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.response-actions { display:flex; align-items:center; gap:2px; white-space:nowrap; }
.response-stat-summary { display:flex; justify-content:space-between; margin-bottom:12px; color:#606266; font-size:13px; }
.response-stat-table { width:370px; max-width:100%; }
.response-detail-table { margin-top:16px; }
</style>
