<template>
  <div>
    <el-card>
      <div class="header">
        <el-button @click="goBack">‹ 返回</el-button>
        <h3 style="margin:0 0 0 12px;display:inline-block">统计: {{ surveyTitle }}</h3>
        <el-button size="small" style="margin-left:auto" @click="exportData">导出 CSV</el-button>
      </div>

      <el-row :gutter="16" style="margin-top:16px">
        <el-col :span="6">
          <el-card shadow="never" class="stat-card">
            <div class="stat-num">{{ stat.total }}</div>
            <div class="stat-label">总答卷数</div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card shadow="never" class="stat-card">
            <div class="stat-num primary">{{ stat.todayCount }}</div>
            <div class="stat-label">今日新增</div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card shadow="never" class="stat-card">
            <div class="stat-num success">{{ stat.deviceStat?.mobile || 0 }}</div>
            <div class="stat-label">移动端</div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card shadow="never" class="stat-card">
            <div class="stat-num warning">{{ stat.deviceStat?.pc || 0 }}</div>
            <div class="stat-label">PC 端</div>
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
            <div class="chart-title">设备分布</div>
            <v-chart :option="deviceOption" style="height:300px" autoresize />
          </el-card>
        </el-col>
      </el-row>

      <el-card shadow="never" style="margin-top:16px">
        <div class="chart-title">每题分析</div>
        <div v-for="fs in stat.fieldStats" :key="fs.questionId" style="margin-top:16px;padding-top:16px;border-top:1px solid #f0f0f0">
          <div class="field-title">{{ fs.title }} <el-tag size="small" type="info">{{ fs.type }}</el-tag></div>
          <el-row :gutter="16" style="margin-top:8px">
            <el-col :span="6">
              <div class="field-meta">填写: {{ fs.nonEmpty }} / {{ fs.totalCount }} ({{ fs.empty }} 空)</div>
            </el-col>
          </el-row>
          <div v-if="fs.numericStat">
            <div class="field-meta">总和 {{ fs.numericStat.sum }} | 平均 {{ fs.numericStat.avg.toFixed(2) }} | 最小 {{ fs.numericStat.min }} | 最大 {{ fs.numericStat.max }}</div>
          </div>
          <div v-if="fs.dist && Object.keys(fs.dist).length > 0" style="margin-top:8px">
            <v-chart :option="distOption(fs.dist)" style="height:200px" autoresize />
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
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart, PieChart, BarChart } from 'echarts/charts'
import { GridComponent, TooltipComponent, LegendComponent } from 'echarts/components'

use([CanvasRenderer, LineChart, PieChart, BarChart, GridComponent, TooltipComponent, LegendComponent])

const route = useRoute()
const router = useRouter()
const surveyId = Number(route.query.surveyId || 0)
const surveyTitle = String(route.query.title || `问卷 #${surveyId}`)
const stat = ref<any>({})

async function load() {
  if (!surveyId) return
  try {
    const res: any = await adminApi.surveyStatistic(surveyId)
    stat.value = res.data || res
  } catch { ElMessage.error('加载失败') }
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

const deviceOption = computed(() => {
  const ds = stat.value.deviceStat || {}
  return {
    tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)' },
    series: [{
      type: 'pie', radius: ['40%', '70%'], center: ['50%', '55%'],
      data: [
        { name: '移动端', value: ds.mobile || 0, itemStyle: { color: '#409eff' } },
        { name: 'PC 端', value: ds.pc || 0, itemStyle: { color: '#67c23a' } }
      ],
      label: { show: true, formatter: '{b}\n{d}%' }
    }]
  }
})

function distOption(dist: Record<string, number>) {
  const entries = Object.entries(dist).sort((a, b) => b[1] - a[1])
  return {
    tooltip: { trigger: 'axis' },
    grid: { left: 120, right: 30, bottom: 30, top: 10 },
    xAxis: { type: 'value' },
    yAxis: { type: 'category', data: entries.map(e => e[0]) },
    series: [{ type: 'bar', data: entries.map(e => e[1]), itemStyle: { color: '#409eff' } }]
  }
}

function exportData() {
  router.push({ path: '/survey/responses', query: { surveyId: String(surveyId), title: surveyTitle } })
}

function goBack() { router.push('/survey') }

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
