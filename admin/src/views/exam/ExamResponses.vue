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
        <el-table-column label="用户信息" min-width="280">
          <template #default="{ row }">
            <el-tooltip :content="formatRecordUserInfo(row)" placement="top" :show-after="400">
              <span class="user-info-cell">{{ formatRecordUserInfo(row) }}</span>
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column label="得分" width="100" align="center">
          <template #default="{ row }">{{ row.score }}/{{ row.totalScore }}</template>
        </el-table-column>
        <el-table-column label="通过" width="70" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.status === 1" type="warning" size="small">待判定</el-tag>
            <el-tag v-else :type="row.pass===1?'success':'danger'" size="small">{{ row.pass===1?'是':'否' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status===1?'warning':(row.status===2?'success':'info')" size="small">
              {{ row.status===1?'待批改':(row.status===2?'已批改':'进行中') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="用时" width="150">
          <template #default="{ row }"><span class="time-cell">{{ formatDuration(row.timeSpent) }}</span></template>
        </el-table-column>
        <el-table-column label="开始时间" width="180">
          <template #default="{ row }"><span class="time-cell">{{ formatTime(row.startTime) }}</span></template>
        </el-table-column>
        <el-table-column label="提交时间" width="180">
          <template #default="{ row }"><span class="time-cell">{{ formatTime(row.submitTime) }}</span></template>
        </el-table-column>
        <el-table-column prop="addIp" label="IP地址" width="160" show-overflow-tooltip>
          <template #default="{ row }">{{ row.addIp || '-' }}</template>
        </el-table-column>
        <el-table-column prop="browser" label="浏览器" width="180" show-overflow-tooltip>
          <template #default="{ row }">{{ row.browser || '-' }}</template>
        </el-table-column>
        <el-table-column prop="deviceType" label="设备类型" width="120" show-overflow-tooltip>
          <template #default="{ row }">{{ row.deviceType || '-' }}</template>
        </el-table-column>
        <el-table-column prop="platformType" label="平台类型" width="140" show-overflow-tooltip>
          <template #default="{ row }">{{ row.platformType || '-' }}</template>
        </el-table-column>
        <el-table-column label="操作" :width="hasExamGradingPermission() ? 330 : 230" fixed="right">
          <template #default="{ row }">
            <div class="record-actions">
              <el-button size="small" @click="viewDetail(row)">查看</el-button>
              <el-button v-if="canOpenGrading(row)" size="small" type="primary" @click="openGrading(row)">{{ row.status === 1 ? '批改' : '批改结果' }}</el-button>
              <el-dropdown trigger="click" @command="(format: ExamExportFormat) => exportRecord(row, format)">
                <el-button size="small" :icon="Download" :loading="exportingRecordId === row.id">导出</el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="pdf">导出 PDF</el-dropdown-item>
                    <el-dropdown-item command="docx">导出 Word</el-dropdown-item>
                    <el-dropdown-item command="xlsx">导出 Excel</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
              <el-button size="small" type="danger" @click="del(row)">删除</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
      <div class="admin-pagination">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :page-sizes="[10, 20, 50, 100]" :total="total" layout="total, sizes, prev, pager, next" @current-change="load" @size-change="handlePageSizeChange" />
      </div>
    </el-card>

    <el-dialog v-model="detailDialog.visible" :title="`记录详情 #${detailDialog.record?.id}`" width="800px">
      <el-descriptions :column="2" border size="small" style="margin-bottom:12px">
        <el-descriptions-item label="用户信息">{{ formatRecordUserInfo(detailDialog.record) }}</el-descriptions-item>
        <el-descriptions-item label="得分">{{ detailDialog.record?.score }}/{{ detailDialog.record?.totalScore }}</el-descriptions-item>
        <el-descriptions-item label="用时">{{ formatDuration(detailDialog.record?.timeSpent) }}</el-descriptions-item>
        <el-descriptions-item label="通过">{{ detailDialog.record?.status===1?'待判定':(detailDialog.record?.pass===1?'是':'否') }}</el-descriptions-item>
        <el-descriptions-item label="开始">{{ formatTime(detailDialog.record?.startTime) }}</el-descriptions-item>
        <el-descriptions-item label="提交">{{ formatTime(detailDialog.record?.submitTime) }}</el-descriptions-item>
        <el-descriptions-item label="IP地址"><span class="client-info-text">{{ detailDialog.record?.addIp || '-' }}</span></el-descriptions-item>
        <el-descriptions-item label="浏览器"><span class="client-info-text">{{ detailDialog.record?.browser || '-' }}</span></el-descriptions-item>
        <el-descriptions-item label="设备类型"><span class="client-info-text">{{ detailDialog.record?.deviceType || '-' }}</span></el-descriptions-item>
        <el-descriptions-item label="平台类型"><span class="client-info-text">{{ detailDialog.record?.platformType || '-' }}</span></el-descriptions-item>
        <el-descriptions-item label="原始UA" :span="2"><span class="client-info-text">{{ detailDialog.record?.device || '-' }}</span></el-descriptions-item>
        <el-descriptions-item label="批改人ID">{{ detailDialog.record?.gradedBy || '-' }}</el-descriptions-item>
        <el-descriptions-item label="批改时间">{{ formatTime(detailDialog.record?.gradedAt) }}</el-descriptions-item>
        <el-descriptions-item label="人工评语" :span="2"><span class="client-info-text">{{ detailDialog.record?.gradingComment || '暂无人工评语' }}</span></el-descriptions-item>
      </el-descriptions>
      <el-table :data="answerRows" stripe size="small" border>
        <el-table-column prop="questionId" label="题号" width="80" />
        <el-table-column prop="title" label="题目" min-width="200" show-overflow-tooltip />
        <el-table-column prop="value" label="答案" min-width="120" />
        <el-table-column label="判定" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="gradeState(row).type" size="small">{{ gradeState(row).label }}</el-tag>
          </template>
        </el-table-column>
      </el-table>
      <template #footer>
        <el-button @click="detailDialog.visible = false">关闭</el-button>
        <el-button v-if="canOpenGrading(detailDialog.record)" type="primary" @click="openGrading(detailDialog.record)">{{ detailDialog.record?.status === 1 ? '开始批改' : '查看批改结果' }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onActivated, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Download } from '@element-plus/icons-vue'
import { adminApi } from '../../api'
import { hasExamGradingPermission } from '@/api/examGrading'
import type { ExamRecord } from '@/types/examGrading'

type ExamExportFormat = 'pdf' | 'docx' | 'xlsx'

const route = useRoute()
const router = useRouter()
const examId = Number(route.query.examId || 0)
const examTitle = String(route.query.title || `考试 #${examId}`)
const listPath = route.fullPath

const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref<any[]>([])
const loading = ref(false)
const questions = ref<any[]>([])
const skipTypes = ['divider', 'description', 'pagination', 'questionSet']
const PERSONAL_INFO_TYPES = new Set(['name', 'phone', 'email', 'studentId', 'employeeId', 'class'])
const PERSONAL_INFO_LABELS: Record<string, string> = { name: '姓名', phone: '手机', email: '邮箱', studentId: '学号', employeeId: '工号', class: '班级' }
const keyword = ref('')
const selectedIds = ref<number[]>([])
const exportingRecordId = ref(0)

function onSelectionChange(rows: any[]) {
  selectedIds.value = rows.map((r: any) => r.id)
}

function formatTime(ms: number) {
  if (!ms) return '-'
  return new Date(ms).toLocaleString()
}
function formatDuration(sec: number) {
  if (!sec) return '-'
  const h = Math.floor(sec / 3600)
  const m = Math.floor(sec / 60)
  const s = sec % 60
  if (h > 0) return `${h}小时${m % 60}分${s}秒`
  return m > 0 ? `${m}分${s}秒` : `${s}秒`
}

function parseRecordAnswers(raw: unknown): Record<string, unknown> {
  if (raw && typeof raw === 'object' && !Array.isArray(raw)) return raw as Record<string, unknown>
  if (typeof raw !== 'string' || !raw.trim()) return {}
  try {
    const parsed = JSON.parse(raw)
    return parsed && typeof parsed === 'object' && !Array.isArray(parsed) ? parsed : {}
  } catch { return {} }
}

function formatRecordUserInfo(row: any): string {
  if (!row) return '匿名'
  const parts: string[] = []
  const userId = String(row.userId ?? '').trim()
  if (userId && userId !== '0') parts.push(`用户ID：${userId}`)
  const values = parseRecordAnswers(row.answers)
  for (const question of questions.value) {
    if (!PERSONAL_INFO_TYPES.has(question.type)) continue
    const value = values[question.id]
    if (value === undefined || value === null || String(value).trim() === '') continue
    const label = stripHtml(question.title || '').trim() || PERSONAL_INFO_LABELS[question.type] || '用户信息'
    parts.push(`${label}：${formatVal(value)}`)
  }
  return parts.length ? parts.join('；') : '匿名'
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

function handlePageSizeChange() {
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

const detailDialog = reactive({ visible: false, record: null as any, answers: {} as any, schema: null as any, scoring: {} as any, results: [] as any[] })
const answerRows = computed(() => {
  if (!detailDialog.record) return []
  const sch = detailDialog.schema
  const ans = detailDialog.answers || {}
  const scoring = detailDialog.scoring || {}
  const resultMap = new Map<number, any>((detailDialog.results || []).map((item: any) => [Number(item.questionId), item]))
  if (!sch) {
    return Object.entries(ans).map(([k, v]) => ({ questionId: k, title: k, value: formatVal(v), grade: resultMap.get(Number(k)) || (scoring[k] === undefined ? null : { correct: scoring[k] }) }))
  }
  return sch.questions
    .map((question: any, schemaIndex: number) => ({ question, schemaIndex }))
    .filter(({ question }: any) => !skipTypes.includes(question.type))
    .map(({ question, schemaIndex }: any, displayIndex: number) => {
      const storedGrade = resultMap.get(schemaIndex + 1) || (scoring[question.id] === undefined ? null : { correct: scoring[question.id] })
      return {
        questionId: displayIndex + 1,
        title: stripHtml(question.title || ''),
        value: formatVal(ans[question.id]),
        grade: isRecordQuestionGraded(question) ? storedGrade : null
      }
    })
})

function isRecordQuestionGraded(question: any): boolean {
  if (!question || PERSONAL_INFO_TYPES.has(question.type)) return false
  if (String(question.examScoreMode || '').trim()) return true
  if (Number(question.examScore) > 0) return true
  return (question.props?.fields || []).some((field: any) => Number(field.examScore) > 0 || field.examCorrect || String(field.examCorrectAnswer || '').trim())
}

function gradeState(row: any): { label: string; type: 'success' | 'danger' | 'warning' | 'info' } {
  if (!row?.grade) return { label: '不计分', type: 'info' }
  if (row.grade.needManual && detailDialog.record?.status !== 2) return { label: '待批改', type: 'warning' }
  if (row.grade.gotScore > 0 && row.grade.gotScore < row.grade.fullScore) return { label: '部分得分', type: 'warning' }
  return row.grade.correct ? { label: '正确', type: 'success' } : { label: '错误', type: 'danger' }
}
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
  detailDialog.results = data.results || []
  detailDialog.schema = data.schema || null
  detailDialog.visible = true
}

async function exportRecord(row: any, format: ExamExportFormat) {
  if (!examId || !row?.id || exportingRecordId.value) return
  exportingRecordId.value = row.id
  try {
    const blob = await adminApi.examRecordExport({ examId, recordId: row.id, format })
    const url = URL.createObjectURL(blob)
    const anchor = document.createElement('a')
    try {
      anchor.href = url
      anchor.download = `考试答卷_${examId}_${row.id}.${format}`
      document.body.appendChild(anchor)
      anchor.click()
    } finally {
      anchor.remove()
      setTimeout(() => URL.revokeObjectURL(url), 1000)
    }
    ElMessage.success('答卷导出成功')
  } finally {
    exportingRecordId.value = 0
  }
}

async function del(row: any) {
  await ElMessageBox.confirm(`确认删除记录 #${row.id}?`, '提示', { type: 'warning' })
  await adminApi.examRecordDel({ id: row.id })
  ElMessage.success('已删除')
  load()
}

function goBack() { router.push('/exam/list') }

function canOpenGrading(record: Pick<ExamRecord, 'id' | 'status'> | null): boolean {
  return hasExamGradingPermission() && !!record && (record.status === 1 || record.status === 2)
}

function openGrading(record: Pick<ExamRecord, 'id' | 'status'> | null) {
  if (!record || !canOpenGrading(record)) return
  detailDialog.visible = false
  void router.push({ path: `/exam/responses/${record.id}/grading`, query: { examId, returnTo: listPath } })
}

onMounted(load)
let firstActivation = true
onActivated(() => {
  if (firstActivation) { firstActivation = false; return }
  // 保留当前搜索与分页，仅刷新缓存列表中的答卷状态。
  void load().catch(() => { /* 请求层已展示错误，可通过查询重试。 */ })
})
</script>

<style scoped>
.header { display:flex; align-items:center; }
.resp-card { overflow:visible; }
.toolbar { position:sticky; top:0; z-index:10; background:#fff; padding:8px 0; }
.toolbar-actions { margin-top:8px; display:flex; gap:8px; justify-content:flex-end; }
.record-actions { display:flex; align-items:center; gap:8px; white-space:nowrap; }
.record-actions :deep(.el-button) { margin-left:0; }
.time-cell { white-space:nowrap; }
.user-info-cell { display:block; line-height:1.6; white-space:normal; word-break:break-all; }
.client-info-text { white-space:pre-wrap; overflow-wrap:anywhere; }
</style>
