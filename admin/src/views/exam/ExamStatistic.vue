<template>
  <div>
    <el-card>
      <div class="header">
        <el-button @click="goBack">‹ 返回</el-button>
        <h3 style="margin:0 0 0 12px;display:inline-block">统计: {{ examTitle }}</h3>
        <el-button size="small" style="margin-left:auto" @click="goResponses">查看记录</el-button>
      </div>

      <el-row :gutter="16" style="margin-top:16px">
        <el-col :span="6">
          <el-card shadow="never" class="stat-card">
            <div class="stat-num">{{ stat.total }}</div>
            <div class="stat-label">总考试人次</div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card shadow="never" class="stat-card">
            <div class="stat-num primary">{{ stat.submitted }}</div>
            <div class="stat-label">已提交</div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card shadow="never" class="stat-card">
            <div class="stat-num success">{{ stat.passed }}</div>
            <div class="stat-label">通过</div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card shadow="never" class="stat-card">
            <div class="stat-num warning">{{ stat.passRate ? (stat.passRate * 100).toFixed(1) + '%' : '-' }}</div>
            <div class="stat-label">通过率</div>
          </el-card>
        </el-col>
      </el-row>

      <el-row :gutter="16" style="margin-top:16px">
        <el-col :span="14">
          <el-card shadow="never">
            <div class="chart-title">近 7 天趋势</div>
            <v-chart :option="dailyOption" style="height:300px" autoresize />
          </el-card>
        </el-col>
        <el-col :span="10">
          <el-card shadow="never">
            <div class="chart-title">分数分布</div>
            <v-chart :option="scoreDistOption" style="height:300px" autoresize />
          </el-card>
        </el-col>
      </el-row>

      <el-card shadow="never" style="margin-top:16px">
        <div class="chart-title">每题分析</div>
        <div v-for="fs in stat.fieldStats" :key="fs.questionId" style="margin-top:16px;padding-top:16px;border-top:1px solid #f0f0f0">
          <div class="field-title">{{ truncate(stripHtml(fs.title), 50) }} <el-tag size="small" type="info">{{ fs.type }}</el-tag></div>
          <el-row :gutter="16" style="margin-top:8px">
            <el-col :span="6">
              <div class="field-meta">作答: {{ fs.nonEmpty }} / {{ fs.totalCount }} ({{ fs.empty }} 空)</div>
            </el-col>
            <el-col :span="6">
              <div class="field-meta">正确率: {{ fs.correctRate ? (fs.correctRate * 100).toFixed(1) + '%' : '-' }}</div>
            </el-col>
          </el-row>
          <div v-if="fs.dist && Object.keys(fs.dist).length > 0" style="margin-top:8px">
            <v-chart :option="distOption(fs)" style="height:200px" autoresize />
          </div>
          <div v-else-if="fs.tableData?.length" style="margin-top:8px;max-height:300px;overflow:auto">
            <el-table :data="fs.tableData" border size="small" style="width:100%">
              <el-table-column type="index" label="#" width="40" fixed />
              <el-table-column v-for="(col, ci) in fs.tableCols" :key="ci" :label="col" min-width="100">
                <template #default="{ row }">{{ stripHtml(row[ci] ?? '') }}</template>
              </el-table-column>
            </el-table>
          </div>
        </div>
        <div v-if="!stat.fieldStats || stat.fieldStats.length === 0" style="color:#aaa;text-align:center;padding:40px 0">暂无数据</div>
      </el-card>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { adminApi } from '../../api'
import { showRequestError } from '../../utils/request'
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart, PieChart, BarChart } from 'echarts/charts'
import { GridComponent, TooltipComponent, LegendComponent } from 'echarts/components'

use([CanvasRenderer, LineChart, PieChart, BarChart, GridComponent, TooltipComponent, LegendComponent])

const route = useRoute()
const router = useRouter()
const examId = Number(route.query.examId || 0)
const examTitle = String(route.query.title || `考试 #${examId}`)
const stat = ref<any>({})

async function load() {
  if (!examId) return
  try {
    const res: any = await adminApi.examStatistics(examId)
    stat.value = res.data || res
  } catch (error) { showRequestError(error, '加载失败') }
}

const dailyOption = computed(() => {
  const daily = stat.value.daily || []
  return {
    tooltip: { trigger: 'axis' },
    grid: { left: 40, right: 20, bottom: 30, top: 20 },
    xAxis: { type: 'category', data: daily.map((d: any) => d.date) },
    yAxis: { type: 'value', minInterval: 1 },
    series: [{ type: 'line', data: daily.map((d: any) => d.count), smooth: true, areaStyle: { opacity: 0.15 }, lineStyle: { color: '#fb454c' }, itemStyle: { color: '#fb454c' } }]
  }
})

const scoreDistOption = computed(() => {
  const dist = stat.value.scoreDist || {}
  const keys = Object.keys(dist).sort((a: any, b: any) => Number(a) - Number(b))
  return {
    tooltip: { trigger: 'axis' },
    grid: { left: 40, right: 20, bottom: 30, top: 20 },
    xAxis: { type: 'category', data: keys },
    yAxis: { type: 'value', minInterval: 1 },
    series: [{ type: 'bar', data: keys.map((k) => dist[k]), itemStyle: { color: '#409eff' } }]
  }
})

function distOption(fs: any) {
  const dist = fs.dist || {}
  const entries = Object.entries(dist).sort((a: any, b: any) => b[1] - a[1])
  const maxVal = Math.max(...entries.map((e: any) => e[1]), 1)
  const axisMax = maxVal + Math.ceil(maxVal * 0.6) + 1
  return {
    tooltip: { trigger: 'axis' },
    grid: { left: 120, right: 30, bottom: 30, top: 10 },
    xAxis: { type: 'value', minInterval: 1, max: axisMax, splitLine: { show: false }, axisTick: { show: false }, axisLabel: { showMinLabel: true, formatter: (v: number) => v > maxVal ? '' : String(v) } },
    yAxis: { type: 'category', data: entries.map((e: any) => stripHtml(e[0] ?? '')) },
    series: [{ type: 'bar', data: entries.map((e: any) => e[1]), itemStyle: { color: '#409eff' } }]
  }
}

function goResponses() {
  router.push({ path: '/exam/responses', query: { examId: String(examId), title: examTitle } })
}

function goBack() { router.push('/exam/list') }

function stripHtml(html: string) {
  return html ? html.replace(/<[^>]+>/g, '') : ''
}
function truncate(s: string, len: number) {
  return s.length > len ? s.slice(0, len) + '…' : s
}

onMounted(load)
</script>

<style scoped>
.header { display:flex; align-items:center; }
.stat-card { text-align:center; padding:8px 0; }
.stat-num { font-size:36px; font-weight:bold; color:#333; }
.stat-num.primary { color:#409eff; }
.stat-num.success { color:#67c23a; }
.stat-num.warning { color:#e6a23c; }
.stat-label { font-size:13px; color:#888; margin-top:4px; }
.chart-title { font-size:15px; font-weight:500; margin-bottom:8px; color:#333; }
.field-title { font-size:14px; font-weight:500; margin-bottom:4px; }
.field-meta { font-size:12px; color:#888; }
</style>
