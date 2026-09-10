<template>
  <div class="setting-wrapper">
    <div class="report-scroll">
      <div class="setting-group">
        <div class="group-title">报表</div>
        <div v-if="!surveyId" class="save-tip">请先保存并发布问卷</div>
        <div v-else-if="loading" class="report-loading">加载中...</div>
        <div v-else-if="!reportData?.fieldStats?.length" class="report-empty">
          <el-empty description="暂无数据" :image-size="50" />
        </div>
        <div v-else class="report-list">
          <div v-for="(field, fieldIndex) in reportData.fieldStats" :key="field?.questionId || fieldIndex" class="report-card">
            <template v-if="field && field.totalCount > 0 && field.dist">
              <div class="report-card-header">
                <span class="report-title">{{ Number(fieldIndex) + 1 }}. {{ firstLine(stripHtml(field.title)) }}</span>
                <span class="chart-type-btns">
                  <button :class="{ active: chartType(field.questionId) === 'bar' }" title="条形图" @click="setChartType(field.questionId, 'bar')">≡</button>
                  <button :class="{ active: chartType(field.questionId) === 'column' }" title="柱形图" @click="setChartType(field.questionId, 'column')">▯</button>
                  <button :class="{ active: chartType(field.questionId) === 'pie' }" title="扇形图" @click="setChartType(field.questionId, 'pie')">◯</button>
                </span>
              </div>
              <div v-if="chartType(field.questionId) === 'bar'" class="bar-chart">
                <div v-for="(count, label) in field.dist" :key="label" class="bar-row">
                  <span class="chart-label">{{ label || '(空)' }}</span>
                  <div class="bar-track"><div class="bar-value" :style="{ width: percentage(count, field.totalCount) + '%' }"></div></div>
                  <span class="chart-count">{{ count }} / {{ field.totalCount }}</span>
                </div>
              </div>
              <div v-else-if="chartType(field.questionId) === 'column'" class="column-chart">
                <div v-for="(count, label) in field.dist" :key="label" class="column-item">
                  <span class="column-count">{{ count }}</span>
                  <div class="column-value" :style="{ height: Math.max(Number(count) / field.totalCount * 140, 4) + 'px' }"></div>
                  <span class="column-label">{{ label || '(空)' }}</span>
                </div>
              </div>
              <div v-else class="pie-chart">
                <svg viewBox="0 0 160 160" class="pie-graphic">
                  <circle cx="80" cy="80" r="70" fill="none" stroke="#f0f0f0" stroke-width="30" />
                  <circle
                    v-for="(count, label) in field.dist"
                    :key="label"
                    cx="80"
                    cy="80"
                    r="70"
                    fill="none"
                    :stroke="pieColor(field.dist, label)"
                    stroke-width="30"
                    :stroke-dasharray="pieDasharray(Number(count), field.totalCount)"
                  />
                </svg>
                <div class="pie-legend">
                  <div v-for="(count, label) in field.dist" :key="label" class="pie-legend-row">
                    <span class="pie-dot" :style="{ background: pieColor(field.dist, label) }"></span>
                    <span class="pie-label">{{ label || '(空)' }}</span>
                    <span class="pie-percent">{{ percentage(count, field.totalCount) }}%</span>
                  </div>
                </div>
              </div>
            </template>
            <template v-else-if="field?.tableData?.length">
              <div class="table-title">{{ Number(fieldIndex) + 1 }}. {{ firstLine(stripHtml(field.title)) }}</div>
              <div class="table-scroll">
                <el-table :data="field.tableData" border size="small">
                  <el-table-column type="index" label="#" width="40" fixed />
                  <el-table-column v-for="(column, columnIndex) in (field.tableCols || ['回答'])" :key="columnIndex" :label="column" min-width="100">
                    <template #default="{ row }">{{ stripHtml(row[columnIndex] ?? '') }}</template>
                  </el-table-column>
                </el-table>
              </div>
            </template>
            <div v-else-if="field?.numericStat" class="numeric-stat">
              总和: {{ field.numericStat.sum }} | 平均: {{ field.numericStat.avg.toFixed(2) }} | 最小: {{ field.numericStat.min }} | 最大: {{ field.numericStat.max }}
            </div>
            <div v-else class="no-stat">暂无统计</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { adminApi } from '../../../api'

interface FieldStatistic {
  questionId: string
  title: string
  totalCount: number
  dist?: Record<string, number>
  tableData?: Array<Record<number, unknown>>
  tableCols?: string[]
  numericStat?: { sum: number; avg: number; min: number; max: number }
}

interface SurveyReportData {
  fieldStats?: FieldStatistic[]
}

const props = defineProps<{ surveyId: number; active: boolean }>()
const reportData = ref<SurveyReportData | null>(null)
const loading = ref(false)
const chartTypes = ref<Record<string, string>>({})
const PIE_COLORS = ['#2563eb', '#06b6d4', '#22c55e', '#f59e0b', '#8b5cf6', '#14b8a6', '#64748b', '#f97316']

watch(
  () => [props.active, props.surveyId] as const,
  ([active, surveyId]) => {
    if (!surveyId) reportData.value = null
    else if (active) void loadReport()
  },
  { immediate: true },
)

async function loadReport() {
  loading.value = true
  try {
    const response: any = await adminApi.surveyStatistic(props.surveyId)
    reportData.value = response.data || response
  } catch {
    reportData.value = null
  } finally {
    loading.value = false
  }
}

function chartType(questionId: string) { return chartTypes.value[questionId] || 'bar' }
function setChartType(questionId: string, type: string) { chartTypes.value[questionId] = type }
function percentage(count: number, total: number) { return Math.max(Math.round(Number(count || 0) / total * 100), 0) }
function firstLine(value: string) { return value.split('\n')[0] || value }
function stripHtml(html: string) {
  const container = document.createElement('div')
  container.innerHTML = html
  return container.textContent || container.innerText || ''
}
function pieDasharray(count: number, total: number) {
  const circumference = 2 * Math.PI * 70
  const ratio = count / total
  return `${ratio * circumference} ${circumference - ratio * circumference}`
}
function pieColor(distribution: Record<string, number>, label: string | number) {
  return PIE_COLORS[Object.keys(distribution).indexOf(String(label)) % PIE_COLORS.length]
}
</script>

<style scoped>
.setting-wrapper { height:100%; overflow-y:auto; background:var(--designer-bg); }
.report-scroll { padding:20px 24px; }
.setting-group { padding:18px 20px; overflow:hidden; background:#fff; border:1px solid var(--designer-border); border-radius:10px; box-shadow:0 8px 18px rgba(15,23,42,.04); }
.group-title { margin-bottom:14px; color:#344054; font-size:14px; font-weight:650; }
.save-tip, .report-loading { padding:60px 0; color:#98a2b3; font-size:14px; text-align:center; }
.report-empty { padding:12px; }
.report-list { max-width:600px; margin:0 auto; padding:12px; }
.report-card { margin-bottom:16px; padding:12px; border:1px solid #eee; border-radius:6px; }
.report-card-header { display:flex; align-items:center; gap:8px; margin-bottom:8px; }
.report-title { flex:1; overflow:hidden; font-size:13px; font-weight:600; text-overflow:ellipsis; white-space:nowrap; }
.chart-type-btns { display:flex; flex-shrink:0; gap:2px; }
.chart-type-btns button { display:flex; width:26px; height:26px; align-items:center; justify-content:center; border:1px solid #e8e8e8; border-radius:4px; background:#fff; color:#888; font-size:14px; line-height:1; cursor:pointer; }
.chart-type-btns button:hover { border-color:var(--designer-accent); color:var(--designer-accent); }
.chart-type-btns button.active { border-color:var(--designer-accent); background:var(--designer-accent); color:#fff; }
.bar-chart { display:flex; flex-direction:column; gap:4px; }
.bar-row { display:flex; align-items:center; gap:8px; }
.chart-label { width:100px; flex-shrink:0; overflow:hidden; font-size:12px; text-overflow:ellipsis; white-space:nowrap; }
.bar-track { flex:1; height:12px; overflow:hidden; border-radius:6px; background:#e8e8e8; }
.bar-value { height:100%; border-radius:6px; background:#409eff; transition:width .3s; }
.chart-count { width:60px; color:#999; font-size:11px; text-align:right; }
.column-chart { display:flex; min-height:160px; align-items:flex-end; justify-content:center; gap:12px; padding:16px 0; }
.column-item { display:flex; max-width:80px; flex:1; flex-direction:column; align-items:center; gap:4px; }
.column-count, .column-label { font-size:11px; }
.column-count { color:#999; }
.column-value { width:100%; border-radius:4px 4px 0 0; background:#2563eb; transition:height .3s; }
.column-label { max-width:100%; overflow:hidden; color:#666; text-align:center; text-overflow:ellipsis; white-space:nowrap; }
.pie-chart { display:flex; justify-content:center; padding:12px 0; }
.pie-graphic { width:160px; height:160px; transform:rotate(-90deg); }
.pie-legend { display:flex; flex-direction:column; justify-content:center; gap:4px; margin-left:16px; font-size:12px; }
.pie-legend-row { display:flex; align-items:center; gap:6px; }
.pie-dot { display:inline-block; width:10px; height:10px; border-radius:50%; }
.pie-label { max-width:80px; overflow:hidden; color:#666; text-overflow:ellipsis; white-space:nowrap; }
.pie-percent { color:#999; }
.table-title { margin-bottom:6px; font-size:13px; font-weight:600; }
.table-scroll { max-height:200px; overflow:auto; }
.numeric-stat { color:#606266; font-size:12px; line-height:2; }
.no-stat { color:#999; font-size:12px; }
</style>
