<template>
  <div>
    <el-card>
      <div class="header">
        <el-button @click="goBack">‹ 返回</el-button>
        <h3 style="margin:0 0 0 12px;display:inline-block">答卷: {{ surveyTitle }}</h3>
        <el-tag v-if="surveyId" type="info" size="small" style="margin-left:8px">SurveyID: {{ surveyId }}</el-tag>
        <el-button size="small" style="margin-left:auto" @click="exportCSV">导出 CSV</el-button>
      </div>

      <el-table :data="list" v-loading="loading" stripe style="margin-top:16px" border>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="userId" label="用户" width="120" />
        <el-table-column prop="nickname" label="昵称" width="140" />
        <el-table-column label="状态" width="70">
          <template #default="{ row }">
            <el-tag :type="row.status===1 ? 'success' : 'info'" size="small">
              {{ row.status===1 ? '完成' : '草稿' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="用时" width="80">
          <template #default="{ row }">{{ formatDuration(row.duration) }}</template>
        </el-table-column>
        <el-table-column v-for="q in questions" :key="q.id" :label="q.title" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">{{ formatVal(row.answers?.[q.id]) }}</template>
        </el-table-column>
        <el-table-column label="设备" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">{{ row.device || '-' }}</template>
        </el-table-column>
        <el-table-column label="提交时间" min-width="150">
          <template #default="{ row }">{{ formatTime(row.submitTime) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="viewDetail(row)">查看</el-button>
            <el-button size="small" type="danger" @click="del(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div style="text-align:center;margin-top:16px">
        <el-pagination v-model:current-page="page" :page-size="pageSize" :total="total" layout="total,prev,pager,next" @current-change="load" />
      </div>
    </el-card>

    <el-dialog v-model="detailDialog.visible" :title="`答卷详情 #${detailDialog.response?.id}`" width="800px">
      <el-descriptions :column="2" border size="small" style="margin-bottom:12px">
        <el-descriptions-item label="用户">{{ detailDialog.response?.userId || '匿名' }}</el-descriptions-item>
        <el-descriptions-item label="用时">{{ formatDuration(detailDialog.response?.duration) }}</el-descriptions-item>
        <el-descriptions-item label="设备">{{ detailDialog.response?.device || '-' }}</el-descriptions-item>
        <el-descriptions-item label="IP">{{ detailDialog.response?.ip || '-' }}</el-descriptions-item>
        <el-descriptions-item label="开始">{{ formatTime(detailDialog.response?.startTime) }}</el-descriptions-item>
        <el-descriptions-item label="提交">{{ formatTime(detailDialog.response?.submitTime) }}</el-descriptions-item>
      </el-descriptions>
      <el-table :data="answerRows" stripe size="small" border>
        <el-table-column prop="questionId" label="题号" width="80" />
        <el-table-column prop="title" label="题目" min-width="200" show-overflow-tooltip />
        <el-table-column prop="value" label="答案" min-width="200" />
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { adminApi } from '../../api'

const route = useRoute()
const router = useRouter()
const surveyId = Number(route.query.surveyId || 0)
const surveyTitle = String(route.query.title || `问卷 #${surveyId}`)

const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref<any[]>([])
const loading = ref(false)
const questions = ref<any[]>([])
const skipTypes = ['divider', 'description', 'richText', 'pagination', 'questionSet']

function formatTime(ms: number) {
  if (!ms) return '-'
  return new Date(ms).toLocaleString()
}
function formatDuration(sec: number) {
  if (!sec) return '-'
  const m = Math.floor(sec / 60)
  const s = sec % 60
  return m > 0 ? `${m}分${s}秒` : `${s}秒`
}

async function load() {
  if (!surveyId) { ElMessage.error('缺少 surveyId'); return }
  loading.value = true
  try {
    const [res, detailRes]: any = await Promise.all([
      adminApi.surveyResponseList({ surveyId, page: page.value, pageSize: pageSize.value }),
      adminApi.surveyDetail(surveyId)
    ])
    list.value = res.data?.list || res.list || []
    total.value = res.data?.total || res.total || 0
    const detail = (detailRes as any).data || detailRes
    const raw = detail?.schema
    const sch = raw ? (typeof raw === 'string' ? JSON.parse(raw) : raw) : { questions: [] }
    questions.value = (sch.questions || []).filter((q: any) => !skipTypes.includes(q.type))
  } finally { loading.value = false }
}

const detailDialog = reactive({ visible: false, response: null as any, answers: {} as any, schema: null as any })
const answerRows = computed(() => {
  if (!detailDialog.response) return []
  const sch = detailDialog.schema
  const ans = detailDialog.answers || {}
  if (!sch) {
    return Object.entries(ans).map(([k, v]) => ({ questionId: k, title: k, value: formatVal(v) }))
  }
  return sch.questions.map((q: any) => ({
    questionId: q.id,
    title: q.title,
    value: formatVal(ans[q.id])
  }))
})
function formatVal(v: any) {
  if (v == null) return '(未填)'
  if (Array.isArray(v)) return v.join(', ')
  if (typeof v === 'object') return JSON.stringify(v)
  return v
}

async function viewDetail(row: any) {
  const res: any = await adminApi.surveyResponseDetail(row.id)
  detailDialog.response = res.response
  detailDialog.answers = res.answers || {}
  let sch = null
  if (res.survey && res.survey.schema) {
    try { sch = JSON.parse(res.survey.schema) } catch {}
  }
  detailDialog.schema = sch
  detailDialog.visible = true
}

async function exportCSV() {
  try {
    const res: any = await adminApi.surveyResponseExport(surveyId)
    const url = URL.createObjectURL(new Blob([res], { type: 'text/csv;charset=utf-8' }))
    const a = document.createElement('a')
    a.href = url; a.download = `survey_${surveyId}.csv`; a.click()
    URL.revokeObjectURL(url)
  } catch (e) { ElMessage.error('导出失败') }
}

async function del(row: any) {
  await ElMessageBox.confirm(`确认删除答卷 #${row.id}?`, '提示', { type: 'warning' })
  await adminApi.surveyResponseDel({ id: row.id })
  ElMessage.success('已删除')
  load()
}

function goBack() { router.push('/survey') }

onMounted(load)
</script>

<style scoped>
.header { display:flex; align-items:center; }
</style>
