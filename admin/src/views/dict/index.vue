<template>
  <div class="admin-page dict-page">
    <template v-if="!selectedType">
      <div class="dict-toolbar-card">
        <div class="admin-toolbar dict-search-toolbar">
          <div class="admin-toolbar__left">
            <el-input v-model="typeKeyword" placeholder="搜索类型名称或编码" clearable class="dict-search" @keyup.enter="searchTypes" />
            <el-button type="primary" @click="searchTypes">搜索</el-button>
          </div>
        </div>
        <div class="admin-toolbar dict-actions-toolbar">
          <div class="admin-toolbar__left">
            <el-button v-if="hasPerm('admin:menu:dict:add')" type="success" icon="Plus" @click="showTypeAdd">新增类型</el-button>
          </div>
          <div class="admin-toolbar__right">
            <el-button circle icon="Refresh" title="刷新" :loading="typeLoading" @click="loadTypes()" />
            <el-button circle icon="Download" title="导出" :disabled="types.length === 0" @click="exportTypes" />
          </div>
        </div>
      </div>

      <div class="dict-workspace">
        <el-table
          class="dict-type-table"
          v-loading="typeLoading"
          :data="filteredTypes"
          row-key="typeCode"
          stripe
          empty-text="暂无字典类型"
          style="width:100%"
          @row-click="selectType"
        >
          <el-table-column prop="typeName" label="类型名称" min-width="180" />
          <el-table-column prop="typeCode" label="类型编码" min-width="180">
            <template #default="{ row }"><code class="item-value">{{ row.typeCode }}</code></template>
          </el-table-column>
          <el-table-column prop="itemCnt" label="数据数量" width="100" align="center" />
          <el-table-column label="状态" width="90" align="center">
            <template #default="{ row }">
              <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">{{ row.status === 1 ? '启用' : '停用' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="remark" label="备注" min-width="180" show-overflow-tooltip />
          <el-table-column label="操作" width="260" fixed="right">
            <template #default="{ row }">
              <div class="admin-table-actions" @click.stop>
                <el-button size="small" type="primary" @click="selectType(row)">查看数据</el-button>
                <el-button v-if="hasPerm('admin:menu:dict:edit')" size="small" @click="showTypeEdit(row)">编辑</el-button>
                <el-button v-if="hasPerm('admin:menu:dict:del')" size="small" type="danger" @click="deleteType(row)">删除</el-button>
              </div>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </template>

    <template v-else>
      <div class="dict-toolbar-card">
        <div class="admin-toolbar dict-search-toolbar">
          <div class="admin-toolbar__left">
            <el-button icon="ArrowLeft" @click="backToTypes">返回字典类型</el-button>
            <el-input v-model="itemKeyword" placeholder="搜索标签、值或备注" clearable class="dict-search" @keyup.enter="searchItems" />
            <el-button type="primary" @click="searchItems">搜索</el-button>
          </div>
        </div>
        <div class="admin-toolbar dict-actions-toolbar">
          <div class="admin-toolbar__left">
            <el-button v-if="hasPerm('admin:menu:dict:add')" type="success" icon="Plus" @click="showItemAdd">新增数据</el-button>
            <el-button v-if="hasPerm('admin:menu:dict:del')" type="danger" plain :disabled="items.length === 0" @click="clearTypeItems">清空数据</el-button>
            <span class="dict-current-type">{{ selectedType.typeName }} <code>{{ selectedType.typeCode }}</code></span>
          </div>
          <div class="admin-toolbar__right">
            <el-button circle icon="Refresh" title="刷新" :loading="itemLoading" @click="loadItems()" />
          </div>
        </div>
      </div>

      <div class="dict-workspace">
        <el-table v-loading="itemLoading" :data="filteredItems" row-key="id" stripe empty-text="暂无字典数据" style="width:100%">
          <el-table-column prop="label" label="标签" min-width="150" show-overflow-tooltip />
          <el-table-column prop="value" label="值" min-width="140" show-overflow-tooltip>
            <template #default="{ row }"><code class="item-value">{{ row.value }}</code></template>
          </el-table-column>
          <el-table-column prop="calculationValue" label="计算值" width="100" align="right">
            <template #default="{ row }"><span>{{ row.calculationValue ?? '-' }}</span></template>
          </el-table-column>
          <el-table-column prop="sort" label="排序" width="90" align="center" />
          <el-table-column label="状态" width="90" align="center">
            <template #default="{ row }">
              <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">{{ row.status === 1 ? '启用' : '停用' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="remark" label="备注" min-width="180" show-overflow-tooltip />
          <el-table-column label="操作" width="180" fixed="right">
            <template #default="{ row }">
              <div class="admin-table-actions">
                <el-button v-if="hasPerm('admin:menu:dict:edit')" size="small" type="primary" @click="showItemEdit(row)">编辑</el-button>
                <el-popconfirm v-if="hasPerm('admin:menu:dict:del')" title="确定删除这条字典数据？" @confirm="deleteItem(row)">
                  <template #reference><el-button size="small" type="danger">删除</el-button></template>
                </el-popconfirm>
              </div>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </template>

    <el-dialog
      v-model="typeDialog.visible"
      :title="typeDialog.isCreate ? '新增字典类型' : '编辑字典类型'"
      width="min(520px, 94vw)"
      append-to-body
      :close-on-click-modal="false"
    >
      <el-form ref="typeRef" :model="typeForm" :rules="typeRules" label-position="top">
        <el-form-item label="类型编码" prop="typeCode">
          <el-input v-model="typeForm.typeCode" :disabled="!typeDialog.isCreate" maxlength="50" placeholder="例如 content_type" />
        </el-form-item>
        <el-form-item label="类型名称" prop="typeName">
          <el-input v-model="typeForm.typeName" maxlength="100" placeholder="例如 内容分类" />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="typeForm.status" :active-value="1" :inactive-value="0" active-text="启用" inactive-text="停用" />
        </el-form-item>
        <el-form-item label="备注" prop="remark">
          <el-input v-model="typeForm.remark" type="textarea" :rows="3" maxlength="500" show-word-limit />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="typeDialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveType">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="itemDialog.visible"
      :title="itemDialog.isCreate ? '新增字典数据' : '编辑字典数据'"
      width="min(560px, 94vw)"
      append-to-body
      :close-on-click-modal="false"
    >
      <el-form ref="itemRef" :model="itemForm" :rules="itemRules" label-position="top">
        <div class="dialog-grid">
          <el-form-item label="标签" prop="label">
            <el-input v-model="itemForm.label" maxlength="100" placeholder="页面显示名称" />
          </el-form-item>
          <el-form-item label="值" prop="value">
            <el-input v-model="itemForm.value" maxlength="200" placeholder="业务存储值" />
          </el-form-item>
          <el-form-item label="计算值">
            <el-input-number v-model="itemForm.calculationValue" :min="-999999999" :max="999999999" :precision="6" controls-position="right" placeholder="可选，供流程计算使用" style="width:100%" />
          </el-form-item>
        </div>
        <div class="dialog-grid dialog-grid--compact">
          <el-form-item label="排序">
            <el-input-number v-model="itemForm.sort" :min="0" :max="999999" controls-position="right" />
          </el-form-item>
          <el-form-item label="状态">
            <el-switch v-model="itemForm.status" :active-value="1" :inactive-value="0" active-text="启用" inactive-text="停用" />
          </el-form-item>
        </div>
        <el-form-item label="备注" prop="remark">
          <el-input v-model="itemForm.remark" type="textarea" :rows="3" maxlength="500" show-word-limit />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="itemDialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveItem">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { adminApi } from '../../api'
import type { DictItem, DictItemPayload, DictTypePayload, DictTypeSummary } from '../../api/types'
import { hasPerm } from '../../utils/permission'

const typeLoading = ref(false)
const itemLoading = ref(false)
const saving = ref(false)
const types = ref<DictTypeSummary[]>([])
const items = ref<DictItem[]>([])
const selectedType = ref<DictTypeSummary | null>(null)
const typeKeyword = ref('')
const itemKeyword = ref('')
let itemRequestVersion = 0

const filteredTypes = computed(() => {
  const keyword = typeKeyword.value.trim().toLowerCase()
  if (!keyword) return types.value
  return types.value.filter(item => item.typeName.toLowerCase().includes(keyword) || item.typeCode.toLowerCase().includes(keyword))
})

const filteredItems = computed(() => {
  const keyword = itemKeyword.value.trim().toLowerCase()
  if (!keyword) return items.value
  return items.value.filter(item => (
    item.label.toLowerCase().includes(keyword)
    || item.value.toLowerCase().includes(keyword)
    || (item.remark || '').toLowerCase().includes(keyword)
  ))
})

function searchTypes() {
  typeKeyword.value = typeKeyword.value.trim()
}

function searchItems() {
  itemKeyword.value = itemKeyword.value.trim()
}

function backToTypes() {
  itemRequestVersion += 1
  selectedType.value = null
  items.value = []
  itemKeyword.value = ''
  itemLoading.value = false
}

async function loadTypes(preferredTypeCode = selectedType.value?.typeCode || '') {
  typeLoading.value = true
  try {
    const response = await adminApi.dictTypes()
    types.value = Array.isArray(response.data) ? response.data : []
    const nextType = preferredTypeCode ? types.value.find(item => item.typeCode === preferredTypeCode) || null : null
    if (nextType) await selectType(nextType)
    else {
      selectedType.value = null
      items.value = []
    }
  } finally {
    typeLoading.value = false
  }
}

async function selectType(row: DictTypeSummary) {
  selectedType.value = row
  itemKeyword.value = ''
  await loadItems(row.typeCode)
}

async function loadItems(typeCode = selectedType.value?.typeCode || '') {
  if (!typeCode) {
    items.value = []
    return
  }
  const requestVersion = ++itemRequestVersion
  itemLoading.value = true
  try {
    const response = await adminApi.dictItems(typeCode)
    if (requestVersion === itemRequestVersion) items.value = Array.isArray(response.data) ? response.data : []
  } finally {
    if (requestVersion === itemRequestVersion) itemLoading.value = false
  }
}

const typeRef = ref<FormInstance>()
const typeDialog = reactive({ visible: false, isCreate: true })
const typeForm = reactive<DictTypePayload>({ typeCode: '', typeName: '', status: 1, remark: '' })
const typeRules: FormRules = {
  typeCode: [
    { required: true, message: '请输入类型编码', trigger: 'blur' },
    { validator: validateTypeCodeRule, trigger: 'blur' },
  ],
  typeName: [{ required: true, message: '请输入类型名称', trigger: 'blur' }],
}

function validateTypeCodeRule(_rule: unknown, value: string, callback: (error?: Error) => void) {
  if (!typeDialog.isCreate || /^[a-z][a-z0-9._-]{0,49}$/.test(value.trim())) callback()
  else callback(new Error('以小写字母开头，只能使用小写字母、数字、点、下划线和连字符'))
}

function showTypeAdd() {
  Object.assign(typeForm, { typeCode: '', typeName: '', status: 1, remark: '' })
  typeDialog.isCreate = true
  typeDialog.visible = true
}

function showTypeEdit(row: DictTypeSummary) {
  Object.assign(typeForm, { typeCode: row.typeCode, typeName: row.typeName, status: row.status, remark: row.remark || '' })
  typeDialog.isCreate = false
  typeDialog.visible = true
}

async function saveType() {
  if (!await typeRef.value?.validate().catch(() => false)) return
  const typeCode = typeForm.typeCode.trim()
  saving.value = true
  try {
    if (typeDialog.isCreate) {
      await adminApi.dictTypeAdd({ ...typeForm, typeCode, typeName: typeForm.typeName.trim(), remark: typeForm.remark.trim() })
    } else {
      await adminApi.dictTypeEdit(typeCode, { typeName: typeForm.typeName.trim(), status: typeForm.status, remark: typeForm.remark.trim() })
    }
    ElMessage.success(typeDialog.isCreate ? '字典类型已创建' : '字典类型已更新')
    typeDialog.visible = false
    await loadTypes(selectedType.value ? typeCode : '')
  } finally {
    saving.value = false
  }
}

async function deleteType(row: DictTypeSummary) {
  try {
    await ElMessageBox.confirm(`删除“${row.typeName}”将同时删除其 ${row.itemCnt} 条数据，且不可恢复。`, '删除字典类型', { type: 'warning' })
  } catch {
    return
  }
  await adminApi.dictTypeDelete(row.typeCode)
  ElMessage.success('字典类型已删除')
  await loadTypes(row.typeCode === selectedType.value?.typeCode ? '' : selectedType.value?.typeCode)
}

async function clearTypeItems() {
  if (!selectedType.value) return
  try {
    await ElMessageBox.confirm(`确定清空“${selectedType.value.typeName}”的全部字典数据？字典类型会保留。`, '清空字典数据', { type: 'warning' })
  } catch {
    return
  }
  await adminApi.dictTypeClearItems(selectedType.value.typeCode)
  ElMessage.success('字典数据已清空')
  await loadTypes(selectedType.value.typeCode)
}

const itemRef = ref<FormInstance>()
const itemDialog = reactive({ visible: false, isCreate: true })
const itemForm = reactive<DictItemPayload & { id?: number }>({ typeCode: '', label: '', value: '', calculationValue: undefined, sort: 0, status: 1, remark: '' })
const itemRules: FormRules = {
  label: [{ required: true, message: '请输入标签', trigger: 'blur' }],
  value: [{ required: true, message: '请输入值', trigger: 'blur' }],
}

function showItemAdd() {
  if (!selectedType.value) return
  Object.assign(itemForm, { id: undefined, typeCode: selectedType.value.typeCode, label: '', value: '', calculationValue: undefined, sort: 0, status: 1, remark: '' })
  itemDialog.isCreate = true
  itemDialog.visible = true
}

function showItemEdit(row: DictItem) {
  Object.assign(itemForm, { id: row.id, typeCode: row.typeCode, label: row.label, value: row.value, calculationValue: row.calculationValue, sort: row.sort, status: row.status, remark: row.remark || '' })
  itemDialog.isCreate = false
  itemDialog.visible = true
}

async function saveItem() {
  if (!await itemRef.value?.validate().catch(() => false)) return
  const payload: DictItemPayload = {
    typeCode: itemForm.typeCode,
    label: itemForm.label.trim(),
    value: itemForm.value.trim(),
    calculationValue: typeof itemForm.calculationValue === 'number' && Number.isFinite(itemForm.calculationValue) ? itemForm.calculationValue : undefined,
    sort: itemForm.sort,
    status: itemForm.status,
    remark: itemForm.remark.trim(),
  }
  saving.value = true
  try {
    if (itemDialog.isCreate) await adminApi.dictAdd(payload)
    else if (itemForm.id) await adminApi.dictEdit({ ...payload, id: itemForm.id })
    ElMessage.success(itemDialog.isCreate ? '字典数据已添加' : '字典数据已更新')
    itemDialog.visible = false
    await loadTypes(selectedType.value?.typeCode || '')
  } finally {
    saving.value = false
  }
}

async function deleteItem(row: DictItem) {
  await adminApi.dictDel({ id: row.id })
  ElMessage.success('字典数据已删除')
  await loadTypes(selectedType.value?.typeCode || '')
}

function exportTypes() {
  const rows = [['类型名称', '类型编码', '状态', '数据条数', '备注']]
  types.value.forEach(item => rows.push([item.typeName, item.typeCode, item.status === 1 ? '启用' : '停用', String(item.itemCnt), item.remark || '']))
  const csv = '\uFEFF' + rows.map(row => row.map(value => `"${String(value).replace(/"/g, '""')}"`).join(',')).join('\n')
  const url = URL.createObjectURL(new Blob([csv], { type: 'text/csv;charset=utf-8;' }))
  const link = document.createElement('a')
  link.href = url
  link.download = '字典类型.csv'
  link.click()
  URL.revokeObjectURL(url)
}

onMounted(() => loadTypes())
</script>

<style scoped>
.dict-search { width: 300px; }
.dict-toolbar-card { overflow: hidden; border: 1px solid var(--admin-ui-color-border); border-radius: 8px; background: var(--admin-ui-color-surface); }
.dict-toolbar-card .admin-toolbar { margin: 0; border: 0; border-radius: 0; }
.dict-toolbar-card .dict-search-toolbar { border-bottom: 1px solid var(--admin-ui-color-border); }
.dict-workspace { overflow: hidden; border: 1px solid var(--admin-ui-color-border); border-radius: 8px; background: var(--admin-ui-color-surface); }
.dict-workspace :deep(.el-table__row) { cursor: default; }
.dict-workspace :deep(.dict-type-table .el-table__row) { cursor: pointer; }
.dict-current-type { display: inline-flex; align-items: center; gap: 8px; color: var(--admin-ui-color-text-secondary); font-size: 13px; }
.dict-current-type code, .item-value { overflow: hidden; color: #52606d; font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; text-overflow: ellipsis; white-space: nowrap; }
.dialog-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 14px; }
.dialog-grid--compact { grid-template-columns: 180px minmax(0, 1fr); }
@media (max-width: 760px) {
  .dict-search { width: min(100%, 300px); }
  .dict-toolbar-card .admin-toolbar { align-items: stretch; flex-direction: column; }
  .dict-toolbar-card .admin-toolbar__left, .dict-toolbar-card .admin-toolbar__right { width: 100%; }
  .dict-current-type { width: 100%; }
  .dialog-grid, .dialog-grid--compact { grid-template-columns: 1fr; gap: 0; }
}
</style>
