<template>
  <div>
    <el-card>
      <div style="display:flex;gap:10px;margin-bottom:12px">
        <el-input v-model="keyword" placeholder="搜索字典名称" clearable style="width:300px" @keyup.enter="search" />
        <el-button type="primary" @click="search">搜索</el-button>
      </div>
      <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:12px">
        <div>
          <el-button v-if="hasPerm('dict:add')" type="success" @click="showTypeAdd">+ 新增字典类型</el-button>
          <el-button v-if="hasPerm('dict:del')" type="danger" :disabled="selected.length === 0" @click="delSelected">批量删除</el-button>
        </div>
        <div>
          <el-button circle icon="Refresh" title="刷新" @click="loadTypes" />
          <el-button circle icon="Download" title="导出" @click="exportData" />
        </div>
      </div>
      <el-table :data="types" v-loading="loading" stripe style="width:100%" @selection-change="selected = $event">
        <el-table-column type="selection" width="45" />
        <el-table-column prop="typeCode" label="类型编码" width="160" />
        <el-table-column prop="typeName" label="类型名称" width="180" />
        <el-table-column prop="itemCnt" label="数据条数" width="100" />
        <el-table-column label="操作" width="320">
          <template #default="{ row }">
            <div class="table-actions">
              <el-button v-if="hasPerm('dict:list')" size="small" type="primary" @click="showItems(row)">管理数据</el-button>
              <el-button v-if="hasPerm('dict:edit')" size="small" @click="showTypeEdit(row)">编辑</el-button>
              <el-button v-if="hasPerm('dict:del')" size="small" type="danger" @click="clearType(row)">清空</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 新增/编辑字典类型 -->
    <el-dialog v-model="typeDialog.visible" :title="typeDialog.title" width="420px">
      <el-form ref="typeRef" :model="typeForm" label-width="80px">
        <el-form-item label="类型编码" prop="typeCode">
          <el-input v-model="typeForm.typeCode" placeholder="如 news_category" />
        </el-form-item>
        <el-form-item label="类型名称" prop="typeName">
          <el-input v-model="typeForm.typeName" placeholder="如 新闻分类" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="typeDialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveType">{{ typeDialog.isCreate ? '确定' : '保存' }}</el-button>
      </template>
    </el-dialog>

    <!-- 字典数据管理 -->
    <el-dialog v-model="itemDialog.visible" :title="'字典数据 - ' + currentTypeName" width="700px">
      <div style="margin-bottom:12px">
        <el-button v-if="hasPerm('dict:add')" type="success" @click="showItemAdd">+ 新增数据</el-button>
      </div>
      <el-table :data="items" stripe style="width:100%">
        <el-table-column prop="value" label="值" width="120" />
        <el-table-column prop="label" label="标签" width="150" />
        <el-table-column prop="sort" label="排序" width="60" />
        <el-table-column label="状态" width="70">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">{{ row.status === 1 ? '启用' : '停用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" min-width="120" />
        <el-table-column label="操作" width="150">
          <template #default="{ row }">
            <div class="table-actions">
              <el-button v-if="hasPerm('dict:edit')" size="small" @click="showItemEdit(row)">编辑</el-button>
              <el-popconfirm v-if="hasPerm('dict:del')" title="确定删除？" @confirm="delItem(row)">
                <template #reference>
                  <el-button size="small" type="danger">删除</el-button>
                </template>
              </el-popconfirm>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <!-- 新增/编辑字典数据 -->
    <el-dialog v-model="itemFormDialog.visible" :title="itemFormDialog.title" width="500px">
      <el-form ref="itemRef" :model="itemForm" label-width="80px">
        <el-form-item label="标签" prop="label">
          <el-input v-model="itemForm.label" placeholder="显示名称" />
        </el-form-item>
        <el-form-item label="值" prop="value">
          <el-input v-model="itemForm.value" placeholder="实际值" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="itemForm.sort" :min="0" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="itemForm.remark" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="itemFormDialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveItem">{{ itemFormDialog.isCreate ? '添加' : '保存' }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script lang="ts" setup>
import { ref, reactive, onMounted } from 'vue'
import { adminApi } from '../../api'
import { ElMessage, ElMessageBox } from 'element-plus'
import { hasPerm } from '../../utils/permission'

const loading = ref(false)
const saving = ref(false)
const types = ref<any[]>([])
const items = ref<any[]>([])
const currentTypeCode = ref('')
const currentTypeName = ref('')
const keyword = ref('')
const selected = ref<any[]>([])

async function loadTypes() {
  loading.value = true
  try {
    const res = await adminApi.dictTypes()
    types.value = res.data || []
  } catch { types.value = [] }
  loading.value = false
}

// 新增/编辑类型
const typeDialog = reactive({ visible: false, title: '', isCreate: false })
const typeForm = reactive({ typeCode: '', typeName: '' })
const oldTypeCode = ref('')
function showTypeAdd() {
  typeForm.typeCode = ''
  typeForm.typeName = ''
  oldTypeCode.value = ''
  typeDialog.isCreate = true
  typeDialog.title = '新增字典类型'
  typeDialog.visible = true
}
function showTypeEdit(row: any) {
  typeForm.typeCode = row.typeCode
  typeForm.typeName = row.typeName
  oldTypeCode.value = row.typeCode
  typeDialog.isCreate = false
  typeDialog.title = '编辑字典类型'
  typeDialog.visible = true
}
async function saveType() {
  if (!typeForm.typeCode || !typeForm.typeName) { ElMessage.warning('请填写完整'); return }
  saving.value = true
  try {
    if (typeDialog.isCreate) {
      await adminApi.dictAdd({ typeCode: typeForm.typeCode, typeName: typeForm.typeName, label: typeForm.typeName, value: '', sort: 0, remark: '' })
    } else {
      await adminApi.dictEditTypeName({ oldTypeCode: oldTypeCode.value, typeCode: typeForm.typeCode, typeName: typeForm.typeName })
    }
    ElMessage.success('保存成功')
    typeDialog.visible = false
    loadTypes()
  } finally { saving.value = false }
}

// 管理数据
const itemDialog = reactive({ visible: false })
async function showItems(row: any) {
  currentTypeCode.value = row.typeCode
  currentTypeName.value = row.typeName
  itemDialog.visible = true
  await loadItems()
}
async function loadItems() {
  try {
    const res = await adminApi.dictItems(currentTypeCode.value)
    items.value = res.data || []
  } catch { items.value = [] }
}

// 清空类型
async function clearType(row: any) {
  try {
    await ElMessageBox.confirm(`确定清空「${row.typeName}」的所有数据？`, '提示')
    await adminApi.dictClear(row.typeCode)
    ElMessage.success('已清空')
    loadTypes()
  } catch {}
}

function search() {
  loadTypes()
}

async function delSelected() {
  if (selected.value.length === 0) return
  try {
    await ElMessageBox.confirm(`确定删除选中的 ${selected.value.length} 个字典类型？`, '提示')
    for (const row of selected.value) {
      await adminApi.dictClear(row.typeCode || row.code)
    }
    ElMessage.success('已删除')
    selected.value = []
    loadTypes()
  } catch {}
}

function exportData() {
  const rows = [['字典名称', '字典类型', '备注']]
  types.value.forEach((r: any) => {
    rows.push([r.typeName || '', r.typeCode || '', r.remark || ''])
  })
  const csv = '\uFEFF' + rows.map(r => r.map(v => '"' + String(v).replace(/"/g, '""') + '"').join(',')).join('\n')
  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
  const a = document.createElement('a')
  a.href = URL.createObjectURL(blob)
  a.download = '字典列表.csv'
  a.click()
  URL.revokeObjectURL(a.href)
}

// 新增/编辑数据
const itemFormDialog = reactive({ visible: false, title: '', isCreate: false })
const itemForm = reactive({ id: null as any, label: '', value: '', sort: 0, remark: '' })
function resetItemForm() {
  Object.assign(itemForm, { id: null, label: '', value: '', sort: 0, remark: '' })
}
function showItemAdd() {
  resetItemForm()
  itemFormDialog.isCreate = true
  itemFormDialog.title = '新增字典数据'
  itemFormDialog.visible = true
}
function showItemEdit(row: any) {
  resetItemForm()
  itemFormDialog.isCreate = false
  itemFormDialog.title = '编辑字典数据'
  itemForm.id = row.id
  itemForm.label = row.label || ''
  itemForm.value = row.value || ''
  itemForm.sort = row.sort ?? 0
  itemForm.remark = row.remark || ''
  itemFormDialog.visible = true
}
async function saveItem() {
  if (!itemForm.label || !itemForm.value) { ElMessage.warning('请填写完整'); return }
  saving.value = true
  try {
    if (itemFormDialog.isCreate) {
      await adminApi.dictAdd({ typeCode: currentTypeCode.value, typeName: currentTypeName.value, label: itemForm.label, value: itemForm.value, sort: itemForm.sort, remark: itemForm.remark })
    } else {
      await adminApi.dictEdit({ id: itemForm.id, label: itemForm.label, value: itemForm.value, sort: itemForm.sort, remark: itemForm.remark })
    }
    ElMessage.success('保存成功')
    itemFormDialog.visible = false
    loadItems()
  } finally { saving.value = false }
}
async function delItem(row: any) {
  await adminApi.dictDel({ id: row.id })
  ElMessage.success('已删除')
  loadItems()
}

onMounted(loadTypes)
</script>
