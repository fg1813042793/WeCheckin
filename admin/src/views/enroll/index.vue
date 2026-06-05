<template>
  <div>
    <el-card>
      <div style="display:flex;gap:10px;margin-bottom:12px">
        <el-input v-model="keyword" placeholder="搜索打卡标题" clearable style="width:300px" @keyup.enter="search" />
        <el-button type="primary" @click="search">搜索</el-button>
      </div>
      <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:12px">
        <div>
          <el-button v-if="hasPerm('enroll:add')" type="success" @click="showAdd">+ 新增打卡</el-button>
          <el-button v-if="hasPerm('enroll:del')" type="danger" :disabled="selected.length === 0" @click="delSelected">批量删除</el-button>
        </div>
        <div class="toolbar-icons">
          <el-button circle icon="Refresh" title="刷新" @click="load" />
          <el-button circle icon="Upload" title="导入" @click="ElMessage.info('导入功能开发中')" />
          <el-button circle icon="Download" title="导出" @click="exportData" />
          <SortPopover :columns="sortColumns" v-model="sortRules" @change="onSortChange" />
        </div>
      </div>
      <el-table :data="list" v-loading="loading" stripe style="width:100%" @selection-change="selected = $event">
        <el-table-column type="selection" width="45" />
        <el-table-column prop="title" label="标题" min-width="160" />
        <el-table-column prop="cateName" label="分类" width="100" />
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">{{ row.status === 1 ? '正常' : '停用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="时间范围" min-width="200">
          <template #default="{ row }">{{ fmtDate(row.timeStart) }} ~ {{ fmtDate(row.timeEnd) }}</template>
        </el-table-column>
        <el-table-column prop="userCnt" label="参与人数" width="80" />
        <el-table-column prop="joinCount" label="打卡数" width="80" />
        <el-table-column label="操作" width="400" fixed="right">
          <template #default="{ row }">
            <div class="table-actions">
              <el-button v-if="hasPerm('enroll:edit')" size="small" type="primary" @click="showEdit(row)">编辑</el-button>
              <el-button v-if="hasPerm('enroll:list')" size="small" @click="showJoins(row)">打卡记录</el-button>
              <el-button v-if="hasPerm('enroll:list')" size="small" @click="showUsers(row)">参与用户</el-button>
              <el-button v-if="hasPerm('enroll:list')" size="small" @click="showStats(row)">统计</el-button>
              <el-dropdown v-if="hasPerm('enroll:edit')" trigger="click" @command="(cmd:string)=>handleMore(cmd,row)">
                <el-button size="small">更多<el-icon><ArrowDown /></el-icon></el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                  <el-dropdown-item command="enable" :disabled="row.status===1">启用</el-dropdown-item>
                  <el-dropdown-item command="disable" :disabled="row.status===0">停用</el-dropdown-item>
                    <el-dropdown-item :command="row.vouch ? 'unvouch' : 'vouch'">{{ row.vouch ? '取消推荐' : '推荐首页' }}</el-dropdown-item>
                    <el-dropdown-item v-if="hasPerm('enroll:del')" command="del" divided>删除</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </template>
        </el-table-column>
      </el-table>
      <div style="text-align:center;margin-top:16px">
        <el-pagination
          v-model:current-page="page"
          :page-size="pageSize"
          :page-sizes="[10,20,50,100]"
          :total="total"
          layout="total,sizes,prev,pager,next"
          @current-change="load"
          @size-change="(val:number) => { pageSize = val; page = 1; load() }"
        />
      </div>
    </el-card>

    <!-- 新增/编辑弹窗 -->
    <el-dialog v-model="formDialog.visible" :title="formDialog.title" width="700px" :close-on-click-modal="false" destroy-on-close>
      <el-form ref="formRef" :model="form" label-width="120px">
        <el-form-item label="标题" prop="title">
          <el-input v-model="form.title" placeholder="必填" />
        </el-form-item>
        <el-form-item label="封面">
          <el-upload action="/upload" :show-file-list="false" :on-success="handleCoverSuccess" :on-error="()=>ElMessage.error('上传失败')" :headers="{ Authorization: token }" accept="image/*">
            <div class="cover-upload">
              <el-image v-if="form.cover" :src="form.cover" class="cover-preview" />
              <div v-else class="cover-placeholder">+</div>
              <div v-if="form.cover" class="cover-overlay" @click.stop>
                <el-button size="small" type="danger" :icon="Delete" circle @click.stop="form.cover=''" />
              </div>
            </div>
          </el-upload>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.desc" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="分类">
          <el-select v-model="form.cateName" style="width:200px">
            <el-option v-for="c in categories" :key="c.value" :label="c.label" :value="c.label" />
          </el-select>
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort" :min="0" />
        </el-form-item>
        <el-form-item label="开始时间">
          <el-date-picker v-model="form.startTime" type="date" value-format="YYYY-MM-DD" style="width:200px" />
        </el-form-item>
        <el-form-item label="结束时间">
          <el-date-picker v-model="form.endTime" type="date" value-format="YYYY-MM-DD" style="width:200px" />
        </el-form-item>
        <el-form-item label="重复打卡">
          <el-switch v-model="form.allowRepeat" :active-value="'1'" :inactive-value="'0'" />
        </el-form-item>
        <el-form-item label="每日上限">
          <el-input-number v-model="form.dailyLimit" :min="1" />
        </el-form-item>
        <el-form-item label="发布部门">
          <el-popover trigger="click" placement="bottom" :width="260" popper-style="padding:0">
            <template #reference>
              <div class="multi-select-input">
                <el-tag v-for="id in (form.publishDeptIds ? form.publishDeptIds.split(',').map(Number) : [])" :key="id" size="small" closable @close.stop="form.publishDeptIds=form.publishDeptIds.split(',').map(Number).filter((k:any)=>k!==id).join(',')">{{ getDeptPath(id) }}</el-tag>
                <span class="ms-placeholder" v-if="!form.publishDeptIds">选择发布部门</span>
                <el-icon class="ms-arrow"><ArrowDown /></el-icon>
              </div>
            </template>
            <el-tree
              ref="deptTreeRef"
              :data="deptTree"
              :props="{ label: 'name', children: 'children' }"
              node-key="id"
              show-checkbox
              check-strictly
              :default-checked-keys="form.publishDeptIds ? form.publishDeptIds.split(',').map(Number) : []"
              @check="(data:any,{checkedKeys}:any)=>{form.publishDeptIds=checkedKeys.filter((k:any)=>k!==0).join(',')}"
            />
          </el-popover>
        </el-form-item>
        <el-form-item label="报名表单字段">
          <div class="multi-select-input" @click="showFormEditor('enrollForms')">
            <el-tag v-for="(f, i) in form.enrollForms || []" :key="i" size="small" closable @close.stop="form.enrollForms.splice(i,1)">{{ f.label || '字段' + (i+1) }}</el-tag>
            <span class="ms-placeholder" v-if="!form.enrollForms || form.enrollForms.length === 0">点击配置报名表单字段</span>
            <el-icon class="ms-arrow"><ArrowDown /></el-icon>
          </div>
        </el-form-item>
        <el-form-item label="打卡表单字段">
          <div class="multi-select-input" @click="showFormEditor('joinForms')">
            <el-tag v-for="(f, i) in form.joinForms || []" :key="i" size="small" closable @close.stop="form.joinForms.splice(i,1)">{{ f.label || '字段' + (i+1) }}</el-tag>
            <span class="ms-placeholder" v-if="!form.joinForms || form.joinForms.length === 0">点击配置打卡表单字段</span>
            <el-icon class="ms-arrow"><ArrowDown /></el-icon>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formDialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveEnroll">{{ formDialog.isCreate ? '创建' : '保存' }}</el-button>
      </template>
    </el-dialog>

    <!-- 表单字段配置弹窗 -->
    <el-dialog v-model="formEditor.visible" :title="'配置表单字段'" width="600px">
      <div v-for="(f, i) in formEditor.fields" :key="i" style="display:flex;gap:8px;align-items:center;margin-bottom:8px">
        <span style="color:#999">{{ i + 1 }}.</span>
        <el-input v-model="f.label" placeholder="字段名称" style="width:140px" />
        <el-select v-model="f.type" style="width:120px">
          <el-option label="文本" value="text" />
          <el-option label="数字" value="number" />
          <el-option label="多行文本" value="textarea" />
          <el-option label="选择" value="select" />
          <el-option label="拍照上传" value="image" />
          <el-option label="位置签到" value="location" />
        </el-select>
        <el-input v-if="f.type==='select'" v-model="f.options" placeholder="选项(逗号分隔)" style="width:160px" />
        <el-checkbox v-model="f.required">必填</el-checkbox>
        <el-button type="danger" :icon="Delete" circle size="small" @click="formEditor.fields.splice(i,1)" />
      </div>
      <el-button @click="formEditor.fields.push({label:'',type:'text',options:'',required:false})">+ 添加字段</el-button>
      <template #footer>
        <el-button @click="confirmFormEditor">完成</el-button>
      </template>
    </el-dialog>

    <!-- 打卡记录 -->
    <el-dialog v-model="joinDialog.visible" :title="joinDialog.title" width="1200px">
      <div style="display:flex;gap:10px;margin-bottom:12px">
        <el-input v-model="joinKeyword" placeholder="搜索用户名/ID" clearable style="width:240px" @keyup.enter="searchJoins" />
        <el-button type="primary" @click="searchJoins">搜索</el-button>
      </div>
      <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:12px">
        <div>
          <el-button type="danger" :disabled="joinSelected.length === 0" @click="delSelectedJoins">批量删除</el-button>
        </div>
        <div class="toolbar-icons">
          <el-date-picker v-model="joinExportRange" type="daterange" range-separator="~" start-placeholder="开始日期" end-placeholder="结束日期" value-format="YYYY-MM-DD" style="width:240px" />
          <el-button type="primary" size="small" @click="exportJoins">导出</el-button>
        </div>
      </div>
      <el-table :data="joinList" v-loading="joinLoading" stripe @selection-change="joinSelected = $event">
        <el-table-column type="selection" width="45" />
        <el-table-column label="用户" min-width="120">
          <template #default="{ row }">{{ row.userName || row.enrollTitle || row.userId }}</template>
        </el-table-column>
        <el-table-column label="部门" min-width="100">
          <template #default="{ row }">{{ row.deptName || '-' }}</template>
        </el-table-column>
        <el-table-column label="顶级部门" min-width="100">
          <template #default="{ row }">{{ row.topDeptName || '-' }}</template>
        </el-table-column>
        <el-table-column label="打卡时间" width="150">
          <template #default="{ row }">{{ fmtTime(row._createTime) }}</template>
        </el-table-column>
        <el-table-column label="内容" min-width="240">
          <template #default="{ row }">
            <div v-if="row._forms && row._forms.length > 0">
              <div v-for="(f, i) in row._forms" :key="i" style="line-height:1.6">
                <span style="color:#909399;margin-right:4px">{{ f.label }}:</span>
                <template v-if="isImageUrl(f.value)">
                  <el-image :src="f.value" style="width:60px;height:60px;border-radius:4px" fit="cover" :preview-src-list="[f.value]" preview-teleported />
                </template>
                <span v-else>{{ f.value || '-' }}</span>
              </div>
            </div>
            <span v-else>{{ formatForms(row.forms) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="80">
          <template #default="{ row }">
            <el-popconfirm title="确定删除？" @confirm="delJoin(row)">
              <template #reference>
                <el-button size="small" type="danger">删除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
      <div style="text-align:center;margin-top:16px">
        <el-pagination v-model:current-page="joinPage" :page-size="joinPageSize" :page-sizes="[10,20,50,100]" :total="joinTotal" layout="total,sizes,prev,pager,next" @current-change="loadJoins" @size-change="(val:number) => { joinPageSize = val; joinPage = 1; loadJoins() }" />
      </div>
    </el-dialog>

    <!-- 参与用户 -->
    <el-dialog v-model="userDialog.visible" :title="userDialog.title" width="1200px">
      <div style="display:flex;gap:10px;margin-bottom:12px">
        <el-input v-model="userKeyword" placeholder="搜索用户名" clearable style="width:240px" @keyup.enter="searchUsers" />
        <el-button type="primary" @click="searchUsers">搜索</el-button>
      </div>
      <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:12px">
        <div>
          <el-button type="danger" :disabled="userSelected.length === 0" @click="delSelectedUsers">批量移除</el-button>
        </div>
        <div class="toolbar-icons">
          <el-button circle icon="Download" title="导出 CSV" @click="exportUsers" />
        </div>
      </div>
      <el-table :data="userList" v-loading="userLoading" stripe @selection-change="userSelected = $event">
        <el-table-column type="selection" width="45" />
        <el-table-column label="用户" min-width="120">
          <template #default="{ row }">{{ row.userName || row.title || row.miniOpenId || row.user_id }}</template>
        </el-table-column>
        <el-table-column label="部门" min-width="100">
          <template #default="{ row }">{{ row.deptName || '-' }}</template>
        </el-table-column>
        <el-table-column label="顶级部门" min-width="100">
          <template #default="{ row }">{{ row.topDeptName || '-' }}</template>
        </el-table-column>
        <el-table-column prop="joinCnt" label="打卡次数" width="80" />
        <el-table-column prop="dayCnt" label="打卡天数" width="80" />
        <el-table-column prop="lastDay" label="最后打卡" width="110" />
        <el-table-column label="参与时间" width="170">
          <template #default="{ row }">{{ fmtTime(row._createTime) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="100">
          <template #default="{ row }">
            <el-popconfirm title="确定移除该用户？" @confirm="removeUser(row)">
              <template #reference>
                <el-button size="small" type="danger">移除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <!-- 统计 -->
    <el-dialog v-model="statsDialog.visible" :title="statsDialog.title" width="700px">
      <div style="display:flex;gap:10px;margin-bottom:12px">
        <el-date-picker v-model="statsRange" type="daterange" range-separator="~" start-placeholder="开始日期" end-placeholder="结束日期" value-format="YYYY-MM-DD" style="width:240px" />
        <el-button type="primary" @click="loadStats">统计</el-button>
      </div>
      <el-table :data="statsList" v-loading="statsLoading" stripe>
        <el-table-column label="用户" min-width="120">
          <template #default="{ row }">{{ row.userName || row.userId }}</template>
        </el-table-column>
        <el-table-column label="部门" min-width="100">
          <template #default="{ row }">{{ row.deptName || '-' }}</template>
        </el-table-column>
        <el-table-column label="顶级部门" min-width="100">
          <template #default="{ row }">{{ row.topDeptName || '-' }}</template>
        </el-table-column>
        <el-table-column prop="joinCnt" label="打卡次数" width="90" />
        <el-table-column prop="dayCnt" label="打卡天数" width="90" />
      </el-table>
    </el-dialog>
  </div>
</template>

<script lang="ts" setup>
import SortPopover from '../../components/SortPopover.vue'
import { ref, reactive, onMounted } from 'vue'
import { Delete, ArrowDown } from '@element-plus/icons-vue'
import { adminApi } from '../../api'
import { ElMessage, ElMessageBox } from 'element-plus'
import { hasPerm } from '../../utils/permission'

const loading = ref(false)
const list = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const keyword = ref('')
const saving = ref(false)
const categories = ref<any[]>([])
const deptTree = ref<any[]>([])
const deptTreeRef = ref()
const token = localStorage.getItem('admin_token') || ''
const selected = ref<any[]>([])
const sortRules = ref<{field:string;order:string}[]>([])
const sortColumns = [
  { label: '标题', field: 'title' },
  { label: '排序', field: 'sort' },
  { label: '状态', field: 'status' },
  { label: '是否推荐', field: 'isVouch' },
  { label: '报名人数', field: 'userCnt' },
  { label: '签到人数', field: 'joinCnt' },
  { label: '创建时间', field: 'addTime' },
]

function handleCoverSuccess(res: any) {
  if (res.data?.url) form.cover = res.data.url
}

function fmtDate(ts: number) {
  if (!ts) return '-'
  return new Date(ts).toLocaleDateString('zh-CN')
}
function formatDate(ts: number) {
  if (!ts) return ''
  const d = new Date(ts)
  return d.getFullYear() + '-' + String(d.getMonth() + 1).padStart(2, '0') + '-' + String(d.getDate()).padStart(2, '0')
}

async function load() {
  loading.value = true
  try {
    const params: any = { page: page.value, pageSize: pageSize.value, keyword: keyword.value }
    if (sortRules.value.length) params.sort = sortRules.value.map(s => s.field + ':' + s.order).join(',')
    const res = await adminApi.enrollList(params)
    list.value = res.data?.list || []
    total.value = res.data?.total || 0
  } catch { list.value = []; total.value = 0 }
  loading.value = false
}
function onSortChange() { page.value = 1; load() }
function search() { page.value = 1; load() }

const formDialog = reactive({ visible: false, title: '', isCreate: false })
const form = reactive({
  id: null as any, title: '', cover: '', desc: '', cateName: '', sort: 9999,
  startTime: '', endTime: '', allowRepeat: '0', dailyLimit: 1,
  enrollForms: [] as any[], joinForms: [] as any[],
  publishDeptIds: ''
})
function resetForm() {
  Object.assign(form, {
    id: null, title: '', cover: '', desc: '', cateName: '', sort: 9999,
    startTime: '', endTime: '', allowRepeat: '0', dailyLimit: 1,
    enrollForms: [], joinForms: [], publishDeptIds: ''
  })
}
function showAdd() {
  resetForm()
  formDialog.isCreate = true
  formDialog.title = '新增打卡项目'
  formDialog.visible = true
}
async function showEdit(row: any) {
  resetForm()
  formDialog.isCreate = false
  formDialog.title = '编辑打卡项目'
  try {
    const res = await adminApi.enrollDetail(row.id)
    const d = res.data || {}
    form.id = d.id
    form.title = d.title || ''
    form.cover = d.img || ''
    form.desc = d.desc || ''
    form.cateName = d.cateName || ''
    form.sort = d.order ?? 9999
    form.allowRepeat = d.allowRepeat ? '1' : '0'
    form.dailyLimit = d.dailyLimit || 1
    form.publishDeptIds = d.publishDeptIds || ''
    if (d.timeStart) form.startTime = formatDate(d.timeStart)
    if (d.timeEnd) form.endTime = formatDate(d.timeEnd)
    if (d.forms) {
      try { form.enrollForms = JSON.parse(d.forms) } catch { form.enrollForms = [] }
    }
    if (d.joinForms) {
      try { form.joinForms = JSON.parse(d.joinForms) } catch { form.joinForms = [] }
    }
    formDialog.visible = true
  } catch {}
}
async function saveEnroll() {
  if (!form.title) { ElMessage.warning('请输入标题'); return }
  saving.value = true
  try {
    const payload: any = {
      title: form.title, cover: form.cover, desc: form.desc, cateName: form.cateName,
      sort: form.sort, startTime: form.startTime, endTime: form.endTime,
      allowRepeat: form.allowRepeat, dailyLimit: form.dailyLimit,
      enrollForms: JSON.stringify(form.enrollForms), joinForms: JSON.stringify(form.joinForms),
      deptId: 0, publishDeptIds: form.publishDeptIds
    }
    if (formDialog.isCreate) {
      await adminApi.enrollInsert(payload)
    } else {
      payload.id = form.id
      await adminApi.enrollEdit(payload)
    }
    ElMessage.success('保存成功')
    formDialog.visible = false
    load()
  } finally { saving.value = false }
}

// 表单字段编辑
const formEditor = reactive({ visible: false, fields: [] as any[], target: '' })
function showFormEditor(target: string) {
  formEditor.target = target
  formEditor.fields = JSON.parse(JSON.stringify(form[target as keyof typeof form] || []))
  formEditor.visible = true
}

function confirmFormEditor() {
  const key = formEditor.target as keyof typeof form
  ;(form as any)[key] = JSON.parse(JSON.stringify(formEditor.fields))
  formEditor.visible = false
}

// 更多操作
async function handleMore(cmd: string, row: any) {
  if (cmd === 'enable') {
    await adminApi.enrollStatus({ id: row.id, status: '1' })
    ElMessage.success('已启用'); load()
  } else if (cmd === 'disable') {
    await adminApi.enrollStatus({ id: row.id, status: '0' })
    ElMessage.success('已停用'); load()
  } else if (cmd === 'del') {
    try {
      await ElMessageBox.confirm('确定删除该打卡项目？', '提示')
      await adminApi.enrollDel({ id: row.id })
      ElMessage.success('已删除'); load()
    } catch {}
  } else if (cmd === 'vouch') {
    await adminApi.enrollVouch({ id: row.id, vouch: '1' })
    ElMessage.success('已推荐到首页'); load()
  } else if (cmd === 'unvouch') {
    await adminApi.enrollVouch({ id: row.id, vouch: '0' })
    ElMessage.success('已取消推荐'); load()
  }
}

async function delSelected() {
  if (selected.value.length === 0) return
  try {
    await ElMessageBox.confirm(`确定删除选中的 ${selected.value.length} 条打卡项目？`, '提示')
    const ids = selected.value.map((r: any) => r.id).join(',')
    await adminApi.enrollDels({ ids })
    ElMessage.success('已删除')
    selected.value = []
    load()
  } catch {}
}

function exportData() {
  const rows = [['标题', '分类', '状态', '开始时间', '结束时间', '参与人数', '打卡数']]
  list.value.forEach((r: any) => {
    rows.push([r.title, r.cateName || '', r.status === 1 ? '正常' : '停用', fmtDate(r.regStart), fmtDate(r.regEnd), String(r.userCnt || 0), String(r.joinCount || 0)])
  })
  const csv = '\uFEFF' + rows.map(r => r.map(v => '"' + String(v).replace(/"/g, '""') + '"').join(',')).join('\n')
  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
  const a = document.createElement('a')
  a.href = URL.createObjectURL(blob)
  a.download = '打卡列表.csv'
  a.click()
  URL.revokeObjectURL(a.href)
}

// 打卡记录
const joinDialog = reactive({ visible: false, title: '' })
const joinList = ref<any[]>([])
const joinTotal = ref(0)
const joinPage = ref(1)
const joinPageSize = ref(20)
const joinLoading = ref(false)
const joinSelected = ref<any[]>([])
let joinEnrollId = 0
const joinKeyword = ref('')
const joinExportRange = ref<[string, string] | null>(null)
async function showJoins(row: any) {
  joinEnrollId = row.id
  joinDialog.title = '打卡记录 - ' + row.title
  joinKeyword.value = ''
  joinPage.value = 1
  joinDialog.visible = true
  await loadJoins()
}
function searchJoins() { joinPage.value = 1; loadJoins() }
async function loadJoins() {
  joinLoading.value = true
  try {
    const res = await adminApi.enrollJoinList({ enrollId: joinEnrollId, page: joinPage.value, pageSize: joinPageSize.value, keyword: joinKeyword.value })
    const list = res.data?.list || []
    joinList.value = list.map((r: any) => {
      let _forms: any[] = []
      try {
        const obj = typeof r.forms === 'string' ? JSON.parse(r.forms) : r.forms
        if (Array.isArray(obj)) _forms = obj
        else _forms = Object.entries(obj).map(([k, v]) => ({ label: k, value: v }))
      } catch {}
      return { ...r, _forms }
    })
    joinTotal.value = res.data?.total || 0
  } catch { joinList.value = []; joinTotal.value = 0 }
  joinLoading.value = false
}
function formatForms(forms: any) {
  if (!forms) return '-'
  try {
    const obj = typeof forms === 'string' ? JSON.parse(forms) : forms
    if (Array.isArray(obj)) {
      return obj.map((f: any) => {
        if (f.label && f.value !== undefined && f.value !== null && f.value !== '') return `${f.label}: ${f.value}`
        return f.value || f.label || ''
      }).filter(Boolean).join(' / ')
    }
    return Object.values(obj).filter(Boolean).join(' / ')
  } catch { return String(forms) }
}

function isImageUrl(val: any): boolean {
  if (!val || typeof val !== 'string') return false
  return /\.(jpg|jpeg|png|gif|webp|bmp)(\?.*)?$/i.test(val) || val.includes('/uploads/') || val.startsWith('http')
}
async function delJoin(row: any) {
  await adminApi.enrollJoinDel({ enrollJoinId: row.id })
  ElMessage.success('已删除')
  loadJoins(); load()
}

async function delSelectedJoins() {
  if (joinSelected.value.length === 0) return
  try {
    await ElMessageBox.confirm(`确定删除选中的 ${joinSelected.value.length} 条打卡记录？`, '提示')
    const ids = joinSelected.value.map((r: any) => r.id).join(',')
    await adminApi.enrollJoinDels({ ids })
    ElMessage.success('已删除')
    joinSelected.value = []
    loadJoins(); load()
  } catch {}
}

async function exportJoins() {
  try {
    const params: any = { enrollId: joinEnrollId }
    if (joinExportRange.value) {
      params.startTime = joinExportRange.value[0]
      params.endTime = joinExportRange.value[1]
    }
    const res = await adminApi.enrollJoinDataExport(params)
    if (res.data) {
      window.open(res.data, '_blank')
    }
  } catch { ElMessage.error('导出失败') }
}

// 参与用户
const userDialog = reactive({ visible: false, title: '' })
const userList = ref<any[]>([])
const userLoading = ref(false)
const userSelected = ref<any[]>([])
const userKeyword = ref('')
let userEnrollId = 0
function searchUsers() { loadUsers() }
async function loadUsers() {
  userLoading.value = true
  try {
    const res = await adminApi.enrollUserList({ enrollId: userEnrollId, keyword: userKeyword.value })
    userList.value = res.data || []
  } catch { userList.value = [] }
  userLoading.value = false
}
async function showUsers(row: any) {
  userEnrollId = row.id
  userKeyword.value = ''
  userDialog.title = '参与用户 - ' + row.title
  userDialog.visible = true
  loadUsers()
}
async function removeUser(row: any) {
  await adminApi.enrollRemoveUser({ enrollId: userEnrollId, userId: row.miniOpenId })
  ElMessage.success('已移除')
  loadUsers()
}
async function delSelectedUsers() {
  if (userSelected.value.length === 0) return
  try {
    await ElMessageBox.confirm(`确定移除选中的 ${userSelected.value.length} 个用户？`, '提示')
    const userIds = userSelected.value.map((r: any) => r.miniOpenId).join(',')
    await adminApi.enrollRemoveUsers({ enrollId: userEnrollId, userIds })
    ElMessage.success('已移除')
    userSelected.value = []
    loadUsers()
  } catch {}
}
function exportUsers() {
  const rows = [['用户', '部门', '顶级部门', '打卡次数', '打卡天数', '最后打卡', '参与时间']]
  userList.value.forEach((r: any) => {
    rows.push([
      r.userName || r.title || r.miniOpenId || '',
      r.deptName || '',
      r.topDeptName || '',
      String(r.joinCnt || 0),
      String(r.dayCnt || 0),
      r.lastDay || '',
      fmtTime(r._createTime),
    ])
  })
  const csv = '\uFEFF' + rows.map(r => r.map(v => '"' + String(v).replace(/"/g, '""') + '"').join(',')).join('\n')
  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
  const a = document.createElement('a')
  a.href = URL.createObjectURL(blob)
  a.download = (userDialog.title.replace('参与用户 - ', '') || '参与用户') + '.csv'
  a.click()
  URL.revokeObjectURL(a.href)
}

function getDeptPath(id: number): string {
  const walk = (nodes: any[], parents: string[]): string => {
    for (const n of nodes) {
      const path = [...parents, n.name]
      if (n.id === id) return path.join('/')
      if (n.children) { const r = walk(n.children, path); if (r) return r }
    }
    return ''
  }
  return walk(deptTree.value, []) || ''
}
function getDeptNames(ids: string): string {
  if (!ids) return '选择部门'
  const names = ids.split(',').map(id => getDeptPath(Number(id))).filter(Boolean)
  return names.length ? names.join(', ') : '选择部门'
}

// 统计
const statsDialog = reactive({ visible: false, title: '' })
const statsList = ref<any[]>([])
const statsLoading = ref(false)
const statsRange = ref<[string, string] | null>(null)
let statsEnrollId = 0
function showStats(row: any) {
  statsEnrollId = row.id
  statsDialog.title = '统计 - ' + row.title
  statsRange.value = null
  statsDialog.visible = true
  loadStats()
}
async function loadStats() {
  statsLoading.value = true
  try {
    const params: any = { enrollId: statsEnrollId }
    if (statsRange.value) {
      params.startTime = statsRange.value[0]
      params.endTime = statsRange.value[1]
    }
    const res = await adminApi.enrollStats(params)
    statsList.value = res.data || []
  } catch { statsList.value = [] }
  statsLoading.value = false
}

function fmtTime(ts: number) {
  if (!ts) return '-'
  return new Date(ts).toLocaleString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

async function loadCategories() {
  try {
    const res = await adminApi.dictItems('check_type')
    categories.value = (res.data || []).map((d: any) => ({ label: d.label, value: d.value }))
  } catch { categories.value = [] }
}

async function loadDeptTree() {
  try {
    const res = await adminApi.deptTree()
    deptTree.value = res.data || []
  } catch { deptTree.value = [] }
}

onMounted(() => { load(); loadCategories(); loadDeptTree() })
</script>

<style scoped>
.cover-upload {
  position: relative;
  width: 100px;
  height: 100px;
  border: 1px dashed #d9d9d9;
  border-radius: 6px;
  cursor: pointer;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: border-color 0.3s;
}
.cover-upload:hover {
  border-color: #409eff;
}
.cover-placeholder {
  font-size: 32px;
  color: #999;
  line-height: 100px;
  text-align: center;
}
.cover-preview {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.cover-overlay {
  position: absolute;
  top: 0;
  right: 0;
  padding: 4px;
}
.multi-select-input {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 4px;
  min-height: 32px;
  width: 100%;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  padding: 2px 8px;
  cursor: pointer;
  transition: border-color 0.2s;
  background: #fff;
  box-sizing: border-box;
}
.multi-select-input:hover {
  border-color: #409eff;
}
.ms-placeholder {
  color: #c0c4cc;
  font-size: 14px;
  flex: 1;
}
.ms-arrow {
  color: #c0c4cc;
  font-size: 12px;
  flex-shrink: 0;
  margin-left: auto;
}
.toolbar-icons {
  display: flex;
  align-items: center;
}
.toolbar-icons > * {
  margin-left: 8px;
}
.toolbar-icons > :first-child {
  margin-left: 0;
}
</style>
<style>
.el-tree { text-align: left; }
</style>
