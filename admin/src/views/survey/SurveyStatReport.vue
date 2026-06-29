<template>
  <div>
    <el-card>
      <div class="header">
        <el-button @click="goBack">‹ 返回</el-button>
        <h3 style="margin:0 0 0 12px;display:inline-block">统计报表: {{ surveyTitle }}</h3>
        <div style="margin-left:auto;display:flex;gap:8px;align-items:center">
          <el-radio-group v-if="showModeSwitch" :model-value="viewMode" size="small" @change="switchMode">
            <el-radio-button value="aggregate">总览</el-radio-button>
            <el-radio-button value="perResponse">逐份</el-radio-button>
          </el-radio-group>
          <el-button size="small" @click="exportCSV">导出 CSV</el-button>
          <el-button size="small" type="primary" @click="load">刷新</el-button>
        </div>
      </div>

      <el-alert v-if="configInfo" :title="configInfo" type="info" show-icon :closable="false" style="margin-top:12px" />

      <!-- ========== 总览模式 ========== -->
      <template v-if="viewMode === 'aggregate'">
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
              <div class="stat-num success">{{ deviceMobile }}</div>
              <div class="stat-label">移动端</div>
            </el-card>
          </el-col>
          <el-col :span="6">
            <el-card shadow="never" class="stat-card">
              <div class="stat-num warning">{{ devicePC }}</div>
              <div class="stat-label">PC 端</div>
            </el-card>
          </el-col>
        </el-row>

        <el-row :gutter="16" style="margin-top:16px">
          <el-col :span="14">
            <el-card shadow="never">
              <div class="chart-title">近 7 天趋势</div>
              <v-chart :option="dailyOption" style="height:280px" autoresize />
            </el-card>
          </el-col>
          <el-col :span="10">
            <el-card shadow="never">
              <div class="chart-title">设备分布</div>
              <v-chart :option="deviceOption" style="height:280px" autoresize />
            </el-card>
          </el-col>
        </el-row>

        <el-card shadow="never" style="margin-top:16px">
          <div class="chart-title">每题分析</div>
          <div v-if="!stat.fieldStats || stat.fieldStats.length === 0" style="color:#aaa;text-align:center;padding:40px 0">暂无数据</div>
          <div v-for="fs in filteredStats" :key="fs.questionId" class="field-section">
            <div class="field-title">
              <span class="field-label">{{ truncate(stripHtml(fs.title), 60) }}</span>
              <el-tag size="small" :type="tagType(fs.type)" class="q-type-tag">{{ fs.type }}</el-tag>
              <el-tag v-if="statMode !== 'count'" size="small" :type="statMode === 'value' ? 'warning' : 'success'" class="mode-tag">{{ statMode === 'value' ? '按值' : '按分' }}</el-tag>
            </div>
            <div class="field-meta-bar">填写: <b>{{ fs.nonEmpty }}</b> / {{ fs.totalCount }} ({{ fs.empty }} 空)</div>

            <div v-if="fs.numericStat" class="numeric-stats">
              <span>总和 {{ fs.numericStat.sum }}</span>
              <span>平均 {{ fs.numericStat.avg.toFixed(2) }}</span>
              <span>最小 {{ fs.numericStat.min }}</span>
              <span>最大 {{ fs.numericStat.max }}</span>
            </div>

            <div v-if="isChoiceType(fs.type)" style="margin-top:8px">
              <div v-if="showChart" class="chart-wrapper">
                <v-chart :option="distOption(fs)" style="height:220px" autoresize />
              </div>
              <div v-if="showDetail" class="detail-table-wrapper">
                <el-table :data="detailRows(fs)" border size="small" style="max-width:500px" :show-header="true">
                  <el-table-column label="选项" min-width="200">
                    <template #default="{ row }">{{ row.label }}</template>
                  </el-table-column>
                  <el-table-column label="次数" width="100" align="center">
                    <template #default="{ row }">{{ row.count }}</template>
                  </el-table-column>
                  <el-table-column label="占比" width="100" align="center">
                    <template #default="{ row }">{{ row.pct }}%</template>
                  </el-table-column>
                </el-table>
              </div>
              <div v-if="!showChart && !showDetail" style="color:#999;font-size:13px;margin-top:8px">
                当前配置隐藏图表和明细，请在配置面板中调整
              </div>
            </div>

            <div v-else-if="fs.tableData?.length" style="margin-top:8px;max-height:300px;overflow:auto">
              <el-table :data="buildMatrixRows(fs)" border size="small" style="width:100%">
                <el-table-column type="index" label="#" width="40" fixed />
                <el-table-column v-for="(col, ci) in fs.tableCols" :key="ci" :label="col" min-width="100">
                  <template #default="{ row }">{{ stripHtml(row[ci] ?? '') }}</template>
                </el-table-column>
              </el-table>
            </div>
          </div>
        </el-card>
      </template>

      <!-- ========== 逐份模式 ========== -->
      <template v-if="viewMode === 'perResponse'">
        <div class="mode-bar">
          <el-input v-model="respKeyword" placeholder="搜索提交人" clearable size="small" style="width:200px" @keyup.enter="loadResponses" />
          <span style="font-size:13px;color:#888;margin-left:12px">共 {{ responses.length }} 份答卷</span>
        </div>

        <div v-if="responses.length === 0" style="color:#aaa;text-align:center;padding:40px 0">暂无答卷数据</div>

        <div v-for="(resp, ri) in responses" :key="resp.id || ri" class="resp-card">
          <div class="resp-header" @click="toggleResp(ri)">
            <span class="resp-index">#{{ ri + 1 }}</span>
            <span class="resp-user">{{ resp.nickname || resp.userName || '匿名' }}</span>
            <span class="resp-meta">{{ resp.addTime ? formatTime(resp.addTime) : '' }}</span>
            <span v-if="resp.answerTime" class="resp-meta">用时 {{ (resp.answerTime / 1000).toFixed(1) }}s</span>
            <span class="resp-toggle">{{ expanded[ri] ? '收起' : '展开' }}</span>
          </div>
          <div v-show="expanded[ri]" class="resp-body">
            <div v-for="(ans, ai) in parseAnswers(resp)" :key="ai" class="resp-answer-row">
              <div class="resp-q-title">{{ ai + 1 }}. {{ stripHtml(ans.title) }}</div>
              <div class="resp-q-answer">{{ ans.display }}</div>
            </div>
            <div v-if="!parseAnswers(resp).length" style="color:#999;font-size:13px;padding:8px">此份答卷无有效答案</div>
          </div>
        </div>
      </template>
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
const surveyTitle = ref('')
const stat = ref<any>({})
const questions = ref<any[]>([])
const settings = ref<any>({})

const statMode = ref('count')
const viewMode = ref('aggregate')
const statRange = ref('all')
const showChart = ref(true)
const showDetail = ref(true)
const exportField = ref('label')

const showModeSwitch = ref(true)

const skipTypes = ['divider', 'description', 'pagination', 'questionSet']
const CHOICE_TYPES = ['radio', 'checkbox', 'select', 'picker', 'matrixRadio', 'matrixCheckbox', 'user', 'dept']
const SCORE_TYPES = ['score', 'rating', 'nps']

const configInfo = computed(() => {
  const modeLabel: Record<string, string> = { value: '按选项值', count: '按计数(标签)', score: '按分值' }
  const viewLabel: Record<string, string> = { aggregate: '总览统计', perResponse: '逐份统计' }
  const mode = modeLabel[statMode.value] || '按计数(标签)'
  const view = viewLabel[viewMode.value] || '总览统计'
  let info = `查看模式：${view} ｜ 统计方式：${mode}`
  if (viewMode.value === 'aggregate') {
    info += ` ｜ ${showChart.value ? '图表' : ''}${showChart.value && showDetail.value ? '+' : ''}${showDetail.value ? '明细' : ''} ｜ 导出：${exportField.value === 'value' ? '选项值' : exportField.value === 'both' ? '值+标签' : '选项标签'}`
  }
  return info
})

const filteredStats = computed(() => {
  if (!stat.value.fieldStats) return []
  let list = stat.value.fieldStats
  if (viewMode.value !== 'aggregate') return list
  if (statRange.value === 'choice') {
    list = list.filter((fs: any) => CHOICE_TYPES.includes(fs.type))
  } else if (statRange.value === 'score') {
    list = list.filter((fs: any) => SCORE_TYPES.includes(fs.type))
  }
  return list
})

const deviceMobile = computed(() => stat.value.deviceStat?.mobile || 0)
const devicePC = computed(() => stat.value.deviceStat?.pc || 0)

// 逐份模式
const responses = ref<any[]>([])
const expanded = ref<Record<number, boolean>>({})
const respKeyword = ref('')

async function loadResponses() {
  try {
    const res: any = await adminApi.surveyResponseList({ surveyId, page: 1, pageSize: 500, keyword: respKeyword.value })
    const data = res.data || res
    responses.value = (data.list || data.rows || data.data || []).filter((r: any) => r.forms)
    expanded.value = {}
  } catch { ElMessage.error('加载答卷失败') }
}

function toggleResp(i: number) {
  expanded.value[i] = !expanded.value[i]
}

function parseAnswers(resp: any) {
  const forms = typeof resp.forms === 'string' ? JSON.parse(resp.forms) : (resp.forms || {})
  const result: { title: string; display: string }[] = []
  for (const q of questions.value) {
    const val = forms[q.id] ?? forms[q.key] ?? forms[q.id.toString()]
    if (val === undefined || val === null || val === '') continue
    const display = formatAnswerValue(q, val)
    result.push({ title: q.title || q.label || '', display })
  }
  return result
}

function formatAnswerValue(q: any, val: any): string {
  if (q.type === 'checkbox' && Array.isArray(val)) {
    return val.map((v: string) => {
      const opt = findOption(q, v)
      return opt ? stripHtml(opt.label) : v
    }).join('、')
  }
  if (['radio', 'select', 'picker'].includes(q.type)) {
    const opt = findOption(q, val)
    return opt ? stripHtml(opt.label) : String(val)
  }
  if (q.type === 'matrixRadio' && typeof val === 'object') {
    return Object.entries(val).map(([k, v]) => `${k}: ${v}`).join('; ')
  }
  if (q.type === 'matrixCheckbox' && typeof val === 'object') {
    return Object.entries(val).map(([k, v]) => `${k}: [${(v as string[]).join(', ')}]`).join('; ')
  }
  if (q.type === 'user' || q.type === 'dept') {
    const opt = findOption(q, val)
    return opt ? stripHtml(opt.label) : String(val)
  }
  if (typeof val === 'object') return JSON.stringify(val)
  return stripHtml(String(val))
}

function findOption(q: any, val: string): any {
  if (!q.props?.options) return null
  return q.props.options.find((opt: any) => String(opt.value ?? opt.id) === String(val) || String(opt.label) === String(val))
}

function switchMode(mode: string) {
  viewMode.value = mode
  if (mode === 'perResponse' && responses.value.length === 0) {
    loadResponses()
  }
}

function tagType(type: string) {
  if (CHOICE_TYPES.includes(type)) return ''
  if (SCORE_TYPES.includes(type)) return 'success'
  return 'info'
}

function isChoiceType(type: string) {
  return CHOICE_TYPES.includes(type)
}

async function load() {
  if (!surveyId) return
  try {
    const [statRes, detailRes]: any = await Promise.all([
      adminApi.surveyStatistic(surveyId),
      adminApi.surveyDetail(surveyId)
    ])
    stat.value = statRes.data || statRes
    surveyTitle.value = String(route.query.title || statRes.data?.survey?.title || `问卷 #${surveyId}`)

    const detail = detailRes.data || detailRes
    const sv = detail?.survey || detail?.data?.survey || {}
    const raw = detail?.schema || detail?.data?.schema
    const sch = raw ? (typeof raw === 'string' ? JSON.parse(raw) : raw) : { questions: [] }
    questions.value = (sch.questions || []).filter((q: any) => !skipTypes.includes(q.type))

    if (sv?.settings) {
      settings.value = typeof sv.settings === 'string' ? JSON.parse(sv.settings) : sv.settings
      const rc = settings.value?.resultConfig || {}
      statMode.value = rc.statType || 'count'
      viewMode.value = rc.viewMode || 'aggregate'
      statRange.value = (rc.scope || ['all'])[0] || 'all'
      showChart.value = rc.showChart !== false
      showDetail.value = rc.showDetail !== false
      exportField.value = rc.exportField || 'label'
    }

    if (viewMode.value === 'perResponse') {
      await loadResponses()
    }
  } catch (e) {
    ElMessage.error('加载失败')
  }
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

function distOption(fs: any) {
  const dist = fs.dist || {}
  const entries = Object.entries(dist).sort((a: any, b: any) => b[1] - a[1])
  const q = questions.value.find((x: any) => x.id === fs.questionId)
  const labelMap: Record<string, string> = {}
  if (q && q.props?.options && ['radio', 'select', 'picker', 'checkbox', 'matrixRadio', 'matrixCheckbox', 'user', 'dept'].includes(q.type)) {
    for (const opt of q.props.options) {
      const val = String(opt.value ?? opt.label ?? opt)
      const lbl = String(opt.label ?? opt.value ?? opt)
      labelMap[val] = lbl
      if (!(val in dist)) {
        entries.push([val, 0])
      }
    }
  }
  const maxVal = Math.max(...entries.map((e: any) => e[1]), 1)
  const axisMax = maxVal + Math.ceil(maxVal * 0.6) + 1

  const displayKey = (k: string) => {
    if (statMode.value === 'value' && labelMap[k]) {
      if (exportField.value === 'value') return k
      if (exportField.value === 'both') return `${k} (${labelMap[k]})`
      return `${labelMap[k]} [${k}]`
    }
    if (exportField.value === 'value') return k
    return stripHtml(labelMap[k] ?? k)
  }

  return {
    tooltip: { trigger: 'axis', formatter: (params: any) => {
      const p = params[0]
      return `${p.name}: ${p.value}`
    } },
    grid: { left: 140, right: 30, bottom: 30, top: 10 },
    xAxis: { type: 'value', minInterval: 1, max: axisMax, splitLine: { show: false }, axisTick: { show: false }, axisLabel: { showMinLabel: true, formatter: (v: number) => v > maxVal ? '' : String(v) } },
    yAxis: { type: 'category', data: entries.map((e: any) => displayKey(e[0])), axisLabel: { fontSize: 12 } },
    series: [{ type: 'bar', data: entries.map((e: any) => e[1]), itemStyle: { color: '#409eff' }, label: { show: true, position: 'right', formatter: (p: any) => p.value > 0 ? String(p.value) : '' } }]
  }
}

function detailRows(fs: any) {
  const dist = fs.dist || {}
  const entries = Object.entries(dist).sort((a: any, b: any) => b[1] - a[1])
  const total = entries.reduce((s: number, e: any) => s + e[1], 0) || 1
  const q = questions.value.find((x: any) => x.id === fs.questionId)
  const labelMap: Record<string, string> = {}
  if (q && q.props?.options) {
    for (const opt of q.props.options) {
      const val = String(opt.value ?? opt.label ?? opt)
      const lbl = String(opt.label ?? opt.value ?? opt)
      labelMap[val] = lbl
    }
  }
  return entries.map(([k, v]: [string, number]) => {
    let label = stripHtml(labelMap[k] ?? k)
    if (statMode.value === 'value' && labelMap[k]) {
      if (exportField.value === 'value') label = k
      else if (exportField.value === 'both') label = `${k} (${labelMap[k]})`
    } else if (statMode.value === 'value') {
      label = k
    }
    return { label, count: v, pct: ((v / total) * 100).toFixed(1) }
  })
}

function buildMatrixRows(fs: any) {
  return fs.tableData || []
}

function exportCSV() {
  router.push({ path: '/survey/responses', query: { surveyId: String(surveyId), title: surveyTitle.value } })
}

function goBack() { router.push('/survey') }

function stripHtml(html: string) {
  return html ? html.replace(/<[^>]+>/g, '') : ''
}

function truncate(s: string, len: number) {
  return s.length > len ? s.slice(0, len) + '…' : s
}

function formatTime(ms: number) {
  if (!ms) return '-'
  return new Date(ms).toLocaleString()
}

onMounted(() => { load() })
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
.field-section { margin-top:16px; padding-top:16px; border-top:1px solid #f0f0f0; }
.field-title { font-size:14px; font-weight:500; margin-bottom:4px; display:flex; align-items:center; gap:6px; }
.field-label { max-width:500px; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.field-meta-bar { font-size:12px; color:#888; margin-bottom:6px; }
.q-type-tag { flex-shrink:0; }
.mode-tag { flex-shrink:0; }
.numeric-stats { font-size:12px; color:#666; display:flex; gap:16px; margin-top:4px; }
.chart-wrapper { margin-top:4px; }
.detail-table-wrapper { margin-top:8px; }
.mode-bar { display:flex; align-items:center; margin-bottom:12px; }
.resp-card { border:1px solid #e8e8e8; border-radius:6px; margin-bottom:8px; overflow:hidden; }
.resp-header { display:flex; align-items:center; gap:12px; padding:10px 14px; background:#fafafa; cursor:pointer; user-select:none; }
.resp-header:hover { background:#f0f0f0; }
.resp-index { font-weight:600; color:#333; min-width:40px; }
.resp-user { font-weight:500; color:#333; }
.resp-meta { font-size:12px; color:#999; }
.resp-toggle { margin-left:auto; font-size:12px; color:#409eff; }
.resp-body { padding:8px 14px 14px; border-top:1px solid #eee; }
.resp-answer-row { display:flex; gap:12px; padding:6px 0; border-bottom:1px solid #f5f5f5; }
.resp-answer-row:last-child { border-bottom:none; }
.resp-q-title { flex:1; font-size:13px; color:#333; min-width:0; }
.resp-q-answer { flex:1; font-size:13px; color:#666; }
</style>
