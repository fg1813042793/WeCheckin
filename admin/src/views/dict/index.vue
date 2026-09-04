<template>
  <div class="admin-page dict-page">
    <el-card class="admin-card dict-shell" shadow="never">
      <div class="admin-toolbar dict-toolbar">
        <div class="admin-toolbar__left">
          <el-input v-model="typeKeyword" prefix-icon="Search" placeholder="搜索类型名称或编码" clearable class="dict-search" />
          <el-button v-if="hasPerm('admin:menu:dict:add')" type="primary" icon="Plus" @click="showTypeAdd">新增类型</el-button>
        </div>
        <div class="admin-toolbar__right">
          <el-tooltip content="刷新" placement="bottom">
            <el-button circle icon="Refresh" :loading="typeLoading" @click="loadTypes()" />
          </el-tooltip>
          <el-tooltip content="导出类型" placement="bottom">
            <el-button circle icon="Download" :disabled="types.length === 0" @click="exportTypes" />
          </el-tooltip>
        </div>
      </div>

      <div class="dict-workspace">
        <section class="dict-types-pane">
          <div class="pane-heading">
            <span>字典类型</span>
            <span class="pane-count">{{ filteredTypes.length }}</span>
          </div>
          <el-table
            v-loading="typeLoading"
            :data="filteredTypes"
            height="100%"
            row-key="typeCode"
            highlight-current-row
            :current-row-key="selectedType?.typeCode"
            :row-class-name="typeRowClassName"
            @row-click="selectType"
          >
            <el-table-column min-width="190">
              <template #default="{ row }">
                <div class="type-cell">
                  <div class="type-cell__title">
                    <span>{{ row.typeName }}</span>
                    <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">{{ row.status === 1 ? '启用' : '停用' }}</el-tag>
                  </div>
                  <div class="type-cell__meta">
                    <code>{{ row.typeCode }}</code>
                    <span>{{ row.itemCnt }} 项</span>
                  </div>
                </div>
              </template>
            </el-table-column>
            <el-table-column width="82" align="right">
              <template #default="{ row }">
                <div class="compact-actions" @click.stop>
                  <el-tooltip v-if="hasPerm('admin:menu:dict:edit')" content="编辑类型">
                    <el-button link icon="Edit" @click="showTypeEdit(row)" />
                  </el-tooltip>
                  <el-tooltip v-if="hasPerm('admin:menu:dict:del')" content="删除类型">
                    <el-button link type="danger" icon="Delete" @click="deleteType(row)" />
                  </el-tooltip>
                </div>
              </template>
            </el-table-column>
            <template #empty>
              <el-empty :image-size="72" description="暂无字典类型" />
            </template>
          </el-table>
        </section>

        <section class="dict-items-pane">
          <template v-if="selectedType">
            <div class="items-heading">
              <div class="items-heading__identity">
                <div class="items-heading__title">
                  <span>{{ selectedType.typeName }}</span>
                  <el-tag :type="selectedType.status === 1 ? 'success' : 'info'" size="small">{{ selectedType.status === 1 ? '启用' : '停用' }}</el-tag>
                </div>
                <code>{{ selectedType.typeCode }}</code>
              </div>
              <div class="items-heading__actions">
                <el-input v-model="itemKeyword" prefix-icon="Search" placeholder="搜索标签或值" clearable class="item-search" />
                <el-button v-if="hasPerm('admin:menu:dict:add')" type="primary" icon="Plus" @click="showItemAdd">新增数据</el-button>
                <el-button
                  v-if="hasPerm('admin:menu:dict:del')"
                  icon="Delete"
                  :disabled="items.length === 0"
                  @click="clearTypeItems"
                >清空数据</el-button>
              </div>
            </div>

            <el-table v-loading="itemLoading" :data="filteredItems" height="100%" row-key="id" stripe>
              <el-table-column prop="label" label="标签" min-width="150" show-overflow-tooltip />
              <el-table-column prop="value" label="值" min-width="140" show-overflow-tooltip>
                <template #default="{ row }"><code class="item-value">{{ row.value }}</code></template>
              </el-table-column>
              <el-table-column prop="sort" label="排序" width="72" align="center" />
              <el-table-column label="状态" width="84" align="center">
                <template #default="{ row }">
                  <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">{{ row.status === 1 ? '启用' : '停用' }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="remark" label="备注" min-width="160" show-overflow-tooltip />
              <el-table-column label="操作" width="96" fixed="right" align="right">
                <template #default="{ row }">
                  <div class="compact-actions">
                    <el-tooltip v-if="hasPerm('admin:menu:dict:edit')" content="编辑数据">
                      <el-button link icon="Edit" @click="showItemEdit(row)" />
                    </el-tooltip>
                    <el-popconfirm v-if="hasPerm('admin:menu:dict:del')" title="确定删除这条字典数据？" @confirm="deleteItem(row)">
                      <template #reference>
                        <el-button link type="danger" icon="Delete" />
                      </template>
                    </el-popconfirm>
                  </div>
                </template>
              </el-table-column>
              <template #empty>
                <el-empty :image-size="88" description="暂无字典数据" />
              </template>
            </el-table>
          </template>
          <el-empty v-else class="type-empty" :image-size="100" description="请选择或新增字典类型" />
        </section>
      </div>
    </el-card>

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

async function loadTypes(preferredTypeCode = selectedType.value?.typeCode || '') {
  typeLoading.value = true
  try {
    const response = await adminApi.dictTypes()
    types.value = Array.isArray(response.data) ? response.data : []
    const nextType = types.value.find(item => item.typeCode === preferredTypeCode) || types.value[0] || null
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

function typeRowClassName({ row }: { row: DictTypeSummary }) {
  return row.typeCode === selectedType.value?.typeCode ? 'is-selected-type' : ''
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
    await loadTypes(typeCode)
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
const itemForm = reactive<DictItemPayload & { id?: number }>({ typeCode: '', label: '', value: '', sort: 0, status: 1, remark: '' })
const itemRules: FormRules = {
  label: [{ required: true, message: '请输入标签', trigger: 'blur' }],
  value: [{ required: true, message: '请输入值', trigger: 'blur' }],
}

function showItemAdd() {
  if (!selectedType.value) return
  Object.assign(itemForm, { id: undefined, typeCode: selectedType.value.typeCode, label: '', value: '', sort: 0, status: 1, remark: '' })
  itemDialog.isCreate = true
  itemDialog.visible = true
}

function showItemEdit(row: DictItem) {
  Object.assign(itemForm, { id: row.id, typeCode: row.typeCode, label: row.label, value: row.value, sort: row.sort, status: row.status, remark: row.remark || '' })
  itemDialog.isCreate = false
  itemDialog.visible = true
}

async function saveItem() {
  if (!await itemRef.value?.validate().catch(() => false)) return
  const payload: DictItemPayload = {
    typeCode: itemForm.typeCode,
    label: itemForm.label.trim(),
    value: itemForm.value.trim(),
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
.dict-shell { min-height: calc(100vh - 126px); }
.dict-toolbar { margin-bottom: 14px; }
.dict-search { width: 300px; }
.dict-workspace { display: grid; grid-template-columns: minmax(320px, 36%) minmax(0, 1fr); min-height: 590px; height: calc(100vh - 210px); border: 1px solid var(--admin-border); border-radius: 6px; overflow: hidden; background: #fff; }
.dict-types-pane, .dict-items-pane { min-width: 0; min-height: 0; }
.dict-types-pane { display: grid; grid-template-rows: auto minmax(0, 1fr); border-right: 1px solid var(--admin-border); }
.dict-items-pane { display: grid; grid-template-rows: auto minmax(0, 1fr); }
.pane-heading { display: flex; align-items: center; justify-content: space-between; min-height: 52px; padding: 0 16px; border-bottom: 1px solid var(--admin-border); font-size: 15px; font-weight: 650; color: var(--admin-text); }
.pane-count { min-width: 24px; color: var(--admin-muted); font-size: 12px; font-weight: 500; text-align: right; }
.type-cell { display: grid; gap: 5px; padding: 4px 0; }
.type-cell__title, .type-cell__meta, .items-heading__title { display: flex; align-items: center; justify-content: space-between; gap: 10px; min-width: 0; }
.type-cell__title > span, .items-heading__title > span { overflow: hidden; font-weight: 600; color: var(--admin-text); text-overflow: ellipsis; white-space: nowrap; }
.type-cell__meta { color: var(--admin-muted); font-size: 12px; }
.type-cell__meta code, .items-heading__identity code, .item-value { overflow: hidden; color: #52606d; font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; text-overflow: ellipsis; white-space: nowrap; }
.compact-actions { display: flex; align-items: center; justify-content: flex-end; gap: 2px; }
.compact-actions :deep(.el-button) { width: 30px; height: 30px; margin: 0; }
.items-heading { display: flex; min-height: 72px; align-items: center; justify-content: space-between; gap: 20px; padding: 10px 16px; border-bottom: 1px solid var(--admin-border); }
.items-heading__identity { display: grid; flex: 0 1 240px; gap: 5px; min-width: 150px; }
.items-heading__identity code { color: var(--admin-muted); font-size: 12px; }
.items-heading__actions { display: flex; align-items: center; justify-content: flex-end; gap: 8px; min-width: 0; }
.item-search { width: 210px; }
.type-empty { place-self: center; grid-row: 1 / -1; }
.dialog-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 14px; }
.dialog-grid--compact { grid-template-columns: 180px minmax(0, 1fr); }
:deep(.dict-types-pane .el-table__row) { cursor: pointer; }
:deep(.dict-types-pane .el-table__row.is-selected-type > td.el-table__cell) { background: #eef6ff; }
@media (max-width: 1040px) {
  .dict-workspace { grid-template-columns: minmax(290px, 38%) minmax(0, 1fr); }
  .items-heading { align-items: flex-start; flex-direction: column; gap: 10px; }
  .items-heading__identity { flex-basis: auto; width: 100%; }
  .items-heading__actions { width: 100%; justify-content: flex-start; flex-wrap: wrap; }
}
@media (max-width: 760px) {
  .dict-search { width: min(100%, 300px); }
  .dict-workspace { display: grid; grid-template-columns: 1fr; height: auto; min-height: 0; overflow: visible; }
  .dict-types-pane { height: 360px; border-right: 0; border-bottom: 1px solid var(--admin-border); }
  .dict-items-pane { height: 520px; }
  .item-search { width: 100%; }
  .dialog-grid, .dialog-grid--compact { grid-template-columns: 1fr; gap: 0; }
}
</style>
