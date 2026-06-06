<template>
  <div class="page">
    <div class="page-banner">
      <div><h2>题库管理</h2><p>创建和管理考试题目</p></div>
      <el-button type="primary" size="large" @click="openAdd"><el-icon style="margin-right:6px"><Plus /></el-icon>新增题目</el-button>
    </div>

    <el-card shadow="never" class="main-card">
      <div class="toolbar">
        <el-input v-model="keyword" placeholder="搜索题目" clearable style="width:200px" @keyup.enter="load" />
        <el-select v-model="typeFilter" placeholder="题型" clearable style="width:130px" @change="load">
          <el-option v-for="t in allTypes" :key="t" :label="t" :value="t" />
        </el-select>
        <el-input v-model="categoryFilter" placeholder="分类" clearable style="width:130px" @keyup.enter="load" />
        <el-button type="primary" @click="load">搜索</el-button>
      </div>

      <el-table :data="list" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column label="题型" width="100">
          <template #default="{row}"><el-tag size="small" round>{{ row.type }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="title" label="题目" min-width="240" show-overflow-tooltip />
        <el-table-column label="分值" width="70" align="center">
          <template #default="{row}"><span class="score-badge">{{ row.score }}</span></template>
        </el-table-column>
        <el-table-column prop="category" label="分类" width="100" />
        <el-table-column prop="difficulty" label="难度" width="80">
          <template #default="{row}"><el-tag :type="diffType(row.difficulty)" size="small">{{ row.difficulty||'普通' }}</el-tag></template>
        </el-table-column>
        <el-table-column label="状态" width="70">
          <template #default="{row}"><el-tag :type="row.status===1?'success':'danger'" size="small">{{ row.status===1?'启用':'停用' }}</el-tag></template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{row}">
            <el-button size="small" @click="openEdit(row)">编辑</el-button>
            <el-popconfirm title="确认删除?" @confirm="del(row)"><template #reference><el-button size="small" type="danger" plain>删除</el-button></template></el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
      <div class="page-bar"><el-pagination v-model:current-page="page" :page-size="pageSize" :total="total" layout="total,prev,pager,next" @current-change="load" background /></div>
    </el-card>

    <el-dialog v-model="dialog.visible" :title="dialog.isEdit?'编辑题目':'新增题目'" width="700px" top="5vh">
      <el-form :model="dialog.form" label-width="80px">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="题型">
              <el-select v-model="dialog.form.type" style="width:100%">
                <el-option v-for="t in allTypes" :key="t" :label="t" :value="t" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="分值"><el-input-number v-model="dialog.form.score" :min="0" :max="1000" style="width:100%" /></el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="题干"><el-input v-model="dialog.form.title" type="textarea" :rows="3" /></el-form-item>
        <el-form-item label="选项">
          <el-input v-model="dialog.form.options" type="textarea" :rows="3" placeholder='JSON 数组，如 ["选项A","选项B"]' />
        </el-form-item>
        <el-form-item label="答案"><el-input v-model="dialog.form.answer" :rows="2" placeholder="正确答案（按题型格式）" /></el-form-item>
        <el-row :gutter="16">
          <el-col :span="8"><el-form-item label="分类"><el-input v-model="dialog.form.category" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="难度"><el-select v-model="dialog.form.difficulty" style="width:100%"><el-option label="简单" value="简单" /><el-option label="普通" value="普通" /><el-option label="困难" value="困难" /></el-select></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="标签"><el-input v-model="dialog.form.tags" placeholder="逗号分隔" /></el-form-item></el-col>
        </el-row>
        <el-form-item label="解析"><el-input v-model="dialog.form.analysis" type="textarea" :rows="2" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialog.visible=false">取消</el-button>
        <el-button type="primary" :loading="dialog.saving" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { adminApi } from '../../../api'

const keyword = ref('')
const typeFilter = ref('')
const categoryFilter = ref('')
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref<any[]>([])
const loading = ref(false)
const allTypes = ['input','textarea','number','select','radio','checkbox','date','time','phone','email','idCard']

function diffType(d: string) { return ({ '简单':'success', '普通':'primary', '困难':'danger' } as any)[d||''] || '' }

const dialog = reactive({ visible: false, isEdit: false, saving: false, editingId: 0, form: { type:'radio', title:'', options:'', answer:'', score:5, category:'', difficulty:'普通', tags:'', analysis:'' } })

function resetForm() { dialog.form = { type:'radio', title:'', options:'', answer:'', score:5, category:'', difficulty:'普通', tags:'', analysis:'' } }

async function load() {
  loading.value = true
  try {
    const res: any = await adminApi.examQuestionList({ page: page.value, pageSize: pageSize.value, keyword: keyword.value, type: typeFilter.value, category: categoryFilter.value })
    list.value = res.list || []; total.value = res.total || 0
  } finally { loading.value = false }
}

function openAdd() { resetForm(); dialog.isEdit = false; dialog.visible = true }
function openEdit(row: any) {
  dialog.isEdit = true; dialog.editingId = row.id
  dialog.form = { type: row.type, title: row.title, options: row.options||'', answer: row.answer||'', score: row.score, category: row.category||'', difficulty: row.difficulty||'普通', tags: row.tags||'', analysis: row.analysis||'' }
  dialog.visible = true
}

async function save() {
  dialog.saving = true
  try {
    const payload = { ...dialog.form, id: dialog.editingId }
    if (dialog.isEdit) await adminApi.examQuestionEdit(payload)
    else await adminApi.examQuestionInsert(payload)
    ElMessage.success(dialog.isEdit?'已更新':'已创建')
    dialog.visible = false; load()
  } catch (e: any) { ElMessage.error(e.msg || '保存失败') }
  finally { dialog.saving = false }
}

async function del(row: any) {
  await adminApi.examQuestionDel({ id: row.id })
  ElMessage.success('已删除'); load()
}

onMounted(load)
</script>

<style scoped>
.page { max-width:1400px; margin:0 auto; }
.page-banner { display:flex; justify-content:space-between; align-items:center; padding:24px 0 16px; }
.page-banner h2 { margin:0; font-size:22px; font-weight:600; color:#1a1a2e; }
.page-banner p { margin:4px 0 0; color:#888; font-size:13px; }
.main-card { border-radius:12px; }
.toolbar { display:flex; gap:8px; margin-bottom:16px; flex-wrap:wrap; }
.score-badge { background:#fff5f5; color:#fb454c; padding:2px 10px; border-radius:4px; font-weight:600; font-size:13px; }
.page-bar { text-align:center; margin-top:16px; }
</style>
