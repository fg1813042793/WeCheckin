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
          <div class="field-title">{{ truncate(stripHtml(fs.title), 50) }} <el-tag size="small" type="info">{{ fs.type }}</el-tag></div>
          <el-row :gutter="16" style="margin-top:8px">
            <el-col :span="6">
              <div class="field-meta">填写: {{ fs.nonEmpty }} / {{ fs.totalCount }} ({{ fs.empty }} 空)</div>
            </el-col>
          </el-row>
          <div v-if="fs.numericStat">
            <div class="field-meta">总和 {{ fs.numericStat.sum }} | 平均 {{ fs.numericStat.avg.toFixed(2) }} | 最小 {{ fs.numericStat.min }} | 最大 {{ fs.numericStat.max }}</div>
          </div>
          <div v-if="fs.dist && Object.keys(fs.dist).length > 0 && !['matrixFillBlank','matrixAuto','file','input','textarea','text','phone','email','idCard','password','multiInput','hInput','location','date','time','dateRange','richText','autopop','signature','name','studentId','employeeId','class'].includes(fs.type)" style="margin-top:8px">
            <v-chart :option="distOption(fs)" style="height:200px" autoresize />
          </div>
          <div v-else-if="fs.type === 'location' && fs.tableData?.length" style="margin-top:8px">
            <div ref="mapRef" class="stat-map" :id="'location-map-' + fs.questionId" style="height:300px;border-radius:6px"></div>
          </div>
          <div v-else-if="fs.tableData?.length" style="margin-top:8px;max-height:300px;overflow:auto">
            <el-table :data="buildMatrixTableRows(fs)" border size="small" style="width:100%">
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
const questions = ref<any[]>([])
const skipTypes = ['divider', 'description', 'pagination', 'questionSet']

async function load() {
  if (!surveyId) return
  try {
    const [statRes, detailRes]: any = await Promise.all([
      adminApi.surveyStatistic(surveyId),
      adminApi.surveyDetail(surveyId)
    ])
    stat.value = statRes.data || statRes
    const detail = detailRes.data || detailRes
    const raw = detail?.schema
    const sch = raw ? (typeof raw === 'string' ? JSON.parse(raw) : raw) : { questions: [] }
    questions.value = (sch.questions || []).filter((q: any) => !skipTypes.includes(q.type))
  } catch { ElMessage.error('加载失败') }
}

const dailyOption = computed(() => {
  const daily = stat.value.daily || []
  return {
    tooltip: { trigger: 'axis' },
    grid: { left: 40, right: 20, bottom: 30, top: 20 },
    xAxis: { type: 'category', data: daily.map((d: any) => d.date) },
    yAxis: { type: 'value', minInterval: 1 },
    series: [{ type: 'line', data: daily.map((d: any) => d.count), smooth: true, areaStyle: { opacity: 0.15, color: '#2563eb' }, lineStyle: { color: '#2563eb' }, itemStyle: { color: '#2563eb' } }]
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

function distOption(fs: any) {
  const dist = fs.dist || {}
  const entries = Object.entries(dist).sort((a: any, b: any) => b[1] - a[1])
  const q = questions.value.find((x: any) => x.id === fs.questionId)
  const labelMap: Record<string, string> = {}
  if (q && q.props?.options && ['radio','select','picker','checkbox','matrixRadio','matrixCheckbox','user','dept'].includes(q.type)) {
    for (const opt of q.props.options) {
      const val = opt.value ?? opt.label ?? opt
      const lbl = opt.label ?? opt.value ?? opt
      labelMap[String(val)] = String(lbl)
      if (!(val in dist)) {
        entries.push([String(val), 0])
      }
    }
  }
  const maxVal = Math.max(...entries.map((e: any) => e[1]), 1)
  const axisMax = maxVal + Math.ceil(maxVal * 0.6) + 1
  return {
    tooltip: { trigger: 'axis' },
    grid: { left: 120, right: 30, bottom: 30, top: 10 },
    xAxis: { type: 'value', minInterval: 1, max: axisMax, splitLine: { show: false }, axisTick: { show: false }, axisLabel: { showMinLabel: true, formatter: (v: number) => v > maxVal ? '' : String(v) } },
    yAxis: { type: 'category', data: entries.map((e: any) => stripHtml(labelMap[e[0]] ?? e[0])) },
    series: [{ type: 'bar', data: entries.map((e: any) => e[1]), itemStyle: { color: '#409eff' } }]
  }
  
}

function exportData() {
  router.push({ path: '/survey/responses', query: { surveyId: String(surveyId), title: surveyTitle } })
}

function goBack() { router.push('/survey') }

function formatTime(ms: number) {
  if (!ms) return '-'
  return new Date(ms).toLocaleString()
}
function buildMatrixTableRows(fs: any) {
  return fs.tableData || []
}
function stripHtml(html: string) {
  return html ? html.replace(/<[^>]+>/g, '') : ''
}
function truncate(s: string, len: number) {
  return s.length > len ? s.slice(0, len) + '…' : s
}

function initLocationMaps() {
  import('leaflet').then(L => {
    delete (L as any).Icon.Default.prototype._getIconUrl
    ;(L as any).Icon.Default.mergeOptions({
      iconRetinaUrl: 'https://cdn.jsdelivr.net/npm/leaflet@1.9.4/dist/images/marker-icon-2x.png',
      iconUrl: 'https://cdn.jsdelivr.net/npm/leaflet@1.9.4/dist/images/marker-icon.png',
      shadowUrl: 'https://cdn.jsdelivr.net/npm/leaflet@1.9.4/dist/images/marker-shadow.png'
    })
    const fss = stat.value.fieldStats || []
    for (const fs of fss) {
      if (fs.type !== 'location' || !fs.tableData?.length) continue
      const containerId = 'location-map-' + fs.questionId
      const el = document.getElementById(containerId)
      if (!el) continue
      const coords: [number, number][] = []
      for (const row of fs.tableData) {
        const s = row[0]
        if (!s) continue
        const parts = s.split(',')
        if (parts.length < 2) continue
        const lat = parseFloat(parts[0])
        const lng = parseFloat(parts[1])
        if (!isNaN(lat) && !isNaN(lng)) coords.push([lat, lng])
      }
      if (coords.length === 0) continue
      const map = L.map(containerId).setView(coords[0], 13)
      L.tileLayer('https://webrd01.is.autonavi.com/appmaptile?lang=zh_cn&size=1&scale=1&style=8&x={x}&y={y}&z={z}', {
        attribution: '&copy; 高德地图',
        maxZoom: 18
      }).addTo(map)
      for (const c of coords) {
        L.marker(c).addTo(map)
      }
    }
  })
}
const locationMapKey = ref(0)
function onLocationMapMounted(_el: any, fs: any) {
  setTimeout(() => {
    initLocationMaps()
  }, 100)
}

onMounted(() => {
  load()
  setTimeout(() => initLocationMaps(), 300)
})
</script>

<style>
@import 'leaflet/dist/leaflet.css';
</style>
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
:deep(.stat-map) { width:100%; z-index:0; }
</style>
