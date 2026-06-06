<template>
  <div class="page">
    <div class="page-banner">
      <div><h2>考试记录</h2><p>{{ examTitle }}</p></div>
      <el-button @click="goBack"><el-icon><ArrowLeft /></el-icon> 返回</el-button>
    </div>

    <el-card shadow="never" class="main-card">
      <el-table :data="list" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column label="用户" width="120"><template #default="{row}">{{ row.userId || '匿名' }}</template></el-table-column>
        <el-table-column label="得分" width="130">
          <template #default="{row}">
            <span :class="row.pass===1?'score-pass':'score-fail'">{{ row.score }}</span>
            <span class="score-divider">/</span>
            <span>{{ row.totalScore }}</span>
            <el-tag v-if="row.pass===1" size="small" type="success" round style="margin-left:6px">通过</el-tag>
            <el-tag v-else size="small" type="danger" round style="margin-left:6px">未通过</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{row}">
            <el-tag v-if="row.status===0" type="info" size="small" round>进行中</el-tag>
            <el-tag v-else-if="row.status===1" type="warning" size="small" round>待批改</el-tag>
            <el-tag v-else type="success" size="small" round>已完成</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="提交时间" min-width="150"><template #default="{row}">{{ fmtTime(row.submitTime) }}</template></el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{row}">
            <el-button size="small" @click="viewDetail(row)">查看</el-button>
            <el-button v-if="row.status===1" size="small" type="warning" @click="openGrade(row)">人工判分</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div class="page-bar"><el-pagination v-model:current-page="page" :page-size="pageSize" :total="total" layout="total,prev,pager,next" @current-change="load" background /></div>
    </el-card>

    <el-dialog v-model="detailDialog.visible" :title="`答题详情 #${detailDialog.record?.id}`" width="800px">
      <el-descriptions :column="2" border size="small" style="margin-bottom:12px">
        <el-descriptions-item label="用户">{{ detailDialog.record?.userId }}</el-descriptions-item>
        <el-descriptions-item label="得分"><span :class="detailDialog.record?.pass===1?'score-pass':'score-fail'">{{ detailDialog.record?.score }}</span> / {{ detailDialog.record?.totalScore }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag v-if="detailDialog.record?.status===0" type="info" size="small">进行中</el-tag>
          <el-tag v-else-if="detailDialog.record?.status===1" type="warning" size="small">待批改</el-tag>
          <el-tag v-else type="success" size="small">已完成</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="提交">{{ fmtTime(detailDialog.record?.submitTime) }}</el-descriptions-item>
      </el-descriptions>
      <el-table :data="detailDialog.results" size="small" border>
        <el-table-column prop="questionId" label="题号" width="70" />
        <el-table-column label="得分" width="120">
          <template #default="{row}"><span class="score-pass">{{ row.gotScore }}</span> / {{ row.fullScore }}</template>
        </el-table-column>
        <el-table-column label="结果" width="100" align="center">
          <template #default="{row}">
            <el-tag v-if="row.needManual" type="warning" size="small" round>待人工</el-tag>
            <el-tag v-else-if="row.correct" type="success" size="small" round>正确</el-tag>
            <el-tag v-else type="danger" size="small" round>错误</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="reason" label="原因" min-width="200" />
      </el-table>
    </el-dialog>

    <el-dialog v-model="gradeDialog.visible" :title="`人工判分 - 记录 #${gradeDialog.record?.id}`" width="480px">
      <el-alert v-if="gradeDialog.record" :title="`当前得分: ${gradeDialog.record.score}/${gradeDialog.record.totalScore}`" type="info" show-icon style="margin-bottom:16px" />
      <el-form label-width="80px">
        <el-form-item label="题号"><el-input-number v-model="gradeDialog.qid" :min="1" style="width:100%" /></el-form-item>
        <el-form-item label="得分"><el-input-number v-model="gradeDialog.score" :min="0" :max="999" style="width:100%" /></el-form-item>
        <div style="color:#888;font-size:12px">设置该题得分，系统将自动重算总分</div>
      </el-form>
      <template #footer>
        <el-button @click="gradeDialog.visible=false">取消</el-button>
        <el-button type="primary" :loading="gradeDialog.saving" @click="submitGrade">提交</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { adminApi } from '../../../api'

const route = useRoute()
const router = useRouter()
const examId = Number(route.query.examId || 0)
const examTitle = String(route.query.title || `考试 #${examId}`)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref<any[]>([])
const loading = ref(false)

function fmtTime(ms: number) { return ms ? new Date(ms).toLocaleString() : '-' }

async function load() {
  if (!examId) return
  loading.value = true
  try {
    const res: any = await adminApi.examRecordList({ examId, page: page.value, pageSize: pageSize.value })
    list.value = res.list || []; total.value = res.total || 0
  } finally { loading.value = false }
}

const detailDialog = reactive({ visible: false, record: null as any, results: [] as any[] })
function viewDetail(row: any) {
  detailDialog.record = row
  try { detailDialog.results = row.result ? JSON.parse(row.result).results || [] : [] } catch { detailDialog.results = [] }
  detailDialog.visible = true
}

const gradeDialog = reactive({ visible: false, record: null as any, qid: 0, score: 0, saving: false })
function openGrade(row: any) {
  gradeDialog.record = row; gradeDialog.qid = 0; gradeDialog.score = 0; gradeDialog.visible = true
}
async function submitGrade() {
  if (!gradeDialog.qid) { ElMessage.warning('请输入题号'); return }
  gradeDialog.saving = true
  try {
    await adminApi.examManualGrade({ recordId: gradeDialog.record.id, qid: gradeDialog.qid, score: gradeDialog.score })
    ElMessage.success('已评分')
    gradeDialog.visible = false; load()
  } catch { ElMessage.error('评分失败') }
  finally { gradeDialog.saving = false }
}

function goBack() { router.back() }

onMounted(load)
</script>

<style scoped>
.page { max-width:1400px; margin:0 auto; }
.page-banner { display:flex; justify-content:space-between; align-items:center; padding:24px 0 16px; }
.page-banner h2 { margin:0; font-size:22px; font-weight:600; color:#1a1a2e; }
.page-banner p { margin:4px 0 0; color:#888; font-size:13px; }
.main-card { border-radius:12px; }
.score-pass { color:#67c23a; font-weight:600; font-size:16px; }
.score-fail { color:#fb454c; font-weight:600; font-size:16px; }
.score-divider { color:#ccc; margin:0 4px; }
.page-bar { text-align:center; margin-top:16px; }
</style>
