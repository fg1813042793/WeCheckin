<template>
  <div>
    <el-tabs v-model="activeTab">
      <el-tab-pane label="用户列表" name="list">
        <el-card>
          <div style="display:flex;gap:10px;margin-bottom:12px">
            <el-input v-model="keyword" placeholder="搜索用户名/姓名" clearable style="width:300px" @keyup.enter="search" />
            <el-button type="primary" @click="search">搜索</el-button>
          </div>
          <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:12px">
            <div>
              <el-button v-if="hasPerm('user:add')" type="success" @click="showAdd">+ 增加用户</el-button>
              <el-button v-if="hasPerm('user:del')" type="danger" :disabled="selected.length === 0" @click="delSelected">批量删除</el-button>
            </div>
            <div>
              <el-button circle icon="Refresh" title="刷新" @click="load" />
              <el-button circle icon="Upload" title="导入" @click="ElMessage.info('导入功能开发中')" />
              <el-button circle icon="Download" title="导出" @click="exportData" />
              <SortPopover :columns="sortColumns" v-model="sortRules" @change="onSortChange" />
            </div>
          </div>
          <el-table :data="list" v-loading="loading" stripe style="width:100%" @selection-change="selected = $event">
            <el-table-column type="selection" width="45" />
            <el-table-column label="头像" width="70">
              <template #default="{ row }">
                <el-avatar :src="row.avatar || row.pic" size="small">{{ row.name?.[0] }}</el-avatar>
              </template>
            </el-table-column>
            <el-table-column prop="name" label="用户名" width="120" />
            <el-table-column prop="mobile" label="手机号" width="130" />
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="statusType(row.status)" size="small">{{ statusLabel(row.status) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="loginCnt" label="登录次数" width="80" />
            <el-table-column label="所属部门" min-width="150">
              <template #default="{ row }">
                <span v-if="!row.deptNames">-</span>
                <span v-else>{{ row.deptNames }}</span>
              </template>
            </el-table-column>
            <el-table-column label="注册时间" width="160">
              <template #default="{ row }">{{ fmtTime(row.addTime) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="500" fixed="right">
              <template #default="{ row }">
                <div class="table-actions">
                  <el-button v-if="hasPerm('user:list')" size="small" @click="showDetail(row)">详情</el-button>
                  <el-button v-if="hasPerm('user:edit')" size="small" type="primary" @click="showEdit(row)">编辑</el-button>
                <template v-if="row.status === 0">
                  <el-button v-if="hasPerm('user:edit')" size="small" type="success" @click="changeStatus(row, '1')">审核通过</el-button>
                </template>
                <template v-else-if="row.status === 1">
                  <el-button v-if="hasPerm('user:edit')" size="small" type="warning" @click="showReason(row)">禁用</el-button>
                </template>
                <template v-else-if="row.status === 2">
                  <el-button v-if="hasPerm('user:edit')" size="small" type="success" @click="changeStatus(row, '1')">恢复正常</el-button>
                </template>
                  <el-button v-if="hasPerm('user:edit')" size="small" @click="resetPwd(row)">重置密码</el-button>
                  <el-popconfirm v-if="hasPerm('user:del')" title="确定删除该用户？" @confirm="remove(row)">
                    <template #reference>
                      <el-button size="small" type="danger">删除</el-button>
                    </template>
                  </el-popconfirm>
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
      </el-tab-pane>
    </el-tabs>

    <el-dialog v-model="dialog.visible" :title="dialog.title" width="600px" :close-on-click-modal="false">
      <el-form ref="formRef" :model="form" label-width="80px">
        <el-form-item label="头像" class="avatar-form-item">
          <el-upload action="/upload" :show-file-list="false" :on-success="handleAvatarSuccess" :on-error="()=>ElMessage.error('上传失败')" :headers="{ Authorization: token }" accept="image/*">
            <div class="avatar-upload">
              <el-avatar v-if="form.pic" :src="form.pic" size="large" />
              <div v-else class="avatar-placeholder">+</div>
            </div>
          </el-upload>
        </el-form-item>
        <el-form-item label="用户名" prop="name">
          <el-input v-model="form.name" placeholder="必填" />
        </el-form-item>
        <el-form-item label="手机号">
          <el-input v-model="form.mobile" placeholder="手机号" />
        </el-form-item>
        <el-form-item label="所属部门">
          <div style="width:100%;position:relative">
            <el-popover trigger="click" placement="bottom-start" :width="420" popper-style="margin-top:4px">
              <template #reference>
                <el-input readonly :model-value="deptDisplayText" placeholder="请选择部门" suffix-icon="ArrowDown" style="width:100%" />
              </template>
              <el-tree
                ref="deptTreeRef"
                :data="deptTreeData"
                :props="{ label: 'name' }"
                show-checkbox
                check-strictly
                node-key="id"
                :default-checked-keys="form.deptIds"
                @check="onDeptCheck"
                style="max-height:300px;overflow-y:auto"
              />
            </el-popover>
          </div>
        </el-form-item>
        <el-form-item v-for="f in formFields" :key="f._key" :label="f.label">
          <el-input v-if="f.type === '文本'" v-model="form.formsData[f._key]" />
          <el-input-number v-else-if="f.type === '数字'" v-model="form.formsData[f._key]" :min="0" />
          <el-input v-else-if="f.type === '多行文本'" v-model="form.formsData[f._key]" type="textarea" />
          <el-select v-else-if="f.type === '选择'" v-model="form.formsData[f._key]" style="width:100%">
            <el-option v-for="opt in (f.options||'').split(',').filter(Boolean)" :key="opt" :label="opt" :value="opt" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveUser">{{ dialog.isCreate ? '创建' : '保存' }}</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="reasonDialog.visible" title="请输入原因" width="400px">
      <el-input v-model="reasonDialog.reason" type="textarea" placeholder="禁用/审核不通过原因" />
      <template #footer>
        <el-button @click="reasonDialog.visible = false">取消</el-button>
        <el-button type="primary" @click="confirmReason">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailDialog.visible" title="用户详情" width="500px">
      <div v-if="detailDialog.detail">
        <div style="text-align:center;margin-bottom:16px">
          <el-avatar :src="detailDialog.detail.avatar || detailDialog.detail.pic" size="large">{{ detailDialog.detail.name?.[0] }}</el-avatar>
          <div style="margin-top:8px;font-size:18px;font-weight:600">{{ detailDialog.detail.name }}</div>
          <el-tag :type="statusType(detailDialog.detail.status)" size="small">{{ statusLabel(detailDialog.detail.status) }}</el-tag>
        </div>
        <el-descriptions :column="1" border>
                <el-descriptions-item label="手机号">{{ detailDialog.detail.mobile || '-' }}</el-descriptions-item>
          <el-descriptions-item label="所属部门">
            {{ deptNameStr(detailDialog.detail.deptIds) }}
          </el-descriptions-item>
          <el-descriptions-item label="注册时间">{{ fmtTime(detailDialog.detail.addTime) }}</el-descriptions-item>
          <el-descriptions-item label="最后登录">{{ fmtTime(detailDialog.detail.loginTime) }}</el-descriptions-item>
          <el-descriptions-item v-for="f in formFields" :key="f._key" :label="f.label">
            {{ parsedForms(f) }}
          </el-descriptions-item>
        </el-descriptions>
      </div>
    </el-dialog>
  </div>
</template>

<script lang="ts" setup>
import SortPopover from '../../components/SortPopover.vue'
import { ref, reactive, computed, onMounted, nextTick } from 'vue'
import { Delete } from '@element-plus/icons-vue'
import { adminApi } from '../../api'
import { ElMessage, ElMessageBox } from 'element-plus'
import { hasPerm } from '../../utils/permission'

const activeTab = ref('list')
const loading = ref(false)
const saving = ref(false)
const list = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const keyword = ref('')
const selected = ref<any[]>([])
const token = localStorage.getItem('admin_token') || ''
const sortRules = ref<{field:string;order:string}[]>([])
const sortColumns = [
  { label: '用户名', field: 'name' },
  { label: '手机号', field: 'mobile' },
  { label: '状态', field: 'status' },
  { label: '登录次数', field: 'loginCnt' },
  { label: '注册时间', field: 'addTime' },
]

const formFields = ref<any[]>([])
const deptTreeData = ref<any[]>([])
const deptTreeRef = ref<any>(null)
const deptDisplayText = computed(() => {
  if (!form.deptIds || form.deptIds.length === 0) return ''
  return form.deptIds.map((id: number) => deptTagName(id)).join('、')
})

function fmtTime(ts: number) {
  if (!ts) return '-'
  return new Date(ts).toLocaleString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}
function statusType(s: any) {
  const m: Record<string, string> = { '0': 'warning', '1': 'success', '2': 'danger', '9': 'info' }
  return m[String(s)] || 'info'
}
function statusLabel(s: any) {
  const m: Record<string, string> = { '0': '待审核', '1': '正常', '2': '禁用', '9': '管理员' }
  return m[String(s)] || String(s)
}

function deptNameMap() {
  const m: Record<number, string> = {}
  function walk(nodes: any[]) {
    for (const n of nodes) {
      m[n.id] = n.name
      if (n.children) walk(n.children)
    }
  }
  walk(deptTreeData.value)
  return m
}

function deptTagName(id: number) {
  return deptNameMap()[id] || id
}

function onDeptCheck() {
  nextTick(() => {
    form.deptIds = deptTreeRef.value?.getCheckedKeys() || []
  })
}

async function loadDepts() {
  try {
    const res = await adminApi.deptTree()
    deptTreeData.value = Array.isArray(res.data) ? res.data : []
  } catch { deptTreeData.value = [] }
}

async function load() {
  loading.value = true
  try {
    const params: any = { page: page.value, pageSize: pageSize.value, keyword: keyword.value }
    if (sortRules.value.length) params.sort = sortRules.value.map(s => s.field + ':' + s.order).join(',')
    const res = await adminApi.userList(params)
    const rawList = res.data?.list || []
    total.value = res.data?.total || 0
    const dmap = deptNameMap()
    list.value = rawList.map((u: any) => ({
      ...u,
      deptNames: (u.deptIds || []).map((id: number) => dmap[id] || id).filter(Boolean).join('、') || '-'
    }))
  } catch { list.value = []; total.value = 0 }
  loading.value = false
}
function onSortChange() { page.value = 1; load() }
function search() { page.value = 1; load() }

// 状态变更
const reasonDialog = reactive({ visible: false, reason: '', row: null as any, targetStatus: '' })
function showReason(row: any) {
  reasonDialog.row = row
  reasonDialog.targetStatus = '2'
  reasonDialog.reason = ''
  reasonDialog.visible = true
}
function showAuditFail(row: any) {
  reasonDialog.row = row
  reasonDialog.targetStatus = '2'
  reasonDialog.reason = ''
  reasonDialog.visible = true
}
async function confirmReason() {
  if (!reasonDialog.reason) { ElMessage.warning('请输入原因'); return }
  await changeStatus(reasonDialog.row, reasonDialog.targetStatus, reasonDialog.reason)
  reasonDialog.visible = false
}
async function changeStatus(row: any, status: string, reason?: string) {
  await adminApi.userStatus({ id: row.id, status, reason: reason || '' })
  ElMessage.success('操作成功')
  load()
}

async function remove(row: any) {
  await adminApi.userDel({ id: row.id })
  ElMessage.success('已删除')
  load()
}

async function delSelected() {
  if (selected.value.length === 0) return
  try {
    await ElMessageBox.confirm(`确定删除选中的 ${selected.value.length} 个用户？`, '提示')
    const ids = selected.value.map((r: any) => r.id).join(',')
    await adminApi.userDels({ ids })
    ElMessage.success('已删除')
    selected.value = []
    load()
  } catch {}
}

function exportData() {
  const rows = [['用户名', '姓名', '手机号', '部门', '状态']]
  list.value.forEach((r: any) => {
    rows.push([r.name || '', r.realName || '', r.phone || '', r.deptName || '', r.status === 1 ? '正常' : '禁用'])
  })
  const csv = '\uFEFF' + rows.map(r => r.map(v => '"' + String(v).replace(/"/g, '""') + '"').join(',')).join('\n')
  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
  const a = document.createElement('a')
  a.href = URL.createObjectURL(blob)
  a.download = '用户列表.csv'
  a.click()
  URL.revokeObjectURL(a.href)
}

async function resetPwd(row: any) {
  try {
    await ElMessageBox.confirm(`确定将用户「${row.name}」的密码重置为 123456？`, '提示')
    await adminApi.userResetPwd({ id: row.id })
    ElMessage.success('密码已重置为 123456')
  } catch {}
}

// 新增/编辑
const dialog = reactive({ visible: false, title: '', isCreate: false })
const form = reactive({ id: null as any, name: '', mobile: '', avatar: '', pic: '', deptIds: [] as number[], formsData: {} as any })
function resetForm() {
  form.id = null
  form.name = ''
  form.mobile = ''
  form.avatar = ''
  form.pic = ''
  form.deptIds = []
  form.formsData = {}
}

function showAdd() {
  resetForm()
  dialog.isCreate = true
  dialog.title = '增加用户'
  dialog.visible = true
}
async function showEdit(row: any) {
  resetForm()
  dialog.isCreate = false
  dialog.title = '编辑用户'
  dialog.visible = true
  form.id = row.id
  form.name = row.name
  form.mobile = row.mobile
  form.pic = row.avatar || row.pic
  form.avatar = row.avatar || ''
  form.deptIds = row.deptIds || []
  try {
    const res = await adminApi.userDetailById(row.id)
    const d = res.data || {}
    form.deptIds = d.deptIds || []
    if (d.forms) {
      try { form.formsData = JSON.parse(d.forms) } catch { form.formsData = {} }
    }
  } catch {}
}

function handleAvatarSuccess(res: any) {
  if (res.data?.url) form.pic = res.data.url
}

async function saveUser() {
  if (!form.name) { ElMessage.warning('请输入用户名'); return }
  saving.value = true
  try {
    const payload: any = { name: form.name, mobile: form.mobile, pic: form.pic }
    if (form.deptIds.length > 0) payload.deptIds = form.deptIds.join(',')
    const hasAny = Object.values(form.formsData).some(v => v !== undefined && v !== null && v !== '')
    if (hasAny) payload.forms = JSON.stringify(form.formsData)
    if (dialog.isCreate) {
      await adminApi.userAdd(payload)
    } else {
      payload.id = form.id
      await adminApi.userEdit(payload)
    }
    ElMessage.success(dialog.isCreate ? '创建成功' : '保存成功')
    dialog.visible = false
    load()
  } finally { saving.value = false }
}

// 详情
const detailDialog: { visible: boolean; detail: any } = reactive({ visible: false, detail: null })
async function showDetail(row: any) {
  try {
    const res = await adminApi.userDetailById(row.id)
    detailDialog.detail = res.data || {}
    detailDialog.visible = true
  } catch {}
}
function deptNameStr(deptIds: number[]) {
  if (!deptIds || deptIds.length === 0) return '-'
  const dmap = deptNameMap()
  return deptIds.map(id => dmap[id] || id).join('、')
}
function parsedForms(f: any) {
  const d: any = detailDialog.detail
  if (!d || !d.forms) return '-'
  try {
    const forms = JSON.parse(d.forms)
    return forms[String(f.id)] || '-'
  } catch { return '-' }
}

async function loadFormFields() {
  try {
    const res = await adminApi.userFormFields()
    formFields.value = (res.data || []).map((f: any, i: number) => ({ ...f, _key: String(f.id || i) }))
  } catch { formFields.value = [] }
}

onMounted(() => {
  load()
  loadFormFields()
  loadDepts()
})
</script>

<style scoped>
.avatar-upload {
  cursor: pointer;
  display: inline-block;
}
.avatar-placeholder {
  width: 56px;
  height: 56px;
  border: 1px dashed #d9d9d9;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  color: #999;
  transition: border-color 0.3s;
}
.avatar-placeholder:hover {
  border-color: #409eff;
  color: #409eff;
}
.avatar-form-item :deep(.el-form-item__content) {
  justify-content: flex-start !important;
  display: flex !important;
}
.avatar-form-item {
  display: flex !important;
  align-items: center !important;
  flex-wrap: nowrap !important;
}
</style>
