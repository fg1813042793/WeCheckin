<template>
  <div>
    <el-card>
      <div style="display:flex;gap:10px;margin-bottom:12px">
        <el-input v-model="keyword" placeholder="搜索管理员" clearable style="width:300px" @keyup.enter="search" />
        <el-button type="primary" @click="search">搜索</el-button>
      </div>
      <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:12px">
        <div>
          <el-button v-if="hasPerm('mgr:add')" type="success" @click="showAdd">+ 添加管理员</el-button>
          <el-button v-if="hasPerm('mgr:del')" type="danger" :disabled="selected.length === 0" @click="delSelected">批量删除</el-button>
        </div>
        <div>
          <el-button circle icon="Refresh" title="刷新" @click="load" />
          <el-button circle icon="Upload" title="导入" @click="ElMessage.info('导入功能开发中')" />
          <el-button circle icon="Download" title="导出" @click="exportData" />

        </div>
      </div>
      <el-table :data="list" v-loading="loading" stripe style="width:100%" @selection-change="selected = $event" :row-key="(r:any)=>r.id">
        <el-table-column type="selection" width="45" :selectable="(r:any)=>r.type!==1" />
        <el-table-column label="头像" width="60">
          <template #default="{ row }">
            <el-avatar :src="row.pic" size="small">{{ row.name?.[0] }}</el-avatar>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="姓名" width="300" />
        <el-table-column prop="desc" label="描述" min-width="100" />
        <el-table-column prop="phone" label="手机号" width="120" />
        <el-table-column label="角色" width="120">
          <template #default="{ row }">
            <el-tag :type="row.type === 1 ? 'danger' : 'info'" size="small">{{ row.roleName || (row.type === 1 ? '超级管理员' : '-') }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="所属部门" min-width="130">
          <template #default="{ row }">
            <span v-if="!row.deptIds || row.deptIds.length === 0" style="color:#999">-</span>
            <span v-else>{{ deptNames(row.deptIds) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="70">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">{{ row.status === 1 ? '正常' : '停用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="loginCnt" label="登录次数" width="80" />
        <el-table-column label="操作" width="300" fixed="right">
          <template #default="{ row }">
            <div class="table-actions">
              <el-button v-if="hasPerm('mgr:edit')" size="small" type="primary" @click="showEdit(row)">编辑</el-button>
              <template v-if="row.type !== 1">
                <el-button v-if="row.status === 1 && hasPerm('mgr:edit')" size="small" type="warning" @click="toggleStatus(row, 0)">停用</el-button>
                <el-button v-else-if="hasPerm('mgr:edit')" size="small" type="success" @click="toggleStatus(row, 1)">启用</el-button>
                <el-popconfirm v-if="hasPerm('mgr:del')" title="确定删除该管理员？" @confirm="remove(row)">
                  <template #reference>
                    <el-button size="small" type="danger">删除</el-button>
                  </template>
                </el-popconfirm>
              </template>
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
    <el-dialog v-model="formDialog.visible" :title="formDialog.title" width="500px" :close-on-click-modal="false">
      <el-form ref="formRef" :model="form" label-width="80px">
        <el-form-item label="头像">
          <el-upload action="/upload" :show-file-list="false" :on-success="handleAvatarSuccess" :headers="{ Authorization: token }" accept="image/*" style="display:inline-block">
            <el-avatar :src="form.pic" size="medium" style="cursor:pointer">{{ form.name?.[0] }}</el-avatar>
          </el-upload>
        </el-form-item>
        <el-form-item label="姓名" prop="name">
          <el-input v-model="form.name" placeholder="必填" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.desc" />
        </el-form-item>
        <el-form-item label="手机号">
          <el-input v-model="form.phone" />
        </el-form-item>
        <el-form-item label="绑定角色" v-if="form.type !== 1">
          <el-select v-model="form.roleId" placeholder="选择角色" clearable style="width:100%">
            <el-option v-for="r in roleList" :key="r.id" :label="r.name" :value="r.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="所属部门" v-if="form.type !== 1">
          <el-popover trigger="click" placement="bottom-start" :width="400">
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
        </el-form-item>
        <el-form-item v-if="formDialog.isCreate" label="密码" prop="password">
          <el-input v-model="form.password" type="password" placeholder="至少6位" />
        </el-form-item>
        <el-form-item v-else label="新密码">
          <el-input v-model="form.password" type="password" placeholder="留空不修改" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formDialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveMgr">{{ formDialog.isCreate ? '创建' : '保存' }}</el-button>
      </template>
    </el-dialog>

    <!-- 修改密码 -->
    <el-dialog v-model="pwdDialog.visible" title="修改密码" width="420px">
      <el-form ref="pwdRef" :model="pwdForm" label-width="100px">
        <el-form-item label="旧密码" prop="oldPassword">
          <el-input v-model="pwdForm.oldPassword" type="password" placeholder="必填" />
        </el-form-item>
        <el-form-item label="新密码" prop="password">
          <el-input v-model="pwdForm.password" type="password" placeholder="至少6位" />
        </el-form-item>
        <el-form-item label="确认密码" prop="password2">
          <el-input v-model="pwdForm.password2" type="password" placeholder="再次输入" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="pwdDialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="pwdLoading" @click="savePwd">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script lang="ts" setup>
import { ref, reactive, computed, onMounted, nextTick } from 'vue'
import { adminApi } from '../../api'
import { ElMessage, ElMessageBox } from 'element-plus'
import { hasPerm } from '../../utils/permission'

const loading = ref(false)
const list = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const saving = ref(false)
const roleList = ref<any[]>([])
const deptTreeData = ref<any[]>([])
const deptTreeRef = ref<any>(null)
const keyword = ref('')
const selected = ref<any[]>([])
const token = localStorage.getItem('admin_token') || ''

function handleAvatarSuccess(res: any) {
  if (res.code === 0) form.pic = res.data?.url || ''
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
function deptNames(ids: number[]) {
  if (!ids || ids.length === 0) return '-'
  const dmap = deptNameMap()
  return ids.map(id => dmap[id] || id).join('、')
}
const deptDisplayText = computed(() => {
  if (!form.deptIds || form.deptIds.length === 0) return ''
  return form.deptIds.map((id: number) => deptNameMap()[id] || id).join('、')
})

async function load() {
  loading.value = true
  try {
    const params: any = { page: page.value, pageSize: pageSize.value, keyword: keyword.value || undefined }
    const [mgrRes, roleRes, deptRes] = await Promise.all([
      adminApi.mgrList(params),
      adminApi.roleList({ page: 1, pageSize: 9999 }),
      adminApi.deptTree()
    ])
    list.value = mgrRes.data?.list || []
    total.value = mgrRes.data?.total || 0
    roleList.value = Array.isArray(roleRes.data?.list) ? roleRes.data.list : (Array.isArray(roleRes.data) ? roleRes.data : [])
    deptTreeData.value = Array.isArray(deptRes.data) ? deptRes.data : []
  } catch { list.value = []; total.value = 0; roleList.value = []; deptTreeData.value = [] }
  loading.value = false
}
function search() {
  page.value = 1
  load()
}

const formDialog = reactive({ visible: false, title: '', isCreate: false })
const form = reactive({ id: null as any, name: '', desc: '', phone: '', pic: '', password: '', type: 0, roleId: null as any, deptIds: [] as number[] })
function resetForm() {
  Object.assign(form, { id: null, name: '', desc: '', phone: '', pic: '', password: '', type: 0, roleId: null, deptIds: [] })
}
function showAdd() {
  resetForm()
  formDialog.isCreate = true
  formDialog.title = '添加管理员'
  formDialog.visible = true
}
async function showEdit(row: any) {
  resetForm()
  formDialog.isCreate = false
  formDialog.title = '编辑管理员'
  formDialog.visible = true
  form.id = row.id
  form.name = row.name
  form.desc = row.desc || ''
  form.phone = row.phone || ''
  form.pic = row.pic || ''
  form.type = row.type
  form.roleId = row.roleId || null
  form.deptIds = row.deptIds || []
  nextTick(() => deptTreeRef.value?.setCheckedKeys(form.deptIds))
}
function onDeptCheck() {
  nextTick(() => {
    form.deptIds = deptTreeRef.value?.getCheckedKeys() || []
  })
}
async function saveMgr() {
  if (!form.name) { ElMessage.warning('请输入姓名'); return }
  if (formDialog.isCreate && !form.password) { ElMessage.warning('请输入密码'); return }
  saving.value = true
  try {
    const payload: any = { name: form.name, desc: form.desc, phone: form.phone, roleId: form.roleId || 0, deptIds: form.deptIds.join(',') }
    if (formDialog.isCreate) {
      payload.password = form.password
      await adminApi.mgrInsert(payload)
    } else {
      payload.id = form.id
      payload.pic = form.pic
      if (form.password) payload.password = form.password
      await adminApi.mgrEdit(payload)
    }
    ElMessage.success('保存成功')
    formDialog.visible = false
    load()
  } finally { saving.value = false }
}

async function toggleStatus(row: any, status: number) {
  await adminApi.mgrStatus({ id: row.id, status })
  ElMessage.success(status === 1 ? '已启用' : '已停用')
  load()
}
async function remove(row: any) {
  if (row.type === 1) { ElMessage.warning('超级管理员不可删除'); return }
  await adminApi.mgrDel({ id: row.id })
  ElMessage.success('已删除')
  load()
}

async function delSelected() {
  if (selected.value.length === 0) return
  try {
    await ElMessageBox.confirm(`确定删除选中的 ${selected.value.length} 个管理员？`, '提示')
    const ids = selected.value.map((r: any) => r.id).join(',')
    await adminApi.mgrDels({ ids })
    ElMessage.success('已删除')
    selected.value = []
    load()
  } catch {}
}

function exportData() {
  const rows = [['用户名', '状态']]
  list.value.forEach((r: any) => {
    rows.push([r.name || '', r.status === 1 ? '正常' : '停用'])
  })
  const csv = '\uFEFF' + rows.map(r => r.map(v => '"' + String(v).replace(/"/g, '""') + '"').join(',')).join('\n')
  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
  const a = document.createElement('a')
  a.href = URL.createObjectURL(blob)
  a.download = '管理员列表.csv'
  a.click()
  URL.revokeObjectURL(a.href)
}

// 修改密码
const pwdDialog = reactive({ visible: false })
const pwdLoading = ref(false)
const pwdForm = reactive({ oldPassword: '', password: '', password2: '' })
function showPwd() {
  Object.assign(pwdForm, { oldPassword: '', password: '', password2: '' })
  pwdDialog.visible = true
}
async function savePwd() {
  if (!pwdForm.oldPassword) { ElMessage.warning('请输入旧密码'); return }
  if (!pwdForm.password) { ElMessage.warning('请输入新密码'); return }
  if (pwdForm.password !== pwdForm.password2) { ElMessage.warning('两次密码不一致'); return }
  if (pwdForm.password.length < 6) { ElMessage.warning('新密码至少6位'); return }
  pwdLoading.value = true
  try {
      await adminApi.mgrPwd({ oldPassword: pwdForm.oldPassword, password: pwdForm.password })
    ElMessage.success('密码修改成功')
    pwdDialog.visible = false
  } finally { pwdLoading.value = false }
}

onMounted(load)
</script>
