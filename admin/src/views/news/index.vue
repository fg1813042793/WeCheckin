<template>
  <div>
    <el-card>
      <div style="display:flex;gap:10px;margin-bottom:12px">
        <el-input v-model="keyword" placeholder="搜索标题" clearable style="width:300px" @keyup.enter="search" />
        <el-button type="primary" @click="search">搜索</el-button>
      </div>
      <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:12px">
        <div>
          <el-button v-if="hasPerm('news:add')" type="success" @click="showAdd">+ 添加通知</el-button>
          <el-button v-if="hasPerm('news:del')" type="danger" :disabled="selected.length === 0" @click="delSelected">批量删除</el-button>
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
        <el-table-column prop="title" label="标题" min-width="180" />
        <el-table-column prop="cateName" label="分类" width="100" />
        <el-table-column prop="order" label="排序" width="60" />
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">{{ row.status === 1 ? '启用' : '停用' }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column label="推荐" width="80" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.vouch === 1" type="warning" size="small">推荐</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="340" fixed="right">
          <template #default="{ row }">
            <div class="table-actions">
              <el-button v-if="hasPerm('news:edit')" size="small" type="primary" @click="showEdit(row)">编辑</el-button>
              <el-button v-if="(row.status === '1') && hasPerm('news:edit')" size="small" type="warning" @click="toggleStatus(row, '0')">停用</el-button>
              <el-button v-else-if="hasPerm('news:edit')" size="small" type="success" @click="toggleStatus(row, '1')">启用</el-button>
              <el-dropdown v-if="hasPerm('news:edit')" trigger="click" @command="(cmd:string)=>handleMore(cmd,row)">
                <el-button size="small">更多<el-icon><ArrowDown /></el-icon></el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item :command="row.vouch === 1 ? 'unvouch' : 'vouch'">
                      {{ row.vouch === 1 ? '取消推荐' : '推荐首页' }}
                    </el-dropdown-item>
                    <el-dropdown-item v-if="hasPerm('news:del')" command="del" divided>删除</el-dropdown-item>
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
    <el-dialog v-model="formDialog.visible" :title="formDialog.title" width="700px" :close-on-click-modal="false">
      <el-form ref="formRef" :model="form" label-width="100px">
        <el-form-item label="标题" prop="title">
          <el-input v-model="form.title" placeholder="必填" />
        </el-form-item>
        <el-form-item label="分类">
          <el-select v-model="form.cateName" style="width:200px">
            <el-option v-for="c in categories" :key="c.value" :label="c.label" :value="c.label" />
          </el-select>
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.order" :min="0" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.desc" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="内容">
          <el-input v-model="form.content" type="textarea" :rows="6" />
        </el-form-item>
        <el-form-item label="封面图片">
          <el-upload action="/upload" :show-file-list="false" :on-success="handleCoverSuccess" :on-error="()=>ElMessage.error('上传失败')" :headers="{ Authorization: token }" accept="image/*">
            <div class="cover-upload">
              <el-image v-if="form.img" :src="form.img" class="cover-preview" />
              <div v-else class="cover-placeholder">+</div>
              <div v-if="form.img" class="cover-overlay" @click.stop>
                <el-button size="small" type="danger" :icon="Delete" circle @click.stop="form.img=''" />
              </div>
            </div>
          </el-upload>
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
      </el-form>
      <template #footer>
        <el-button @click="formDialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveNews">{{ formDialog.isCreate ? '创建' : '保存' }}</el-button>
      </template>
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
  { label: '类型', field: 'type' },
  { label: '排序', field: 'order' },
  { label: '状态', field: 'status' },
  { label: '推荐', field: 'vouch' },
  { label: '创建时间', field: 'addTime' },
]

function handleCoverSuccess(res: any) {
  if (res.data?.url) form.img = res.data.url
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

async function loadCategories() {
  try {
    const res = await adminApi.dictItems('content_type')
    categories.value = (res.data || []).map((d: any) => ({ label: d.label, value: d.value }))
  } catch { categories.value = [] }
}

async function loadDeptTree() {
  try {
    const res = await adminApi.deptTree()
    deptTree.value = res.data || []
  } catch { deptTree.value = [] }
}

async function load() {
  loading.value = true
  try {
    const params: any = { page: page.value, pageSize: pageSize.value, keyword: keyword.value }
    if (sortRules.value.length) params.sort = sortRules.value.map(s => s.field + ':' + s.order).join(',')
    const res = await adminApi.newsList(params)
    list.value = res.data?.list || []
    total.value = res.data?.total || 0
  } catch { list.value = []; total.value = 0 }
  loading.value = false
}
function onSortChange() { page.value = 1; load() }
function search() { page.value = 1; load() }

const formDialog = reactive({ visible: false, title: '', isCreate: false })
const form = reactive({ id: null as any, title: '', cateName: '', order: 9999, desc: '', content: '', img: '', publishDeptIds: '' })
function resetForm() {
  Object.assign(form, { id: null, title: '', cateName: '', order: 9999, desc: '', content: '', img: '', publishDeptIds: '' })
}
function showAdd() {
  resetForm()
  formDialog.isCreate = true
  formDialog.title = '添加通知'
  formDialog.visible = true
}
async function showEdit(row: any) {
  resetForm()
  formDialog.isCreate = false
  formDialog.title = '编辑通知'
  formDialog.visible = true
  try {
    const res = await adminApi.newsDetail(row.id)
    const d = res.data || {}
    form.id = d.id
    form.title = d.title || ''
    form.cateName = d.cateName || ''
    form.order = d.order ?? 9999
    form.desc = d.desc || ''
    form.content = d.content || ''
    form.img = d.img || ''
    form.publishDeptIds = d.publishDeptIds || ''
  } catch {}
}
async function saveNews() {
  if (!form.title) { ElMessage.warning('请输入标题'); return }
  saving.value = true
  try {
    const payload: any = {
      title: form.title, cateName: form.cateName, order: form.order,
      desc: form.desc, content: form.content, img: form.img,
      deptId: 0, publishDeptIds: form.publishDeptIds
    }
    if (formDialog.isCreate) {
      await adminApi.newsInsert(payload)
    } else {
      payload.id = form.id
      await adminApi.newsEdit(payload)
    }
    ElMessage.success('保存成功')
    formDialog.visible = false
    load()
  } finally { saving.value = false }
}

async function toggleStatus(row: any, status: string) {
  await adminApi.newsStatus({ id: row.id, status })
  ElMessage.success(status === '1' ? '已启用' : '已停用')
  load()
}

async function handleMore(cmd: string, row: any) {
  if (cmd === 'vouch') {
    await adminApi.newsVouch({ id: row.id, isVouch: '1' })
    ElMessage.success('已推荐'); load()
  } else if (cmd === 'unvouch') {
    await adminApi.newsVouch({ id: row.id, isVouch: '0' })
    ElMessage.success('已取消推荐'); load()
  } else if (cmd === 'del') {
    try {
      await ElMessageBox.confirm('确定删除该通知？', '提示')
      await adminApi.newsDel({ id: row.id })
      ElMessage.success('已删除'); load()
    } catch {}
  }
}

async function delSelected() {
  if (selected.value.length === 0) return
  try {
    await ElMessageBox.confirm(`确定删除选中的 ${selected.value.length} 条通知？`, '提示')
    const ids = selected.value.map((r: any) => r.id).join(',')
    await adminApi.newsDels({ ids })
    ElMessage.success('已删除')
    selected.value = []
    load()
  } catch {}
}

function exportData() {
  const headers = ['标题', '分类', '排序', '状态', '推荐']
  const rows = [headers]
  list.value.forEach((r: any) => {
    rows.push([r.title, r.cateName || '', r.order || '0', r.status === 1 ? '正常' : '停用', r.isVouch === 1 ? '是' : '否'])
  })
  const csv = '\uFEFF' + rows.map(r => r.map(v => '"' + String(v).replace(/"/g, '""') + '"').join(',')).join('\n')
  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
  const a = document.createElement('a')
  a.href = URL.createObjectURL(blob)
  a.download = '通知列表.csv'
  a.click()
  URL.revokeObjectURL(a.href)
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
</style>
<style>
.el-tree { text-align: left; }
</style>
