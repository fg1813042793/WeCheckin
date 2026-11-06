<script setup lang="ts">
import { onBeforeUnmount, ref, watch } from 'vue'
import { AdminSearchBar, AdminTablePanel } from '@/components/admin-ui'
import { publicPortalApi } from '@/api/publicPortal'
import type { PublicPortalKind, PublicPortalSourceOption } from '@/types/publicPortal'
import { activityCategory, errorText } from './portalModel'

const props = defineProps<{ kind: PublicPortalKind; selectedId: number; disabled: boolean }>()
const emit = defineEmits<{ select: [source: PublicPortalSourceOption] }>()
const keyword = ref('')
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const list = ref<PublicPortalSourceOption[]>([])
const loading = ref(false)
const error = ref('')
let sequence = 0
async function load() {
  if (props.disabled) return
  const current = ++sequence
  loading.value = true
  error.value = ''
  try {
    const { data } = await publicPortalApi.options({ kind: props.kind, keyword: keyword.value.trim(), page: page.value, pageSize: pageSize.value })
    if (current !== sequence) return
    list.value = data.list
    total.value = data.total
  } catch (cause) {
    if (current === sequence) { error.value = errorText(cause, '候选内容加载失败，请重试'); list.value = []; total.value = 0 }
  } finally { if (current === sequence) loading.value = false }
}
function search() { page.value = 1; void load() }
function reset() { keyword.value = ''; search() }
function select(source: PublicPortalSourceOption) { if (!props.disabled && !loading.value) emit('select', source) }
watch(() => props.kind, () => { keyword.value = ''; list.value = []; search() }, { immediate: true })
onBeforeUnmount(() => { ++sequence })
</script>

<template>
  <div class="source-picker">
    <p class="picker-hint">仅列出您有管理权限、已发布且公开的内容。</p>
    <AdminSearchBar :loading="loading" @search="search" @reset="reset">
      <el-form-item><el-input v-model="keyword" clearable :disabled="disabled" maxlength="255" aria-label="搜索来源管理名称或显示名称" placeholder="搜索管理名称或显示名称" /></el-form-item>
    </AdminSearchBar>
    <AdminTablePanel :loading="loading" :error="error" :empty="!list.length" empty-text="没有符合条件的公开内容，请检查发布状态或更换关键词" :min-height="180">
      <el-table :data="list" row-key="sourceId" highlight-current-row>
        <el-table-column label="分类" width="120" show-overflow-tooltip><template #default="{ row }">{{ activityCategory(row.kind, row.category) }}</template></el-table-column>
        <el-table-column label="管理名称" prop="managementTitle" min-width="150" show-overflow-tooltip />
        <el-table-column label="显示名称" prop="displayTitle" min-width="150" show-overflow-tooltip />
        <el-table-column label="ID" prop="sourceId" width="90" />
        <el-table-column label="选择" width="90" fixed="right"><template #default="{ row }"><el-button :type="selectedId === row.sourceId ? 'primary' : 'default'" :disabled="disabled || loading" @click="select(row)">{{ selectedId === row.sourceId ? '已选' : '选择' }}</el-button></template></el-table-column>
      </el-table>
      <template #footer>
        <el-button v-if="error" :disabled="disabled" @click="load">重试加载</el-button>
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50, 100]" layout="total, sizes, prev, pager, next" :disabled="disabled || loading" @current-change="load" @size-change="search" />
      </template>
    </AdminTablePanel>
  </div>
</template>

<style scoped>
.source-picker { display: grid; gap: var(--admin-ui-space-3); min-width: 0; }
.picker-hint { margin: 0; color: var(--admin-ui-color-text-secondary); font-size: var(--admin-ui-font-size-caption); }
.source-picker :deep(.el-pagination) { margin-left: auto; max-width: 100%; overflow-x: auto; }
</style>
