<template>
  <div>
    <el-card class="resp-card">
      <div class="header">
        <el-button @click="goBack">‹ 返回</el-button>
        <h3 style="margin:0 0 0 12px;display:inline-block">考试记录: {{ examTitle }}</h3>
        <el-tag v-if="examId" type="info" size="small" style="margin-left:8px">ExamID: {{ examId }}</el-tag>
      </div>
      <div class="toolbar">
        <el-input v-model="keyword" placeholder="搜索用户ID" clearable style="width:260px" @clear="search" @keyup.enter="search" />
        <div class="toolbar-actions">
          <el-button size="small" type="danger" :disabled="!selectedIds.length" @click="batchDel">批量删除</el-button>
        </div>
      </div>

      <el-table :data="list" v-loading="loading" stripe style="margin-top:16px" border @selection-change="onSelectionChange">
        <el-table-column type="selection" width="40" />
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="userId" label="用户" width="120" />
        <el-table-column label="得分" width="100" align="center">
          <template #default="{ row }">{{ row.score }}/{{ row.totalScore }}</template>
        </el-table-column>
        <el-table-column label="通过" width="70" align="center">
          <template #default="{ row }">
            <el-tag :type="row.pass===1?'success':'danger'" size="small">{{ row.pass===1?'是':'否' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status===1?'success':(row.status===2?'warning':'info')" size="small">
              {{ row.status===1?'已提交':(row.status===2?'已批改':'进行中') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="用时" width="80">
          <template #default="{ row }">{{ formatDuration(row.timeSpent) }}</template>
        </el-table-column>
        <el-table-column label="开始时间" min-width="140">
          <template #default="{ row }">{{ formatTime(row.startTime) }}</template>
        </el-table-column>
        <el-table-column label="提交时间" min-width="150">
          <template #default="{ row }">{{ formatTime(row.submitTime) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="140" fixed="right">
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

    <el-dialog v-model="detailDialog.visible" :title="`记录详情 #${detailDialog.record?.id}`" width="800px">
      <el-descriptions :column="2" border size="small" style="margin-bottom:12px">
        <el-descriptions-item label="用户">{{ detailDialog.record?.userId || '匿名' }}</el-descriptions-item>
        <el-descriptions-item label="得分">{{ detailDialog.record?.score }}/{{ detailDialog.record?.totalScore }}</el-descriptions-item>
        <el-descriptions-item label="用时">{{ formatDuration(detailDialog.record?.timeSpent) }}</el-descriptions-item>
        <el-descriptions-item label="通过">{{ detailDialog.record?.pass===1?'是':'否' }}</el-descriptions-item>
        <el-descriptions-item label="开始">{{ formatTime(detailDialog.record?.startTime) }}</el-descriptions-item>
        <el-descriptions-item label="提交">{{ formatTime(detailDialog.record?.submitTime) }}</el-descriptions-item>
      </el-descriptions>
      <el-table :data="answerRows" stripe size="small" border>
        <el-table-column prop="questionId" label="题号" width="80" />
        <el-table-column prop="title" label="题目" min-width="200" show-overflow-tooltip />
        <el-table-column prop="value" label="答案" min-width="120" />
        <el-table-column prop="correct" label="正确" width="60" align="center">
          <template #default="{ row }">
            <el-tag :type="row.correct?'success':'danger'" size="small">{{ row.correct?'对':'错' }}</el-tag>
          </template>
        </el-table-column>
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
const examId = Number(route.query.examId || 0)
const examTitle = String(route.query.title || `考试 #${examId}`)

const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref<any[]>([])
const loading = ref(false)
const questions = ref<any[]>([])
const skipTypes = ['divider', 'description', 'pagination', 'questionSet']
const keyword = ref('')
const selectedIds = ref<number[]>([])

function onSelectionChange(rows: any[]) {
  selectedIds.value = rows.map((r: any) => r.id)
}

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
  if (!examId) { ElMessage.error('缺少 examId'); return }
  loading.value = true
  try {
    const [res, detailRes]: any = await Promise.all([
      adminApi.examRecordList({ examId, page: page.value, pageSize: pageSize.value, keyword: keyword.value || undefined }),
      adminApi.examDetail(examId)
    ])
    list.value = res.data?.list || res.list || []
    total.value = res.data?.total || res.total || 0
    const detail = (detailRes as any).data || detailRes
    const survey = detail?.survey || detail
    const raw = survey?.schema
    const sch = raw ? (typeof raw === 'string' ? JSON.parse(raw) : raw) : { questions: [] }
    questions.value = (sch.questions || []).filter((q: any) => !skipTypes.includes(q.type))
  } finally { loading.value = false }
}

function search() {
  page.value = 1
  load()
}

async function batchDel() {
  if (!selectedIds.value.length) return
  await ElMessageBox.confirm(`确认删除选中的 ${selectedIds.value.length} 条记录？`, '提示', { type: 'warning' })
  await adminApi.examRecordBatchDel({ ids: selectedIds.value.join(',') })
  ElMessage.success('已删除')
  selectedIds.value = []
  load()
}

const detailDialog = reactive({ visible: false, record: null as any, answers: {} as any, schema: null as any, scoring: {} as any })
const answerRows = computed(() => {
  if (!detailDialog.record) return []
  const sch = detailDialog.schema
  const ans = detailDialog.answers || {}
  const scoring = detailDialog.scoring || {}
  if (!sch) {
    return Object.entries(ans).map(([k, v]) => ({ questionId: k, title: k, value: formatVal(v), correct: scoring[k] }))
  }
  return sch.questions.filter((q: any) => !skipTypes.includes(q.type)).map((q: any, i: number) => ({
    questionId: i + 1,
    title: stripHtml(q.title || ''),
    value: formatVal(ans[q.id]),
    correct: scoring[q.id]
  }))
})
function stripHtml(s: string) {
  const d = document.createElement('div')
  d.innerHTML = s
  return d.textContent || d.innerText || ''
}
function formatVal(v: any) {
  if (v == null) return '(未填)'
  if (Array.isArray(v)) return v.join(', ')
  if (typeof v === 'object') return JSON.stringify(v)
  return v
}

async function viewDetail(row: any) {
  const raw: any = await adminApi.examRecordDetail(row.id)
  const data = raw.data || raw
  detailDialog.record = data.record || data
  detailDialog.answers = data.answers || {}
  detailDialog.scoring = data.scoring || {}
  detailDialog.schema = data.schema || null
  detailDialog.visible = true
}

async function del(row: any) {
  await ElMessageBox.confirm(`确认删除记录 #${row.id}?`, '提示', { type: 'warning' })
  await adminApi.examRecordDel({ id: row.id })
  ElMessage.success('已删除')
  load()
}

function goBack() { router.push('/exam/list') }

onMounted(load)
</script>

<style scoped>
.header { display:flex; align-items:center; }
.resp-card { overflow:visible; }
.toolbar { position:sticky; top:0; z-index:10; background:#fff; padding:8px 0; }
.toolbar-actions { margin-top:8px; display:flex; gap:8px; justify-content:flex-end; }
</style>
