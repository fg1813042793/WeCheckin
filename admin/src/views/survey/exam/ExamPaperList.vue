<template>
  <div class="page">
    <div class="page-banner">
      <div><h2>试卷管理</h2><p>组卷出题，灵活配置考试</p></div>
      <el-button type="primary" size="large" @click="showAdd"><el-icon style="margin-right:6px"><Plus /></el-icon>新建试卷</el-button>
    </div>

    <el-card shadow="never" class="main-card">
      <div class="toolbar">
        <el-input v-model="keyword" placeholder="搜索标题" clearable style="width:220px" @keyup.enter="load" />
        <el-button type="primary" @click="load">搜索</el-button>
      </div>
      <el-table :data="list" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="title" label="标题" min-width="200" />
        <el-table-column prop="description" label="描述" min-width="180" show-overflow-tooltip />
        <el-table-column label="分数" width="110">
          <template #default="{ row }"><span class="score-badge">总分 {{ row.totalScore }}</span> / <span class="pass-badge">及格 {{ row.passScore }}</span></template>
        </el-table-column>
        <el-table-column label="时长" width="90"><template #default="{ row }">{{ row.timeLimit || '-' }} 分</template></el-table-column>
        <el-table-column label="显示答案" width="80" align="center">
          <template #default="{ row }"><el-tag :type="row.showAnswer===1?'success':'info'" size="small" round>{{ row.showAnswer===1?'是':'否' }}</el-tag></template>
        </el-table-column>
        <el-table-column label="状态" width="70" align="center">
          <template #default="{ row }"><el-tag :type="row.status===1?'success':'danger'" size="small" round>{{ row.status===1?'启用':'停用' }}</el-tag></template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="showEdit(row)">编辑</el-button>
            <el-button size="small" @click="preview(row)">预览</el-button>
            <el-popconfirm title="确认删除?" @confirm="del(row)"><template #reference><el-button size="small" type="danger" plain>删除</el-button></template></el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
      <div class="page-bar"><el-pagination v-model:current-page="page" :page-size="pageSize" :total="total" layout="total,prev,pager,next" @current-change="load" background /></div>
    </el-card>

    <!-- 新增/编辑弹窗 -->
    <el-dialog v-model="dialog.visible" :title="dialog.title" width="900px" :close-on-click-modal="false" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="标题" prop="title">
          <el-input v-model="form.title" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="题目">
          <el-button size="small" @click="openPicker">+ 从题库选题</el-button>
          <el-button size="small" type="danger" @click="clearQ" v-if="picked.length">清空</el-button>
          <div style="margin-top:8px;border:1px solid #eee;border-radius:4px;max-height:300px;overflow:auto">
            <el-table :data="picked" size="small" @selection-change="selPicked = $event">
              <el-table-column type="selection" width="40" />
              <el-table-column prop="id" label="ID" width="60" />
              <el-table-column label="类型" width="80">
                <template #default="{ row }">
                  <el-tag size="small">{{ row.type }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="title" label="题干" min-width="200" show-overflow-tooltip />
              <el-table-column prop="score" label="分值" width="70" />
            </el-table>
          </div>
          <div style="color:#888;margin-top:6px;font-size:12px">
            已选 {{ picked.length }} 题，自动合计总分: {{ pickedScore }}
          </div>
        </el-form-item>
        <el-form-item label="时长(分)">
          <el-input-number v-model="form.timeLimit" :min="1" :max="600" />
        </el-form-item>
        <el-form-item label="及格分">
          <el-input-number v-model="form.passScore" :min="0" />
        </el-form-item>
        <el-form-item label="交卷显示答案">
          <el-switch v-model="form.showAnswerBool" :active-value="1" :inactive-value="0" />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="form.statusBool" :active-value="1" :inactive-value="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </template>
    </el-dialog>

    <!-- 题库选择弹窗 -->
    <el-dialog v-model="picker.visible" title="选择题库" width="900px" :close-on-click-modal="false">
      <div style="display:flex;gap:8px;margin-bottom:8px">
        <el-input v-model="picker.q" placeholder="搜索" clearable style="width:200px" @keyup.enter="pickerLoad" />
        <el-button @click="pickerLoad">搜索</el-button>
      </div>
      <el-table :data="picker.list" v-loading="picker.loading" @selection-change="picker.sel = $event" max-height="420">
        <el-table-column type="selection" width="40" />
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="type" label="类型" width="80" />
        <el-table-column prop="title" label="题干" min-width="240" show-overflow-tooltip />
        <el-table-column prop="category" label="分类" width="100" />
        <el-table-column prop="score" label="分值" width="70" />
      </el-table>
      <div style="text-align:center;margin-top:8px">
        <el-pagination v-model:current-page="picker.page" :page-size="picker.pageSize" :total="picker.total" layout="total,prev,pager,next" @current-change="pickerLoad" />
      </div>
      <template #footer>
        <el-button @click="picker.visible = false">取消</el-button>
        <el-button type="primary" @click="confirmPick">确认选择</el-button>
      </template>
    </el-dialog>

    <!-- 预览弹窗 -->
    <el-dialog v-model="previewDialog.visible" :title="'预览: ' + previewDialog.paper.title" width="800px">
      <div v-for="q in previewDialog.questions" :key="q.id" style="margin-bottom:14px">
        <div style="font-weight:500">{{ q.id }}. {{ q.title }} ({{ q.score }}分)</div>
        <div style="color:#888;font-size:12px;margin-top:2px">类型: {{ q.type }} | 答案: {{ q.answer }}</div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { adminApi } from '../../../api'

const keyword = ref('')
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref<any[]>([])
const loading = ref(false)

async function load() {
  loading.value = true
  try {
    const res: any = await adminApi.examPaperList({ page: page.value, pageSize: pageSize.value, keyword: keyword.value })
    list.value = res.list || []
    total.value = res.total || 0
  } finally { loading.value = false }
}

const dialog = reactive({ visible: false, title: '', isCreate: true })
const formRef = ref()
const saving = ref(false)
const form = reactive<any>({
  id: 0, title: '', description: '',
  timeLimit: 60, passScore: 60,
  showAnswerBool: 0, statusBool: 1
})
const picked = ref<any[]>([])
const selPicked = ref<any[]>([])
const pickedScore = computed(() => picked.value.reduce((s, q) => s + (q.score || 0), 0))

const rules = {
  title: [{ required: true, message: '请输入标题', trigger: 'blur' }]
}

const picker = reactive({
  visible: false, q: '', page: 1, pageSize: 20, total: 0,
  list: [] as any[], loading: false, sel: [] as any[]
})
async function pickerLoad() {
  picker.loading = true
  try {
    const res: any = await adminApi.examQuestionList({ page: picker.page, pageSize: picker.pageSize, keyword: picker.q })
    picker.list = res.list || []
    picker.total = res.total || 0
  } finally { picker.loading = false }
}
function openPicker() {
  picker.page = 1; picker.q = ''; picker.sel = []
  picker.visible = true
  pickerLoad()
}
function confirmPick() {
  const existing = new Set(picked.value.map((q: any) => q.id))
  for (const q of picker.sel) {
    if (!existing.has(q.id)) picked.value.push(q)
  }
  picker.visible = false
}
function clearQ() {
  if (selPicked.value.length) {
    const ids = new Set(selPicked.value.map((q: any) => q.id))
    picked.value = picked.value.filter((q: any) => !ids.has(q.id))
  } else {
    picked.value = []
  }
}

function showAdd() {
  Object.assign(form, { id: 0, title: '', description: '', timeLimit: 60, passScore: 60, showAnswerBool: 0, statusBool: 1 })
  picked.value = []
  dialog.isCreate = true
  dialog.title = '新建试卷'
  dialog.visible = true
}

async function showEdit(row: any) {
  Object.assign(form, {
    id: row.id, title: row.title, description: row.description,
    timeLimit: row.timeLimit, passScore: row.passScore,
    showAnswerBool: row.showAnswer, statusBool: row.status
  })
  // 加载题目详情
  const detail: any = await adminApi.examPaperDetail(row.id)
  picked.value = detail.questions || []
  dialog.isCreate = false
  dialog.title = '编辑试卷 #' + row.id
  dialog.visible = true
}

async function save() {
  await formRef.value.validate()
  if (picked.value.length === 0) { ElMessage.warning('请至少选一题'); return }
  saving.value = true
  try {
    const payload: any = {
      title: form.title, description: form.description,
      timeLimit: form.timeLimit, passScore: form.passScore,
      showAnswer: form.showAnswerBool, status: form.statusBool,
      questionIds: JSON.stringify(picked.value.map((q: any) => q.id)),
      totalScore: pickedScore.value
    }
    if (dialog.isCreate) {
      await adminApi.examPaperInsert(payload)
      ElMessage.success('创建成功')
    } else {
      payload.id = form.id
      await adminApi.examPaperEdit(payload)
      ElMessage.success('更新成功')
    }
    dialog.visible = false
    load()
  } finally { saving.value = false }
}

async function del(row: any) {
  await ElMessageBox.confirm(`确认删除试卷「${row.title}」?`, '提示', { type: 'warning' })
  await adminApi.examPaperDel({ id: row.id })
  ElMessage.success('已删除')
  load()
}

const previewDialog = reactive({ visible: false, paper: {} as any, questions: [] as any[] })
async function preview(row: any) {
  const detail: any = await adminApi.examPaperDetail(row.id)
  previewDialog.paper = detail.paper
  previewDialog.questions = detail.questions || []
  previewDialog.visible = true
}

onMounted(load)
</script>

<style scoped>
.page { max-width:1400px; margin:0 auto; }
.page-banner { display:flex; justify-content:space-between; align-items:center; padding:24px 0 16px; }
.page-banner h2 { margin:0; font-size:22px; font-weight:600; color:#1a1a2e; }
.page-banner p { margin:4px 0 0; color:#888; font-size:13px; }
.main-card { border-radius:12px; }
.toolbar { display:flex; gap:8px; margin-bottom:16px; }
.score-badge { color:#fb454c; font-weight:500; }
.pass-badge { color:#67c23a; font-weight:500; }
.page-bar { text-align:center; margin-top:16px; }
</style>
