<template>
  <div class="page">
    <div class="page-banner">
      <div><h2>考试管理</h2><p>发布考试、监控数据、管理成绩</p></div>
      <el-button type="primary" size="large" @click="showAdd"><el-icon style="margin-right:6px"><Plus /></el-icon>新建考试</el-button>
    </div>

    <el-card shadow="never" class="main-card">
      <div class="toolbar">
        <el-input v-model="keyword" placeholder="搜索考试" clearable style="width:220px" @keyup.enter="load" />
        <el-button type="primary" @click="load">搜索</el-button>
      </div>

      <el-table :data="list" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="title" label="标题" min-width="200" />
        <el-table-column label="时间" min-width="200">
          <template #default="{row}">
            <div class="time-cell">
              <span>{{ fmtTime(row.startTime) }}</span>
              <span style="color:#ccc;margin:0 4px">→</span>
              <span>{{ fmtTime(row.endTime) }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="时长" width="70" align="center"><template #default="{row}">{{ row.duration || '-' }}分</template></el-table-column>
        <el-table-column label="次数" width="60" align="center"><template #default="{row}">{{ row.maxAttempts || 1 }}</template></el-table-column>
        <el-table-column label="状态" width="70" align="center">
          <template #default="{row}"><el-tag :type="row.status===1?'success':'danger'" size="small" round>{{ row.status===1?'启用':'停用' }}</el-tag></template>
        </el-table-column>
        <el-table-column label="操作" width="320" fixed="right">
          <template #default="{row}">
            <el-button size="small" type="primary" @click="goDesigner(row)">设计</el-button>
            <el-button size="small" @click="showEdit(row)">编辑</el-button>
            <el-popconfirm title="确认删除?" @confirm="del(row)"><template #reference><el-button size="small" type="danger" plain>删除</el-button></template></el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
      <div class="page-bar"><el-pagination v-model:current-page="page" :page-size="pageSize" :total="total" layout="total,prev,pager,next" @current-change="load" background /></div>
    </el-card>

    <el-dialog v-model="dialog.visible" :title="dialog.isCreate?'新建考试':'编辑考试'" width="700px">
      <el-form :model="form" label-width="110px">
        <el-form-item label="考试标题"><el-input v-model="form.title" /></el-form-item>
        <el-row :gutter="16">
          <el-col :span="12"><el-form-item label="开始时间"><el-date-picker v-model="form.startDate" type="datetime" placeholder="不限" value-format="x" style="width:100%" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="结束时间"><el-date-picker v-model="form.endDate" type="datetime" placeholder="不限" value-format="x" style="width:100%" /></el-form-item></el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="8"><el-form-item label="时长(分)"><el-input-number v-model="form.duration" :min="1" style="width:100%" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="最大次数"><el-input-number v-model="form.maxAttempts" :min="1" :max="99" style="width:100%" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="显示分数"><el-switch v-model="form.showScoreBool" :active-value="1" :inactive-value="0" /></el-form-item></el-col>
        </el-row>
        <el-form-item label="状态"><el-switch v-model="form.statusBool" :active-value="1" :inactive-value="0" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialog.visible=false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'
import { adminApi } from '../../api'

const router = useRouter()
const keyword = ref('')
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref<any[]>([])
const loading = ref(false)

function fmtTime(ms: number) { return ms ? new Date(ms).toLocaleDateString() : '-' }

const dialog = reactive({ visible: false, isCreate: true })
const saving = ref(false)
const form = reactive<any>({ id: 0, title: '', startDate: null, endDate: null, duration: 60, maxAttempts: 1, showScoreBool: 1, statusBool: 1 })

async function load() {
  loading.value = true
  try {
    const res: any = await adminApi.examList({ page: page.value, pageSize: pageSize.value, keyword: keyword.value })
    list.value = res.list || []; total.value = res.total || 0
  } finally { loading.value = false }
}

function showAdd() {
  Object.assign(form, { id: 0, title: '', startDate: null, endDate: null, duration: 60, maxAttempts: 1, showScoreBool: 1, statusBool: 1 })
  dialog.isCreate = true; dialog.visible = true
}

function showEdit(row: any) {
  Object.assign(form, {
    id: row.id, title: row.title,
    startDate: row.startTime || null, endDate: row.endTime || null,
    duration: row.duration, maxAttempts: row.maxAttempts,
    showScoreBool: row.showScore, statusBool: row.status
  })
  dialog.isCreate = false; dialog.visible = true
}

async function save() {
  saving.value = true
  try {
    const payload: any = {
      id: form.id, title: form.title,
      startTime: form.startDate || 0, endTime: form.endDate || 0,
      duration: form.duration, maxAttempts: form.maxAttempts,
      showScore: form.showScoreBool, status: form.statusBool,
      schema: '', settings: '{}'
    }
    await adminApi.examSave(payload)
    ElMessage.success(dialog.isCreate ? '已创建' : '已更新')
    dialog.visible = false; load()
  } catch { ElMessage.error('保存失败') }
  finally { saving.value = false }
}

async function del(row: any) {
  await adminApi.examDelete({ id: row.id })
  ElMessage.success('已删除'); load()
}

function goDesigner(row: any) {
  router.push({ path: '/exam/designer', query: { id: String(row.id) } })
}

onMounted(() => { load() })
</script>

<style scoped>
.page { max-width:1400px; margin:0 auto; }
.page-banner { display:flex; justify-content:space-between; align-items:center; padding:24px 0 16px; }
.page-banner h2 { margin:0; font-size:22px; font-weight:600; color:#1a1a2e; }
.page-banner p { margin:4px 0 0; color:#888; font-size:13px; }
.main-card { border-radius:12px; }
.toolbar { display:flex; gap:8px; margin-bottom:16px; }
.time-cell { font-size:12px; color:#888; }
.page-bar { text-align:center; margin-top:16px; }
</style>
